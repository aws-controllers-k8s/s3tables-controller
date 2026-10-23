<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-315: CreateTable returns only {tableARN, versionToken}; GetTable adds 12 members (metadataLocation only once the table has metadata)
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-315"></a>**S3T-TABLE-315** `read-gap` · impact high · handled · verified 2026-10-14
  **CreateTable returns only {tableARN, versionToken}; GetTable adds 12 members (metadataLocation only once the table has metadata)**
  CreateTableBucket wire body keys ['arn']; CreateNamespace -> HTTP 200 keys ['namespace', 'tableBucketARN']
  (echoes the parent tableBucketARN: True); GetNamespace right after -> HTTP 200 keys ['createdAt',
  'createdBy', 'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId']. CreateTable(ICEBERG, no
  metadata) and CreateTable(ICEBERG, 2-field schema) -> HTTP 200/200 with wire keys ['tableARN',
  'versionToken'] (no tableBucketARN member; the tableARN starts with '<bucket ARN>/': True). GetTable right
  after Create (HTTP 200) had keys ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace',
  'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']
  for the table without metadata and ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt',
  'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] for the table with a schema: present only in the Get: ['createdAt', 'createdBy',
  'format', 'modifiedAt', 'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId', 'type',
  'warehouseLocation'] (+ metadataLocation only with metadata). Modeled GetTable members absent from both wire
  bodies on a fresh table: ['managedByService', 'managedTableInformation', 'modifiedBy']. versionToken in Get ==
  Create: True. ListTables entry wire keys: ['createdAt', 'modifiedAt', 'name', 'namespace', 'tableARN',
  'type'] (type equals GetTable.type: True). At +3 s the GetTable key sets were unchanged: True.
  - ACK: hook:sdk_create_post_set_output, is_read_only, late_initialize, is_arn_primary_key · ops:
    CreateNamespace, GetNamespace, CreateTable, GetTable, ListTables · fields: tableARN, versionToken,
    metadataLocation, warehouseLocation, namespaceId, tableBucketId, createdAt, createdBy, modifiedAt,
    modifiedBy, type, format, managedByService
  - repro: CreateTableBucket -> CreateNamespace -> GetNamespace -> CreateTable (no metadata) + CreateTable
    (schema) -> GetTable each immediately and at +3 s -> ListTables; compare raw JSON key sets
  - measurements: n_tables=2, n=2
  - handling: handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-014](../table-metadata-location.md#s3t-table-014), [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-033](../table-schema.md#s3t-table-033), [S3T-TABLE-309](../table-maintenance.md#s3t-table-309), [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001) ·
    evidence: table/round-trip/fu-12-create-vs-get-members

## Notes

FU-12 (and FU-10: ListTables.type agrees with GetTable.type). Exact wire names are camelCase with 'Id'
(namespaceId, tableBucketId, ownerAccountId). modifiedBy and managedByService are NOT returned on a fresh
table even though the model has them (see the FU-11 finding for when modifiedBy appears); ListTables entries
omit namespaceId/tableBucketId/managedByService on the wire. Single run, two tables.
