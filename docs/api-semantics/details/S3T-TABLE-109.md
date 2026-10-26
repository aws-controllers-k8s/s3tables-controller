<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-109: GetTable under an IAM explicit Deny or with no s3tables permissions -> AccessDeniedException 403 (IAM message), NOT ForbiddenException
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-109"></a>**S3T-TABLE-109** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable under an IAM explicit Deny or with no s3tables permissions -> AccessDeniedException 403 (IAM message), NOT ForbiddenException**
  Role with Allow s3tables:* plus explicit Deny s3tables:GetTable: get_composite -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34 with
  an explicit deny in an identity-based policy'; get_arn -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34 with
  an explicit deny in an identity-based policy'; get_nonexistent_composite -> AccessDeniedException (HTTP 403)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/* with an explicit deny in an
  identity-based policy'; get_table_bucket -> HTTP 200; list_tables -> HTTP 200; get_table_metadata_location
  -> HTTP 200; delete_nonexistent -> NotFoundException (HTTP 404) 'The specified table does not exist.'. Role
  with no s3tables permissions: get_composite -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTable action'; get_arn -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTable action'; get_nonexistent_composite ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/*
  because no identity-based policy allows the s3tables:GetTable action'; get_table_bucket ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam
  because no identity-based policy allows the s3tables:GetTableBucket action'; list_tables ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTables on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam
  because no identity-based policy allows the s3tables:ListTables action'; get_table_metadata_location ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTableMetadataLocation action'; delete_nonexistent ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:DeleteTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/* because no identity-based policy allows
  the s3tables:DeleteTable action'. Control (probe principal) GetTable -> HTTP 200. x-amzn-ErrorType header on
  the deny GetTable: AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/. The
  IAM-denied error differs from the service's own 403: header
  x-amzn-ErrorType=AccessDeniedException:...com.amazon.coral.service/ with body key 'Message' ('User:
  <principal> is not authorized to perform: s3tables:GetTable on resource: <table ARN> with an explicit deny
  in an identity-based policy' / '... because no identity-based policy allows the s3tables:GetTable action'),
  whereas a foreign-account ARN yields ForbiddenException 'Unauthorized' (header namespace
  com.amazon.s3tablebuckets, body key 'message'). With no permissions, DeleteTable of a nonexistent table ->
  AccessDeniedException (auth precedes the existence check); with the Deny-on-GetTable role it ->
  NotFoundException.
  - ACK: terminal_codes · ops: GetTable, GetTableBucket, ListTables, GetTableMetadataLocation, DeleteTable ·
    fields: tableArn, tableBucketARN, namespace, name
  - repro: CreateRole (trust: own account) with inline Allow s3tables:* + Deny s3tables:GetTable; AssumeRole;
    GetTable composite and by ARN; same with a role lacking s3tables permissions
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](../table-policy.md#s3t-table-026),
    [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLE-132](../table-policy.md#s3t-table-132), [S3T-TABLE-140](../table-policy.md#s3t-table-140), [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109),
    [S3T-TABLEBUCKET-067](../tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](../table-policy.md#s3t-table-083), [S3T-TABLE-120](../table.md#s3t-table-120) · hypotheses: H-T-044 · evidence:
    table/error-taxonomy/iam-deny-forbidden

## Notes

H-T-044 REFUTED: IAM denials (explicit Deny and no-Allow) surface as AccessDeniedException, not
ForbiddenException; ForbiddenException 'Unauthorized' is the code for a cross-account/foreign ARN (see the
identifier-forms and create-validation probes). A controller's 'insufficient permissions' classification must
therefore match AccessDeniedException, and ForbiddenException should be treated as 'wrong account/ARN'. Both
roles assumed on the first attempt; inline policies were given 10 s to propagate before the measured calls
(single pass).
