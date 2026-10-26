<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-073: partitionSpec transform not grammar-checked: 'bucket' (no arity), 'days', '' written verbatim, 'DAY' lower-cased; only day(long) -> 400
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-073"></a>**S3T-TABLE-073** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec transform not grammar-checked: 'bucket' (no arity), 'days', '' written verbatim, 'DAY' lower-cased; only day(long) -> 400**
  On schemaV2 {1 ts timestamp, 2 id long}, one partition field per attempt. Accepted (partition-specs as
  written): {"identity(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "identity", "source-id":
  1, "field-id": 1000}]}], "year(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "year",
  "source-id": 1, "field-id": 1000}]}], "month(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform":
  "month", "source-id": 1, "field-id": 1000}]}], "day(ts)": [{"spec-id": 0, "fields": [{"name": "pcol",
  "transform": "day", "source-id": 1, "field-id": 1000}]}], "hour(ts)": [{"spec-id": 0, "fields": [{"name":
  "pcol", "transform": "hour", "source-id": 1, "field-id": 1000}]}], "bucket\[16\](id)": [{"spec-id": 0,
  "fields": [{"name": "pcol", "transform": "bucket[16]", "source-id": 2, "field-id": 1000}]}],
  "truncate\[4\](id)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "truncate[4]", "source-id": 2,
  "field-id": 1000}]}], "void(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "void",
  "source-id": 1, "field-id": 1000}]}], "bucket(id)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform":
  "bucket", "source-id": 2, "field-id": 1000}]}], "days(ts)": [{"spec-id": 0, "fields": [{"name": "pcol",
  "transform": "days", "source-id": 1, "field-id": 1000}]}], "DAY(ts)": [{"spec-id": 0, "fields": [{"name":
  "pcol", "transform": "day", "source-id": 1, "field-id": 1000}]}], "(ts)": [{"spec-id": 0, "fields":
  [{"name": "pcol", "transform": "", "source-id": 1, "field-id": 1000}]}]}. Rejected: {"day(id)":
  "BadRequestException 400 'The specified metadata is not valid'"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.transform
  - repro: Loop CreateTable over the transforms above; GetObject metadata.json; DeleteTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-072](../table-schema.md#s3t-table-072), [S3T-TABLE-074](../table-schema.md#s3t-table-074), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-038](../table-schema.md#s3t-table-038), [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-042](../table-schema.md#s3t-table-042) ·
    hypotheses: H-T-121 · evidence: table/weird-inputs/partition-sort-variants

## Notes

H-T-121 REFUTED for the grammar part: identity/year/month/day/hour/bucket[16]/truncate[4]/void accepted as
expected, but 'bucket' (no arity), 'days' and the empty string '' were ALSO accepted and persisted verbatim in
partition-specs (an Iceberg engine cannot parse 'bucket' or ''), and 'DAY' was accepted and normalized to
'day'. Confirmed only for the type-compatibility part: transform 'day' on a long field -> BadRequestException
400 'The specified metadata is not valid'. One attempt per transform.

Full original title: partitionSpec transform is NOT validated against the Iceberg grammar: 'bucket' (no
arity), 'days' and '' are written VERBATIM; 'DAY' lower-cased; only day(long) type mismatch -> 400
