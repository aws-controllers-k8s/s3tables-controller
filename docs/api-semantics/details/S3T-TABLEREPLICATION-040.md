<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-040: Destination bucket deleted under a completed config: 'failed: The specified destination bucket does not exist.' in 1 of 2 runs; re-Put 200
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-040"></a>**S3T-TABLEREPLICATION-040** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Destination bucket deleted under a completed config: 'failed: The specified destination bucket does not exist.' in 1 of 2 runs; re-Put 200**
  Replica table + namespace deleted in D, then DeleteTableBucket(D) -> HTTP 204 while the source's
  configuration still named D. GetTableReplicationStatus + GetTableReplication every 5 s for 183s: [{"t": 0.0,
  "status": "failed:The specified destination bucket does not exist.", "get_config": "HTTP 200"}]. Identical
  re-Put naming the deleted bucket (current token) -> HTTP 200; status right after 'failed:The specified
  destination bucket does not exist.', after 20 s 'failed:The specified destination bucket does not exist.'.
  PutTableReplication switching to a new bucket D2 -> HTTP 200; status right after 'failed:The specified
  destination bucket does not exist.'; watch: [{"elapsed_s": 0.03, "value": "failed:The specified destination
  bucket does not exist."}]; final entries [{"replicationStatus": "failed", "failureMessage": "The specified
  destination bucket does not exist."}].
  - ACK: requeue, terminal_codes, custom_update · ops: DeleteTableBucket, GetTableReplicationStatus,
    GetTableReplication, PutTableReplication · fields: destinations[].replicationStatus,
    destinations[].failureMessage, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: completed replication -> empty + DeleteTableBucket(destination) -> poll status 3 min -> re-Put same
    config -> Put new destination -> watch
  - measurements: lost_destination_window_s=182.7, status_change_at_s=null, heal_completed_at_s=0.03, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118),
    [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023), [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048) ·
    evidence: tablereplication/creative/reput-storm-and-lost-destination

## Notes

Cross-service analogy (ElastiCache CacheSubnetGroupNotFound at call time; DynamoDB replica of a deleted
table): S3 Tables accepts a configuration whose destination no longer exists (re-Put naming the deleted bucket
-> HTTP 200, as does a Put to a brand-new bucket D2) and reports the problem only through
GetTableReplicationStatus, which showed 'failed' with failureMessage 'The specified destination bucket does
not exist.' on the first poll (t=0.0 s) after DeleteTableBucket(D) returned, i.e. the status appears to be
evaluated against the live destination rather than by an asynchronous job; it stayed so for 183 s and
GetTableReplication kept returning the configuration (200). The heal after switching to D2 was NOT observed in
this run (the watch stopped on the still-listed old 'failed' entry at +0.03 s); see the follow-up probe
tablereplication/creative/lost-destination-heal ([S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042)). Related: [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030)
notes saw the same message incidentally on a cross-region destination. Empty source, n=1. CORRECTION from that
follow-up: repeating the same delete sequence on a fresh setup WITHOUT the preceding re-Put storm left the
status at 'completed' for the deleted destination (0 s and ~60 s later) until a re-pointing Put replaced the
entry - so 'at once' is not reproducible (n=1 vs n=1); the stable facts are: Put accepts a deleted/nonexistent
destination (200) and the status is the only place the problem can surface.

Contradiction with [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042): how a deleted destination bucket is reported under a completed
config: 040 saw 'failed: The specified destination bucket does not exist.' on the first poll after
DeleteTableBucket(D) and for 183 s, also after re-Puts naming it; 042 (same sequence) saw the dead D stay
'completed' (0 s and ~60 s later) until a re-pointing Put replaced the entry, and a config [dead D, D2] then
showed NO entry for D at all (0 s, 30 s); n=1 each, the only setup difference is a 10-call re-Put storm ~60 s
before the delete in 040; both agree Put accepts the dead destination with 200 Resolution: re-probe: creative -
completed table-level config; DeleteTableBucket(destination) with and without a preceding 10-call re-Put storm
(n>=3 each); poll GetTableReplicationStatus at 1 s for 180 s and record first 'failed' vs 'completed'
persistence; then Put [dead, new] and record whether the dead destination has a 'failed' entry or none

Contradiction with [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042): FU2-05 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: both readings are right under different
conditions, and the preceding re-Put storm is the deciding condition: for the 3 tables whose config received
10 identical re-Puts ~60 s before DeleteTableBucket(destination), GetTableReplicationStatus showed 'failed:
The specified destination bucket does not exist.' within 4.4-4.6 s of the delete (one of them already on the
first poll right after the Resolution: keep all; [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048) is canonical for the settled rule
