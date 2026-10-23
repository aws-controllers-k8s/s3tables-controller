<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-130: UntagResource on a just-deleted table's ARN -> InternalServerErrorException 500 (2/2); TagResource/ListTagsForResource on it -> 404
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-130"></a>**S3T-TABLE-130** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **UntagResource on a just-deleted table's ARN -> InternalServerErrorException 500 (2/2); TagResource/ListTagsForResource on it -> 404**
  Table created with tags {a:1}, then DeleteTable -> 204. UntagResource(old tableARN, ['a']) immediately after
  -> HTTP 500 InternalServerErrorException 'An internal error occurred. Try again.' (latency ~320 ms); the
  same call at t+5 s -> the same 500 (2/2). TagResource(old ARN) and ListTagsForResource(old ARN) over the
  same period -> NotFoundException 404 'The specified table does not exist.' every time (12/12). The table's
  bucket still existed; the triple was re-created afterwards under a new UUID.
  - ACK: exceptions.404, terminal_codes, pre-delete-cleanup · ops: UntagResource, DeleteTable, TagResource,
    ListTagsForResource · fields: resourceArn, tagKeys
  - repro: CreateTable(tags {a:1}) -> DeleteTable -> UntagResource(old tableARN, ['a']) at t+0 and t+5s;
    compare with TagResource/ListTagsForResource on the same ARN
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLE-267](../table.md#s3t-table-267), [S3T-TABLE-053](../table.md#s3t-table-053), [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLE-145](../table.md#s3t-table-145) ·
    hypotheses: H-T-095 · evidence: table/idempotency/untag-and-post-delete

## Notes

Side finding from H-T-095, n=2 in one probe run (confidence medium until reproduced): a finalizer or tag
reconciler that untags a table whose deletion raced ahead receives a retryable-looking 500 for a permanent
condition and would retry forever; it should treat a 500 from UntagResource on a resource it cannot GetTable
as gone. Whether UntagResource returns 500 for a never-existing UUID as well is checked in
table/identity/tag-arn-forms.

Full original title: UntagResource on a just-deleted table's ARN returns InternalServerErrorException 500 'Try
again' (TagResource/ListTagsForResource return 404)
