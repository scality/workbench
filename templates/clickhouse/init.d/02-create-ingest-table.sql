CREATE TABLE IF NOT EXISTS logs.access_logs_ingest
(
    -- Common
    timestamp              DateTime,
    insertedAt             DateTime DEFAULT now(),
    hostname               LowCardinality(String),

    -- Analytics
    action                 LowCardinality(String),
    accountName            String,
    accountDisplayName     String,
    userName               String,
    clientPort             UInt32,
    httpMethod             LowCardinality(String),
    bytesDeleted           UInt64,
    bytesReceived          UInt64,
    bodyLength             UInt64,
    contentLength          UInt64,
    elapsed_ms             Float32,

    -- AWS access server logs fields https://docs.aws.amazon.com/AmazonS3/latest/userguide/LogFormat.html
    startTime              DateTime64(3), -- AWS "Time" field
    requester              String,
    operation              String,
    requestURI             String,
    errorCode              String,
    objectSize             UInt64,
    totalTime              Float32,
    turnAroundTime         Float32,
    referer                String,
    userAgent              String,
    versionId              String,
    signatureVersion       LowCardinality(String),
    cipherSuite            LowCardinality(String),
    authenticationType     LowCardinality(String),
    hostHeader             String,
    tlsVersion             LowCardinality(String),
    aclRequired            LowCardinality(String),

    -- Shared between AWS access server logs and Analytics logs
    bucketOwner            String, -- AWS "Bucket Owner" field
    bucketName             String, -- AWS "Bucket" field
    req_id                 String, -- AWS "Request ID" field
    bytesSent              UInt64, -- AWS "Bytes Sent" field
    clientIP               String, -- AWS "Remote IP" field
    httpCode               UInt16, -- AWS "HTTP Status" field
    objectKey              String, -- AWS "Key" field

    -- Scality server access logs extra fields.
    logFormatVersion       LowCardinality(String),
    loggingEnabled         Bool,
    loggingTargetBucket    String,
    loggingTargetPrefix    String,
    awsAccessKeyID         String,
    raftSessionID          UInt16
)
Engine = Null();
