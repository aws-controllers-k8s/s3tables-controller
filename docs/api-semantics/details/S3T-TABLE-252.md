<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-252: ARN of another region on 7 ops (GetTable, GetTableBucket, GetTableEncryption, TagResource, DeleteTable, ...) -> 400 'Region does not match'
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-252"></a>**S3T-TABLE-252** `cross-region` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **ARN of another region on 7 ops (GetTable, GetTableBucket, GetTableEncryption, TagResource, DeleteTable, ...) -> 400 'Region does not match'**
  Table created in us-east-1. On the us-west-2 endpoint every call carrying the us-east-1 ARN fails with
  BadRequestException HTTP 400 'The provided Region does not match the Region of the resource.'
  (x-amzn-ErrorType BadRequestException:...s3tablebuckets/): GetTable(tableArn),
  GetTable(tableBucketARN+namespace+name), GetTableBucket, ListNamespaces, GetTableEncryption,
  ListTagsForResource(table and bucket ARN), TagResource, DeleteTable - no cross-region 200 and nothing
  changed at home (table exists, tags unchanged). The check is ARN-region-vs-endpoint, not existence: on the
  HOME endpoint (us-east-1) the same ARN with the region segment rewritten to us-west-2 or to 'xx-fake-1' also
  -> 400 'The provided Region does not match...'. The region-rewritten ARN sent to the endpoint that matches
  it (us-west-2, where no such table exists) -> NotFoundException 404 'The specified table does not exist.'
  (by tableArn and by bucketARN+names). A tableArn whose table id is not a UUID -> BadRequestException 400
  'The specified table ARN is not valid.'. ListTableBuckets(prefix=<us-east-1 bucket name>) on us-west-2 ->
  200 with 0 buckets (listing is regional).
  - ACK: is_arn_primary_key, terminal_codes · ops: GetTable, GetTableBucket, ListTagsForResource, TagResource,
    DeleteTable, GetTableEncryption, ListNamespaces · fields: tableArn, tableBucketARN, resourceArn
  - repro: CreateTableBucket/Namespace/Table in us-east-1 -> GetTable(tableArn) and the other reads on the
    us-west-2 endpoint; rewrite the ARN region segment and repeat
  - measurements: get_table_foreign_latency_ms=[31, 21], get_table_home_latency_ms=97, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-011](../table-encryption-storage.md#s3t-table-011), [S3T-TABLE-062](../table-encryption-storage.md#s3t-table-062), [S3T-TABLE-063](../table-encryption-storage.md#s3t-table-063) · hypotheses: H-T-007 · evidence:
    table/identity/foreign-region-arn-lookup

## Notes

H-T-007 confirmed: BadRequestException 400 whose message names the Region, not NotFoundException and not a
cross-region success; the same code/message applies to GetTableBucket, ListNamespaces, GetTableEncryption,
ListTagsForResource, TagResource and DeleteTable. A controller resolving by ARN must pick the client region
from the ARN; a 400 'Region does not match' is a client-side region mismatch, not an invalid spec. Single run,
2 GetTable samples.
