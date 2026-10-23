<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-054: PutTableBucketMaintenanceConfiguration merges into the stored value...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-054"></a>**S3T-TABLEBUCKET-054** `server-default` · impact high · handled · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration merges into the stored value...** (hypothesis refuted; behavior confirmed)
  All 12 Puts returned HTTP 204. Sequence (sent -> Get status/unreferencedDays/nonCurrentDays): enabled 7/20
  -> enabled/7/20; {status: disabled} (no settings) -> disabled/7/20; {status: enabled} (no settings) ->
  enabled/7/20; enabled 5/20 -> enabled/5/20; {enabled, unreferencedDays: 5 only} -> enabled/5/20; {enabled,
  nonCurrentDays: 20 only} -> enabled/5/20; enabled 7/4 -> enabled/7/4; disabled 7/4 -> disabled/7/4; enabled
  7/4 -> enabled/7/4; {status: disabled} -> disabled/7/4; {status: disabled} -> disabled/7/4; {settings 5/15,
  no status} -> disabled/5/15.
  - ACK: compare.nil_equals_zero_value, late_initialize, custom_update · ops:
    PutTableBucketMaintenanceConfiguration, GetTableBucketMaintenanceConfiguration · fields: value.status,
    value.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: Fresh bucket -> the 12 Put/Get pairs above in order (each Get immediately after its Put)
  - measurements: put_latency_ms_min_max=[28, 519], n=12
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275; pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-055](../tablebucket-maintenance-metrics.md#s3t-tablebucket-055), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](../tablebucket-maintenance-metrics.md#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](../tablebucket-maintenance-metrics.md#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](../tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](../tablebucket-maintenance-metrics.md#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075) · hypotheses: H-S-027,
    H-S-028, H-S-109, H-S-030 · evidence: tablebucket/sub-resources/maintenance-defaults-fill

## Notes

H-S-027 refuted: {status: disabled} without settings keeps the stored 7/20 (no reset to 3/10); {status:
enabled} without settings keeps 7/20 too. H-S-028 refuted: Put is a merge-patch, not a whole-value replace:
{enabled, unreferencedDays: 5} kept nonCurrentDays 20 and {enabled, nonCurrentDays: 20} kept unreferencedDays 5.
H-S-109 qualified: 'disabled 7/4' keeps 7/4 (first half confirmed) but 'disabled without settings' reads back
identically (disabled/7/4), so the two disable spellings do NOT differ (second half refuted). H-S-030 part 1
qualified: a value with settings but no status is accepted (HTTP 204) but status stays at the stored value
('disabled' here), it is not forced to 'enabled'. All Puts returned HTTP 204 (the hypotheses say 200). One
bucket, 12 Put/Get pairs; every Get read immediately after its Put.

Full original title: PutTableBucketMaintenanceConfiguration merges into the stored value: status-only keeps
the days, one day keeps the other, settings-only keeps status
