<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-175: After RenameTable the pre-rename table ARN still resolves: GetTable(old ARN) name=new, tableARN unchanged; old triple -> NotFoundException
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-175"></a>**S3T-TABLE-175** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After RenameTable the pre-rename table ARN still resolves: GetTable(old ARN) name=new, tableARN unchanged; old triple -> NotFoundException**
  RenameTable(newName) -> HTTP 204. Then with the OLD (pre-rename) table ARN:
  GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405) (same as before the rename:
  the table still resolves); GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  GetTableReplication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.' (same as before); GetTable(tableArn=old ARN) -> HTTP 200 with the NEW name, tableARN identical to
  the pre-rename ARN, versionToken identical to the one returned by CreateTable.
  GetTable(bucketARN+namespace+OLD name) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(new triple) -> HTTP 200 with the same tableARN.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTable, GetTableRecordExpirationConfiguration,
    GetTableReplication · fields: tableArn, tableARN, name
  - repro: CreateTable -> RenameTable(newName) -> Get* with the old table ARN; GetTable with old and new
    triples
  - measurements: renames=1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001), [S3T-TABLE-030](../table-encryption-storage.md#s3t-table-030), [S3T-TABLE-173](../table-replication.md#s3t-table-173) · hypotheses: H-U-104 · evidence:
    table/identity/expiration-arn-forms-rename-delete

## Notes

H-U-104 qualified: the UUID-based ARN survives the rename (same ARN, GetTable by old ARN shows the new name,
old triple 404); the 'unchanged expiration configuration' part is untestable on a customer table because the
record-expiration ops answer 405 before and after the rename (the table still resolves). Rename to a new
namespace was not tried.
