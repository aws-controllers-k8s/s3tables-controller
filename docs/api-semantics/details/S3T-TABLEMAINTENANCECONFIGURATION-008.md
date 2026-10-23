<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEMAINTENANCECONFIGURATION-008: Disabled job status is stable: compaction + snapshot mgmt stayed 'Disabled' 60 min after Put(disabled); unconfigured UFR stayed Not_Yet_Run
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablemaintenanceconfiguration-008"></a>**S3T-TABLEMAINTENANCECONFIGURATION-008** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Disabled job status is stable: compaction + snapshot mgmt stayed 'Disabled' 60 min after Put(disabled); unconfigured UFR stayed Not_Yet_Run**
  After PutTableMaintenanceConfiguration disabled for icebergSnapshotManagement and icebergCompaction,
  GetTableMaintenanceJobStatus every 60 s for 60 min: final
  'icebergCompaction=Disabled;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run',
  transitions 0. icebergUnreferencedFileRemoval (bucket-level type, never configured on this table) read
  'Not_Yet_Run' throughout.
  - ACK: synced.when · ops: GetTableMaintenanceJobStatus, PutTableMaintenanceConfiguration · fields:
    status.*.status
  - repro: Put disabled (both types) -> GetTableMaintenanceJobStatus every 60 s for 60 min
  - measurements: soak_window_min=[60.0], n=61
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-002](../table-maintenance.md#s3t-tablemaintenanceconfiguration-002), [S3T-TABLEMAINTENANCECONFIGURATION-010](../table-maintenance.md#s3t-tablemaintenanceconfiguration-010),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-265](../table-maintenance.md#s3t-table-265), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-160](../table-maintenance.md#s3t-table-160) · hypotheses: H-U-057 · evidence:
    tablemaintenanceconfiguration/state-machine/job-status-soak

## Notes

H-U-057 second half (status flips only at the next scheduled pass): the Disabled value seen immediately after
the Put did not change in the window; whether a scheduled pass later rewrites it is beyond the 60-min cap.

Full original title: Disabled job status is stable: both types stayed 'Disabled' for 60 min after
Put(disabled) (61 samples); unconfigured icebergUnreferencedFileRemoval stayed Not_Yet_Run
