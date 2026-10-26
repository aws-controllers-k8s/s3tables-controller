<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-297: Doc claim C055 PARTLY: strategy=sort/z-order -> 204 ONLY on a table whose metadata has a sort order; otherwise 400 (S3T-TABLE-204/232)
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-297"></a>**S3T-TABLE-297** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C055 PARTLY: strategy=sort/z-order -> 204 ONLY on a table whose metadata has a sort order; otherwise 400 ([S3T-TABLE-204](../table-maintenance.md#s3t-table-204)/232)**
  Live: CreateTable with schema + writeOrder -> HTTP 200 (metadataLocation present afterwards: True); fresh
  GetTableMaintenanceConfiguration compaction entry {"status": "enabled", "settings": {"icebergCompaction":
  {"targetFileSizeMB": 512, "strategy": "auto"}}}. Put icebergCompaction {enabled, strategy sort} -> HTTP 204,
  Get strategy afterwards 'sort'; Put {enabled, strategy z-order} -> HTTP 204, Get strategy 'z-order'; control
  Put {strategy binpack} -> HTTP 204. Existing: table without metadata -> BadRequestException 400 'The table
  metadata location is not specified.' ([S3T-TABLE-204](../table-maintenance.md#s3t-table-204)); table with schema but no sort order -> 400 'The
  specified compaction strategy requires sort order in the table metadata...' ([S3T-TABLE-232](../table-maintenance.md#s3t-table-232)); a role without
  s3tables:GetTableData -> AccessDeniedException 403 'Insufficient access to update the table maintenance
  configuration' ([S3T-TABLE-234](../table-maintenance.md#s3t-table-234)).
  - ACK: requeue, docs-only · ops: PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration,
    CreateTable · fields: value.settings.icebergCompaction.strategy, metadata.iceberg.writeOrder
  - repro: CreateTable(ICEBERG, metadata.iceberg.schema [id int, ts timestamp] + writeOrder {fields:
    [{sourceId 1, transform identity, asc, nulls-first}]}) -> PutTableMaintenanceConfiguration
    icebergCompaction {enabled, strategy sort} -> Get -> Put {strategy z-order} -> Get -> Put {strategy
    binpack}
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-204](../table-maintenance.md#s3t-table-204), [S3T-TABLE-232](../table-maintenance.md#s3t-table-232), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230), [S3T-TABLE-294](../table-maintenance.md#s3t-table-294) · evidence:
    service/static/doc-claims-2

## Notes

VERDICT: PARTLY - live: with a sort order in the Iceberg metadata both Puts returned HTTP 204 and Get read
back the strategy; the claim's unconditional 'accepted by the control plane' is false for tables without
metadata (400 'The table metadata location is not specified.') or without a sort order (400 'requires sort
order'); the GetTableData permission IS enforced for a role lacking it (403, [S3T-TABLE-234](../table-maintenance.md#s3t-table-234)), not for the
full-access caller.
