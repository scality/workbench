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

echo "[setup] Management account and access key setup completed successfully 🎉"

# === Create acccess key for lifecycle service user ===
# echo "[setup] Creating access key for lifecycle service user..."
# LIFECYCLE_CREDS_JSON=$(jq '.extensions.lifecycle.auth.sts' /config/backbeat/config.json)
# LEFECYCLE_ACCESS_KEY=$(echo "$LIFECYCLE_CREDS_JSON" | jq -r '.accessKey')
# LIFECYCLE_SECRET_KEY=$(echo "$LIFECYCLE_CREDS_JSON" | jq -r '.secretKey')

# # === Generate management access key ===
# echo "[setup] Generating management access key..."
# MGMT_CREDS_JSON=$(./node_modules/vaultclient/bin/vaultclient generate-account-access-key --name management --host vault --port 8600)

# MANAGEMENT_ACCESS_KEY=$(echo "$MGMT_CREDS_JSON" | jq -r '.id')
# MANAGEMENT_SECRET_KEY=$(echo "$MGMT_CREDS_JSON" | jq -r '.value')

# echo "[setup] Management credentials:"
# echo "MANAGEMENT_ACCESS_KEY=$MANAGEMENT_ACCESS_KEY"
# echo "MANAGEMENT_SECRET_KEY=$MANAGEMENT_SECRET_KEY"
# echo

# # === Create lifecycle service user ===
# echo "[setup] Creating lifecycle service user..."
# SERVICE_CREDS_JSON=$(AWS_ACCESS_KEY_ID="$MANAGEMENT_ACCESS_KEY" \
#                       AWS_SECRET_ACCESS_KEY="$MANAGEMENT_SECRET_KEY" \
#                       AWS_REGION="$REGION" \
#                       ./bin/ensureServiceUser apply lifecycle --iam-endpoint http://vault:8600)

# SERVICE_ACCESS_KEY=$(echo "$SERVICE_CREDS_JSON" | jq -r '.data.AccessKeyId')
# SERVICE_SECRET_KEY=$(echo "$SERVICE_CREDS_JSON" | jq -r '.data.SecretAccessKey')

# echo "[setup] Lifecycle service user credentials:"
# echo "SERVICE_ACCESS_KEY=$SERVICE_ACCESS_KEY"
# echo "SERVICE_SECRET_KEY=$SERVICE_SECRET_KEY"
# echo

# # === Update backbeat-config.json ===
# echo "[setup] Updating backbeat-config.json with service user credentials..."
# jq --arg ak "$SERVICE_ACCESS_KEY" --arg sk "$SERVICE_SECRET_KEY" \
#   '.extensions.lifecycle.auth.sts.accessKey = $ak | .extensions.lifecycle.auth.sts.secretKey = $sk' \
#   "$CONFIG_FILE" > /tmp/backbeat-config.updated.json

# mv /tmp/backbeat-config.updated.json "$CONFIG_FILE"
# echo "[setup] backbeat-config.json successfully updated!"
# echo

# # === Done ===
# echo "[setup] Setup service users completed successfully 🎉"
