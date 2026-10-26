<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-075: writeOrder.orderId IGNORED (omitted/2/7/-1 -> order-id 1) except orderId=0 with fields -> 400; unknown sourceId -> 400
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-075"></a>**S3T-TABLE-075** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **writeOrder.orderId IGNORED (omitted/2/7/-1 -> order-id 1) except orderId=0 with fields -> 400; unknown sourceId -> 400**
  schemaV2 {1 ts timestamp, 2 id long}; sort field identity/asc/nulls-first on id unless noted. sourceId:9 ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'. orderId omitted -> HTTP 200 accepted
  -> sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=0 (non-empty fields) -> BadRequestException
  (HTTP 400) 'The specified metadata is not valid'. orderId=2 -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=7 -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=-1 -> HTTP 200 accepted. Two sort fields
  (day(ts) asc nulls-first, id desc nulls-last) -> HTTP 200 accepted -> sort-orders=[{"order-id": 1, "fields":
  [{"transform": "day", "source-id": 1, "direction": "asc", "null-order": "nulls-first"}, {"transform":
  "identity", "source-id": 2, "direction": "desc", "null-order": "nulls-last"}]}] default-sort-order-id=1.
  - ACK: terminal_codes, docs-only, compare.nil_equals_zero_value · ops: CreateTable · fields:
    metadata.iceberg.writeOrder.orderId, metadata.iceberg.writeOrder.fields.sourceId
  - repro: CreateTable with the writeOrder variants above (parameter_validation=False for the omitted
    orderId); GetObject metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-284](../table-schema.md#s3t-table-284), [S3T-TABLE-285](../table-schema.md#s3t-table-285), [S3T-TABLE-286](../table-schema.md#s3t-table-286), [S3T-TABLE-287](../table-schema.md#s3t-table-287), [S3T-TABLE-288](../table-schema.md#s3t-table-288), [S3T-TABLE-289](../table-schema.md#s3t-table-289),
    [S3T-TABLE-074](../table-schema.md#s3t-table-074), [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-077](../table-schema.md#s3t-table-077), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-018](../table-schema.md#s3t-table-018) · hypotheses: H-T-123, H-T-193,
    H-T-194 · evidence: table/weird-inputs/partition-sort-variants

## Notes

H-T-123 QUALIFIED: unknown sourceId 9 -> 400 'The specified metadata is not valid' (confirmed); omitted
orderId -> order-id 1 / default-sort-order-id 1 (confirmed); but orderId=2 is written as order-id 1 (not 2).
H-T-193 REFUTED: orderId=7 -> accepted and written as order-id 1 (not honored); orderId=-1 -> ACCEPTED (not
400), also written as 1. H-T-194 CONFIRMED: orderId=0 with a non-empty fields list -> BadRequestException 400
(a Go zero-value orderId would permanently fail). Two sort fields (day(ts) asc nulls-first, identity(id) desc
nulls-last) -> written verbatim in order. One attempt each.

Full original title: writeOrder.orderId is IGNORED (omitted/2/7/-1 all written as order-id 1,
default-sort-order-id 1) except orderId=0 with fields -> 400; unknown sourceId -> 400
