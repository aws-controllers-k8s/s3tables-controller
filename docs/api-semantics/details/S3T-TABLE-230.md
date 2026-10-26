<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-230: Enums exact lowercase: status ENABLED/Enabled/on -> 400; strategy zorder/ZORDER/BINPACK/Auto -> 400; binpack ok; z-order/sort pass enum
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-230"></a>**S3T-TABLE-230** `request-validation` · impact medium · handled · verified 2026-10-13
  **Enums exact lowercase: status ENABLED/Enabled/on -> 400; strategy zorder/ZORDER/BINPACK/Auto -> 400; binpack ok; z-order/sort pass enum**
  status values: ENABLED -> BadRequestException/400; Enabled -> BadRequestException/400; enabled -> HTTP 204;
  DISABLED -> BadRequestException/400; disabled -> HTTP 204; on -> BadRequestException/400; empty ->
  BadRequestException/400. strategy values: zorder -> BadRequestException/400; ZORDER ->
  BadRequestException/400; Z-ORDER -> BadRequestException/400; z-order -> BadRequestException/400; binpack ->
  HTTP 204; BINPACK -> BadRequestException/400; Binpack -> BadRequestException/400; Auto ->
  BadRequestException/400; sort -> BadRequestException/400; SORT -> BadRequestException/400; none ->
  BadRequestException/400; empty -> BadRequestException/400; null -> HTTP 204. Messages for rejections:
  {"ENABLED": "The specified status is invalid. Value must be either 'enabled' or 'disabled.'", "Enabled":
  "The specified status is invalid. Value must be either 'enabled' or 'disabled.'", "DISABLED": "The specified
  status is invalid. Value must be either 'enabled' or 'disabled.'", "on": "The specified status is invalid.
  Value must be either 'enabled' or 'disabled.'", "empty": "The compaction strategy specified is not valid.",
  "zorder": "The compaction strategy specified is not valid.", "ZORDER": "The compaction strategy specified is
  not valid.", "Z-ORDER": "The compaction strategy specified is not valid.", "z-order": "The table metadata
  location is not specified.", "BINPACK": "The compaction strategy specified is not valid.", "Binpack": "The
  compaction strategy specified is not valid.", "Auto": "The compaction strategy specified is not valid.",
  "sort": "The table metadata location is not specified.", "SORT": "The compaction strategy specified is not
  valid.", "none": "The compaction strategy specified is not valid."}.
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration · fields: value.status,
    value.settings.icebergCompaction.strategy
  - repro: Put compaction with each status / strategy spelling (parameter_validation=False)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-297](../table-maintenance.md#s3t-table-297), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLE-226](../table-maintenance.md#s3t-table-226),
    [S3T-TABLE-227](../table-maintenance.md#s3t-table-227), [S3T-TABLE-228](../table-maintenance.md#s3t-table-228), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLEMAINTENANCECONFIGURATION-012](../table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-233](../table-maintenance.md#s3t-table-233) ·
    hypotheses: H-U-046 · evidence: table/weird-inputs/maintenance-bounds

## Notes

H-U-046 confirmed: status accepts only 'enabled'/'disabled' (400 'The specified status is invalid. Value must
be either 'enabled' or 'disabled.''); strategy accepts binpack and the hyphenated z-order (z-order passed enum
validation and failed later with 'The table metadata location is not specified.' on this metadata-less table),
while zorder, ZORDER, Z-ORDER, BINPACK, Binpack, Auto, SORT, none and '' -> 400 'The compaction strategy
specified is not valid.'. A JSON null strategy is accepted and ignored (stored strategy kept).
