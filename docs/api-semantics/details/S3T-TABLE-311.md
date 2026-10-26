<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-311: CreateTable metadata={} -> client-side ParamValidationError (validation on) / BadRequestException 400 'metadata is not valid' (off)
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-311"></a>**S3T-TABLE-311** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable metadata={} -> client-side ParamValidationError (validation on) / BadRequestException 400 'metadata is not valid' (off)**
  With botocore parameter validation ON, CreateTable(metadata={}) -> client-side ParamValidationError:
  Parameter validation failed: Must set one of the following keys for tagged unionstructure metadata:
  iceberg.. With validation OFF the same request -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid' (second attempt: BadRequestException (HTTP 400)); wire body {"name": "ackq_8d6c9c_empty_off1",
  "format": "ICEBERG", "metadata": {}}. metadata={iceberg:{}}: validation on -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'; off -> BadRequestException (HTTP 400).
  metadata.iceberg.schema.fields=[] -> HTTP 200. A field with type 'bogus' -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'. Unknown union key {notIceberg:{schema}} (validation off) ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'; iceberg plus an unknown sibling key ->
  HTTP 200; schema and schemaV2 both set -> BadRequestException (HTTP 400) 'Both 'schema' and 'schemaV2' were
  specified. Use either 'schema' or 'schemaV2' and retry the create table command.'; JSON null metadata on the
  wire -> HTTP 200.
  - ACK: terminal_codes, input_wrapper_field_path · ops: CreateTable · fields: metadata, metadata.iceberg,
    metadata.iceberg.schema.fields, metadata.iceberg.schema.fields.type
  - repro: CreateTable(format=ICEBERG, metadata={}) with botocore parameter_validation True then False; then
    the variants listed in behavior
  - measurements: n_empty_validation_off=2, n_per_other_cell=1, n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-036](../table-schema.md#s3t-table-036), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-041](../table-schema.md#s3t-table-041), [S3T-TABLE-194](../table-schema.md#s3t-table-194), [S3T-TABLE-015](../table-schema.md#s3t-table-015), [S3T-TABLE-034](../table-metadata-location.md#s3t-table-034),
    [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316) · evidence: table/weird-inputs/fu-03-metadata-empty-union

## Notes

FU-03 items 2-3. The Go SDK v2 tagged-union (interface) encoding cannot be observed from Python; what IS
observable is whether the server accepts an empty union object and the wire code it returns. For every
accepted cell GetTable.metadataLocation present (= a schema was applied): {'schema_fields_empty_list': True,
'iceberg_plus_unknown_key_off': True, 'metadata_null_wire': False}. Tables created by any accepted cell are
deleted at the end.
