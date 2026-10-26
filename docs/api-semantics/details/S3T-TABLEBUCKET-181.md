<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-181: PutTableBucketEncryption(aws:kms) needs caller kms:DescribeKey, CreateTable needs kms:GenerateDataKey; denied -> ForbiddenException 403
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-181"></a>**S3T-TABLEBUCKET-181** `error-code` · impact medium · handled · verified 2026-10-14
  **PutTableBucketEncryption(aws:kms) needs caller kms:DescribeKey, CreateTable needs kms:GenerateDataKey; denied -> ForbiddenException 403**
  Assumed role, inline policy only, bucket already SSE-KMS with the same key. V1 (Allow s3tables:*; Deny
  kms:DescribeKey; no kms Allow): PutTableBucketEncryption -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' /
  ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' (12 s
  and 22 s after the policy write); CreateTable(schema) -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:GenerateDataKey on
  resource: arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a because no
  identity-base'; GetTableBucketEncryption -> HTTP 200. V2 (Allow s3tables:*, Allow kms:*, Deny
  kms:DescribeKey): Put -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' /
  ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ';
  CreateTable -> HTTP 200; Get -> HTTP 200. V3 (Allow s3tables:*, Allow kms:*): Put -> HTTP 200 / HTTP 200;
  CreateTable -> HTTP 200; Get -> HTTP 200.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, CreateTable, GetTableBucketEncryption ·
    fields: encryptionConfiguration.kmsKeyArn, encryptionConfiguration.sseAlgorithm
  - repro: iam CreateRole(trust: account) + PutRolePolicy(variant) -> sts AssumeRole ->
    PutTableBucketEncryption(aws:kms, key) x2, CreateTable(schema), GetTableBucketEncryption
  - measurements: n_put_per_variant=2, iam_propagation_wait_s=12, n=2
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLE-131](../table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-034](../tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035), [S3T-TABLEBUCKET-106](../tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](../table-policy.md#s3t-table-086)
    · evidence: tablebucket/sub-resources/fu-01-kms-key-policy-disable

## Notes

FU-01 part 2. The key policy at this point: root + maintenance.s3tables.amazonaws.com. Caller-side KMS
permission checks happen on the caller's identity (the controller's role), so the controller's IAM policy
needs whatever V1->V3 shows. Two attempts per variant guard against IAM propagation; see result.yaml
B_V1..B_V3 for messages. Summary: V1 and V2 Put -> ForbiddenException 403 'not authorized to perform:
kms:DescribeKey ... with an explicit deny in an identity-based policy' (the service calls kms:DescribeKey with
the caller's identity to validate the key); V1 CreateTable -> ForbiddenException 403 'not authorized to
perform: kms:GenerateDataKey ... because no identity-based policy allows the kms:GenerateDataKey action'; V2
CreateTable -> 200 (CreateTable does not call DescribeKey); GetTableBucketEncryption -> 200 under every
variant (reads need no KMS access). Wire code is ForbiddenException (modeled), not AccessDeniedException;
compare the resource-policy Deny case in FU-02 which returns AccessDeniedException.
