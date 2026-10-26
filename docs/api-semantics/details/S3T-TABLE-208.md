<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-208: GetTableMaintenanceConfiguration immediately after Put: 0 stale reads in 20 trials (write->read gap 40-64 ms)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-208"></a>**S3T-TABLE-208** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceConfiguration immediately after Put: 0 stale reads in 20 trials (write->read gap 40-64 ms)**
  20 alternating Puts of targetFileSizeMB 128/256 each followed immediately by a Get: 0 stale reads; the Get
  returned the new value 40-64 ms after the Put started.
  - ACK: none · ops: PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration
  - repro: loop 20x: Put compaction 128|256; Get
  - measurements: trials=20, stale_reads=0, gap_ms_min=40, gap_ms_max=64, n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-207](../table-maintenance.md#s3t-table-207), [S3T-TABLE-209](../table-maintenance.md#s3t-table-209), [S3T-TABLE-191](../table-maintenance.md#s3t-table-191) · hypotheses: H-U-050 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-050 confirmed (20 trials, reads at ~0 ms only, 1 Get/s budget).
