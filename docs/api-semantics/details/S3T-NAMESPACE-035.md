<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-035: CreateNamespace back-to-back: 14/20 then 429 (2 setup calls used budget); retry every 250 ms succeeds after 0.54 s; after a 5 s pause 16/20
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-035"></a>**S3T-NAMESPACE-035** `quota-limit` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace back-to-back: 14/20 then 429 (2 setup calls used budget); retry every 250 ms succeeds after 0.54 s; after a 5 s pause 16/20**
  20 sequential CreateNamespace (distinct names, no pause, ~2 ms each) -> {"200": 14,
  "TooManyRequestsException/429": 6} in 0.45s, first TooManyRequestsException at index 14 (message 'Too Many
  Requests', Retry-After []). Retrying the first rejected name every 250 ms succeeded after 0.54 s (attempts
  {'TooManyRequestsException (HTTP 429)': 2, 'HTTP 200': 1}). After a 5 s pause, 20 more sequential creates ->
  {"200": 16, "TooManyRequestsException/429": 4} (first failure at index 16). After another 5 s, 31 sequential
  DeleteNamespace -> {"204": 16, "TooManyRequestsException/429": 15} in 0.44s (first failure at index 16); a
  throttled delete succeeded again after 0.54 s.
  - ACK: requeue, e2e-timing · ops: CreateNamespace, DeleteNamespace · fields: namespace
  - repro: 20 x CreateNamespace back-to-back; retry the first 429 every 250 ms; sleep 5; 20 more; sleep 5;
    DeleteNamespace all back-to-back
  - measurements: create_wall_s=0.45, create_latency_ms_min_max=[2, 40], first_429_index_create=14,
    refill_lag_s_create=0.54, ok_after_5s_pause=16, first_429_index_delete=16, refill_lag_s_delete=0.54, n=71
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040) · hypotheses: H-N-043 · evidence:
    table/limits/default-page-120-tables

## Notes

Follow-up to the Namespace 16-then-429 finding (H-N-043 mutation half, qualified further): the limiter behaves
like a token bucket of capacity 16 per operation - 2 setup CreateNamespace calls a few seconds earlier left 14
tokens (14 of 20 succeeded, 429 from index 14), after a 5 s pause exactly 16 of 20 succeeded, and 31
back-to-back DeleteNamespace gave 16 x 204 then 15 x 429. A throttled create/delete retried every 250 ms
succeeded after 0.54 s in both cases (2 more 429s then success), i.e. a refill of roughly 2 tokens/s. 429
message 'Too Many Requests', no Retry-After header, 429s answered in 2-3 ms. Scope (per account vs per bucket)
unknown; 3 other probes were running in this checkout and may have shared the bucket.

Contradiction with [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033): 033 title says 'exactly 16 succeed then every call is 429' from a single
30-call burst per op; 035 (n=71) saw 14 of 20 succeed when 2 setup creates had already used budget, 16 of 20
after a 5 s pause, and a throttled create succeed again after 0.54 s - the count is ~16 not exactly 16 and the
429 is transient, not 'every call'; 040 adds 16-17 (once 22, once 11) Resolution: keep both; [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033)
over-generalizes a single burst, see title_fixes

Contradiction with [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040): limiter model disagrees: 033/035 describe a
per-operation token bucket with a slow refill (035 notes infer ~2 tokens/s from one retry that succeeded at
0.54 s, scope 'unknown'); 040 (3 clean runs) observes 12/13 ok at 4/s, 106/106 paced <= 5/s and whole batches
passing or failing, i.e. a ~16-per-second fixed window shared by every table bucket of the account-region - a
2 tokens/s refill would have throttled the 4/s pacing Resolution: keep both; [S3T-NAMESPACE-040](../namespace.md#s3t-namespace-040) is canonical
for the limiter's key, shape and safe pacing
