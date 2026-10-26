<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-179: tableBucketARN of another account -> ForbiddenException/403; other region in ARN -> BadRequestException/400...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-179"></a>**S3T-TABLE-179** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **tableBucketARN of another account -> ForbiddenException/403; other region in ARN -> BadRequestException/400...**
  GetTablePolicy with a well-formed ARN whose account id is not the caller's -> ForbiddenException/403
  'Unauthorized'; GetTableMaintenanceConfiguration same ARN -> ForbiddenException/403 'Unauthorized'. ARN with
  region us-east-1 sent to the us-west-2 endpoint -> BadRequestException/400 'The provided Region does not
  match the Region of the resource.'. ARN missing the bucket name ('...:bucket') -> BadRequestException/400
  'The specified table bucket ARN is not valid.'. A bare bucket name instead of an ARN ->
  BadRequestException/400 'The specified table bucket ARN is not valid.'. Nonexistent bucket in the caller's
  account -> NotFoundException/404 'The specified table does not exist.'.
  - ACK: terminal_codes, exceptions.404 · ops: GetTablePolicy, GetTableMaintenanceConfiguration · fields:
    tableBucketARN
  - repro: GetTablePolicy / GetTableMaintenanceConfiguration with each ARN variant
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-126](../tablebucket-encryption-storage.md#s3t-tablebucket-126), [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-243](../table-record-expiration.md#s3t-table-243),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190), [S3T-TABLE-178](../table-maintenance.md#s3t-table-178), [S3T-TABLE-180](../table-maintenance.md#s3t-table-180), [S3T-TABLE-182](../table-maintenance.md#s3t-table-182),
    [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-141](../table-policy.md#s3t-table-141), [S3T-TABLE-216](../table-policy.md#s3t-table-216) · hypotheses: H-U-018 · evidence:
    table/identity/subresource-identifiers

## Notes

H-U-018 confirmed.

Full original title: tableBucketARN of another account -> ForbiddenException/403; other region in ARN ->
BadRequestException/400; nonexistent own bucket -> NotFoundException/404 (GetTablePolicy)
