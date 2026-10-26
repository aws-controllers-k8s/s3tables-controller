<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-024: GetTable with no identifier -> BadRequestException; partial composite key (bucket+name, bucket+namespace...
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-024"></a>**S3T-TABLE-024** `request-validation` · impact medium · handled · verified 2026-10-13
  **GetTable with no identifier -> BadRequestException; partial composite key (bucket+name, bucket+namespace...**
  All four GetTable input members are optional in the model and boto3 sends the request. no params ->
  BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table bucket
  arn, namespace, table name) together'; namespace only -> BadRequestException (HTTP 400) 'Invalid input: Must
  provide either table arn alone OR all of (table bucket arn, namespace, table name) together';
  tableBucketARN+name (no namespace) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either
  table arn alone OR all of (table bucket arn, namespace, table name) together'; tableBucketARN+namespace (no
  name) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table
  bucket arn, namespace, table name) together'; tableBucketARN only -> BadRequestException (HTTP 400) 'Invalid
  input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name) together';
  namespace+name (no bucket) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn
  alone OR all of (table bucket arn, namespace, table name) together'.
  - ACK: custom_find, terminal_codes · ops: GetTable · fields: tableArn, tableBucketARN, namespace, name
  - repro: GetTable() / GetTable(namespace=ns) / GetTable(tableBucketARN, name) / GetTable(tableBucketARN,
    namespace) / GetTable(tableBucketARN)
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-269](../table-metadata-location.md#s3t-table-269), [S3T-TABLE-270](../table.md#s3t-table-270), [S3T-TABLE-023](../table-metadata-location.md#s3t-table-023), [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001), [S3T-TABLE-025](../table.md#s3t-table-025), [S3T-TABLE-267](../table.md#s3t-table-267) ·
    hypotheses: H-T-003, H-T-006, H-U-190 · evidence: table/identity/get-table-identifier-forms

## Notes

H-T-003 confirmed (no client-side validation, server BadRequestException (HTTP 400) 'Invalid input: Must
provide either table arn alone OR all of (table bucket arn, namespace, table name) together'). H-T-006
confirmed (partial key -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone
OR all of (table bucket arn, namespace, table name) together', no namespace-wide fallback search). H-U-190 (no
identifier) confirmed.

Full original title: GetTable with no identifier -> BadRequestException; partial composite key (bucket+name,
bucket+namespace, bucket only) -> BadRequestException
