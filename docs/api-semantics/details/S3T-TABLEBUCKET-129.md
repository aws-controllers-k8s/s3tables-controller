<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-129: Role with only Create/Get/DeleteTableBucket - plain CreateTableBucket 200; with tags, encryption or storageClass -> AccessDeniedException
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-129"></a>**S3T-TABLEBUCKET-129** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Role with only Create/Get/DeleteTableBucket - plain CreateTableBucket 200; with tags, encryption or storageClass -> AccessDeniedException**
  Role policy = Allow s3tables:CreateTableBucket, GetTableBucket, DeleteTableBucket on * (nothing else).
  CreateTableBucket(name) -> HTTP 200 (bucket exists; DeleteTableBucket as the role -> 204).
  CreateTableBucket(name, tags={a: b}) -> AccessDeniedException 403 '... is not authorized to perform:
  s3tables:TagResource on resource: <bucket ARN of the requested name> because no identity-based policy allows
  the s3tables:TagResource action'. CreateTableBucket(name, encryptionConfiguration={sseAlgorithm: AES256}) ->
  AccessDeniedException 403 naming s3tables:PutTableBucketEncryption (even though AES256 is the default).
  CreateTableBucket(name, storageClassConfiguration={storageClass: STANDARD}) -> AccessDeniedException 403
  naming s3tables:PutTableBucketStorageClass. CreateTableBucket(name, tags + encryptionConfiguration) ->
  AccessDeniedException 403 naming s3tables:TagResource (tags checked first). After every 403 the admin's
  GetTableBucket on the constructed ARN -> NotFoundException 404 (no bucket created). The same role's implicit
  deny on ListTagsForResource / TagResource -> AccessDeniedException 403 'because no identity-based policy
  allows the ... action'.
  - ACK: docs-only, terminal_codes · ops: CreateTableBucket, TagResource · fields: tags,
    encryptionConfiguration, storageClassConfiguration
  - repro: IAM role with Allow s3tables:CreateTableBucket/GetTableBucket/DeleteTableBucket only;
    CreateTableBucket with and without tags / encryptionConfiguration(AES256) /
    storageClassConfiguration(STANDARD)
  - measurements: create_table_bucket_latency_ms_allowed=1628, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLE-140](../table-policy.md#s3t-table-140),
    [S3T-TABLE-131](../table-policy.md#s3t-table-131) · hypotheses: H-B-039 · evidence: table/error-taxonomy/iam-deny-rename-tag

## Notes

H-B-039 qualified: the permission requirement is confirmed (tags need s3tables:TagResource;
encryptionConfiguration needs PutTableBucketEncryption; storageClassConfiguration needs
PutTableBucketStorageClass; no bucket is created on the 403), but the code is AccessDeniedException 403
(unmodeled for CreateTableBucket), not the modeled ForbiddenException the hypothesis predicted. The message
names the missing action, so it is actionable. Single run.

Full original title: Role allowed only Create/Get/DeleteTableBucket: CreateTableBucket plain -> 200; with
tags, encryptionConfiguration or storageClassConfiguration -> AccessDeniedException 403, nothing created
