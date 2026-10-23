<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-245: Ghost table name or UUID in an other-region ARN -> 400 region mismatch (check precedes lookup); same-named triple in 2 regions = 2 tables
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-245"></a>**S3T-TABLE-245** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Ghost table name or UUID in an other-region ARN -> 400 region mismatch (check precedes lookup); same-named triple in 2 regions = 2 tables**
  us-west-2 endpoint, us-east-1 bucket ARN with a table name that exists in neither region and a tableArn with
  an all-zero UUID: GetTable(triple, ghost name), GetTableMetadataLocation(ghost name), GetTable(ghost
  tableArn) -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'. us-east-1 endpoint, us-west-2 ARNs: GetTable(triple, ghost name), GetTableMetadataLocation(ghost
  name), GetTable(ghost tableArn) -> BadRequestException (HTTP 400) 'The provided Region does not match the
  Region of the resource.'. The same-named bucket/namespace/table created in both regions are unrelated
  objects: table UUID equal = False (ARNs differ in segment(s) [3, 5]), namespaceId equal = False,
  tableBucketId equal = False.
  - ACK: terminal_codes, exceptions.404, is_arn_primary_key · ops: GetTable, GetTableMetadataLocation,
    CreateTable · fields: tableBucketARN, tableArn, name
  - repro: GetTable/GetTableMetadataLocation with a ghost name under the other region's bucket ARN;
    GetTable(tableArn) with the other region's bucket segment and a zero UUID
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses: H-T-213, H-T-215 ·
    evidence: table/cross-region/foreign-region-arn-reads-writes

## Notes

Supports H-T-213 and the refutation of H-T-215: a wrong-region ARN can never be mistaken for a deleted table
(the lookup is not reached). A controller must never map this 400 to 'recreate'.

Full original title: Nonexistent table name or UUID inside an other-region ARN -> BadRequestException (region
check runs before the table lookup); same-named triple in two regions = two tables
