<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-303: 10 concurrent CreateTable into a FRESH namespace: wall 10.5 s (n=2; median 0.66 s, one call 10.45 s); warm namespace 0.7 s, 5 fresh <= 1.6 s
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-303"></a>**S3T-TABLE-303** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **10 concurrent CreateTable into a FRESH namespace: wall 10.5 s (n=2; median 0.66 s, one call 10.45 s); warm namespace 0.7 s, 5 fresh <= 1.6 s**
  Rounds of concurrent calls (ThreadPool, 6 s apart); contention scope: first-use only (second burst in the
  same namespace fast). CreateTable x10 in one fresh namespace: wall 10.46 s, latency 424/656/10453 ms
  (min/median/max), codes {"200": 10}; CreateTable x10 again in that namespace: wall 0.65 s, latency
  346/455/642 ms (min/median/max), codes {"200": 10}; CreateTable 5+5 in two buckets: wall 1.63 s, latency
  328/506/1616 ms (min/median/max), codes {"200": 10}, per group median {"A1": 401, "B1": 612}; CreateTable
  5+5 in two namespaces of one bucket: wall 0.87 s, latency 325/499/862 ms (min/median/max), codes {"200":
  10}, per group median {"A1": 375, "A2": 549}; CreateTable x10 in a namespace holding 5 tables: wall 0.74 s,
  latency 319/474/730 ms (min/median/max), codes {"200": 10}; CreateTable x7 in one namespace: wall 0.67 s,
  latency 333/408/669 ms (min/median/max), codes {"200": 7}; PutTableMaintenanceConfiguration on 10 different
  tables: wall 0.09 s, latency 52/65/82 ms (min/median/max), codes {"204": 10}; TagResource on 10 different
  tables: wall 0.17 s, latency 124/138/165 ms (min/median/max), codes {"200": 10}; CreateNamespace x10
  distinct names: wall 0.15 s, latency 33/82/137 ms (min/median/max), codes {"200": 10}; DeleteTable x10
  distinct tables: wall 0.44 s, latency 134/236/429 ms (min/median/max), codes {"204": 10}.
  - ACK: e2e-timing, requeue, one-per-reconcile · ops: CreateTable, CreateNamespace,
    PutTableMaintenanceConfiguration, TagResource, DeleteTable
  - repro: 2 buckets, 3 namespaces; 7 rounds of 10 concurrent calls as listed; record wall and per-call
    latency
  - measurements: R1_create_table_one_namespace.wall_s=10.46,
    R1_create_table_one_namespace.lat_ms_sorted=[424, 452, 476, 524, 524, 788, 816, 900, 909, 10453],
    R1_create_table_one_namespace.codes.200=10, R2_create_table_two_buckets.wall_s=1.63,
    R2_create_table_two_buckets.lat_ms_sorted=[328, 358, 377, 390, 401, 612, 663, 795, 856, 1616],
    R2_create_table_two_buckets.codes.200=10, R3_create_table_two_namespaces.wall_s=0.87,
    R3_create_table_two_namespaces.lat_ms_sorted=[325, 337, 352, 375, 459, 540, 549, 606, 655, 862],
    R3_create_table_two_namespaces.codes.200=10, R1b_create_table_same_namespace_again.wall_s=0.65,
    R1b_create_table_same_namespace_again.lat_ms_sorted=[346, 359, 409, 426, 426, 485, 516, 570, 631, 642],
    R1b_create_table_same_namespace_again.codes.200=10, R8_create_table_10_in_ns2.wall_s=0.74,
    R8_create_table_10_in_ns2.lat_ms_sorted=[319, 332, 341, 350, 379, 569, 588, 598, 627, 730],
    R8_create_table_10_in_ns2.codes.200=10, R9_create_table_7_in_b1.wall_s=0.67,
    R9_create_table_7_in_b1.lat_ms_sorted=[333, 343, 370, 408, 427, 496, 669],
    R9_create_table_7_in_b1.codes.200=7, R4_put_maintenance_10_tables.wall_s=0.09,
    R4_put_maintenance_10_tables.lat_ms_sorted=[52, 55, 59, 61, 65, 65, 69, 70, 71, 82],
    R4_put_maintenance_10_tables.codes.204=10, R5_tag_10_tables.wall_s=0.17,
    R5_tag_10_tables.lat_ms_sorted=[124, 127, 128, 131, 132, 145, 146, 153, 158, 165],
    R5_tag_10_tables.codes.200=10, R6_create_namespace_x10.wall_s=0.15,
    R6_create_namespace_x10.lat_ms_sorted=[33, 36, 38, 42, 67, 97, 107, 112, 116, 137],
    R6_create_namespace_x10.codes.200=10, R7_delete_table_x10.wall_s=0.44,
    R7_delete_table_x10.lat_ms_sorted=[134, 155, 162, 173, 195, 278, 307, 324, 366, 429],
    R7_delete_table_x10.codes.204=10, n=1, n_R1=2, R1_wall_s_runs=[10.49, 10.46]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-188](../table.md#s3t-table-188), [S3T-TABLE-218](../table.md#s3t-table-218), [S3T-TABLE-186](../table.md#s3t-table-186) · evidence: table/creative/write-serialization-scope

## Notes

extends [S3T-TABLE-188](../table.md#s3t-table-188) (20 concurrent CreateTable serialized to ~10 s): the ~10 s is a first-use effect of an
EMPTY namespace under >5 concurrent creates (a few finish in ~0.4 s, the rest wait 4-10 s, as if a
per-namespace initialisation lock is retried with backoff), not a steady per-bucket rate: the second burst of
10 into the same namespace, 10 into a warm namespace with 5 tables and 7 into a warm namespace all completed
in < 0.8 s, and 5 into a fresh namespace in 0.9-1.6 s. Maintenance Puts, Tags, CreateNamespace and DeleteTable
bursts on 10 distinct targets are not serialized. Decision rule: when a Namespace CR and many Table CRs are
applied together, expect the first wave of creates in that namespace to take up to ~10 s each (set the create
timeout >= 15 s or bound concurrency per new namespace to <= 5); no pacing is needed for later creates or the
other writes. R1 measured twice (10.49 s, 10.46 s).
