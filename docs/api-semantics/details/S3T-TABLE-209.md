<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-209: GetTableMaintenanceJobStatus.icebergCompaction shows Disabled 0.2 s after Put status=disabled and Not_Yet_Run 0.2 s after re-enable
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-209"></a>**S3T-TABLE-209** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceJobStatus.icebergCompaction shows Disabled 0.2 s after Put status=disabled and Not_Yet_Run 0.2 s after re-enable**
  Put icebergCompaction {disabled, 512}; the first GetTableMaintenanceJobStatus poll 0.2 s later already
  reported icebergCompaction.status=Disabled. Put {enabled, 512}; the first poll 0.2 s later reported
  Not_Yet_Run again (the job status is configuration-driven, not run-driven). Snapshot management disabled
  earlier showed Disabled the same way.
  - ACK: custom_field, requeue · ops: PutTableMaintenanceConfiguration, GetTableMaintenanceJobStatus · fields:
    status.icebergCompaction.status
  - repro: Put compaction disabled; poll GetTableMaintenanceJobStatus 5 s for <=120 s; Put enabled; poll <=60
    s
  - measurements: disabled_visible_after_s=0.2, back_after_s=null, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-207](../table-maintenance.md#s3t-table-207), [S3T-TABLE-208](../table-maintenance.md#s3t-table-208), [S3T-TABLE-191](../table-maintenance.md#s3t-table-191) · hypotheses: H-U-056, H-U-061 · evidence:
    table/mutation-matrix/maintenance-config

## Notes

H-U-056 confirmed (both directions, immediate; single table). H-U-061 raw strings: Not_Yet_Run, Disabled.
