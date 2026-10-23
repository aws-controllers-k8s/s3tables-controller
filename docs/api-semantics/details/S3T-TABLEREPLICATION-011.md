<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-011: PutTableReplication (client validation off): rules=[], destinations=[], no rules, two rules -> ValidationException; no role -> BadRequest
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-011"></a>**S3T-TABLEREPLICATION-011** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication (client validation off): rules=[], destinations=[], no rules, two rules -> ValidationException; no role -> BadRequest**
  With botocore parameter validation disabled (the SDK itself rejects these shapes: rules min/max 1,
  destinations min 1 max 5): configuration.rules=[] -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'configuration.rules' failed to satisfy constraint: Member must have length greater than
  or equal to 1'; rules=[{destinations: []}] -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length
  greater than or equal to 1'; configuration without rules -> ValidationException (HTTP 400) '1 validation
  error detected: Value at 'configuration.rules' failed to satisfy constraint: Member must not be null';
  configuration without role -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn
  alone OR all of (table bucket arn, namespace, table name) together'; two rules [{D}, {E}] ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'configuration.rules' failed to
  satisfy constraint: Member must have length less than or equal to 1'. GetTableReplication after the rejected
  calls -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (nothing
  stored).
  - ACK: terminal_codes, custom_delete · ops: PutTableReplication, GetTableReplication · fields:
    configuration.rules, configuration.rules[].destinations, configuration.role
  - repro: Fresh table; PutTableReplication with each malformed configuration (parameter_validation=False);
    GetTableReplication after
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012), [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003) · hypotheses: H-U-143, H-U-144 · evidence:
    table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-143 confirmed: the empty-list shapes are rejected with ['ValidationException']. 'Remove all destinations'
cannot be expressed through Put; the CRD needs MinItems=1. H-U-144 confirmed: two rules -> ValidationException
(HTTP 400) '1 validation error detected: Value at 'configuration.rules' failed to satisfy constraint: Member
must have length less than or equal to 1'.

Full original title: PutTableReplication rules=[] -> ValidationException; destinations=[] ->
ValidationException; rules omitted -> ValidationException; two rules -> ValidationException (server side,
client validation off)
