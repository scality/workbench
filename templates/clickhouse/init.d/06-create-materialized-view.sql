CREATE MATERIALIZED VIEW IF NOT EXISTS logs.access_logs_ingest_mv
TO logs.access_logs_federated
AS
SELECT
    insertedAt,
    hostname,

    startTime,
    requester,
    operation,
    requestURI,
    errorCode,
    objectSize,
    totalTime,
    turnAroundTime,
    referer,
    userAgent,
    versionId,
    signatureVersion,
    cipherSuite,
    authenticationType,
    hostHeader,
    tlsVersion,
    aclRequired,

    bucketOwner,
    bucketName,
    req_id,
    bytesSent,
    clientIP,
    httpCode,
    objectKey,

    logFormatVersion,
    loggingEnabled,
    loggingTargetBucket,
    loggingTargetPrefix,
    awsAccessKeyID,
    raftSessionID
FROM logs.access_logs_ingest
WHERE loggingEnabled = true;
