<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-178: GetTableBucket and ListTableBuckets carry no policy member; a bucket policy is visible only through GetTableBucketPolicy (404 when absent)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-178"></a>**S3T-TABLEBUCKET-178** `read-gap` · impact medium · handled · verified 2026-10-14
  **GetTableBucket and ListTableBuckets carry no policy member; a bucket policy is visible only through GetTableBucketPolicy (404 when absent)**
  Raw GetTableBucket JSON keys as A: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'],
  as B: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'] (B -> HTTP 200);
  ListTableBuckets entry keys: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']. Keys
  containing 'polic': none. Whether a policy exists is learnt only from GetTableBucketPolicy
  (NotFoundException (HTTP 404) when absent).
  - ACK: custom_find, exceptions.404 · ops: GetTableBucket, ListTableBuckets, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: GetTableBucket(arn), ListTableBuckets(prefix) with and without a bucket policy;
    GetTableBucketPolicy when none is set
  - handling: handled via `generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-061](../tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065),
    [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLE-149](../table-policy.md#s3t-table-149), [S3T-TABLE-181](../table-policy.md#s3t-table-181), [S3T-TABLE-184](../table-policy.md#s3t-table-184) · evidence:
    tablebucketpolicy/sub-resources/fu-02-delete-policy-principal-b

## Notes

FU-02 (GetTableBucket body check). Single run; the key set is deterministic from the model (6 output members).

Full original title: GetTableBucket and ListTableBuckets carry no policy member (keys ['arn', 'createdAt',
'name', 'ownerAccountId', 'tableBucketId', 'type']); the policy is only visible through GetTableBucketPolicy
