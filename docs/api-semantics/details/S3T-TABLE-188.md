<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-188: 20 concurrent CreateTable (distinct names, one namespace) all 200 but serialized (wall 9.7 s, latency 0.3-9.7 s); 20 DeleteTable at 4 TPS ok
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-188"></a>**S3T-TABLE-188** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **20 concurrent CreateTable (distinct names, one namespace) all 200 but serialized (wall 9.7 s, latency 0.3-9.7 s); 20 DeleteTable at 4 TPS ok**
  20 CreateTable calls for distinct names in one namespace fired concurrently (20 threads, wall 9.69s) ->
  {"200": 20}. ListTables afterwards listed 20 of the 20 names, none of the rejected names exists. Deleting
  the created tables sequentially at 4 TPS with the no-retry client -> {"204": 20}.
  - ACK: requeue, e2e-timing · ops: CreateTable, ListTables, DeleteTable · fields: name
  - repro: 20 threads x CreateTable(distinct names); retry a rejected one every 250 ms; ListTables;
    DeleteTable each at 4 TPS
  - measurements: burst_wall_s=9.69, create_latency_ms_min_max=[331, 9679], throttle_refill_s=null,
    delete_latency_ms_min_max=[140, 195], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLE-217](../table.md#s3t-table-217), [S3T-TABLE-186](../table.md#s3t-table-186) · hypotheses: H-T-045 · evidence:
    table/limits/tag-count-and-read-bursts

## Notes

H-T-045 create half confirmed (no 429/409 for 20 parallel CreateTable) but qualified: the service appears to
serialize table creation within a namespace/bucket - 20 concurrent creates completed one after another
(latencies 331-9679 ms, wall 9.69 s) instead of in parallel, so a controller creating many tables at once pays
O(n) latency on each call. Contrast: 30 sequential CreateNamespace hit TooManyRequestsException after 16 (see
the Namespace finding). Single burst; 5 other probes running.

Full original title: 20 concurrent CreateTable (distinct names, one namespace) all 200 but serialized: wall
9.7 s, latencies 0.3-9.7 s; 20 DeleteTable at 4 TPS all 204
