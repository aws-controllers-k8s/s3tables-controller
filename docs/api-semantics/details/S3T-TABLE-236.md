<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-236: Key grant revoked: UnreferencedFileRemoval job Not_Yet_Run -> Failed 5.0 min after a metadata-only commit; CreateTable maintenance text, n=1
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-236"></a>**S3T-TABLE-236** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key grant revoked: UnreferencedFileRemoval job Not_Yet_Run -> Failed 5.0 min after a metadata-only commit; CreateTable maintenance text, n=1**
  Table in an SSE-KMS bucket; key state 'Enabled' (phase revoked). GetTableMaintenanceJobStatus every 60 s:
  icebergUnreferencedFileRemoval changed Not_Yet_Run -> Failed+ts+msg at +302 s into the phase (5.0 min after
  the metadata-only commit made under the revoked key). failureMessage: {"icebergUnreferencedFileRemoval":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'."}. Raw body: {"status": {"icebergCompaction": {"status":
  "Not_Yet_Run"}, "icebergUnreferencedFileRemoval": {"failureMessage": "Insufficient access to perform table
  maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.",
  "lastRunTimestamp": "2026-10-13T23:51:56.672Z", "status": "Failed"}, "icebergSnapshotManagement": {"status":
  "Not_Yet_Run"}}, "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-abb4b0-kmsjob/table/7752f10a-bdd2-4877-a789-a56ea35d13ce"}.
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus,
    UpdateTableMetadataLocation, PutKeyPolicy · fields: status.*.status, status.*.failureMessage,
    status.*.lastRunTimestamp, encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS table -> PutKeyPolicy(root only) -> UpdateTableMetadataLocation(copy of metadata.json) ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: since_commit_min=[5.0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-235](../table-maintenance.md#s3t-table-235), [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238), [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011),
    [S3T-TABLE-242](../table-encryption-storage.md#s3t-table-242), [S3T-TABLEMAINTENANCECONFIGURATION-003](../table-maintenance.md#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](../table-maintenance.md#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](../table-maintenance.md#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-009](../table-maintenance.md#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-126, H-U-062, H-U-060 · evidence:
    table/dependencies/kms-key-revoked-jobs

## Notes

Transition recorded immediately during the 'revoked' soak (key Enabled, policy = account root only). Failed
entry members: status, lastRunTimestamp, failureMessage. The wire value is 'Failed' (model casing), not the
user guide's 'FAILED'. The message is identical to CreateTable's synchronous rejection text ([S3T-TABLE-054](../table-maintenance.md#s3t-table-054)).
H-U-126 (Failed within 48 h with a KMS/AccessDenied-flavoured message) qualified: reached within 5 min via the
maintenance job status (record expiration is 405 on customer tables); the message says 'Insufficient access
... key ... maintenance.s3tables.amazonaws.com', not 'KMS'/'AccessDenied' literally. H-U-060: Failed is
reachable and spelled 'Failed'. H-U-062 Failed cell: all three members present.

Full original title: Key grant revoked: UnreferencedFileRemoval job Not_Yet_Run -> 'Failed' 5.0 min after a
metadata-only commit; failureMessage = the CreateTable 'Insufficient access to perform table maintenance' text
