<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-062: GetTableEncryption/GetTableStorageClass take only the exact bucketARN+namespace+name triple: table ARN -> 400; old name after Rename -> 404
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-062"></a>**S3T-TABLE-062** `identity` · impact medium · handled · verified 2026-10-13
  **GetTableEncryption/GetTableStorageClass take only the exact bucketARN+namespace+name triple: table ARN -> 400; old name after Rename -> 404**
  On the {IT + aws:kms} table: tableBucketARN=<table ARN> -> enc BadRequestException 400 / sc
  BadRequestException 400; tableBucketARN with trailing slash -> enc BadRequestException 400 / sc
  BadRequestException 400; tableBucketARN bucket name UPPERCASED -> enc BadRequestException 400 / sc
  BadRequestException 400; name UPPERCASED -> enc NotFoundException 404 / sc NotFoundException 404; namespace
  as the namespace id -> enc NotFoundException 404 / sc NotFoundException 404. RenameTable(newName) -> HTTP
  204; old triple -> enc NotFoundException 404 / sc NotFoundException 404; new triple -> enc {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"} / sc
  {"storageClass": "INTELLIGENT_TIERING"}; GetTable(tableArn=<ARN from before the rename>) -> HTTP 200,
  name=ackq_5fda45_it_kms_renamed.
  - ACK: is_arn_primary_key, custom_find, references · ops: GetTableEncryption, GetTableStorageClass,
    RenameTable, GetTable · fields: tableBucketARN, namespace, name
  - repro: GetTableEncryption/GetTableStorageClass with each identifier form, then RenameTable and re-read
    with old and new names
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-011](../table-encryption-storage.md#s3t-table-011), [S3T-TABLE-063](../table-encryption-storage.md#s3t-table-063), [S3T-TABLE-252](../table-encryption-storage.md#s3t-table-252), [S3T-TABLEBUCKET-085](../tablebucket-encryption-storage.md#s3t-tablebucket-085) · hypotheses: H-T-180 ·
    evidence: table/round-trip/explicit-encryption-storage

## Notes

H-T-180 confirmed (table ARN in tableBucketARN gives BadRequestException, not NotFoundException; rename: old
NotFoundException / new 200).

Full original title: GetTableEncryption/GetTableStorageClass take only the bucketARN+namespace+name triple:
table ARN in tableBucketARN -> BadRequestException 400; after RenameTable the old name -> NotFoundException
404, new name keeps the config
