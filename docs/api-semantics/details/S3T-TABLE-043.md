<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-043: schemaV2 field ids are RENUMBERED 1..n (sparse 1,5,42 -> 1,2,3; 0 and -1 accepted); only duplicate ids rejected (400)
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-043"></a>**S3T-TABLE-043** `normalization` · impact high · handled · verified 2026-10-13
  **schemaV2 field ids are RENUMBERED 1..n (sparse 1,5,42 -> 1,2,3; 0 and -1 accepted); only duplicate ids rejected (400)**
  ids {1,1} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'. ids {0,1} -> HTTP 200
  accepted (metadata.json fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name':
  'b', 'type': 'string', 'required': False}], last-column-id=2). ids {-1,1} -> HTTP 200 accepted
  (metadata.json fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name': 'b',
  'type': 'string', 'required': False}], last-column-id=2). ids {1,5,42} -> HTTP 200 accepted (metadata.json
  fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name': 'b', 'type': 'string',
  'required': False}, {'id': 3, 'name': 'c', 'type': 'long', 'required': False}], last-column-id=3).
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schemaV2.fields.id
  - repro: CreateTable schemaV2 with the four id sets; GetObject metadata.json for accepted ones
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; 5c1bda2; 007c152`
  - related: [S3T-TABLE-041](../table-schema.md#s3t-table-041), [S3T-TABLE-042](../table-schema.md#s3t-table-042), [S3T-TABLE-044](../table-schema.md#s3t-table-044), [S3T-TABLE-045](../table-schema.md#s3t-table-045), [S3T-TABLE-046](../table-schema.md#s3t-table-046), [S3T-TABLE-039](../table-schema.md#s3t-table-039) ·
    hypotheses: H-T-116 · evidence: table/weird-inputs/schema-variants

## Notes

H-T-116 REFUTED in its pass-through part: ids {0,1} and {-1,1} are accepted (written 1,2), sparse {1,5,42}
accepted but written as 1,2,3 with last-column-id=3; duplicates {1,1} -> 400 'The specified metadata is not
valid'. Nested ids are also reassigned (struct child sent id 2 -> written 3; list element-id 4 kept at 4
because of the top-level-first assignment order). schemaV2 cannot pin Iceberg field ids. One attempt each.
