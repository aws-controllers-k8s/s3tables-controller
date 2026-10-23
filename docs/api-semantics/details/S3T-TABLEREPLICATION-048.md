<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-048: Deleted destination bucket: 'failed ... does not exist.' within 5 s only after a re-Put storm (3/3); otherwise 'completed' for 180 s (3/3)
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-048"></a>**S3T-TABLEREPLICATION-048** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Deleted destination bucket: 'failed ... does not exist.' within 5 s only after a re-Put storm (3/3); otherwise 'completed' for 180 s (3/3)**
  Six empty tables, each with a completed table-level config to its own same-region bucket E_k. t0-t2: 10
  identical re-Puts (current token chain) -> {'t0': ['HTTP 200'], 't1': ['HTTP 200'], 't2': ['HTTP 200']},
  status right after the storm {'t0': 'E0=completed', 't1': 'E1=completed', 't2': 'E2=completed'}; ~60 s later
  all six destination buckets were emptied and deleted ({'t0': 'HTTP 204', 't1': 'HTTP 204', 't2': 'HTTP 204',
  't3': 'HTTP 204', 't4': 'HTTP 204', 't5': 'HTTP 204'}). First GetTableReplicationStatus after each delete:
  {'t0': 'E0=completed', 't1': 'E1=completed', 't2': 'E2=failed:The specified destination bucket does not
  exist.', 't3': 'E3=completed', 't4': 'E4=completed', 't5': 'E5=completed'}. Round-robin polling (<= 1
  call/s, each table every ~6 s) for 180 s, transitions relative to the delete: {"t0": [[4.4, "E0=failed:The
  specified destination bucket does not exist."]], "t1": [[4.6, "E1=failed:The specified destination bucket
  does not exist."]], "t2": [[4.5, "E2=failed:The specified destination bucket does not exist."]], "t3":
  [[4.7, "E3=completed"]], "t4": [[5.0, "E4=completed"]], "t5": [[5.3, "E5=completed"]]}. Then
  PutTableReplication [dead E_k, new N_k] -> {'t0': 'HTTP 200', 't1': 'HTTP 200', 't2': 'HTTP 200', 't3':
  'HTTP 200', 't4': 'HTTP 200', 't5': 'HTTP 200'}; over 150 s the dead destination was listed at all and was
  'failed'; new destination completed at {'t0': 63.2, 't1': 64.0, 't2': 71.1, 't3': 72.0, 't4': 66.7, 't5':
  80.0} s; final views {'t0': 'E0=failed:The specified destination bucket does not exist.|N0=completed', 't1':
  'E1=failed:The specified destination bucket does not exist.|N1=completed', 't2': 'E2=failed:The specified
  destination bucket does not exist.|N2=completed', 't3': 'E3=failed:The specified destination bucket does not
  exist.|N3=completed', 't4': 'E4=failed:The specified destination bucket does not exist.|N4=completed', 't5':
  'E5=failed:The specified destination bucket does not exist.|N5=completed'}.
  - ACK: requeue, synced.when, custom_update · ops: PutTableReplication, DeleteTableBucket,
    GetTableReplicationStatus · fields: destinations[].replicationStatus, destinations[].failureMessage,
    destinations[].destinationTableBucketArn
  - repro: 6 tables with completed table-level replication -> 10 re-Puts on 3 of them -> 60 s ->
    empty+DeleteTableBucket(destination) x6 -> poll status 180 s -> PutTableReplication [dead, new] x6 -> poll
    150 s
  - measurements: storm_first_failed_s=[4.4, 4.6, 4.5], no_storm_first_failed_s=[null, null, null],
    dead_poll_s=180, heal_poll_s=150, heal_new_completed_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0], n=6,
    heal_dead_failed_entry_appears_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042) · evidence:
    tablereplication/creative/fu2-03-05-orphan-body-dead-destination

## Notes

FU2-05: both readings are right under different conditions, and the preceding re-Put storm is the deciding
condition: for the 3 tables whose config received 10 identical re-Puts ~60 s before
DeleteTableBucket(destination), GetTableReplicationStatus showed 'failed: The specified destination bucket
does not exist.' within 4.4-4.6 s of the delete (one of them already on the first poll right after the delete,
as 040 saw), and kept it for 180 s; for the 3 tables without a storm the dead destination kept reading
'completed' for the whole 180 s (as 042 saw). Explanation consistent with [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047): each Put
schedules an evaluation ~61 s later - the storm's last Put was evaluated right after the delete and found the
bucket missing, while an untouched completed config is not re-evaluated. After PutTableReplication [dead E,
new N] (HTTP 200, 6/6) the status kept the previous view for ~60 s (old failed entry for the storm group,
stale 'completed' for the other) and then, 63-80 s after the Put, showed BOTH 'E=failed: The specified
destination bucket does not exist.' and 'N=completed' (6/6). 042's 'no entry for the dead destination at 0/30
s' was the ~60 s stale view; the dead destination does get a 'failed' entry once the Put is evaluated. The
status of an untouched config therefore does not reflect a deleted destination until something triggers an
evaluation.

Contradiction with [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042): FU2-05 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: both readings are right under different
conditions, and the preceding re-Put storm is the deciding condition: for the 3 tables whose config received
10 identical re-Puts ~60 s before DeleteTableBucket(destination), GetTableReplicationStatus showed 'failed:
The specified destination bucket does not exist.' within 4.4-4.6 s of the delete (one of them already on the
first poll right after the Resolution: keep all; [S3T-TABLEREPLICATION-048](../table-replication-status.md#s3t-tablereplication-048) is canonical for the settled rule
