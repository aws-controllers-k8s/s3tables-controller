<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-234: Put strategy sort/z-order by a role without s3tables:GetTableData -> AccessDeniedException 403 'Insufficient access...'; binpack/size ok
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-234"></a>**S3T-TABLE-234** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Put strategy sort/z-order by a role without s3tables:GetTableData -> AccessDeniedException 403 'Insufficient access...'; binpack/size ok**
  Role allowed Get/PutTableMaintenanceConfiguration + GetTable but NOT s3tables:GetTableData, on a table with
  Iceberg metadata (schema only): Put {binpack,256} -> HTTP 204; Put {targetFileSizeMB 128} -> HTTP 204; Put
  {strategy sort} -> AccessDeniedException/403 'Insufficient access to update the table maintenance
  configuration'; Put {strategy z-order} -> AccessDeniedException/403 'Insufficient access to update the table
  maintenance configuration'. After adding s3tables:GetTableData to the role: sort ->
  AccessDeniedException/403 'Insufficient access to update the table maintenance configuration' (87.9 s after
  the policy update); z-order -> AccessDeniedException/403. Final compaction config: {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "binpack"}}}.
  - ACK: terminal_codes, docs-only · ops: PutTableMaintenanceConfiguration · fields:
    value.settings.icebergCompaction.strategy
  - repro: assumed role without GetTableData: Put strategy binpack / sort / z-order; add GetTableData; Put
    sort again
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-297](../table-maintenance.md#s3t-table-297), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294), [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-SERVICE-024](../service.md#s3t-service-024),
    [S3T-SERVICE-027](../service.md#s3t-service-027), [S3T-SERVICE-011](../service.md#s3t-service-011), [S3T-SERVICE-012](../service.md#s3t-service-012), [S3T-TABLE-233](../table-maintenance.md#s3t-table-233), [S3T-TABLE-251](../table-policy.md#s3t-table-251) · hypotheses: H-U-053 ·
    evidence: table/error-taxonomy/iam-deny-subresource-codes

## Notes

H-U-053 qualified: without GetTableData the sort/z-order Put fails at Put time with AccessDeniedException 403
'Insufficient access to update the table maintenance configuration' (a service message, not the IAM 'not
authorized to perform' form) while binpack and targetFileSizeMB Puts by the same role succeed. Adding only
s3tables:GetTableData did NOT make sort succeed within 88 s (still the same 403) - the check evidently needs
more than that one action (candidates: s3tables:GetTableMetadataLocation; not tested), and the admin caller on
the same table gets 400 'requires sort order' instead. Single role, single table.

Contradiction with [S3T-TABLE-232](../table-maintenance.md#s3t-table-232): Same Put icebergCompaction {strategy sort|z-order} on a table with Iceberg
metadata (schema only, no sort order): 232 reports BadRequestException 400 'requires sort order in the table
metadata', 234 reports AccessDeniedException 403 'Insufficient access to update the table maintenance
configuration'. 234's own notes reconcile: the caller differs - the restricted role
(Get/PutTableMaintenanceConfiguration + GetTable, no GetTableData) hits the permission check first, the admin
caller on the same table gets the sort-order 400. Consequential for ACK: 400 is terminal ([GT-S3T-039](../service.md#gt-s3t-039) (controller hooks catalog entry)) while
403 is requeued, so the same user mistake is terminal or an endless requeue depending on the controller's IAM
role Resolution: keep both; [S3T-TABLE-234](../table-maintenance.md#s3t-table-234) is canonical for a restricted role, [S3T-TABLE-232](../table-maintenance.md#s3t-table-232) for a
full-permission caller; 232 over-generalizes (omits the caller), see title_fixes
