#!/usr/bin/env sh
set -x

# === Constants ===
VAULT_ENDPOINT=http://127.0.0.1:8600
REGION=us-east-1
PROFILE=scality-internal-services
# CONFIG_FILE=/config/backbeat-config.json

# === Environment Echo ===
echo "[setup] Environment Variables:"
echo "VAULT_ENDPOINT=$VAULT_ENDPOINT"
echo "CONFIG_FILE=$CONFIG_FILE"
echo "PROFILE=$PROFILE"
echo "REGION=$REGION"
echo

# === Wait for Vault service to be ready ===
echo "[setup] Waiting for Vault service at $VAULT_ENDPOINT..."
until curl -s "$VAULT_ENDPOINT" > /dev/null; do
  echo "[setup] Vault not ready, waiting 2s..."
  sleep 2
done
echo "[setup] Vault service is ready!"
echo

# === Configure vaultclient ===
echo "[setup] Configuring vaultclient..."

# Copy test keys to config file
cat ./tests/utils/admincredentials.json | jq 'to_entries | first | {accessKey: .key, secretKeyValue: .value}' > /tmp/vaultclient.conf

export VAULT_CONFIG=/tmp/vaultclient.conf
echo "[setup] vaultclient configured with /tmp/vaultclient.conf"
echo

# === Create management account ===

MANAGEMENT_ACCESS_KEY=$(jq -r '.accessKey' /conf/management-creds.json)
MANAGEMENT_SECRET_KEY=$(jq -r '.secretKey' /conf/management-creds.json)

echo "[setup] Ensure management account is configured..."
resp=$(./node_modules/vaultclient/bin/vaultclient \
        ensure-internal-services-account \
        --host 127.0.0.1 \
        --port 8600 \
        --accesskey "$MANAGEMENT_ACCESS_KEY" \
        --secretkey "$MANAGEMENT_SECRET_KEY")

if [ $? -ne 0 ]; then
  echo "[setup] Error configuring management account:"
  echo "$resp"
  exit 1
fi

echo "[setup] Management account and access key setup completed successfully"

# === Create test data account ===
# This account is seeded with the lifecycle role via vault's accountSeeds,
# which allows backbeat to AssumeRole into it for lifecycle operations.
TEST_ACCOUNT_ACCESS_KEY="WBTKACCESSI9O3YKIRQ0"
TEST_ACCOUNT_SECRET_KEY="ICxmNTBbOqijy4rMq/MOP1EPlTMqfsEBLjROcAbN"

echo "[setup] Creating test data account..."
resp=$(./node_modules/vaultclient/bin/vaultclient \
        create-account \
        --name testaccount \
        --email testaccount@test.com \
        --host 127.0.0.1 \
        --port 8600 2>&1) || {
  if echo "$resp" | grep -q "EntityAlreadyExists"; then
    echo "[setup] Test data account already exists, skipping creation"
  else
    echo "[setup] Error creating test data account:"
    echo "$resp"
    exit 1
  fi
}

echo "[setup] Generating access key for test data account..."
resp=$(./node_modules/vaultclient/bin/vaultclient \
        generate-account-access-key \
        --name testaccount \
        --accesskey "$TEST_ACCOUNT_ACCESS_KEY" \
        --secretkey "$TEST_ACCOUNT_SECRET_KEY" \
        --host 127.0.0.1 \
        --port 8600 2>&1) || {
  if echo "$resp" | grep -q "EntityAlreadyExists"; then
    echo "[setup] Test data account access key already exists, skipping"
  else
    echo "[setup] Error generating access key for test data account:"
    echo "$resp"
    exit 1
  fi
}

echo "[setup] Test data account ready (accessKey=$TEST_ACCOUNT_ACCESS_KEY)"

# === Create lifecycle service user ===
BACKBEAT_CONFIG_FILE=/conf/backbeat/config.json
IAM_ENDPOINT=http://127.0.0.1:8600
export AWS_ACCESS_KEY_ID="$MANAGEMENT_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$MANAGEMENT_SECRET_KEY"
export AWS_DEFAULT_REGION="$REGION"

