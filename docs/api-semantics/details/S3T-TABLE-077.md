<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-077: Empty writeOrder ({fields:[]}, {orderId:0|1, fields:[]}, {}) -> 200, written as the unsorted order (order-id 0, default-sort-order-id 0)
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-077"></a>**S3T-TABLE-077** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Empty writeOrder ({fields:[]}, {orderId:0|1, fields:[]}, {}) -> 200, written as the unsorted order (order-id 0, default-sort-order-id 0)**
  writeOrder={fields:[]} -> HTTP 200 accepted -> sort-orders=[{"order-id": 0, "fields": []}]
  default-sort-order-id=0. {orderId:0, fields:[]} -> HTTP 200 accepted -> sort-orders=[{"order-id": 0,
  "fields": []}] default-sort-order-id=0. {orderId:1, fields:[]} -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 0, "fields": []}] default-sort-order-id=0. writeOrder={} -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 0, "fields": []}] default-sort-order-id=0. (schema.fields=[] is covered by the
  schema-variants probe.)
  - ACK: terminal_codes, hook:sdk_create_pre_build_request · ops: CreateTable · fields:
    metadata.iceberg.writeOrder, metadata.iceberg.writeOrder.fields
  - repro: CreateTable with the empty writeOrder shapes above (parameter_validation=False); GetObject
    metadata.json if accepted
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-075](../table-schema.md#s3t-table-075), [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-018](../table-schema.md#s3t-table-018) · hypotheses: H-T-196 · evidence:
    table/weird-inputs/partition-sort-variants

## Notes

H-T-196 QUALIFIED: the accept part holds (no 400, unlike nothing - note schema.fields=[] is also accepted per
the schema-variants probe), but the written sort order is {order-id 0, fields []} with default-sort-order-id 0
for all four shapes, including orderId=1 - i.e. an empty writeOrder is equivalent to omitting it; a controller
that always emits the struct is harmless here. One attempt each.

Full original title: Empty writeOrder ({fields:[]}, {orderId:0|1, fields:[]}, {}) is accepted and written as
the unsorted order (order-id 0, default-sort-order-id 0), NOT order-id 1
