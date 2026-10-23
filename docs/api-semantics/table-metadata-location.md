<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table metadata location and versionToken (UpdateTableMetadataLocation, GetTableMetadataLocation)
_metadataLocation as engine-managed state (set synchronously by CreateTable-with-schema, moved by every commit), GetTableMetadataLocation, UpdateTableMetadataLocation validation, versionToken as a function of metadataLocation, ConflictException CAS semantics on Delete/Rename/Update._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-metadata-location.md`; every claim cites the finding ids that carry the full
evidence below.

### Orientation

This document covers `metadataLocation` (exposed by GetTable and GetTableMetadataLocation, moved by
UpdateTableMetadataLocation), the `versionToken` that gates DeleteTable, RenameTable and UpdateTableMetadataLocation with
ConflictException 409, and the identity, 404 and post-delete behaviors those operations share. Shared model with
`table-schema.md`: `metadata.iceberg` is write-only (the schema lives only in the metadata.json at `metadataLocation`),
`metadataLocation` is engine-managed state, and `versionToken` is a deterministic function of `metadataLocation`. The key
is absent on a table created without a schema, the first UpdateTableMetadataLocation with the CreateTable token adds it,
every later update moves it with reads consistent at 0 s, and CreateTable-with-schema sets it synchronously
([S3T-TABLE-014](#s3t-table-014), -031, -315, -123, -101, -102; [S3T-TABLE-015](table-schema.md#s3t-table-015), -033 in `table-schema.md`). `versionToken` is a 20-char hex
string, not a per-commit nonce: unchanged by reads, renames and tag writes, back to the create-time value when the location
is restored, so a 409 'Provided version token does not match the table version token.' means only "the location differs
from the one your token encodes" ([S3T-TABLE-034](#s3t-table-034), -005, -124, -134, -108, -302); the object behind a location is never
verified ([S3T-TABLE-100](#s3t-table-100)).

### Rules

1. Do not reconcile `metadataLocation` as desired state: on an empty table the first UpdateTableMetadataLocation(CreateTable
   token, `<warehouseLocation>/metadata/<file>`) -> 200 adds the key and a new token (the create token -> 409 next), every
   later commit moves it, and UpdateTableMetadataLocation(current token, old location) is accepted as a rollback with no
   monotonicity check (20 moves away, then back to the 00000-* file -> 200, response token == the CreateTable token), so a
   pinned spec clobbers the engine's commit chain after every commit; on a schema-at-create table the key is non-nil from
   the first GetTable ([S3T-TABLE-015](table-schema.md#s3t-table-015), -033 in `table-schema.md`) ([S3T-TABLE-123](#s3t-table-123), -101, -277, -108, -134, -302).
2. A `versionToken` match is not proof the table is unchanged: A->B->A->B gives A0==A1==A2 and B1==B2 while another table
   at the same relative key gets a different token; a stale token revived by another writer's rollback lets
   DeleteTable(A0) -> 204 and UpdateTableMetadataLocation(C, A0) -> 200; a same-location re-send (x3) -> 200 with the
   token unchanged but modifiedAt advancing ([S3T-TABLE-134](#s3t-table-134), -302, -104).
3. Only a location change moves the token: 1 distinct token across 10 reads over 54 s with no writes, across 30
   back-to-back renames, across name, namespace and both-fields renames, and across TagResource/UntagResource (which
   advance modifiedAt only); modifiedBy stays absent after a commit until RenameTable populates it ([S3T-TABLE-034](#s3t-table-034), -193,
   -005, -321, -006, -007, -124, -316).
4. Tell the ConflictException 409s apart by message: duplicate CreateTable -> 'A table with an identical name already
   exists in the namespace.' at t+0/5/10 s and with different metadata or tags (permanent; existing table untouched);
   'Provided version token does not match the table version token.' (stale or garbage token on DeleteTable, RenameTable,
   UpdateTableMetadataLocation) clears after a re-read; 'The table was modified concurrently. Try again.' is the
   concurrent-rename 409 ([S3T-TABLE-064](#s3t-table-064), -051, -103, -065, -193).
5. Re-read or omit the token before DeleteTable/RenameTable: the create-time token after one commit that moved the
   location -> 409 while the current token -> 204; a tokenless DeleteTable -> 204 in 20/20 and a tokenless RenameTable
   -> 204; garbage, another table's, upper-cased or 2048-char tokens -> 409 (no 400 seen); an empty token on
   UpdateTableMetadataLocation -> ValidationException 400 ([S3T-TABLE-098](#s3t-table-098), -089, -275, -065, -103, -321, -282).
6. Do not infer anything from a location being accepted: the check is a bare string prefix of `warehouseLocation`
   (sibling table, other bucket, foreign bucket, 'S3://', 's3a://', https -> 400 'The specified metadata location is not
   valid.' but `s3://<warehouse-host>x/..` -> 200) plus the case-sensitive substring '.metadata.json' (x.json,
   x.metadata.JSON, x.avro -> 400; x.metadata.json.gz, .json.gzip, .json.tar, a trailing '/' -> 200); object existence
   is never verified (a dangling path -> 200); '//', '%20' and a literal space are stored byte-exact; the warehouse host
   (3 uuid groups + opaque suffix + '--table-s3') is unknowable before CreateTable ([S3T-TABLE-105](#s3t-table-105), -278, -106, -107, -279,
   -135, -280, -100, -281, -192, -028).
7. One writer wins per token: two concurrent UpdateTableMetadataLocation calls with the same fresh token -> exactly one
   200 and one 409 in 10/10 rounds, reads then show the winner's location and a changed token; 10 concurrent tokenless
   RenameTable -> 1x204, 4x409, 1x500, 4x404 and the table ends under one of the 2xx names ([S3T-TABLE-192](#s3t-table-192), -193).
8. Resolve 404s and region errors before deciding "gone": GetTable, GetTableMetadataLocation and
   UpdateTableMetadataLocation on a missing table, namespace or table bucket all -> NotFoundException 404 'The specified
   table does not exist.'; after a rename the old triple -> 404 within 0.2 s (GetTable, a replayed RenameTable,
   DeleteTable, UpdateTableMetadataLocation) while GetTable(tableArn) returns the current triple, which deletes with the
   pre-rename token and accepts UpdateTableMetadataLocation; an other-region ARN -> 400 'The provided Region does not match
   the Region of the resource.' before any lookup ([S3T-TABLE-121](#s3t-table-121), -111, -002, -009, -003, -320, -245).
9. After DeleteTable (204, empty body) the table is gone on the next call: GetTable, GetTableMetadataLocation and a second
   DeleteTable -> 404 from the first sample in 20/20, 3/3 and 2/2 cycles; TagResource/ListTagsForResource on the old ARN
   -> 404 but UntagResource -> 500; a same-name re-create -> 200 at once with a new tableARN, warehouseLocation and token
   and no `metadataLocation` key; old metadata.json -> HeadObject 403 ([S3T-TABLE-089](#s3t-table-089), -099, -122, -323, -066, -129, -067).
10. Read `versionToken`, `metadataLocation` and `warehouseLocation` only from GetTable (triple or tableArn - identical
    payload; 13 keys without metadata, 14 with) or GetTableMetadataLocation: CreateTable returns only {tableARN,
    versionToken}; ListTables carries only createdAt, modifiedAt, name, namespace, tableARN, type ([S3T-TABLE-090](#s3t-table-090), -023,
    -001, -014, -031, -315, -269, -292).

### Timing ranges

- UpdateTableMetadataLocation: 11-17 ms over 21 accepted updates (n=21, [S3T-TABLE-108](#s3t-table-108)); 13-17 ms for three same-location
  re-sends (n=3, -104); 13-16 ms for the two updates of the consistency series (n=2, -102); the single calls of -101 and
  -123 (n=1 each) fall inside 12-13 ms.
- Read-after-write on token/location: 0 stale among 38 alternating GetTable/GetTableMetadataLocation samples over ~8 s
  after two updates (n=2 updates, -102); response == reads at 0/1/5 s (n=1, -101); a same-token CAS round settles in
  0.01-0.03 s wall (n=10, -192) - no poll loop is needed, only a re-read before token-guarded writes.
- No writes: the `metadataLocation` key stays absent at 0 s and 5 s on a no-schema table (n=1, -014) and the GetTable
  key set is unchanged at +3 s (n=2 tables, -315); 1 distinct token across 10 GetTable reads at 6 s intervals over 54 s
  (n=10, -034) and 1 distinct body across 20 back-to-back GetTableMetadataLocation reads (-019).
- DeleteTable: 105-168 ms on empty tables (n=20, -089); 118-192 ms with committed metadata and a data object (n=3, -099);
  146-154 ms (n=2, -323). First read after the 204 was 404 at 7-46 ms by triple and 10-58 ms by ARN in 20/20 (-089), at
  9-11 ms in 3/3 (-099) and in 6/6 samples over 3 s in 2/2 cycles (-323); no 200 window in any series
  (`last_200_after_delete_s` null; -122, -323).
- Same-name re-create: 200 in 10/10 when issued 308-403 ms after the DeleteTable 204, 409 in 0 (n=10, -067).
- RenameTable: 121-182 ms over 30 back-to-back renames (n=30, -193); 13-152 ms across the single-shot rename, move and
  both-fields calls (n=5 across -002, -003, -006, -007); the old name is 404 on GetTable and GetTableMetadataLocation
  0.17 s after the 204 (n=1, -111); 10 concurrent renames completed in 3.13 s wall (single run, -193).
- Duplicate CreateTable: ConflictException at t+0, +5 and +10 s with no clearing window observed (n=3, -064).

### Handling gaps

Today `MetadataLocation` is a spec field from GetTable (generator.yaml:76-83) with a post-create requeue when desired
(generator.yaml:106-110); `customUpdateTable` sends UpdateTableMetadataLocation(latest `Status.VersionToken`, desired
location) on any delta (hooks.go:96-100, 145-171); the read-only `VersionToken` (generator.yaml:51-56) goes to RenameTable
(hooks.go:132-134) and DeleteTable (sdk.go:489-491); `renameTable` never sets `NewNamespaceName` (hooks.go:126-131);
BadRequestException is the only terminal code (generator.yaml:99-104). Three findings are suspect-bug; none is handled yet.

- [S3T-TABLE-108](#s3t-table-108) (suspect-bug, not yet handled): `metadataLocation` is reconciled as desired state. Evidence: after 20
  moves, UpdateTableMetadataLocation(current token, 00000-*) -> 200 with the CreateTable token back (n=21); with -134/-302
  a pinned spec re-issues that rollback after each engine commit and the next engine commit fails 409. Fix per the finding:
  read-only status (or late-init + compare-ignore), no create-time requeue (a desired location on a fresh table can only
  fail 400 -> Terminal: server-assigned warehouse host, -105, -028).
- [S3T-TABLE-098](#s3t-table-098) (suspect-bug, not yet handled): the cached `Status.VersionToken` is passed to DeleteTable/RenameTable
  although optional. Evidence: DeleteTable(T0) after one commit that moved the location -> 409, DeleteTable(T1) -> 204;
  a tokenless delete -> 204 (-089, -067); a token match proves only the location (-134, -302). Omit the token, or map
  this 409 to requeue-after-re-read and never to Terminal (the same code is permanent for duplicate CreateTable, -064).
- [S3T-TABLE-006](#s3t-table-006) (suspect-bug, not yet handled): after an out-of-band RenameTable(newNamespaceName) the ARN-keyed read
  puts the new namespace into latest `Spec.Namespace`, delta.go:146-152 flags it every reconcile and `customUpdateTable`
  ignores it (hooks.go:90-109) - no RenameTable back, no condition, Synced=True; ConflictException is not terminal, so a
  permanent 409 name collision on RenameTable retries forever. Evidence: the move is synchronous (204, n=1), namespaceId
  follows the destination, tableARN and versionToken unchanged.

### Where next

- `table-schema.md` is the other half of the shared model: CreateTable-with-schema sets `metadataLocation` on the first
  GetTable with the create token (-015, -033, -283), what the committed metadata.json carries and how it differs from the
  request (-016, -043, -312, -322), the `schema={}` -> 500 loop (-133, -036) and the write-only read gap (-018, -313).
- `table.md` (catch-all) is behind rules 8-9: identity by ARN vs triple, 404 message shapes (FU2 "GetTable 404 message
  depends on name shape" qualifies -121 for well-formed names), RenameTable under collision (the -006 note points at
  [S3T-TABLE-048](table.md#s3t-table-048), -050, -070 there; -112 is a duplicate of -051 in this document's appendix), tags. `table-encryption-storage.md`: GetTable key set (-308), sub-reads after
  rename (-319).
- `table-maintenance.md`: token unchanged by maintenance Puts, a metadata-only commit as the job trigger, the -233
  correction. `table-policy.md`: AccessDeniedException guarding UpdateTableMetadataLocation (-132). `table-replication.md`:
  configuration versionToken vs the table versionToken of rule 2; `tablebucket-replication.md`: UpdateTableMetadataLocation
  as the source commit, 400 on a replica ([S3T-TABLEBUCKETREPLICATION-029](tablebucket-replication.md#s3t-tablebucketreplication-029), -015); `table-record-expiration.md`: -148, -222.
- Long catalogues render in `details/<id>.md`: the metadataLocation prefix/scheme/suffix matrices ([S3T-TABLE-105](#s3t-table-105), -106,
  -107, -279, -135), the 404/region matrix (-121, -245), the post-delete series (-089, -122, -323, -129), the races (-192,
  -193). The doc-claim confirmations (-269, -275, -277, -278, -280, -281, -282, -292) restate rules 1, 5, 6, 10 (appendix).
- Open experiments (STATUS.md, Phase 8): the re-verification corrected [S3T-TABLE-122](#s3t-table-122) via -323 (the deleted metadata.json
  answers HeadObject 403 from 0.03 s, or 200 up to 10 s and 403 from 20 s if read before the delete; the control-plane
  404s hold). Rules 1-2 and the -108/-098/-006 fixes are code-reading conclusions, not an in-cluster reproduction; n=1 and
  single-run items are in `table-metadata-location.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 57 (high 29 / medium 18 / low 10); duplicates folded into the appendix: 1
- handling: handled 29 · partial 0 · tracked 0 · unhandled 25 · suspect-bug 3 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 4 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: idempotency 9, identity 9, other 8, delete-semantics 6, error-code 5, request-validation 5,
  response-fidelity 5, read-gap 2, eventual-consistency 1, normalization 1, prerequisite 1, quota-limit 1,
  server-default 1, sub-resource-api 1, tag-semantics 1, update-granularity 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateNamespace | create | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetNamespace | read | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-table-001"></a>**S3T-TABLE-001** `identity` · impact high · handled · verified 2026-10-13
  **Table ARN is <bucketARN>/table/<uuid> (namespace/name not in it); GetTable accepts tableArn alone or bucketARN+namespace+name**
  CreateTable (format=ICEBERG, no metadata) returned tableARN
  'arn:aws:s3tables:<region>:<account>:bucket/<bucket>/table/<uuid>'; neither the namespace nor the table name
  appears in it, so the ARN cannot be derived from the spec. GetTable(tableBucketARN+namespace+name) -> HTTP
  200; GetTable(tableArn=<tableARN>) -> HTTP 200, identical payload. The fresh table's GetTable payload has
  ['type', 'format', 'versionToken', 'warehouseLocation', 'namespaceId', 'tableBucketId'] and omits
  ['metadataLocation', 'managedByService'] (type=customer; warehouseLocation form
  's3://<table-uuid-prefix><opaque>--table-s3').
  - ACK: is_arn_primary_key · ops: CreateTable, GetTable · fields: tableARN, tableArn, metadataLocation,
    warehouseLocation, versionToken
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(format=ICEBERG) -> GetTable by names and by
    tableArn
  - measurements: get_table_visible_after_create_s=0.03, n=1
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-269](#s3t-table-269), [S3T-TABLE-270](table.md#s3t-table-270), [S3T-TABLE-023](#s3t-table-023), [S3T-TABLE-024](table.md#s3t-table-024), [S3T-TABLE-025](table.md#s3t-table-025), [S3T-TABLE-030](table-encryption-storage.md#s3t-table-030),
    [S3T-TABLE-173](table-replication.md#s3t-table-173), [S3T-TABLE-175](table-replication.md#s3t-table-175), [S3T-TABLE-315](#s3t-table-315), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-031](#s3t-table-031) · evidence:
    tablebucket/round-trip/smoke
  - notes: Smoke probe, single run: ARN shape and lookup forms are deterministic; the latency is one sample.
    versionToken is identical across CreateTable, GetTable and GetTableMetadataLocation.

- <a id="s3t-table-002"></a>**S3T-TABLE-002** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable(newName) keeps tableARN/ids/locations; GetTable by old ARN -> 200 new name...**
  On an ICEBERG table with schema, tags and a policy: RenameTable(newName) -> HTTP 204 (body keys: []). Then
  GetTable(tableArn=<pre-rename ARN>) -> HTTP 200 with name=<newName>; GetTable(bucket, ns, newName) -> HTTP
  200 and its payload is identical to the ARN read; GetTable(old triple) -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'. Fields unchanged: tableARN, tableBucketId, namespaceId, warehouseLocation,
  metadataLocation, createdAt, createdBy, format, type, ownerAccountId. Fields changed: modifiedAt,
  modifiedBy, name.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTable · fields: tableARN, tableBucketId,
    namespaceId, warehouseLocation, versionToken, name
  - repro: CreateTable(ns1,t1,schema,tags) -> PutTablePolicy -> RenameTable(newName=t2) -> GetTable(tableArn),
    GetTable(ns1,t2), GetTable(ns1,t1)
  - measurements: rename_latency_ms=[149], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-004](table-policy.md#s3t-table-004), [S3T-TABLE-032](table.md#s3t-table-032) · hypotheses: H-T-050, H-T-055 · evidence:
    table/identity/rename-identity
  - notes: H-T-050 confirmed (ARN byte-identical: True; tableBucketId/namespaceId unchanged: True). H-T-055:
    rename response is HTTP 204 with empty body; modifiedAt changed: True; modifiedBy changed: True; createdAt
    changed: False; createdBy changed: False.
  - full notes: [details/S3T-TABLE-002.md](details/S3T-TABLE-002.md)

- <a id="s3t-table-006"></a>**S3T-TABLE-006** `identity` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId follows the destination, tableARN unchanged**
  RenameTable(ns1,t2, newNamespaceName=ns2) with no newName -> HTTP 204. GetTable(ns2,t2) -> HTTP 200;
  GetTable(tableArn) -> HTTP 200 with namespace=['<ns2>']. Changed fields vs before the move: modifiedAt,
  namespace, namespaceId; namespaceId equals GetNamespace(ns2).namespaceId: True. ListTables(namespace=ns1) ->
  HTTP 200 with 0 tables; ListTables(namespace=ns2) -> 1 tables.
  - ACK: is_immutable, custom_update, is_arn_primary_key · ops: RenameTable, GetTable, ListTables,
    GetNamespace · fields: newNamespaceName, namespace, namespaceId, tableARN, versionToken, warehouseLocation
  - repro: CreateNamespace ns1, ns2; CreateTable ns1/t; RenameTable(ns1,t,newNamespaceName=ns2);
    GetTable(ns2,t), GetTable(tableArn); ListTables(namespace=ns1)
  - measurements: move_latency_ms=[143], n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-022](table.md#s3t-table-022), [S3T-TABLE-029](table.md#s3t-table-029), [S3T-TABLE-008](table.md#s3t-table-008), [S3T-TABLE-068](table.md#s3t-table-068), [S3T-TABLE-069](table.md#s3t-table-069), [S3T-TABLE-070](table.md#s3t-table-070),
    [S3T-TABLE-071](table.md#s3t-table-071) · hypotheses: H-T-057 · evidence: table/identity/rename-identity
  - notes: H-T-057 confirmed: versionToken changed on move: False; warehouseLocation changed: False;
    tableBucketId changed: False. DeleteNamespace of the emptied source ns is exercised later in this probe.
  - full notes: [details/S3T-TABLE-006.md](details/S3T-TABLE-006.md)

- <a id="s3t-table-009"></a>**S3T-TABLE-009** `identity` · impact high · handled · verified 2026-10-15, re-verified
  **After an out-of-band rename DeleteTable(old triple) -> 404; GetTable(tableArn) recovers the current triple; pre-rename token still deletes**
  Table created as ns1/t1, renamed 3 times (now ns1/t3). DeleteTable(ns1,t1) -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. GetTable(tableArn=<stored ARN>) -> HTTP 200 returning
  namespace=[<current ns>], name=<current name>. DeleteTable(recovered triple with versionToken=<CreateTable
  token>) -> HTTP 204. GetTable(tableArn) afterwards -> NotFoundException (HTTP 404) 'The specified table does
  not exist.'.
  - ACK: custom_delete, is_arn_primary_key, exceptions.404 · ops: DeleteTable, GetTable, RenameTable · fields:
    tableArn, namespace, name, versionToken
  - repro: CreateTable ns1/t1 -> RenameTable (x3) -> DeleteTable(ns1,t1) -> GetTable(tableArn) ->
    DeleteTable(current namespace/name) -> GetTable(tableArn)
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:10-13; generator.yaml:31-39`
  - related: [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-111](#s3t-table-111), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-005](#s3t-table-005),
    [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-316](#s3t-table-316) · hypotheses: H-T-099, H-T-054 · evidence: table/identity/rename-identity
  - notes: H-T-099 confirmed. H-T-054: DeleteTable guarded with the CreateTable-time versionToken after three
    renames -> HTTP 204 (token still valid: True).
  - full notes: [details/S3T-TABLE-009.md](details/S3T-TABLE-009.md)

- <a id="s3t-table-023"></a>**S3T-TABLE-023** `identity` · impact high · handled · verified 2026-10-13
  **GetTable(tableArn) payload identical to GetTable(tableBucketARN+namespace+name); 13 keys on a no-metadata table (no metadataLocation key)**
  GetTable(tableArn=<tableARN>) -> HTTP 200; GetTable(composite) -> HTTP 200; payloads (ResponseMetadata
  excluded) are identical. Keys returned: ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name',
  'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation']. namespace is a list of length 1 (element 0 == created namespace: True); the ARN output
  key is spelled ['tableARN'] while the input member is 'tableArn'.
  - ACK: is_arn_primary_key, custom_find · ops: GetTable · fields: tableArn, tableARN, namespace, name,
    namespaceId, tableBucketId, ownerAccountId
  - repro: Create t1; GetTable(tableArn) and GetTable(tableBucketARN, namespace, name) within 1 s; diff bodies
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:10-13; generator.yaml:31-39; generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-269](#s3t-table-269), [S3T-TABLE-270](table.md#s3t-table-270), [S3T-TABLE-001](#s3t-table-001), [S3T-TABLE-024](table.md#s3t-table-024), [S3T-TABLE-025](table.md#s3t-table-025), [S3T-TABLE-018](table-schema.md#s3t-table-018),
    [S3T-TABLE-310](table-schema.md#s3t-table-310), [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-090](#s3t-table-090), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-028](#s3t-table-028) · hypotheses: H-T-002, H-U-195,
    H-T-012 · evidence: table/identity/get-table-identifier-forms
  - notes: H-T-002 confirmed; H-U-195 confirmed (ARN-only read returns name, namespace, tableBucketId,
    namespaceId, ownerAccountId so the triple is recoverable from the ARN); H-T-012 confirmed (namespace sent
    as scalar string on CreateTable, returned as list).

- <a id="s3t-table-028"></a>**S3T-TABLE-028** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **warehouseLocation host = first 3 uuid groups (18-19 chars) of the table uuid + opaque suffix + '--table-s3'; no names in it; stable, unique**
  GetTable warehouseLocation for t1 is 's3://ef44dd0a-f123-4ce4-nuqppejam7deeueahexkquthetdnqusw2b--table-s3'
  (host length 63): it shares only its first 19 characters with the table uuid from the ARN (full uuid prefix:
  False), ends with '--table-s3': True, contains bucket name: False, namespace: False, table name: False; t2
  likewise shares 19 chars with its uuid. Stable across two reads: t1=True t2=True; different between t1 and
  t2: True.
  - ACK: is_read_only · ops: GetTable · fields: warehouseLocation
  - repro: Create t1,t2; GetTable twice each; inspect warehouseLocation host
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-018](table-schema.md#s3t-table-018), [S3T-TABLE-310](table-schema.md#s3t-table-310), [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-090](#s3t-table-090), [S3T-TABLE-023](#s3t-table-023), [S3T-TABLE-113](table.md#s3t-table-113),
    [S3T-TABLE-067](#s3t-table-067), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-089](#s3t-table-089) · hypotheses: H-T-031 ·
    evidence: table/identity/get-table-identifier-forms
  - notes: H-T-031 qualified: the host is NOT the full table uuid (only the first three uuid groups, an 18-19
    char shared prefix across two runs), the rest is opaque; the other claims (no names in it, --table-s3
    suffix, unique, stable) hold. The controller must copy warehouseLocation from GetTable; it cannot...
  - full notes: [details/S3T-TABLE-028.md](details/S3T-TABLE-028.md)

- <a id="s3t-table-067"></a>**S3T-TABLE-067** `identity` · impact high · handled · verified 2026-10-13
  **Immediate CreateTable after DeleteTable (same name) -> 200 in 10/10 (409 in 0); new tableARN every time, old ARN -> 404 in 10/10**
  10 trials of CreateTable -> DeleteTable(no token) -> CreateTable(same name) issued 308-403 ms after the
  delete response: immediate HTTP 200 in 10, ConflictException in 0 (retries to 200: []). The re-created table
  has a different tableARN (uuid) in all trials: True, a different versionToken: True, a different
  warehouseLocation (trial 0): True. GetTable(old tableArn) afterwards -> NotFoundException (HTTP 404) 'The
  specified table does not exist.' in 10/10 trials; GetTable(new tableArn) -> HTTP 200.
  - ACK: is_arn_primary_key, custom_find, requeue · ops: CreateTable, DeleteTable, GetTable · fields:
    tableARN, versionToken, warehouseLocation
  - repro: Loop 10x: CreateTable t; DeleteTable t; CreateTable t at once (retry each 1 s on 409); GetTable by
    old and new ARN; DeleteTable t
  - measurements: recreate_gap_ms=[361, 326, 308, 338, 336, 403, 316, 326, 329, 368], immediate_200=10,
    conflict_409=0, n=10
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-028](#s3t-table-028) · hypotheses: H-T-041,
    H-T-042 · evidence: table/idempotency/duplicate-create-double-delete
  - notes: H-T-041 confirmed. H-T-042 (409 name reservation for 1-2 s) refuted. A controller holding the old
    ARN sees 404 by ARN while the composite read returns the NEW table: it must compare ARNs and re-adopt.
  - full notes: [details/S3T-TABLE-067.md](details/S3T-TABLE-067.md)

- <a id="s3t-table-111"></a>**S3T-TABLE-111** `identity` · impact high · handled · verified 2026-10-13
  **RenameTable (newName, then newNamespaceName): tableARN, warehouseLocation, metadataLocation stable; old identifiers -> 404 within 0.2 s**
  RenameTable(newName) -> HTTP 204; under the new name GetTable shows the same tableARN True,
  warehouseLocation True, metadataLocation True, createdAt True; GetTable/GetTableMetadataLocation under the
  OLD name (0.17 s after the rename response) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.'. RenameTable(newNamespaceName)
  -> HTTP 204; tableARN same True, warehouseLocation same True, metadataLocation same True, namespaceId
  changed True; old namespace+name reads -> NotFoundException (HTTP 404) 'The specified table does not exist.'
  / NotFoundException (HTTP 404) 'The specified table does not exist.'. GetTable(tableArn) after both renames
  -> HTTP 200 and reports the new namespace/name.
  - ACK: is_arn_primary_key, custom_update · ops: RenameTable, GetTable, GetTableMetadataLocation · fields:
    tableARN, warehouseLocation, metadataLocation, namespaceId, name, namespace
  - repro: CreateTable with schema -> RenameTable(newName) -> GetTable/GetTableMetadataLocation under old and
    new names -> RenameTable(newNamespaceName) -> same reads -> GetTable(tableArn)
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-007](#s3t-table-007) · hypotheses: H-T-144 ·
    evidence: table/identity/version-token-rename
  - notes: H-T-144 confirmed: tableARN, warehouseLocation, metadataLocation and createdAt identical under the
    new name and the new namespace (namespaceId changes with the namespace move); old name and old
    namespace+name -> NotFoundException 404 'The specified table does not exist.' on both GetTable and...
  - full notes: [details/S3T-TABLE-111.md](details/S3T-TABLE-111.md)

- <a id="s3t-table-321"></a>**S3T-TABLE-321** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **RenameTable keeps the table versionToken; DeleteTable(current triple) with a bogus token -> ConflictException...**
  Fresh ICEBERG table (no metadata): RenameTable(newName, no versionToken) -> HTTP 204 (4 renames); afterwards
  GetTable(tableArn).versionToken does NOT change (create token 20 chars, post-rename token 20 chars;
  GetTable.modifiedAt strictly increased across the 3 renames: True). DeleteTable(renamed triple,
  versionToken='ackqbogus0000000000000') -> ConflictException (HTTP 409) 'Provided version token does not
  match the table version token.'. DeleteTable(renamed triple, versionToken=<CreateTable token>) -> HTTP 204
  (second table) and -> HTTP 204 (first table, after 3 renames). DeleteTable(old pre-rename triple) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: is_arn_primary_key, custom_delete · ops: RenameTable, DeleteTable, GetTable · fields: versionToken,
    name, namespace
  - repro: CreateTable -> RenameTable(newName) -> GetTable(tableArn) compare versionToken ->
    DeleteTable(versionToken=bogus) -> DeleteTable(versionToken=create token)
  - measurements: tables=2, renames=4, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-009](#s3t-table-009) · evidence: table/creative/reverify-set-a
  - notes: Re-verification precision gain for [S3T-TABLE-009](#s3t-table-009): states whether the 'pre-rename token still
    deletes' clause is because rename does not rotate the token or because DeleteTable ignores/loosely checks
    versionToken. n=2 tables.
  - full notes: [details/S3T-TABLE-321.md](details/S3T-TABLE-321.md)

## Idempotency

- <a id="s3t-table-003"></a>**S3T-TABLE-003** `idempotency` · impact medium · handled · verified 2026-10-13
  **Replaying an identical RenameTable right after it succeeded -> NotFoundException (HTTP 404) 'The specified table does not exist.'**
  RenameTable(ns1,t1,newName=t2) without versionToken -> HTTP 204; the identical call re-sent immediately (old
  triple no longer exists) -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  GetTable(tableArn) afterwards shows name=newName.
  - ACK: custom_update, is_arn_primary_key · ops: RenameTable · fields: newName, versionToken
  - repro: RenameTable(ns1,t1,newName=t2) twice back to back without versionToken
  - measurements: rename_latency_ms=[149, 13], n=2
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122`
  - related: [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-111](#s3t-table-111), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-121](#s3t-table-121),
    [S3T-TABLE-245](#s3t-table-245) · hypotheses: H-T-062 · evidence: table/identity/rename-identity
  - notes: H-T-062 confirmed: first call HTTP 204, replay NotFoundException (HTTP 404) 'The specified table
    does not exist.' (hypothesis predicted 204 then NotFoundException 404).

- <a id="s3t-table-064"></a>**S3T-TABLE-064** `idempotency` · impact high · handled · verified 2026-10-13
  **Duplicate CreateTable (same bucket/namespace/name) -> ConflictException 409 at t+0/5/10 s and with different metadata/tags; table untouched**
  After CreateTable(t1, format=ICEBERG, tags) -> 200: identical CreateTable at t+0 -> ConflictException (HTTP 409)
  'A table with an identical name already exists in the namespace.'; t+5 s -> ConflictException (HTTP 409) 'A
  table with an identical name already exists in the namespace.'; t+10 s -> ConflictException (HTTP 409) 'A
  table with an identical name already exists in the namespace.'; with metadata.iceberg.schema ->
  ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'; without tags
  -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'; with an
  extra tag -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'.
  GetTable afterwards: tableARN/versionToken/createdAt/modifiedAt/warehouseLocation unchanged = {'tableARN':
  True, 'versionToken': True, 'createdAt': True, 'modifiedAt': True, 'warehouseLocation': True}; ListTables
  shows 1 entry named t1.
  - ACK: custom_create, terminal_codes · ops: CreateTable, GetTable · fields: name, namespace, versionToken
  - repro: CreateTable t1; CreateTable t1 again at t+0, t+5 s, t+10 s with identical params, then with
    metadata, without tags, with extra tags; GetTable; ListTables
  - measurements: retry_spacing_s=[0, 5, 10], n=3
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-NAMESPACE-036](namespace.md#s3t-namespace-036), [S3T-TABLE-274](table.md#s3t-table-274), [S3T-TABLE-276](table.md#s3t-table-276), [S3T-NAMESPACE-021](namespace.md#s3t-namespace-021), [S3T-TABLE-050](table.md#s3t-table-050), [S3T-TABLE-065](#s3t-table-065),
    [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-067](#s3t-table-067), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-028](#s3t-table-028) ·
    hypotheses: H-T-023, H-T-024, H-T-025 · evidence: table/idempotency/duplicate-create-double-delete
  - notes: H-T-023 confirmed. H-T-024 (PUT-like idempotent 200) refuted. H-T-025 (409 is permanent across 3
    spaced retries) confirmed. Message text: 'A table with an identical name already exists in the
    namespace.'. The duplicate-create ConflictException message differs from the versionToken-mismatch...
  - full notes: [details/S3T-TABLE-064.md](details/S3T-TABLE-064.md)

- <a id="s3t-table-066"></a>**S3T-TABLE-066** `idempotency` · impact high · handled · verified 2026-10-13
  **DeleteTable -> 204 empty body; GetTable (composite and ARN) -> 404 at once; 2nd DeleteTable -> 404, same message as a never-created name**
  DeleteTable(correct versionToken) -> HTTP 204 (parsed body empty: True). GetTable composite ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTable(tableArn) -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'. Second DeleteTable without token -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; with the old token -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; DeleteTable of a never-created name -> NotFoundException (HTTP 404) 'The
  specified table does not exist.' (identical message to the second delete: True). DeleteTable of t2 WITHOUT
  versionToken -> HTTP 204.
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable · fields: versionToken
  - repro: Create t1; DeleteTable(t1, token); GetTable both forms; DeleteTable(t1) twice more;
    DeleteTable(never_existed)
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-292](#s3t-table-292), [S3T-TABLE-283](table-schema.md#s3t-table-283), [S3T-TABLE-291](table.md#s3t-table-291), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-088](table.md#s3t-table-088), [S3T-TABLE-089](#s3t-table-089),
    [S3T-TABLE-099](#s3t-table-099), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-121](#s3t-table-121) · hypotheses: H-T-039, H-T-036 · evidence:
    table/idempotency/duplicate-create-double-delete
  - notes: H-T-039 confirmed: no distinct already-deleted signal. H-T-036 partially covered here (single
    trial; the 20-trial window measurement is in table/consistency-windows/create-delete-visibility).
  - full notes: [details/S3T-TABLE-066.md](details/S3T-TABLE-066.md)

- <a id="s3t-table-104"></a>**S3T-TABLE-104** `idempotency` · impact high · handled · verified 2026-10-13
  **Re-sending UpdateTableMetadataLocation with the current token and the SAME location (x3) -> 200, token unchanged, but modifiedAt advances**
  Three consecutive UpdateTableMetadataLocation calls, each with the then-current token and the location
  already in effect: outcomes ['HTTP 200', 'HTTP 200', 'HTTP 200']; token changed per call: [False, False,
  False] (0 distinct tokens over 3 calls); GetTable.modifiedAt 2026-10-13 22:33:19.776587+00:00 -> 2026-10-13
  22:33:24.985951+00:00.
  - ACK: compare.is_ignored+delta_pre_compare, hook:sdk_update_pre_build_request · ops:
    UpdateTableMetadataLocation, GetTable · fields: metadataLocation, versionToken, modifiedAt
  - repro: UpdateTableMetadataLocation(T, loc) -> UpdateTableMetadataLocation(T', loc) x3 at 1 s intervals ->
    GetTable
  - measurements: update_latency_ms=[13, 14, 17], n=3
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-102](#s3t-table-102), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-108](#s3t-table-108),
    [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-124](#s3t-table-124), [S3T-TABLE-316](#s3t-table-316),
    [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-135, H-T-136 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-135 QUALIFIED: the no-op re-send is accepted (200) but does NOT mint a new token - the response
    echoes the same versionToken (1 distinct token over 3 calls); GetTable.modifiedAt nevertheless advanced
    from the pre-resend value to the time of the last re-send. H-T-136 REFUTED: no...
  - full notes: [details/S3T-TABLE-104.md](details/S3T-TABLE-104.md)

- <a id="s3t-table-108"></a>**S3T-TABLE-108** `idempotency` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200; response versionToken == the CreateTable token**
  After 20 accepted updates away from the create-time file (last location '<WH>/a/b/c/d.metadata.json'),
  UpdateTableMetadataLocation(current token, <loc0 = 00000-<uuid>.metadata.json>) -> HTTP 200;
  GetTableMetadataLocation shows loc0 again: True. HeadObject: loc0 still exists=True; the dangling 00001 path
  exists=False (never created). No earlier metadata file was deleted by any update.
  - ACK: docs-only, compare.is_ignored+delta_pre_compare · ops: UpdateTableMetadataLocation,
    GetTableMetadataLocation · fields: metadataLocation
  - repro: CreateTable with schema (loc0) -> UpdateTableMetadataLocation to
    <wh>/metadata/00001-<uuid>.metadata.json -> UpdateTableMetadataLocation back to loc0 ->
    GetTableMetadataLocation
  - measurements: update_latency_ms_min=11, update_latency_ms_max=17, n=21
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-100](#s3t-table-100), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-104](#s3t-table-104),
    [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-103](#s3t-table-103) · hypotheses: H-T-147 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-147 confirmed (no monotonicity check; both files untouched; the probe wrote no objects).
    SURPRISE: the token returned for the rollback equals the create-time token exactly, and the same-location
    re-sends also returned the same token - versionToken appears to be a deterministic function of the...
  - full notes: [details/S3T-TABLE-108.md](details/S3T-TABLE-108.md)

- <a id="s3t-table-134"></a>**S3T-TABLE-134** `idempotency` · impact medium · handled · verified 2026-10-13
  **versionToken is a deterministic function of the full metadataLocation: A->B->A->B gives A0==A1==A2, B1==B2; another table's same key differs**
  Table with location A (create token A0). UpdateTableMetadataLocation A->B, B->A, A->B, B->A (all HTTP 200):
  token after each return to A equals A0: True/True; the two B tokens are equal: True; A and B tokens differ:
  True. A second table pointing at the same relative key under ITS warehouse gets the same token as B: False
  (so the token depends on the full location, including the warehouse). Token is 20 chars, hex: True.
  - ACK: hook:sdk_update_pre_build_request, docs-only · ops: UpdateTableMetadataLocation, GetTable · fields:
    versionToken, metadataLocation
  - repro: CreateTable with schema -> UpdateTableMetadataLocation to B, back to A, to B, back to A -> compare
    the five tokens
  - measurements: updates=4, n=4
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-192](#s3t-table-192) ·
    hypotheses: H-T-135, H-T-139 · evidence: table/creative/token-and-suffix-followup
  - notes: Creative follow-up: explains why same-location re-sends (H-T-135), renames/tags (H-T-139) and the
    rollback (H-T-147) did not change the token - it is derived from the location, not a write counter.
    Consequence: a token cached before an external commit becomes valid again if the location is restored;...
  - full notes: [details/S3T-TABLE-134.md](details/S3T-TABLE-134.md)

- <a id="s3t-table-192"></a>**S3T-TABLE-192** `idempotency` · impact high · handled · verified 2026-10-13
  **UpdateTableMetadataLocation same-token race: exactly one 200 + one 409 in 10/10 rounds; also 200 for a missing .metadata.json object**
  UpdateTableMetadataLocation(versionToken=current,
  metadataLocation=<warehouseLocation>/metadata/00000-ackq-probe.metadata.json) on a fresh table where that
  object does NOT exist -> HTTP 200, versionToken changed in the response. Then 10 rounds of two concurrent
  updates carrying the same (fresh) versionToken and different locations: code pairs {'ConflictException (HTTP
  409)+HTTP 200': 10}; messages {"B": "Provided version token does not match the table version token.", "A":
  "Provided version token does not match the table version token."}; in 10/10 rounds GetTableMetadataLocation
  afterwards showed the winner's location and in 10/10 rounds the versionToken had changed.
  - ACK: requeue, custom_update · ops: UpdateTableMetadataLocation, GetTableMetadataLocation · fields:
    versionToken, metadataLocation
  - repro: GetTableMetadataLocation -> token; 2 threads x UpdateTableMetadataLocation(token, locA|locB);
    GetTableMetadataLocation; repeat 10x
  - measurements: cas_round_wall_s=[0.02, 0.01, 0.01, 0.02, 0.02, 0.03, 0.02, 0.01, 0.01, 0.01], n=10
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:84-88; generator.yaml:103-104`
  - related: [S3T-TABLE-100](#s3t-table-100), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-104](#s3t-table-104),
    [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-193](#s3t-table-193) · hypotheses:
    H-T-145 · evidence: table/limits/policy-size-write-races
  - notes: H-T-145 confirmed: in 10/10 rounds two concurrent UpdateTableMetadataLocation calls with the same
    fresh versionToken returned exactly one 200 and one ConflictException 409 'Provided version token does not
    match the table version token.'; GetTableMetadataLocation then showed the winner's location and...
  - full notes: [details/S3T-TABLE-192.md](details/S3T-TABLE-192.md)

- <a id="s3t-table-302"></a>**S3T-TABLE-302** `idempotency` · impact medium · handled · verified 2026-10-14
  **Stale versionToken revived by another writer's rollback (A->B->A): DeleteTable(A0) -> 204, UpdateTableMetadataLocation(C, A0) -> 200**
  Table at location A with the controller's cached token A0. External UpdateTableMetadataLocation to B -> HTTP
  200; the controller's A0 while at B -> ConflictException (HTTP 409) 'Provided version token does not match
  the table version token.'; external rollback to A -> HTTP 200 (token A2 == A0: True); the external writer's
  B token after the rollback -> ConflictException (HTTP 409) 'Provided version token does not match the table
  version token.'. The controller's DeleteTable with the revived A0 -> HTTP 204 (table gone). On a second
  table with the same A->B->A history, UpdateTableMetadataLocation(C, A0) -> HTTP 200 and GetTable then points
  at C: True. Three fresh tables without metadataLocation had 3 distinct tokens.
  - ACK: annotation-shadow-state, custom_delete · ops: UpdateTableMetadataLocation, DeleteTable, GetTable ·
    fields: versionToken, metadataLocation
  - repro: CreateTable -> Update(A) token A0 -> Update(B, A0) -> Update(B, A0) 409 -> Update(A, B1) ->
    DeleteTable(versionToken=A0) / Update(C, A0)
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-192](#s3t-table-192) ·
    evidence: table/creative/token-revival
  - notes: Extends [S3T-TABLE-134](#s3t-table-134) (token = f(location)) with the consequence for optimistic locking:
    ConflictException means 'the location differs from the one your token encodes', NOT 'the table changed
    since you read it' - two intervening commits that end at the cached location are invisible, so a
    controller...
  - full notes: [details/S3T-TABLE-302.md](details/S3T-TABLE-302.md)

## Errors

- <a id="s3t-table-051"></a>**S3T-TABLE-051** `error-code` · impact high · handled · verified 2026-10-13
  **RenameTable with a stale or garbage versionToken -> ConflictException 409 'Provided version token does not match'; current token -> 204**
  UpdateTableMetadataLocation(v0, <warehouse>/metadata/<new file>) -> HTTP 200 (versionToken changed: True).
  RenameTable(newName=u, versionToken=<pre-update token>) -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'; RenameTable(newName=u, versionToken='xyz') ->
  ConflictException (HTTP 409) 'Provided version token does not match the table version token.'; GetTable then
  shows the name unchanged: True. RenameTable(newName=u, versionToken=<current token>) -> HTTP 204.
  - ACK: requeue, terminal_codes, custom_update · ops: RenameTable, UpdateTableMetadataLocation, GetTable ·
    fields: versionToken, metadataLocation
  - repro: CreateTable(schema) -> GetTable v0 -> UpdateTableMetadataLocation(v0, new path) -> v1 ->
    RenameTable(versionToken=v0), RenameTable(versionToken='xyz'), RenameTable(versionToken=v1)
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-100](#s3t-table-100), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-065](#s3t-table-065),
    [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-123](#s3t-table-123) · hypotheses: H-T-061 · evidence:
    table/error-taxonomy/rename-errors
  - notes: H-T-061 confirmed: stale token (from before UpdateTableMetadataLocation) and the garbage token
    'xyz' both -> ConflictException 409 with the same message 'Provided version token does not match the table
    version token.'; the table was not renamed; the current token -> 204. Side observation (outside...
  - full notes: [details/S3T-TABLE-051.md](details/S3T-TABLE-051.md)

- <a id="s3t-table-065"></a>**S3T-TABLE-065** `error-code` · impact medium · handled · verified 2026-10-13
  **DeleteTable with a garbage/other table's/random/upper-cased/2048-char versionToken -> ConflictException 409, table untouched; no 400 seen**
  On an existing table: versionToken=garbage_abc -> ConflictException (HTTP 409) 'Provided version token does
  not match the table version token.'; versionToken=other_tables_token -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'; versionToken=empty -> ParamValidationError
  (HTTP None) 'Parameter validation failed:
  Invalid length for parameter versionToken, value: 0, valid min length: 1';
  versionToken=same_length_hex_random -> ConflictException (HTTP 409) 'Provided version token does not match
  the table version token.'; versionToken=uppercased_own_token -> ConflictException (HTTP 409) 'Provided
  version token does not match the table version token.'; versionToken=very_long -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'. GetTable afterwards -> table exists with
  its original token: True.
  - ACK: custom_delete, requeue · ops: DeleteTable · fields: versionToken
  - repro: Create t1,t2; DeleteTable(t1, versionToken='abc'), (t1, token(t2)), (t1, ''), (t1, random hex),
    (t1, upper-cased own token), (t1, 2048 x 'a')
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-064](#s3t-table-064),
    [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-192](#s3t-table-192) · hypotheses: H-T-038 · evidence:
    table/idempotency/duplicate-create-double-delete
  - notes: H-T-038 confirmed: every server-side mismatch is ConflictException 'Provided version token does not
    match the table version token.'; there is no BadRequestException for a malformed token. The empty-string
    cell never reached the server: boto3 rejects it client-side (ParamValidationError, min length...
  - full notes: [details/S3T-TABLE-065.md](details/S3T-TABLE-065.md)

- <a id="s3t-table-103"></a>**S3T-TABLE-103** `error-code` · impact high · handled · verified 2026-10-13
  **UpdateTableMetadataLocation: superseded-location or garbage versionToken -> ConflictException 409; empty token -> ValidationException 400**
  Stale token (the CreateTable token after 3 later commits) -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'. Garbage token 'bogus-token-ackq' -> ConflictException (HTTP 409)
  'Provided version token does not match the table version token.'. Empty string token (SDK validation
  bypassed) -> ValidationException (HTTP 400) '1 validation error detected: Value at 'versionToken' failed to
  satisfy constraint: Member must have length greater than or equal to 1'. GetTableMetadataLocation after each
  still returned the last accepted token/location: True/True/True.
  - ACK: requeue, terminal_codes · ops: UpdateTableMetadataLocation · fields: versionToken
  - repro: UpdateTableMetadataLocation with the pre-commit token, with 'bogus-token-ackq', and with '' ->
    GetTableMetadataLocation
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-192](#s3t-table-192),
    [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-320](#s3t-table-320) · hypotheses: H-T-134 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-134 confirmed: stale and garbage tokens -> ConflictException 409 'Provided version token does
    not match the table version token.'; the empty-string token -> ValidationException 400 '1 validation error
    detected: Value at 'versionToken' failed to satisfy constraint: Member must have length greater...
  - full notes: [details/S3T-TABLE-103.md](details/S3T-TABLE-103.md)

- <a id="s3t-table-121"></a>**S3T-TABLE-121** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **GetTable/GTML/UTML on a missing table, namespace or table bucket -> NotFoundException 404, one message 'The specified table does not exist.'**
  Missing table in an existing namespace: GetTable -> NotFoundException (HTTP 404) 'The specified table does
  not exist.'; GetTableMetadataLocation -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  UpdateTableMetadataLocation(bogus token, s3://<nonexistent>/metadata/x.metadata.json) -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'. Missing namespace: NotFoundException (HTTP 404) 'The
  specified table does not exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.' /
  NotFoundException (HTTP 404) 'The specified table does not exist.'. Missing table bucket ARN:
  NotFoundException (HTTP 404) 'The specified table does not exist.' / NotFoundException (HTTP 404) 'The
  specified table does not exist.' / NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: exceptions.404 · ops: GetTable, GetTableMetadataLocation, UpdateTableMetadataLocation · fields: name,
    namespace, tableBucketARN
  - repro: Call the three operations with a bogus table name, a bogus namespace and a bogus table bucket ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](#s3t-table-066), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-099](#s3t-table-099), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-245](#s3t-table-245),
    [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-146 · evidence:
    table/idempotency/metadata-location-after-delete
  - notes: H-T-146 NotFound part QUALIFIED: all nine cells are NotFoundException (HTTP 404), but the message
    is identical for a missing table, a missing namespace and a missing table bucket ARN ('The specified table
    does not exist.'), so the message does NOT distinguish which level is missing (hypothesis...
  - full notes: [details/S3T-TABLE-121.md](details/S3T-TABLE-121.md)

- <a id="s3t-table-245"></a>**S3T-TABLE-245** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Ghost table name or UUID in an other-region ARN -> 400 region mismatch (check precedes lookup); same-named triple in 2 regions = 2 tables**
  us-west-2 endpoint, us-east-1 bucket ARN with a table name that exists in neither region and a tableArn with
  an all-zero UUID: GetTable(triple, ghost name), GetTableMetadataLocation(ghost name), GetTable(ghost
  tableArn) -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'. us-east-1 endpoint, us-west-2 ARNs: GetTable(triple, ghost name), GetTableMetadataLocation(ghost
  name), GetTable(ghost tableArn) -> BadRequestException (HTTP 400) 'The provided Region does not match the
  Region of the resource.'. The same-named bucket/namespace/table created in both regions are unrelated
  objects: table UUID equal = False (ARNs differ in segment(s) [3, 5]), namespaceId equal = False,
  tableBucketId equal = False.
  - ACK: terminal_codes, exceptions.404, is_arn_primary_key · ops: GetTable, GetTableMetadataLocation,
    CreateTable · fields: tableBucketARN, tableArn, name
  - repro: GetTable/GetTableMetadataLocation with a ghost name under the other region's bucket ARN;
    GetTable(tableArn) with the other region's bucket segment and a zero UUID
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-213, H-T-215 ·
    evidence: table/cross-region/foreign-region-arn-reads-writes
  - notes: Supports H-T-213 and the refutation of H-T-215: a wrong-region ARN can never be mistaken for a
    deleted table (the lookup is not reached). A controller must never map this 400 to 'recreate'.
  - full notes: [details/S3T-TABLE-245.md](details/S3T-TABLE-245.md)

## Request validation

- <a id="s3t-table-100"></a>**S3T-TABLE-100** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **UpdateTableMetadataLocation to a NON-EXISTENT object under the warehouse -> HTTP 200; the table now points at a dangling metadata file**
  On a table with a committed metadata.json (loc0), S3 HeadObject on
  <warehouseLocation>/metadata/00001-<uuid>.metadata.json -> 404 (HTTP 404) 'Not Found' (absent).
  UpdateTableMetadataLocation(current token, that path) -> HTTP 200 with keys ['metadataLocation', 'name',
  'namespace', 'tableARN', 'versionToken']. GetTableMetadataLocation/GetTable immediately after return the
  dangling path (True) with the new token; the service did not verify the object exists.
  - ACK: docs-only, hook:sdk_update_pre_build_request · ops: UpdateTableMetadataLocation,
    GetTableMetadataLocation · fields: metadataLocation, versionToken
  - repro: CreateTable with schema -> HeadObject <wh>/metadata/00001-<uuid>.metadata.json (404) ->
    UpdateTableMetadataLocation(token, that path) -> GetTableMetadataLocation
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-108](#s3t-table-108) · hypotheses: H-T-132 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-132 confirmed. n=1.

- <a id="s3t-table-105"></a>**S3T-TABLE-105** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation prefix check is a bare string-prefix match: sibling/other-bucket/foreign paths -> 400, but s3://<warehouse-host>x/.. -> 200**
  With the current token: another table's warehouse in the same bucket -> BadRequestException (HTTP 400) 'The
  specified metadata location is not valid.'; a table's warehouse in another table bucket ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; s3://<non-existent
  general-purpose bucket>/... -> BadRequestException (HTTP 400) 'The specified metadata location is not
  valid.'; a bucket whose name is the warehouse host plus 'x' (string-prefix match but a different bucket) ->
  HTTP 200. GetTableMetadataLocation unchanged after each rejection: {'sibling_table_warehouse': True,
  'other_table_bucket_warehouse': True, 'foreign_gp_bucket': True, 'scheme_upper_S3': True, 'scheme_s3a':
  True, 'https_virtual_hosted': True}.
  - ACK: terminal_codes, hook:sdk_update_pre_build_request · ops: UpdateTableMetadataLocation · fields:
    metadataLocation, warehouseLocation
  - repro: UpdateTableMetadataLocation(A, token, <B.warehouseLocation>/metadata/x.metadata.json), then
    s3://<other>/..., then s3://<A-host>x/...
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-106](#s3t-table-106), [S3T-TABLE-107](#s3t-table-107),
    [S3T-TABLE-135](#s3t-table-135) · hypotheses: H-T-131 · evidence: table/mutation-matrix/metadata-location-update
  - notes: H-T-131 confirmed for sibling-table, other-table-bucket and foreign-bucket paths
    (BadRequestException 400 'The specified metadata location is not valid.'; location unchanged after each).
    SURPRISE: a location whose bucket name is the warehouse host plus 'x' (string-prefix match, a different
    bucket)...
  - full notes: [details/S3T-TABLE-105.md](details/S3T-TABLE-105.md)

- <a id="s3t-table-107"></a>**S3T-TABLE-107** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation: x.metadata.json, x.metadata.json.gz, x.metadata.json.gzip, any folder under the warehouse -> 200; .json/.JSON/.avro -> 400**
  Under the table's own warehouse, with fresh tokens: '<WH>/metadata/x.json' -> BadRequestException (HTTP 400)
  'The specified metadata location is not valid.'; '<WH>/metadata/x.metadata.JSON' -> BadRequestException
  (HTTP 400) 'The specified metadata location is not valid.'; '<WH>/metadata/x.metadata.json.gzip' -> HTTP
  200; '<WH>/metadata/x.avro' -> BadRequestException (HTTP 400) 'The specified metadata location is not
  valid.'; '<WH>/metadata/x.metadata.json' -> HTTP 200; '<WH>/metadata/x.metadata.json.gz' -> HTTP 200;
  '<WH>/metadata/' -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; '<WH>'
  -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; '<WH>/metadata' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.';
  '<WH>/metadata/.metadata.json' -> HTTP 200; '<WH>/data/v2.metadata.json' -> HTTP 200;
  '<WH>/v2.metadata.json' -> HTTP 200; '<WH>/a/b/c/d.metadata.json' -> HTTP 200.
  - ACK: terminal_codes, hook:sdk_update_pre_build_request, docs-only · ops: UpdateTableMetadataLocation ·
    fields: metadataLocation
  - repro: UpdateTableMetadataLocation with each suffix/path variant under the warehouse (objects need not
    exist)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-106](#s3t-table-106),
    [S3T-TABLE-135](#s3t-table-135) · hypotheses: H-T-133, H-T-141 · evidence: table/mutation-matrix/metadata-location-update
  - notes: H-T-133 QUALIFIED: .metadata.json and .metadata.json.gz accepted; .json, .metadata.JSON
    (case-sensitive), .avro, trailing '/', bare warehouse and bare '/metadata' rejected with
    BadRequestException 400 'The specified metadata location is not valid.'; but '.metadata.json.gzip' was
    ACCEPTED (contrary...
  - full notes: [details/S3T-TABLE-107.md](details/S3T-TABLE-107.md)

- <a id="s3t-table-279"></a>**S3T-TABLE-279** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C087 PARTLY: <WH>/metadata/x.json -> 400 as documented, but x.metadata.json.gzip -> 200 (rule: contains '.metadata.json')**
  '<WH>/metadata/x.json' -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.' (so
  do x.avro, x.metadata.JSON, xmetadata.json and bare metadata.json). But '<WH>/metadata/x.metadata.json.gzip'
  -> HTTP 200, as do x.metadata.jsonfoo, x.metadata.json.tar, x.metadata.json.GZ, x.metadata.json.gz.bak and a
  trailing '/': the implemented rule is a case-sensitive 'contains the substring .metadata.json', not the
  documented 'ends with .metadata.json or .metadata.json.gz'.
  - ACK: docs-only · ops: UpdateTableMetadataLocation · fields: metadataLocation
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-107](#s3t-table-107), [S3T-TABLE-135](#s3t-table-135), [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-105](#s3t-table-105),
    [S3T-TABLE-106](#s3t-table-106) · evidence: service/static/doc-claims-3
  - notes: VERDICT: PARTLY - '.json' is rejected as documented but '.metadata.json.gzip' is accepted; a
    webhook enforcing the doc rule would be stricter than the service.

## Update granularity and ordering

- <a id="s3t-table-005"></a>**S3T-TABLE-005** `update-granularity` · impact high · handled · verified 2026-10-13
  **H-T-054 refuted: RenameTable(newName) does NOT bump GetTable.versionToken (byte-identical); pre-rename token still deletes the table** (hypothesis refuted; behavior confirmed)
  GetTable.versionToken is byte-identical before and after RenameTable(newName) (and equal to the CreateTable
  token: True); the pre-rename token stays valid for token-guarded calls (DeleteTable with it is exercised at
  the end of this probe).
  - ACK: custom_update, requeue · ops: RenameTable, GetTable, DeleteTable · fields: versionToken
  - repro: GetTable -> v0; RenameTable(newName); GetTable -> v1; compare; DeleteTable(versionToken=v0)
  - handling: handled via `generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-009](#s3t-table-009) ·
    hypotheses: H-T-054 · evidence: table/identity/rename-identity
  - notes: H-T-054 (contrarian: rename bumps versionToken) refuted: token unchanged by rename.

- <a id="s3t-table-123"></a>**S3T-TABLE-123** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Empty table: first UpdateTableMetadataLocation with the CreateTable token -> 200, GetTable gains metadataLocation; reusing that token -> 409**
  On a table created without metadata (GetTable has no metadataLocation key), S3 PutObject of a metadata.json
  copy under <warehouseLocation>/metadata/ -> HTTP 200; UpdateTableMetadataLocation(versionToken=<CreateTable
  response token>, that path) -> HTTP 200; afterwards GetTable has metadataLocation: True and
  GetTableMetadataLocation returns the path byte-exact: True; new token minted: True. A second
  UpdateTableMetadataLocation with the same CreateTable token -> ConflictException (HTTP 409) 'Provided
  version token does not match the table version token.'.
  - ACK: hook:sdk_create_post_set_output, custom_update · ops: CreateTable, UpdateTableMetadataLocation,
    GetTable, GetTableMetadataLocation · fields: versionToken, metadataLocation
  - repro: CreateTable(no metadata) -> PutObject <wh>/metadata/00000-<uuid>.metadata.json ->
    UpdateTableMetadataLocation(create token, path) -> GetTable -> repeat with the same token
  - measurements: update_latency_ms=[12], n=1
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:76-83; generator.yaml:105-110`
  - related: [S3T-TABLE-292](#s3t-table-292), [S3T-TABLE-277](#s3t-table-277), [S3T-TABLE-271](table.md#s3t-table-271), [S3T-TABLE-275](#s3t-table-275), [S3T-TABLE-282](#s3t-table-282), [S3T-TABLE-291](table.md#s3t-table-291),
    [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-034](#s3t-table-034), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-311](table-schema.md#s3t-table-311), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-102](#s3t-table-102),
    [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-320](#s3t-table-320) · hypotheses:
    H-T-140 · evidence: table/idempotency/metadata-location-after-delete
  - notes: H-T-140 confirmed: the CreateTable-response token is a valid CAS token for an empty table; after
    the first commit GetTable gains metadataLocation and a new token is minted; reusing the create token ->
    ConflictException 409 'Provided version token does not match the table version token.'. One...
  - full notes: [details/S3T-TABLE-123.md](details/S3T-TABLE-123.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-090"></a>**S3T-TABLE-090** `read-gap` · impact medium · handled · verified 2026-10-13
  **ListTables summary carries only createdAt, modifiedAt, name, namespace, tableARN, type; no versionToken/metadataLocation/warehouseLocation**
  ListTables(namespace, prefix=<name>) summary keys: ['createdAt', 'modifiedAt', 'name', 'namespace',
  'tableARN', 'type']. GetTable keys absent from the summary: ['createdBy', 'format', 'metadataLocation',
  'namespaceId', 'ownerAccountId', 'tableBucketId', 'versionToken', 'warehouseLocation']. Shared fields
  byte-identical to GetTable: True. namespace is a 1-element list in both; timestamps compare equal at
  microsecond precision: True.
  - ACK: list_operation.match_fields, custom_find · ops: ListTables, GetTable · fields: tables, versionToken,
    metadataLocation, warehouseLocation, format, namespaceId, tableBucketId, modifiedAt
  - repro: CreateTable(schema) -> ListTables(namespace, prefix=name) and GetTable(triple); compare key sets
    and values
  - handling: handled via `generator.yaml:10-13; generator.yaml:31-39; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-018](table-schema.md#s3t-table-018), [S3T-TABLE-310](table-schema.md#s3t-table-310), [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-023](#s3t-table-023), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-028](#s3t-table-028) ·
    hypotheses: H-T-076 · evidence: table/weird-inputs/list-tables-bounds
  - notes: H-T-076 confirmed (predicted summary = namespace, name, type, tableARN, createdAt, modifiedAt [+
    namespaceId/tableBucketId]; no versionToken/metadataLocation/warehouseLocation/format).
  - full notes: [details/S3T-TABLE-090.md](details/S3T-TABLE-090.md)

- <a id="s3t-table-106"></a>**S3T-TABLE-106** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **metadataLocation: 'S3://', 's3a://', https:// -> 400; '//' path, '%20' and a literal space -> 200, read back byte-exact (no normalization)**
  scheme_upper_S3: sent
  'S3://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3/metadata/x.metadata.json' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; scheme_s3a: sent
  's3a://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3/metadata/x.metadata.json' ->
  BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; https_virtual_hosted: sent
  'https://1ae9703d-2fb4-4723-ewxhmr5r89mkgmuaf63a1ddwsxxxeusw2b--table-s3.s3.us-west-2.amazonaws.com/metadata/x.metadata.json'
  -> BadRequestException (HTTP 400) 'The specified metadata location is not valid.'; double_slashes: sent
  '<WH>//metadata//x.metadata.json' -> HTTP 200, GetTableMetadataLocation returns
  '<WH>//metadata//x.metadata.json' (byte-identical: True); percent20: sent
  '<WH>/metadata/a%20b.metadata.json' -> HTTP 200, GetTableMetadataLocation returns
  '<WH>/metadata/a%20b.metadata.json' (byte-identical: True); space: sent '<WH>/metadata/a b.metadata.json' ->
  HTTP 200, GetTableMetadataLocation returns '<WH>/metadata/a b.metadata.json' (byte-identical: True).
  - ACK: docs-only, compare.is_ignored+delta_pre_compare · ops: UpdateTableMetadataLocation,
    GetTableMetadataLocation · fields: metadataLocation
  - repro: UpdateTableMetadataLocation with each rewritten location string -> GetTableMetadataLocation
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-107](#s3t-table-107),
    [S3T-TABLE-135](#s3t-table-135) · hypotheses: H-T-142 · evidence: table/mutation-matrix/metadata-location-update
  - notes: H-T-142 confirmed. One attempt each.
  - full notes: [details/S3T-TABLE-106.md](details/S3T-TABLE-106.md)

- <a id="s3t-table-315"></a>**S3T-TABLE-315** `read-gap` · impact high · handled · verified 2026-10-14
  **CreateTable returns only {tableARN, versionToken}; GetTable adds 12 members (metadataLocation only once the table has metadata)**
  CreateTableBucket wire body keys ['arn']; CreateNamespace -> HTTP 200 keys ['namespace', 'tableBucketARN']
  (echoes the parent tableBucketARN: True); GetNamespace right after -> HTTP 200 keys ['createdAt',
  'createdBy', 'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId']. CreateTable(ICEBERG, no
  metadata) and CreateTable(ICEBERG, 2-field schema) -> HTTP 200/200 with wire keys ['tableARN',
  'versionToken'] (no tableBucketARN member; the tableARN starts with '<bucket ARN>/': True). GetTable right
  after Create (HTTP 200) had keys ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace',
  'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']
  for the table without metadata and ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt',
  'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] for the table with a schema: present only in the Get: ['createdAt', 'createdBy',
  'format', 'modifiedAt', 'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId', 'type',
  'warehouseLocation'] (+ metadataLocation only with metadata). Modeled GetTable members absent from both wire
  bodies on a fresh table: ['managedByService', 'managedTableInformation', 'modifiedBy']. versionToken in Get ==
  Create: True. ListTables entry wire keys: ['createdAt', 'modifiedAt', 'name', 'namespace', 'tableARN',
  'type'] (type equals GetTable.type: True). At +3 s the GetTable key sets were unchanged: True.
  - ACK: hook:sdk_create_post_set_output, is_read_only, late_initialize, is_arn_primary_key · ops:
    CreateNamespace, GetNamespace, CreateTable, GetTable, ListTables · fields: tableARN, versionToken,
    metadataLocation, warehouseLocation, namespaceId, tableBucketId, createdAt, createdBy, modifiedAt,
    modifiedBy, type, format, managedByService
  - repro: CreateTableBucket -> CreateNamespace -> GetNamespace -> CreateTable (no metadata) + CreateTable
    (schema) -> GetTable each immediately and at +3 s -> ListTables; compare raw JSON key sets
  - measurements: n_tables=2, n=2
  - handling: handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-014](#s3t-table-014), [S3T-TABLE-031](#s3t-table-031), [S3T-TABLE-033](table-schema.md#s3t-table-033), [S3T-TABLE-309](table-maintenance.md#s3t-table-309), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-001](#s3t-table-001) ·
    evidence: table/round-trip/fu-12-create-vs-get-members
  - notes: FU-12 (and FU-10: ListTables.type agrees with GetTable.type). Exact wire names are camelCase with
    'Id' (namespaceId, tableBucketId, ownerAccountId). modifiedBy and managedByService are NOT returned on a
    fresh table even though the model has them (see the FU-11 finding for when modifiedBy appears);...
  - full notes: [details/S3T-TABLE-315.md](details/S3T-TABLE-315.md)

- <a id="s3t-table-316"></a>**S3T-TABLE-316** `server-default` · impact medium · handled · verified 2026-10-14
  **modifiedBy is absent on a fresh table and still absent after UpdateTableMetadataLocation; RenameTable populates it (== caller)**
  GetNamespace: ownerAccountId present True (== caller account True), createdBy present True (== caller True),
  no modifiedBy/modifiedAt member. GetTable on both fresh tables: ownerAccountId present/== caller True/True,
  createdBy present/== caller True/True, modifiedBy present False, modifiedAt == createdAt True; still absent
  at +3 s: True. UpdateTableMetadataLocation(versionToken, <warehouse>/metadata/00001-<uuid>.metadata.json) ->
  HTTP 200 (wire keys ['metadataLocation', 'name', 'namespace', 'tableARN', 'versionToken']): afterwards
  modifiedBy present False (== caller False), modifiedAt changed True, versionToken changed True (Update
  response versionToken == GetTable's: True). RenameTable(newName only, same namespace) -> HTTP 204:
  afterwards modifiedBy present True (== caller True), modifiedAt changed True, versionToken changed False.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare, late_initialize · ops: GetNamespace, GetTable,
    UpdateTableMetadataLocation, RenameTable · fields: ownerAccountId, createdBy, modifiedBy, modifiedAt,
    versionToken
  - repro: CreateNamespace+GetNamespace; CreateTable x2 + GetTable at 0 s and +3 s;
    UpdateTableMetadataLocation on one table, RenameTable(newName) on the other; GetTable after each
  - measurements: n_tables=2, n=2
  - handling: handled via `test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-034](#s3t-table-034), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-311](table-schema.md#s3t-table-311), [S3T-TABLE-005](#s3t-table-005),
    [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104),
    [S3T-TABLE-124](#s3t-table-124) · evidence: table/round-trip/fu-12-create-vs-get-members
  - notes: FU-11. Single-principal lab: createdBy/modifiedBy/ownerAccountId all equal the caller's 12-digit
    account id, so a modifiedBy change to a different principal is not observable (only presence and equality
    were recorded). UpdateTableMetadataLocation bumped modifiedAt and versionToken but did NOT...
  - full notes: [details/S3T-TABLE-316.md](details/S3T-TABLE-316.md)

## Response fidelity and consistency

- <a id="s3t-table-007"></a>**S3T-TABLE-007** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **Across 3 renames (name, namespace, both) GetTable(tableArn) changed only: modifiedAt, modifiedBy, name, namespace, namespaceId**
  RenameTable(ns2,t2, newNamespaceName=ns1, newName=t3) (both fields) -> HTTP 204. Keys that changed in
  GetTable(tableArn) across the three renames: modifiedAt, modifiedBy, name, namespace, namespaceId; keys
  never changed: createdAt, createdBy, format, metadataLocation, ownerAccountId, tableARN, tableBucketId,
  type, versionToken, warehouseLocation. After the combined call namespaceId equals
  GetNamespace(ns1).namespaceId: True; name=newName: True.
  - ACK: compare.is_ignored+delta_pre_compare, is_arn_primary_key · ops: RenameTable, GetTable · fields: name,
    namespace, namespaceId, modifiedAt, modifiedBy, versionToken, createdAt, createdBy, tableBucketId,
    metadataLocation, warehouseLocation
  - repro: GetTable(tableArn) body A -> RenameTable(both fields) -> GetTable(tableArn) body B -> diff;
    GetNamespace(destination)
  - measurements: rename_both_latency_ms=[152], n=1
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-111](#s3t-table-111), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-005](#s3t-table-005),
    [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-124](#s3t-table-124), [S3T-TABLE-019](#s3t-table-019) · hypotheses:
    H-T-197, H-T-060 · evidence: table/identity/rename-identity
  - notes: H-T-197 confirmed (hypothesis: only name, namespace, namespaceId, modifiedAt, modifiedBy,
    versionToken change). H-T-060 (both fields accepted in one call): confirmed (204).

- <a id="s3t-table-014"></a>**S3T-TABLE-014** `response-fidelity` · impact high · handled · verified 2026-10-13
  **CreateTable without metadata: GetTable/GetTableMetadataLocation omit metadataLocation (key absent), warehouseLocation set...**
  CreateTable(format=ICEBERG, no metadata) -> HTTP 200 with keys ['tableARN', 'versionToken']. Immediately and
  at 5 s GetTable returned keys ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace',
  'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']
  (metadataLocation key absent at 0 s and 5 s; warehouseLocation
  's3://3e467aed-5b02-470e-y5yt6ojsai4r8h5gyw1iypjgr3znhusw2b--table-s3'; type=customer; managedByService
  <absent>; managedTableInformation <absent>). GetTableMetadataLocation -> HTTP 200 with keys ['versionToken',
  'warehouseLocation'] (metadataLocation absent). versionToken identical across CreateTable response, GetTable
  and GetTableMetadataLocation: True.
  - ACK: hook:sdk_read_one_post_set_output, docs-only · ops: CreateTable, GetTable, GetTableMetadataLocation ·
    fields: metadataLocation, warehouseLocation, versionToken, type
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(format=ICEBERG) with no metadata -> GetTable +
    GetTableMetadataLocation at 0 s and 5 s
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:51-56; pkg/resource/table/sdk.go:489-491; generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-031](#s3t-table-031), [S3T-TABLE-033](table-schema.md#s3t-table-033), [S3T-TABLE-309](table-maintenance.md#s3t-table-309), [S3T-TABLE-315](#s3t-table-315) · hypotheses: H-T-101, H-T-143,
    H-T-129 · evidence: table/round-trip/metadata-create-readback
  - notes: H-T-101 confirmed (metadataLocation key absent, not empty string; warehouseLocation already set).
    H-T-143 confirmed (GetTableMetadataLocation is 200 on an empty table, not 404). H-T-129 qualified: a
    CreateTable-created table reads type=customer; managedByService <absent>; no aws-managed table was...
  - full notes: [details/S3T-TABLE-014.md](details/S3T-TABLE-014.md)

- <a id="s3t-table-031"></a>**S3T-TABLE-031** `response-fidelity` · impact high · handled · verified 2026-10-13
  **Fresh table without metadata: raw GetTable JSON has metadataLocation absent, managedByService absent, managedTableInformation absent...**
  Raw GetTable(tableArn) body for a table created with format=ICEBERG and no metadata (HTTP 200): keys
  ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace', 'namespaceId', 'ownerAccountId',
  'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']. metadataLocation: absent;
  warehouseLocation: present(str) (s3:// URI: True); managedByService: absent value="<absent>";
  managedTableInformation: absent; type=customer; namespace raw=["<ns>"]. GetTableMetadataLocation -> HTTP 200
  with metadataLocation absent, warehouseLocation present(str), versionToken equal to the create token: True.
  - ACK: compare.nil_equals_zero_value, is_read_only, late_initialize · ops: GetTable,
    GetTableMetadataLocation · fields: metadataLocation, warehouseLocation, managedByService,
    managedTableInformation, type, namespace
  - repro: CreateTable(format=ICEBERG) without metadata; GetTable(tableArn) and inspect the raw HTTP body
    (botocore drops JSON nulls); GetTableMetadataLocation
  - handling: handled via `generator.yaml:10-13; generator.yaml:31-39; generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-014](#s3t-table-014), [S3T-TABLE-033](table-schema.md#s3t-table-033), [S3T-TABLE-309](table-maintenance.md#s3t-table-309), [S3T-TABLE-315](#s3t-table-315), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-001](#s3t-table-001),
    [S3T-TABLE-136](table-replication-status.md#s3t-table-136), [S3T-TABLE-223](table-replication-status.md#s3t-table-223), [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](table-replication.md#s3t-table-177) · hypotheses: H-T-029, H-T-033, H-U-189,
    H-T-012 · evidence: table/round-trip/fresh-table-fields
  - notes: H-T-029 confirmed (metadataLocation absent on both reads; warehouseLocation always present).
    H-U-189 REFUTED: managedByService is not returned as false, it is absent, so both managed* members are
    omitted symmetrically (managedTableInformation absent). H-T-033 confirmed (type=customer, managed* keys...
  - full notes: [details/S3T-TABLE-031.md](details/S3T-TABLE-031.md)

- <a id="s3t-table-034"></a>**S3T-TABLE-034** `response-fidelity` · impact high · handled · verified 2026-10-13
  **versionToken from CreateTable is unchanged across 10 GetTable reads over 54 s with no writes; 20-char hex string, not an integer**
  CreateTable(no metadata) returned versionToken of length 20 (charset hex, shape hhhhhhhhhhhhhhhhhhhh,
  integer-parsable: False). 10 GetTable reads at 6 s intervals returned 1 distinct token(s); all equal to the
  create-time token: True. GetTableMetadataLocation returned the same token: True. A second table's token has
  length 20 and equals the first: False.
  - ACK: is_read_only, custom_delete · ops: CreateTable, GetTable, GetTableMetadataLocation · fields:
    versionToken
  - repro: CreateTable; GetTable x10 at 6 s intervals; compare versionToken strings
  - measurements: token_stability_reads=10, span_s=54, n=10, token_len=20
  - handling: handled via `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
  - related: [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-311](table-schema.md#s3t-table-311), [S3T-TABLE-316](#s3t-table-316) · hypotheses: H-T-028 ·
    evidence: table/round-trip/fresh-table-fields
  - notes: H-T-028 confirmed: the create-time token can be persisted and reused for
    DeleteTable/UpdateTableMetadataLocation as long as nobody else commits; do not parse it.

- <a id="s3t-table-101"></a>**S3T-TABLE-101** `response-fidelity` · impact high · handled · verified 2026-10-13
  **UpdateTableMetadataLocation response versionToken/metadataLocation == GetTable and GTML at 0/1/5 s; modifiedAt advanced, createdAt unchanged**
  UpdateTableMetadataLocation -> HTTP 200, response keys ['metadataLocation', 'name', 'namespace', 'tableARN',
  'versionToken']. Token changed vs the pre-update token: True. Response versionToken equals
  GetTable.versionToken at 0 s: True and GetTableMetadataLocation.versionToken: True; response
  metadataLocation equals the reads: True/True; reads at 1 s and 5 s identical: True. GetTable.modifiedAt
  2026-10-13 22:31:29.246753+00:00 -> 2026-10-13 22:33:07.628423+00:00 (createdAt unchanged: True);
  warehouseLocation unchanged: True.
  - ACK: hook:sdk_update_post_set_output · ops: UpdateTableMetadataLocation, GetTable,
    GetTableMetadataLocation · fields: versionToken, metadataLocation, modifiedAt, createdAt
  - repro: UpdateTableMetadataLocation(T0, loc1) -> GetTable + GetTableMetadataLocation at 0 s, 1 s, 5 s ->
    diff versionToken/metadataLocation/modifiedAt/createdAt
  - measurements: update_latency_ms=[13], n=1
  - handling: handled via `generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-102](#s3t-table-102), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-124](#s3t-table-124),
    [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-137 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-137 confirmed. n=1 update, 6 reads. Full original title: UpdateTableMetadataLocation response:
    new versionToken == GetTable/GetTableMetadataLocation at 0/1/5 s; metadataLocation echoed exactly;
    modifiedAt advanced, createdAt unchanged

- <a id="s3t-table-102"></a>**S3T-TABLE-102** `eventual-consistency` · impact medium · handled · verified 2026-10-13
  **After UpdateTableMetadataLocation, 38 alternating GetTable/GTML samples over ~8 s: 0 stale token/location; ListTables modifiedAt == GetTable**
  Two updates (2 s apart), each followed by alternating GetTable / GetTableMetadataLocation reads every 200 ms
  (38 samples total, 0 errors): 0 returned a versionToken or metadataLocation older than the latest update
  response. ListTables read ~2.7 s after the second update shows modifiedAt 2026-10-13 22:33:19.776587+00:00,
  equal to GetTable.modifiedAt: True.
  - ACK: none · ops: UpdateTableMetadataLocation, GetTable, GetTableMetadataLocation, ListTables · fields:
    versionToken, metadataLocation, modifiedAt
  - repro: UpdateTableMetadataLocation -> alternate GetTable/GetTableMetadataLocation every 200 ms for ~5 s ->
    second update -> poll ~3 s -> ListTables
  - measurements: samples=38, stale_samples=0, update_latency_ms=[13, 16], n=2
  - handling: handled via `generator.yaml:76-83; generator.yaml:105-110`
  - related: [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-015](table-schema.md#s3t-table-015) · hypotheses: H-T-138 · evidence:
    table/mutation-matrix/metadata-location-update
  - notes: H-T-138 confirmed (polled at 200 ms, not 100 ms, to respect the per-probe read budget). n=2
    updates, 38 reads. Full original title: After UpdateTableMetadataLocation, 38 alternating
    GetTable/GetTableMetadataLocation samples over ~8 s showed 0 stale token/location; ListTables modifiedAt ==
    GetTable

## Sub-resources

- <a id="s3t-table-320"></a>**S3T-TABLE-320** `sub-resource-api` · impact medium · handled · verified 2026-10-14
  **Renamed table accepts UpdateTableMetadataLocation (200) and TagResource (200); stale versionToken -> ConflictException 409**
  On the renamed table (new triple) UpdateTableMetadataLocation(versionToken=<current from GetTable>,
  metadataLocation=<warehouseLocation>/metadata/00001-<uuid>.metadata.json (no object was written there)) ->
  HTTP 200; the same call addressed to the OLD triple -> NotFoundException (HTTP 404) 'The specified table
  does not exist.'; with the pre-rename/stale versionToken -> ConflictException (HTTP 409) 'Provided version
  token does not match the table version token.'. GetTable after the update: metadataLocation changed to the
  sent URI, versionToken changed: True, modifiedAt changed: True. TagResource(resourceArn=<tableARN>) -> HTTP
  200, ListTagsForResource -> HTTP 200 tags=['ackq-run', 'after-rename']. RenameTable with the stale
  versionToken -> ConflictException (HTTP 409) 'Provided version token does not match the table version
  token.'; RenameTable into a namespace that does not exist -> NotFoundException (HTTP 404) 'The specified
  destination namespace does not exist.'. Across create/rename/update: createdBy==caller account True,
  modifiedBy present True and ==caller account True, ownerAccountId==caller True.
  - ACK: custom_update, exceptions.404, terminal_codes · ops: UpdateTableMetadataLocation, TagResource,
    ListTagsForResource, RenameTable, GetTable · fields: versionToken, metadataLocation, modifiedAt,
    modifiedBy, createdBy, ownerAccountId, tags
  - repro: RenameTable -> GetTable(new triple) for versionToken -> UpdateTableMetadataLocation(new triple,
    token, <warehouse>/metadata/00001-<uuid>.metadata.json); same on old triple; with stale token;
    TagResource(tableARN); RenameTable(stale token); RenameTable(newNamespaceName=<missing>)
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-100](#s3t-table-100), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-103](#s3t-table-103),
    [S3T-TABLE-098](#s3t-table-098), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-111](#s3t-table-111), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-007](#s3t-table-007),
    [S3T-TABLE-005](#s3t-table-005), [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-124](#s3t-table-124), [S3T-TABLE-019](#s3t-table-019),
    [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-245](#s3t-table-245) · evidence: table/sub-resources/fu-07-rename-kms-table
  - notes: FU-07 (mutations after rename) and FU-11 (modifiedBy/createdBy/ownerAccountId are all the caller's
    12-digit account id in a single-principal lab, so a change of modifiedBy to another principal is not
    observable: only presence and equality were recorded; modifiedAt/versionToken changes are recorded...
  - full notes: [details/S3T-TABLE-320.md](details/S3T-TABLE-320.md)

## Tags

- <a id="s3t-table-124"></a>**S3T-TABLE-124** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **TagResource/UntagResource bump GetTable.modifiedAt (not versionToken); the pre-tag versionToken still deletes the table**
  GetTable before, after TagResource({z:1}) (-> HTTP 200) and after UntagResource(['z']) (-> HTTP 204):
  modifiedAt CHANGED 2026-10-13 22:37:45.178044+00:00 -> 2026-10-13 22:37:45.660173+00:00 -> 2026-10-13
  22:37:45.807230+00:00, versionToken identical, modifiedBy identical (before: modifiedAt=2026-10-13
  22:37:45.178044+00:00 versionToken=<20 chars>).
  - ACK: compare.is_ignored+delta_pre_compare, tags.custom-sync · ops: TagResource, UntagResource, GetTable ·
    fields: modifiedAt, versionToken, modifiedBy, tags
  - repro: GetTable; TagResource({z:1}); GetTable; UntagResource(['z']); GetTable; compare
    modifiedAt/versionToken (DeleteTable with the original token is at the end of this probe)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-019](#s3t-table-019),
    [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-096 · evidence: table/idempotency/untag-and-post-delete
  - notes: H-T-096 qualified: the contrarian claim holds for modifiedAt (advanced by ~0.5 s after TagResource
    and again after UntagResource, 1 trial each) but not for versionToken (byte-identical, 20 chars) nor
    modifiedBy; DeleteTable guarded with the CreateTable-time versionToken after these tag writes ->...
  - full notes: [details/S3T-TABLE-124.md](details/S3T-TABLE-124.md)

## Delete semantics

- <a id="s3t-table-089"></a>**S3T-TABLE-089** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable (no versionToken, empty table) -> 204 empty body in 20/20; GetTable composite and by ARN -> 404 on the next call in 20/20**
  20 trials: DeleteTable without versionToken on a freshly created empty table -> HTTP 204 (parsed body empty:
  True; call latency 105-168 ms). GetTable composite issued 7-46 ms later -> NotFoundException in 20/20;
  GetTable(tableArn) issued 10-58 ms later -> NotFoundException in 20/20. Outcomes seen after delete:
  ['NotFoundException (HTTP 404)'].
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable · fields: versionToken
  - repro: Loop 20x: CreateTable; DeleteTable(composite, no token); GetTable composite and by tableArn at once
  - measurements: after_delete_composite_first_call_ms=[9, 27, 9, 24, 13, 26, 8, 20, 8, 18, 7, 22, 8, 18, 8,
    18, 46, 19, 8, 22], after_delete_arn_first_call_ms=[20, 13, 20, 16, 23, 17, 18, 11, 17, 11, 17, 11, 17,
    11, 19, 10, 58, 11, 19, 13], delete_latency_ms=[130, 129, 119, 128, 111, 111, 140, 168, 111, 120, 118,
    119, 116, 105, 121, 110, 132, 135, 115, 121], n=20, still_visible_windows=[]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](#s3t-table-066), [S3T-TABLE-099](#s3t-table-099), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-067](#s3t-table-067),
    [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-028](#s3t-table-028), [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-231](table.md#s3t-table-231) · hypotheses: H-T-036 · evidence:
    table/consistency-windows/create-delete-visibility
  - notes: H-T-036 confirmed: delete of an empty table is synchronous and the token is optional; the
    controller may finalize on 204 without polling (empty tables; non-empty tables are covered by
    table/mutation-matrix/delete-version-token).
  - full notes: [details/S3T-TABLE-089.md](details/S3T-TABLE-089.md)

- <a id="s3t-table-098"></a>**S3T-TABLE-098** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation (location moved) -> 409; current token -> 204**
  Table created with metadata.iceberg.schema (token T0, metadataLocation 00000-<uuid>.metadata.json).
  UpdateTableMetadataLocation(versionToken=T0, metadataLocation=<copied 00001 file>) -> HTTP 200 (token
  changed: True; response keys ['metadataLocation', 'name', 'namespace', 'tableARN', 'versionToken']; GetTable
  token equals the update response token: True; GetTable metadataLocation equals the sent location: True;
  modifiedAt changed: True; modifiedBy now present: False). Then UpdateTableMetadataLocation(T0) ->
  ConflictException (HTTP 409) 'Provided version token does not match the table version token.';
  DeleteTable(versionToken=T0) -> ConflictException (HTTP 409) 'Provided version token does not match the
  table version token.'; GetTable -> table still exists with token T1: True; DeleteTable(versionToken=T1) ->
  HTTP 204; GetTable -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: custom_delete, requeue · ops: DeleteTable, UpdateTableMetadataLocation, GetTable · fields:
    versionToken, metadataLocation, modifiedAt, modifiedBy
  - repro: CreateTable with schema (T0) -> S3 CopyObject of the 00000 metadata file to 00001 ->
    UpdateTableMetadataLocation(T0) -> DeleteTable(T0) -> GetTable -> DeleteTable(T1)
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-134](#s3t-table-134), [S3T-TABLE-108](#s3t-table-108), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-302](#s3t-table-302), [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-192](#s3t-table-192),
    [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-320](#s3t-table-320) · hypotheses: H-T-037, H-T-030 · evidence:
    table/mutation-matrix/delete-version-token
  - notes: H-T-037 confirmed. A controller caching the create-time token must re-read it (GetTable) before
    DeleteTable or omit versionToken entirely (a tokenless delete succeeds, see the idempotency probe).
    H-T-030 re-confirmed: the schema-at-create table had metadataLocation and a token usable for the first...
  - full notes: [details/S3T-TABLE-098.md](details/S3T-TABLE-098.md)

- <a id="s3t-table-099"></a>**S3T-TABLE-099** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable of a table with committed metadata + a data object -> HTTP 204 in 3/3; GetTable -> NotFoundException (HTTP 404) on the next call**
  3 trials: table with schema, one UpdateTableMetadataLocation commit and one 2 KB object under data/
  (['None'] objects listed in the warehouse via S3 ListObjectsV2 before delete). DeleteTable(current token) ->
  HTTP 204 (latency [140, 192, 118] ms). First GetTable composite [9, 11, 9] ms later -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; first GetTable(tableArn) -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. After the delete, S3 ListObjectsV2 on the warehouse ->
  MethodNotAllowed (HTTP 405) 'The specified method is not allowed against this resource.'; HeadObject on the
  data file -> HTTP 200. Data-plane access to the warehouse location through the S3 API: CopyObject ->
  MethodNotAllowed (HTTP 405) 'The specified method is not allowed against this resource.', ListObjectsV2 ->
  MethodNotAllowed (HTTP 405) (before and after the delete); GetObject, PutObject and HeadObject -> HTTP 200;
  HeadObject on the data object still returned 200 immediately AFTER the table was deleted (3/3), so the
  storage is drained asynchronously even though the table itself is gone from GetTable at once.
  - ACK: exceptions.404, custom_delete · ops: DeleteTable, GetTable, UpdateTableMetadataLocation · fields:
    versionToken, warehouseLocation
  - repro: CreateTable with schema; CopyObject metadata; UpdateTableMetadataLocation; PutObject data/...;
    DeleteTable(token); GetTable both forms; poll at 0.5 s if still visible
  - measurements: delete_latency_ms=[140, 192, 118], first_get_after_delete_ms=[9, 11, 9],
    visible_windows_s=[], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-066](#s3t-table-066), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-121](#s3t-table-121) · hypotheses: H-T-040 ·
    evidence: table/mutation-matrix/delete-version-token
  - notes: H-T-040 (delete becomes asynchronous once the table is non-empty) refuted: GetTable returned 404 on
    the very next call in every trial. Only 3 trials with a tiny data object; a large table may behave
    differently. Surprise for follow-up: the warehouse objects remain readable via HeadObject right after...
  - full notes: [details/S3T-TABLE-099.md](details/S3T-TABLE-099.md)

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
  - related: [S3T-TABLE-066](#s3t-table-066), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-099](#s3t-table-099), [S3T-TABLE-129](#s3t-table-129), [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-067](#s3t-table-067),
    [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-028](#s3t-table-028) · hypotheses: H-T-211, H-T-212 · evidence:
    table/idempotency/metadata-location-after-delete
  - notes: H-T-211 confirmed: the first GetTableMetadataLocation after delete (0.22 s) -> NotFoundException
    404 'The specified table does not exist.' (same wire code and message as the missing-table case);
    CreateTable of the same bucket/namespace/name -> 200 with a different warehouseLocation (new UUID
    host),...
  - full notes: [details/S3T-TABLE-122.md](details/S3T-TABLE-122.md)

- <a id="s3t-table-129"></a>**S3T-TABLE-129** `delete-semantics` · impact high · handled · verified 2026-10-13
  **After DeleteTable: TagResource/ListTagsForResource on the old ARN -> 404 from t+0, UntagResource -> 500 (2/2); re-create: new ARN, tags {}**
  DeleteTable(versionToken=<CreateTable token>) -> HTTP 204. Immediately after: TagResource(old ARN) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; UntagResource(old ARN) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'; ListTagsForResource(old
  ARN) -> NotFoundException (HTTP 404) 'The specified table does not exist.'. ListTagsForResource polled every
  0.5 s for 4 s: statuses ['NotFoundException']. At t+5 s: NotFoundException / InternalServerErrorException /
  NotFoundException. CreateTable with the same (bucket, namespace, name) -> HTTP 200, new tableARN differs
  from the old: True; ListTagsForResource(new ARN) -> {}; ListTagsForResource(old ARN) after the re-create ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: is_arn_primary_key, tags.custom-sync, exceptions.404 · ops: DeleteTable, TagResource, UntagResource,
    ListTagsForResource, CreateTable · fields: resourceArn, tableARN, tags, versionToken
  - repro: CreateTable(tags) -> DeleteTable -> Tag/Untag/ListTags on the old ARN at t+0, every 0.5s, t+5s ->
    CreateTable same triple -> compare ARN, ListTagsForResource both ARNs
  - measurements: post_delete_list_statuses=[NotFoundException, NotFoundException, NotFoundException,
    NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException], n=8
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-066](#s3t-table-066), [S3T-TABLE-089](#s3t-table-089), [S3T-TABLE-099](#s3t-table-099), [S3T-TABLE-122](#s3t-table-122), [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-067](#s3t-table-067),
    [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-028](#s3t-table-028), [S3T-TABLE-133](table-schema.md#s3t-table-133), [S3T-TABLE-036](table-schema.md#s3t-table-036), [S3T-TABLE-193](#s3t-table-193) · hypotheses: H-T-095, H-T-096 ·
    evidence: table/idempotency/untag-and-post-delete
  - notes: H-T-095 qualified: no window - ListTagsForResource and TagResource on the deleted table's ARN
    return NotFoundException 404 'The specified table does not exist.' from the first call after DeleteTable
    returned (10 reads over 7 s); a same-triple CreateTable returns a different tableARN (new UUID) with...
  - full notes: [details/S3T-TABLE-129.md](details/S3T-TABLE-129.md)

- <a id="s3t-table-323"></a>**S3T-TABLE-323** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Re-verification [S3T-TABLE-122](#s3t-table-122): DeleteTable 204 and 404s from the first sample hold (n=2); the old metadata.json is NOT readable after delete**
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
  - related: [S3T-TABLE-122](#s3t-table-122) · evidence: table/creative/reverify-set-b
  - notes: Independent re-verification (set B), n=2 per variant (original n=1). All API-side clauses
    reproduce. The HeadObject-200 clause is at best transient: readability of the warehouse object is revoked
    with the table (403, so existence is unobservable from this account); a prior read keeps it readable
    for...
  - full notes: [details/S3T-TABLE-323.md](details/S3T-TABLE-323.md)

## Quotas and rate limits

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
  - related: [S3T-TABLE-064](#s3t-table-064), [S3T-TABLE-065](#s3t-table-065), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-003](#s3t-table-003), [S3T-TABLE-009](#s3t-table-009), [S3T-TABLE-111](#s3t-table-111),
    [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-005](#s3t-table-005), [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104),
    [S3T-TABLE-124](#s3t-table-124), [S3T-TABLE-019](#s3t-table-019), [S3T-TABLE-121](#s3t-table-121), [S3T-TABLE-245](#s3t-table-245), [S3T-TABLE-133](table-schema.md#s3t-table-133), [S3T-TABLE-036](table-schema.md#s3t-table-036), [S3T-TABLE-129](#s3t-table-129) ·
    hypotheses: H-T-200, H-T-201 · evidence: table/limits/policy-size-write-races
  - notes: H-T-200 qualified: 30 sequential RenameTable alternating two names with no versionToken and no
    pause -> 30 x 204 in 4.5 s total (121-182 ms each), no 409/429 - but the GetTable versionToken after every
    rename was IDENTICAL (1 distinct token across 30 reads), refuting the 'versionToken differs after...
  - full notes: [details/S3T-TABLE-193.md](details/S3T-TABLE-193.md)

## Handling gaps (bugs to file)

- [S3T-TABLE-006](#s3t-table-006) - RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId
  follows the destination, tableARN unchanged (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven. (1) After an out-of-band
  RenameTable(newNamespaceName) the ARN-keyed read writes the new namespace into latest.Spec.Namespace, the
  delta flags Spec.Namespace every reconcile, and customUpdateTable silently ignores it - no RenameTable back,
  no condition, Synced=True - so the CR permanently reports a namespace the table is no longer in (not the
  feared immutable-field Terminal; a user edit of spec.namespace is simply rejected by the CRD CEL rule); (2)
  ConflictException is not a Table terminal code, so a permanent 409 name collision on RenameTable is retried
  with backoff forever (the 'loops on collision' sign), while the transient 409s (concurrent modification,
  stale versionToken) do recover because each retry re-reads versionToken - either surface the namespace drift
  (condition or RenameTable newNamespaceName back to spec) and classify 409 'identical name already exists' as
  terminal by message. code: pkg/resource/table/sdk.go:159-161
  (templates/hooks/table/sdk_read_one_post_set_output.go.tpl:3-5) sets ko.Spec.Namespace = resp.Namespace[0]
  on the latest read by ARN (sdk.go:84-91); pkg/resource/table/ delta.go:146-152 adds Spec.Namespace;
  pkg/resource/table/hooks.go:76-112 customUpdateTable handles only Spec.Name, Spec.MetadataLocation,
  Spec.Tags and returns desired.DeepCopy() so nothing is written back and no error is raised; hooks.go:126-131
  renameTable never sets NewNamespaceName; apis/v1alpha1/table.go:52 / generator.yaml:34-36 Namespace
  is_immutable (CEL self == oldSelf). generator.yaml:103-104 Table terminal_codes = [BadRequestException] ->
  pkg/resource/table/sdk.go:606-611 ConflictException is not terminal; a same-name RenameTable cannot be
  issued because delta.go:139-145 fires only when desired.Name != latest.Name and hooks.go:126-134 sends
  Name=latest, NewName=desired with latest.Status.VersionToken (re-read each reconcile at sdk.go:145-149).
  Delete and metadata-location updates still target the right table because they use the observed spec
  (runtime reconciler.go:1453-1473 passes observed to Delete; hooks.go:159-166). finding: [S3T-TABLE-006](#s3t-table-006)
  (RenameTable newNamespaceName moves the table, ARN unchanged, GetTable(tableArn) reports the new namespace),
  [S3T-TABLE-070](table.md#s3t-table-070), [S3T-TABLE-048](table.md#s3t-table-048) (collision 409 permanent), [S3T-TABLE-050](table.md#s3t-table-050) (same-name 409), [S3T-TABLE-193](#s3t-table-193)
  (concurrent-modification 409 transient), [S3T-TABLE-112](#s3t-table-112) (stale versionToken 409 transient).
  - handling_ref: `generator.yaml:18-23; generator.yaml:119-122; generator.yaml:10-13; generator.yaml:31-39;
    generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; generator.yaml:179-187;
    templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:57-98;
    generator.yaml:137-157`
- [S3T-TABLE-098](#s3t-table-098) - DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation
  (location moved) -> 409; current token -> 204 (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven. The controller passes the
  cached Status.VersionToken to DeleteTable and RenameTable although the token is optional there, so any
  engine commit between the reconcile's GetTable and the write turns a wanted delete/rename into
  ConflictException 409 'Provided version token does not match the table version token.' ([S3T-TABLE-098](#s3t-table-098) for
  DeleteTable, [S3T-TABLE-051](#s3t-table-051) for RenameTable) while a tokenless DeleteTable succeeds ([S3T-TABLE-067](#s3t-table-067)) and a
  token match is not proof the table is unchanged anyway (token = f(location), [S3T-TABLE-134](#s3t-table-134)/302); omit the
  token on DeleteTable/RenameTable, or at minimum map this 409 to requeue-after-re-read, never to Terminal
  (the same code is permanent for duplicate CreateTable, [S3T-TABLE-064](#s3t-table-064)).
  - handling_ref: `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`
- [S3T-TABLE-108](#s3t-table-108) - Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200;
  response versionToken == the CreateTable token (suspected bug)
  Suspected controller bug confirmed by evidence: metadataLocation is engine-managed state and must not be a
  reconciled spec field: on any table in use the observed value is non-nil from the first read
  (schema-at-create writes 00000-<uuid>.metadata.json synchronously, [S3T-TABLE-015](table-schema.md#s3t-table-015)/033) and moves with every
  commit ([S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-123](#s3t-table-123)), so an unset spec diffs forever, while a pinned spec makes the
  controller re-issue UpdateTableMetadataLocation(latest token, old location) after each engine commit - the
  service accepts the rollback with no monotonicity check and even returns the old token ([S3T-TABLE-108](#s3t-table-108),
  [S3T-TABLE-134](#s3t-table-134)) and the engine's next commit then fails with ConflictException 409 'Provided version token
  does not match the table version token.' ([S3T-TABLE-302](#s3t-table-302)). Also drop the create-time Synced=False/requeue
  path: a desired metadataLocation on a fresh table can only fail with BadRequestException 400 -> Terminal,
  because the location must string-prefix the server-assigned warehouseLocation ([S3T-TABLE-105](#s3t-table-105)) that is
  unknowable before CreateTable ([S3T-TABLE-028](#s3t-table-028), [S3T-TABLE-122](#s3t-table-122)). Fix: surface metadataLocation as read-only
  status (or late-initialize and ignore it in compare).
  - handling_ref: `generator.yaml:51-56; pkg/resource/table/sdk.go:489-491`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-001](#s3t-table-001) | Table ARN is <bucketARN>/table/<uuid> (namespace/name not in it); GetTable accepts tableArn alone or bucketARN+namespace+name | get_table_visible_after_create_s=0.03 | 1 |
| [S3T-TABLE-002](#s3t-table-002) | RenameTable(newName) keeps tableARN/ids/locations; GetTable by old ARN -> 200 new name... | rename_latency_ms=[149] | 1 |
| [S3T-TABLE-003](#s3t-table-003) | Replaying an identical RenameTable right after it succeeded -> NotFoundException (HTTP 404) 'The specified table does not exist.' | rename_latency_ms=[149, 13] | 2 |
| [S3T-TABLE-006](#s3t-table-006) | RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId follows the destination, tableARN unchanged | move_latency_ms=[143] | 1 |
| [S3T-TABLE-007](#s3t-table-007) | Across 3 renames (name, namespace, both) GetTable(tableArn) changed only: modifiedAt, modifiedBy, name, namespace, namespaceId | rename_both_latency_ms=[152] | 1 |
| [S3T-TABLE-019](#s3t-table-019) | 20 back-to-back GetTableMetadataLocation reads return 1 distinct body(ies); versionToken/modifiedAt unchanged | get_table_metadata_location_latency_ms_min=6, get_table_metadata_location_latency_ms_max=17 | 20 |
| [S3T-TABLE-034](#s3t-table-034) | versionToken from CreateTable is unchanged across 10 GetTable reads over 54 s with no writes; 20-char hex string, not an integer | token_stability_reads=10, span_s=54, token_len=20 | 10 |
| [S3T-TABLE-064](#s3t-table-064) | Duplicate CreateTable (same bucket/namespace/name) -> ConflictException 409 at t+0/5/10 s and with different metadata/tags; table untouched | retry_spacing_s=[0, 5, 10] | 3 |
| [S3T-TABLE-067](#s3t-table-067) | Immediate CreateTable after DeleteTable (same name) -> 200 in 10/10 (409 in 0); new tableARN every time, old ARN -> 404 in 10/10 | recreate_gap_ms=[361, 326, 308, 338, 336, 403, 316, 326, 329, 368], immediate_200=10, conflict_409=0 | 10 |
| [S3T-TABLE-089](#s3t-table-089) | DeleteTable (no versionToken, empty table) -> 204 empty body in 20/20; GetTable composite and by ARN -> 404 on the next call in 20/20 | after_delete_composite_first_call_ms=[9, 27, 9, 24, 13, 26, 8, 20, 8, 18, 7, 22, 8, 18, 8, 18, 46, 19, 8, 22], after_delete_arn_first_call_ms=[20, 13, 20, 16, 23, 17, 18, 11, 17, 11, 17, 11, 17, 11, 19, 10, 58, 11, 19, 13], delete_latency_ms=[130, 129, 119, 128, 111, 111, 140, 168, 111, 120, 118, 119, 116, 105, 121, 110, 132, 135, 115, 121], still_visible_windows=[] | 20 |
| [S3T-TABLE-099](#s3t-table-099) | DeleteTable of a table with committed metadata + a data object -> HTTP 204 in 3/3; GetTable -> NotFoundException (HTTP 404) on the next call | delete_latency_ms=[140, 192, 118], first_get_after_delete_ms=[9, 11, 9], visible_windows_s=[] | 3 |
| [S3T-TABLE-101](#s3t-table-101) | UpdateTableMetadataLocation response versionToken/metadataLocation == GetTable and GTML at 0/1/5 s; modifiedAt advanced, createdAt unchanged | update_latency_ms=[13] | 1 |
| [S3T-TABLE-102](#s3t-table-102) | After UpdateTableMetadataLocation, 38 alternating GetTable/GTML samples over ~8 s: 0 stale token/location; ListTables modifiedAt == GetTable | samples=38, stale_samples=0, update_latency_ms=[13, 16] | 2 |
| [S3T-TABLE-104](#s3t-table-104) | Re-sending UpdateTableMetadataLocation with the current token and the SAME location (x3) -> 200, token unchanged, but modifiedAt advances | update_latency_ms=[13, 14, 17] | 3 |
| [S3T-TABLE-108](#s3t-table-108) | Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200; response versionToken == the CreateTable token | update_latency_ms_min=11, update_latency_ms_max=17 | 21 |
| [S3T-TABLE-122](#s3t-table-122) | After DeleteTable (204) GetTable and GetTableMetadataLocation are 404 from the first sample (10 ms); re-create gets a new warehouseLocation | delete_latency_ms=[129], poll_samples=6, poll_window_s=1.062, last_200_after_delete_s=0 | 1 |
| [S3T-TABLE-123](#s3t-table-123) | Empty table: first UpdateTableMetadataLocation with the CreateTable token -> 200, GetTable gains metadataLocation; reusing that token -> 409 | update_latency_ms=[12] | 1 |
| [S3T-TABLE-129](#s3t-table-129) | After DeleteTable: TagResource/ListTagsForResource on the old ARN -> 404 from t+0, UntagResource -> 500 (2/2); re-create: new ARN, tags {} | post_delete_list_statuses=[NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException, NotFoundException] | 8 |
| [S3T-TABLE-134](#s3t-table-134) | versionToken is a deterministic function of the full metadataLocation: A->B->A->B gives A0==A1==A2, B1==B2; another table's same key differs | updates=4 | 4 |
| [S3T-TABLE-192](#s3t-table-192) | UpdateTableMetadataLocation same-token race: exactly one 200 + one 409 in 10/10 rounds; also 200 for a missing .metadata.json object | cas_round_wall_s=[0.02, 0.01, 0.01, 0.02, 0.02, 0.03, 0.02, 0.01, 0.01, 0.01] | 10 |
| [S3T-TABLE-193](#s3t-table-193) | 30 back-to-back RenameTable all 204 (121-182 ms) and versionToken NEVER changes; 10 concurrent renames -> 1 x 204, 4 x 409, 1 x 500, 4 x 404 | rename_30_wall_s=4.5, rename_latency_ms_min_max=[121, 182], rename_race_wall_s=3.13 | 40 |
| [S3T-TABLE-315](#s3t-table-315) | CreateTable returns only {tableARN, versionToken}; GetTable adds 12 members (metadataLocation only once the table has metadata) | n_tables=2 | 2 |
| [S3T-TABLE-316](#s3t-table-316) | modifiedBy is absent on a fresh table and still absent after UpdateTableMetadataLocation; RenameTable populates it (== caller) | n_tables=2 | 2 |
| [S3T-TABLE-321](#s3t-table-321) | RenameTable keeps the table versionToken; DeleteTable(current triple) with a bogus token -> ConflictException... | tables=2, renames=4 | 2 |
| [S3T-TABLE-323](#s3t-table-323) | Re-verification [S3T-TABLE-122](#s3t-table-122): DeleteTable 204 and 404s from the first sample hold (n=2); the old metadata.json is NOT readable after delete | delete_latency_ms=[154, 146], first_sample_after_delete_s=[0.009, 0.009], poll_samples=[6, 6], poll_window_s=3.0, last_200_after_delete_s=[null, null], headobject_first_non_200_after_delete_s=[20.068, 20.068, 0.033, 0.031], headobject_last_200_after_delete_s=[10.035, 10.029, -1, -1], headobject_pre_read=[true, true, false, false], headobject_tables_n=5 | 2 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-metadata-location.md` - unverified, n=1 or single-run items worth a future probe; one line each.

- [S3T-TABLE-134](#s3t-table-134) / -302: token = f(full metadataLocation) is inferred from A->B->A->B on 2 tables; tables WITHOUT a
  location still get distinct tokens (3/3 in -302; -034), so what the token encodes before the first commit is unknown.
- [S3T-TABLE-100](#s3t-table-100) / -192 / -105: a dangling location (object absent) and a `<warehouse-host>x` bucket are accepted (n=1
  each); what the engine, maintenance jobs or replication do with such a table afterwards was not probed.
- [S3T-TABLE-108](#s3t-table-108): the rollback to the 00000-* file is accepted (n=21 updates, one table); whether a subsequent engine
  commit or compaction job re-bases or fails was not probed - the "engine's next commit fails 409" consequence is
  reasoning from -302, not an observed engine run.
- [S3T-TABLE-006](#s3t-table-006): move latency n=1; the Spec.Namespace drift and Synced=True described in the note is a code-reading
  conclusion (delta.go/hooks.go), not an in-cluster reproduction.
- [S3T-TABLE-098](#s3t-table-098): the controller-side "cached token -> 409 on delete" is reasoned from the API behavior (n=1 table); not
  reproduced through a reconcile with an intervening external commit.
- [S3T-TABLE-193](#s3t-table-193): 1x500 among 10 concurrent tokenless renames is a single run; reproducibility and whether the 500 ever
  leaves the table renamed are open.
- [S3T-TABLE-121](#s3t-table-121): one 404 message for missing table/namespace/bucket, single run, vs FU2 "GetTable 404 message depends on
  name shape" (STATUS.md Phase 8, `table.md`); check which name shapes GetTableMetadataLocation and
  UpdateTableMetadataLocation follow.
- [S3T-TABLE-245](#s3t-table-245): ghost-name/all-zero-UUID region mismatch precedes lookup - single run, two region pairs.
- [S3T-TABLE-123](#s3t-table-123): first commit on an empty table consuming the create token is n=1.
- [S3T-TABLE-104](#s3t-table-104): three same-location re-sends advanced modifiedAt without changing the token (n=3); whether any metadata
  object is rewritten was not checked (-108 notes the probe wrote no objects).
- [S3T-TABLE-064](#s3t-table-064): the message-based split of the two ConflictException 409s is the only discriminator observed; no probe
  saw the 'identical name' 409 turn transient ([S3T-TABLE-067](#s3t-table-067) saw 0 x 409 in 10/10 immediate same-name re-creates after a
  delete), so the controller's "permanent by message" rule rests on 3 spaced retries over 10 s.
- [S3T-TABLE-323](#s3t-table-323): the HeadObject 403-vs-200 split after DeleteTable (pre-read keeps 200 for 10-20 s) is n=2 per variant and
  reads like a cached authorization; whether the object itself survives is unobservable from the owning account.
- [S3T-TABLE-129](#s3t-table-129): UntagResource -> 500 on a deleted table's ARN is 2/2 from one probe; whether it ever turns 404 is open.
- [S3T-TABLE-316](#s3t-table-316): modifiedBy absent after an UpdateTableMetadataLocation commit is n=2 tables at +3 s; a later appearance
  was not polled.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-019"></a>**S3T-TABLE-019** | idempotency | low | confirmed | 20 back-to-back GetTableMetadataLocation reads return 1 distinct body(ies); versionToken/modifiedAt unchanged | [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-034](#s3t-table-034), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-311](table-schema.md#s3t-table-311), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-104](#s3t-table-104), [S3T-TABLE-124](#s3t-table-124), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-193](#s3t-table-193) | - |
| <a id="s3t-table-112"></a>**S3T-TABLE-112** | idempotency | medium | confirmed | RenameTable optional versionToken: stale (create-time) token -> ConflictException (HTTP 409)... | [S3T-TABLE-005](#s3t-table-005), [S3T-TABLE-007](#s3t-table-007), [S3T-TABLE-193](#s3t-table-193), [S3T-TABLE-316](#s3t-table-316), [S3T-TABLE-320](#s3t-table-320), [S3T-TABLE-009](#s3t-table-009) | [S3T-TABLE-051](#s3t-table-051) |
| <a id="s3t-table-135"></a>**S3T-TABLE-135** | request-validation | low | confirmed | metadataLocation check = case-sensitive substring '.metadata.json' (leading dot required): x.metadata.jsonfoo, .json.tar, trailing '/' ok | [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-106](#s3t-table-106), [S3T-TABLE-107](#s3t-table-107) | - |
| <a id="s3t-table-269"></a>**S3T-TABLE-269** | other | low | confirmed | Doc claim C077 TRUE: GetTable(tableArn) alone returns the same table as GetTable(tableBucketARN+namespace+name) | [S3T-TABLE-023](#s3t-table-023), [S3T-TABLE-001](#s3t-table-001), [S3T-TABLE-270](table.md#s3t-table-270), [S3T-TABLE-024](table.md#s3t-table-024), [S3T-TABLE-025](table.md#s3t-table-025) | - |
| <a id="s3t-table-275"></a>**S3T-TABLE-275** | other | low | confirmed | Doc claim C083 TRUE: RenameTable with a non-matching versionToken -> ConflictException, not renamed; without versionToken it succeeds | [S3T-TABLE-051](#s3t-table-051), [S3T-TABLE-112](#s3t-table-112), [S3T-TABLE-002](#s3t-table-002), [S3T-TABLE-006](#s3t-table-006), [S3T-TABLE-271](table.md#s3t-table-271), [S3T-TABLE-272](table.md#s3t-table-272), [S3T-TABLE-273](table.md#s3t-table-273), [S3T-TABLE-274](table.md#s3t-table-274), [S3T-TABLE-276](table.md#s3t-table-276), [S3T-TABLE-050](table.md#s3t-table-050), [S3T-TABLE-292](#s3t-table-292), [S3T-TABLE-277](#s3t-table-277), [S3T-TABLE-282](#s3t-table-282), [S3T-TABLE-291](table.md#s3t-table-291), [S3T-TABLE-123](#s3t-table-123) | - |
| <a id="s3t-table-277"></a>**S3T-TABLE-277** | other | low | confirmed | Doc claim C085 TRUE: UpdateTableMetadataLocation(current token, <warehouse>/metadata/00001-x.metadata.json) -> 200, new token; reads agree | [S3T-TABLE-101](#s3t-table-101), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-292](#s3t-table-292), [S3T-TABLE-271](table.md#s3t-table-271), [S3T-TABLE-275](#s3t-table-275), [S3T-TABLE-282](#s3t-table-282), [S3T-TABLE-291](table.md#s3t-table-291) | - |
| <a id="s3t-table-278"></a>**S3T-TABLE-278** | other | low | confirmed | Doc claim C086 PARTLY: metadataLocation outside the warehouse (other table/bucket, s3a, https) -> 400; literal prefix: <host>x/.. passes | [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-106](#s3t-table-106), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-107](#s3t-table-107), [S3T-TABLE-135](#s3t-table-135) | - |
| <a id="s3t-table-280"></a>**S3T-TABLE-280** | other | low | confirmed | Doc claim C088 TRUE: metadataLocation under the warehouse ending in .metadata.json.gz is accepted | [S3T-TABLE-107](#s3t-table-107), [S3T-TABLE-135](#s3t-table-135), [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-281](#s3t-table-281), [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-106](#s3t-table-106) | - |
| <a id="s3t-table-281"></a>**S3T-TABLE-281** | other | low | confirmed | Doc claim C089 TRUE: UpdateTableMetadataLocation does not verify that an object exists at metadataLocation | [S3T-TABLE-100](#s3t-table-100), [S3T-TABLE-192](#s3t-table-192), [S3T-TABLE-278](#s3t-table-278), [S3T-TABLE-279](#s3t-table-279), [S3T-TABLE-280](#s3t-table-280), [S3T-TABLE-105](#s3t-table-105), [S3T-TABLE-106](#s3t-table-106), [S3T-TABLE-107](#s3t-table-107), [S3T-TABLE-135](#s3t-table-135) | - |
| <a id="s3t-table-282"></a>**S3T-TABLE-282** | other | low | confirmed | Doc claim C090 TRUE: UpdateTableMetadataLocation with a stale/garbage versionToken -> ConflictException 409, unchanged; empty token -> 400 | [S3T-TABLE-103](#s3t-table-103), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-292](#s3t-table-292), [S3T-TABLE-277](#s3t-table-277), [S3T-TABLE-271](table.md#s3t-table-271), [S3T-TABLE-275](#s3t-table-275), [S3T-TABLE-291](table.md#s3t-table-291) | - |
| <a id="s3t-table-292"></a>**S3T-TABLE-292** | other | low | confirmed | Doc claim C116 TRUE: CreateTable returns tableARN+versionToken; an immediate GetTable returns the same tableARN and versionToken | [S3T-TABLE-088](table.md#s3t-table-088), [S3T-TABLE-022](table.md#s3t-table-022), [S3T-TABLE-034](#s3t-table-034), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-277](#s3t-table-277), [S3T-TABLE-271](table.md#s3t-table-271), [S3T-TABLE-275](#s3t-table-275), [S3T-TABLE-282](#s3t-table-282), [S3T-TABLE-291](table.md#s3t-table-291), [S3T-TABLE-123](#s3t-table-123), [S3T-TABLE-283](table-schema.md#s3t-table-283), [S3T-TABLE-066](#s3t-table-066) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
