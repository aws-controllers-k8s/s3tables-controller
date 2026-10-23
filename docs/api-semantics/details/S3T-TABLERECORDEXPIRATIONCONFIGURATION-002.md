<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLERECORDEXPIRATIONCONFIGURATION-002: Record-expiration check order (customer table): enum typo -> ValidationException 400, ghost/deleted table -> 404, bad ARN -> 400, then 405
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablerecordexpirationconfiguration-002"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-002** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Record-expiration check order (customer table): enum typo -> ValidationException 400, ghost/deleted table -> 404, bad ARN -> 400, then 405**
  On a customer table (where every valid call is MethodNotAllowedException 405): Put days=0 and days=-1 are
  rejected CLIENT-SIDE by botocore (ParamValidationError 'valid min value: 1', the model's minimum) and never
  reach the service; status='Enabled' (capitalized) -> ValidationException 400 '1 validation error detected:
  Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set: [enabled,
  disabled]'. Nonexistent table id under the same bucket: GetTableRecordExpirationJobStatus ->
  NotFoundException 404 'The specified table does not exist.'; GetTableRecordExpirationConfiguration ->
  NotFoundException 404; Put -> NotFoundException 404. Malformed table ARN (non-uuid id) ->
  BadRequestException 400 'The specified table ARN is not valid.'. Right after DeleteTable (HTTP 204): job
  status -> NotFoundException 404 'The specified table does not exist.'; Put -> NotFoundException 404; config
  -> NotFoundException 404.
  - ACK: exceptions.404, terminal_codes · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, GetTableRecordExpirationConfiguration, DeleteTable · fields: tableArn,
    value.settings.days, value.status
  - repro: customer table: Put days=0/-1/'Enabled'; random table uuid; malformed ARN; DeleteTable then Get/Put
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](../table-record-expiration.md#s3t-table-169), [S3T-TABLE-170](../table-record-expiration.md#s3t-table-170), [S3T-TABLE-172](../table-record-expiration.md#s3t-table-172), [S3T-TABLE-189](../table-record-expiration.md#s3t-table-189), [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190), [S3T-TABLE-243](../table-record-expiration.md#s3t-table-243),
    [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179) · hypotheses: H-U-127 · evidence:
    tablerecordexpirationconfiguration/state-machine/job-status-soak

## Notes

Check order observed: input enum validation (ValidationException 400, a wire code that is NOT in the
operation's modeled error list) -> table existence (NotFoundException 404) -> 'allowed for this table'
(MethodNotAllowedException 405). A controller can therefore tell 'gone' (404) from 'not applicable' (405).
Single run.

Full original title: Record-expiration check order on a customer table: enum typo 'Enabled' ->
ValidationException 400 (unmodeled) BEFORE the 405; nonexistent/deleted table -> NotFoundException 404;
malformed ARN -> BadRequestException 400
