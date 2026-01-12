CREATE TABLE IF NOT EXISTS logs.offsets
(
    bucketName                String,
    raftSessionID             UInt16,
    lastProcessedInsertedAt   DateTime,
    lastProcessedTimestamp    DateTime64(3),
    lastProcessedReqId        String
)
ENGINE = ReplacingMergeTree(lastProcessedInsertedAt)
ORDER BY (bucketName, raftSessionID);
