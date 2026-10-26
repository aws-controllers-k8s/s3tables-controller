<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-121: Hypothesis of a transient 404 after CreateTableBucket refuted: Get, List(prefix) and ListTags all 200 on poll #1 in 5/5, no flapping
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-121"></a>**S3T-TABLEBUCKET-121** `eventual-consistency` · impact medium · handled · verified 2026-10-14
  **Hypothesis of a transient 404 after CreateTableBucket refuted: Get, List(prefix) and ListTags all 200 on poll #1 in 5/5, no flapping** (hypothesis refuted; behavior confirmed)
  5 fresh buckets (CreateTableBucket latency 1454-1678 ms, tags={seed:1}); right after Create returned,
  GetTableBucket, ListTableBuckets(prefix=<name>) and ListTagsForResource(arn) were polled in lockstep every
  100 ms until all agreed for 5 polls. First-success offsets since Create returned (ms): Get [0, 0, 0, 0, 0],
  List(prefix) [0, 0, 0, 0, 0] (bucket present), ListTags [0, 0, 0, 0, 0] (seed tag present). Codes seen
  before the first success: none. Successes followed by a failure (flapping): none.
  - ACK: none · ops: CreateTableBucket, GetTableBucket, ListTableBuckets, ListTagsForResource · fields:
    tableBucketARN, prefix, tags
  - repro: CreateTableBucket(tags) -> poll GetTableBucket + ListTableBuckets(prefix) + ListTagsForResource
    every 100 ms for 5 agreeing polls; x5
  - measurements: first_ok_ms_get=[0, 0, 0, 0, 0], first_ok_ms_list=[0, 0, 0, 0, 0], first_ok_ms_tags=[0, 0,
    0, 0, 0], create_latency_ms=[1678, 1454, 1540, 1458, 1538], polls_per_trial=[5, 5, 5, 5, 5],
    poll_interval_s=0.1, n=5
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-004](../tablebucket.md#s3t-tablebucket-004), [S3T-TABLE-128](../table.md#s3t-table-128), [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](../tablebucket.md#s3t-tablebucket-122),
    [S3T-TABLEBUCKET-156](../tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](../tablebucket.md#s3t-tablebucket-157), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-127](../table.md#s3t-table-127), [S3T-TABLEBUCKET-028](../tablebucket.md#s3t-tablebucket-028),
    [S3T-TABLEBUCKET-007](../tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-160](../tablebucket.md#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](../tablebucket.md#s3t-tablebucket-167), [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-NAMESPACE-019](../namespace.md#s3t-namespace-019) · hypotheses: H-B-003, H-B-046 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5

## Notes

H-B-003 (contrarian: transient 404 within 1-2 s after Create) refuted for GetTableBucket and extended to
ListTableBuckets(prefix) and ListTagsForResource: every read path was correct on poll #1. H-B-046
create-time-tags part confirmed: tags passed to CreateTableBucket were present on the first
ListTagsForResource in 5/5. Brief n=5 confirmation of [S3T-TABLEBUCKET-004](../tablebucket.md#s3t-tablebucket-004) (Get) at 100 ms.

Full original title: GetTableBucket, ListTableBuckets(prefix) and ListTagsForResource all 200 on the first
poll after CreateTableBucket (5/5, no 404 window, no flapping)
