<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-076: writeOrder field: direction/nullOrder omitted -> 400 (no default); 'ASC' lower-cased; 'NULLS_FIRST' -> 400; transform 'days' kept verbatim
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-076"></a>**S3T-TABLE-076** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **writeOrder field: direction/nullOrder omitted -> 400 (no default); 'ASC' lower-cased; 'NULLS_FIRST' -> 400; transform 'days' kept verbatim**
  With SDK validation disabled: direction omitted -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid'. nullOrder omitted -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  direction='ASC' -> HTTP 200 accepted -> sort-orders=[{"order-id": 1, "fields": [{"transform": "identity",
  "source-id": 2, "direction": "asc", "null-order": "nulls-first"}]}] default-sort-order-id=1.
  nullOrder='NULLS_FIRST' -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  transform='days' -> HTTP 200 accepted.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.writeOrder.fields.direction,
    metadata.iceberg.writeOrder.fields.nullOrder, metadata.iceberg.writeOrder.fields.transform
  - repro: CreateTable writeOrder with each variant (parameter_validation=False)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-075](../table-schema.md#s3t-table-075), [S3T-TABLE-077](../table-schema.md#s3t-table-077), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-018](../table-schema.md#s3t-table-018), [S3T-TABLE-038](../table-schema.md#s3t-table-038), [S3T-TABLE-073](../table-schema.md#s3t-table-073),
    [S3T-TABLE-042](../table-schema.md#s3t-table-042) · hypotheses: H-T-124 · evidence: table/weird-inputs/partition-sort-variants

## Notes

H-T-124 QUALIFIED: omitting direction or nullOrder -> BadRequestException 400 'The specified metadata is not
valid' (confirmed: no defaulting); 'NULLS_FIRST' -> 400 (confirmed); but 'ASC' is ACCEPTED and normalized to
'asc' in metadata.json (refutes the upper-case part for direction). Sort-field transform 'days' is accepted
and written verbatim (no grammar check, same as partitionSpec). One attempt each.

Full original title: writeOrder field: direction/nullOrder omitted -> 400 (no server default); 'ASC' accepted
and lower-cased to 'asc'; 'NULLS_FIRST' -> 400; sort transform 'days' persisted verbatim
