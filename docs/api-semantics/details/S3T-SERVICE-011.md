<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-011: Permissions-boundary deny -> AccessDeniedException 403 on Get/List/Create; ARN region/name/syntax BadRequestException 400 precede authz
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-011"></a>**S3T-SERVICE-011** `error-code` · impact medium · handled · verified 2026-10-14
  **Permissions-boundary deny -> AccessDeniedException 403 on Get/List/Create; ARN region/name/syntax BadRequestException 400 precede authz**
  Role with inline Allow s3tables:* but PermissionsBoundary=AmazonS3ReadOnlyAccess (no s3tables):
  GetTableBucket(existing) -> AccessDeniedException/403 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL>
  is not authorized to perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:'; GetTableBucket(ghost)
  -> AccessDeniedException/403 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:'; GetTableBucket(other-region ARN) ->
  BadRequestException/400 'The provided Region does not match the Region of the resource.';
  GetTableBucket(invalid name in ARN) -> BadRequestException/400 "The specified bucket name isn't valid.
  Specify a different bucket name, and then try again."; GetTableBucket(not an ARN) -> BadRequestException/400
  'The specified table bucket ARN is not valid.'; ListTableBuckets -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:ListTableBuckets on
  resource: arn:aws:s3table'; GetTable(tableArn) -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-wes'; CreateNamespace -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:CreateNamespace on
  resource: arn:aws:s3tables'.
  - ACK: terminal_codes, requeue · ops: GetTableBucket, ListTableBuckets, GetTable, CreateNamespace · fields:
    tableBucketARN
  - repro: CreateRole(PermissionsBoundary=AmazonS3ReadOnlyAccess) + inline Allow s3tables:*; AssumeRole;
    GetTableBucket with existing/ghost/other-region/invalid/non-ARN values
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-024](../service.md#s3t-service-024), [S3T-SERVICE-027](../service.md#s3t-service-027), [S3T-SERVICE-012](../service.md#s3t-service-012), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127),
    [S3T-TABLEBUCKET-022](../tablebucket.md#s3t-tablebucket-022) · evidence: service/creative/iam-authz-variants

## Notes

Creative round, seed error-messages: fills the AccessDeniedException message-variant gap (boundary) and fixes
the check order for the catalogue: which of ARN-syntax 400 / region 400 / authz 403 / lookup 404 wins is in
behavior. AUTHZ class regardless of the reason suffix.

Full original title: Permissions-boundary deny: AccessDeniedException/403 'no permissions boundary allows the
s3tables:G'; under deny GetTableBucket ghost -> AccessDeniedException, other-region -> BadRequestException,
invalid name -> BadRequestException, not-ARN -> BadRequestException
