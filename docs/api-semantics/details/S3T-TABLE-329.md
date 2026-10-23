<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-329: Orphan replica after DeleteTableReplication keeps managedTableInformation.replicationInformation.sourceTableARN (0-479 s, same+cross region)
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-329"></a>**S3T-TABLE-329** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Orphan replica after DeleteTableReplication keeps managedTableInformation.replicationInformation.sourceTableARN (0-479 s, same+cross region)**
  Completed table-level replicas (empty ICEBERG sources): baseline GetTable(replica) raw
  managedTableInformation = {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/e5e8835c-55bd-4fe2-b696-2097ecb5f0e8"}}.
  DeleteTableReplication(tS -> same-region DS) -> HTTP 204, DeleteTableReplication(tX -> cross-region DX) ->
  HTTP 204, DeleteTable(source tD, replica in DS) -> HTTP 204; GetTableReplicationStatus(tS) right after ->
  NotFoundException (HTTP 404). Raw GetTable(replica) at [0.0, 10.0, 60.1, 479.3] s after the deletes - tS:
  managedByService ['replication.s3tables.amazonaws.com'], managedTableInformation present [True], raw
  ['{"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/e5e8835c-55bd-4fe2-b696-2097ecb5f0e8"}}'],
  nested replicationInformation [True], top-level replicationInformation key [False], sourceTableARN == source
  [True]; tX: ['replication.s3tables.amazonaws.com'] / [True] / ['{"replicationInformation":
  {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/a0072454-62b8-4f9a-8fc5-42addb7954bc"}}']
  / [True] / [False] / [True]; tD (source deleted): ['replication.s3tables.amazonaws.com'] / [True] /
  ['{"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/b366eebe-2b77-43ed-bfdb-2e5a574c2ecd"}}']
  / [True] / [False] / [True].
  - ACK: is_read_only, pre-delete-cleanup, custom_find · ops: DeleteTableReplication, DeleteTable, GetTable,
    GetTableReplicationStatus · fields: managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, type
  - repro: table-level replication to a same-region and a cross-region bucket -> completed ->
    DeleteTableReplication (and DeleteTable(source) for a third table) -> raw GetTable(replica) at
    0/10/60/300+ s
  - measurements: sample_offsets_s=[0.0, 10.0, 60.1, 479.3], replicas=3, n=12
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196) · evidence:
    tablereplication/creative/fu2-03-05-orphan-body-dead-destination

## Notes

FU2-03: 225 holds and 307's 'replicationInformation present=False' is a key-path artifact, not a different
behavior: in the GetTable body replicationInformation is NESTED under managedTableInformation (there is no
top-level replicationInformation key, so a top-level check is always False). The orphan left by
DeleteTableReplication keeps type=customer, managedByService=replication.s3tables.amazonaws.com and
managedTableInformation={"replicationInformation": {"sourceTableARN": <source table ARN>}} byte-identical to
the pre-delete baseline on every sample at 0/10/60/479 s, for a same-region (us-west-2) and a cross-region
(us-east-1) replica alike. The DeleteTable(source) orphan (tD, as in 196) has the same body shape and still
names the deleted source's ARN on every sample. GetTableReplicationStatus(source) right after
DeleteTableReplication -> NotFoundException 404. A controller adopting a replica can read the source ARN from
managedTableInformation.replicationInformation.sourceTableARN indefinitely after the configuration is gone; it
does NOT indicate an active replication.

Contradiction with [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196): FU2-03 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: 225 holds and 307's 'replicationInformation
present=False' is a key-path artifact, not a different behavior: in the GetTable body replicationInformation
is NESTED under managedTableInformation (there is no top-level replicationInformation key, so a top-level
check is always False). The orphan left by DeleteTableReplication keeps type=customer,
managedByService=replication.s3tables.amazonaw Resolution: keep all; [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) is canonical for the
settled rule
