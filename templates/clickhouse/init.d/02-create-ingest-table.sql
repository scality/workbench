CREATE TABLE IF NOT EXISTS logs.access_logs_ingest
(
    -- Common
    insertedAt             DateTime DEFAULT now(),
    hostname               LowCardinality(Nullable(String)),

    -- AWS access server logs fields https://docs.aws.amazon.com/AmazonS3/latest/userguide/LogFormat.html
    startTime              Int64, -- AWS "Time" field (epoch milliseconds)
    requester              Nullable(String),
    operation              Nullable(String),
    requestURI             Nullable(String),
    errorCode              Nullable(String),
    objectSize             Nullable(UInt64),
    totalTime              Nullable(Float32),
    turnAroundTime         Nullable(Float32),
    referer                Nullable(String),
    userAgent              Nullable(String),
    versionId              Nullable(String),
    signatureVersion       LowCardinality(Nullable(String)),
    cipherSuite            LowCardinality(Nullable(String)),
    authenticationType     LowCardinality(Nullable(String)),
    hostHeader             Nullable(String),
    tlsVersion             LowCardinality(Nullable(String)),
    aclRequired            LowCardinality(Nullable(String)),

    -- Shared between AWS access server logs and Analytics logs
    bucketOwner            Nullable(String), -- AWS "Bucket Owner" field
    bucketName             String DEFAULT '', -- AWS "Bucket" field
    req_id                 String DEFAULT '', -- AWS "Request ID" field
    bytesSent              Nullable(UInt64), -- AWS "Bytes Sent" field
    clientIP               Nullable(String), -- AWS "Remote IP" field
    httpCode               Nullable(UInt16), -- AWS "HTTP Status" field
    objectKey              Nullable(String), -- AWS "Key" field

    -- Scality server access logs extra fields.
    logFormatVersion       LowCardinality(Nullable(String)),
    loggingEnabled         Bool DEFAULT false,
    loggingTargetBucket    String DEFAULT '',
    loggingTargetPrefix    String DEFAULT '',
    awsAccessKeyID         Nullable(String),
    raftSessionID          UInt16 DEFAULT 0
)
Engine = Null();
