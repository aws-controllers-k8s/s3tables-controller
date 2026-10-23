<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-308: GetTable (by triple or tableArn) omits encryptionConfiguration, storageClassConfiguration, metadata/schema and tags; each needs its own read
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-308"></a>**S3T-TABLE-308** `read-gap` · impact high · handled · verified 2026-10-14
  **GetTable (by triple or tableArn) omits encryptionConfiguration, storageClassConfiguration, metadata/schema and tags; each needs its own read**
  CreateTable(ICEBERG, metadata.iceberg.schema with 2 fields, encryptionConfiguration={aws:kms,K},
  storageClassConfiguration=INTELLIGENT_TIERING, 3 tags) -> HTTP 200, body keys ['tableARN', 'versionToken'].
  GetTable by (tableBucketARN, namespace, name) -> HTTP 200 with exactly the raw JSON keys ['createdAt',
  'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name', 'namespace', 'namespaceId',
  'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']; GetTable by
  tableArn -> HTTP 200 with keys ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt',
  'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] (identical key set: True; differing values: []). None of ['encryptionConfiguration',
  'storageClassConfiguration', 'metadata', 'schema', 'tags'] appears in either body nor in the ListTables
  entry (keys ['createdAt', 'modifiedAt', 'name', 'namespace', 'tableARN', 'type']). The inputs are readable
  only via: GetTableEncryption -> HTTP 200 {"kmsKeyArn": "<the key ARN sent>", "sseAlgorithm": "aws:kms"};
  GetTableStorageClass -> HTTP 200 {"storageClass": "INTELLIGENT_TIERING"}; ListTagsForResource(tableArn) ->
  HTTP 200 same 3 keys/values as sent; GetTableMetadataLocation -> HTTP 200 keys ['metadataLocation',
  'versionToken', 'warehouseLocation'] (metadataLocation present
  (s3://<table-bucket-uuid>--table-s3/metadata/<file>.metadata.json)); the schema itself is never returned by
  any API (botocore model 2018-05-10: operations whose OUTPUT carries a
  schema/schemaV2/partitionSpec/sortOrder/writeOrder/properties/fields member: none; INPUTs carrying them:
  ['CreateTable']).
  - ACK: hook:sdk_read_one_post_set_output, custom_find, tags.custom-sync, is_immutable · ops: CreateTable,
    GetTable, ListTables, GetTableEncryption, GetTableStorageClass, ListTagsForResource,
    GetTableMetadataLocation · fields: metadata, metadata.iceberg.schema, encryptionConfiguration,
    storageClassConfiguration, tags, metadataLocation
  - repro: CreateTable(format=ICEBERG, metadata.iceberg.schema, encryptionConfiguration={aws:kms,K},
    storageClassConfiguration, tags) -> GetTable(triple), GetTable(tableArn), ListTables, GetTableEncryption,
    GetTableStorageClass, ListTagsForResource(tableArn), GetTableMetadataLocation
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181`
  - related: [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-010](../tablebucket-encryption-storage.md#s3t-tablebucket-010), [S3T-TABLE-060](../table-encryption-storage.md#s3t-table-060) · evidence:
    table/round-trip/fu-06-full-create-read-paths

## Notes

FU-06. Single table, one run (the key set is deterministic from the model: GetTable output members are
['name', 'type', 'tableARN', 'namespace', 'namespaceId', 'versionToken', 'metadataLocation',
'warehouseLocation', 'createdAt', 'createdBy', 'managedByService', 'modifiedAt', 'modifiedBy',
'ownerAccountId', 'format', 'tableBucketId', 'managedTableInformation']). ReadOne needs three extra calls
(encryption, storage class, tags) to see what was sent; the Iceberg schema is only inside the metadata.json at
metadataLocation (data plane, not read here), so a controller cannot diff spec.metadata against anything.
