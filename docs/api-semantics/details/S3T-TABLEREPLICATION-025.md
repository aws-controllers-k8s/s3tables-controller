<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-025: Collision with a user table in D: Put 200, 'D:failed' (identical name exists), user table untouched; same-region dest/existing ns: completed
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-025"></a>**S3T-TABLEREPLICATION-025** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Collision with a user table in D: Put 200, 'D:failed' (identical name exists), user table untouched; same-region dest/existing ns: completed**
  (t1) A user-created table with the source's namespace/name already exists in D: PutTableReplication -> HTTP
  200; status transitions [(0.1, 'no-destinations'), (62.0, 'D:failed')]; terminal entry {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "failureMessage": "A table with an identical name already exists at the destination.", "replicationStatus":
  "failed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/58ee5109-3a05-4669-a8e6-a78158cdf347"}.
  GetTable(user table in D) after: versionToken unchanged, managedByService=None, modifiedAt unchanged.
  ListTables(D, ns) -> [{"arn_is_user_table": true, "name": "ackq_9d291a_t1"}, {"arn_is_user_table": false,
  "name": "ackq_9d291a_t3"}]. (t2) Destination E in the SAME region as the source: Put -> HTTP 200;
  transitions [(1.1, 'no-destinations'), (71.2, 'E:pending'), (72.3, 'E:completed')]; GetTable(E replica) ->
  {"managedByService": "replication.s3tables.amazonaws.com", "managedTableInformation":
  {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/21e11778-a08e-4f39-a7f6-f0dd8e00f285"}},
  "type": "customer"}. (t3) The destination namespace already existed (created by the user) but not the table:
  Put -> HTTP 200; transitions [(2.2, 'no-destinations'), (66.1, 'D:pending'), (68.1, 'D:completed')];
  GetTable(D replica) -> {"managedByService": "replication.s3tables.amazonaws.com", "managedTableInformation":
  {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/f102c252-f65c-4530-a177-ae241e61e940"}},
  "type": "customer"}.
  - ACK: synced.when, requeue, pre-delete-cleanup · ops: PutTableReplication, GetTableReplicationStatus,
    GetTable, ListTables, CreateTable, CreateNamespace · fields: destinations[].replicationStatus,
    destinations[].failureMessage, managedByService
  - repro: CreateNamespace+CreateTable(same names) in D -> Put t1->D; Put t2->E(same region); Put t3->D
    (namespace exists); poll 1 call/s round-robin
  - measurements: t1_terminal_s=[62.0], t2_terminal_s=[72.3], t3_terminal_s=[68.1], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224) · hypotheses: H-U-177, H-U-149, H-U-150, H-U-162 · evidence:
    table/cross-region/repl-same-region-collision

## Notes

H-U-177 confirmed (accepted, then 'failed' without touching the user table). H-U-149 confirmed / H-U-150
refuted (same-region table-level destination). H-U-162 refuted again: a pre-existing namespace is simply
reused ('D:completed').

Full original title: Destination name collision with a user table: Put -> 200, status 'D:failed' (A table with
an identical name already exists at the destination.); same-region destination -> 'E:completed'; pre-existing
namespace -> 'D:completed'
