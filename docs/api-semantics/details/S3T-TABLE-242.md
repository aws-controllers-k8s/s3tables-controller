<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-242: KMS key PendingDeletion (after Disabled): DeleteTable, DeleteNamespace, DeleteTableBucket -> 204; deletion does not need a usable key (n=1)
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-242"></a>**S3T-TABLE-242** `delete-semantics` · impact high · handled · verified 2026-10-14
  **KMS key PendingDeletion (after Disabled): DeleteTable, DeleteNamespace, DeleteTableBucket -> 204; deletion does not need a usable key (n=1)**
  With the bucket's KMS key in state 'PendingDeletion' (ScheduleKeyDeletion 7 days, after the maintenance
  grant had been revoked and the key disabled): DeleteTable -> HTTP 204; GetTable afterwards ->
  NotFoundException 404; DeleteNamespace -> HTTP 204; DeleteTableBucket -> HTTP 204. Final key state
  'PendingDeletion'.
  - ACK: deletable.when, pre-delete-cleanup, terminal_codes · ops: DeleteTable, DeleteNamespace,
    DeleteTableBucket, ScheduleKeyDeletion · fields: encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS bucket + table -> DisableKey -> ScheduleKeyDeletion(7) -> DeleteTable -> DeleteNamespace ->
    DeleteTableBucket
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-235](../table-maintenance.md#s3t-table-235), [S3T-TABLE-236](../table-maintenance.md#s3t-table-236), [S3T-TABLE-237](../table-maintenance.md#s3t-table-237), [S3T-TABLE-238](../table-maintenance.md#s3t-table-238),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](../table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs

## Notes

Deletion-window behavior for a controller's finalizer path: a table/bucket whose KMS key is disabled or
pending deletion can be deleted without restoring the key; GetTable afterwards -> 404. The key itself stays
PendingDeletion for 7 days (lab leftover, registered). n=1 table + 1 extra table created in the baseline phase
(also 204).

Full original title: KMS key PendingDeletion (and previously Disabled): DeleteTable -> 204, DeleteNamespace ->
204, DeleteTableBucket -> 204 - deletion does not need a usable key
