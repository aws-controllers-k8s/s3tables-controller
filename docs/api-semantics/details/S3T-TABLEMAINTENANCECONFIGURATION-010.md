<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-010: Re-enabling snapshot management after 60 min Disabled (never run) -> 'Not_Yet_Run' on the first read; compaction left Disabled
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-010"></a>**S3T-TABLEMAINTENANCECONFIGURATION-010** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Re-enabling snapshot management after 60 min Disabled (never run) -> 'Not_Yet_Run' on the first read; compaction left Disabled**
  t_dis had both types disabled for the whole soak (status Disabled, no run ever).
  PutTableMaintenanceConfiguration(icebergSnapshotManagement, {enabled, 1/120}) -> HTTP 204.
  GetTableMaintenanceJobStatus 342 ms later ->
  'icebergCompaction=Disabled;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run';
  1/s sampling for up to 30 s: [[0.21,
  "icebergCompaction=Disabled;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run"]];
  final
  'icebergCompaction=Disabled;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run'.
  Raw presence after re-enable: {"top": ["status", "tableARN"], "status_map": {"icebergCompaction":
  ["status"], "icebergSnapshotManagement": ["status"], "icebergUnreferencedFileRemoval": ["status"]}}.
  - ACK: synced.when, hook:sdk_read_one_post_set_output · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus · fields: value.status, status.icebergSnapshotManagement.status
  - repro: Put disabled -> wait -> Put enabled -> GetTableMaintenanceJobStatus 1/s x 30 s
  - measurements: disabled_for_min=[60.0], burst_polls=[1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-002](../table-maintenance.md#s3t-tablemaintenanceconfiguration-002), [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265) · hypotheses: H-U-057, H-U-062 · evidence:
    tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

Re-enable returns the never-run job to 'Not_Yet_Run' (the configuration write, not a run, drives the job
status value). n=1.
