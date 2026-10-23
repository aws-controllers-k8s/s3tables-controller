<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-126: UntagResource with one 129-char key anywhere in tagKeys -> ValidationException 400 (unmodeled code) and nothing is removed
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-126"></a>**S3T-TABLE-126** `request-validation` · impact medium · handled · verified 2026-10-13
  **UntagResource with one 129-char key anywhere in tagKeys -> ValidationException 400 (unmodeled code) and nothing is removed**
  Table tagged {k1:v}: UntagResource(['k1', <129-char key>]) -> ValidationException (HTTP 400) "1 validation
  error detected: Value at 'tagKeys' failed to satisfy constraint: Member must satisfy constraint: [Member
  must have length less than or equal to 128", k1 still present: True. UntagResource([<129-char key>, 'k1'])
  -> ValidationException (HTTP 400) "1 validation error detected: Value at 'tagKeys' failed to satisfy
  constraint: Member must satisfy constraint: [Member must have length less than or equal to 128", k1 still
  present: True. UntagResource(['k1', '']) -> ParamValidationError (HTTP None) 'Parameter validation
  failed:\nInvalid length for parameter tagKeys[1], value: 0, valid min length: 1' (client side: True), k1
  still present: True.
  - ACK: tags.custom-sync, terminal_codes · ops: UntagResource, ListTagsForResource · fields: tagKeys
  - repro: TagResource({k1:v}); UntagResource(['k1','x'*129]); ListTagsForResource;
    UntagResource(['x'*129,'k1']); ListTagsForResource
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-119](../table.md#s3t-table-119), [S3T-TABLE-117](../table.md#s3t-table-117), [S3T-TABLE-118](../table.md#s3t-table-118), [S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLE-125](../table.md#s3t-table-125) · hypotheses: H-T-205 ·
    evidence: table/idempotency/untag-and-post-delete

## Notes

H-T-205 qualified: whole-list validation is atomic and order-independent (k1 kept in both orders) as
predicted, but the wire code is ValidationException (HTTP 400, 'Value at tagKeys failed to satisfy constraint:
Member must have length less than or equal to 128'), not the modeled BadRequestException. An empty-string key
is caught client-side by the SDK (min length 1).
