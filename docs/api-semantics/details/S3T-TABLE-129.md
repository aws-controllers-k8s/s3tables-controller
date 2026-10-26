<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-129: After DeleteTable: TagResource/ListTagsForResource on the old ARN -> 404 from t+0, UntagResource -> 500 (2/2); re-create: new ARN, tags {}
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-129"></a>**S3T-TABLE-129** `delete-semantics` · impact high · handled · verified 2026-10-13
  **After DeleteTable: TagResource/ListTagsForResource on the old ARN -> 404 from t+0, UntagResource -> 500 (2/2); re-create: new ARN, tags {}**
  DeleteTable(versionToken=<CreateTable token>) -> HTTP 204. Immediately after: TagResource(old ARN) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; UntagResource(old ARN) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'; ListTagsForResource(old
  ARN) -> NotFoundException (HTTP 404) 'The specified table does not exist.'. ListTagsForResource polled every
  0.5 s for 4 s: statuses ['NotFoundException']. At t+5 s: NotFoundException / InternalServerErrorException /
  NotFoundException. CreateTable with the same (bucket, namespace, name) -> HTTP 200, new tableARN differs
  from the old: True; ListTagsForResource(new ARN) -> {}; ListTagsForResource(old ARN) after the re-create ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: is_arn_primary_key, tags.custom-sync, exceptions.404 · ops: DeleteTable, TagResource, UntagResource,
    ListTagsForResource, CreateTable · fields: resourceArn, tableARN, tags, versionToken
  - repro: CreateTable(tags) -> DeleteTable -> Tag/Untag/ListTags on the old ARN at t+0, every 0.5s, t+5s ->
    CreateTable same triple -> compare ARN, ListTagsForResource both ARNs
  - measurements: post_delete_list_statuses=[NotFoundException, NotFoundException, NotFoundException,
    NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException], n=8
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-066](../table-metadata-location.md#s3t-table-066), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-099](../table-metadata-location.md#s3t-table-099), [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-067](../table-metadata-location.md#s3t-table-067),
    [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-036](../table-schema.md#s3t-table-036), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193) · hypotheses: H-T-095, H-T-096 ·
    evidence: table/idempotency/untag-and-post-delete

## Notes

H-T-095 qualified: no window - ListTagsForResource and TagResource on the deleted table's ARN return
NotFoundException 404 'The specified table does not exist.' from the first call after DeleteTable returned (10
reads over 7 s); a same-triple CreateTable returns a different tableARN (new UUID) with tags {} and the old
ARN keeps answering 404. BUT UntagResource on the deleted ARN returned InternalServerErrorException (HTTP 500)
'An internal error occurred. Try again.' both at t+0 and t+5 s (2/2) instead of 404 - see the dedicated
error-code finding. H-T-096 clause: DeleteTable with the original (CreateTable-time) versionToken after tag
writes -> HTTP 204.

Full original title: After DeleteTable: TagResource/ListTagsForResource on the old ARN -> 404 from t+0, but
UntagResource -> 500 InternalServerErrorException; re-create gets a new UUID, empty tags
