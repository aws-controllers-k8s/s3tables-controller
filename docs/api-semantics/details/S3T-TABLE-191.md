<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-191: Concurrent writes to one table's policy/maintenance -> ConflictException 409 (1 of 10 also 500); compaction and snapshot Puts conflict
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-191"></a>**S3T-TABLE-191** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Concurrent writes to one table's policy/maintenance -> ConflictException 409 (1 of 10 also 500); compaction and snapshot Puts conflict**
  10 PutTablePolicy with distinct Sids fired concurrently on one table (wall 6.22s, latencies [155, 6210] ms)
  -> {"200": 2, "ConflictException/409": 7, "InternalServerErrorException/500": 1} messages
  {"ConflictException": "The resource was modified by a concurrent request. Please retry your request.",
  "InternalServerErrorException": "An internal error occurred. Try again."}; GetTablePolicy then held Sid P04,
  which is one of the 2 successful puts; rejected puts retried sequentially -> [('P01', 'ConflictException',
  'HTTP 200'), ('P02', 'ConflictException', 'HTTP 200'), ('P03', 'ConflictException', 'HTTP 200'), ('P05',
  'InternalServerErrorException', 'HTTP 200'), ('P06', 'ConflictException', 'HTTP 200'), ('P07',
  'ConflictException', 'HTTP 200'), ('P08', 'ConflictException', 'HTTP 200'), ('P09', 'ConflictException',
  'HTTP 200')]. Ten rounds of PutTableMaintenanceConfiguration(icebergCompaction) and
  (icebergSnapshotManagement) fired together (distinct values per round) -> {"HTTP 204": 10,
  "ConflictException (HTTP 409)": 10}, per-round wall 1.04-1.08 s messages {"snapshot": "The resource was
  modified by a concurrent request. Please retry your request.", "compaction": "The resource was modified by a
  concurrent request. Please retry your request."}; GetTableMaintenanceConfiguration afterwards listed types
  ['icebergCompaction', 'icebergSnapshotManagement'] with compaction {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 73, "strategy": "binpack"}}} and snapshot {"status": "enabled",
  "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 8, "maxSnapshotAgeHours": 31}}} (the last
  round sent targetFileSizeMB=73, minSnapshotsToKeep=10, maxSnapshotAgeHours=33).
  - ACK: requeue, one-per-reconcile, custom_update · ops: PutTablePolicy, GetTablePolicy,
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration · fields: resourcePolicy,
    configuration.icebergCompaction, configuration.icebergSnapshotManagement
  - repro: 10 threads x PutTablePolicy(distinct Sid); 10 x
    [PutTableMaintenanceConfiguration(icebergCompaction) || (icebergSnapshotManagement)]; Get* after
  - measurements: policy_race_wall_s=6.22, policy_latency_ms_min_max=[155, 6210], pair_wall_s=[1.08, 1.04,
    1.05, 1.05, 1.04, 1.05, 1.04, 1.04, 1.04, 1.04], n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-207](../table-maintenance.md#s3t-table-207), [S3T-TABLE-208](../table-maintenance.md#s3t-table-208), [S3T-TABLE-209](../table-maintenance.md#s3t-table-209) · hypotheses: H-U-064 · evidence:
    table/limits/policy-size-write-races

## Notes

H-U-064 refuted: (1) 10 concurrent PutTablePolicy -> 2 x 200, 7 x ConflictException 409 'The resource was
modified by a concurrent request. Please retry your request.', 1 x InternalServerErrorException 500 'An
internal error occurred. Try again.'; the stored policy was one of the two successes; all 8 rejected puts
succeeded on sequential retry; the burst took 6.2 s wall (per-table write serialization). (2) In 10 of 10
rounds, firing PutTableMaintenanceConfiguration(icebergCompaction) and (icebergSnapshotManagement) together
yielded exactly one 204 and one 409 with the same 'modified by a concurrent request' message (the loser
alternated), each round ~1.05 s wall with the loser answering after ~1.04 s - the two maintenance types share
one write lock, so a controller must send them sequentially (or requeue on 409). Final Get showed both types
present with the values of the last successful round of each type (targetFileSizeMB=73 = round 10;
minSnapshotsToKeep=8/maxSnapshotAgeHours=31 = round 8). Contrast: concurrent PutTableBucketPolicy on a BUCKET
produced 500s and no 409. 3-4 other probes running.

Full original title: Concurrent writes to one table's policy/maintenance -> ConflictException 409 'modified by
a concurrent request' (plus some 500); compaction vs snapshot Puts conflict with each other
