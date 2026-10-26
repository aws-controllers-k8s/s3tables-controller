<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-180: A table ARN in the tableBucketARN slot -> BadRequestException/400 for GetTablePolicy (all 6 ops rejected)...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-180"></a>**S3T-TABLE-180** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **A table ARN in the tableBucketARN slot -> BadRequestException/400 for GetTablePolicy (all 6 ops rejected)...**
  tableBucketARN=<table ARN> with the correct namespace/name: GetTablePolicy -> BadRequestException/400 'The
  specified bucket name isn't valid. Specify a different bucket name, and then try again.'; DeleteTablePolicy
  -> BadRequestException/400 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; GetTableMaintenanceConfiguration -> BadRequestException/400 'The specified bucket name isn't
  valid. Specify a different bucket name, and then try again.'; GetTableMaintenanceJobStatus ->
  BadRequestException/400 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; PutTablePolicy -> BadRequestException/400 'The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.'; PutTableMaintenanceConfiguration -> BadRequestException/400
  'The specified bucket name isn't valid. Specify a different bucket name, and then try again.'. With a wrong
  namespace: GetTablePolicy -> BadRequestException/400 'The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.'. With a wrong name: BadRequestException/400. Passing tableArn=
  to get_table_policy raises ParamValidationError client-side (the model has no such member);
  GetTable(tableArn) -> HTTP 200. Policy visible via the triple after the table-ARN-form Put: HTTP 200.
  - ACK: is_arn_primary_key, custom_find · ops: GetTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, PutTablePolicy,
    PutTableMaintenanceConfiguration, GetTable · fields: tableBucketARN, tableArn
  - repro: GetTablePolicy etc. with tableBucketARN=<tableARN>; get_table_policy(tableArn=...)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-178](../table-maintenance.md#s3t-table-178), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-182](../table-maintenance.md#s3t-table-182), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183) · hypotheses: H-U-023, H-U-024,
    H-U-071 · evidence: table/identity/subresource-identifiers

## Notes

H-U-023 confirmed, H-U-024 refuted; H-U-071 confirmed at the model level (policy/maintenance ops have only
tableBucketARN+namespace+name; GetTable also takes tableArn).

Full original title: A table ARN in the tableBucketARN slot -> BadRequestException/400 for GetTablePolicy (all
6 ops rejected); tableArn kwarg -> ParamValidationError (client-side)
