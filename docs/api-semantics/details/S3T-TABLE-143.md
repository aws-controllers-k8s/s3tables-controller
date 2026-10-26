<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-143: Tag ops on a table ARN with a random UUID -> list/tag NotFoundException 404 but untag InternalServerError 500; malformed ARN forms -> 400
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-143"></a>**S3T-TABLE-143** `error-code` · impact medium · handled · verified 2026-10-13
  **Tag ops on a table ARN with a random UUID -> list/tag NotFoundException 404 but untag InternalServerError 500; malformed ARN forms -> 400**
  With a real table ARN <bucketARN>/table/<lowercase uuid> as the base, ListTagsForResource /
  TagResource({probe}) / UntagResource(['t']) returned: random_uuid: list NotFoundException (HTTP 404) 'The
  specified table does not exist.', tag NotFoundException (HTTP 404) 'The specified table does not exist.',
  untag InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.';
  wrong_service_plain_s3: list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.'; truncated_after_table: list BadRequestException (HTTP 400) 'The
  specified resourceArn is not valid.', tag BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.', untag BadRequestException (HTTP 400) 'The specified resourceArn is not valid.'; uuid_uppercased:
  list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400) 'The specified resourceArn
  is not valid.'; trailing_slash: list BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.', tag BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.'; other_region: list
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.', tag
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.', untag
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  tables_plural_segment: list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.'. The real table's tags afterwards: {'t': '1'} (t intact: True).
  - ACK: exceptions.404, terminal_codes, is_arn_primary_key · ops: ListTagsForResource, TagResource,
    UntagResource · fields: resourceArn
  - repro: Take the tableARN; call the three tag ops with the UUID replaced by uuid4(),
    'arn:aws:s3:::<bucket>', ARN cut after '/table/', UUID upper-cased, ARN + '/', region swapped, '/tables/'
    segment
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-263](../table.md#s3t-table-263), [S3T-TABLE-264](../table.md#s3t-table-264), [S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLE-113](../table.md#s3t-table-113), [S3T-TABLE-114](../table.md#s3t-table-114), [S3T-TABLE-125](../table.md#s3t-table-125),
    [S3T-TABLEBUCKET-032](../tablebucket.md#s3t-tablebucket-032), [S3T-TABLE-130](../table.md#s3t-table-130), [S3T-TABLE-145](../table.md#s3t-table-145), [S3T-TABLE-267](../table.md#s3t-table-267), [S3T-TABLE-027](../table.md#s3t-table-027), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLE-244](../table.md#s3t-table-244), [S3T-TABLE-097](../table.md#s3t-table-097) · hypotheses: H-T-086, H-T-207 · evidence: table/identity/tag-arn-forms

## Notes

H-T-086 qualified: see codes (predicted: unknown UUID 404; wrong-service and truncated 400; upper-cased UUID
404 without case folding). H-T-207 clauses upper-cased UUID -> 404 and trailing slash -> 400: qualified: upper
BadRequestException, slash BadRequestException. Messages: random uuid 'The specified table does not exist.';
plain s3 'The specified resourceArn is not valid.'; truncated 'The specified resourceArn is not valid.'; upper
'The specified resourceArn is not valid.'; slash 'The specified resourceArn is not valid.'.

Full original title: Tag ops by ARN form (list/tag/untag): random UUID
NotFoundException/NotFoundException/InternalServerErrorException; plain S3 ARN
BadRequestException/BadRequestException/BadRequestException; truncated
BadRequestException/BadRequestException/BadRequestException; upper-cased UUID
BadRequestException/BadRequestException/BadRequestException; trailing slash
BadRequestException/BadRequestException/BadRequestException
