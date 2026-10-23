<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-018: CreateTableBucket same-name 409 storm at ~8 calls/s is throttled: TooManyRequestsException 429 in runs of 4-6 after 5-9 accepted (~5 TPS)
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-018"></a>**S3T-TABLEBUCKET-018** `quota-limit` · impact medium · handled · verified 2026-10-13
  **CreateTableBucket same-name 409 storm at ~8 calls/s is throttled: TooManyRequestsException 429 in runs of 4-6 after 5-9 accepted (~5 TPS)**
  500 CreateTableBucket calls for one name at ~100 ms spacing (~8/s) over 60 s: 314 reached the service (409
  ConflictException) and 186 were throttled with TooManyRequestsException HTTP 429 'Too Many Requests', in a
  regular pattern of 5-9 accepted then 4-6 throttled, consistent with a ~5 TPS token bucket for
  CreateTableBucket. Other probers shared the account, so the exact limit is an estimate.
  - ACK: requeue, terminal_codes · ops: CreateTableBucket
  - repro: CreateTableBucket(same name) every 100 ms for 60 s; count 429s
  - measurements: calls=500, throttled_429=186, window_s=60, approx_accept_rate_per_s=5, n=1
  - handling: handled via `test/e2e/tests/test_table_bucket.py:48-69; test/e2e/tests/test_table_bucket.py:225-233; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-068](../tablebucket.md#s3t-tablebucket-068), [S3T-TABLEBUCKET-179](../tablebucket.md#s3t-tablebucket-179), [S3T-TABLEBUCKET-071](../tablebucket.md#s3t-tablebucket-071), [S3T-TABLEBUCKET-165](../tablebucket.md#s3t-tablebucket-165),
    [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](../tablebucket-policy.md#s3t-tablebucket-081), [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094),
    [S3T-SERVICE-004](../service.md#s3t-service-004), [S3T-SERVICE-008](../service.md#s3t-service-008), [S3T-SERVICE-018](../service.md#s3t-service-018) · evidence:
    tablebucket/idempotency/dup-create-delete-recreate

## Notes

Observed incidentally; TooManyRequestsException is a modeled error on every bucket operation and must be
retried, not treated as terminal.

Full original title: CreateTableBucket at ~8 calls/s is throttled: TooManyRequestsException (HTTP 429, 'Too
Many Requests') in bursts of 4-6 after 5-9 accepted calls
