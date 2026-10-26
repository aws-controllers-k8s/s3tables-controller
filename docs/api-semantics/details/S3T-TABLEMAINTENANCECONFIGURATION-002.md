<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-002: PutTableMaintenanceConfiguration(disabled) -> job status 'Disabled' on the first read (<1 s) for snapshot mgmt and compaction...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-002"></a>**S3T-TABLEMAINTENANCECONFIGURATION-002** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration(disabled) -> job status 'Disabled' on the first read (<1 s) for snapshot mgmt and compaction...** (hypothesis refuted; behavior confirmed)
  t_dis (fresh table with schema, all jobs Not_Yet_Run).
  PutTableMaintenanceConfiguration(type=icebergSnapshotManagement, value={status: disabled, settings
  {minSnapshotsToKeep 1, maxSnapshotAgeHours 120}}) -> HTTP 204; GetTableMaintenanceJobStatus 95 ms later ->
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run;
  1/s sampling: [[0.15,
  "icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run"]].
  GetTableMaintenanceConfiguration afterwards -> HTTP 200 {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}},
  "icebergSnapshotManagement": {"status": "disabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}. Then
  PutTableMaintenanceConfiguration(type=icebergCompaction, {status: disabled, settings {512, auto}}) -> HTTP
  204; first read ->
  icebergCompaction=Disabled;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run;
  1/s sampling: [[0.15,
  "icebergCompaction=Disabled;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run"]].
  Raw member presence in the Disabled state: {"icebergCompaction": ["status"], "icebergSnapshotManagement":
  ["status"], "icebergUnreferencedFileRemoval": ["status"]} (lastRunTimestamp absent, failureMessage absent).
  - ACK: synced.when, hook:sdk_read_one_post_set_output · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus, GetTableMaintenanceConfiguration · fields: type, value.status,
    status.icebergSnapshotManagement.status, status.icebergCompaction.status
  - repro: CreateTable(schema) -> Put icebergSnapshotManagement disabled -> GetTableMaintenanceJobStatus 1/s x
    30 s -> Put icebergCompaction disabled -> same
  - measurements: disabled_visible_after_put_s=[0.0, 0.0], burst_polls=[1, 1], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008), [S3T-TABLEMAINTENANCECONFIGURATION-010](../table-maintenance.md#s3t-tablemaintenanceconfiguration-010),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265) · hypotheses: H-U-057, H-U-062 · evidence:
    tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-057 (Disabled appears only after the next scheduled pass, hours later) REFUTED: the job status reflects
the configuration write immediately. H-U-062 Disabled cell on a never-run job: no lastRunTimestamp (there was
no previous run) and no failureMessage.

Full original title: PutTableMaintenanceConfiguration(disabled) -> job status 'Disabled' on the first read (<1
s) for snapshot mgmt and compaction; entry = {status} only
