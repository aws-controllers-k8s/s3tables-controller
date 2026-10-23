<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-133: CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created; partitionSpec={} and writeOrder={orderId:1} -> 200
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-133"></a>**S3T-TABLE-133** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created; partitionSpec={} and writeOrder={orderId:1} -> 200**
  metadata.iceberg.schema={} (the SDK would reject this client-side; sent with validation disabled) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' /
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' /
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'. schemaV2={type:struct}
  without fields -> BadRequestException (HTTP 400) 'The specified metadata is not valid'. partitionSpec={} ->
  HTTP 200. writeOrder={orderId:1} without fields -> HTTP 200. No table was created by any 5xx/4xx attempt
  (nothing registered).
  - ACK: terminal_codes, requeue · ops: CreateTable · fields: metadata.iceberg.schema,
    metadata.iceberg.schemaV2.fields, metadata.iceberg.partitionSpec, metadata.iceberg.writeOrder.fields
  - repro: CreateTable(format=ICEBERG, metadata={iceberg:{schema:{}}}) with botocore
    parameter_validation=False, three times
  - measurements: attempts=3, http_500=3, n=3
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-036](../table-schema.md#s3t-table-036), [S3T-TABLE-311](../table-schema.md#s3t-table-311), [S3T-TABLE-041](../table-schema.md#s3t-table-041), [S3T-TABLE-194](../table-schema.md#s3t-table-194), [S3T-TABLE-072](../table-schema.md#s3t-table-072), [S3T-TABLE-073](../table-schema.md#s3t-table-073),
    [S3T-TABLE-074](../table-schema.md#s3t-table-074), [S3T-TABLE-075](../table-schema.md#s3t-table-075), [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-077](../table-schema.md#s3t-table-077), [S3T-TABLE-018](../table-schema.md#s3t-table-018), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) ·
    hypotheses: H-T-107 · evidence: table/creative/token-and-suffix-followup

## Notes

Follow-up to the schema-variants probe: 3 of 3 fresh attempts (plus the earlier one) returned
InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' and no table was created;
schemaV2={type:struct} without fields -> BadRequestException 400 'The specified metadata is not valid';
partitionSpec={} and writeOrder={orderId:1} (fields omitted) -> HTTP 200 (tables created and deleted). A
controller whose generated code emits an empty v1 schema struct would see a retryable-looking 5xx forever; the
SDK's client-side required-member check is the only guard. n=3 (+1).

Full original title: CreateTable schema={} (fields omitted on the wire) -> InternalServerErrorException 500
reproducibly (4/4); partitionSpec={} and writeOrder={orderId:1} without fields are ACCEPTED

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. A Table CR whose
metadata.iceberg.schema has no fields never reaches the terminal BadRequestException the controller expects:
schema={} on the wire -> InternalServerErrorException 500 reproducibly (4/4, no table created; [S3T-TABLE-133](../table-schema.md#s3t-table-133)),
so the reconciler requeues forever, and schema.fields=[] -> HTTP 200 creating an empty-schema table reported
as Synced ([S3T-TABLE-036](../table-schema.md#s3t-table-036)); add CRD validation (minItems=1 on schema.fields, or require exactly one of
schema/schemaV2) or a pre-build guard.
