<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-007: Metadata-only UpdateTableMetadataLocation triggers all 3 maintenance jobs within 7 min (Successful +2.8/+4.8/+6.8 min)...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-007"></a>**S3T-TABLEMAINTENANCECONFIGURATION-007** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Metadata-only UpdateTableMetadataLocation triggers all 3 maintenance jobs within 7 min (Successful +2.8/+4.8/+6.8 min)...** (hypothesis refuted; behavior confirmed)
  t_prop: CreateTable(schema) wrote metadata.json with keys ['current-schema-id', 'current-snapshot-id',
  'default-sort-order-id', 'default-spec-id', 'format-version', 'last-column-id', 'last-partition-id',
  'last-sequence-number', 'last-updated-ms', 'location', 'metadata-log', 'partition-specs',
  'partition-statistics', 'properties', 'refs', 'schemas', 'snapshot-log', 'snapshots', 'sort-orders',
  'statistics', 'table-uuid'] (snapshots: 0, properties: {"write.parquet.compression-codec": "zstd"}, refs:
  {}). A copy with properties['history.expire.max-snapshot-age-ms']='1' (+ metadata-log entry, still no
  snapshot) was written to <warehouse>/metadata/00001-<uuid>.metadata.json and
  UpdateTableMetadataLocation(versionToken) -> HTTP 200 (GetTable then shows the new metadataLocation: True,
  versionToken changed: True). Job status right after the commit:
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run'.
  Sampled every 60 s: icebergCompaction -> Successful+ts first seen 2.8 min after the commit (lastRunTimestamp
  2026-10-13T23:29:52.066Z); icebergSnapshotManagement -> Successful+ts first seen 4.8 min after the commit
  (lastRunTimestamp 2026-10-13T23:31:07.881Z); icebergUnreferencedFileRemoval -> Successful+ts first seen 6.8
  min after the commit (lastRunTimestamp 2026-10-13T23:33:02.148Z). Final raw body: {"status":
  {"icebergCompaction": {"lastRunTimestamp": "2026-10-13T23:29:52.066Z", "status": "Successful"},
  "icebergUnreferencedFileRemoval": {"lastRunTimestamp": "2026-10-13T23:33:02.148Z", "status": "Successful"},
  "icebergSnapshotManagement": {"lastRunTimestamp": "2026-10-13T23:31:07.881Z", "status": "Successful"}},
  "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-cf508b-mjob/table/e707e087-85b8-4393-aa1e-fd9550ae04f2"}.
  Member presence per job type in the final state: {"icebergCompaction": ["lastRunTimestamp", "status"],
  "icebergUnreferencedFileRemoval": ["lastRunTimestamp", "status"], "icebergSnapshotManagement":
  ["lastRunTimestamp", "status"]}. Soak window 60 min, 61 samples.
  - ACK: synced.when, e2e-timing, requeue · ops: UpdateTableMetadataLocation, GetTableMaintenanceJobStatus,
    GetTable · fields: metadataLocation, versionToken, status.*.status, status.*.lastRunTimestamp,
    status.*.failureMessage
  - repro: CreateTable(schema) -> GetObject(metadata.json) -> PutObject(copy +
    history.expire.max-snapshot-age-ms=1) -> UpdateTableMetadataLocation -> GetTableMaintenanceJobStatus every
    60 s
  - measurements: commit_to_first_successful_min=[2.8, 4.8, 6.8], soak_window_min=[60.0], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](../table-maintenance.md#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](../table-maintenance.md#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-006](../table-maintenance.md#s3t-tablemaintenanceconfiguration-006) · hypotheses: H-U-060,
    H-U-059, H-U-062 · evidence: tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-060 REFUTED on the property variant: snapshot management did not report Failed - it reported
'Successful+ts' with a lastRunTimestamp 4.8 min after the commit (the hypothesis expected Failed within 48 h);
the branch/tag-ref variant was not attempted (a ref needs an existing snapshot). H-U-059 qualified: jobs DO
run on a table without data once a metadata commit happens (minutes, not 24-48 h), while untouched tables
stayed Not_Yet_Run. H-U-062 Successful cell: members ["lastRunTimestamp", "status"]. Wire casing is the
model's (Successful, Not_Yet_Run), not the user guide's upper case. n=1 table.

Full original title: Metadata-only UpdateTableMetadataLocation triggers all 3 maintenance jobs within 7 min
(Successful +2.8/+4.8/+6.8 min); history.expire.max-snapshot-age-ms=1 -> Successful, not Failed
