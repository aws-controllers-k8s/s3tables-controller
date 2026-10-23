<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-193: 30 back-to-back RenameTable all 204 (121-182 ms) and versionToken NEVER changes; 10 concurrent renames -> 1 x 204, 4 x 409, 1 x 500, 4 x 404
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-193"></a>**S3T-TABLE-193** `quota-limit` · impact high · handled · verified 2026-10-13
  **30 back-to-back RenameTable all 204 (121-182 ms) and versionToken NEVER changes; 10 concurrent renames -> 1 x 204, 4 x 409, 1 x 500, 4 x 404**
  30 sequential RenameTable calls alternating between two names on one table with no versionToken and no pause
  (each followed by GetTable(tableArn)): {"HTTP 204": 30} in 4.50s, first non-2xx at index None, RenameTable
  latency [121, 182] ms; 1 distinct versionTokens seen across the 30 GetTable reads (0 consecutive changes).
  Messages: none. Then 10 concurrent RenameTable to 10 distinct new names (no versionToken, wall 3.13s) ->
  {"ConflictException/409": 4, "InternalServerErrorException/500": 1, "204": 1, "NotFoundException/404": 4}
  messages {"ConflictException": "The table was modified concurrently. Try again.",
  "InternalServerErrorException": "An internal error occurred. Try again.", "NotFoundException": "The
  specified table does not exist."}; GetTable(tableArn) afterwards -> name ackq_62ee3a_n02, which is one of
  the 1 names whose rename returned 2xx; ListTables shows 1 table(s) in the namespace.
  - ACK: requeue, custom_update · ops: RenameTable, GetTable, ListTables · fields: newName, versionToken
  - repro: 30 x RenameTable(t<->t_alt) sequential with GetTable after each; then 10 threads x RenameTable(cur
    -> n_i); GetTable(tableArn); ListTables
  - measurements: rename_30_wall_s=4.5, rename_latency_ms_min_max=[121, 182], rename_race_wall_s=3.13, n=40
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-064](../table-metadata-location.md#s3t-table-064), [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065), [S3T-TABLE-192](../table-metadata-location.md#s3t-table-192), [S3T-TABLE-003](../table-metadata-location.md#s3t-table-003), [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009), [S3T-TABLE-111](../table-metadata-location.md#s3t-table-111),
    [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007), [S3T-TABLE-005](../table-metadata-location.md#s3t-table-005), [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112), [S3T-TABLE-316](../table-metadata-location.md#s3t-table-316), [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104),
    [S3T-TABLE-124](../table-metadata-location.md#s3t-table-124), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019), [S3T-TABLE-121](../table-metadata-location.md#s3t-table-121), [S3T-TABLE-245](../table-metadata-location.md#s3t-table-245), [S3T-TABLE-133](../table-schema.md#s3t-table-133), [S3T-TABLE-036](../table-schema.md#s3t-table-036), [S3T-TABLE-129](../table-metadata-location.md#s3t-table-129) ·
    hypotheses: H-T-200, H-T-201 · evidence: table/limits/policy-size-write-races

## Notes

H-T-200 qualified: 30 sequential RenameTable alternating two names with no versionToken and no pause -> 30 x
204 in 4.5 s total (121-182 ms each), no 409/429 - but the GetTable versionToken after every rename was
IDENTICAL (1 distinct token across 30 reads), refuting the 'versionToken differs after each rename' clause:
renames do not bump the token (UpdateTableMetadataLocation does). H-T-201 qualified: 10 concurrent RenameTable
to distinct new names gave exactly one 204 (GetTable(tableArn) shows that name, ListTables shows 1 table), but
the nine losers were a MIX: 4 x ConflictException 409 'The table was modified concurrently. Try again.', 1 x
InternalServerErrorException 500 'An internal error occurred. Try again.', 4 x NotFoundException 404 'The
specified table does not exist.' (the old name no longer existed once the winner landed) - so a 404 from
RenameTable can mean 'already renamed by someone else', not 'table gone'. Burst wall 3.13 s (max latency 3127
ms). 3-4 other probes running.
