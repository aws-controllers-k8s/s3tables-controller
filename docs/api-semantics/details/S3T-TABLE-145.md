<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-145: Table ARN with the bucket segment swapped to another existing bucket -> list/tag NotFoundException 404, untag 500; UUID alone resolves none
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-145"></a>**S3T-TABLE-145** `identity` · impact medium · handled · verified 2026-10-13
  **Table ARN with the bucket segment swapped to another existing bucket -> list/tag NotFoundException 404, untag 500; UUID alone resolves none** (hypothesis refuted; behavior confirmed)
  ARN = real table ARN with 'bucket/<B1>/' replaced by 'bucket/<B2>/' (B2 exists, empty): ListTagsForResource
  -> NotFoundException (HTTP 404) 'The specified table does not exist.'; TagResource({probe}) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; UntagResource(['t']) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'. Real table tags afterwards
  {'t': '1'}; B2 tags {'b2': '1'}. The service does NOT resolve a table by its UUID alone.
  - ACK: is_arn_primary_key, exceptions.404 · ops: ListTagsForResource, TagResource, UntagResource · fields:
    resourceArn
  - repro: Two buckets; table in B1 tagged t=1; build the table ARN with B2's name in the bucket segment;
    ListTagsForResource/TagResource/UntagResource on it; ListTagsForResource(real ARN)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-130](../table.md#s3t-table-130), [S3T-TABLE-143](../table.md#s3t-table-143) · hypotheses: H-T-208 · evidence: table/identity/tag-arn-forms

## Notes

H-T-208 (contrarian: UUID alone resolves the table across bucket segments) refuted: the mismatched ARN is
rejected and the real table's tags are intact. Message: 'An internal error occurred. Try again.'.

Full original title: Table ARN with the bucket segment swapped for another existing bucket: list/tag/untag ->
NotFoundException/NotFoundException/InternalServerErrorException; real table's tags untouched
