<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-323: Re-verification S3T-TABLE-122: DeleteTable 204 and 404s from the first sample hold (n=2); the old metadata.json is NOT readable after delete
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-323"></a>**S3T-TABLE-323** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Re-verification [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122): DeleteTable 204 and 404s from the first sample hold (n=2); the old metadata.json is NOT readable after delete**
  Two delete cycles on one bucket/namespace/name (cycle 1: table created with a schema; cycle 2: the
  re-created table without metadata): DeleteTable -> HTTP 204 (154, 146 ms); alternating
  GetTable/GetTableMetadataLocation every 0.5 s for 3 s starting 0.009 s after the 204 -> NotFoundException
  (HTTP 404) 'The specified table does not exist.' in 6/6 samples per cycle (no HTTP 200 ever). CreateTable of
  the same triple without metadata -> HTTP 200 with a new tableARN; GetTableMetadataLocation then returns a
  different warehouseLocation, a different versionToken and no metadataLocation key. DEVIATION from the
  original's 'HeadObject on the deleted table's metadata.json -> HTTP 200': S3 HeadObject(metadataLocation) by
  the same caller returned 403 Forbidden (not 404) from 0.03 s after the 204 and through 30 s in 2/2 tables
  that had NOT been read before the delete; in 2/2 tables read once before the delete (HTTP 200) it kept
  returning HTTP 200 up to the 10 s sample and 403 Forbidden from the 20 s sample on. HeadObject on a live
  table's metadata.json -> HTTP 200.
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable, GetTableMetadataLocation, CreateTable ·
    fields: warehouseLocation, metadataLocation, versionToken
  - repro: CreateTable(schema) -> GTML -> DeleteTable -> alternate GetTable/GTML every 0.5 s for 3 s ->
    CreateTable same triple (no metadata) -> GTML -> DeleteTable -> poll again; separately CreateTable(schema)
    -> HeadObject(metadataLocation) -> DeleteTable -> HeadObject at 0..30 s (x2)
  - measurements: delete_latency_ms=[154, 146], first_sample_after_delete_s=[0.009, 0.009], poll_samples=[6,
    6], poll_window_s=3.0, last_200_after_delete_s=[null, null],
    headobject_first_non_200_after_delete_s=[20.068, 20.068, 0.033, 0.031],
    headobject_last_200_after_delete_s=[10.035, 10.029, -1, -1], headobject_pre_read=[true, true, false,
    false], headobject_tables_n=5, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-122](../table-metadata-location.md#s3t-table-122) · evidence: table/creative/reverify-set-b

## Notes

Independent re-verification (set B), n=2 per variant (original n=1). All API-side clauses reproduce. The
HeadObject-200 clause is at best transient: readability of the warehouse object is revoked with the table
(403, so existence is unobservable from this account); a prior read keeps it readable for 10-20 s, consistent
with a cached authorization rather than the object surviving. A controller must not rely on the old metadata
file after DeleteTable.
