<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-202: PutTableMaintenanceConfiguration merges partial values: {status} only, {settings} only, {} and settings {} all HTTP 204
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-202"></a>**S3T-TABLE-202** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration merges partial values: {status} only, {settings} only, {} and settings {} all HTTP 204**
  PutTableMaintenanceConfiguration type=icebergCompaction with value {"status":"disabled"} (no settings) ->
  HTTP 204, map after {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 64,
  "strategy": "auto"}}}. {"status":"enabled"} only -> HTTP 204, map after {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 64, "strategy": "auto"}}}.
  {"settings":{"icebergCompaction":{"targetFileSizeMB":256}}} (no status) -> HTTP 204, map after {"status":
  "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}. value {} ->
  HTTP 204. {"status":"enabled","settings":{}} -> HTTP 204.
  {"status":"enabled","settings":{"icebergCompaction":{}}} -> HTTP 204, map after {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}.
  - ACK: terminal_codes, custom_update · ops: PutTableMaintenanceConfiguration · fields: value.status,
    value.settings
  - repro: Put compaction with each partial value (boto3 parameter validation disabled)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-295](../table-maintenance.md#s3t-table-295), [S3T-TABLE-296](../table-maintenance.md#s3t-table-296), [S3T-TABLE-298](../table-maintenance.md#s3t-table-298), [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201),
    [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-200](../table-maintenance.md#s3t-table-200), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199) · hypotheses: H-U-037, H-U-043, H-U-049 ·
    evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-037 REFUTED: settings is not conditionally required; {status:disabled} alone flips status and keeps the
stored settings (64 MB stayed), {settings} alone keeps the stored status, value {} and settings {} are
accepted as no-ops (nothing changed). H-U-043 empty-settings cell: accepted (204), not 400. H-U-049: there is
no reset input; {} does nothing.
