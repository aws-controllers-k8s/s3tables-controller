<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-322: CreateTable metadata.iceberg.properties: metadata.json keeps the user property and adds write.parquet.compression-codec=zstd
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-322"></a>**S3T-TABLE-322** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable metadata.iceberg.properties: metadata.json keeps the user property and adds write.parquet.compression-codec=zstd**
  CreateTable(format=ICEBERG, metadata.iceberg with schema [evt_ts_qx timestamp, evt_id_qx long],
  partitionSpec day(evt_ts_qx) as evt_ts_day_qx, writeOrder orderId=1 identity(evt_id_qx) asc nulls-first,
  properties {ackq_prop_marker: ackq_prop_value_qx}) -> HTTP 200. S3 GetObject of
  GetTableMetadataLocation.metadataLocation (s3://<warehouse>--table-s3/metadata/00000-<uuid>.metadata.json)
  -> HTTP 200; metadata.json format-version=2, properties={"ackq_prop_marker": "ackq_prop_value_qx",
  "write.parquet.compression-codec": "zstd"}, schemas=[{"type": "struct", "schema-id": 0, "fields": [{"id": 1,
  "name": "evt_ts_qx", "required": false, "type": "timestamp"}, {"id": 2, "name": "evt_id_qx", "required":
  false, "type": "long"}]}], partition-specs=[{"spec-id": 0, "fields": [{"name": "evt_ts_day_qx", "transform":
  "day", "source-id": 1, "field-id": 1000}]}], sort-orders=[{"order-id": 1, "fields": [{"transform":
  "identity", "source-id": 2, "direction": "asc", "null-order": "nulls-first"}]}], default-sort-order-id=1.
  The raw bodies of GetTable, GetTableMetadataLocation and ListTables contain 0 of 19 probed tokens (field
  names, partition name, property marker/value, transform/direction/null-order words, 'schema', 'properties').
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: CreateTable, GetTableMetadataLocation,
    GetTable, ListTables · fields: metadata.iceberg.properties, metadata.iceberg.schema,
    metadata.iceberg.partitionSpec, metadata.iceberg.writeOrder, metadataLocation
  - repro: CreateTable(schema+partitionSpec+writeOrder+properties) -> raw
    GetTable/GetTableMetadataLocation/ListTables bodies -> s3 GetObject metadataLocation
  - measurements: api_body_token_hits=0, tokens_probed=19, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-018](../table-schema.md#s3t-table-018) · evidence: table/creative/reverify-set-a

## Notes

Re-verification of [S3T-TABLE-018](../table-schema.md#s3t-table-018) (confirmed: zero API-body hits, no sibling-name false positive this time).
Precision gain: the only way to read metadata.iceberg back is the metadata.json in S3, and that document is
not byte-equal to the request (server adds Iceberg defaults such as the compression codec, ids and spec/order
ids), so even an S3-based read-back needs a semantic compare.

Full original title: CreateTable metadata.iceberg.properties: the committed metadata.json carries the user
property plus server-added write.parquet.compression-codec=zstd
