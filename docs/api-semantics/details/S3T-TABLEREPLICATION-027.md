<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-027: Re-Put (current token) on a 'failed' collision dest after the fix: 'completed' +63.3s; re-Put after DeleteTable(replica) re-creates it +75s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-027"></a>**S3T-TABLEREPLICATION-027** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Re-Put (current token) on a 'failed' collision dest after the fix: 'completed' +63.3s; re-Put after DeleteTable(replica) re-creates it +75s**
  First Puts: t1 (user table with the same name in D) [(0.1, 'no-destinations'), (72.5, 'failed')] -> failed
  (message 'A table with an identical name already exists at the destination.'); t2 [(1.4, 'no-destinations'),
  (61.0, 'pending'), (64.1, 'completed')]; t3 [(2.6, 'no-destinations'), (67.3, 'completed')]. Fixes:
  DeleteTable(user table) -> HTTP 204; DeleteTable(replica t2) -> HTTP 204; DeleteTable(replica t3) -> HTTP
  204; status right after the deletes: {'t1': 'failed', 't2': 'completed', 't3': 'completed'}. Re-Put
  (identical config, current token) on t1 -> HTTP 200 and t2 -> HTTP 200; t3 gets no Put. Status over the next
  4 min (1 call/s round-robin, ListTables(D) every 5 rounds): t1 [(0.1, 'failed'), (63.3, 'completed')]
  (listed in D at +63.4s, final message None); t2 [(1.2, 'completed'), (66.7, 'pending'), (68.7, 'completed')]
  (listed at +75.0s, final message None); t3 [(2.3, 'completed')] (listed at +Nones). Final GetTable(D, ns,
  <name>): {"t1": {"managedByService": "replication.s3tables.amazonaws.com", "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/6f26a844-8319-4864-b62c-f3c103842fd4"},
  "t2": {"managedByService": "replication.s3tables.amazonaws.com", "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/f8d4a971-f9a8-4523-bbf8-decb4a4e5457"},
  "t3": {"error": "NotFoundException"}}.
  - ACK: requeue, custom_update, synced.when, pre-delete-cleanup · ops: PutTableReplication,
    GetTableReplicationStatus, DeleteTable, ListTables, GetTable · fields: destinations[].replicationStatus,
    destinations[].failureMessage, destinations[].destinationTableArn
  - repro: Put x3 -> terminal; fix causes (delete user table / replicas); re-Put t1, t2 (token); poll 4 min
    ignoring the stale first entry
  - measurements: t1_fix_re_put_to_terminal_s=[63.3], t2_replica_recreated_s=[75.0], t1_first_failed_s=[72.5],
    n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028) ·
    hypotheses: H-U-155, H-U-156, H-U-167, H-U-176 · evidence: table/state-machine/repl-failed-recovery

## Notes

H-U-155 confirmed (a token-refreshing re-Put is the recovery path once the cause is gone). H-U-167: a deleted
replica IS re-created by a re-Put; the control t3 (no Put) stayed 'completed'. Right after any Put the status
still returns the PREVIOUS entry for ~60 s - a controller must not judge a re-Put from the first read.

Full original title: Re-Put on a 'failed' (collision) destination after the fix -> 'completed' in 63.4s;
re-Put after DeleteTable(replica) -> 'completed', replica re-created at +75.0s; no re-Put (control) ->
'completed'
