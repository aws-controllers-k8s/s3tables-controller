<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-020: Re-Put after DeleteTableReplication with the service-managed replica left behind -> 200 but 'failed: identical name already exists' at +62s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-020"></a>**S3T-TABLEREPLICATION-020** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Re-Put after DeleteTableReplication with the service-managed replica left behind -> 200 but 'failed: identical name already exists' at +62s** (hypothesis refuted; behavior confirmed)
  (a) With the leftover replica present: PutTableReplication(identical config, no token) -> HTTP 200; status
  transitions [(0.05, 'no-destinations'), (60.95, 'pending'), (61.99, 'failed')]; destinationTableArn DIFFERS
  from the pre-delete replica ARN; ListTables(D, ns) afterwards: ['ackq_f81240_t']. (b) Then
  DeleteTable(replica) -> HTTP 204; status right after: 'failed'; PutTableReplication(identical, current
  token) -> HTTP 200; the watch stopped at once on the STALE 'failed' entry (0.05 s), so whether the re-Put
  re-creates the replica was NOT observed here (re-run in table/state-machine/repl-failed-recovery); final
  status body {"destinations": [{"destinationTableBucketArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest", "failureMessage": "A table with an
  identical name already exists at the destination.", "replicationStatus": "failed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"};
  ListTables(D, ns) at the end: [].
  - ACK: custom_create, requeue, pre-delete-cleanup · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTable, ListTables · fields: destinations[].destinationTableArn, destinations[].replicationStatus,
    destinations[].failureMessage
  - repro: DeleteTableReplication -> Put identical -> watch; DeleteTable(replica) -> Put identical (token) ->
    watch 7 min + ListTables(D)
  - measurements: re_put_to_failed_s=[61.99], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224),
    [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008) · hypotheses: H-U-176, H-U-166, H-U-167 · evidence:
    table/state-machine/repl-status-lifecycle

## Notes

H-U-176 REFUTED: the leftover service-managed replica is NOT re-attached; the new configuration fails with 'A
table with an identical name already exists at the destination.' (the same failure as a user-table collision),
so after a Delete+re-Put cycle the controller must delete the leftover replica first. H-U-166 refuted /
H-U-167 qualified again: DeleteTable(replica) -> 204 while configured. Re-creation after a replica delete +
re-Put: inconclusive here (stale status), see repl-failed-recovery.

Full original title: Re-Put after DeleteTableReplication with the replica left behind -> 200 but 'failed: A
table with an identical name already exists at the destination' at +62 s
