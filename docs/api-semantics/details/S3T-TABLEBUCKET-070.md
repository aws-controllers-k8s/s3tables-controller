<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-070: 10 concurrent TagResource on one bucket -> 8-9/10 InternalServerErrorException 500 (never 409); some 500-failed calls still applied the tag
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-070"></a>**S3T-TABLEBUCKET-070** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **10 concurrent TagResource on one bucket -> 8-9/10 InternalServerErrorException 500 (never 409); some 500-failed calls still applied the tag**
  Two bursts of 10 concurrent TagResource calls (one distinct key each) on one table bucket: run 1
  {"InternalServerErrorException/500": 9, "200": 1} (wall 0.63s), run 2 {"InternalServerErrorException/500":
  8, "200": 2} (wall 0.18s); ListTagsForResource afterwards showed 4/1 and 4/2 of the successfully-tagged
  keys. Rejected calls retried sequentially: [('k00', 'InternalServerErrorException', 'HTTP 200'), ('k01',
  'InternalServerErrorException', 'HTTP 200'), ('k02', 'InternalServerErrorException', 'HTTP 200'), ('k03',
  'InternalServerErrorException', 'HTTP 200'), ('k04', 'InternalServerErrorException', 'HTTP 200'), ('k05',
  'InternalServerErrorException', 'HTTP 200'), ('k07', 'InternalServerErrorException', 'HTTP 200'), ('k08',
  'InternalServerErrorException', 'HTTP 200'), ('k09', 'InternalServerErrorException', 'HTTP 200')]. Duplicate
  CreateTableBucket(existing name) -> ConflictException (HTTP 409) 'The bucket that you tried to create
  already exists, and you own it.'. DeleteTableBucket(bucket with one namespace) -> BadRequestException (HTTP 400)
  'The bucket that you tried to delete is not empty.'.
  - ACK: tags.custom-sync, requeue, one-per-reconcile · ops: TagResource, CreateTableBucket,
    DeleteTableBucket, ListTagsForResource · fields: tags
  - repro: CreateTableBucket; 10 threads x TagResource(arn, {k_i: v}); ListTagsForResource;
    CreateTableBucket(same name); CreateNamespace; DeleteTableBucket
  - measurements: tag_burst_wall_s=[0.63, 0.18], tag_latency_ms_min_max=[[50,631], [29,178]], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-159](../tablebucket.md#s3t-tablebucket-159), [S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017), [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKET-019](../tablebucket.md#s3t-tablebucket-019), [S3T-TABLEBUCKET-044](../tablebucket.md#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](../tablebucket.md#s3t-tablebucket-132), [S3T-NAMESPACE-024](../namespace.md#s3t-namespace-024), [S3T-TABLEBUCKET-033](../tablebucket.md#s3t-tablebucket-033),
    [S3T-TABLEBUCKET-043](../tablebucket.md#s3t-tablebucket-043), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015) · hypotheses: H-B-017 · evidence:
    tablebucket/limits/bucket-quota-create-burst

## Notes

H-B-017 refuted for the concurrent-write half: concurrent writes to one bucket's tags surface as
InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' and never as
ConflictException. Run 1: 1 call returned 200 but ListTagsForResource showed 4 keys (k00,k04,k05 were applied
by calls that got a 500); run 2: 2 calls returned 200 but 4 keys were present. So a 500 from TagResource is
NOT a clean failure - the write may have landed. All 9 rejected calls succeeded when retried sequentially. The
two terminal 'conflict' cases use different codes: duplicate CreateTableBucket -> ConflictException 409 'The
bucket that you tried to create already exists, and you own it.'; DeleteTableBucket of a non-empty bucket ->
BadRequestException 400 'The bucket that you tried to delete is not empty.' (not 409). Shared account, 2 other
probes running; both bursts gave the same picture.

Full original title: 10 concurrent TagResource on one table bucket -> 8-9/10 InternalServerErrorException 500
(never 409); some 500-failed calls still applied their tag
