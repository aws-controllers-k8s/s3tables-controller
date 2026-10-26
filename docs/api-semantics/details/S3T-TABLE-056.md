<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-056: Maintenance-principal key-policy Condition at CreateTable: <bucket>/table/<name>/* -> 400; <bucket>/table/*, <bucket>/*, no Condition -> 200
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-056"></a>**S3T-TABLE-056** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance-principal key-policy Condition at CreateTable: <bucket>/table/<name>/* -> 400; <bucket>/table/*, <bucket>/*, no Condition -> 200**
  Key policy = root statement + Allow maintenance.s3tables.amazonaws.com kms:GenerateDataKey/kms:Decrypt with
  Condition StringLike kms:EncryptionContext:aws:s3:arn = X, 10 s after each PutKeyPolicy;
  CreateTable(aws:kms, key ARN) with a fresh name each time. X=<bucketARN>/table/<planned table name>/* ->
  BadRequestException (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used
  to encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.' after 3 attempts (41.4 s after PutKeyPolicy);
  X=<bucketARN>/table/* -> HTTP 200 (first attempt, 10.6 s after PutKeyPolicy); GetTableEncryption
  {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}; X=<bucketARN>/* -> HTTP 200
  (first attempt, 10.5 s after PutKeyPolicy); GetTableEncryption {"encryptionConfiguration": {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}; no
  Condition -> HTTP 200 (first attempt, 10.6 s after PutKeyPolicy); GetTableEncryption
  {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: PutKeyPolicy(maintenance statement with the given Condition) -> sleep 10 s -> CreateTable(aws:kms,
    key) (up to 3 attempts 15 s apart)
  - measurements: attempts_per_cell=[3, 1, 1, 1], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-141](../table-policy.md#s3t-table-141) · hypotheses: H-T-166, H-T-163 ·
    evidence: table/dependencies/kms-maintenance-principal

## Notes

H-T-166 confirmed: the bucket-scoped condition passed and the name-scoped condition failed (note the real
table ARN carries a UUID, not the name, so a name-scoped pattern cannot match any table ARN either); the
<bucket>/table/* pattern passed, which contradicts the 'bucket-scoped zero-byte object' explanation. H-T-163
second half (success once the maintenance statement is present) confirmed.

Full original title: Maintenance-principal key-policy condition scoping at CreateTable:
<bucket>/table/<name>/* REJECTED, <bucket>/table/* OK, <bucket>/* OK, no condition OK
