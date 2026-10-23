<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-042: Re-pointing after a lost destination: Put [D2] -> D2 pending at +61 s, completed at +66 s, dead D entry gone; Put [dead D, D2] lists only D2
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-042"></a>**S3T-TABLEREPLICATION-042** `stale-response` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Re-pointing after a lost destination: Put [D2] -> D2 pending at +61 s, completed at +66 s, dead D entry gone; Put [dead D, D2] lists only D2**
  After DeleteTableBucket(D) (replica + namespace removed first) the status read 'DST=completed'.
  CreateTableBucket(D2) then PutTableReplication([D2], current token) -> HTTP 200; status right after:
  'DST=completed'; transitions while watching (5 s): [[0.02, "DST=completed"], [60.68,
  "DST=completed;DST2=pending"], [65.73, "DST2=completed"]]; final 'DST2=completed'; D2 contents
  {"ackq_62668e_ns": ["ackq_62668e_t"]}. Then PutTableReplication([D, D2]) with D deleted -> HTTP 200; status
  right after 'DST2=completed', after 30 s 'DST2=completed'; raw entries [{"replicationStatus": "completed",
  "destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-62668e-ldh-dst2/table/f4862249-85b2-4af0-9642-c6881a826dc1",
  "dest": "DST2"}].
  - ACK: requeue, custom_update, terminal_codes · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTableBucket, ListTables · fields: destinations[].replicationStatus, destinations[].failureMessage,
    destinations[].destinationTableBucketArn
  - repro: completed replication -> delete destination bucket -> PutTableReplication([new bucket]) -> watch
    per-destination status -> PutTableReplication([dead, new])
  - measurements: d2_completed_s=65.73, heal_window_s=360, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027),
    [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047), [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048) · evidence:
    tablereplication/creative/lost-destination-heal

## Notes

Follow-up to [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040) (status 'failed: The specified destination bucket does not exist.' at
once after the destination is deleted; Put accepted). Cross-service analogy (ElastiCache/DynamoDB: a deleted
dependency must be re-pointed by the controller): whether re-pointing is enough and how long the heal takes.
Extends [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017) (same-region destination switch while healthy) and
[S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027)/029 (re-Put heals a failed collision; policy fix alone does not). Empty source, n=1.
CONTRAST with [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040) (same sequence, n=1 each): there the status flipped to 'failed: The
specified destination bucket does not exist.' on the first read after DeleteTableBucket(D); here it kept
reading 'completed' for D (0 s and ~60 s later) until the re-pointing Put replaced the entry - the only
difference in setup was a 10-call re-Put storm ~60 s before the delete in the earlier run. Treat 'how fast a
dead destination is reported' as unresolved; what is consistent is that Put accepts dead destinations (200)
and that a configuration naming a deleted bucket shows NO entry for it in GetTableReplicationStatus (K stage:
config [D, D2] -> status [D2=completed] only, 0 s and 30 s), so a controller cannot rely on a status entry per
configured destination.

Contradiction with [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017): time until a newly configured destination appears in
GetTableReplicationStatus after a Put on an existing completed table-level config: 017 (replace [D]->[E], same
region, empty table) saw E:pending at +1.06 s and E:completed at +2.1 s; 042 (replace [dead D]->[D2]) saw
D2:pending at +60.68 s and completed at +65.73 s, and 021 (+61.04 s) / 022 (+61.17 s) / 031 ('previous entry
shown for ~60 s after any Put') agree with 042; all n=1 - 017's Put landed 63 s after the first Put, i.e.
possibly right at a ~60 s evaluation tick Resolution: re-probe: state-machine - completed table-level config
on an empty table, Put [E] replacing [D] issued at varied offsets after the last status change (5/20/40/55/70
s, n>=5), poll GetTableReplicationStatus at 1 s, record first E entry and E:completed; test whether the ~60 s
delay is a periodic evaluation tick rather than a fixed latency

Contradiction with [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040): how a deleted destination bucket is reported under a completed
config: 040 saw 'failed: The specified destination bucket does not exist.' on the first poll after
DeleteTableBucket(D) and for 183 s, also after re-Puts naming it; 042 (same sequence) saw the dead D stay
'completed' (0 s and ~60 s later) until a re-pointing Put replaced the entry, and a config [dead D, D2] then
showed NO entry for D at all (0 s, 30 s); n=1 each, the only setup difference is a 10-call re-Put storm ~60 s
before the delete in 040; both agree Put accepts the dead destination with 200 Resolution: re-probe: creative -
completed table-level config; DeleteTableBucket(destination) with and without a preceding 10-call re-Put storm
(n>=3 each); poll GetTableReplicationStatus at 1 s for 180 s and record first 'failed' vs 'completed'
persistence; then Put [dead, new] and record whether the dead destination has a 'failed' entry or none

Contradiction with [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
[S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022): FU2-04 settled by follow-up probe
tablereplication/state-machine/fu2-04-put-offset-tick: 042/021/022 hold; 017's +1.06 s did not reproduce. The
delay is a fixed ~61 s per Put, not a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last
observed status change (67-565 s after the first Put) all showed the new destination as 'pending' 61.0-64.1 s
later and 'completed' 1.0-1.1 s after that (empty table, same-region buckets); the first Put behaved
identically (pendi Resolution: keep all; [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) is canonical for the settled rule

Contradiction with [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048), [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040): FU2-05 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: both readings are right under different
conditions, and the preceding re-Put storm is the deciding condition: for the 3 tables whose config received
10 identical re-Puts ~60 s before DeleteTableBucket(destination), GetTableReplicationStatus showed 'failed:
The specified destination bucket does not exist.' within 4.4-4.6 s of the delete (one of them already on the
first poll right after the Resolution: keep all; [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048) is canonical for the settled rule
