<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-173: tableArn-keyed ops with a bucket ARN, name-based pseudo ARN or 'ns/table' -> BadRequestException 400 'The specified table ARN is not valid.'
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-173"></a>**S3T-TABLE-173** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **tableArn-keyed ops with a bucket ARN, name-based pseudo ARN or 'ns/table' -> BadRequestException 400 'The specified table ARN is not valid.'**
  All five tableArn-keyed operations (GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration(enabled, days=1),
  GetTableReplication, GetTableReplicationStatus) answered BadRequestException (HTTP 400) 'The specified table
  ARN is not valid.' for tableArn = (a) the table bucket ARN, (b) the name-based
  '<bucketARN>/table/<namespace>/<table>', (c) the bare '<namespace>/<table>'. botocore sent all three to the
  wire (it does not enforce the model's tableArn regex). For reference the real UUID-based table ARN gives:
  the three record-expiration ops -> MethodNotAllowedException (HTTP 405) 'Table record expiration
  configuration API calls to this table are not allowed'; GetTableReplication and GetTableReplicationStatus ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'
  - ACK: is_arn_primary_key, custom_find, terminal_codes · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration, GetTableReplication,
    GetTableReplicationStatus · fields: tableArn
  - repro: CreateTable -> call the five tableArn-keyed ops with the bucket ARN, a name-based pseudo ARN and a
    non-ARN (SDK client, then parameter_validation=False)
  - measurements: forms=3, ops=5, n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001), [S3T-TABLE-030](../table-encryption-storage.md#s3t-table-030), [S3T-TABLE-175](../table-replication.md#s3t-table-175), [S3T-TABLE-174](../table-replication.md#s3t-table-174), [S3T-TABLE-176](../table-replication.md#s3t-table-176), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147) · hypotheses: H-U-103 · evidence: table/identity/expiration-arn-forms-rename-delete

## Notes

H-U-103 confirmed for the three forms tried (bucket ARN, name-based pseudo ARN, non-ARN): BadRequestException
400 about the ARN, never NotFoundException. The real ARN cannot be derived from bucket+namespace+name; a
controller must read it from CreateTable/GetTable. 400 is terminal (retrying never helps). Also visible: on a
customer table without replication, GetTableReplication(Status) is a 404 whose message names the
configuration, not the table. n=1 per cell.
