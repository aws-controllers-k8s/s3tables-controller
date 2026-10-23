<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-003: CreateTableBucket returns only {arn}; ARN is arn:aws:s3tables:<region>:<account>:bucket/<name>, constructible before the call
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-003"></a>**S3T-TABLEBUCKET-003** `identity` · impact high · handled · verified 2026-10-15, re-verified
  **CreateTableBucket returns only {arn}; ARN is arn:aws:s3tables:<region>:<account>:bucket/<name>, constructible before the call**
  In 5/5 creates the CreateTableBucket response body (HTTP 200) had exactly the keys ['arn'] and the arn
  equalled the string built in advance from (region, caller account, name) with resource segment
  'bucket/<name>' (no suffix, no trailing slash). GetTableBucket(tableBucketARN=<constructed arn>) returned
  HTTP 200 with keys ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'], and its arn was
  byte-identical to the Create response arn in 5/5 runs.
  - ACK: is_arn_primary_key, custom_find · ops: CreateTableBucket, GetTableBucket · fields: arn, name,
    tableBucketARN
  - repro: expected='arn:aws:s3tables:<region>:<account>:bucket/<name>'; CreateTableBucket(name) -> compare
    arn; GetTableBucket(expected)
  - measurements: create_latency_ms=[1580, 1556, 1477, 1470, 1514], n=5
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98`
  - related: [S3T-TABLEBUCKET-004](../tablebucket.md#s3t-tablebucket-004), [S3T-TABLEBUCKET-121](../tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-122](../tablebucket.md#s3t-tablebucket-122), [S3T-TABLEBUCKET-022](../tablebucket.md#s3t-tablebucket-022),
    [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](../tablebucket.md#s3t-tablebucket-033), [S3T-TABLEBUCKET-160](../tablebucket.md#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](../tablebucket.md#s3t-tablebucket-167),
    [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-NAMESPACE-019](../namespace.md#s3t-namespace-019) · hypotheses: H-B-001, H-B-066 · evidence:
    tablebucket/round-trip/create-identity-defaults

## Notes

Confirms H-B-001 (deterministic ARN) and H-B-066 (Create output carries only arn; Status fields other than arn
must come from an immediate GetTableBucket). The same arn string was accepted verbatim as tableBucketARN by
GetTableBucket and DeleteTableBucket (see later steps for resourceArn on the tag operations).
