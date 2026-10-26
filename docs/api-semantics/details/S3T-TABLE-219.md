<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-219: PutTablePolicy / PutTableMaintenanceConfiguration 0-182 ms after CreateTable returned: HTTP 200 / HTTP 204 in 20/20 trials
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-219"></a>**S3T-TABLE-219** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy / PutTableMaintenanceConfiguration 0-182 ms after CreateTable returned: HTTP 200 / HTTP 204 in 20/20 trials**
  20 fresh tables: PutTablePolicy issued 0-0 ms after CreateTable returned -> ['HTTP 200'];
  PutTableMaintenanceConfiguration issued 108-182 ms after -> ['HTTP 204']. No ConflictException or
  NotFoundException on the new table's sub-resources. CreateTable latency 307-416 ms.
  - ACK: none · ops: CreateTable, PutTablePolicy, PutTableMaintenanceConfiguration
  - repro: 20x: CreateTable -> PutTablePolicy -> PutTableMaintenanceConfiguration -> DeleteTable
  - measurements: put_policy_offset_ms_min=0, put_policy_offset_ms_max=0, put_maint_offset_ms_min=108,
    put_maint_offset_ms_max=182, create_latency_ms_min=307, create_latency_ms_max=416, n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-220](../table-maintenance.md#s3t-table-220), [S3T-TABLEMAINTENANCECONFIGURATION-001](../table-maintenance.md#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-066 · evidence:
    table/consistency-windows/post-create-subresources

## Notes

H-U-066 confirmed.
