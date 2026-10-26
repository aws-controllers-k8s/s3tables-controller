<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-036: CreateTable schema.fields=[] -> 200 (empty-schema table committed); schema={} -> 500; iceberg={}, metadata={}, properties-only -> 400
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-036"></a>**S3T-TABLE-036** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable schema.fields=[] -> 200 (empty-schema table committed); schema={} -> 500; iceberg={}, metadata={}, properties-only -> 400**
  With SDK validation disabled: metadata.iceberg.schema.fields=[] -> HTTP 200 accepted (metadata.json
  fields=[], last-column-id=0). schema={} -> InternalServerErrorException (HTTP 500) 'An internal error
  occurred. Try again.'. iceberg={} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  metadata={} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  iceberg={properties:{...}} without any schema -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid'.
  - ACK: terminal_codes, hook:sdk_create_pre_build_request · ops: CreateTable · fields: metadata,
    metadata.iceberg, metadata.iceberg.schema, metadata.iceberg.schema.fields, metadata.iceberg.properties
  - repro: CreateTable(format=ICEBERG) with each of the five metadata shapes above
    (parameter_validation=False); GetTable afterwards
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-311](../table-schema.md#s3t-table-311), [S3T-TABLE-041](../table-schema.md#s3t-table-041), [S3T-TABLE-194](../table-schema.md#s3t-table-194), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) ·
    hypotheses: H-T-107, H-T-113 · evidence: table/weird-inputs/schema-variants

## Notes

H-T-107 REFUTED: fields=[] -> HTTP 200 and a committed metadata.json with schemas[0].fields=[] and
last-column-id=0 (table then deleted); schema={} (fields omitted on the wire) -> HTTP 500
InternalServerErrorException 'An internal error occurred. Try again.' with no table created - a controller
sending an empty schema struct would create a useless table or loop on a 5xx. H-T-113 'neither schema set'
part confirmed: iceberg={} and properties-only -> 400 'The specified metadata is not valid'. All 400 messages
are the same generic text. One attempt each.

Full original title: CreateTable schema.fields=[] is ACCEPTED (empty-schema table committed); schema={} ->
InternalServerErrorException 500; iceberg={}/metadata={}/properties-only -> BadRequestException 400
