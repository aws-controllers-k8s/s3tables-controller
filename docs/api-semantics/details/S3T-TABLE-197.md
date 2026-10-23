<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-197: Fresh table GetTableMaintenanceConfiguration: compaction enabled/512 MB strategy auto; snapshots enabled 1/120h...
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-197"></a>**S3T-TABLE-197** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table GetTableMaintenanceConfiguration: compaction enabled/512 MB strategy auto; snapshots enabled 1/120h...**
  Table created in a bucket whose bucket-level icebergUnreferencedFileRemoval was set to status=disabled,
  unreferencedDays=7, nonCurrentDays=5 BEFORE the table existed. GetTableMaintenanceConfiguration 619 ms after
  the CreateTable request was issued (includes CreateTable latency) -> HTTP 200; the full configuration map is
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}. Nothing from the
  bucket-level config appears in the table-level map.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: GetTableMaintenanceConfiguration,
    CreateTable, PutTableBucketMaintenanceConfiguration · fields: configuration,
    configuration.icebergCompaction.settings.icebergCompaction.strategy, targetFileSizeMB, minSnapshotsToKeep,
    maxSnapshotAgeHours
  - repro: PutTableBucketMaintenanceConfiguration(disabled, 7/5) -> CreateTable ->
    GetTableMaintenanceConfiguration
  - measurements: first_get_after_create_ms=619, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-295](../table-maintenance.md#s3t-table-295), [S3T-TABLE-296](../table-maintenance.md#s3t-table-296), [S3T-TABLE-298](../table-maintenance.md#s3t-table-298), [S3T-TABLE-201](../table-maintenance.md#s3t-table-201), [S3T-TABLE-202](../table-maintenance.md#s3t-table-202),
    [S3T-TABLE-205](../table-maintenance.md#s3t-table-205), [S3T-TABLE-160](../table-maintenance.md#s3t-table-160), [S3T-TABLE-198](../table-maintenance.md#s3t-table-198), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-183](../table-maintenance.md#s3t-table-183), [S3T-TABLE-211](../table-maintenance.md#s3t-table-211), [S3T-TABLE-199](../table-maintenance.md#s3t-table-199),
    [S3T-TABLE-212](../table-maintenance.md#s3t-table-212) · hypotheses: H-U-031, H-U-032, H-U-033, H-U-067 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-031 confirmed for the values (512; 1/120; both enabled; no bucket inheritance; no
icebergUnreferencedFileRemoval key) but the compaction strategy IS present as 'auto' on a never-configured
table. H-U-032 refuted (full map, not empty/404). H-U-033 fresh-table half REFUTED: strategy is reported as
'auto', not omitted. H-U-067 maintenance-config read at ~0.6 s -> 200 (single sample; 20 trials in
post-create-subresources).

Full original title: Fresh table GetTableMaintenanceConfiguration: compaction enabled/512 MB strategy auto;
snapshots enabled 1/120h; unreferencedFileRemoval key absent

Contradiction with [S3T-TABLE-212](../table-maintenance.md#s3t-table-212): 212's behavior says a never-configured table returns compaction with
'strategy absent' and that an explicit Put 'makes auto appear permanently'; 197 (direct
GetTableMaintenanceConfiguration 619 ms after CreateTable, full map) reads strategy 'auto' on the fresh table,
as do 199 ('auto is present on a never-configured table') and 205 (fresh='auto'). 212 is a derived finding
(repro: 'derived from the fresh-table defaults and strategy findings', confidence medium) that restated the
refuted hypothesis H-U-033 instead of the observation Resolution: keep both; [S3T-TABLE-197](../table-maintenance.md#s3t-table-197) is canonical for
the fresh-table strategy value ('auto' present); 212 remains valid for its design claim (no reset input, nil
spec = unmanaged)

Contradiction with [S3T-TABLE-199](../table-maintenance.md#s3t-table-199): 199's 'Full compaction entries: fresh {targetFileSizeMB 128, strategy
binpack}' disagrees with 197's fresh table {512, auto} and with 199's own first sentence ('strategy on a fresh
table: auto'); the 'fresh' label in 199 denotes the baseline of the already-mutated probe table (after 203's
binpack/128 Puts), not a never-configured table Resolution: keep both; [S3T-TABLE-197](../table-maintenance.md#s3t-table-197) is canonical for
fresh-table values; read 199's 'fresh' entry as 'before this sub-experiment'
