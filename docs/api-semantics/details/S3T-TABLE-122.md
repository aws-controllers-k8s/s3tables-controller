<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-122: After DeleteTable (204) GetTable and GetTableMetadataLocation are 404 from the first sample (10 ms); re-create gets a new warehouseLocation
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-122"></a>**S3T-TABLE-122** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **After DeleteTable (204) GetTable and GetTableMetadataLocation are 404 from the first sample (10 ms); re-create gets a new warehouseLocation**
  DeleteTable -> HTTP 204 (129 ms). Alternating GetTable / GetTableMetadataLocation every 200 ms starting 0.01
  s after the delete response: outcomes ['NotFoundException 404']; the first GetTableMetadataLocation after
  delete -> NotFoundException 404 'The specified table does not exist.'; last HTTP 200 from
  GetTableMetadataLocation at None s and from GetTable at None s (None = never). HeadObject on the deleted
  table's metadata.json -> HTTP 200. CreateTable of the same bucket/namespace/name without metadata -> HTTP
  200; GetTableMetadataLocation then returns a DIFFERENT warehouseLocation (True), metadataLocation key
  present: False, versionToken differs from the deleted table's: True.
  - ACK: exceptions.404, hook:sdk_read_one_post_set_output · ops: DeleteTable, GetTable,
    GetTableMetadataLocation, CreateTable · fields: warehouseLocation, metadataLocation, versionToken
  - repro: CreateTable with schema -> GetTableMetadataLocation -> DeleteTable -> alternate
    GetTable/GetTableMetadataLocation every 200 ms -> CreateTable same triple -> GetTableMetadataLocation
  - measurements: delete_latency_ms=[129], poll_samples=6, poll_window_s=1.062, last_200_after_delete_s=0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](../table-metadata-location.md#s3t-table-066), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089), [S3T-TABLE-099](../table-metadata-location.md#s3t-table-099), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-067](../table-metadata-location.md#s3t-table-067),
    [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-028](../table-metadata-location.md#s3t-table-028) · hypotheses: H-T-211, H-T-212 · evidence:
    table/idempotency/metadata-location-after-delete

## Notes

H-T-211 confirmed: the first GetTableMetadataLocation after delete (0.22 s) -> NotFoundException 404 'The
specified table does not exist.' (same wire code and message as the missing-table case); CreateTable of the
same bucket/namespace/name -> 200 with a different warehouseLocation (new UUID host), new versionToken and no
metadataLocation key. H-T-212 REFUTED: no divergence between the two read paths - all 6 alternating samples
from 0.01 s to 1.06 s after the delete response were 404 (polling stopped after 6 consecutive 404s). Side
observation: S3 HeadObject on the deleted table's metadata.json still returned 200 ~1 s after the delete. n=1
delete cycle.

Full original title: After DeleteTable (204) both GetTable and GetTableMetadataLocation are 404 from the first
sample (10 ms); recreate of the same triple gets a new warehouseLocation and no metadataLocation

Re-verification CORRECTED (2026-10-15, independent probe table/creative/reverify-set-b): API clauses reproduce
with n=2: DeleteTable 204 (154/146 ms); GetTable and GTML -> NotFoundException 404 from the first sample at
0.009 s (6/6 per cycle, never 200); same-triple re-create -> 200, new tableARN, different warehouseLocation
and versionToken, no metadataLocation key. Differs: HeadObject on the deleted table's metadata.json is NOT 200 -
it is 403 Forbidden from 0.03 s on in 2/2 tables never read before the delete, and 200 only until the 10 s
sample (403 from 20 s) in 2/2 tables read once before the delete; new finding recorded with related
[S3T-TABLE-122](../table-metadata-location.md#s3t-table-122).
