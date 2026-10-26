<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-199: Compaction strategy is always reported ('auto' on a fresh table); Put without strategy keeps the stored strategy (merge)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-199"></a>**S3T-TABLE-199** `response-fidelity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Compaction strategy is always reported ('auto' on a fresh table); Put without strategy keeps the stored strategy (merge)**
  GetTableMaintenanceConfiguration strategy on a fresh table: auto. After Put icebergCompaction {enabled,
  {targetFileSizeMB 512, strategy auto}} -> Put HTTP 204, strategy read back auto. After Put {enabled,
  {targetFileSizeMB 512}} (no strategy) -> strategy read back auto. Full compaction entries: fresh {"status":
  "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "binpack"}}}; after auto
  {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}};
  after plain {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy":
  "auto"}}}.
  - ACK: compare.nil_equals_zero_value, compare.is_ignored+delta_pre_compare · ops:
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration · fields:
    configuration.icebergCompaction.settings.icebergCompaction.strategy
  - repro: Get fresh; Put compaction {512, auto}; Get; Put compaction {512}; Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-197](../table-maintenance.md#s3t-table-197), [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-212](../table-maintenance.md#s3t-table-212), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-200](../table-maintenance.md#s3t-table-200), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201),
    [S3T-TABLE-202](../table-maintenance.md#s3t-table-202), [S3T-TABLE-203](../table-maintenance.md#s3t-table-203), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211), [S3T-TABLE-016](../table-schema.md#s3t-table-016) · hypotheses: H-U-033 ·
    evidence: table/mutation-matrix/maintenance-config

## Notes

H-U-033 refuted: 'auto' is present on a never-configured table, so writing strategy=auto does not create a
read-back difference; and omitting strategy in a later Put does not remove it (see the merge finding).

Contradiction with [S3T-TABLE-197](../table-maintenance.md#s3t-table-197): 199's 'Full compaction entries: fresh {targetFileSizeMB 128, strategy
binpack}' disagrees with 197's fresh table {512, auto} and with 199's own first sentence ('strategy on a fresh
table: auto'); the 'fresh' label in 199 denotes the baseline of the already-mutated probe table (after 203's
binpack/128 Puts), not a never-configured table Resolution: keep both; [S3T-TABLE-197](../table-maintenance.md#s3t-table-197) is canonical for
fresh-table values; read 199's 'fresh' entry as 'before this sub-experiment'
