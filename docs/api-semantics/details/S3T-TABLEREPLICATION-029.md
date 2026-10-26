<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-029: Fixing the role policy alone does NOT heal a 'failed' dest (180 s); identical re-Put (current token): B2 pending 66s, completed 521s; B1 ok
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-029"></a>**S3T-TABLEREPLICATION-029** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Fixing the role policy alone does NOT heal a 'failed' dest (180 s); identical re-Put (current token): B2 pending 66s, completed 521s; B1 ok**
  With B2 'failed' and B1 'completed': PutRolePolicy adding B2 to the same inline policy (IAM only, no
  s3tables call) -> GetTableReplicationStatus every 5 s for 180 s: unchanged 'B1 completed / B2 failed'. Then
  GetTableReplication -> current versionToken; PutTableReplication(byte-identical configuration, that token)
  -> HTTP 200 {new versionToken, status 'success'}; status right after the Put still 'B2 failed'; +66 s B2
  'pending' (the failureMessage gone, destinationTableArn now present); +521 s B2 'completed' with
  lastSuccessfulReplicatedUpdate pointing at the same source metadata file as B1. B1 stayed 'completed'
  throughout (never flipped to pending). A replica table then exists in B2.
  - ACK: requeue, custom_update, synced.when · ops: PutRolePolicy, GetTableReplicationStatus,
    GetTableReplication, PutTableReplication · fields: destinations[].replicationStatus,
    destinations[].failureMessage, versionToken
  - repro: B2 failed -> PutRolePolicy(+B2) -> watch 3 min -> GetTableReplication -> PutTableReplication(same
    config, token) -> watch
  - measurements: self_heal_watch_s=[180], re_put_to_b2_pending_s=[65.85], re_put_to_b2_completed_s=[520.83],
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](../table-replication.md#s3t-tablereplication-039), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044),
    [S3T-TABLEREPLICATION-014](../table-replication.md#s3t-tablereplication-014), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019) ·
    hypotheses: H-U-154 · evidence: tablereplication/state-machine/per-destination-failure

## Notes

Extends H-U-154 with the recovery path: a controller that only fixes the IAM prerequisite sees the destination
stay 'failed' (at least 3 min); a token-refreshing re-Put of the unchanged configuration re-triggers the
failed destination only, without disturbing the completed one. Mirrors [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008) (bad trust) at
the per-destination level. n=1.

Full original title: Fixing the role policy alone does NOT heal a failed destination (180 s); re-Put of the
identical config (current token) -> B2 pending at +66 s, completed at +521 s; B1 undisturbed

Contradiction with [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019): when destinationTableArn is present in a status entry: 019's
first-backfill 'pending' entry has only destinationTableBucketArn+replicationStatus and its notes conclude the
ARN is carried 'only at terminal'; 029's B2 'pending' after the corrective re-Put already carries
destinationTableArn (failureMessage gone) 455 s before 'completed'; 028 adds that 'failed' entries never carry
it; n=1 each Resolution: keep both; [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029) narrows 019: destinationTableArn is absent on
'failed' and on the first-backfill 'pending' but present on a re-Put 'pending' and at 'completed' - presence
is per entry and optional in every non-completed state, not a terminal marker
