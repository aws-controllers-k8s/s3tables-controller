<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-031: 50-tag limit is cumulative per bucket: 30+25 -> BadRequestException 400 (none applied); +20 -> 200 (50); 51st -> 400; update at 50 -> 200
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-031"></a>**S3T-TABLEBUCKET-031** `quota-limit` · impact medium · handled · verified 2026-10-13
  **50-tag limit is cumulative per bucket: 30+25 -> BadRequestException 400 (none applied); +20 -> 200 (50); 51st -> 400; update at 50 -> 200**
  Bucket created with 30 tags (count 30). TagResource(25 new keys) -> BadRequestException HTTP 400 'The
  provided tags are invalid. Number of user tags must not exceed 50', count afterwards 30. TagResource(20 new
  keys) -> HTTP 200, count 50. TagResource({k51}) -> BadRequestException HTTP 400 'The provided tags are
  invalid. Number of user tags must not exceed 50', count 50. At 50 tags TagResource({k01: updated}) -> HTTP
  200 and TagResource({k02: updated, k52: x}) -> BadRequestException HTTP 400 'The provided tags are invalid.
  Number of user tags must not exceed 50'; final count 50, k01='updated', k02='v', k52 present=False.
  - ACK: tags.custom-sync, terminal_codes · ops: CreateTableBucket, TagResource, ListTagsForResource · fields:
    tags
  - repro: CreateTableBucket(30 tags) -> TagResource(25 new) -> TagResource(20 new) -> TagResource(1 new) ->
    TagResource(existing key) -> TagResource(existing+new)
  - measurements: max_tags_observed=50, n=1
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025), [S3T-TABLEBUCKET-030](../tablebucket.md#s3t-tablebucket-030), [S3T-TABLEBUCKET-029](../tablebucket.md#s3t-tablebucket-029), [S3T-TABLEBUCKET-148](../tablebucket.md#s3t-tablebucket-148),
    [S3T-TABLEBUCKET-149](../tablebucket.md#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](../tablebucket.md#s3t-tablebucket-150), [S3T-TABLE-185](../table.md#s3t-table-185) · hypotheses: H-B-041, H-B-042 · evidence:
    tablebucket/tags/upsert-limits-chars-scope

## Notes

H-B-041 confirmed; H-B-042 refuted. The limit counts keys on the resource, so the controller's tag sync must
account for out-of-band tags.

Full original title: 50-tag limit is cumulative per bucket: 30 existing + 25 new -> BadRequestException 400
(none applied); +20 -> HTTP 200 (50); 51st -> BadRequestException 400; update of an existing key at 50 -> HTTP
200
