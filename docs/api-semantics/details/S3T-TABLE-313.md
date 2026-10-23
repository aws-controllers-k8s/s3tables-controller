<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-313: Table 'type' is 'customer' in GetTable and ListTables (agree); schema field types live only in metadata.json, in no API output
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-313"></a>**S3T-TABLE-313** `read-gap` · impact medium · partially handled · verified 2026-10-14
  **Table 'type' is 'customer' in GetTable and ListTables (agree); schema field types live only in metadata.json, in no API output**
  GetTable wire JSON members: ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name',
  'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] (type='customer', format='ICEBERG'); ListTables summary wire members: ['createdAt',
  'modifiedAt', 'name', 'namespace', 'tableARN', 'type'] (type='customer', format present: False); Get and
  List agree on type: True. No s3tables output shape carries schema/fields (model walk: only
  GetTable.format/type and ListTables.tables[].type exist). The field type strings sent in CreateTable ({"id":
  "int", "label": "string", "ts": "long"}) appear in the Iceberg metadata.json as {"id": "int", "label":
  "string", "ts": "long"}; explicit field ids {"a": 5, "b": 7, "c": 9} sent for the second table came back as
  {"a": 1, "b": 2, "c": 3} with types {"a": "int", "b": "decimal(10, 2)", "c": "timestamp"}.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: GetTable, ListTables, CreateTable · fields:
    type, format, metadata.iceberg.schema.fields[].type, metadata.iceberg.schema.fields[].id
  - repro: CreateTable(ICEBERG, schema) -> GetTable + ListTables(namespace) raw JSON; model walk for
    schema-bearing outputs; s3:GetObject(metadataLocation)
  - handling: partially handled via `generator.yaml:45-46; generator.yaml:60-65` - see Handling gaps
  - related: [S3T-TABLE-037](../table-schema.md#s3t-table-037), [S3T-TABLE-038](../table-schema.md#s3t-table-038), [S3T-TABLE-039](../table-schema.md#s3t-table-039), [S3T-TABLE-040](../table-schema.md#s3t-table-040), [S3T-TABLE-314](../table-schema.md#s3t-table-314), [S3T-TABLE-018](../table-schema.md#s3t-table-018),
    [S3T-TABLE-310](../table-schema.md#s3t-table-310), [S3T-TABLE-090](../table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](../table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028), [S3T-TABLE-231](../table.md#s3t-table-231), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089) ·
    evidence: table/round-trip/fu-08-schema-required-metadata-json

## Notes

FU-10. The table type member is literally named 'type' with the value shown (TableType enum customer|aws); a
spec.metadata.iceberg.schema can never be diffed against the API, only against metadata.json. Creative extra
(second table): explicit SchemaField.id values 5/7/9 were IGNORED and renumbered 1/2/3 in metadata.json, and
the type string 'decimal(10,2)' was normalized to 'decimal(10, 2)' (space inserted); 'timestamp' and
primitives were kept verbatim. ListTables summaries carried only createdAt, modifiedAt, name, namespace,
tableARN, type on the wire (no namespaceId/tableBucketId/managedByService although the model has them).

Partially handled in controller via: [GT-S3T-041](../service.md#gt-s3t-041) (controller hooks catalog entry): go_tag: json:"type,omitempty" overrides on Table.Type,
TableBucket.Type and Metadata.Iceberg.Schema.Fields.Type so the CRD exposes `type`. (scorer verdict: partial
match)
