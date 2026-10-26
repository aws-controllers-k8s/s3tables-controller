<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-097: ListTables: empty bucket -> 200 tables=[] (key present, no token); deleted or never-created bucket -> NotFoundException 404; S3 ARN 400
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-097"></a>**S3T-TABLE-097** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables: empty bucket -> 200 tables=[] (key present, no token); deleted or never-created bucket -> NotFoundException 404; S3 ARN 400**
  Fresh empty bucket: ListTables -> HTTP 200, 'tables' key present: True (value []), continuationToken
  present: False; emptied bucket before deletion -> HTTP 200 with 0 tables. Right after DeleteTableBucket
  returned HTTP 204: ListTables -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; final
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  tableBucketARN='arn:aws:s3:::<bucket>' -> BadRequestException (HTTP 400) 'The specified table bucket ARN is
  not valid.'; ARN with another region -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; well-formed ARN of a never-created bucket -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'; 'not-an-arn' -> BadRequestException (HTTP 400) 'The specified table
  bucket ARN is not valid.' (client_side=False).
  - ACK: exceptions.404, terminal_codes, pre-delete-cleanup · ops: ListTables, DeleteTableBucket · fields:
    tables, tableBucketARN
  - repro: CreateTableBucket; ListTables; ... DeleteTableBucket; ListTables;
    ListTables(tableBucketARN='arn:aws:s3:::x'); ListTables(unknown bucket ARN)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-093](../table.md#s3t-table-093), [S3T-TABLE-094](../table.md#s3t-table-094), [S3T-TABLE-244](../table.md#s3t-table-244), [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLE-267](../table.md#s3t-table-267) · hypotheses: H-T-080 ·
    evidence: table/weird-inputs/list-tables-bounds

## Notes

H-T-080 confirmed (predicted: empty -> 200 with tables=[] present and no token; deleted bucket ->
NotFoundException 404; plain S3 ARN -> BadRequestException 400). Messages: deleted 'The specified bucket does
not exist.', plain-s3 'The specified table bucket ARN is not valid.', unknown 'The specified bucket does not
exist.'.

Full original title: ListTables on an empty bucket -> HTTP 200 tables=[] (key present); deleted bucket ->
NotFoundException; plain S3 ARN -> BadRequestException; unknown bucket -> NotFoundException
