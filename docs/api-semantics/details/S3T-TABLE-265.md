<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-265: Bucket-level Put(disabled) on icebergUnreferencedFileRemoval -> only that job reads 'Disabled' in each table's job status; newborns too
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-265"></a>**S3T-TABLE-265** `requested-vs-effective` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Bucket-level Put(disabled) on icebergUnreferencedFileRemoval -> only that job reads 'Disabled' in each table's job status; newborns too**
  Two fresh tables read
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run /
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run.
  PutTableBucketMaintenanceConfiguration(type=icebergUnreferencedFileRemoval, value={status: disabled}) ->
  HTTP 204. GetTableMaintenanceJobStatus(t1) 0.55s later: UnreferencedFileRemoval=Disabled; watch (1/s) ->
  {"value": "Disabled", "at_s": 0.11, "polls": 1}. t2 after:
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Disabled. A table created
  while the bucket setting was disabled:
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Disabled at birth.
  GetTable(t1) keys changed by the bucket-level write: []; GetTableMaintenanceConfiguration(t1) unchanged:
  True (it lists only compaction+snapshot). Table-level
  PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) -> BadRequestException (HTTP 400)
  'Invalid table maintenance type: icebergUnreferencedFileRemoval'. Re-enable at bucket level -> 204; t1 job
  status -> {"value": "Not_Yet_Run", "at_s": 0.49, "polls": 1}; t3 ->
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run; t2 ->
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run.
  - ACK: scope:field-on-parent, hook:sdk_read_one_post_set_output · ops:
    PutTableBucketMaintenanceConfiguration, GetTableMaintenanceJobStatus, PutTableMaintenanceConfiguration,
    GetTable · fields: status.icebergUnreferencedFileRemoval.status, value.status
  - repro: CreateTableBucket -> CreateNamespace -> 2x CreateTable ->
    PutTableBucketMaintenanceConfiguration(disabled) -> GetTableMaintenanceJobStatus per table (1/s) ->
    CreateTable -> re-enable
  - measurements: t1_first_read_after_put_s=0.545, t1_change_at_s=0.11, t1_reenable_at_s=0.49, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLEMAINTENANCECONFIGURATION-002](../table-maintenance.md#s3t-tablemaintenanceconfiguration-002),
    [S3T-TABLEMAINTENANCECONFIGURATION-008](../table-maintenance.md#s3t-tablemaintenanceconfiguration-008), [S3T-TABLEMAINTENANCECONFIGURATION-010](../table-maintenance.md#s3t-tablemaintenanceconfiguration-010),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-319](../table-encryption-storage.md#s3t-table-319)
    · evidence: table/creative/bucket-maintenance-propagation

## Notes

Cross-service analogy (ElastiCache parameter-group apply status / DynamoDB ContinuousBackupsStatus): a
per-table status that is controlled by a parent-bucket setting. A Table controller surfacing job status sees
changes it did not cause; the table-level Put cannot override the bucket-level job. Single bucket, 3 tables,
n=1 per cell. Extends [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001) (which only listed the bucket-level job type on a
fresh table) and [S3T-TABLE-160](../table-maintenance.md#s3t-table-160) (table-level Put of that type -> 400).

Full original title: Bucket-level maintenance Put(disabled) flips every table's
GetTableMaintenanceJobStatus.icebergUnreferencedFileRemoval to 'Disabled' (first read 0.5s); new table born
'Disabled'
