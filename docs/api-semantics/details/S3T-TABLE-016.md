<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-016: v1 schema: metadata.json fields get ids [1, 2, 3], required [False, True, False]...
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-016"></a>**S3T-TABLE-016** `server-default` · impact medium · handled · verified 2026-10-13
  **v1 schema: metadata.json fields get ids [1, 2, 3], required [False, True, False]...**
  For CreateTable metadata.iceberg.schema.fields=[{fld_alpha,int}(required
  omitted),{fld_bravo,string,required=true},{fld_charlie,long,required=false}] the metadata.json at
  metadataLocation has format-version=2, current-schema-id=0, schemas=[{'schema-id': 0, 'type': 'struct',
  'n_fields': 3}], last-column-id=3, partition-specs=[{'spec-id': 0, 'fields': []}] (default-spec-id=0,
  last-partition-id=999), sort-orders=[{'order-id': 0, 'fields': []}] (default-sort-order-id=0),
  current-snapshot-id=-1, snapshots=[], identifier-field-ids=None. Written fields: [{'id': 1, 'name':
  'fld_alpha', 'required': False, 'type': 'int'}, {'id': 2, 'name': 'fld_bravo', 'required': True, 'type':
  'string'}, {'id': 3, 'name': 'fld_charlie', 'required': False, 'type': 'long'}].
  - ACK: docs-only, is_immutable · ops: CreateTable · fields: metadata.iceberg.schema.fields,
    metadata.iceberg.schema.fields.required, metadata.iceberg.writeOrder.orderId
  - repro: CreateTable with the 3-field v1 schema above -> GetTable.metadataLocation -> S3 GetObject ->
    inspect schemas/sort-orders/partition-specs
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; 5c1bda2; 007c152`
  - related: [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211) · hypotheses: H-T-111, H-T-112,
    H-T-195 · evidence: table/round-trip/metadata-create-readback

## Notes

H-T-111 confirmed (required omitted -> False; required=true -> True; no extra identifier field:
identifier-field-ids=None). H-T-112 confirmed (ids [1, 2, 3], schema-id 0, format-version 2). H-T-195
confirmed (writeOrder omitted -> default-sort-order-id=0, sort-orders=[{'order-id': 0, 'fields': []}]). n=1
table.

Full original title: v1 schema: metadata.json fields get ids [1, 2, 3], required [False, True, False];
format-version 2, schema-id 0, default-sort-order-id 0, default-spec-id 0
