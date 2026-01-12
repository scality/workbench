CREATE TABLE IF NOT EXISTS logs.access_logs_federated AS logs.access_logs
ENGINE = Distributed(workbench_cluster, logs, access_logs, raftSessionID);

CREATE TABLE IF NOT EXISTS logs.offsets_federated AS logs.offsets
ENGINE = Distributed(workbench_cluster, logs, offsets, raftSessionID);
