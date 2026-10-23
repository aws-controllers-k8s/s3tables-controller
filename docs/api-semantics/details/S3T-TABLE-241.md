<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-241: KMS key PendingDeletion: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 10 min
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-241"></a>**S3T-TABLE-241** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **KMS key PendingDeletion: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 10 min**
  Phase 'pending-deletion' (DescribeKey.KeyState 'PendingDeletion'). Right after the key change (15 s):
  GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; GetTableEncryption -> HTTP 200;
  GetTableBucketEncryption -> HTTP 200; GetTableMaintenanceConfiguration -> HTTP 200;
  PutTableMaintenanceConfiguration -> HTTP 204; GetTableMaintenanceJobStatus -> HTTP 200;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405 'Table record expiration configuration
  API calls to this table are not allowed'; ListTables -> HTTP 200; CreateTable -> BadRequestException 400
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting acce'; TagResource -> HTTP 200. Baseline (key fully
  granted) for comparison: GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; GetTableEncryption ->
  HTTP 200; GetTableBucketEncryption -> HTTP 200; GetTableMaintenanceConfiguration -> HTTP 200;
  PutTableMaintenanceConfiguration -> TooManyRequestsException 429; GetTableMaintenanceJobStatus -> HTTP 200;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405; ListTables -> HTTP 200; CreateTable ->
  HTTP 200; TagResource -> HTTP 200. Maintenance job status sampled every 60 s for 10 min: start
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions []. Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> ScheduleKeyDeletion(7) -> the read/write set ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: soak_window_min=[10.0], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239), [S3T-TABLE-240](../table-record-expiration.md#s3t-table-240) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs

## Notes

H-U-126 qualified (key pending deletion variant). ScheduleKeyDeletion(7 days) changed nothing on the control
plane versus Disabled: same codes for the 11 operations, CreateTable 400 with the same message. The Failed
entries did not change and no job re-ran in 10 min. Total time the Failed entries stayed unchanged across the
disabled + pending phases: ~25 min.

Full original title: KMS key PendingDeletion: control plane unchanged (reads/Put config/Tag 200, CreateTable
400 same message); Failed job entries frozen for 10 min
