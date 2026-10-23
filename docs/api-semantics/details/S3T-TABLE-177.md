<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-177: GetTable(source) is unchanged by PutTableReplication (no field changed) and DeleteTableReplication (changed: ['modifiedAt'])...
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-177"></a>**S3T-TABLE-177** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTable(source) is unchanged by PutTableReplication (no field changed) and DeleteTableReplication (changed: ['modifiedAt'])...**
  GetTable(source) before the first Put, after the Put (+status populated) and after DeleteTableReplication:
  fields differing before->after Put: none; after Put->after Delete: ['modifiedAt']. managedByService=None and
  managedTableInformation=None on the source throughout.
  - ACK: custom_find, synced.when · ops: GetTable, PutTableReplication, DeleteTableReplication · fields:
    versionToken, modifiedAt, managedByService, managedTableInformation
  - repro: GetTable -> PutTableReplication -> GetTable -> ... -> DeleteTableReplication -> GetTable; diff
    bodies
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299) ·
    hypotheses: H-U-188 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-188 qualified: versionToken, managedByService and managedTableInformation never changed; modifiedAt was
identical before and after the Put but had moved (to a timestamp ~0.2 s before the DeleteTableReplication
call, i.e. while the same-region destination E was completing its first replication) when read after the
Delete - at n=1 the trigger (the Delete or the replication to E completing) is ambiguous. A controller should
not rely on modifiedAt being stable across replication activity.
