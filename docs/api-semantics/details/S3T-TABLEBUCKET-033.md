<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-033: Tag ops on a table ARN with an unknown uuid: ListTags/Tag -> NotFoundException 404 but UntagResource -> InternalServerErrorException 500
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-033"></a>**S3T-TABLEBUCKET-033** `identity` · impact high · handled · verified 2026-10-13
  **Tag ops on a table ARN with an unknown uuid: ListTags/Tag -> NotFoundException 404 but UntagResource -> InternalServerErrorException 500**
  ListTagsForResource / TagResource({a:b}) / UntagResource([a]) per ARN form: s3-general-purpose-bucket ->
  BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The
  specified resourceArn is not valid.' | BadRequestException HTTP 400 'The specified resourceArn is not
  valid.'; namespace-path -> BadRequestException HTTP 400 'The specified resourceArn is not valid.' |
  BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The
  specified resourceArn is not valid.'; glue-table -> BadRequestException HTTP 400 'The specified resourceArn
  is not valid.' | BadRequestException HTTP 400 'The specified resourceArn is not valid.' |
  BadRequestException HTTP 400 'The specified resourceArn is not valid.'; table-arn-missing-uuid ->
  NotFoundException HTTP 404 'The specified table does not exist.' | NotFoundException HTTP 404 'The specified
  table does not exist.' | InternalServerErrorException HTTP 500 'An internal error occurred. Try again.';
  table-arn-real -> HTTP 200 {"tags": {"t": "1"}} | HTTP 200 | HTTP 204 ; bucket-arn-missing ->
  NotFoundException HTTP 404 'The specified bucket does not exist.' | NotFoundException HTTP 404 'The
  specified bucket does not exist.' | NotFoundException HTTP 404 'The specified bucket does not exist.';
  plain-name -> BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException
  HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The specified resourceArn
  is not valid.'; bucket-arn-trailing-slash -> BadRequestException HTTP 400 'The specified resourceArn is not
  valid.' | BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP
  400 'The specified resourceArn is not valid.'.
  - ACK: terminal_codes, exceptions.404, tags.custom-sync · ops: ListTagsForResource, TagResource,
    UntagResource · fields: resourceArn
  - repro: ListTagsForResource/TagResource/UntagResource with an S3 bucket ARN, a namespace path, a Glue table
    ARN, a table ARN with an unknown uuid, a missing bucket ARN, a plain name
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-022](../tablebucket.md#s3t-tablebucket-022), [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003),
    [S3T-TABLEBUCKET-043](../tablebucket.md#s3t-tablebucket-043), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015) · hypotheses: H-B-048 · evidence:
    tablebucket/tags/upsert-limits-chars-scope

## Notes

Confirms H-B-048 for the wrong-type forms (S3 bucket ARN, namespace path, Glue table ARN, plain name, trailing
slash -> BadRequestException 400 'The specified resourceArn is not valid.') and for a missing bucket ARN
(404). SURPRISE: UntagResource on a well-formed table ARN whose uuid does not exist returned
InternalServerErrorException HTTP 500 'An internal error occurred. Try again.' while ListTagsForResource and
TagResource on the same ARN returned NotFoundException 404 'The specified table does not exist.' (single
observation; re-verified in tablebucket/error-taxonomy/arn-forms-get-delete-tags).
