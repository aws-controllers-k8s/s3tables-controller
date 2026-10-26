<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-325: CreateTable with metadata.iceberg.schema needs s3tables:PutTableData: same policy + it -> 200, without -> ForbiddenException 403
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-325"></a>**S3T-TABLE-325** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable with metadata.iceberg.schema needs s3tables:PutTableData: same policy + it -> 200, without -> ForbiddenException 403**
  Three runs, identical. Two assumed roles, both Allow s3tables:CreateTable/GetTable/DeleteTable on
  <bucketARN> and <bucketARN>/*. Without PutTableData: CreateTable(no metadata) -> HTTP 200 at the first
  attempt; CreateTable(metadata.iceberg.schema) -> ForbiddenException (HTTP 403) 'Access Denied' (402-449 ms),
  retry -> the same; admin GetTable on that name -> NotFoundException 404 and admin ListTables does not list
  it; admin CreateTable of the same name/metadata -> HTTP 200. Properties-only metadata
  ({iceberg:{properties:{k:v}}}) by that role -> BadRequestException (HTTP 400) 'The specified metadata is not
  valid' (7-8 ms, before any authorization). With s3tables:PutTableData added: CreateTable(no metadata) ->
  HTTP 200; CreateTable(metadata.iceberg.schema) -> HTTP 200 (378-486 ms) and admin GetTableMetadataLocation
  shows a metadataLocation.
  - ACK: docs-only, terminal_codes · ops: CreateTable · fields: metadata.iceberg.schema
  - repro: Two IAM roles (with/without s3tables:PutTableData) -> AssumeRole -> CreateTable with and without
    metadata.iceberg.schema
  - measurements: assume_role_first_attempt_ok_s=0.0, create_with_metadata_403_latency_ms=[449, 422, 402],
    create_with_metadata_200_latency_ms=[404, 401, 486], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-131](../table-policy.md#s3t-table-131) · evidence: table/creative/reverify-set-c

## Notes

Confirms the causal clause in the [S3T-TABLE-131](../table-policy.md#s3t-table-131) title (which only tested the role lacking the action) with a
positive control: the controller's IAM policy must include s3tables:PutTableData for Table specs that carry
metadata.iceberg.schema; the error it gets otherwise is ForbiddenException 'Access Denied' (not the IAM
AccessDeniedException form), and nothing is created.
