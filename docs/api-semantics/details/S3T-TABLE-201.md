<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-201: Put compaction status=disabled with settings stores both: Get -> {"status": "disabled", "settings": {"icebergCompaction"...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-201"></a>**S3T-TABLE-201** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Put compaction status=disabled with settings stores both: Get -> {"status": "disabled", "settings": {"icebergCompaction"...**
  Put {disabled, {targetFileSizeMB 256}} -> HTTP 204, Get {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}. Put {enabled, {128}} -> HTTP 204, Get
  {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "auto"}}}. Put
  {disabled, {64}} -> HTTP 204, Get {"status": "disabled", "settings": {"icebergCompaction":
  {"targetFileSizeMB": 64, "strategy": "auto"}}}. Settings are stored and returned alongside status=disabled.
  - ACK: none · ops: PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration · fields:
    value.status, value.settings
  - repro: Put compaction disabled+256; Get; Put enabled+128; Get; Put disabled+64; Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-295](../table-maintenance.md#s3t-table-295), [S3T-TABLE-296](../table-maintenance.md#s3t-table-296), [S3T-TABLE-298](../table-maintenance.md#s3t-table-298), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-202](../table-maintenance.md#s3t-table-202),
    [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-200](../table-maintenance.md#s3t-table-200), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199) · hypotheses: H-U-036 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-036 confirmed.

Full original title: Put compaction status=disabled with settings stores both: Get -> {"status": "disabled",
"settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}; re-enable with 128 ->
{"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "auto"}}};
disabled+64 -> {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 64, "strategy":
"auto"}}}
