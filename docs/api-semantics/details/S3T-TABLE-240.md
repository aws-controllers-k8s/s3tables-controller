<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-240: KMS key Disabled: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 15 min
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-240"></a>**S3T-TABLE-240** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **KMS key Disabled: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 15 min**
  Phase 'disabled' (DescribeKey.KeyState 'Disabled'). Right after the key change (15 s): GetTable -> HTTP 200;
  GetTableMetadataLocation -> HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> HTTP 204;
  GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405
  'Table record expiration configuration API calls to this table are not allowed'; ListTables -> HTTP 200;
  CreateTable -> BadRequestException 400 'Insufficient access to perform table maintenance. Ensure that the
  key used to encrypt the table still exists, is active, and has a resource policy granting acce'; TagResource
  -> HTTP 200. Baseline (key fully granted) for comparison: GetTable -> HTTP 200; GetTableMetadataLocation ->
  HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> TooManyRequestsException
  429; GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus ->
  MethodNotAllowedException 405; ListTables -> HTTP 200; CreateTable -> HTTP 200; TagResource -> HTTP 200.
  Maintenance job status sampled every 60 s for 15 min: start
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions []. Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> DisableKey -> the read/write set -> GetTableMaintenanceJobStatus every 60
    s
  - measurements: soak_window_min=[15.0], n=17
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239), [S3T-TABLE-241](../table-record-expiration.md#s3t-table-241) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs

## Notes

H-U-126 qualified (key disabled variant). DisableKey changed nothing on the control plane compared with the
revoked-grant phase: the same 11 operations gave the same codes, CreateTable's 400 message is the same
'Insufficient access to perform table maintenance ... key ... still exists, is active ...' text. The three
Failed entries (lastRunTimestamp from the revoked phase) did not change and no job re-ran in 15 min (no new
commit). GetTableEncryption still reports aws:kms + the disabled key ARN.

Full original title: KMS key Disabled: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same
message); Failed job entries frozen for 15 min, no re-run
