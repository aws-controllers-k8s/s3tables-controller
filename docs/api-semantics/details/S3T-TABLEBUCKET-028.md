<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-028: UntagResource absent keys -> 204, empty tagKeys -> BadRequestException 400; TagResource {} -> 200; all tags removed -> ListTags {"tags":{}}
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-028"></a>**S3T-TABLEBUCKET-028** `tag-semantics` · impact medium · handled · verified 2026-10-13
  **UntagResource absent keys -> 204, empty tagKeys -> BadRequestException 400; TagResource {} -> 200; all tags removed -> ListTags {"tags":{}}**
  UntagResource([zzz]) (key absent) -> HTTP 204; UntagResource([b, zzz]) -> HTTP 204 and ListTagsForResource
  -> {"ackq-run": "47dd85", "c": "3"} (b removed). UntagResource(tagKeys=[]) -> BadRequestException HTTP 400
  'The provided tags are invalid.'; TagResource(tags={}) -> HTTP 200; TagResource re-sent with an identical
  existing pair -> HTTP 200. After UntagResource of every remaining key (-> HTTP 204) ListTagsForResource ->
  HTTP 200 with raw body {"tags":{}}; repeating that UntagResource -> HTTP 204.
  - ACK: tags.custom-sync, compare.nil_equals_zero_value · ops: UntagResource, TagResource,
    ListTagsForResource · fields: tags, tagKeys
  - repro: UntagResource([absent]); UntagResource([]); TagResource({}); UntagResource(all keys) ->
    ListTagsForResource raw body
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-027](../tablebucket.md#s3t-tablebucket-027), [S3T-TABLEBUCKET-007](../tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-124](../tablebucket.md#s3t-tablebucket-124), [S3T-TABLEBUCKET-032](../tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-128](../table.md#s3t-table-128), [S3T-TABLEBUCKET-153](../tablebucket.md#s3t-tablebucket-153), [S3T-TABLEBUCKET-154](../tablebucket.md#s3t-tablebucket-154), [S3T-TABLE-125](../table.md#s3t-table-125), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023),
    [S3T-TABLEBUCKET-155](../tablebucket.md#s3t-tablebucket-155), [S3T-TABLE-116](../table.md#s3t-table-116), [S3T-TABLEBUCKET-156](../tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](../tablebucket.md#s3t-tablebucket-157), [S3T-TABLE-113](../table.md#s3t-table-113),
    [S3T-TABLE-127](../table.md#s3t-table-127), [S3T-TABLEBUCKET-121](../tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008) · hypotheses: H-B-043, H-B-071 · evidence:
    tablebucket/tags/upsert-limits-chars-scope

## Notes

H-B-043: confirmed (lenient for absent keys: HTTP 204; empty tagKeys: BadRequestException HTTP 400 'The
provided tags are invalid.'; empty map on the Tag side: HTTP 200). Also covers the
TagResource-re-sent-unchanged cell of H-B-071 (HTTP 200).

Full original title: UntagResource with absent keys -> HTTP 204; empty tagKeys -> BadRequestException 400;
TagResource with empty map -> HTTP 200; after removing all tags ListTagsForResource returns raw body
{"tags":{}}
