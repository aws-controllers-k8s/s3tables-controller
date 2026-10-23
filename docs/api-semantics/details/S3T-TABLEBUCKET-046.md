<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-046: IAM explicit Deny on the seven bucket-core operations yields AccessDeniedException/403 on every one...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-046"></a>**S3T-TABLEBUCKET-046** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM explicit Deny on the seven bucket-core operations yields AccessDeniedException/403 on every one...**
  Role = AdministratorAccess + inline Deny on ['s3tables:GetTableBucket', 's3tables:ListTableBuckets',
  's3tables:CreateTableBucket', 's3tables:DeleteTableBucket', 's3tables:TagResource',
  's3tables:UntagResource', 's3tables:ListTagsForResource']. As that role: GetTableBucket ->
  AccessDeniedException/403; ListTableBuckets -> AccessDeniedException/403; CreateTableBucket ->
  AccessDeniedException/403; DeleteTableBucket -> AccessDeniedException/403; TagResource ->
  AccessDeniedException/403; UntagResource -> AccessDeniedException/403; ListTagsForResource ->
  AccessDeniedException/403; GetTableBucketEncryption (not denied) -> 200. Messages: ['User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:CreateTableBucket on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-denied with an explicit deny in an
  identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucket on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with
  an explicit deny in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:GetTableBucket on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:ListTableBuckets on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/* with an explicit deny
  in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTagsForResource on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:TagResource on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an
  explicit deny in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:UntagResource on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy'].
  - ACK: terminal_codes · ops: GetTableBucket, ListTableBuckets, CreateTableBucket, DeleteTableBucket,
    TagResource, UntagResource, ListTagsForResource
  - repro: IAM role with AdministratorAccess + Deny on the seven ops; AssumeRole; call each op against an
    existing bucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLE-109](../table-policy.md#s3t-table-109),
    [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131) · hypotheses:
    H-B-036, H-B-037 · evidence: tablebucket/error-taxonomy/iam-deny-codes

## Notes

H-B-036 (AccessDeniedException for Get/List, ForbiddenException for the rest): refuted. H-B-037
(ForbiddenException everywhere for IAM denies): refuted/qualified: ['AccessDeniedException/403'].

Full original title: IAM explicit Deny on the seven bucket-core operations yields AccessDeniedException/403 on
every one (no AccessDenied/Forbidden split by operation)

Contradiction with [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005): 005's behavior text calls ForbiddenException the code 'seen for
IAM-level/foreign-account denials'; 046 (seven ops), 045, 091 and 129 show every IAM-level denial
(identity-policy explicit Deny and implicit deny) as AccessDeniedException 403 - ForbiddenException belongs
only to foreign-account ARNs (038) and downstream KMS/PutTableData checks (181, [S3T-TABLE-131](../table-policy.md#s3t-table-131)) Resolution:
keep both; [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046) is canonical for the IAM-level code; 005's remark conflates IAM-level with
foreign-account/KMS ForbiddenException
