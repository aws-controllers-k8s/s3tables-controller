<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-125: UntagResource: absent key -> 204, mixed list removes the present key, tagKeys=[] -> BadRequestException 400; TagResource({}) -> 200 no-op
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-125"></a>**S3T-TABLE-125** `idempotency` · impact high · handled · verified 2026-10-13
  **UntagResource: absent key -> 204, mixed list removes the present key, tagKeys=[] -> BadRequestException 400; TagResource({}) -> 200 no-op**
  On a table tagged {a:1,b:2}: UntagResource(['zzz']) -> HTTP 204; UntagResource(['a','zzz']) -> HTTP 204, map
  afterwards {'b': '2'}. UntagResource(tagKeys=[]) -> BadRequestException (HTTP 400) 'The provided tags are
  invalid.' (client side: False; the SDK sends no tagKeys query parameter). TagResource(tags={}) -> HTTP 200.
  UntagResource(['a','a']) -> HTTP 204 (a removed: True). UntagResource with 50 keys of which only 'a' exists
  -> HTTP 204 (a removed: True); with 51 keys -> HTTP 204 (a removed: True).
  - ACK: tags.custom-sync, terminal_codes · ops: UntagResource, TagResource, ListTagsForResource · fields:
    tagKeys, tags
  - repro: UntagResource(['zzz']); UntagResource(['a','zzz']); UntagResource([]); TagResource({});
    UntagResource(['a','a']); UntagResource(['a', 49 absent keys]); UntagResource(51 keys)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-263](../table.md#s3t-table-263), [S3T-TABLE-264](../table.md#s3t-table-264), [S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-114](../table.md#s3t-table-114), [S3T-TABLEBUCKET-032](../tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLEBUCKET-153](../tablebucket.md#s3t-tablebucket-153), [S3T-TABLEBUCKET-154](../tablebucket.md#s3t-tablebucket-154), [S3T-TABLEBUCKET-028](../tablebucket.md#s3t-tablebucket-028), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023),
    [S3T-TABLEBUCKET-155](../tablebucket.md#s3t-tablebucket-155), [S3T-TABLE-116](../table.md#s3t-table-116), [S3T-TABLE-127](../table.md#s3t-table-127), [S3T-TABLE-128](../table.md#s3t-table-128), [S3T-TABLE-117](../table.md#s3t-table-117), [S3T-TABLE-119](../table.md#s3t-table-119),
    [S3T-TABLE-126](../table.md#s3t-table-126), [S3T-TABLE-118](../table.md#s3t-table-118) · hypotheses: H-T-093, H-T-204 · evidence:
    table/idempotency/untag-and-post-delete

## Notes

H-T-093 qualified: absent-key 204 and mixed removal confirmed; UntagResource([]) -> BadRequestException 400
'The provided tags are invalid.' confirmed (the SDK sends no tagKeys query parameter at all); but
TagResource(tags={}) is ACCEPTED with HTTP 200 (predicted 400) - an empty-diff TagResource is a harmless
no-op. H-T-204 confirmed: duplicates 204 + removal, 50-key list with one existing key 204 + removal; a 51-key
list is also accepted (204), so the 50-tag quota does not cap tagKeys.

Full original title: UntagResource: absent key -> 204; mixed list removes the present key; tagKeys=[] ->
BadRequestException 400; TagResource(tags={}) -> HTTP 200 (accepted no-op)
