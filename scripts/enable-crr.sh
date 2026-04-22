#!/usr/bin/env bash
# enable-crr.sh — create source + destination buckets and configure replication.
#
# Usage:
#   scripts/enable-crr.sh --source <bucket> --destination <bucket> \
#       [--prefix <pfx>] [--endpoint <url>]
#
# Defaults:
#   --endpoint  http://127.0.0.1:8000
#   --prefix    ""   (replicate everything)
#
# Idempotent: re-running is a no-op once resources exist.

set -eu

# Pinned to match templates/vault/create-management-account.sh and the
# replication-role accountSeed in templates/vault/config.json.
ROLE_ARN="arn:aws:iam::123456789012:role/scality-internal/replication-role"

SOURCE=""
DESTINATION=""
PREFIX=""
ENDPOINT="http://127.0.0.1:8000"

while [ $# -gt 0 ]; do
    case "$1" in
        --source)       SOURCE="$2"; shift 2 ;;
        --destination)  DESTINATION="$2"; shift 2 ;;
        --prefix)       PREFIX="$2"; shift 2 ;;
        --endpoint)     ENDPOINT="$2"; shift 2 ;;
        -h|--help)
            sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *) echo "unknown flag: $1" >&2; exit 2 ;;
    esac
done

if [ -z "$SOURCE" ] || [ -z "$DESTINATION" ]; then
    echo "error: --source and --destination are required" >&2
    exit 2
fi

# testaccount credentials are fixed in templates/vault/create-management-account.sh
export AWS_ACCESS_KEY_ID="WBTKACCESSI9O3YKIRQ0"
export AWS_SECRET_ACCESS_KEY="ICxmNTBbOqijy4rMq/MOP1EPlTMqfsEBLjROcAbN"
export AWS_DEFAULT_REGION="us-east-1"

AWS="aws --endpoint-url $ENDPOINT"

create_bucket() {
    local bucket="$1"
    if $AWS s3api create-bucket --bucket "$bucket" >/dev/null 2>&1; then
        echo "[crr] created bucket $bucket"
    else
        # swallow "already exists and owned by you" — treat anything else as fatal
        if $AWS s3api head-bucket --bucket "$bucket" >/dev/null 2>&1; then
            echo "[crr] bucket $bucket already exists"
        else
            echo "error: failed to create bucket $bucket" >&2
            $AWS s3api create-bucket --bucket "$bucket"
            exit 1
        fi
    fi
}

enable_versioning() {
    local bucket="$1"
    $AWS s3api put-bucket-versioning \
        --bucket "$bucket" \
        --versioning-configuration Status=Enabled
    echo "[crr] versioning enabled on $bucket"
}

create_bucket "$SOURCE"
create_bucket "$DESTINATION"
enable_versioning "$SOURCE"
enable_versioning "$DESTINATION"

REPLICATION_CONFIG=$(cat <<EOF
{
  "Role": "${ROLE_ARN},${ROLE_ARN}",
  "Rules": [
    {
      "ID": "workbench-crr",
      "Status": "Enabled",
      "Prefix": "${PREFIX}",
      "Destination": {
        "Bucket": "arn:aws:s3:::${DESTINATION}",
        "StorageClass": "sf"
      }
    }
  ]
}
EOF
)

$AWS s3api put-bucket-replication \
    --bucket "$SOURCE" \
    --replication-configuration "$REPLICATION_CONFIG"

echo "[crr] replication configured: $SOURCE -> $DESTINATION (prefix='$PREFIX')"
