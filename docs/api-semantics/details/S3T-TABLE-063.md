<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-063: After DeleteTable (204) GetTable, GetTableEncryption, GetTableStorageClass -> NotFoundException at once; 0 stale hits in 33 polls over 30 s
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-063"></a>**S3T-TABLE-063** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable (204) GetTable, GetTableEncryption, GetTableStorageClass -> NotFoundException at once; 0 stale hits in 33 polls over 30 s**
  Table {IT + aws:kms K} (pre-delete reads {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"} / {"storageClass":
  "INTELLIGENT_TIERING"}). DeleteTable -> HTTP 204. Polled right after and then round-robin 1/s for 30 s:
  GetTable -> ['NotFoundException:The specified table does not exist.']; GetTableEncryption ->
  ['NotFoundException:The specified table does not exist.']; GetTableStorageClass -> ['NotFoundException:The
  specified table does not exist.']. CreateTable with the same name and {STANDARD + AES256} -> HTTP 200;
  GetTableEncryption -> {"sseAlgorithm": "AES256"}; GetTableStorageClass -> {"storageClass": "STANDARD"} (no
  cached old values).
  - ACK: exceptions.404 · ops: DeleteTable, GetTable, GetTableEncryption, GetTableStorageClass, CreateTable ·
    fields: encryptionConfiguration, storageClassConfiguration
  - repro: CreateTable({IT + aws:kms}) -> DeleteTable -> poll the three reads 30 s -> CreateTable same name
    {STANDARD + AES256} -> re-read
  - measurements: post_delete_polls=33, stale_hits=0, window_s=30, delete_table_latency_ms=[130], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-011](../table-encryption-storage.md#s3t-table-011), [S3T-TABLE-062](../table-encryption-storage.md#s3t-table-062), [S3T-TABLE-252](../table-encryption-storage.md#s3t-table-252), [S3T-TABLE-249](../table-encryption-storage.md#s3t-table-249) · hypotheses: H-T-179 · evidence:
    table/round-trip/explicit-encryption-storage

## Notes

H-T-179 confirmed: the three 404 messages are identical ('The specified table does not exist.').

Full original title: After DeleteTable (HTTP 204) GetTable, GetTableEncryption and GetTableStorageClass all
return NotFoundException immediately; 0 stale hits in 33 polls over 30 s; same-name re-create reads the new
config
