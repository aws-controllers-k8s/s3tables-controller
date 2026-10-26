<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-185: Table tag limit is 50 user tags cumulative, enforced atomically (BadRequestException 400, nothing applied); updating a key at 50 is allowed
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-185"></a>**S3T-TABLE-185** `quota-limit` · impact high · handled · verified 2026-10-13
  **Table tag limit is 50 user tags cumulative, enforced atomically (BadRequestException 400, nothing applied); updating a key at 50 is allowed**
  TagResource with 50 new keys on a table that already had 1 tag (51 cumulative) -> BadRequestException (HTTP 400)
  'The provided tags are invalid. Number of user tags must not exceed 50' (1 keys afterwards); then 49 new
  keys (-> 50 cumulative) -> HTTP 200, 50 keys. With 50 keys on the table: TagResource({new key}) ->
  BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user tags must not exceed 50', 50
  keys afterwards; TagResource({existing key: new value}) -> HTTP 200, value afterwards 'changed'; one
  TagResource mixing an existing-key update with a new key -> BadRequestException (HTTP 400) 'The provided
  tags are invalid. Number of user tags must not exceed 50', 50 keys afterwards, existing key value 'v'.
  TagResource with 51 keys on an untagged table -> BadRequestException (HTTP 400) 'The provided tags are
  invalid. Number of user tags must not exceed 50', 0 keys afterwards. CreateTable with 51 tags ->
  BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user tags must not exceed 50';
  GetTable on that name -> NotFoundException (HTTP 404); ListTables(prefix) found 0 table(s). UntagResource
  with 49 keys in one call -> HTTP 204, 0 non-run keys left.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, UntagResource, ListTagsForResource, CreateTable,
    GetTable · fields: tags, tagKeys
  - repro: CreateTable(1 tag); TagResource(50 keys); TagResource(1 new); TagResource(1 existing);
    TagResource(51 keys) on an untagged table; CreateTable(51 tags)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-263](../table.md#s3t-table-263), [S3T-TABLE-264](../table.md#s3t-table-264), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-114](../table.md#s3t-table-114), [S3T-TABLE-125](../table.md#s3t-table-125), [S3T-TABLEBUCKET-032](../tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLEBUCKET-148](../tablebucket.md#s3t-tablebucket-148), [S3T-TABLEBUCKET-149](../tablebucket.md#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](../tablebucket.md#s3t-tablebucket-150), [S3T-TABLEBUCKET-031](../tablebucket.md#s3t-tablebucket-031),
    [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025), [S3T-TABLE-119](../table.md#s3t-table-119), [S3T-TABLE-117](../table.md#s3t-table-117), [S3T-TABLE-126](../table.md#s3t-table-126), [S3T-TABLE-118](../table.md#s3t-table-118) · hypotheses: H-T-087,
    H-T-088 · evidence: table/limits/tag-count-and-read-bursts

## Notes

H-T-087 confirmed: with 50 tags (the harness 'ackq-run' tag counts as a user tag) TagResource({new}) ->
BadRequestException 400 'The provided tags are invalid. Number of user tags must not exceed 50' and nothing
applied; TagResource({existing: new value}) -> 200 and the value changed; a single call mixing one update and
one new key is rejected as a whole (the update is NOT applied). 51 keys on an untagged table -> the same 400,
0 keys applied. H-T-088 confirmed: CreateTable with 51 tags -> the same BadRequestException 400, GetTable ->
NotFoundException 404 'The specified table does not exist.', ListTables(prefix) empty (atomic, no untagged
table left). UntagResource accepted 49 keys in one call (204).

Full original title: Table tag limit is 50 user tags cumulative, enforced atomically with BadRequestException
400; updating an existing key at the limit is allowed
