CREATE TABLE IF NOT EXISTS logs.offsets
(
    bucketName                String,
    raftSessionID             UInt16,
    lastProcessedInsertedAt   DateTime,
    lastProcessedStartTime    Int64,
    lastProcessedReqId        String
)
ENGINE = ReplacingMergeTree(lastProcessedInsertedAt)
ORDER BY (bucketName, raftSessionID);
