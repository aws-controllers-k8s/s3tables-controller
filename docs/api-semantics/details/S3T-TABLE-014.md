<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-014: CreateTable without metadata: GetTable/GetTableMetadataLocation omit metadataLocation (key absent), warehouseLocation set...
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-014"></a>**S3T-TABLE-014** `response-fidelity` · impact high · handled · verified 2026-10-13
  **CreateTable without metadata: GetTable/GetTableMetadataLocation omit metadataLocation (key absent), warehouseLocation set...**
  CreateTable(format=ICEBERG, no metadata) -> HTTP 200 with keys ['tableARN', 'versionToken']. Immediately and
  at 5 s GetTable returned keys ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace',
  'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']
  (metadataLocation key absent at 0 s and 5 s; warehouseLocation
  's3://3e467aed-5b02-470e-y5yt6ojsai4r8h5gyw1iypjgr3znhusw2b--table-s3'; type=customer; managedByService
  <absent>; managedTableInformation <absent>). GetTableMetadataLocation -> HTTP 200 with keys ['versionToken',
  'warehouseLocation'] (metadataLocation absent). versionToken identical across CreateTable response, GetTable
  and GetTableMetadataLocation: True.
  - ACK: hook:sdk_read_one_post_set_output, docs-only · ops: CreateTable, GetTable, GetTableMetadataLocation ·
    fields: metadataLocation, warehouseLocation, versionToken, type
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(format=ICEBERG) with no metadata -> GetTable +
    GetTableMetadataLocation at 0 s and 5 s
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-033](../table-schema.md#s3t-table-033), [S3T-TABLE-309](../table-maintenance.md#s3t-table-309), [S3T-TABLE-315](../table-metadata-location.md#s3t-table-315) · hypotheses: H-T-101, H-T-143,
    H-T-129 · evidence: table/round-trip/metadata-create-readback

## Notes

H-T-101 confirmed (metadataLocation key absent, not empty string; warehouseLocation already set). H-T-143
confirmed (GetTableMetadataLocation is 200 on an empty table, not 404). H-T-129 qualified: a
CreateTable-created table reads type=customer; managedByService <absent>; no aws-managed table was available
to compare. Single table, n=1 per read.

Full original title: CreateTable without metadata: GetTable/GetTableMetadataLocation omit metadataLocation
(key absent), warehouseLocation set, versionToken present
