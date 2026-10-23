<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-218: DeleteTable back-to-back at ~6.5/s: 68 x 204 then 52 x TooManyRequestsException 429 (n=120); 120 CreateTable back-to-back at ~3/s: no 429
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-218"></a>**S3T-TABLE-218** `quota-limit` · impact high · handled · verified 2026-10-13
  **DeleteTable back-to-back at ~6.5/s: 68 x 204 then 52 x TooManyRequestsException 429 (n=120); 120 CreateTable back-to-back at ~3/s: no 429**
  120 CreateTable (format=ICEBERG, no metadata) issued one after another with no pause across 2 namespaces ->
  {"200": 120} in 39.5s (3.0/s; latency [278, 395] ms). 120 DeleteTable likewise -> {"204": 68,
  "TooManyRequestsException/429": 52} in 10.4s (11.5/s; latency [2, 203] ms) messages
  {"TooManyRequestsException": "Too Many Requests"}.
  - ACK: e2e-timing, requeue · ops: CreateTable, DeleteTable
  - repro: 120 x CreateTable back-to-back; 120 x DeleteTable back-to-back; no-retry client
  - measurements: create_wall_s=39.53, create_latency_ms_min_max=[278, 395], delete_wall_s=10.41,
    delete_latency_ms_min_max=[2, 203], n=240
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-165](../tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](../tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018),
    [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLE-188](../table.md#s3t-table-188), [S3T-TABLE-217](../table.md#s3t-table-217), [S3T-TABLE-186](../table.md#s3t-table-186) · hypotheses: H-T-071,
    H-N-043 · evidence: table/limits/default-page-120-tables

## Notes

Rate observation (H-T-071 setup, H-N-043 context): 120 CreateTable one after another (278-395 ms each, 3.0/s)
-> 120 x 200, no 429; 120 DeleteTable one after another (130-200 ms each, ~6.5/s) -> the first 67 returned 204
and calls 68-120 all returned TooManyRequestsException 429 'Too Many Requests' in 2-3 ms (no Retry-After). The
remaining 53 tables were then deleted with 0.3 s pacing (~2.2/s) without a single 429. So DeleteTable is
rate-limited well below its own latency-bound pace while CreateTable at 3/s is not; a controller deleting many
tables (e.g. namespace teardown) must back off on 429. 4 other probes running in this checkout.

Full original title: DeleteTable back-to-back: 67 x 204 (~6.5/s) then TooManyRequestsException 429 'Too Many
Requests' for the rest; 120 CreateTable back-to-back at ~3/s never throttled
