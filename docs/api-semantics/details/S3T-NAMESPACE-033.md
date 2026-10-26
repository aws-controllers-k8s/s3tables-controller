<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-033: Back-to-back CreateNamespace or DeleteNamespace (~30/s): 16 of 30 succeed, rest 429 'Too Many Requests' (1 burst each); 30 parallel Gets 200
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-033"></a>**S3T-NAMESPACE-033** `quota-limit` · impact high · handled · verified 2026-10-13
  **Back-to-back CreateNamespace or DeleteNamespace (~30/s): 16 of 30 succeed, rest 429 'Too Many Requests' (1 burst each); 30 parallel Gets 200**
  30 GetNamespace calls fired concurrently (20 threads) on one namespace: run 1 {"200": 30} (wall 0.16s), run
  2 after 20 s {"200": 30} (wall 0.09s). 30 CreateNamespace for distinct fresh names sent sequentially with no
  pause: {"200": 16, "TooManyRequestsException/429": 14} in 0.57s (latency [2, 37] ms); 30 DeleteNamespace
  likewise: {"204": 16, "TooManyRequestsException/429": 14} in 0.51s (latency [2, 52] ms). Error messages:
  {"TooManyRequestsException": "Too Many Requests"}; Retry-After headers: [].
  - ACK: requeue, e2e-timing · ops: GetNamespace, CreateNamespace, DeleteNamespace
  - repro: 20 threads x 30 GetNamespace; 30 sequential CreateNamespace(fresh); 30 sequential DeleteNamespace
  - measurements: get_burst_wall_s=[0.16, 0.09], get_latency_ms_min_max=[[11,156], [10,78]],
    create_30_wall_s=0.57, delete_30_wall_s=0.51, create_latency_ms_min_max=[2, 37],
    delete_latency_ms_min_max=[2, 52], n=120
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-165](../tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](../tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018),
    [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-NAMESPACE-035](../namespace.md#s3t-namespace-035), [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040) · hypotheses: H-N-043 · evidence:
    namespace/limits/pagination-and-bursts

## Notes

H-N-043 qualified: the read half is refuted at the tried scale (2 x 30 concurrent GetNamespace, all 200, wall
0.09-0.16 s; the 200-call cell is untested), the mutation half is confirmed and sharper than predicted: 30
sequential CreateNamespace (distinct names, ~33 ms each, ~30 TPS) returned 200 for calls 1-16 and
TooManyRequestsException 429 for calls 17-30, the 429s answered in 2-3 ms; 30 sequential DeleteNamespace
started 0.1 s later showed the identical 16-then-429 pattern, i.e. the limiter looks like a per-operation
token bucket of ~16 with a slow refill (the delete bucket was full although creates had just been throttled).
The 429 message is 'Too Many Requests' (not the modeled 'The limit on the number of requests per second was
exceeded.'), no Retry-After header. Each 429'd namespace was simply not created (cleanup found nothing extra).
Single burst each; 2 other probes running in this checkout.

Full original title: Back-to-back CreateNamespace/DeleteNamespace: exactly 16 succeed then every call is
TooManyRequestsException 429 'Too Many Requests' (~30 TPS); 30 concurrent GetNamespace all 200

Contradiction with [S3T-NAMESPACE-035](../namespace.md#s3t-namespace-035): 033 title says 'exactly 16 succeed then every call is 429' from a single
30-call burst per op; 035 (n=71) saw 14 of 20 succeed when 2 setup creates had already used budget, 16 of 20
after a 5 s pause, and a throttled create succeed again after 0.54 s - the count is ~16 not exactly 16 and the
429 is transient, not 'every call'; 040 adds 16-17 (once 22, once 11) Resolution: keep both; [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033)
over-generalizes a single burst, see title_fixes

Contradiction with [S3T-NAMESPACE-035](../namespace.md#s3t-namespace-035), [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040): limiter model disagrees: 033/035 describe a
per-operation token bucket with a slow refill (035 notes infer ~2 tokens/s from one retry that succeeded at
0.54 s, scope 'unknown'); 040 (3 clean runs) observes 12/13 ok at 4/s, 106/106 paced <= 5/s and whole batches
passing or failing, i.e. a ~16-per-second fixed window shared by every table bucket of the account-region - a
2 tokens/s refill would have throttled the 4/s pacing Resolution: keep both; [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040) is canonical
for the limiter's key, shape and safe pacing
