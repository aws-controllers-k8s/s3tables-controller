<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-055: PutTableBucketMaintenanceConfiguration value {} -> HTTP 204 but a no-op (Get keeps enabled/9/9)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-055"></a>**S3T-TABLEBUCKET-055** `unsettable-field` · impact medium · handled · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration value {} -> HTTP 204 but a no-op (Get keeps enabled/9/9)...** (hypothesis refuted; behavior confirmed)
  After Put enabled 9/9: Put value={} -> HTTP 204; Get -> enabled/9/9 (fresh-bucket body was enabled/3/10, so
  {} does NOT reset to the fresh default). Put {status: enabled, settings: {}} (empty union) ->
  ParamValidationError (HTTP None) 'Parameter validation failed:
  Must set one of the following keys for tagged unionstructure value.settings:
  icebergUnreferencedFileRemoval.'; Get -> enabled/9/9. Put {status: enabled, settings:
  {icebergUnreferencedFileRemoval: {}}} -> HTTP 204; Get -> enabled/9/9. Put with no value parameter at all ->
  ParamValidationError (client-side botocore validation).
  - ACK: custom_update, compare.nil_equals_zero_value · ops: PutTableBucketMaintenanceConfiguration · fields:
    value, value.settings, value.settings.icebergUnreferencedFileRemoval
  - repro: Put enabled 9/9 -> Put value {} -> Get -> Put {enabled, settings {}} -> Get -> Put {enabled,
    settings {icebergUnreferencedFileRemoval: {}}} -> Get -> Put without value
  - handling: handled via `pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; pkg/resource/table_bucket/hooks.go:173-195; pkg/resource/table_bucket/hooks.go:389-391`
  - related: [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-054](../tablebucket-maintenance-metrics.md#s3t-tablebucket-054), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](../tablebucket-maintenance-metrics.md#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](../tablebucket-maintenance-metrics.md#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](../tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](../tablebucket-maintenance-metrics.md#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075) · hypotheses: H-S-112,
    H-S-030 · evidence: tablebucket/sub-resources/maintenance-defaults-fill

## Notes

H-S-112 refuted: value {} is accepted (204) but does NOT reset to the fresh default enabled/3/10; the stored
enabled/9/9 is unchanged, so {} is not a 'reset'. settings {} (empty union) is rejected by botocore's
client-side tagged-union validation (ParamValidationError, no request sent) - the server cell is re-tried with
parameter validation disabled in tablebucket/weird-inputs/maintenance-validation. settings
{icebergUnreferencedFileRemoval: {}} is accepted (204) and is also a no-op. H-S-030 part 2 refuted on the
reset claim ({} does not reset) and untestable through boto3 for the 'missing value -> 400' cell (client-side
ParamValidationError).

Full original title: PutTableBucketMaintenanceConfiguration value {} -> HTTP 204 but a no-op (Get keeps
enabled/9/9); empty union settings {} rejected client-side
