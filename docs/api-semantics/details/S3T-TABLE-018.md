<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-018: metadata.iceberg (schema/partitionSpec/writeOrder/properties) is write-only: GetTable, GetTableMetadataLocation, ListTables echo none of it
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-018"></a>**S3T-TABLE-018** `read-gap` · impact high · handled · verified 2026-10-15, re-verified
  **metadata.iceberg (schema/partitionSpec/writeOrder/properties) is write-only: GetTable, GetTableMetadataLocation, ListTables echo none of it**
  CreateTable with schema fields evt_ts_qx/evt_id_qx, partitionSpec(day on evt_ts_qx), writeOrder(identity asc
  nulls-first on evt_id_qx) and a marker property -> HTTP 200. The concatenated raw bodies of GetTable +
  GetTableMetadataLocation + ListTables contain none of the field names, partition name, property marker,
  transform/direction/null-order words or any orderId/sort/partition key (the only token hit, 'schema', is the
  name of a sibling table 'ackq_<run>_schema' in the ListTables body). The metadata.json at metadataLocation
  contains them: {'evt_ts_qx': True, 'evt_id_qx': True, 'evt_ts_day_qx': True, 'ackq_prop_marker': True,
  'order-id': True, 'sort-order': True, 'partition': True, 'schema': True, 'properties': True, 'identity':
  True, 'nulls-first': True, 'timestamp': True} (schemas=[{'type': 'struct', 'schema-id': 0, 'fields': [{'id':
  1, 'name': 'evt_ts_qx', 'required': False, 'type': 'timestamp'}, {'id': 2, 'name': 'evt_id_qx', 'required':
  False, 'type': 'long'}]}]; partition-specs=[{'spec-id': 0, 'fields': [{'name': 'evt_ts_day_qx', 'transform':
  'day', 'source-id': 1, 'field-id': 1000}]}]; sort-orders=[{'order-id': 1, 'fields': [{'transform':
  'identity', 'source-id': 2, 'direction': 'asc', 'null-order': 'nulls-first'}]}]; default-sort-order-id=1).
  - ACK: compare.is_ignored+delta_pre_compare, is_immutable, docs-only · ops: CreateTable, GetTable,
    GetTableMetadataLocation, ListTables · fields: metadata.iceberg.schema, metadata.iceberg.partitionSpec,
    metadata.iceberg.writeOrder, metadata.iceberg.properties
  - repro: CreateTable with schema+partitionSpec+writeOrder+properties -> dump
    GetTable/GetTableMetadataLocation/ListTables -> grep field names -> S3 GetObject metadataLocation
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-310](../table-schema.md#s3t-table-310), [S3T-TABLE-075](../table-schema.md#s3t-table-075), [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-077](../table-schema.md#s3t-table-077), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-017](../table-schema.md#s3t-table-017),
    [S3T-TABLE-194](../table-schema.md#s3t-table-194), [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-090](../table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](../table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028) · hypotheses:
    H-T-103, H-T-192 · evidence: table/round-trip/metadata-create-readback

## Notes

H-T-103 confirmed. H-T-192 confirmed: writeOrder with orderId omitted (SDK validation bypassed; the model
marks orderId required) -> HTTP 200; metadata.json default-sort-order-id=1, sort-orders order-ids=[1]; no
control-plane read exposes orderId. n=1 table. Side observation: S3 ListObjectsV2 on the warehouse bucket ->
MethodNotAllowed (HTTP 405); GetObject on the metadataLocation key works.

Full original title: metadata.iceberg (schema/partitionSpec/writeOrder/properties) is write-only: GetTable,
GetTableMetadataLocation and ListTables echo none of it
