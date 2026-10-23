<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-072: partitionSpec sourceId: unknown 99, 0 or past-last 3 -> 400 'metadata is not valid'; id 1 accepted on schemaV2 and server-numbered v1 schema
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-072"></a>**S3T-TABLE-072** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec sourceId: unknown 99, 0 or past-last 3 -> 400 'metadata is not valid'; id 1 accepted on schemaV2 and server-numbered v1 schema**
  schemaV2 {1 ts timestamp, 2 id long} + partitionSpec [{sourceId:99, fieldId:1000, name:p, transform:day}] ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'. sourceId:1 -> HTTP 200 accepted ->
  partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id": 1,
  "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. v1 schema [{ts,timestamp},{id,long}]
  (server-assigned ids) + sourceId:1 -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields":
  [{"name": "ts_day", "transform": "day", "source-id": 1, "field-id": 1000}]}] default-spec-id=0
  last-partition-id=1000. sourceId:0 -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  sourceId:3 -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.sourceId, metadata.iceberg.schema
  - repro: CreateTable with the partitionSpec variants above; GetObject metadata.json for accepted ones
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-290](../table-schema.md#s3t-table-290), [S3T-TABLE-287](../table-schema.md#s3t-table-287), [S3T-TABLE-289](../table-schema.md#s3t-table-289), [S3T-TABLE-073](../table-schema.md#s3t-table-073), [S3T-TABLE-074](../table-schema.md#s3t-table-074), [S3T-TABLE-133](../table-schema.md#s3t-table-133) ·
    hypotheses: H-T-119, H-T-120 · evidence: table/weird-inputs/partition-sort-variants

## Notes

H-T-119 confirmed (unknown sourceId 99 -> 400 'The specified metadata is not valid' - the message does NOT
name the source id; sourceId 1 -> 200, partition-specs written with source-id 1, field-id 1000). H-T-120
confirmed: with the v1 schema sourceId 1 maps to the first (server-numbered) field; sourceId 0 and sourceId 3
(past the 2 fields) -> 400. One attempt each.

Full original title: partitionSpec sourceId: unknown 99 -> BadRequestException 400; valid (schemaV2 id 1) ->
accepted; v1 schema sourceId 1 -> accepted, 0 -> BadRequestException 400, 3 (past last field) ->
BadRequestException 400
