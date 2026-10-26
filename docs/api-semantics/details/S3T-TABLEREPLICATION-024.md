<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-024: Caller scoped to the source bucket only (+iam:PassRole): PutTableReplication 200 (no caller perms on dest); without PassRole -> AccessDenied
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-024"></a>**S3T-TABLEREPLICATION-024** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Caller scoped to the source bucket only (+iam:PassRole): PutTableReplication 200 (no caller perms on dest); without PassRole -> AccessDenied** (hypothesis refuted; behavior confirmed)
  Assumed-role sessions on the shard role. Session policy allowing s3tables:* only on the source bucket and
  its tables plus iam:PassRole: GetTableReplication -> HTTP 200; PutTableReplication(t -> D, current token) ->
  HTTP 200. Session policy also allowing s3tables:* on D and its tables: Put -> HTTP 200. Session policy
  allowing s3tables:* everywhere but NO iam:PassRole: Put -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> because no session policy allows the iam:PassRole action';
  GetTableReplication -> HTTP 200.
  - ACK: references, terminal_codes · ops: PutTableReplication, GetTableReplication · fields:
    configuration.role, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: AssumeRole with a session policy scoped to S (+PassRole) -> Put t->D; widen to D -> Put; drop
    PassRole -> Put
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023), [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLE-026](../table-policy.md#s3t-table-026) · hypotheses: H-U-185 · evidence:
    table/error-taxonomy/repl-forbidden-vs-denied

## Notes

H-U-185 REFUTED: the Put succeeded without any caller permission on the destination bucket (only the role
needs it).

Full original title: Caller scoped to the source bucket only (+PassRole): PutTableReplication -> 200; scoped
to source+destination -> 200; without iam:PassRole -> AccessDeniedException