if [ -f "$BACKBEAT_CONFIG_FILE" ]; then
  echo "[setup] Creating lifecycle service user..."
  resp=$(aws iam create-user \
    --user-name lifecycle \
    --endpoint-url "$IAM_ENDPOINT" 2>&1) || {
    if echo "$resp" | grep -q "EntityAlreadyExists"; then
      echo "[setup] Lifecycle user already exists, skipping creation"
    else
      echo "[setup] Error creating lifecycle user:"
      echo "$resp"
      exit 1
    fi
  }

  echo "[setup] Generating lifecycle access key..."
  LIFECYCLE_CREDS=$(aws iam create-access-key \
    --user-name lifecycle \
    --endpoint-url "$IAM_ENDPOINT" 2>&1) || {
    if echo "$LIFECYCLE_CREDS" | grep -q "LimitExceeded\|EntityAlreadyExists"; then
      echo "[setup] Lifecycle access key already exists, reading from backbeat config"
      SERVICE_ACCESS_KEY=$(jq -r '.extensions.lifecycle.auth.sts.accessKey' "$BACKBEAT_CONFIG_FILE")
      SERVICE_SECRET_KEY=$(jq -r '.extensions.lifecycle.auth.sts.secretKey' "$BACKBEAT_CONFIG_FILE")
      LIFECYCLE_CREDS=""
    else
      echo "[setup] Error generating lifecycle access key:"
      echo "$LIFECYCLE_CREDS"
      exit 1
    fi
  }

  if [ -n "$LIFECYCLE_CREDS" ]; then
    SERVICE_ACCESS_KEY=$(echo "$LIFECYCLE_CREDS" | jq -r '.AccessKey.AccessKeyId')
    SERVICE_SECRET_KEY=$(echo "$LIFECYCLE_CREDS" | jq -r '.AccessKey.SecretAccessKey')
  fi

  echo "[setup] Lifecycle user created"
  echo "[setup] SERVICE_ACCESS_KEY=$SERVICE_ACCESS_KEY"

  # === Grant lifecycle user permission to assume roles ===
  echo "[setup] Creating assume-role policy..."
  ASSUME_POLICY='{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"sts:AssumeRole","Resource":"*"}}'
  resp=$(aws iam create-policy \
    --policy-name lifecycle-assume-role \
    --policy-document "$ASSUME_POLICY" \
    --endpoint-url "$IAM_ENDPOINT" 2>&1) || {
    if echo "$resp" | grep -q "EntityAlreadyExists"; then
      echo "[setup] Assume-role policy already exists, skipping creation"
    else
      echo "[setup] Error creating assume-role policy:"
      echo "$resp"
      exit 1
    fi
  }

  aws iam attach-user-policy \
    --policy-arn "arn:aws:iam::000000000000:policy/lifecycle-assume-role" \
    --user-name lifecycle \
    --endpoint-url "$IAM_ENDPOINT"

  if [ $? -ne 0 ]; then
    echo "[setup] Error attaching assume-role policy to lifecycle user"
    exit 1
  fi
  echo "[setup] Assume-role policy attached to lifecycle user"

  # === Create lifecycle role in the internal services account ===
  # Backbeat hardcodes account 000000000000 for AssumeRole (VAULT-238 workaround),
  # and accountSeeds only apply to accounts created via the normal flow,
  # so we must create the role explicitly in the internal services account.
  echo "[setup] Creating lifecycle role in internal services account..."
  TRUST_POLICY="{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"arn:aws:iam::000000000000:user/lifecycle\"},\"Action\":\"sts:AssumeRole\",\"Condition\":{}}]}"
  resp=$(aws iam create-role \
    --role-name lifecycle-role \
    --path "/scality-internal/" \
    --assume-role-policy-document "$TRUST_POLICY" \
    --endpoint-url "$IAM_ENDPOINT" 2>&1) || {
    if echo "$resp" | grep -q "EntityAlreadyExists"; then
      echo "[setup] Lifecycle role already exists, skipping creation"
    else
      echo "[setup] Error creating lifecycle role:"
      echo "$resp"
      exit 1
    fi
  }

  S3_FULL_ACCESS='{"Version":"2012-10-17","Statement":[{"Sid":"LifecycleFullAccess","Effect":"Allow","Action":["s3:*"],"Resource":["*"]}]}'
  resp=$(aws iam create-policy \
    --policy-name lifecycle-s3-access \
    --policy-document "$S3_FULL_ACCESS" \
    --endpoint-url "$IAM_ENDPOINT" 2>&1) || {
    if echo "$resp" | grep -q "EntityAlreadyExists"; then
      echo "[setup] Lifecycle S3 access policy already exists, skipping creation"
    else
      echo "[setup] Error creating lifecycle S3 access policy:"
      echo "$resp"
      exit 1
    fi
  }

  aws iam attach-role-policy \
    --role-name lifecycle-role \
    --policy-arn "arn:aws:iam::000000000000:policy/lifecycle-s3-access" \
    --endpoint-url "$IAM_ENDPOINT"

  if [ $? -ne 0 ]; then
    echo "[setup] Error setting up lifecycle role"
    exit 1
  fi
  echo "[setup] Lifecycle role created in internal services account"

  # === Update backbeat config.json with lifecycle credentials ===
  echo "[setup] Updating backbeat config.json with lifecycle credentials..."
  jq --arg ak "$SERVICE_ACCESS_KEY" --arg sk "$SERVICE_SECRET_KEY" \
    '.extensions.lifecycle.auth.sts.accessKey = $ak | .extensions.lifecycle.auth.sts.secretKey = $sk' \
    "$BACKBEAT_CONFIG_FILE" > /tmp/backbeat-config.updated.json

  mv /tmp/backbeat-config.updated.json "$BACKBEAT_CONFIG_FILE"
  echo "[setup] Backbeat config.json updated with lifecycle credentials"
fi

echo "[setup] Setup completed successfully"
