<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-120: TagResource over-long key/value and empty key return wire code ValidationException (HTTP 400), which the s3tables model does not declare
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-120"></a>**S3T-TABLE-120** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **TagResource over-long key/value and empty key return wire code ValidationException (HTTP 400), which the s3tables model does not declare**
  TagResource(tableARN, {<129-char key>: v}) -> HTTP 400 with wire code ValidationException and message "1
  validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy constraint:
  [Member must have length less than or equal to 128, Member must have length greater than or equal to 1]";
  {k: <257-char value>} -> ValidationException 400 (Map value ... less than or equal to 256); a raw request
  with an empty key -> ValidationException 400 (same keys constraint). Every other rejection on the tag path
  (charset, aws: prefix) is BadRequestException 400. ValidationException is not among the modeled errors of
  TagResource (BadRequestException, ConflictException, ForbiddenException, InternalServerErrorException,
  NotFoundException, TooManyRequestsException), so the SDK surfaces it as a generic, untyped error.
  - ACK: terminal_codes, tags.custom-sync · ops: TagResource · fields: tags
  - repro: TagResource(tableARN, {'k'*129: 'v'}); TagResource(tableARN, {'k': 'v'*257}); raw POST /tag/<arn>
    {tags: {'': 'v'}}
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKET-067](../tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](../table-policy.md#s3t-table-083), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-109](../table-policy.md#s3t-table-109),
    [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090) · hypotheses: H-T-092, H-T-090 ·
    evidence: table/tags/table-tag-semantics

## Notes

Side finding from H-T-092/H-T-090: terminal-error matching by typed exception (BadRequestException) would miss
these permanent user errors and requeue forever; match the wire code string 'ValidationException' too. The
SDK's own client-side validation catches the empty key (min length 1) but not the max lengths (botocore
enforces min only), so a controller does reach the server with a 129-char key.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven: the service returns the
unmodeled wire code ValidationException (HTTP 400) for permanent spec errors - TagResource on a table ARN with
a key >128 chars, a value >256 chars or an empty key ([S3T-TABLE-120](../table.md#s3t-table-120)) and PutTablePolicy with a document >20480
bytes ([S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001)) - which terminal_codes=[BadRequestException] never matches, so a Table whose
spec.tags break the length bounds (and any future TablePolicy with an oversize document) is requeued as
recoverable forever instead of reaching a Terminal condition; add ValidationException to terminal_codes
(botocore stops only the empty key client-side, -120, so the over-long values reach the server) and/or enforce
the bounds in the CRD.
