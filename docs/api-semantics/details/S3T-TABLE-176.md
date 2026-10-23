<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-176: After DeleteTable all four tableArn-keyed Gets return NotFoundException 404 'The specified table does not exist.' on the first call (<0.1 s)
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-176"></a>**S3T-TABLE-176** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable all four tableArn-keyed Gets return NotFoundException 404 'The specified table does not exist.' on the first call (<0.1 s)**
  Customer table, no replication. Before DeleteTable: GetTableRecordExpirationConfiguration and
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405; GetTableReplication and
  GetTableReplicationStatus -> NotFoundException 404 'The specified replication configuration does not
  exist.'. DeleteTable -> HTTP 204. First call per op right after the delete returned (0.01 / 0.03 / 0.04 /
  0.05 s): all four -> NotFoundException (HTTP 404) 'The specified table does not exist.'; no 200 or 405 was
  seen after the delete (one call each sufficed, so no 1/s polling followed).
  - ACK: exceptions.404, pre-delete-cleanup · ops: DeleteTable, GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, GetTableReplication, GetTableReplicationStatus · fields: tableArn
  - repro: CreateTable -> DeleteTable -> the four Gets immediately and every 1 s for up to 30 s
  - measurements: first_404_s.GetTableRecordExpirationConfiguration=0.01,
    first_404_s.GetTableRecordExpirationJobStatus=0.03, first_404_s.GetTableReplication=0.04,
    first_404_s.GetTableReplicationStatus=0.05, n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-173](../table-replication.md#s3t-table-173), [S3T-TABLE-174](../table-replication.md#s3t-table-174), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147) · hypotheses: H-U-105 · evidence:
    table/identity/expiration-arn-forms-rename-delete

## Notes

H-U-105 confirmed for the four Gets (no stale window; first call < 0.1 s). Note the replication Gets are 404
both before and after the delete and only the message changes ('replication configuration does not exist' ->
'table does not exist'), so for those two ops a controller cannot use the code alone to tell 'config absent'
from 'table gone'. Single table, one deletion.
