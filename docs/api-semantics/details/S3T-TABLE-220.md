<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-220: GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable: all 200, 0 stale maintenance / 0 stale policy reads (n=20)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-220"></a>**S3T-TABLE-220** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable: all 200, 0 stale maintenance / 0 stale policy reads (n=20)**
  First reads 165-244 ms after CreateTable returned (and right after the two Puts):
  GetTableMaintenanceConfiguration -> ['HTTP 200'] with (config keys, job keys) = [(2, 3)];
  GetTableMaintenanceJobStatus -> ['HTTP 200']; GetTablePolicy -> ['HTTP 200']. Stale reads (maintenance not
  yet 256 MB / policy Sid missing) at the first read: 0 / 0; later reads at ~200 ms and ~1 s showed no further
  change.
  - ACK: none · ops: GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetTablePolicy
  - repro: 20x: CreateTable -> Puts -> Gets at ~0/200/1000 ms
  - measurements: first_read_offset_ms_min=165, first_read_offset_ms_max=244, stale_maint=0, stale_policy=0,
    n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-219](../table-maintenance.md#s3t-table-219), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-067 · evidence:
    table/consistency-windows/post-create-subresources

## Notes

H-U-067 confirmed: full default map (2 keys) and 3 job keys at the first read in every trial=True.

Full original title: GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable: HTTP 200
/ HTTP 200 / HTTP 200; 0 stale maintenance, 0 stale policy reads (n=20)
