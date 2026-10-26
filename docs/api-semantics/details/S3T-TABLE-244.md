<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-244: CreateTable/RenameTable/DeleteTable/Tag/UntagResource with the other region's ARNs -> 400 region mismatch; nothing changed in either region
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-244"></a>**S3T-TABLE-244** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable/RenameTable/DeleteTable/Tag/UntagResource with the other region's ARNs -> 400 region mismatch; nothing changed in either region**
  us-west-2 endpoint with us-east-1 identifiers: TagResource, UntagResource, CreateTable, RenameTable,
  DeleteTable -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'. us-east-1 endpoint with us-west-2 identifiers: TagResource, UntagResource, CreateTable,
  RenameTable, DeleteTable -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.'. Verification via each region's own endpoint afterwards: tables in the namespace
  ['ackq_5d5312_tbl'] / ['ackq_5d5312_tbl'] (no new table, no rename), table tags {'ackq-run': '5d5312',
  'side': 'us-west-2'} / {'ackq-run': '5d5312', 'side': 'us-east-1'} (no foreign tag, 'side' not removed),
  GetTable(tableArn) HTTP 200 / HTTP 200.
  - ACK: terminal_codes, references · ops: CreateTable, RenameTable, DeleteTable, TagResource, UntagResource ·
    fields: tableBucketARN, resourceArn, namespace, name
  - repro: Same-named triple in both regions; from each endpoint CreateTable(new name)/RenameTable/DeleteTable
    with the other region's bucket ARN and Tag/Untag with the other region's tableArn;
    ListTables/ListTags/GetTable in each region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLE-097](../table.md#s3t-table-097), [S3T-TABLE-267](../table.md#s3t-table-267) · hypotheses: H-T-022, H-T-075 · evidence:
    table/cross-region/foreign-region-arn-reads-writes

## Notes

H-T-022 confirmed (CreateTable on the us-west-2 endpoint with a real us-east-1 bucket ARN ->
BadRequestException, no remote create, no local create); H-T-075 write part confirmed
(TagResource/UntagResource -> 400, no tag landed). State in both regions unchanged: True.

Full original title: CreateTable/RenameTable/DeleteTable/TagResource/UntagResource with the other region's
ARNs -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
nothing created, renamed, deleted or tagged in either region
