<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001: PutTableBucketMaintenanceConfiguration silently clamps nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketmaintenanceconfiguration-001"></a>**S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001** `request-validation` · impact medium · partially handled · verified 2026-10-14
  **PutTableBucketMaintenanceConfiguration silently clamps nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204**
  Client validation off; baseline unreferencedDays=7/nonCurrentDays=11 re-Put before every attempt.
  nonCurrentDays (unreferencedDays=1): 2147483648 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3);
  4294967296 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3); 9223372036854775807 -> Put HTTP 204, Get
  shows clamped-to-2147483647 (n=3). unreferencedDays (nonCurrentDays=1): 2147483648 -> Put HTTP 204, Get
  shows clamped-to-2147483647 (n=3); 4294967296 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3);
  9223372036854775807 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3). Messages: none. Wire body of one
  attempt: {"value": {"status": "enabled", "settings": {"icebergUnreferencedFileRemoval": {"nonCurrentDays":
  9223372036854775807, "unreferencedDays": 1}}}}.
  - ACK: compare.is_ignored+delta_pre_compare, terminal_codes · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields:
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays,
    value.settings.icebergUnreferencedFileRemoval.unreferencedDays
  - repro: PutTableBucketMaintenanceConfiguration(type=icebergUnreferencedFileRemoval,
    value.settings.%s=2147483648|4294967296|9223372036854775807, other=1) with parameter validation disabled
    -> GetTableBucketMaintenanceConfiguration; x3 per cell
  - measurements: n_per_cell=3, cells=18, values=[2147483648, 4294967296, 9223372036854775807], n=3
  - handling: partially handled via `pkg/resource/table_bucket/hooks.go:31-40; pkg/resource/table_bucket/hooks.go:376-388` - see Handling gaps
  - related: [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075), [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](../tablebucket-maintenance-metrics.md#s3t-tablebucketmaintenanceconfiguration-002) · evidence:
    tablebucket/weird-inputs/fu-04-maintenance-int-overflow

## Notes

FU-04. The modeled range is 1..2147483647 (int32); a Go int64 spec field would let a user send these values
and the controller must map the wire code to a terminal condition. See extra cells (0, -1, 2^63, 2^64, 1.5) in
the companion finding.

Partially handled in controller via: [GT-S3T-019](../service.md#gt-s3t-019) (controller hooks catalog entry): Hand-written maintenanceDaysToInt32 returns
ackerr.NewTerminalError on values outside 0..MaxInt32 for unreferencedDays/nonCurrentDays; the generator emits
math. (scorer verdict: partial match)
