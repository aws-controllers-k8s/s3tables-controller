<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-074: partitionSpec fieldId (0, 5, 2) and specId (3) IGNORED (field-id 1000,1001..., spec-id 0); fields=[] accepted as an unpartitioned spec
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-074"></a>**S3T-TABLE-074** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec fieldId (0, 5, 2) and specId (3) IGNORED (field-id 1000,1001..., spec-id 0); fields=[] accepted as an unpartitioned spec**
  fieldId omitted -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day",
  "transform": "day", "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=0
  -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day",
  "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=5 -> HTTP 200
  accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id":
  1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=2 (equals a schema field id) ->
  HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day",
  "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. specId=3 -> HTTP 200 accepted
  -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id": 1,
  "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. Two fields without fieldId -> HTTP 200
  accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id":
  1, "field-id": 1000}, {"name": "id_bucket", "transform": "bucket[8]", "source-id": 2, "field-id": 1001}]}]
  default-spec-id=0 last-partition-id=1001. partitionSpec.fields=[] -> HTTP 200 accepted ->
  partition-specs=[{"spec-id": 0, "fields": []}] default-spec-id=0 last-partition-id=999.
  - ACK: docs-only, late_initialize · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.fieldId, metadata.iceberg.partitionSpec.specId
  - repro: CreateTable with the partitionSpec variants above; GetObject metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-284](../table-schema.md#s3t-table-284), [S3T-TABLE-285](../table-schema.md#s3t-table-285), [S3T-TABLE-286](../table-schema.md#s3t-table-286), [S3T-TABLE-287](../table-schema.md#s3t-table-287), [S3T-TABLE-288](../table-schema.md#s3t-table-288), [S3T-TABLE-289](../table-schema.md#s3t-table-289),
    [S3T-TABLE-075](../table-schema.md#s3t-table-075), [S3T-TABLE-072](../table-schema.md#s3t-table-072), [S3T-TABLE-073](../table-schema.md#s3t-table-073), [S3T-TABLE-133](../table-schema.md#s3t-table-133) · hypotheses: H-T-122 · evidence:
    table/weird-inputs/partition-sort-variants

## Notes

H-T-122 QUALIFIED/REFUTED: fieldId omitted -> 1000 (confirmed), but explicit fieldId 0, 5 and 2 are all
rewritten to 1000 (not written verbatim), specId=3 is written as spec-id 0 / default-spec-id 0 (not 3); two
fields without fieldId -> 1000, 1001 with last-partition-id 1001; partitionSpec.fields=[] -> accepted,
partition-specs=[{spec-id 0, fields []}], last-partition-id 999. One attempt each.

Full original title: partitionSpec fieldId (0, 5, 2) and specId (3) are IGNORED: field-id always 1000,1001,...
and spec-id 0; fields=[] accepted as an unpartitioned spec
