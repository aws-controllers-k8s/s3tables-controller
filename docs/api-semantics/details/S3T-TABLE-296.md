<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-296: Doc claim C054 FALSE: Put icebergCompaction {status: disabled} w/o settings keeps the stored targetFileSizeMB/strategy (merge), no reset
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-296"></a>**S3T-TABLE-296** `update-granularity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C054 FALSE: Put icebergCompaction {status: disabled} w/o settings keeps the stored targetFileSizeMB/strategy (merge), no reset**
  With 64 MB stored: Put {"status":"disabled"} (no settings) -> HTTP 204, map after {"status": "disabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 64, "strategy": "auto"}}}; {"status":"enabled"} only
  -> enabled, 64 kept; {settings 256, no status} -> status kept ([S3T-TABLE-202](../table-maintenance.md#s3t-table-202)). Put without strategy keeps
  the stored strategy ([S3T-TABLE-199](../table-maintenance.md#s3t-table-199)); even re-sending the default values leaves a previously sent 'binpack'
  in place - there is no reset path ([S3T-TABLE-205](../table-maintenance.md#s3t-table-205)).
  - ACK: custom_update, compare.is_ignored+delta_pre_compare · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: value.status, value.settings.icebergCompaction
  - repro: Settled from existing finding(s) [S3T-TABLE-202](../table-maintenance.md#s3t-table-202), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201); no new
    API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-202](../table-maintenance.md#s3t-table-202), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-295](../table-maintenance.md#s3t-table-295),
    [S3T-TABLE-298](../table-maintenance.md#s3t-table-298), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197) · evidence: service/static/doc-claims-2

## Notes

VERDICT: FALSE - partial values merge field by field; the previously stored settings survive a status-only
Put.
