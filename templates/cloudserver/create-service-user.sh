#!/usr/bin/env sh
set -xe

MANAGEMENT_ACCESS_KEY=$(jq -r '.accessKey' /secrets/management-creds.json)
MANAGEMENT_SECRET_KEY=$(jq -r '.secretKey' /secrets/management-creds.json)

# === Create rate-limit service user ===
echo "[setup] Creating rate-limit service user..."
SERVICE_CREDS_JSON=$(AWS_ACCESS_KEY_ID="$MANAGEMENT_ACCESS_KEY" \
                      AWS_SECRET_ACCESS_KEY="$MANAGEMENT_SECRET_KEY" \
                      AWS_REGION="us-east-1" \
                      ./bin/ensureServiceUser apply service-rate-limit-user --iam-endpoint http://127.0.0.1:8600)

SERVICE_ACCESS_KEY=$(echo "$SERVICE_CREDS_JSON" | jq -r '.data.AccessKeyId')
SERVICE_SECRET_KEY=$(echo "$SERVICE_CREDS_JSON" | jq -r '.data.SecretAccessKey')

echo "[setup] rate-limit service user credentials:"
echo "SERVICE_ACCESS_KEY=$SERVICE_ACCESS_KEY"
echo "SERVICE_SECRET_KEY=$SERVICE_SECRET_KEY"
echo

# === Update rate-limit-service-creds.json ===
echo "[setup] Updating rate-limit-service-creds.json with service user credentials..."
jq --null-input --arg ak "$SERVICE_ACCESS_KEY" --arg sk "$SERVICE_SECRET_KEY" \
  '{accessKey: $ak, secretKey: $sk}' > /secrets/rate-limit-service-creds.json
