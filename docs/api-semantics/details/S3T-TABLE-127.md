<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-127: After removing every tag ListTagsForResource -> HTTP 200 body {"tags":{}}; never-tagged table -> {"tags":{}}...
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-127"></a>**S3T-TABLE-127** `response-fidelity` · impact high · handled · verified 2026-10-13
  **After removing every tag ListTagsForResource -> HTTP 200 body {"tags":{}}; never-tagged table -> {"tags":{}}...**
  UntagResource(all 3 keys) -> HTTP 204. Raw GET /tag/<arn> body afterwards: {"tags":{}} (SDK parsed: tags key
  present True, value {}). Never-tagged table raw body: {"tags":{}} (SDK: present True, value {}).
  TagResource({a:9}) right after the removal -> HTTP 200; ListTagsForResource -> {'a': '9'}.
  - ACK: compare.nil_equals_zero_value, tags.custom-sync · ops: UntagResource, ListTagsForResource,
    TagResource · fields: tags
  - repro: UntagResource(all keys); raw GET /tag/<tableARN>; raw GET /tag/<never-tagged tableARN>;
    TagResource({a:9}); ListTagsForResource
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-156](../tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](../tablebucket.md#s3t-tablebucket-157), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLEBUCKET-028](../tablebucket.md#s3t-tablebucket-028),
    [S3T-TABLEBUCKET-121](../tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-007](../tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLE-114](../table.md#s3t-table-114), [S3T-TABLE-125](../table.md#s3t-table-125),
    [S3T-TABLE-128](../table.md#s3t-table-128), [S3T-TABLE-116](../table.md#s3t-table-116), [S3T-TABLE-117](../table.md#s3t-table-117) · hypotheses: H-T-203 · evidence:
    table/idempotency/untag-and-post-delete

## Notes

H-T-203 confirmed (predicted: 200 with tags={} present, identical to a never-tagged table; re-add -> success
and immediate read-back).

Full original title: After removing every tag ListTagsForResource -> HTTP 200 body {"tags":{}}; never-tagged
table -> {"tags":{}}; re-adding 'a' with a new value -> HTTP 200
