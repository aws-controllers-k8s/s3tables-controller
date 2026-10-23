<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-013: PutTableReplication role: IAM user ARN -> ValidationException; aws-cn role -> 200; bare name -> BadRequest; other-account role -> Forbidden
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-013"></a>**S3T-TABLEREPLICATION-013** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication role: IAM user ARN -> ValidationException; aws-cn role -> 200; bare name -> BadRequest; other-account role -> Forbidden**
  Fresh table: role = IAM user ARN -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.role' failed to satisfy constraint: Member must satisfy regular expression pattern:
  arn:.+:iam::[0-9]{12}:role/<PRINCIPAL>'. role = role ARN in the aws-cn partition -> HTTP 200 (echoed
  'arn:aws-cn:iam::<ACCOUNT>:role/<PRINCIPAL>'). role = bare role name -> BadRequestException (HTTP 400)
  'Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name)
  together'. role = role ARN of another account -> ForbiddenException (HTTP 403) 'Cross-account pass role is
  not allowed.'. GetTableReplication afterwards -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'. The shard role ARN carries a path (role/<PRINCIPAL>) and is echoed
  byte-identical by GetTableReplication (phase 1).
  - ACK: references, terminal_codes · ops: PutTableReplication, GetTableReplication · fields:
    configuration.role
  - repro: Fresh table: Put with each role form (parameter_validation=False); Get after each; Delete when
    accepted
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEREPLICATION-011](../table-replication.md#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023), [S3T-TABLEREPLICATION-024](../table-replication.md#s3t-tablereplication-024),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](../table-policy.md#s3t-table-026) · hypotheses: H-U-186 · evidence:
    table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-186 qualified: see behavior. The model pattern is arn:.+:iam::[0-9]{12}:role/<PRINCIPAL> (any partition).

Full original title: PutTableReplication configuration.role: IAM user ARN -> ValidationException; aws-cn role
ARN -> 200; bare name -> BadRequestException; other-account role ARN -> ForbiddenException
