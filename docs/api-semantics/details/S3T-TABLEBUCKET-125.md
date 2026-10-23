<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-125: Same bucket name in us-west-2 and us-east-1 -> both Create 200, ARNs differ only in region, distinct tableBucketIds, List/tags per region
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-125"></a>**S3T-TABLEBUCKET-125** `cross-region` · impact high · handled · verified 2026-10-14
  **Same bucket name in us-west-2 and us-east-1 -> both Create 200, ARNs differ only in region, distinct tableBucketIds, List/tags per region**
  CreateTableBucket(name=X) in us-west-2 then the same name in us-east-1 -> both HTTP 200 (latency 1911 / 1253
  ms); the ARNs differ only in ARN segment(s) [3] (index 3 = region); GetTableBucket per region: same
  tableBucketId = False. ListTableBuckets(prefix=X) on the us-west-2 endpoint -> [('ackq-5443b8-xr',
  'us-west-2')]; on the us-east-1 endpoint -> [('ackq-5443b8-xr', 'us-east-1')] (each region lists only its
  own bucket). Tags are independent: us-west-2 tags {'ackq-run': '5443b8', 'side': 'A'}, us-east-1 tags
  {'ackq-run': '5443b8', 'side': 'B'}.
  - ACK: is_arn_primary_key, list_operation.match_fields · ops: CreateTableBucket, GetTableBucket,
    ListTableBuckets, ListTagsForResource · fields: name, arn, tableBucketId
  - repro: CreateTableBucket(X) in us-west-2; CreateTableBucket(X) in us-east-1; GetTableBucket each;
    ListTableBuckets(prefix=X) each; ListTagsForResource each
  - measurements: create_latency_ms=[1911, 1253], n=2
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155`
  - related: [S3T-TABLEBUCKET-040](../tablebucket.md#s3t-tablebucket-040), [S3T-TABLEBUCKET-041](../tablebucket.md#s3t-tablebucket-041), [S3T-TABLEBUCKET-069](../tablebucket.md#s3t-tablebucket-069), [S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130),
    [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-159](../tablebucket.md#s3t-tablebucket-159) · hypotheses: H-B-062 · evidence:
    tablebucket/cross-region/same-name-two-regions-arn-routing

## Notes

H-B-062 (per-account AND per-region name scope) confirmed for the create/identity part; the routing part is in
the companion finding. A controller must pin each CR to one region: the same spec name in two regions is two
unrelated buckets.

Full original title: Same table bucket name in us-west-2 and us-east-1: both CreateTableBucket 200; ARNs
differ only in the region segment; distinct tableBucketIds; List/tags scoped per region
