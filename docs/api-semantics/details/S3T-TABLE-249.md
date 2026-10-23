<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-249: DeleteTable on an SSE-KMS table -> HTTP 204 with the key Disabled and PendingDeletion; GetTableEncryption still returns the key ARN
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-249"></a>**S3T-TABLE-249** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **DeleteTable on an SSE-KMS table -> HTTP 204 with the key Disabled and PendingDeletion; GetTableEncryption still returns the key ARN**
  t1 created with aws:kms K1 while Enabled. With K1 Disabled: GetTableEncryption(t1) -> {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/6d730d5b-b3f2-41da-816f-c4549508aca6"};
  GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; DeleteTable -> HTTP 204; GetTable afterwards ->
  NotFoundException 404. t3 created after re-enabling; with K1 PendingDeletion: GetTableEncryption(t3) ->
  {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6d730d5b-b3f2-41da-816f-c4549508aca6"}; GetTable -> HTTP 200;
  DeleteTable -> HTTP 204; GetTable afterwards -> NotFoundException 404.
  - ACK: none · ops: DeleteTable, GetTableEncryption, GetTable, GetTableMetadataLocation · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms K) -> DisableKey(K) -> GetTableEncryption, GetTable, DeleteTable; second table
    -> ScheduleKeyDeletion(K) -> GetTableEncryption, DeleteTable
  - measurements: delete_table_latency_ms=[166, 147], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLE-063](../table-encryption-storage.md#s3t-table-063) · hypotheses: H-T-184 · evidence:
    table/error-taxonomy/kms-key-states-types

## Notes

H-T-184 confirmed: DeleteTable -> HTTP 204 with the key Disabled (t1) and with the key PendingDeletion (t3);
GetTable, GetTableMetadataLocation and GetTableEncryption all still succeed and GetTableEncryption returns the
dead key's ARN. No finalizer-stuck scenario from a dead key at the control-plane level. n=1 per state.

Full original title: DeleteTable on an SSE-KMS table with its key Disabled -> HTTP 204, PendingDeletion ->
HTTP 204; GetTableEncryption keeps returning the key ARN
