<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-090: ListTables summary carries only createdAt, modifiedAt, name, namespace, tableARN, type; no versionToken/metadataLocation/warehouseLocation
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-090"></a>**S3T-TABLE-090** `read-gap` · impact medium · handled · verified 2026-10-13
  **ListTables summary carries only createdAt, modifiedAt, name, namespace, tableARN, type; no versionToken/metadataLocation/warehouseLocation**
  ListTables(namespace, prefix=<name>) summary keys: ['createdAt', 'modifiedAt', 'name', 'namespace',
  'tableARN', 'type']. GetTable keys absent from the summary: ['createdBy', 'format', 'metadataLocation',
  'namespaceId', 'ownerAccountId', 'tableBucketId', 'versionToken', 'warehouseLocation']. Shared fields
  byte-identical to GetTable: True. namespace is a 1-element list in both; timestamps compare equal at
  microsecond precision: True.
  - ACK: list_operation.match_fields, custom_find · ops: ListTables, GetTable · fields: tables, versionToken,
    metadataLocation, warehouseLocation, format, namespaceId, tableBucketId, modifiedAt
  - repro: CreateTable(schema) -> ListTables(namespace, prefix=name) and GetTable(triple); compare key sets
    and values
  - handling: handled via `generator.yaml:10-13; generator.yaml:31-39; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-018](../table-schema.md#s3t-table-018), [S3T-TABLE-310](../table-schema.md#s3t-table-310), [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-023](../table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028) ·
    hypotheses: H-T-076 · evidence: table/weird-inputs/list-tables-bounds

## Notes

H-T-076 confirmed (predicted summary = namespace, name, type, tableARN, createdAt, modifiedAt [+
namespaceId/tableBucketId]; no versionToken/metadataLocation/warehouseLocation/format).

Full original title: ListTables summary carries createdAt, modifiedAt, name, namespace, tableARN, type;
GetTable-only fields: createdBy, format, metadataLocation, namespaceId, ownerAccountId, tableBucketId,
versionToken, warehouseLocation
