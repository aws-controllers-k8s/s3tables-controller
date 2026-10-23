<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-239: Key grant revoked after creation: only CreateTable -> 400; existing-table reads/Put config/Tag 200/204; 3 maintenance jobs Failed in 8 min
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-239"></a>**S3T-TABLE-239** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Key grant revoked after creation: only CreateTable -> 400; existing-table reads/Put config/Tag 200/204; 3 maintenance jobs Failed in 8 min**
  Phase 'revoked' (DescribeKey.KeyState 'Enabled'). Right after the key change (15 s): GetTable -> HTTP 200;
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
  Maintenance job status sampled every 60 s for 20 min: start
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions [[302,
  "icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Failed+ts+msg"],
  [362,
  "icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Failed+ts+msg"],
  [483,
  "icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg"]].
  Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> PutKeyPolicy(root only) -> the read/write set ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: soak_window_min=[20.0], n=21
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-240](../table-record-expiration.md#s3t-table-240), [S3T-TABLE-241](../table-record-expiration.md#s3t-table-241), [S3T-TABLE-253](../table-maintenance.md#s3t-table-253), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254), [S3T-TABLE-246](../table-maintenance.md#s3t-table-246),
    [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054) · hypotheses: H-U-126 · evidence: table/dependencies/kms-key-revoked-jobs

## Notes

H-U-126 qualified: a table cannot be created on a key without the maintenance grant (CreateTable rejects,
[S3T-TABLE-054](../table-maintenance.md#s3t-table-054)), so the state was reached by revoking the grant AFTER creation. Existing-table operations
(GetTable, GetTableMetadataLocation, GetTableEncryption, GetTableBucketEncryption,
Get/PutTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, ListTables, TagResource) are NOT affected;
only CreateTable in the bucket fails. Job status: all three types went Not_Yet_Run -> Failed between +5.0 and
+8.1 min after the one metadata-only commit (transition findings); no re-run (lastRunTimestamp unchanged) in
the remaining 12 min. The baseline PutTableMaintenanceConfiguration 429 was a shared-account throttle, not a
key effect.

Full original title: Key grant revoked after creation: CreateTable in the bucket -> BadRequestException 400;
every other read/write on the existing table -> 200/204; its 3 jobs Failed within 8 min of a commit
