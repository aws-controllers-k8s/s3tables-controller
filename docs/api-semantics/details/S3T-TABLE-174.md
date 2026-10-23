<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-174: Ghost table UUID in tableArn -> 404 'The specified table does not exist.' on all 5 ops; ghost bucket: replication Gets say 'bucket' instead
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-174"></a>**S3T-TABLE-174** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Ghost table UUID in tableArn -> 404 'The specified table does not exist.' on all 5 ops; ghost bucket: replication Gets say 'bucket' instead**
  tableArn = <real bucket ARN>/table/<uuid4 that never existed>: GetTableRecordExpirationConfiguration ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableRecordExpirationJobStatus ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; PutTableRecordExpirationConfiguration ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableReplication -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. Same ARN shape but the table UUID of an existing table under a
  NON-existing bucket name: GetTableRecordExpirationConfiguration -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; GetTableRecordExpirationJobStatus -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; PutTableRecordExpirationConfiguration -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; GetTableReplication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'. Existing table ARN with the region segment swapped to us-east-1, sent to the us-west-2 endpoint:
  GetTableRecordExpirationConfiguration -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; GetTableRecordExpirationJobStatus -> BadRequestException (HTTP 400) 'The
  provided Region does not match the Region of the resource.'; PutTableRecordExpirationConfiguration ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  GetTableReplication -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'; GetTableReplicationStatus -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'.
  - ACK: exceptions.404, terminal_codes · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration, GetTableReplication,
    GetTableReplicationStatus · fields: tableArn
  - repro: Call the five ops with <bucketARN>/table/<random uuid>, with a non-existing bucket name, and with a
    foreign region segment
  - measurements: n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-173](../table-replication.md#s3t-table-173), [S3T-TABLE-176](../table-replication.md#s3t-table-176), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147) · hypotheses: H-U-109, H-U-102 ·
    evidence: table/identity/expiration-arn-forms-rename-delete

## Notes

H-U-109 confirmed: Put and GetJobStatus (and the other three) -> NotFoundException 404 with a message naming
the table ('The specified table does not exist.'); with a non-existing BUCKET name in the ARN the
record-expiration ops still say 'table does not exist' while the replication Gets say 'The specified bucket
does not exist.'. A foreign region segment -> BadRequestException 400 'The provided Region does not match the
Region of the resource.'. Compared with the 405 of an existing customer table, a controller can distinguish
'table gone' (404) from 'table exists but ineligible' (405) on these ops. n=1 per cell.
