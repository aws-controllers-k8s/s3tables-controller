<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-046: schemaV2 field type as a JSON document (nested struct + list) -> accepted; 'varchar' -> 400; field without 'required' -> 400 (no default)
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-046"></a>**S3T-TABLE-046** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **schemaV2 field type as a JSON document (nested struct + list) -> accepted; 'varchar' -> 400; field without 'required' -> 400 (no default)**
  schemaV2 fields [{1,s,{type:struct,fields:[{2,x,int}]}},{3,l,{type:list,element-id:4,element:string}}] ->
  HTTP 200 accepted (metadata.json fields=[{'id': 1, 'name': 's', 'type': {'type': 'struct', 'fields': [{'id':
  3, 'name': 'x', 'required': False, 'type': 'int'}]}, 'required': False}, {'id': 2, 'name': 'l', 'type':
  {'type': 'list', 'element-id': 4, 'element': 'string', 'element-required': False}, 'required': False}],
  last-column-id=4). schemaV2 field type 'varchar' -> BadRequestException (HTTP 400) 'The specified metadata
  is not valid'. schemaV2 field without 'required' (model: required) -> BadRequestException (HTTP 400) 'The
  specified metadata is not valid'.
  - ACK: docs-only, is_document · ops: CreateTable · fields: metadata.iceberg.schemaV2.fields.type,
    metadata.iceberg.schemaV2.fields.required
  - repro: CreateTable schemaV2 with a struct-typed and a list-typed field (Document type); GetObject
    metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-041](../table-schema.md#s3t-table-041), [S3T-TABLE-042](../table-schema.md#s3t-table-042), [S3T-TABLE-043](../table-schema.md#s3t-table-043), [S3T-TABLE-044](../table-schema.md#s3t-table-044), [S3T-TABLE-045](../table-schema.md#s3t-table-045), [S3T-TABLE-039](../table-schema.md#s3t-table-039) ·
    hypotheses: H-T-110, H-T-111 · evidence: table/weird-inputs/schema-variants

## Notes

H-T-110 complement: nested struct/list types are accepted through schemaV2 (Document-typed 'type') and written
as Iceberg nested types (ids reassigned); schemaV2 field type 'varchar' -> 400; schemaV2 field without
'required' -> 400 (not defaulted). H-T-111 qualified: in schemaV2 'required' must be sent. One attempt each.

Full original title: schemaV2 field type as a JSON document (nested struct + list) -> accepted; schemaV2
'varchar' -> BadRequestException 400; schemaV2 required omitted -> BadRequestException 400
