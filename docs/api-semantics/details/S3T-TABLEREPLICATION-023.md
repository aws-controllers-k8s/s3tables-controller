<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-023: Explicit IAM Deny: Get/Put/Delete TableReplication -> AccessDeniedException; other-account tableArn -> ForbiddenException; foreign dest 200
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-023"></a>**S3T-TABLEREPLICATION-023** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Explicit IAM Deny: Get/Put/Delete TableReplication -> AccessDeniedException; other-account tableArn -> ForbiddenException; foreign dest 200**
  (a) Session policy with an explicit Deny on the three replication actions: GetTableReplication ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableReplication on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a session '; GetTableReplicationStatus (not denied) -> HTTP 200;
  PutTableReplication -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableReplication
  on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a session '; DeleteTableReplication -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableReplication on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a sessi'; GetTable -> HTTP 200. (b) tableArn with the account segment replaced by
  <ACCOUNT>: GetTableReplication -> ForbiddenException (HTTP 403) 'Unauthorized'; GetTableReplicationStatus ->
  ForbiddenException (HTTP 403) 'Unauthorized'; PutTableReplication -> ForbiddenException (HTTP 403)
  'Unauthorized'; DeleteTableReplication -> ForbiddenException (HTTP 403) 'Unauthorized'. A
  destinationTableBucketARN of another account (same bucket name) -> HTTP 200.
  - ACK: terminal_codes, requeue · ops: GetTableReplication, GetTableReplicationStatus, PutTableReplication,
    DeleteTableReplication · fields: tableArn, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: AssumeRole(shard role, session policy Deny) -> each op; admin -> each op with the ARN account
    replaced by <ACCOUNT>
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040),
    [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012), [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEREPLICATION-024](../table-replication.md#s3t-tablereplication-024),
    [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028) · hypotheses: H-U-184 · evidence:
    table/error-taxonomy/repl-forbidden-vs-denied

## Notes

H-U-184 confirmed: IAM deny -> ['AccessDeniedException']; foreign account identifier ->
['ForbiddenException'].

Full original title: Explicit IAM Deny: Get/Put/Delete TableReplication ->
AccessDeniedException/AccessDeniedException/AccessDeniedException; other-account tableArn ->
ForbiddenException (Get) ForbiddenException (Put); other-account destination -> 200
