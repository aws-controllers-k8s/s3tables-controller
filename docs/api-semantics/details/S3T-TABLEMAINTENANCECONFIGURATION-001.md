<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-001: Fresh table: GetTableMaintenanceJobStatus -> 200 listing all 3 job types (incl. bucket-level UFR) Not_Yet_Run, entry = {status} only (n=4)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-001"></a>**S3T-TABLEMAINTENANCECONFIGURATION-001** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table: GetTableMaintenanceJobStatus -> 200 listing all 3 job types (incl. bucket-level UFR) Not_Yet_Run, entry = {status} only (n=4)**
  GetTableMaintenanceJobStatus 1-1 s after CreateTable on four fresh tables (with schema, without metadata,
  both kinds): t_schema -> HTTP 200
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run;
  t_nometa -> HTTP 200
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run;
  t_dis -> HTTP 200
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run;
  t_prop -> HTTP 200
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run.
  Raw JSON body (t_schema): {"status": {"icebergCompaction": {"status": "Not_Yet_Run"},
  "icebergUnreferencedFileRemoval": {"status": "Not_Yet_Run"}, "icebergSnapshotManagement": {"status":
  "Not_Yet_Run"}}, "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-cf508b-mjob/table/a6c592b2-62d0-4755-b899-1e41583183ad"}.
  Member presence per cell (top-level / per job type): {"t_dis": {"status_map": {"icebergCompaction":
  ["status"], "icebergSnapshotManagement": ["status"], "icebergUnreferencedFileRemoval": ["status"]}, "top":
  ["status", "tableARN"]}, "t_nometa": {"status_map": {"icebergCompaction": ["status"],
  "icebergSnapshotManagement": ["status"], "icebergUnreferencedFileRemoval": ["status"]}, "top": ["status",
  "tableARN"]}, "t_prop": {"status_map": {"icebergCompaction": ["status"], "icebergSnapshotManagement":
  ["status"], "icebergUnreferencedFileRemoval": ["status"]}, "top": ["status", "tableARN"]}, "t_schema":
  {"status_map": {"icebergCompaction": ["status"], "icebergSnapshotManagement": ["status"],
  "icebergUnreferencedFileRemoval": ["status"]}, "top": ["status", "tableARN"]}}. Job types present:
  ['icebergCompaction', 'icebergSnapshotManagement', 'icebergUnreferencedFileRemoval']
  (icebergUnreferencedFileRemoval IS listed even though it is configured at bucket level only).
  GetTableMaintenanceConfiguration(t_schema) -> HTTP 200 {"tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-cf508b-mjob/table/a6c592b2-62d0-4755-b899-1e41583183ad",
  "configuration": {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction":
  {"targetFileSizeMB": 512, "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled",
  "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 1, "maxSnapshotAg.
  - ACK: late_initialize, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus, CreateTable,
    GetTableMaintenanceConfiguration · fields: status, status.*.status, status.*.lastRunTimestamp,
    status.*.failureMessage
  - repro: CreateTable (with and without metadata.iceberg.schema) -> GetTableMaintenanceJobStatus right away
  - measurements: age_at_sample_s=[1.4, 1.1, 0.9, 0.7], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-219](../table-maintenance.md#s3t-table-219), [S3T-TABLE-220](../table-maintenance.md#s3t-table-220), [S3T-TABLEMAINTENANCECONFIGURATION-009](../table-maintenance.md#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236),
    [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008),
    [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-206](../table-maintenance.md#s3t-table-206) ·
    hypotheses: H-U-062, H-U-059 · evidence: tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-062 Not_Yet_Run cell: confirmed - neither lastRunTimestamp nor failureMessage is present (absent, not
null). H-U-059 baseline: every job type starts at the status shown; wire casing is exactly as printed (model
enum Not_Yet_Run/Successful/Failed/Disabled). Successful/Failed cells are not reachable within this probe's
60-min cap.

Full original title: Fresh table: GetTableMaintenanceJobStatus -> 200 with all 3 job types (incl. bucket-level
icebergUnreferencedFileRemoval) 'Not_Yet_Run', entry = {status} only
