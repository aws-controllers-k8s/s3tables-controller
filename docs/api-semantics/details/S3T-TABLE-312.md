<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-312: Schema 'required' round-trips into metadata.json: required=true -> True, required=false -> False, omitted -> False
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-312"></a>**S3T-TABLE-312** `normalization` · impact medium · partially handled · verified 2026-10-14
  **Schema 'required' round-trips into metadata.json: required=true -> True, required=false -> False, omitted -> False**
  CreateTable(format=ICEBERG,
  metadata.iceberg.schema.fields=[{id:int,required:true},{label:string,required:false},{ts:long (required
  omitted)}]) -> HTTP 200. GetTable.metadataLocation =
  s3://<table-uuid-prefix><opaque>--table-s3/metadata/00000-<uuid>.metadata.json; S3 GetObject on it -> HTTP
  200 (809 bytes, ServerSideEncryption=AES256). The Iceberg metadata (format-version 2, current-schema-id 0, 1
  schema(s), v1 'schema' member present: False) lists fields ['id', 'label', 'ts']: 'required' per field =
  {"id": true, "label": false, "ts": false}, 'type' per field = {"id": "int", "label": "string", "ts":
  "long"}, field ids = {"id": 1, "label": 2, "ts": 3}. The metadata has 0 snapshots,
  location=<warehouseLocation>, properties keys ['write.parquet.compression-codec'].
  - ACK: none, is_immutable · ops: CreateTable, GetTable, GetTableMetadataLocation · fields:
    metadata.iceberg.schema.fields[].required, metadata.iceberg.schema.fields[].type, metadataLocation
  - repro: CreateTable with the three fields -> GetTable -> parse metadataLocation ->
    s3:GetObject(bucket=<warehouse bucket>, key=metadata/00000-<uuid>.metadata.json) ->
    schemas[current].fields[].required
  - handling: partially handled via `5c1bda2; 007c152` - see Handling gaps
  - related: [S3T-TABLE-318](../table-maintenance.md#s3t-table-318), [S3T-TABLE-319](../table-encryption-storage.md#s3t-table-319) · evidence: table/round-trip/fu-08-schema-required-metadata-json

## Notes

FU-08. required=true is preserved as a non-null column (true), required=false and an omitted required both
become False in the metadata; the SDK-side pointer/omission handling itself is not observable from the API
(only its effect in metadata.json is). The schema is never returned by any s3tables operation (GetTable keys:
['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name', 'namespace', 'namespaceId',
'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']), so the read back
needs s3:GetObject on the warehouse bucket (allowed for the account's caller here).

Partially handled in controller via: [GT-S3T-037](../service.md#gt-s3t-037) (controller hooks catalog entry): Originally dropped via ignore.field_paths
SchemaField.Required; now generated as `elem.Required = *iter.Required` so non-null Iceberg columns can be
created. (scorer verdict: partial match)
