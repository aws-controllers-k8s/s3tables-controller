<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-300: Replica settings are a replication-time snapshot: later source maintenance and policy changes did not reach the replica in 184 s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-300"></a>**S3T-TABLE-300** `immutable-field` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Replica settings are a replication-time snapshot: later source maintenance and policy changes did not reach the replica in 184 s**
  With the destination 'completed': source PutTableMaintenanceConfiguration compaction 384/binpack -> HTTP
  204, snapshots disabled 3/72 -> HTTP 204, TagResource {stage: v2} -> InternalServerErrorException (HTTP 500)
  'An internal error occurred. Try again.', PutTablePolicy (Sid SrcAllowGetV2) -> HTTP 200. Replica polled
  every 5 s (tags, policy, maintenance) for 184s: 36 samples, first change none; final replica vs pre-change
  snapshot differs in []; final replica maintenance {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5,
  "maxSnapshotAgeHours": 48}}}} vs source {"icebergCompaction": {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 384, "strategy": "binpack"}}}, "icebergSnapshotManagement":
  {"status": "disabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 3,
  "maxSnapshotAgeHours": 72}}}}. GetTableReplicationStatus during the window: [(0.2, 'completed')].
  - ACK: requeue, scope:skip · ops: PutTableMaintenanceConfiguration, TagResource, PutTablePolicy,
    GetTableMaintenanceConfiguration, ListTagsForResource, GetTablePolicy, GetTableReplicationStatus · fields:
    configuration, tags, resourcePolicy, destinations[].replicationStatus
  - repro: replica completed -> change source maintenance config + tags + policy -> poll the replica's reads
    every 5 s for 3 min
  - measurements: window_s=184.3, first_change_s=null, n=36
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299), [S3T-TABLE-306](../table-replication-status.md#s3t-table-306), [S3T-SERVICE-007](../service.md#s3t-service-007),
    [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-TABLE-304](../table-replication.md#s3t-table-304) · evidence: table/creative/replica-settings-fidelity

## Notes

Cross-service analogy (ElastiCache global datastore propagates parameter changes; DynamoDB replica settings
are per-region): tells a controller whether managing the source's sub-resources implicitly manages the
replica's. Config-only changes (no metadata commit) on an empty table; a commit-triggered sync
([S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), ~7 min) was not waited for. n=1. The TagResource on the SOURCE in this step (one
sequential call, ~100 ms after two PutTableMaintenanceConfiguration calls) failed with
InternalServerErrorException 500 'An internal error occurred. Try again.' so the tag cell of the live-sync
test is untested; see the follow-up probe table/creative/sequential-write-500.
