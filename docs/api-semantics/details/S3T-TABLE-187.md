<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-187: 30 concurrent table reads x2 (GetTable, ListTables, GetTableEncryption+StorageClass, GetTablePolicy+Maintenance): 0 TooManyRequests
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-187"></a>**S3T-TABLE-187** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **30 concurrent table reads x2 (GetTable, ListTables, GetTableEncryption+StorageClass, GetTablePolicy+Maintenance): 0 TooManyRequests**
  Bursts of 30 concurrent reads (20 threads) against one table, two runs 20 s apart: GetTable -> {"200": 30} /
  {"200": 30} (wall 0.16 / 0.15 s); ListTables -> {"200": 30} / {"200": 30} (wall 0.10 / 0.08 s);
  GetTableEncryption+GetTableStorageClass -> {"200": 30} / {"200": 30} (wall 0.05 / 0.05 s);
  GetTablePolicy+GetTableMaintenanceConfiguration -> {"NotFoundException/404": 15, "200": 15} /
  {"NotFoundException/404": 15, "200": 15} (wall 0.13 / 0.11 s); GetTableRecordExpirationJobStatus ->
  {"ParamValidationError/None": 30} / {"ParamValidationError/None": 30} (wall 0.00 / 0.00 s). 30 ListTables at
  10 TPS -> {"200": 30}. GetTablePolicy on a table with no policy -> NotFoundException 404 'The specified
  table policy does not exist.' (15 of 15 per run); every other read returned 200.
  - ACK: requeue, e2e-timing · ops: GetTable, ListTables, GetTableEncryption, GetTableStorageClass,
    GetTablePolicy, GetTableMaintenanceConfiguration
  - repro: 20 threads x 30 reads per operation group on one table; repeat after 20 s; 30 sequential job-status
    reads; 30 ListTables at 10 TPS
  - measurements: burst_wall_s.GetTable=[0.16, 0.15], burst_wall_s.ListTables=[0.1, 0.08],
    burst_wall_s.GetTableEncryption+GetTableStorageClass=[0.05, 0.05],
    burst_wall_s.GetTablePolicy+GetTableMaintenanceConfiguration=[0.13, 0.11],
    latency_ms_min_max.GetTable=[[9,149], [10,125]], latency_ms_min_max.ListTables=[[12,80], [11,59]],
    latency_ms_min_max.GetTableEncryption+GetTableStorageClass=[[6,40], [6,37]],
    latency_ms_min_max.GetTablePolicy+GetTableMaintenanceConfiguration=[[6,112], [6,90]], n=270
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-003) · hypotheses: H-T-045, H-T-081, H-T-187, H-U-065 ·
    evidence: table/limits/tag-count-and-read-bursts

## Notes

Read halves of H-T-045/H-T-081/H-T-187/H-U-065 refuted at the 30-concurrent scale (no 429 in 8 bursts of 30
plus 30 ListTables at 10 TPS; the 100-200-call cells are untested per the shard's burst cap). Burst wall times
0.05-0.16 s, max latency 149 ms. The GetTableRecordExpirationJobStatus calls of this probe never reached AWS
(wrong parameters, client-side ParamValidationError) - H-U-130 is covered by a later probe. 2-5 other probes
running.
