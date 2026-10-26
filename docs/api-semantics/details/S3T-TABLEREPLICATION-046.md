<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-046: DeleteTableReplication: table token and garbage token -> the same 409 message...
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-046"></a>**S3T-TABLEREPLICATION-046** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **DeleteTableReplication: table token and garbage token -> the same 409 message...**
  With the configuration still 'pending' (300 s after Put): DeleteTableReplication(versionToken=<GetTable
  token of the source table>) and DeleteTableReplication(versionToken='0123456789abcdefghij') both ->
  ConflictException 409 'Provided version token does not match the table version token.';
  PutTableReplication(same configuration, versionToken=<table token>) -> ConflictException 409 'The specified
  version token does not match the configuration.'. DeleteTableReplication(current GetTableReplication token)
  -> HTTP 204 with an empty body; GetTableReplication, GetTableReplicationStatus and a repeated Delete ->
  NotFoundException 404 'The specified replication configuration does not exist.'; GetTable(source) -> 200
  with no changed member. Immediately afterwards ListTables(destination bucket) still listed the replica table
  (same namespace and name as the source, type 'customer'); DeleteTable on it in the destination region ->
  204, then DeleteNamespace and DeleteTableBucket -> 204.
  - ACK: custom_delete, pre-delete-cleanup · ops: DeleteTableReplication, PutTableReplication,
    GetTableReplication, GetTableReplicationStatus, ListTables, DeleteTable · fields: versionToken
  - repro: PutTableReplication -> wait 300 s (still pending) -> Delete(table token) -> Delete(garbage token)
    -> Put(table token) -> Delete(current) -> Get/Status/Delete again -> GetTable -> ListTables(destination)
    -> DeleteTable(replica)
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007) · evidence: tablereplication/creative/reverify-set-e-repl

## Notes

Precision gain while re-verifying [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007) (all its codes and messages reproduced, here with
the replication still pending rather than completed): the 409 message does not distinguish a stale token from
a random one, and deleting the configuration does not remove the replica.

Full original title: DeleteTableReplication: table token and garbage token -> the same 409 message; after the
204 the replica table stays in the destination bucket (type customer) and DeleteTable removes it
