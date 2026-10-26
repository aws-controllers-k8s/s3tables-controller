<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-040: CreateNamespace 429 after ~16 back-to-back is per account-region (all buckets), not shared with Delete/CreateTable/Tag; <= 4/s is safe
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-040"></a>**S3T-NAMESPACE-040** `quota-limit` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **CreateNamespace 429 after ~16 back-to-back is per account-region (all buckets), not shared with Delete/CreateTable/Tag; <= 4/s is safe**
  CreateNamespace back-to-back (~27-30 calls/s) in one table bucket: 16 (sometimes 17) x 200 then
  TooManyRequestsException HTTP 429 'Too Many Requests' within ~0.55 s; the 429 carries no Retry-After header
  (only x-amzn-ErrorType and Date). Immediately after the 429, 10 back-to-back calls alternating X with
  CreateNamespace(A) ('.'=200, 'T'=429), 3 clean runs: X=CreateNamespace in ANOTHER table bucket of the same
  region -> X TTTTT A TTTTT in 4 of 6 trials (both '.....' in the other 2, where A had recovered too): the
  limit is shared by all table buckets of the account-region. X=CreateNamespace in the other region -> X
  '.....' in 3/3 (A 'TTTT.' in one): separate limiter. X=DeleteNamespace -> X '.....' in 3/3 while A
  'TTTTT'/'TTTTT'/'TTT..': deletes are NOT counted against it. X=CreateTable -> X '...' in 3/3 (A 'TTT' once);
  X=TagResource(bucket) -> X '...' in 3/3 (A 'TTT' twice): separate. After a drain, 13 CreateNamespace at 4/s
  -> 12/13 ok in 3/3 runs (only the first call, inside the same window, was rejected); in an earlier run 90
  CreateNamespace paced at <= 5/s and 16 at 4/s gave 0 x 429. The recovery pattern (whole batches passing or
  failing) looks like a fixed window of ~16 per second rather than a smooth token bucket.
  - ACK: requeue, one-per-reconcile, terminal_codes · ops: CreateNamespace, DeleteNamespace, CreateTable,
    TagResource · fields: namespace
  - repro: CreateNamespace back-to-back in A until 429, then 10 back-to-back calls alternating <other op or
    bucket> with CreateNamespace(A); repeat per X after 12 s idle; drain then 4/s
  - measurements: drain_ok_before_429=[16, 16, 16, 16, 16, 17, 16, 16, 16, 16, 17, 16, 17, 22, 11],
    same_region_other_bucket_X_patterns=[TTTTT, TTTTT, TTTTT, TTTTT, ....., .....],
    other_region_X_patterns=[....., ....., .....], delete_namespace_X_patterns=[....., ....., .....],
    create_table_X_patterns=[..., ..., ...], tag_bucket_X_patterns=[..., ..., ...],
    refill_ok_of_13_at_4_per_s=[12, 12, 12], paced_le_5_per_s_ok=106, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-NAMESPACE-035](../namespace.md#s3t-namespace-035) · evidence: namespace/creative/throttle-key

## Notes

extends [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033)/-035 with the KEY of the limit and the pacing rule: the budget is per account-region
and shared by every table bucket, so N parallel workers creating Namespace CRs in different buckets still
compete; DeleteNamespace (contrary to the '-033' wording that bundles create/delete), CreateTable and
TagResource do not consume it. Decision rule: pace CreateNamespace to <= 4/s per account-region across all
workers; treat TooManyRequestsException (no Retry-After) as requeue with a ~1 s backoff, never terminal. Other
probers share the account (one drain showed only 11 successes); 3 clean runs plus one partial.

Full original title: CreateNamespace 429 after 16 back-to-back is per account-region (both buckets throttled),
not shared with Delete/CreateTable/Tag; 4/s is safe

Contradiction with [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-NAMESPACE-035](../namespace.md#s3t-namespace-035): limiter model disagrees: 033/035 describe a
per-operation token bucket with a slow refill (035 notes infer ~2 tokens/s from one retry that succeeded at
0.54 s, scope 'unknown'); 040 (3 clean runs) observes 12/13 ok at 4/s, 106/106 paced <= 5/s and whole batches
passing or failing, i.e. a ~16-per-second fixed window shared by every table bucket of the account-region - a
2 tokens/s refill would have throttled the 4/s pacing Resolution: keep both; [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040) is canonical
for the limiter's key, shape and safe pacing
