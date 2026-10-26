<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table (lifecycle, identity, rename, errors, tags)
_CreateTable/GetTable/DeleteTable/RenameTable/ListTables: minimal Create echo, Get-only members, lookup by ARN or triple, rename across namespaces, 404 message shapes, tags (ValidationException on length, UntagResource 500 on vanished tables), concurrency; catch-all for Table findings not covered above._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Table (lifecycle, identity, rename, errors, tags) - hand-authored overview seed for `table.md` (s3tables). Rules and
timings cite canonical finding ids; generated sections below this block are regenerated on every `semantics render`.

### Orientation

This document covers the Table core surface: CreateTable, GetTable, DeleteTable, RenameTable, ListTables and
TagResource/UntagResource/ListTagsForResource on a table ARN, plus name/format validation, 404 message shapes,
cross-region calls and concurrency; metadata/versionToken, encryption/storage class, policy, maintenance, replication
and record expiration live in the sibling documents under "Where next". There is no state machine and no status field:
CreateTable is synchronous (GetTable by triple and by tableArn -> 200 on the first call, 20/20 each, [S3T-TABLE-088](#s3t-table-088)),
DeleteTable is synchronous (ListTables drops the table on the first poll 10/10, -096; GetTable by name and by ARN ->
404 afterwards and a second DeleteTable -> 404, -291), and RenameTable flips old-triple 404 / new-triple 200 with no
both-200 overlap and no both-404 gap at 150 ms polling (-068). Identity is the tableARN `<tableBucketARN>/table/<uuid>`:
it carries neither namespace nor table name, GetTable returns the identical string, and it survives RenameTable(newName)
and a cross-namespace move (-022, -271, -272); GetTable accepts that ARN alone OR the full triple, never both (-025).
CreateTable's response is minimal (-022); the rest is Get-only - GetTable's keys are createdAt, createdBy, format,
modifiedAt, name, namespace, namespaceId, tableBucketId, ownerAccountId, tableARN, type, versionToken, warehouseLocation
and no tags (-113); namespaceId/tableBucketId equal GetNamespace.namespaceId / GetTableBucket.tableBucketId (-029);
modifiedAt == createdAt and modifiedBy is absent right after create (-032). Tags live only behind ListTagsForResource,
which is read-after-write (-113, -128); writes are per-ARN upserts/removals under a cumulative, atomic 50-tag limit
(-114, -125, -185); the one hard trap is UntagResource on a table that no longer exists: InternalServerErrorException
500 instead of 404 for as long as was measured (-324). The controller keys GetTable off the ARN once known, renames by
newName only (Namespace is immutable) and reads tags inside sdkFind via ListTagsForResource.

### Rules

1. UntagResource on a table ARN whose table does not exist (just deleted, never created, or gone with its bucket) ->
   InternalServerErrorException 500 'An internal error occurred. Try again.' in 23/23 calls - at t+0, t+5 s, +36 s,
   +181 s and 8.3-11.7 min after the delete, for present and never-present keys alike - while TagResource and
   ListTagsForResource on the same ARN -> NotFoundException 404 in 35/35; retrying that 500 as a server error loops
   forever: treat an UntagResource 500 on a table GetTable cannot find as not-found ([S3T-TABLE-324](#s3t-table-324), -130, -143, -145).
2. Serialize tag writes per table: 10 concurrent TagResource (distinct keys) on one table -> 8-9/10
   InternalServerErrorException 500 (never 409) in 3/3 rounds, the losing keys not applied and no successful key lost;
   10 concurrent UntagResource -> 9/10 500 with the 9 keys remaining; 30 sequential UntagResource with no pause -> 30 x
   204 and no 429 ([S3T-TABLE-186](#s3t-table-186)).
3. Tag length violations are permanent errors with an unmodeled code: TagResource with a 129-char key, a 257-char
   value or an empty key -> HTTP 400 wire code ValidationException (not in the s3tables model, so the SDK surfaces an
   untyped error); UntagResource with one 129-char key anywhere in tagKeys -> ValidationException 400 and nothing
   removed; charset violations ('|', '#', emoji key, '☃' value) and 'aws:'-prefixed keys of any case are
   BadRequestException 400 - a terminal match on BadRequestException alone misses the length cases ([S3T-TABLE-120](#s3t-table-120),
   -119, -126, -118).
4. The table tag limit is 50 user tags cumulative, enforced atomically: 51 cumulative -> BadRequestException 400 'The
   provided tags are invalid. Number of user tags must not exceed 50' with nothing applied; at 50, updating an existing
   key -> 200 but one call mixing an update with a new key -> 400 and nothing applied; CreateTable with 51 tags -> 400
   and no table exists afterwards ([S3T-TABLE-185](#s3t-table-185), -263).
5. RenameTable is never a no-op: newName = the current name, newNamespaceName = the current namespace, or both ->
   ConflictException 409 'A table with an identical name already exists in the namespace.' (modifiedAt unchanged) - the
   identical 409 also answers a real collision in the target namespace (source table unchanged); neither field ->
   BadRequestException 400 'Neither a new namespace name nor a new table name is specified.'; the same name in another
   namespace renames fine, names are scoped per namespace ([S3T-TABLE-050](#s3t-table-050), -048, -274, -276).
6. Key the resource on the tableARN `<tableBucketARN>/table/<uuid>`: it carries neither namespace nor table name, is
   unchanged by RenameTable(newName) and by a move to another namespace (tableARN, tableBucketId, namespaceId kept),
   tags are keyed by it and survive the move (TagResource right after -> 200), and GetTable(tableArn) stayed 200 across
   5 renames while the old triple went 404 ([S3T-TABLE-022](#s3t-table-022), -271, -272, -008, -068).
7. GetTable takes the ARN alone OR the full triple: tableArn together with tableBucketARN+namespace+name ->
   BadRequestException 400 'Must provide either table arn alone OR all of (table bucket arn, namespace, table name)
   together' even when both name the SAME table; no identifier or a partial triple -> the same 400 ([S3T-TABLE-025](#s3t-table-025),
   -024, -270).
8. Parents are prerequisites, never auto-created: CreateTable into a nonexistent namespace -> NotFoundException 404
   'The specified namespace does not exist.' and ListNamespaces does not show it afterwards; a nonexistent bucket ARN
   -> 404 'The specified bucket does not exist.'; a bucket-less, s3: or garbage ARN -> BadRequestException 400 'The
   specified table bucket ARN is not valid.'; RenameTable(newNamespaceName=never created or just deleted) -> 404 'The
   specified destination namespace does not exist.' and no namespace is created ([S3T-TABLE-080](#s3t-table-080), -081, -049, -273).
9. Do not infer which parent is missing from a GetTable 404: a never-created bucket, namespace or table all return
   'The specified table does not exist.' (3 trials per cell), while an upper-cased or trailing-space namespace returns
   'The specified namespace does not exist.'; only DeleteTable names the missing parent (3 distinct messages);
   RenameTable on a just-deleted table -> 404 'table does not exist' at t+0 and t+2 s but 404 'destination namespace
   does not exist' once the bucket is deleted too ([S3T-TABLE-326](#s3t-table-326), -082, -027, -053).
10. Names and identifiers are case-sensitive and never normalized: GetTable/DeleteTable with the existing name
    upper-cased -> 404 and the table is untouched; an upper-cased uuid/prefix/segment or stray whitespace in
    tableArn/resourceArn -> 400 'not valid'; format must be exactly 'ICEBERG' ('iceberg', 'Iceberg', DELTA, HUDI,
    PARQUET, '' or omitted -> 400 'The specified open table format is not valid'); a table name of 255 chars, 'a__b',
    trailing '_', leading digit or 'aws_*' is accepted while 256 chars, leading '_', '__', upper case, '-', '.', space
    or non-ASCII -> 400 'The specified table name is not valid.' for CreateTable and RenameTable.newName alike; a raw
    CreateTable with an extra 'type':'aws' member -> 400 'Creating tables of this type are not allowed.' ([S3T-TABLE-027](#s3t-table-027),
    -267, -047, -079, -078, -052, -035).
11. CreateTable, RenameTable, DeleteTable, TagResource and UntagResource called with the other region's ARNs ->
    BadRequestException 400 'The provided Region does not match the Region of the resource.' in both directions, and
    nothing changes in either region ([S3T-TABLE-244](#s3t-table-244), -143).
12. Tag semantics: TagResource is an upsert (untouched keys kept; identical re-send -> 200 with the map unchanged;
    TagResource({}) -> 200 no-op); UntagResource of an absent key -> 204, a mixed list removes the present keys,
    tagKeys=[] -> BadRequestException 400; keys are case-sensitive ('Env' and 'env' coexist; untag of a non-matching
    case -> 204, removes nothing) and stored verbatim (' k ' distinct from 'k', '' value kept); after removing every tag
    the body is {"tags":{}} like a never-tagged table; bucket and table stores are separate and UntagResource on the
    bucket ARN acts on the bucket's tags ([S3T-TABLE-114](#s3t-table-114), -125, -116, -117, -127, -115, -144, -264).
13. ListTables: namespace = never existed | prefix of an existing one | just deleted | upper-cased -> NotFoundException
    404 'The specified namespace does not exist.' while an existing empty namespace -> 200 with 0 tables; a deleted or
    never-created bucket -> 404, an empty bucket -> 200 with tables=[] and no token; prefix is a literal case-sensitive
    table-name prefix across namespaces, ANDed with namespace ('*' matches none, '' is unfiltered, 300 chars -> 400);
    maxTables=0 or -1 -> 400, omitted -> all 120 of 120 tables with NO continuationToken, 100 -> 100 + token; order is
    table name ascending across namespaces with an opaque tie-break; a token survives DeleteTable and is multi-use
    (maxTables may change) but a tampered, foreign-bucket or filter-changed token -> 400 'The continuation token is no
    longer valid.' ([S3T-TABLE-094](#s3t-table-094), -097, -093, -168, -217, -091, -305, -095, -293).
14. Pace deletes: 120 DeleteTable back to back -> 68 x 204 then 52 x TooManyRequestsException 429 'Too Many Requests'
    (n=120), while 120 sequential CreateTable at ~3/s and 20 DeleteTable at 4 TPS saw no 429; 20 concurrent
    CreateTable (distinct names, one namespace) all returned 200 but were serialized server-side (latency 0.3-9.7 s)
    ([S3T-TABLE-218](#s3t-table-218), -188).
15. No settling wait is needed anywhere measured: GetTable by triple and by ARN -> 200 on the first call after
    CreateTable (20/20 each); ListTables shows a new table and drops a deleted one on the first poll (10/10 each);
    ListTagsForResource reflects Tag/UntagResource on the first read (20/20) and CreateTable's tags on the very next
    call; ListTables reflects a rename on poll #0 (3/3) with modifiedAt equal to GetTable's; RenameTable followed at
    once by the reverse rename -> 204/204 (3/3); a combined move+rename never exposes an intermediate triple at 200 ms
    polling (2/2) ([S3T-TABLE-088](#s3t-table-088), -096, -128, -113, -071, -069, -070, -068).

### Timing ranges

- CreateTable (format=ICEBERG, no metadata) call latency: 293-481 ms (n=20, [S3T-TABLE-088](#s3t-table-088)); 327-569 ms (n=10,
  -096); 278-395 ms across a 120-call sequential run at ~3.0/s with 0 x 429 (n=120, -218, -217); 331-9679 ms when 20
  calls are fired concurrently into one namespace, i.e. serialized server-side (n=20, -188). The first GetTable after
  the Create response, issued 27-62 ms (triple) / 34-86 ms (tableArn) later, returned 200 in 20/20 each (-088).
- DeleteTable latency: 127-159 ms (n=10, [S3T-TABLE-096](#s3t-table-096)); 140-195 ms at 4 TPS (n=20, -188); 2-203 ms across a 120-call
  burst at ~11.5/s whose calls 69-120 were 429s (n=120, -218); the 429 refill time was not measured (-188).
- RenameTable latency: 146-178 ms (n=5, [S3T-TABLE-068](#s3t-table-068)); 133-167 ms over forward+reverse pairs (n=3 pairs, -069);
  131-200 ms for a combined move+rename (n=2, -070); a ConflictException collision answers in 11-55 ms (n=2, -048).
- Rename visibility: the first new-triple 200 and the first old-triple 404 both completed 0.04-0.07 s BEFORE the
  RenameTable call returned (150 ms polling, both-200 overlap 0 s, both-404 gap 0 s, n=5, [S3T-TABLE-068](#s3t-table-068)); for a
  combined move+rename the first new/new 200 landed 0.006-0.075 s and the first old/old 404 0.004-0.072 s after the
  return (200 ms polling, n=2, -070); ListTables was consistent on the first poll, issued 0.02-0.03 s after the
  return (n=3, -071).
- Tag writes: TagResource on a table ARN 97-128 ms vs 44-60 ms on a bucket ARN (n=10 each, [S3T-TABLE-128](#s3t-table-128)); sequential
  UntagResource 88-133 ms (n=30, -186); a 10-call concurrent TagResource burst takes 3.15-6.18 s wall with 8-9 of 10
  failing 500 (n=3 rounds, -186).
- UntagResource 500 on a nonexistent table: 314-388 ms per call (n=23), reproduced at t+0 and t+5 s (12/12), at +36 s
  and +181 s after the delete, and on ARNs deleted 8.3-11.7 min earlier whose bucket was gone too ([S3T-TABLE-324](#s3t-table-324)).

### Handling gaps

- [S3T-TABLE-120](#s3t-table-120) (suspect-bug, high) - today: `terminal_codes: [BadRequestException]` (generator.yaml:103-104) and
  `terminalAWSError` matches only that typed code (pkg/resource/table/sdk.go:606-609); `syncTags` sends the whole
  "added" map in one TagResource (pkg/resource/table/hooks.go:216-232). Evidence: TagResource with a 129-char key, a
  257-char value or an empty key -> HTTP 400 wire code ValidationException, which the s3tables model does not declare
  for TagResource, while every other tag rejection (charset, aws: prefix) is BadRequestException 400; botocore stops
  only the empty key client-side, so the over-long values reach the server. A Table whose spec.tags break the length
  bounds is therefore requeued as recoverable instead of reaching Terminal. Not yet handled.
- [S3T-TABLE-119](#s3t-table-119) (suspect-bug, low) - same mechanism as -120 seen per cell: key_128 and value_256 -> 200 and read back
  verbatim; key_129 / value_257 -> ValidationException 400 'Member must have length less than or equal to 128/256';
  '|', '#' or emoji in a key and '☃' in a value -> BadRequestException 400 'Tag key/value can only contain
  alphanumeric characters, spaces, and _.:/=+@-'. Today the charset cases go Terminal and the length cases requeue
  forever; not yet handled (neither `ValidationException` in terminal_codes nor length bounds in the CRD).
- [S3T-TABLE-113](#s3t-table-113) (partial, high) - today: GetTable has no tags member, so `sdkFind` calls ListTagsForResource through
  `setResourceTags` (pkg/resource/table/sdk.go:175-179, hooks.go:39-67) and maps an empty tag map to `Spec.Tags = nil`
  (hooks.go:60-64); `newResourceDelta` compares tags with `MapStringStringEqual` (delta.go:177-181). Evidence:
  CreateTable tags are readable on the very first ListTagsForResource (4/4 keys), GetTable's key list has no 'tags', and
  an untagged table (or one whose last tag was removed) returns {"tags":{}} with the key present (-113, -127;
  read-after-write 20/20 in -128). The read path is in place; the scorer rated it a partial match against [GT-S3T-030](service.md#gt-s3t-030) (controller hooks catalog entry) and
  [GT-S3T-043](service.md#gt-s3t-043) (controller hooks catalog entry), and the nil-vs-`{}` corner (`compare.nil_equals_zero_value`) has no finding of its own - partially
  handled, not yet verified end to end.

### Where next

- The six Table-family siblings carry everything this catch-all leaves out: `table-metadata-schema.md` (CreateTable
  metadata.iceberg, UpdateTableMetadataLocation, versionToken and ConflictException CAS semantics - the token the
  controller passes on RenameTable and DeleteTable is explained there), `table-encryption-storage.md` (create-only
  encryptionConfiguration/storageClassConfiguration, KMS effects on CreateTable), `table-policy.md` (Put/Get/Delete
  TablePolicy, per-operation IAM actions, AccessDeniedException vs ForbiddenException), `table-maintenance.md`
  (PutTableMaintenanceConfiguration, job status, KMS maintenance principal), `table-replication.md` (Put/Get/Delete
  TableReplication, ~60 s status evaluation, orphan replicas) and `table-record-expiration.md` (MethodNotAllowedException
  405 on every creatable table; scope verdict).
- Parents and cross-cutting: `namespace.md` (the namespace CreateTable requires and ListTables 404s on; not-empty
  DeleteNamespace as 400; the ~16 req/s per account-region limiter), `tablebucket.md` (bucket ARN construction, the
  ~10 min post-delete name reservation, bucket-ARN tags - the other half of [S3T-TABLE-115](#s3t-table-115)/-144) and `service.md`
  (service-wide findings, the ValidationException-vs-BadRequestException split across operations, the handling-gaps
  summary that -120/-119 feed into).
- Long catalogues render in `details/<id>.md` rather than inline: the CreateTable name cells (-078) and
  RenameTable.newName cells (-052), the tag-ARN form matrix (-143) and identifier normalization (-267), the bucket-ARN
  variants on CreateTable (-081), the 9-cell GetTable 404 message table (-326), the tag bound cells (-119) and
  verbatim-storage cases (-117), the ListTables prefix (-093) and continuation-token (-095) cells, and the three-run
  UntagResource-500 evidence (-324).
- Open experiments (STATUS.md, Phase 8): the GetTable 404 message dependence on name shape was settled by
  F-followups-2 (-326); nothing is running and no blocker is open. Still thin here: the 429 refill time after a
  DeleteTable burst (-188 `throttle_refill_s` null, -218), move+rename atomicity at 200 ms with n=2 (-070), the
  collision latency with n=2 (-048) and the controller's nil-vs-`{}` tag mapping (-113) - see `table.open-questions.md`.
  Account state: the undeletable bucket `ackq-9ba36e-polnorm` (us-west-2) holds table `ackq_9ba36e_t` behind a
  root-principal Deny policy; its AccessDeniedException belongs to `table-policy.md`, not to this document.
<!-- preserved:end -->

## At a glance

- canonical findings: 69 (high 20 / medium 31 / low 18); duplicates folded into the appendix: 1
- handling: handled 32 · partial 2 · tracked 0 · unhandled 33 · suspect-bug 2 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: request-validation 13, error-code 10, other 10, eventual-consistency 5, identity 5,
  response-fidelity 5, tag-semantics 5, quota-limit 3, async-state-machine 2, delete-semantics 2, idempotency
  2, prerequisite 2, read-gap 2, cross-region 1, normalization 1, server-default 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateNamespace | create | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetNamespace | read | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

- <a id="s3t-table-069"></a>**S3T-TABLE-069** `async-state-machine` · impact medium · handled · verified 2026-10-13
  **RenameTable immediately followed by the reverse RenameTable: 3/3 second calls -> 204 (no settling wait needed)**
  RenameTable(a->b) then RenameTable(b->a) issued back to back without any wait, 3 trials: first call [204,
  204, 204], second call [204, 204, 204]; GetTable(tableArn) then shows the original name in 3/3 trials.
  - ACK: none · ops: RenameTable, GetTable · fields: newName
  - repro: RenameTable(ns,a,newName=b); RenameTable(ns,b,newName=a) immediately; GetTable(tableArn)
  - measurements: first_latency_ms=[142, 133, 144], second_latency_ms=[134, 167, 140], n=3
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122`
  - related: [S3T-TABLE-068](#s3t-table-068), [S3T-TABLE-070](#s3t-table-070), [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006) · hypotheses: H-T-198 · evidence:
    table/consistency-windows/rename-visibility
  - notes: H-T-198 second clause (a second RenameTable issued immediately after the first returns 204)
    confirmed.

- <a id="s3t-table-070"></a>**S3T-TABLE-070** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Combined RenameTable(newNamespaceName+newName): intermediate triples never resolve at 200 ms polling (2 trials)**
  Four GetTable pollers at 200 ms on (old ns, old name), (new ns, old name), (old ns, new name), (new ns, new
  name) around a combined move+rename, 2 trials: RenameTable -> [204, 204]; triples that ever returned 200:
  ['new_ns/new_name', 'old_ns/old_name']; statuses per triple: {'old_ns/old_name': ['200',
  'NotFoundException'], 'new_ns/old_name': ['NotFoundException'], 'old_ns/new_name': ['NotFoundException'],
  'new_ns/new_name': ['200', 'NotFoundException']}. First (new ns, new name) 200 relative to the call
  returning: [0.006, 0.075] s; first (old ns, old name) 404: [0.004, 0.072] s.
  - ACK: custom_update, requeue · ops: RenameTable, GetTable · fields: newNamespaceName, newName
  - repro: 4 GetTable pollers at 200 ms; RenameTable(newNamespaceName=ns2, newName=u); stop 1 s later; 2
    trials
  - measurements: rename_both_latency_ms=[200, 131], poll_interval_s=0.2, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-068](#s3t-table-068), [S3T-TABLE-069](#s3t-table-069), [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006) · hypotheses: H-T-060 · evidence:
    table/consistency-windows/rename-visibility
  - notes: H-T-060 second half (both fields move and rename atomically, no intermediate state) confirmed at
    200 ms resolution.

## Field matrix

C = accepted by the create input (CreateTable), U = by the update input, R = present in the read output.

| leaf | C | U | R | type |
| --- | --- | --- | --- | --- |
| createdAt | - | - | x | timestamp |
| createdBy | - | - | x | string |
| direction | x | - | - | enum:IcebergSortDirection |
| doc | x | - | - | string |
| encryptionConfiguration | x | - | - | struct:EncryptionConfiguration |
| fieldId | x | - | - | integer |
| fields | x | - | - | list<struct:SchemaField> |
| format | x | - | x | enum:OpenTableFormat |
| iceberg | x | - | - | struct:IcebergMetadata |
| id | x | - | - | integer |
| identifierFieldIds | x | - | - | list<integer> |
| kmsKeyArn | x | - | - | string |
| managedByService | - | - | x | string |
| managedTableInformation | - | - | x | struct:ManagedTableInformation |
| metadata | x | - | - | struct:TableMetadata |
| metadataLocation | - | - | x | string |
| modifiedAt | - | - | x | timestamp |
| modifiedBy | - | - | x | string |
| name | x | x | x | string |
| namespace | x | x | x | string |
| namespaceId | - | - | x | string |
| newName | - | x | - | string |
| newNamespaceName | - | x | - | string |
| nullOrder | x | - | - | enum:IcebergNullOrder |
| orderId | x | - | - | integer |
| ownerAccountId | - | - | x | string |
| partitionSpec | x | - | - | struct:IcebergPartitionSpec |
| properties | x | - | - | map<string,string> |
| replicationInformation | - | - | x | struct:ReplicationInformation |
| required | x | - | - | boolean |
| schema | x | - | - | struct:IcebergSchema |
| schemaId | x | - | - | integer |
| schemaV2 | x | - | - | struct:IcebergSchemaV2 |
| sourceId | x | - | - | integer |
| sourceTableARN | - | - | x | string |
| specId | x | - | - | integer |
| sseAlgorithm | x | - | - | enum:SSEAlgorithm |
| storageClass | x | - | - | enum:StorageClass |
| storageClassConfiguration | x | - | - | struct:StorageClassConfiguration |
| tableARN | - | - | x | string |
| tableBucketARN | x | x | - | string |
| tableBucketId | - | - | x | string |
| tags | x | - | - | map<string,string> |
| transform | x | - | - | string |
| type | x | - | x | string |
| versionToken | - | x | x | string |
| warehouseLocation | - | - | x | string |
| writeOrder | x | - | - | struct:IcebergSortOrder |

## Identity and lookup

- <a id="s3t-table-022"></a>**S3T-TABLE-022** `identity` · impact high · handled · verified 2026-10-13
  **CreateTable tableARN is <tableBucketARN>/table/<uuid>: namespace and table name absent; GetTable returns the identical string**
  CreateTable(format=ICEBERG, no metadata) returned tableARN
  'arn:aws:s3tables:<region>:<account>:bucket/<bucket>/table/<uuid>' (resource part has 4 segments; last
  segment is a UUID; table name in ARN: False; namespace in ARN: False; starts with the bucket ARN +
  '/table/': True). GetTable(tableBucketARN+namespace+name) -> HTTP 200 with tableARN equal to the create
  response: True.
  - ACK: is_arn_primary_key, is_read_only · ops: CreateTable, GetTable · fields: tableARN
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable -> parse tableARN -> GetTable by composite key
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122; templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; test/e2e/tests/test_table.py:175-193; test/e2e/resources/table.yaml; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-029](#s3t-table-029), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006), [S3T-TABLE-008](#s3t-table-008) · hypotheses: H-T-001 · evidence:
    table/identity/get-table-identifier-forms
  - notes: H-T-001 confirmed: the ARN cannot be derived from spec fields (bucket ARN + service-generated table
    id); it must be captured from the create response.

- <a id="s3t-table-144"></a>**S3T-TABLE-144** `identity` · impact medium · handled · verified 2026-10-13
  **UntagResource(table-bucket ARN, ['t','b']) -> HTTP 204 and acts on the BUCKET's tags (b removed: True); the table's 't' stays**
  Bucket tagged {b:1}, its table tagged {t:1}. On the bucket ARN: ListTagsForResource -> HTTP 200 ({'b':
  '1'}); TagResource({probe}) -> HTTP 200; UntagResource(['t','b']) -> HTTP 204. Afterwards bucket tags
  {'probe': 'bucket_arn'}, table tags {'t': '1'}.
  - ACK: is_arn_primary_key, tags.custom-sync · ops: UntagResource, TagResource, ListTagsForResource · fields:
    resourceArn, tagKeys
  - repro: CreateTableBucket(tags b=1) -> CreateTable(tags t=1) -> UntagResource(bucketARN, ['t','b']) ->
    ListTagsForResource on both ARNs
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-115](#s3t-table-115), [S3T-TABLE-008](#s3t-table-008) · hypotheses: H-T-207 · evidence: table/identity/tag-arn-forms
  - notes: H-T-207 bucket-ARN clause confirmed (predicted 204 on the bucket's own tag set, table untouched).

- <a id="s3t-table-145"></a>**S3T-TABLE-145** `identity` · impact medium · handled · verified 2026-10-13
  **Table ARN with the bucket segment swapped to another existing bucket -> list/tag NotFoundException 404, untag 500; UUID alone resolves none** (hypothesis refuted; behavior confirmed)
  ARN = real table ARN with 'bucket/<B1>/' replaced by 'bucket/<B2>/' (B2 exists, empty): ListTagsForResource
  -> NotFoundException (HTTP 404) 'The specified table does not exist.'; TagResource({probe}) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; UntagResource(['t']) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'. Real table tags afterwards
  {'t': '1'}; B2 tags {'b2': '1'}. The service does NOT resolve a table by its UUID alone.
  - ACK: is_arn_primary_key, exceptions.404 · ops: ListTagsForResource, TagResource, UntagResource · fields:
    resourceArn
  - repro: Two buckets; table in B1 tagged t=1; build the table ARN with B2's name in the bucket segment;
    ListTagsForResource/TagResource/UntagResource on it; ListTagsForResource(real ARN)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-143](#s3t-table-143) · hypotheses: H-T-208 · evidence: table/identity/tag-arn-forms
  - notes: H-T-208 (contrarian: UUID alone resolves the table across bucket segments) refuted: the mismatched
    ARN is rejected and the real table's tags are intact. Message: 'An internal error occurred. Try again.'.
  - full notes: [details/S3T-TABLE-145.md](details/S3T-TABLE-145.md)

## Idempotency

- <a id="s3t-table-095"></a>**S3T-TABLE-095** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables continuationToken survives DeleteTable, is multi-use (maxTables may change); tampered, foreign-bucket or filter-changed -> 400**
  ListTables(maxTables=2) -> page ['ns2/a', 'ns1/a'] + token (720 chars). DeleteTable(ns1/abc) (3rd in order)
  -> HTTP 204. Token reused -> ['ns1/abd', 'ns2/abz']; reused again -> ['ns1/abd', 'ns2/abz']; reused with
  maxTables=5 -> ['ns1/abd', 'ns2/abz', 'ns2/b', 'ns1/xyz', 'ns1/z']; reused with an added namespace filter ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'. Token+'x' ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'; 'not-a-token' ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'; token from another bucket ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'.
  - ACK: custom_find, requeue · ops: ListTables, DeleteTable · fields: continuationToken, maxTables
  - repro: ListTables(maxTables=2) -> token; DeleteTable(3rd table); ListTables(token) x2;
    ListTables(token+'x'); ListTables(token from bucket2)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-091](#s3t-table-091), [S3T-TABLE-305](#s3t-table-305), [S3T-TABLE-168](#s3t-table-168), [S3T-TABLE-217](#s3t-table-217) · hypotheses: H-T-074 · evidence:
    table/weird-inputs/list-tables-bounds
  - notes: H-T-074 confirmed: the token stayed valid after a DeleteTable of the next table in order (deleted
    table absent, no duplicates with page 1), was reused twice with identical results and with a different
    maxTables; token+'x', 'not-a-token' and a token from another bucket all -> BadRequestException 400...
  - full notes: [details/S3T-TABLE-095.md](details/S3T-TABLE-095.md)

- <a id="s3t-table-125"></a>**S3T-TABLE-125** `idempotency` · impact high · handled · verified 2026-10-13
  **UntagResource: absent key -> 204, mixed list removes the present key, tagKeys=[] -> BadRequestException 400; TagResource({}) -> 200 no-op**
  On a table tagged {a:1,b:2}: UntagResource(['zzz']) -> HTTP 204; UntagResource(['a','zzz']) -> HTTP 204, map
  afterwards {'b': '2'}. UntagResource(tagKeys=[]) -> BadRequestException (HTTP 400) 'The provided tags are
  invalid.' (client side: False; the SDK sends no tagKeys query parameter). TagResource(tags={}) -> HTTP 200.
  UntagResource(['a','a']) -> HTTP 204 (a removed: True). UntagResource with 50 keys of which only 'a' exists
  -> HTTP 204 (a removed: True); with 51 keys -> HTTP 204 (a removed: True).
  - ACK: tags.custom-sync, terminal_codes · ops: UntagResource, TagResource, ListTagsForResource · fields:
    tagKeys, tags
  - repro: UntagResource(['zzz']); UntagResource(['a','zzz']); UntagResource([]); TagResource({});
    UntagResource(['a','a']); UntagResource(['a', 49 absent keys]); UntagResource(51 keys)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](#s3t-table-143), [S3T-TABLEBUCKET-153](tablebucket.md#s3t-tablebucket-153), [S3T-TABLEBUCKET-154](tablebucket.md#s3t-tablebucket-154), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023),
    [S3T-TABLEBUCKET-155](tablebucket.md#s3t-tablebucket-155), [S3T-TABLE-116](#s3t-table-116), [S3T-TABLE-127](#s3t-table-127), [S3T-TABLE-128](#s3t-table-128), [S3T-TABLE-117](#s3t-table-117), [S3T-TABLE-119](#s3t-table-119),
    [S3T-TABLE-126](#s3t-table-126), [S3T-TABLE-118](#s3t-table-118) · hypotheses: H-T-093, H-T-204 · evidence:
    table/idempotency/untag-and-post-delete
  - notes: H-T-093 qualified: absent-key 204 and mixed removal confirmed; UntagResource([]) ->
    BadRequestException 400 'The provided tags are invalid.' confirmed (the SDK sends no tagKeys query
    parameter at all); but TagResource(tags={}) is ACCEPTED with HTTP 200 (predicted 400) - an empty-diff
    TagResource is...
  - full notes: [details/S3T-TABLE-125.md](details/S3T-TABLE-125.md)

## Errors

- <a id="s3t-table-048"></a>**S3T-TABLE-048** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable to a name that already exists in the target namespace -> ConflictException; same name in another namespace is allowed**
  RenameTable(ns1,a, newName=b) with ns1/b existing -> ConflictException (HTTP 409) 'A table with an identical
  name already exists in the namespace.'. RenameTable(ns1,a, newNamespaceName=ns2, newName=b) with ns2/b
  existing -> ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.'.
  GetTable(ns1,a) after each -> HTTP 200 / HTTP 200 (source unchanged: True). RenameTable(ns1,a, newName=zz)
  where zz exists only in ns2 -> HTTP 204 (names are scoped per namespace); renamed back -> HTTP 204.
  - ACK: terminal_codes, custom_update · ops: RenameTable · fields: newName, newNamespaceName
  - repro: Create ns1/a, ns1/b, ns2/b, ns2/zz -> RenameTable(ns1,a,newName=b);
    RenameTable(ns1,a,newNamespaceName=ns2,newName=b); RenameTable(ns1,a,newName=zz)
  - measurements: rename_collision_latency_ms=[55, 11], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-050](#s3t-table-050), [S3T-TABLE-049](#s3t-table-049) · hypotheses: H-T-056 · evidence: table/error-taxonomy/rename-errors
  - notes: H-T-056 confirmed (predicted ConflictException 409 for both collisions, success for a name existing
    only in another namespace). The 409 message is specific ('A table with an identical name already exists in
    the namespace.'), not the generic 'retry the request' shape text; the collision is permanent...
  - full notes: [details/S3T-TABLE-048.md](details/S3T-TABLE-048.md)

- <a id="s3t-table-049"></a>**S3T-TABLE-049** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable(newNamespaceName=<missing namespace>) -> NotFoundException; just-deleted namespace -> NotFoundException**
  RenameTable(ns1,a, newNamespaceName=<never existed>) -> NotFoundException (HTTP 404) 'The specified
  destination namespace does not exist.'. After CreateNamespace+DeleteNamespace(ns3) (delete -> HTTP 204):
  RenameTable(newNamespaceName=ns3) -> NotFoundException (HTTP 404) 'The specified destination namespace does
  not exist.'. With both newNamespaceName=<missing> and newName -> NotFoundException (HTTP 404) 'The specified
  destination namespace does not exist.'. GetTable(ns1,a) afterwards -> HTTP 200, name unchanged: True.
  - ACK: exceptions.404, custom_update · ops: RenameTable, DeleteNamespace · fields: newNamespaceName
  - repro: RenameTable(ns1,a,newNamespaceName=nope); CreateNamespace ns3; DeleteNamespace ns3;
    RenameTable(ns1,a,newNamespaceName=ns3)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-080](#s3t-table-080), [S3T-TABLE-094](#s3t-table-094), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-048](#s3t-table-048), [S3T-TABLE-050](#s3t-table-050) ·
    hypotheses: H-T-058 · evidence: table/error-taxonomy/rename-errors
  - notes: H-T-058 confirmed. The 404 message is 'The specified destination namespace does not exist.' -
    compare with the table-missing message 'The specified table does not exist.' to see whether the code alone
    can tell 'table gone' from 'target namespace missing'.

- <a id="s3t-table-082"></a>**S3T-TABLE-082** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable/DeleteTable with a missing namespace, table or bucket -> NotFoundException 404 (6/6 cells); DeleteTable messages name the parent**
  GetTable: missing_namespace -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  missing_table -> NotFoundException (HTTP 404) 'The specified table does not exist.'; missing_bucket ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'. DeleteTable: missing_namespace ->
  NotFoundException (HTTP 404) 'The specified namespace does not exist.'; missing_table -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; missing_bucket -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'. Distinct GetTable messages: 1; distinct DeleteTable messages: 3.
  - ACK: exceptions.404, requeue · ops: GetTable, DeleteTable · fields: namespace, name, tableBucketARN
  - repro: GetTable and DeleteTable with (a) missing namespace, (b) missing table, (c) missing bucket in an
    existing bucket/namespace
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-027](#s3t-table-027), [S3T-TABLE-267](#s3t-table-267), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-080](#s3t-table-080), [S3T-TABLE-094](#s3t-table-094),
    [S3T-TABLE-049](#s3t-table-049), [S3T-TABLE-326](#s3t-table-326) · hypotheses: H-T-043 · evidence: table/error-taxonomy/create-validation
  - notes: H-T-043 qualified: one code (NotFoundException) for all six calls; only DeleteTable's message
    distinguishes the missing parent, GetTable's does not. A ReadOne cannot tell 'namespace/bucket gone' from
    'table gone' (same code AND message); a controller wanting to requeue on a missing parent must check...
  - full notes: [details/S3T-TABLE-082.md](details/S3T-TABLE-082.md)

- <a id="s3t-table-094"></a>**S3T-TABLE-094** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables(namespace = missing | prefix of existing | just-deleted | upper-cased) -> NotFoundException 404; existing empty namespace -> 200**
  ListTables(namespace=<never existed>) -> NotFoundException (HTTP 404) 'The specified namespace does not
  exist.'. ListTables(namespace=<ns1 minus its last char>) -> NotFoundException (HTTP 404) 'The specified
  namespace does not exist.'. ListTables(namespace=<existing empty ns3>) -> HTTP 200 with 0 tables. After
  DeleteNamespace(ns3) (-> HTTP 204): ListTables(namespace=ns3) -> NotFoundException (HTTP 404) 'The specified
  namespace does not exist.'. Upper-cased namespace -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.' (no case folding).
  - ACK: exceptions.404, pre-delete-cleanup · ops: ListTables, DeleteNamespace · fields: namespace
  - repro: ListTables(namespace=nope); ListTables(namespace=<prefix of ns1>); CreateNamespace ns3;
    ListTables(namespace=ns3); DeleteNamespace ns3; ListTables(namespace=ns3)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-080](#s3t-table-080), [S3T-TABLE-049](#s3t-table-049), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-093](#s3t-table-093), [S3T-TABLE-097](#s3t-table-097) ·
    hypotheses: H-T-073 · evidence: table/weird-inputs/list-tables-bounds
  - notes: H-T-073 confirmed (predicted NotFoundException 404 for missing and just-deleted namespaces and
    exact-match semantics so a prefix also yields 404). Message: 'The specified namespace does not exist.'.
  - full notes: [details/S3T-TABLE-094.md](details/S3T-TABLE-094.md)

- <a id="s3t-table-120"></a>**S3T-TABLE-120** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **TagResource over-long key/value and empty key return wire code ValidationException (HTTP 400), which the s3tables model does not declare**
  TagResource(tableARN, {<129-char key>: v}) -> HTTP 400 with wire code ValidationException and message "1
  validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy constraint:
  [Member must have length less than or equal to 128, Member must have length greater than or equal to 1]";
  {k: <257-char value>} -> ValidationException 400 (Map value ... less than or equal to 256); a raw request
  with an empty key -> ValidationException 400 (same keys constraint). Every other rejection on the tag path
  (charset, aws: prefix) is BadRequestException 400. ValidationException is not among the modeled errors of
  TagResource (BadRequestException, ConflictException, ForbiddenException, InternalServerErrorException,
  NotFoundException, TooManyRequestsException), so the SDK surfaces it as a generic, untyped error.
  - ACK: terminal_codes, tags.custom-sync · ops: TagResource · fields: tags
  - repro: TagResource(tableARN, {'k'*129: 'v'}); TagResource(tableARN, {'k': 'v'*257}); raw POST /tag/<arn>
    {tags: {'': 'v'}}
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKET-067](tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](table-policy.md#s3t-table-083), [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLE-109](table-policy.md#s3t-table-109),
    [S3T-TABLEPOLICY-001](table-policy.md#s3t-tablepolicy-001), [S3T-TABLEBUCKETPOLICY-002](tablebucket-policy.md#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKET-090](tablebucket-policy.md#s3t-tablebucket-090) · hypotheses: H-T-092, H-T-090 ·
    evidence: table/tags/table-tag-semantics
  - notes: Side finding from H-T-092/H-T-090: terminal-error matching by typed exception (BadRequestException)
    would miss these permanent user errors and requeue forever; match the wire code string
    'ValidationException' too. The SDK's own client-side validation catches the empty key (min length 1) but
    not the...
  - full notes: [details/S3T-TABLE-120.md](details/S3T-TABLE-120.md)

- <a id="s3t-table-130"></a>**S3T-TABLE-130** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **UntagResource on a just-deleted table's ARN -> InternalServerErrorException 500 (2/2); TagResource/ListTagsForResource on it -> 404**
  Table created with tags {a:1}, then DeleteTable -> 204. UntagResource(old tableARN, ['a']) immediately after
  -> HTTP 500 InternalServerErrorException 'An internal error occurred. Try again.' (latency ~320 ms); the
  same call at t+5 s -> the same 500 (2/2). TagResource(old ARN) and ListTagsForResource(old ARN) over the
  same period -> NotFoundException 404 'The specified table does not exist.' every time (12/12). The table's
  bucket still existed; the triple was re-created afterwards under a new UUID.
  - ACK: exceptions.404, terminal_codes, pre-delete-cleanup · ops: UntagResource, DeleteTable, TagResource,
    ListTagsForResource · fields: resourceArn, tagKeys
  - repro: CreateTable(tags {a:1}) -> DeleteTable -> UntagResource(old tableARN, ['a']) at t+0 and t+5s;
    compare with TagResource/ListTagsForResource on the same ARN
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-027](#s3t-table-027), [S3T-TABLE-267](#s3t-table-267), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-143](#s3t-table-143), [S3T-TABLE-145](#s3t-table-145) ·
    hypotheses: H-T-095 · evidence: table/idempotency/untag-and-post-delete
  - notes: Side finding from H-T-095, n=2 in one probe run (confidence medium until reproduced): a finalizer
    or tag reconciler that untags a table whose deletion raced ahead receives a retryable-looking 500 for a
    permanent condition and would retry forever; it should treat a 500 from UntagResource on a...
  - full notes: [details/S3T-TABLE-130.md](details/S3T-TABLE-130.md)

- <a id="s3t-table-143"></a>**S3T-TABLE-143** `error-code` · impact medium · handled · verified 2026-10-13
  **Tag ops on a table ARN with a random UUID -> list/tag NotFoundException 404 but untag InternalServerError 500; malformed ARN forms -> 400**
  With a real table ARN <bucketARN>/table/<lowercase uuid> as the base, ListTagsForResource /
  TagResource({probe}) / UntagResource(['t']) returned: random_uuid: list NotFoundException (HTTP 404) 'The
  specified table does not exist.', tag NotFoundException (HTTP 404) 'The specified table does not exist.',
  untag InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.';
  wrong_service_plain_s3: list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.'; truncated_after_table: list BadRequestException (HTTP 400) 'The
  specified resourceArn is not valid.', tag BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.', untag BadRequestException (HTTP 400) 'The specified resourceArn is not valid.'; uuid_uppercased:
  list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400) 'The specified resourceArn
  is not valid.'; trailing_slash: list BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.', tag BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.'; other_region: list
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.', tag
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.', untag
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  tables_plural_segment: list BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', tag
  BadRequestException (HTTP 400) 'The specified resourceArn is not valid.', untag BadRequestException (HTTP 400)
  'The specified resourceArn is not valid.'. The real table's tags afterwards: {'t': '1'} (t intact: True).
  - ACK: exceptions.404, terminal_codes, is_arn_primary_key · ops: ListTagsForResource, TagResource,
    UntagResource · fields: resourceArn
  - repro: Take the tableARN; call the three tag ops with the UUID replaced by uuid4(),
    'arn:aws:s3:::<bucket>', ARN cut after '/table/', UUID upper-cased, ARN + '/', region swapped, '/tables/'
    segment
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125),
    [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032), [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-145](#s3t-table-145), [S3T-TABLE-267](#s3t-table-267), [S3T-TABLE-027](#s3t-table-027), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLE-244](#s3t-table-244), [S3T-TABLE-097](#s3t-table-097) · hypotheses: H-T-086, H-T-207 · evidence: table/identity/tag-arn-forms
  - notes: H-T-086 qualified: see codes (predicted: unknown UUID 404; wrong-service and truncated 400;
    upper-cased UUID 404 without case folding). H-T-207 clauses upper-cased UUID -> 404 and trailing slash ->
    400: qualified: upper BadRequestException, slash BadRequestException. Messages: random uuid 'The...
  - full notes: [details/S3T-TABLE-143.md](details/S3T-TABLE-143.md)

- <a id="s3t-table-186"></a>**S3T-TABLE-186** `error-code` · impact high · handled · verified 2026-10-13
  **10 concurrent TagResource on one table -> 8-9/10 InternalServerErrorException 500 (no 409; 3 rounds), 3-6 s wall; concurrent Untag 9/10 500**
  Three rounds of 10 concurrent TagResource (distinct keys) on one table: per-round codes [{'200': 2,
  'InternalServerErrorException/500': 8}, {'200': 1, 'InternalServerErrorException/500': 9}, {'200': 2,
  'InternalServerErrorException/500': 8}] (wall [6.18, 3.15, 3.16] s); ListTagsForResource afterwards showed
  [2, 1, 2] keys present per round, of which [[], [], []] were applied by calls that returned an error (0) and
  [[], [], []] successful calls lost their key. Error messages: {"InternalServerErrorException": "An internal
  error occurred. Try again."}. 10 concurrent UntagResource (one existing key each) -> {"204": 1,
  "InternalServerErrorException/500": 9} (wall 4.14s); 9 keys remained, [] removed despite an error, [] still
  present despite a 2xx. 30 sequential UntagResource with no pause -> {"204": 30} in 3.08s (first error at
  index None), 0 of the 30 keys left.
  - ACK: tags.custom-sync, requeue, one-per-reconcile · ops: TagResource, UntagResource, ListTagsForResource ·
    fields: tags, tagKeys
  - repro: 10 threads x TagResource(tableArn, {k_i: v}) x3; TagResource(10 keys) then 10 threads x
    UntagResource([k_i]); TagResource(30 keys) then 30 sequential UntagResource
  - measurements: tag_race_wall_s=[6.18, 3.15, 3.16], untag_race_wall_s=4.14, untag_sequential_wall_s=3.08,
    untag_sequential_latency_ms_min_max=[88, 133], n=70
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-188](#s3t-table-188), [S3T-TABLE-218](#s3t-table-218), [S3T-TABLE-217](#s3t-table-217) · hypotheses: H-T-098, H-T-209 · evidence:
    table/limits/tag-count-and-read-bursts
  - notes: H-T-098 refuted on the code: concurrent tag writes to one table fail with
    InternalServerErrorException 500 'An internal error occurred. Try again.' (8-9 of 10 in each of 3 rounds),
    never ConflictException; in these 3 rounds the losing keys were NOT applied and no successful key was lost
    (on a table...
  - full notes: [details/S3T-TABLE-186.md](details/S3T-TABLE-186.md)

- <a id="s3t-table-324"></a>**S3T-TABLE-324** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **UntagResource on any nonexistent table ARN (deleted or never-created) -> InternalServerErrorException 500 for 12 min; Tag/ListTags -> 404**
  Three clean runs. Two tables per run created with tags {a:1}, DeleteTable -> HTTP 204; UntagResource(old
  tableARN, ['a']) at t+0 and t+5 s -> InternalServerErrorException (HTTP 500) 'An internal error occurred.
  Try again.' in 12/12 (x-amzn-ErrorType
  'InternalServerErrorException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', raw body
  {"message":"An internal error occurred. Try again."}, 314-388 ms), while TagResource and ListTagsForResource
  on the same ARN -> NotFoundException (HTTP 404) 'The specified table does not exist.' in 24/24. The same 500
  came back for UntagResource with a key the table never had (['zzz'], 3/3), for a well-formed table ARN whose
  uuid never existed in the live bucket (3/3; TagResource/ListTagsForResource on it -> 404), 180 s after the
  delete (2/2; +36 s in run 1), and for ARNs deleted 8.3 and 11.7 min earlier whose table BUCKET had been
  deleted too (2/2, ListTagsForResource -> 404 'The specified table does not exist.'). In total 23/23
  UntagResource calls on a nonexistent table id -> 500; 35/35 TagResource/ListTagsForResource -> 404.
  - ACK: exceptions.404, terminal_codes, tags.custom-sync, pre-delete-cleanup · ops: UntagResource,
    TagResource, ListTagsForResource, DeleteTable · fields: resourceArn, tagKeys
  - repro: CreateTable(tags {a:1}) x2 -> DeleteTable -> UntagResource/TagResource/ListTagsForResource on the
    old ARN at t+0, t+5 s and minutes later; UntagResource on <bucketARN>/table/<random uuid>
  - measurements: untag_500_latency_ms.min=314, untag_500_latency_ms.max=388, untag_500_latency_ms.n=23,
    late_samples_after_delete_s=[36, 181, 181], bucket_deleted_samples_age_min=[11.7, 8.3], n=23
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-130](#s3t-table-130) · evidence: table/creative/reverify-set-c
  - notes: Precision gain on [S3T-TABLE-130](#s3t-table-130) (which reproduced: 500 at t+0 and t+5 s, Tag/ListTags 404). The 500
    is not a transient post-delete artefact: it is the UntagResource not-found path itself (any unknown table
    uuid, any key, minutes later, even after the bucket is gone), so a controller must treat...
  - full notes: [details/S3T-TABLE-324.md](details/S3T-TABLE-324.md)

- <a id="s3t-table-326"></a>**S3T-TABLE-326** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **GetTable NotFound message: well-formed missing namespace -> 'table does not exist'; upper-cased/trailing-space -> 'namespace does not exist'**
  Existing bucket/namespace/table; GetTable(tableBucketARN, namespace, name) 3 trials per cell, outcomes per
  cell: {"bucket_never_created": ["NotFoundException (HTTP 404) 'The specified table does not exist.'"],
  "control_existing": ["HTTP 200"], "ns_and_table_never_created": ["NotFoundException (HTTP 404) 'The
  specified table does not exist.'"], "ns_just_deleted": ["NotFoundException (HTTP 404) 'The specified table
  does not exist.'"], "ns_never_created": ["NotFoundException (HTTP 404) 'The specified table does not
  exist.'"], "ns_trailing_space": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "ns_upper_cased": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "ns_upper_table_never": ["NotFoundException (HTTP 404) 'The specified namespace does not exist.'"],
  "table_never_created": ["NotFoundException (HTTP 404) 'The specified table does not exist.'"]}. DeleteTable
  controls (1 trial): {"bucket_never_created": "NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'", "ns_never_created": "NotFoundException (HTTP 404) 'The specified namespace does not exist.'",
  "table_never_created": "NotFoundException (HTTP 404) 'The specified table does not exist.'"}.
  - ACK: exceptions.404, requeue · ops: GetTable, DeleteTable · fields: namespace, name, tableBucketARN
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable; GetTable with the namespace never created /
    created+deleted / upper-cased / with a trailing space, with a never-created bucket ARN, with a
    never-created table name; 3 trials each
  - measurements: trials_per_cell=3, cells=9, n=27
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-027](#s3t-table-027), [S3T-TABLE-267](#s3t-table-267) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa
  - notes: FU2-06: both readings are right under different conditions - the GetTable message depends on the
    SHAPE of the missing namespace, not on whether it was deleted or never created: a well-formed lowercase
    namespace that does not exist (never created 3/3, created-then-deleted 3/3, never-created bucket...
  - full notes: [details/S3T-TABLE-326.md](details/S3T-TABLE-326.md)

## Request validation

- <a id="s3t-table-024"></a>**S3T-TABLE-024** `request-validation` · impact medium · handled · verified 2026-10-13
  **GetTable with no identifier -> BadRequestException; partial composite key (bucket+name, bucket+namespace...**
  All four GetTable input members are optional in the model and boto3 sends the request. no params ->
  BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table bucket
  arn, namespace, table name) together'; namespace only -> BadRequestException (HTTP 400) 'Invalid input: Must
  provide either table arn alone OR all of (table bucket arn, namespace, table name) together';
  tableBucketARN+name (no namespace) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either
  table arn alone OR all of (table bucket arn, namespace, table name) together'; tableBucketARN+namespace (no
  name) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table
  bucket arn, namespace, table name) together'; tableBucketARN only -> BadRequestException (HTTP 400) 'Invalid
  input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name) together';
  namespace+name (no bucket) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn
  alone OR all of (table bucket arn, namespace, table name) together'.
  - ACK: custom_find, terminal_codes · ops: GetTable · fields: tableArn, tableBucketARN, namespace, name
  - repro: GetTable() / GetTable(namespace=ns) / GetTable(tableBucketARN, name) / GetTable(tableBucketARN,
    namespace) / GetTable(tableBucketARN)
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-269](table-metadata-location.md#s3t-table-269), [S3T-TABLE-270](#s3t-table-270), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-025](#s3t-table-025), [S3T-TABLE-267](#s3t-table-267) ·
    hypotheses: H-T-003, H-T-006, H-U-190 · evidence: table/identity/get-table-identifier-forms
  - notes: H-T-003 confirmed (no client-side validation, server BadRequestException (HTTP 400) 'Invalid input:
    Must provide either table arn alone OR all of (table bucket arn, namespace, table name) together').
    H-T-006 confirmed (partial key -> BadRequestException (HTTP 400) 'Invalid input: Must provide either...
  - full notes: [details/S3T-TABLE-024.md](details/S3T-TABLE-024.md)

- <a id="s3t-table-025"></a>**S3T-TABLE-025** `request-validation` · impact high · handled · verified 2026-10-15, re-verified
  **GetTable with tableArn AND bucket+namespace+name -> BadRequestException (HTTP 400) even when both name the SAME table**
  tableArn(t1)+triple(t2) -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn
  alone OR all of (table bucket arn, namespace, table name) together'; tableArn(t2)+triple(t1) ->
  BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table bucket
  arn, namespace, table name) together'; tableArn(t1)+triple(t1) -> BadRequestException (HTTP 400) 'Invalid
  input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name) together';
  tableArn(t1)+triple(nonexistent ns/name) -> BadRequestException (HTTP 400) 'Invalid input: Must provide
  either table arn alone OR all of (table bucket arn, namespace, table name) together'; tableArn(t1)+name(t2)
  only -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn alone OR all of (table
  bucket arn, namespace, table name) together'; tableArn='not-an-arn'+triple(t1) -> BadRequestException (HTTP 400)
  'Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name)
  together'.
  - ACK: custom_find · ops: GetTable · fields: tableArn, tableBucketARN, namespace, name
  - repro: Create t1,t2 in one namespace; GetTable(tableArn=arn(t1), tableBucketARN, namespace, name=t2) and
    the reverse; also tableArn + nonexistent triple
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-269](table-metadata-location.md#s3t-table-269), [S3T-TABLE-270](#s3t-table-270), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-024](#s3t-table-024), [S3T-TABLE-267](#s3t-table-267) ·
    hypotheses: H-T-004, H-T-005, H-U-190 · evidence: table/identity/get-table-identifier-forms
  - notes: H-T-004 (400 mutually exclusive) confirmed; H-T-005 (200, ARN wins silently) refuted; H-U-190 (both
    forms -> 400) confirmed. No identifier was honored: every mixed request was rejected before lookup (same
    message for a nonexistent triple and for a garbage ARN with a valid triple).

- <a id="s3t-table-027"></a>**S3T-TABLE-027** `request-validation` · impact medium · handled · verified 2026-10-13
  **GetTable/DeleteTable with the existing table's name upper-cased -> NotFoundException / NotFoundException; table untouched**
  For existing table '<lower>' GetTable(name=<UPPER>) -> NotFoundException (HTTP 404) 'The specified table
  does not exist.'; DeleteTable(name=<UPPER>) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(namespace=<UPPER>, name=<lower>) -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.'; afterwards GetTable(name=<lower>) -> HTTP 200 (table still exists).
  - ACK: terminal_codes · ops: GetTable, DeleteTable · fields: name, namespace
  - repro: Create t1 (lower-case); GetTable/DeleteTable with name=t1.upper(); GetTable(name=t1)
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:10-13; generator.yaml:31-39`
  - related: [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-267](#s3t-table-267), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-052](#s3t-table-052),
    [S3T-TABLE-143](#s3t-table-143), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLE-326](#s3t-table-326) · hypotheses: H-T-011 · evidence:
    table/identity/get-table-identifier-forms
  - notes: H-T-011 REFUTED on the error code: no case folding (the lower-case table is untouched: True); an
    upper-case name is NOT rejected as a pattern violation (no 400); it is simply looked up and yields
    NotFoundException (HTTP 404), indistinguishable from a missing table.
  - full notes: [details/S3T-TABLE-027.md](details/S3T-TABLE-027.md)

- <a id="s3t-table-047"></a>**S3T-TABLE-047** `request-validation` · impact medium · handled · verified 2026-10-13
  **CreateTable format other than exact 'ICEBERG' ('iceberg', DELTA, HUDI, '' or omitted) -> BadRequestException 400 'format is not valid'**
  With SDK validation disabled, no metadata: format='iceberg' -> BadRequestException (HTTP 400) 'The specified
  open table format is not valid'. 'DELTA' -> BadRequestException (HTTP 400) 'The specified open table format
  is not valid'. 'HUDI' -> BadRequestException (HTTP 400) 'The specified open table format is not valid'. ''
  -> BadRequestException (HTTP 400) 'The specified open table format is not valid'. format omitted ->
  BadRequestException (HTTP 400) 'The specified open table format is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: format
  - repro: CreateTable with each format value (parameter_validation=False); GetTable afterwards
  - handling: handled via `test/e2e/tests/test_table.py:175-193; test/e2e/resources/table.yaml; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-079](#s3t-table-079), [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-035](#s3t-table-035), [S3T-TABLE-080](#s3t-table-080) · hypotheses: H-T-127, H-T-128 ·
    evidence: table/weird-inputs/schema-variants
  - notes: H-T-127 confirmed; H-T-128 refuted (omitted format -> BadRequestException 400). One attempt each.
  - full notes: [details/S3T-TABLE-047.md](details/S3T-TABLE-047.md)

- <a id="s3t-table-050"></a>**S3T-TABLE-050** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable to the CURRENT name/namespace is not a no-op: ConflictException 409 'identical name already exists'; neither field -> 400**
  On ns1/a: RenameTable(newName=a) -> ConflictException (HTTP 409) 'A table with an identical name already
  exists in the namespace.' (modifiedAt changed: False); RenameTable(newNamespaceName=ns1) ->
  ConflictException (HTTP 409) 'A table with an identical name already exists in the namespace.' (modifiedAt
  changed: False); RenameTable(newName=a, newNamespaceName=ns1) -> ConflictException (HTTP 409) 'A table with
  an identical name already exists in the namespace.' (modifiedAt changed: False); RenameTable with neither
  newName nor newNamespaceName -> BadRequestException (HTTP 400) 'Neither a new namespace name nor a new table
  name is specified.'.
  - ACK: custom_update, compare.is_ignored+delta_pre_compare · ops: RenameTable · fields: newName,
    newNamespaceName
  - repro: RenameTable(ns1,a,newName=a); RenameTable(ns1,a,newNamespaceName=ns1);
    RenameTable(ns1,a,newName=a,newNamespaceName=ns1); RenameTable(ns1,a)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-272](#s3t-table-272), [S3T-TABLE-273](#s3t-table-273), [S3T-TABLE-274](#s3t-table-274), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](#s3t-table-276),
    [S3T-NAMESPACE-036](namespace.md#s3t-namespace-036), [S3T-NAMESPACE-021](namespace.md#s3t-namespace-021), [S3T-TABLE-064](table-metadata-location.md#s3t-table-064), [S3T-TABLE-048](#s3t-table-048), [S3T-TABLE-049](#s3t-table-049) · hypotheses: H-T-059,
    H-T-060 · evidence: table/error-taxonomy/rename-errors
  - notes: H-T-059 qualified: the contrarian claim 'same-name rename is rejected, not a harmless no-op' holds,
    but the code is ConflictException (HTTP 409) 'A table with an identical name already exists in the
    namespace.' rather than BadRequestException 400; the same 409 for newNamespaceName=<current ns> alone...
  - full notes: [details/S3T-TABLE-050.md](details/S3T-TABLE-050.md)

- <a id="s3t-table-052"></a>**S3T-TABLE-052** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable.newName: 'aws_' prefix, 255 chars accepted; upper, '-', leading '_', 256 chars, '.', non-ASCII, '' -> 400 'name is not valid'**
  Per newName candidate (SDK pattern/max not enforced client-side, so the server answered): uppercase ->
  BadRequestException (HTTP 400) 'The specified table name is not valid.'; hyphen -> BadRequestException (HTTP 400)
  'The specified table name is not valid.'; leading_underscore -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; aws_prefix -> HTTP 204; len_256 -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; len_255 -> HTTP 204; dot -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; unicode -> BadRequestException (HTTP 400) 'The specified table name is
  not valid.'. Empty newName via a raw signed request -> HTTP 400 BadRequestException {"message":"A table name
  is not specified"}. Final GetTable(ns1,a) -> HTTP 200 (name restored: True).
  - ACK: terminal_codes, custom_update · ops: RenameTable · fields: newName
  - repro: RenameTable(ns1,a,newName=<candidate>) for uppercase, hyphen, leading underscore, aws_ prefix, 256
    chars, 255 chars, dot, unicode; raw PUT .../rename {newName:''}
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-027](#s3t-table-027) · hypotheses: H-T-065 · evidence: table/error-taxonomy/rename-errors
  - notes: H-T-065 qualified: uppercase, hyphen, leading underscore, 256 chars, '.', non-ASCII and empty are
    rejected with BadRequestException 400, but every rejection carries the same message 'The specified table
    name is not valid.' (empty: 'A table name is not specified') - it does not name the rule; an...
  - full notes: [details/S3T-TABLE-052.md](details/S3T-TABLE-052.md)

- <a id="s3t-table-078"></a>**S3T-TABLE-078** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTable name: 255 chars, 'a__b', trailing '_', leading digit, 'aws_*' ok; 256, leading '_', '__', upper, '-', '.', space, non-ASCII 400**
  CreateTable(format=ICEBERG) per name cell: len255 -> HTTP 200; len256 -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; non_ascii -> BadRequestException (HTTP 400) 'The specified table name
  is not valid.'; hyphen -> BadRequestException (HTTP 400) 'The specified table name is not valid.'; dot ->
  BadRequestException (HTTP 400) 'The specified table name is not valid.'; space -> BadRequestException (HTTP 400)
  'The specified table name is not valid.'; uppercase -> BadRequestException (HTTP 400) 'The specified table
  name is not valid.'; leading_underscore -> BadRequestException (HTTP 400) 'The specified table name is not
  valid.'; trailing_underscore -> HTTP 200; leading_digit -> HTTP 200; double_underscore_inside -> HTTP 200;
  only_underscores -> BadRequestException (HTTP 400) 'The specified table name is not valid.'; aws_prefix ->
  HTTP 200; aws_literal -> HTTP 200; aws_s3_metadata_prefix -> HTTP 200. Server messages seen: ['The specified
  table name is not valid.']. Raw (non-boto3) request with name='' -> HTTP 400
  BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/ "{\"message\":\"A table name
  is not specified\"}"; without the name member -> HTTP 400 "{\"message\":\"A table name is not specified\"}".
  After the uppercase attempt, GetTable(lower-cased name) -> NotFoundException (HTTP 404) 'The specified table
  does not exist.' and GetTable(uppercase name) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(hyphen name) -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: terminal_codes · ops: CreateTable, GetTable · fields: name
  - repro: CreateTable with each name cell (255/256 chars, non-ASCII, hyphen, dot, space, uppercase,
    leading/trailing underscore, leading digit, '__', 'aws_*', 'aws')
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:99-104; generator.yaml:170-175; generator.yaml:45-46; generator.yaml:60-65`
  - related: [S3T-TABLE-047](#s3t-table-047), [S3T-TABLE-079](#s3t-table-079), [S3T-TABLE-035](#s3t-table-035), [S3T-TABLE-080](#s3t-table-080), [S3T-TABLE-052](#s3t-table-052), [S3T-TABLE-027](#s3t-table-027) ·
    hypotheses: H-T-013, H-T-014, H-T-015, H-T-016, H-T-017 · evidence: table/error-taxonomy/create-validation
  - notes: H-T-013 confirmed (255 -> HTTP 200; 256 -> BadRequestException (HTTP 400) 'The specified table name
    is not valid.'; empty -> client-side ParamValidationError in boto3, server 400; non-ASCII ->
    BadRequestException (HTTP 400) 'The specified table name is not valid.'). H-T-014 confirmed (hyphen ->...
  - full notes: [details/S3T-TABLE-078.md](details/S3T-TABLE-078.md)

- <a id="s3t-table-079"></a>**S3T-TABLE-079** `request-validation` · impact medium · handled · verified 2026-10-13
  **CreateTable format is case-sensitive server-side: 'iceberg'/'Iceberg'/'DELTA'/'PARQUET'/'' -> BadRequestException**
  CreateTable with format values: DELTA -> BadRequestException (HTTP 400) 'The specified open table format is
  not valid'; iceberg_lower -> BadRequestException (HTTP 400) 'The specified open table format is not valid';
  Iceberg_mixed -> BadRequestException (HTTP 400) 'The specified open table format is not valid'; empty ->
  BadRequestException (HTTP 400) 'The specified open table format is not valid'; PARQUET ->
  BadRequestException (HTTP 400) 'The specified open table format is not valid'. boto3 does not validate the
  enum client-side (client_side=False). GetTable for the lower-case attempt afterwards -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'.
  - ACK: terminal_codes · ops: CreateTable · fields: format
  - repro: CreateTable(format='DELTA'), ('iceberg'), ('Iceberg'), (''), ('PARQUET'); GetTable the name
  - handling: handled via `test/e2e/tests/test_table.py:175-193; test/e2e/resources/table.yaml; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-047](#s3t-table-047), [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-035](#s3t-table-035), [S3T-TABLE-080](#s3t-table-080) · hypotheses: H-T-019 · evidence:
    table/error-taxonomy/create-validation
  - notes: H-T-019 confirmed.

- <a id="s3t-table-093"></a>**S3T-TABLE-093** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables.prefix: literal, case-sensitive table-name prefix across namespaces, ANDed with namespace; '*' matches none; '' -> 200 unfiltered**
  prefix=<base>_ab without namespace -> ['abc', 'abd', 'abz'] (tables from both namespaces); with
  namespace=ns1 -> ['abc', 'abd'] (ANDed); prefix=<namespace name> -> []; prefix='<ns>.<name>' -> [];
  prefix='*' -> []; upper-cased prefix -> []; exact full name -> ['abc']; 300-char prefix ->
  BadRequestException (HTTP 400) 'The specified search prefix is not valid. A search prefix can not exceed 255
  characters.'. prefix='' via SDK -> ParamValidationError (HTTP None) 'Parameter validation failed:\nInvalid
  length for parameter prefix, value: 0, valid min length: 1'; via raw request -> HTTP 200
  {"tables":[{"createdAt":"2026-10-13T22:30:56.252921345Z","modifiedAt":"2026-10-13T22:30:56.252921345Z","name":"ackq_212a
  [truncated raw body]
  - ACK: list_operation.match_fields, custom_find · ops: ListTables · fields: prefix, namespace
  - repro: Tables ns1/abc, ns1/abd, ns2/abz, ns1/xyz...; ListTables with prefix in {ab, <ns1>, <ns1>.<name>,
    *, AB, <full name>, 300 chars, ''} and (namespace=ns1, prefix=ab)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-094](#s3t-table-094), [S3T-TABLE-097](#s3t-table-097) · hypotheses: H-T-072 · evidence:
    table/weird-inputs/list-tables-bounds
  - notes: H-T-072 qualified: all semantics clauses confirmed (literal, case-sensitive, name only - a
    namespace name or 'ns.name' as prefix matches nothing - across namespaces, ANDed with the namespace
    filter, '*' is literal and matches nothing), but the empty-prefix clause is refuted: the SDK rejects ''...
  - full notes: [details/S3T-TABLE-093.md](details/S3T-TABLE-093.md)

- <a id="s3t-table-126"></a>**S3T-TABLE-126** `request-validation` · impact medium · handled · verified 2026-10-13
  **UntagResource with one 129-char key anywhere in tagKeys -> ValidationException 400 (unmodeled code) and nothing is removed**
  Table tagged {k1:v}: UntagResource(['k1', <129-char key>]) -> ValidationException (HTTP 400) "1 validation
  error detected: Value at 'tagKeys' failed to satisfy constraint: Member must satisfy constraint: [Member
  must have length less than or equal to 128", k1 still present: True. UntagResource([<129-char key>, 'k1'])
  -> ValidationException (HTTP 400) "1 validation error detected: Value at 'tagKeys' failed to satisfy
  constraint: Member must satisfy constraint: [Member must have length less than or equal to 128", k1 still
  present: True. UntagResource(['k1', '']) -> ParamValidationError (HTTP None) 'Parameter validation
  failed:\nInvalid length for parameter tagKeys[1], value: 0, valid min length: 1' (client side: True), k1
  still present: True.
  - ACK: tags.custom-sync, terminal_codes · ops: UntagResource, ListTagsForResource · fields: tagKeys
  - repro: TagResource({k1:v}); UntagResource(['k1','x'*129]); ListTagsForResource;
    UntagResource(['x'*129,'k1']); ListTagsForResource
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-119](#s3t-table-119), [S3T-TABLE-117](#s3t-table-117), [S3T-TABLE-118](#s3t-table-118), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-125](#s3t-table-125) · hypotheses: H-T-205 ·
    evidence: table/idempotency/untag-and-post-delete
  - notes: H-T-205 qualified: whole-list validation is atomic and order-independent (k1 kept in both orders)
    as predicted, but the wire code is ValidationException (HTTP 400, 'Value at tagKeys failed to satisfy
    constraint: Member must have length less than or equal to 128'), not the modeled...
  - full notes: [details/S3T-TABLE-126.md](details/S3T-TABLE-126.md)

- <a id="s3t-table-168"></a>**S3T-TABLE-168** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables maxTables=0 or -1 -> BadRequestException 400 'limit should be a positive number'; 1000 and 1001 -> 200; maxTables=1 walk: 3 pages**
  Namespace holding exactly 3 tables, client validation off: maxTables=0 -> BadRequestException (HTTP 400)
  'The specified limit is not valid. A limit should be a positive number'; maxTables=1000 -> HTTP 200, 3
  items, token absent; maxTables=1001 -> HTTP 200, 3 items, token absent; maxTables=-1 -> BadRequestException
  (HTTP 400) 'The specified limit is not valid. A limit should be a positive number'. maxTables=1 + token
  following: 3 pages, items ['ackq_f745ff_t1', 'ackq_f745ff_t2', 'ackq_f745ff_t3'] (0 duplicates),
  continuationToken lengths [640, 640, 0], final page omits continuationToken.
  - ACK: custom_find, pre-delete-cleanup · ops: ListTables · fields: maxTables, continuationToken, namespace
  - repro: 3 x CreateTable in one namespace; ListTables(maxTables=0|1000|1001|-1); ListTables(maxTables=1)
    following tokens
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-TABLE-293](#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLE-092](#s3t-table-092),
    [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](#s3t-table-091), [S3T-TABLE-305](#s3t-table-305), [S3T-TABLE-095](#s3t-table-095), [S3T-TABLE-217](#s3t-table-217) · hypotheses: H-N-032,
    H-T-070 · evidence: namespace/limits/pagination-and-bursts
  - notes: H-N-032 confirmed for pagination (1 item per page, 3 pages, no duplicates, 640-char tokens, token
    absent on the last page) and for maxTables=0 -> BadRequestException 400; H-T-070 qualified: maxTables=1000
    -> 200 and -1 -> 400 as predicted, but maxTables=1001 -> HTTP 200 (accepted, not a 400), and...
  - full notes: [details/S3T-TABLE-168.md](details/S3T-TABLE-168.md)

## Update granularity and ordering

- <a id="s3t-table-080"></a>**S3T-TABLE-080** `prerequisite` · impact high · handled · verified 2026-10-13
  **CreateTable into a nonexistent namespace -> NotFoundException (HTTP 404); the namespace is NOT auto-created**
  CreateTable(existing bucket, namespace=<never created>, name, format=ICEBERG) -> NotFoundException (HTTP 404)
  'The specified namespace does not exist.'. GetNamespace(<never created>) afterwards -> NotFoundException
  (HTTP 404) 'The specified namespace does not exist.'; ListNamespaces shows it: False.
  - ACK: requeue, references · ops: CreateTable, GetNamespace · fields: namespace
  - repro: CreateTable with a namespace that was never created; GetNamespace; ListNamespaces
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-094](#s3t-table-094), [S3T-TABLE-049](#s3t-table-049), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-047](#s3t-table-047), [S3T-TABLE-079](#s3t-table-079),
    [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-035](#s3t-table-035) · hypotheses: H-T-020 · evidence: table/error-taxonomy/create-validation
  - notes: H-T-020 confirmed. The same NotFoundException code is returned for a missing table on GetTable (see
    the NotFound-variants finding); only the message differs.

- <a id="s3t-table-081"></a>**S3T-TABLE-081** `prerequisite` · impact medium · handled · verified 2026-10-13
  **CreateTable: nonexistent bucket ARN -> NotFoundException; bucket-less/s3/garbage ARN -> BadRequestException...**
  CreateTable(tableBucketARN=variant, namespace, name, format=ICEBERG): nonexistent_bucket ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; no_bucket_prefix -> BadRequestException
  (HTTP 400) 'The specified table bucket ARN is not valid.'; s3_bucket_arn -> BadRequestException (HTTP 400)
  'The specified table bucket ARN is not valid.'; not_an_arn -> BadRequestException (HTTP 400) 'The specified
  table bucket ARN is not valid.'; foreign_account_bucket -> ForbiddenException (HTTP 403) 'Unauthorized';
  table_arn_as_bucket_arn -> BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.'.
  - ACK: requeue, terminal_codes, references · ops: CreateTable · fields: tableBucketARN
  - repro: CreateTable with tableBucketARN = never-created bucket /
    'arn:aws:s3tables:<region>:<acct>:nobucketprefix' / 'arn:aws:s3:::<bucket>' / 'not-an-arn' / foreign
    account / a table ARN
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLE-083](table-policy.md#s3t-table-083) · hypotheses: H-T-021 ·
    evidence: table/error-taxonomy/create-validation
  - notes: H-T-021 confirmed. Full original title: CreateTable: nonexistent bucket ARN -> NotFoundException;
    bucket-less/s3/garbage ARN -> BadRequestException; foreign-account bucket ARN -> ForbiddenException

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-113"></a>**S3T-TABLE-113** `read-gap` · impact high · handled · verified 2026-10-13
  **CreateTable tags readable on the first ListTagsForResource (HTTP 200, 4 keys); GetTable has no tags field; untagged table -> tags {} present**
  CreateTable(tags={a,b,env,ackq-run}) then ListTagsForResource(tableARN) as the very next call -> HTTP 200
  with 4 keys (all four present: True). GetTable right after -> HTTP 200; its keys are ['createdAt',
  'createdBy', 'format', 'modifiedAt', 'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableARN',
  'tableBucketId', 'type', 'versionToken', 'warehouseLocation'] (no 'tags' key: True). For a table created
  without tags, the first ListTagsForResource -> HTTP 200, 'tags' key present in the parsed body: True, value
  {}.
  - ACK: tags.custom-sync, compare.nil_equals_zero_value · ops: CreateTable, ListTagsForResource, GetTable ·
    fields: tags
  - repro: CreateTable(tags) -> ListTagsForResource(tableARN); GetTable; CreateTable(no tags) ->
    ListTagsForResource
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](#s3t-table-143), [S3T-TABLEBUCKET-156](tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](tablebucket.md#s3t-tablebucket-157), [S3T-TABLE-127](#s3t-table-127), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028),
    [S3T-TABLEBUCKET-121](tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-007](tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLE-018](table-schema.md#s3t-table-018), [S3T-TABLE-310](table-schema.md#s3t-table-310),
    [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-090](table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-028](table-metadata-location.md#s3t-table-028) · hypotheses: H-T-083 · evidence:
    table/tags/table-tag-semantics
  - notes: H-T-083 confirmed.

- <a id="s3t-table-117"></a>**S3T-TABLE-117** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Tag keys/values stored verbatim (' k ' distinct from 'k', empty value kept as ''); empty key -> HTTP 400 wire code ValidationException**
  TagResource({' k ':v, k:v2, e:''}) -> HTTP 200; ListTagsForResource -> {'team': 'b', 'k': 'v2', ' k ': 'v',
  'e': ''} (' k ' present with spaces: True; e == '': True). Empty key via SDK -> ParamValidationError (HTTP
  None) 'Parameter validation failed:\nInvalid length for parameter tags (key: ), value: 0, valid min length:
  1' (client side: True); via raw request -> HTTP 400 ValidationException {"message":"1 validation error
  detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy constraint:. Value ' spaced
  value ' -> HTTP 200, read back as ' spaced value '.
  - ACK: tags.custom-sync · ops: TagResource, ListTagsForResource · fields: tags
  - repro: TagResource({' k ':'v','k':'v2','e':''}); ListTagsForResource; raw POST /tag/<arn> {tags:{'':'v'}}
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLE-127](#s3t-table-127), [S3T-TABLE-128](#s3t-table-128), [S3T-TABLE-116](#s3t-table-116), [S3T-TABLE-119](#s3t-table-119),
    [S3T-TABLE-126](#s3t-table-126), [S3T-TABLE-118](#s3t-table-118), [S3T-TABLE-185](#s3t-table-185) · hypotheses: H-T-090 · evidence:
    table/tags/table-tag-semantics
  - notes: H-T-090 confirmed (predicted: verbatim incl. spaces, empty key 400, empty value accepted and
    returned as '').

- <a id="s3t-table-217"></a>**S3T-TABLE-217** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables without maxTables on a bucket with 120 tables returns 120 items with NO continuationToken; maxTables=100 returns 100 + token**
  120 tables were created sequentially across 2 namespaces ({"200": 120} in 39.5s, latency [278, 395] ms).
  ListTables(tableBucketARN) with maxTables omitted -> 120 items, continuationToken ABSENT; per namespace ->
  {'ackq_4e333a_ns1': {'n_items': 60, 'token_present': False}, 'ackq_4e333a_ns2': {'n_items': 60,
  'token_present': False}}. ListTables(maxTables=100) -> 100 items, continuationToken present.
  - ACK: custom_find, pre-delete-cleanup · ops: ListTables, CreateTable · fields: maxTables, continuationToken
  - repro: 120 x CreateTable across 2 namespaces; ListTables(bucket) without maxTables;
    ListTables(maxTables=100)
  - measurements: create_120_wall_s=39.53, create_latency_ms_min_max=[278, 395], n=120
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-091](#s3t-table-091), [S3T-TABLE-305](#s3t-table-305), [S3T-TABLE-168](#s3t-table-168), [S3T-TABLE-095](#s3t-table-095), [S3T-TABLE-188](#s3t-table-188), [S3T-TABLE-218](#s3t-table-218),
    [S3T-TABLE-186](#s3t-table-186) · hypotheses: H-T-071 · evidence: table/limits/default-page-120-tables
  - notes: H-T-071 confirmed (predicted all 120 in one page with no token, i.e. default page size >= 120, not
    100). The sequential CreateTable loop with no pause also shows whether table creation is rate-limited like
    CreateNamespace: no 429 in 120 calls.

- <a id="s3t-table-317"></a>**S3T-TABLE-317** `read-gap` · impact medium · handled · verified 2026-10-14
  **ListTables entries on fresh customer tables omit namespaceId, tableBucketId and managedByService although the model declares them**
  ListTables(tableBucketARN, namespace) on fresh customer tables (one without metadata, one with a schema,
  plus the two tables of probe table/round-trip/fu-08-schema-required-metadata-json) returned entries whose
  raw JSON keys were exactly ['createdAt', 'modifiedAt', 'name', 'namespace', 'tableARN', 'type']; the modeled
  TableSummary members namespaceId, tableBucketId and managedByService were absent. GetTable on the same
  tables returned namespaceId and tableBucketId (uuids) but also no managedByService. type was 'customer' in
  both.
  - ACK: list_operation.match_fields, custom_find, is_read_only · ops: ListTables, GetTable · fields:
    namespaceId, tableBucketId, managedByService, type
  - repro: CreateTable x2 -> ListTables(namespace) -> compare raw JSON entry keys with the TableSummary shape
    in the botocore model
  - measurements: n_tables=4, n=4
  - handling: handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-315](table-metadata-location.md#s3t-table-315), [S3T-TABLE-031](table-metadata-location.md#s3t-table-031), [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-136](table-replication-status.md#s3t-table-136), [S3T-TABLE-223](table-replication-status.md#s3t-table-223), [S3T-TABLE-221](table-record-expiration.md#s3t-table-221),
    [S3T-TABLE-177](table-replication.md#s3t-table-177) · evidence: table/round-trip/fu-12-create-vs-get-members,
    table/round-trip/fu-08-schema-required-metadata-json
  - notes: FU-12 / FU-10 side observation. A find-by-list that matches on namespaceId/tableBucketId would
    never match; match on namespace+name or tableARN. Whether managedByService/namespaceId appear for
    'aws'-type (managed) tables is not observable here (no managed table in the lab).

## Response fidelity and consistency

- <a id="s3t-table-068"></a>**S3T-TABLE-068** `eventual-consistency` · impact high · handled · verified 2026-10-13
  **RenameTable visibility at 150 ms polling: both-200 overlap never, both-404 gap never, ARN read always 200 across 5 renames**
  Three pollers (GetTable old triple, new triple, tableArn) at 150 ms around RenameTable(newName) in 5 trials
  (27 samples/trial): RenameTable -> HTTP 204 in [153, 146, 178, 147, 150] ms. First new-triple 200 completed
  [-0.04, -0.04, -0.07, -0.04, -0.05] s relative to the rename call returning (negative = before); first
  old-triple 404 completed [-0.05, -0.04, -0.07, -0.04, -0.05] s relative to the return. Samples where both
  triples returned 200: none; where both returned 404: none. The tableArn read returned ['200'] throughout and
  the name it reported was one of old/new (names seen per trial: [2, 2, 2, 2, 2]).
  - ACK: requeue, is_arn_primary_key · ops: RenameTable, GetTable · fields: name, tableArn
  - repro: Start 3 GetTable pollers at 150 ms (old triple, new triple, tableArn); RenameTable(newName); stop 1
    s later; repeat 5x alternating names
  - measurements: rename_latency_ms=[153, 146, 178, 147, 150], new_visible_rel_rename_end_s=[-0.045, -0.042,
    -0.071, -0.042, -0.046], old_gone_rel_rename_end_s=[-0.047, -0.041, -0.071, -0.041, -0.046],
    both_200_overlap_s=[0, 0, 0, 0, 0], both_404_gap_s=[0, 0, 0, 0, 0], poll_interval_s=0.15, n=5
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122`
  - related: [S3T-TABLE-069](#s3t-table-069), [S3T-TABLE-070](#s3t-table-070), [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006) · hypotheses: H-T-051, H-T-198,
    H-T-199 · evidence: table/consistency-windows/rename-visibility
  - notes: H-T-051 (contrarian: old triple still 200 / new triple 404 after the call returns) refuted at this
    resolution. H-T-198 (sync, atomic, no transitional view) confirmed at this resolution. H-T-199
    (contrarian: both-404 window) refuted at this resolution. Resolution is the 150 ms poll interval;
    windows...
  - full notes: [details/S3T-TABLE-068.md](details/S3T-TABLE-068.md)

- <a id="s3t-table-071"></a>**S3T-TABLE-071** `eventual-consistency` · impact medium · handled · verified 2026-10-13
  **ListTables reflects RenameTable on poll #0/0/0 (first poll right after the call); list modifiedAt equals GetTable modifiedAt on poll #0/0/0**
  After RenameTable (3 trials, trial 0 also moved namespaces), ListTables(bucket) + GetTable(new triple) were
  polled every ~250 ms: the new name was present and the old absent from poll index [0, 0, 0] (0 = first poll,
  issued [0.02, 0.03, 0.02] s after the rename returned); the summary modifiedAt equalled GetTable.modifiedAt
  from poll index [0, 0, 0]. Summary tableARN equals the pre-rename ARN: True. ListTables(namespace=<old ns>)
  after the move -> 200 with 0 tables.
  - ACK: list_operation.match_fields, custom_find · ops: RenameTable, ListTables, GetTable · fields: tables,
    modifiedAt, tableARN, namespace, name
  - repro: RenameTable(newName[,newNamespaceName]); poll ListTables(bucket) and GetTable(new triple) every 250
    ms until stable for 2 polls
  - measurements: first_consistent_poll_index=[0, 0, 0], first_modifiedAt_equal_poll_index=[0, 0, 0],
    poll_interval_s=0.25, n=3
  - handling: handled via `generator.yaml:18-23; generator.yaml:119-122`
  - related: [S3T-TABLE-068](#s3t-table-068), [S3T-TABLE-069](#s3t-table-069), [S3T-TABLE-070](#s3t-table-070), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006), [S3T-TABLE-088](#s3t-table-088), [S3T-TABLE-096](#s3t-table-096),
    [S3T-TABLE-128](#s3t-table-128) · hypotheses: H-T-063, H-T-077 · evidence: table/consistency-windows/rename-visibility
  - notes: H-T-063 (list reflects the rename within 1 s, ARN unchanged) confirmed. H-T-077 (contrarian: list
    modifiedAt lags GetTable by seconds) refuted.

- <a id="s3t-table-088"></a>**S3T-TABLE-088** `eventual-consistency` · impact high · handled · verified 2026-10-13
  **CreateTable is synchronous: GetTable by composite key and by tableArn -> 200 on the first call in 20/20 and 20/20 trials; no status field**
  20 trials, alternating which read goes first: GetTable(tableBucketARN+namespace+name) issued 27-62 ms after
  the CreateTable response returned 200 in 20/20; GetTable(tableArn) issued 34-86 ms after returned 200 in
  20/20. The GetTable body has no key containing 'status' or 'state' (none). CreateTable(no metadata) call
  latency 293-481 ms.
  - ACK: synced.when, none · ops: CreateTable, GetTable · fields: tableArn, tableBucketARN, namespace, name
  - repro: Loop 20x: CreateTable; GetTable composite and GetTable(tableArn) at once (order alternating);
    DeleteTable
  - measurements: after_create_composite_first_call_ms=[32, 62, 33, 53, 33, 51, 32, 50, 37, 46, 27, 46, 39,
    46, 33, 51, 34, 43, 40, 42], after_create_arn_first_call_ms=[41, 54, 45, 46, 86, 42, 44, 43, 52, 40, 38,
    39, 50, 40, 43, 43, 45, 35, 53, 34], create_latency_ms=[481, 395, 348, 330, 319, 332, 332, 293, 305, 357,
    367, 359, 322, 315, 433, 340, 309, 395, 364, 407], n=20, not_visible_windows=[]
  - handling: handled via `templates/hooks/table/sdk_read_one_post_build_request.go.tpl:1-13; pkg/resource/table/sdk.go:79-91; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-292](table-metadata-location.md#s3t-table-292), [S3T-TABLE-283](table-schema.md#s3t-table-283), [S3T-TABLE-291](#s3t-table-291), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-066](table-metadata-location.md#s3t-table-066), [S3T-TABLE-096](#s3t-table-096),
    [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-128](#s3t-table-128) · hypotheses: H-T-026, H-T-027 · evidence:
    table/consistency-windows/create-delete-visibility
  - notes: H-T-026 confirmed. H-T-027 (ARN index lags the name index for 2-5 s) refuted: GetTable(tableArn)
    was visible on the first call in 20/20 trials.

- <a id="s3t-table-096"></a>**S3T-TABLE-096** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables is read-after-write for CreateTable (10/10 first-poll) and DeleteTable (10/10 first-poll) at 200 ms polling**
  10 trials of CreateTable -> ListTables(namespace, prefix=name) every 200 ms until present, then DeleteTable
  -> ListTables until absent: polls needed after create [1, 1, 1, 1, 1, 1, 1, 1, 1, 1], after delete [1, 1, 1,
  1, 1, 1, 1, 1, 1, 1] (1 = visible on the first list issued right after the mutating call returned).
  CreateTable latency ms [345, 354, 356, 327, 402, 344, 569, 428, 344, 349]; DeleteTable latency ms [137, 146,
  156, 147, 141, 141, 143, 127, 159, 137].
  - ACK: custom_find, requeue · ops: CreateTable, DeleteTable, ListTables · fields: tables, prefix
  - repro: loop 10x: CreateTable; ListTables(prefix=name) @200ms until present; DeleteTable; ListTables until
    absent
  - measurements: create_polls_to_visible=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], delete_polls_to_absent=[1, 1, 1, 1,
    1, 1, 1, 1, 1, 1], create_latency_ms=[345, 354, 356, 327, 402, 344, 569, 428, 344, 349],
    delete_latency_ms=[137, 146, 156, 147, 141, 141, 143, 127, 159, 137], poll_interval_s=0.2, n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-088](#s3t-table-088), [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-128](#s3t-table-128) · hypotheses: H-T-079 · evidence:
    table/weird-inputs/list-tables-bounds
  - notes: H-T-079 confirmed (predicted 0 polls of lag in 10/10 trials).

- <a id="s3t-table-097"></a>**S3T-TABLE-097** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables: empty bucket -> 200 tables=[] (key present, no token); deleted or never-created bucket -> NotFoundException 404; S3 ARN 400**
  Fresh empty bucket: ListTables -> HTTP 200, 'tables' key present: True (value []), continuationToken
  present: False; emptied bucket before deletion -> HTTP 200 with 0 tables. Right after DeleteTableBucket
  returned HTTP 204: ListTables -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; final
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  tableBucketARN='arn:aws:s3:::<bucket>' -> BadRequestException (HTTP 400) 'The specified table bucket ARN is
  not valid.'; ARN with another region -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; well-formed ARN of a never-created bucket -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'; 'not-an-arn' -> BadRequestException (HTTP 400) 'The specified table
  bucket ARN is not valid.' (client_side=False).
  - ACK: exceptions.404, terminal_codes, pre-delete-cleanup · ops: ListTables, DeleteTableBucket · fields:
    tables, tableBucketARN
  - repro: CreateTableBucket; ListTables; ... DeleteTableBucket; ListTables;
    ListTables(tableBucketARN='arn:aws:s3:::x'); ListTables(unknown bucket ARN)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-093](#s3t-table-093), [S3T-TABLE-094](#s3t-table-094), [S3T-TABLE-244](#s3t-table-244), [S3T-TABLE-143](#s3t-table-143), [S3T-TABLE-267](#s3t-table-267) · hypotheses: H-T-080 ·
    evidence: table/weird-inputs/list-tables-bounds
  - notes: H-T-080 confirmed (predicted: empty -> 200 with tables=[] present and no token; deleted bucket ->
    NotFoundException 404; plain S3 ARN -> BadRequestException 400). Messages: deleted 'The specified bucket
    does not exist.', plain-s3 'The specified table bucket ARN is not valid.', unknown 'The specified...
  - full notes: [details/S3T-TABLE-097.md](details/S3T-TABLE-097.md)

- <a id="s3t-table-127"></a>**S3T-TABLE-127** `response-fidelity` · impact high · handled · verified 2026-10-13
  **After removing every tag ListTagsForResource -> HTTP 200 body {"tags":{}}; never-tagged table -> {"tags":{}}...**
  UntagResource(all 3 keys) -> HTTP 204. Raw GET /tag/<arn> body afterwards: {"tags":{}} (SDK parsed: tags key
  present True, value {}). Never-tagged table raw body: {"tags":{}} (SDK: present True, value {}).
  TagResource({a:9}) right after the removal -> HTTP 200; ListTagsForResource -> {'a': '9'}.
  - ACK: compare.nil_equals_zero_value, tags.custom-sync · ops: UntagResource, ListTagsForResource,
    TagResource · fields: tags
  - repro: UntagResource(all keys); raw GET /tag/<tableARN>; raw GET /tag/<never-tagged tableARN>;
    TagResource({a:9}); ListTagsForResource
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-156](tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](tablebucket.md#s3t-tablebucket-157), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028),
    [S3T-TABLEBUCKET-121](tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-007](tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125),
    [S3T-TABLE-128](#s3t-table-128), [S3T-TABLE-116](#s3t-table-116), [S3T-TABLE-117](#s3t-table-117) · hypotheses: H-T-203 · evidence:
    table/idempotency/untag-and-post-delete
  - notes: H-T-203 confirmed (predicted: 200 with tags={} present, identical to a never-tagged table; re-add
    -> success and immediate read-back).
  - full notes: [details/S3T-TABLE-127.md](details/S3T-TABLE-127.md)

- <a id="s3t-table-128"></a>**S3T-TABLE-128** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTagsForResource is read-after-write for TagResource/UntagResource on table and bucket ARNs (20/20 first-poll each)**
  10 trials per ARN: TagResource({n:i}) then ListTagsForResource every 100 ms until n==i, UntagResource(['n'])
  then poll until absent. Polls needed (1 = first read): table tag [1, 1, 1, 1, 1, 1, 1, 1, 1, 1] untag [1, 1,
  1, 1, 1, 1, 1, 1, 1, 1]; bucket tag [1, 1, 1, 1, 1, 1, 1, 1, 1, 1] untag [1, 1, 1, 1, 1, 1, 1, 1, 1, 1].
  TagResource latency ms table [100, 108, 103, 108, 105, 99, 99, 128, 97, 105] bucket [47, 44, 52, 52, 60, 46,
  50, 50, 47, 46].
  - ACK: tags.custom-sync · ops: TagResource, UntagResource, ListTagsForResource · fields: tags
  - repro: loop 10x per ARN: TagResource({n:i}); ListTagsForResource @100ms until n==i; UntagResource(['n']);
    ListTagsForResource until absent
  - measurements: table_tag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], table_untag_polls=[1, 1, 1, 1, 1, 1, 1, 1,
    1, 1], bucket_tag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], bucket_untag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    table_tag_latency_ms=[100, 108, 103, 108, 105, 99, 99, 128, 97, 105], bucket_tag_latency_ms=[47, 44, 52,
    52, 60, 46, 50, 50, 47, 46], poll_interval_s=0.1, n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-027](tablebucket.md#s3t-tablebucket-027), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028), [S3T-TABLEBUCKET-007](tablebucket.md#s3t-tablebucket-007), [S3T-TABLEBUCKET-124](tablebucket.md#s3t-tablebucket-124),
    [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032), [S3T-TABLE-088](#s3t-table-088), [S3T-TABLE-096](#s3t-table-096), [S3T-TABLE-071](#s3t-table-071), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125),
    [S3T-TABLE-127](#s3t-table-127), [S3T-TABLE-116](#s3t-table-116), [S3T-TABLE-117](#s3t-table-117) · hypotheses: H-T-094 · evidence:
    table/idempotency/untag-and-post-delete
  - notes: H-T-094 confirmed (predicted 10/10 first-poll for both ARN kinds).

## Tags

- <a id="s3t-table-008"></a>**S3T-TABLE-008** `tag-semantics` · impact medium · handled · verified 2026-10-13
  **Tags are keyed by table ARN and survive a cross-namespace RenameTable; TagResource right after the rename -> HTTP 200**
  ListTagsForResource(tableARN) before the combined move+rename, immediately after and 5 s after return the
  identical map (3 keys). TagResource({c:3}) issued right after -> HTTP 200; ListTagsForResource then -> 4
  keys including c: True.
  - ACK: tags.custom-sync, is_arn_primary_key · ops: RenameTable, ListTagsForResource, TagResource · fields:
    tags, resourceArn
  - repro: CreateTable(tags) -> RenameTable(newNamespaceName, newName) -> ListTagsForResource at t+0, t+5s ->
    TagResource -> ListTagsForResource
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-022](#s3t-table-022), [S3T-TABLE-029](#s3t-table-029), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006), [S3T-TABLE-115](#s3t-table-115), [S3T-TABLE-144](#s3t-table-144) · hypotheses: H-T-052 ·
    evidence: table/identity/rename-identity
  - notes: H-T-052 confirmed.

- <a id="s3t-table-114"></a>**S3T-TABLE-114** `tag-semantics` · impact high · handled · verified 2026-10-13
  **TagResource on a table is an upsert (untouched keys kept), returns HTTP 200; identical re-send -> HTTP 200 with the map unchanged**
  Table tagged {a:1,b:2,env:table}: TagResource({b:9,c:3}) -> HTTP 200; ListTagsForResource -> {'a': '1', 'b':
  '9', 'c': '3', 'env': 'table'}. Re-sending the identical TagResource -> HTTP 200; map afterwards {'a': '1',
  'b': '9', 'c': '3', 'env': 'table'} (unchanged: True). Note TagResource answers HTTP 200 with an empty body
  (the model declares 200; UntagResource declares 204).
  - ACK: tags.custom-sync · ops: TagResource, ListTagsForResource · fields: tags
  - repro: CreateTable(tags {a:1,b:2}) -> TagResource({b:9,c:3}) -> ListTagsForResource ->
    TagResource({b:9,c:3}) -> ListTagsForResource
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](#s3t-table-143), [S3T-TABLEBUCKET-151](tablebucket.md#s3t-tablebucket-151), [S3T-TABLEBUCKET-027](tablebucket.md#s3t-tablebucket-027), [S3T-TABLE-127](#s3t-table-127), [S3T-TABLE-128](#s3t-table-128), [S3T-TABLE-116](#s3t-table-116),
    [S3T-TABLE-117](#s3t-table-117) · hypotheses: H-T-084 · evidence: table/tags/table-tag-semantics
  - notes: H-T-084 confirmed (predicted upsert {a:1,b:9,c:3} and 204 on identical resend; observed HTTP 200 on
    both calls).

- <a id="s3t-table-115"></a>**S3T-TABLE-115** `tag-semantics` · impact medium · handled · verified 2026-10-13
  **Tag stores are per ARN: bucket tags do not appear on the table nor table tags on the bucket; the same key 'env' holds different values**
  Bucket created with {env:bucket, shared:from-bucket}; table with {env:table}. ListTagsForResource(bucket) ->
  {'shared': 'from-bucket', 'env': 'bucket'}; (table) -> {'a': '1', 'b': '9', 'c': '3', 'env': 'table'}.
  TagResource(table,{x:1}) -> HTTP 200 then bucket tags -> {'shared': 'from-bucket', 'env': 'bucket'}.
  TagResource(bucket,{y:1}) -> HTTP 200 then table tags -> {'x': '1', 'a': '1', 'b': '9', 'c': '3', 'env':
  'table'}.
  - ACK: tags.custom-sync · ops: TagResource, ListTagsForResource, CreateTableBucket, CreateTable · fields:
    tags, resourceArn
  - repro: CreateTableBucket(tags env=bucket) -> CreateTable(tags env=table) -> ListTagsForResource both ->
    TagResource(table) -> ListTagsForResource(bucket) -> TagResource(bucket) -> ListTagsForResource(table)
  - handling: handled via `templates/hooks/table_bucket/sdk_read_one_post_set_output.go.tpl:6-12; pkg/resource/table_bucket/hooks.go:51-171; pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-144](#s3t-table-144), [S3T-TABLE-008](#s3t-table-008) · hypotheses: H-T-085 · evidence: table/tags/table-tag-semantics
  - notes: H-T-085 confirmed. Full original title: Tag stores are per ARN: bucket tags do not appear on the
    table, table tags do not appear on the bucket (same key 'env' holds different values)

- <a id="s3t-table-116"></a>**S3T-TABLE-116** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Tag keys are case-sensitive ('Env' and 'env' coexist: True); UntagResource of a non-matching case -> HTTP 204 and removes nothing**
  TagResource({Env:a}) -> HTTP 200, TagResource({env:b}) -> HTTP 200; ListTagsForResource -> {'a': '1', 'b':
  '9', 'c': '3', 'x': '1', 'env': 'b', 'Env': 'a'}. UntagResource(['ENV']) -> HTTP 204, map afterwards {'a':
  '1', 'b': '9', 'c': '3', 'x': '1', 'env': 'b', 'Env': 'a'} (unchanged: True). UntagResource(['Env']) -> HTTP
  204, map {'a': '1', 'b': '9', 'c': '3', 'x': '1', 'env': 'b'}. On a second table tagged {env:a, team:b}:
  UntagResource(['ENV']) -> HTTP 204, map {'env': 'a', 'team': 'b'}; UntagResource(['env']) -> HTTP 204, map
  {'team': 'b'}. GetTable fields changed by the tag/untag calls: {'modifiedAt': ('2026-10-13
  22:34:25.268233+00:00', '2026-10-13 22:34:25.490540+00:00')}.
  - ACK: tags.custom-sync · ops: TagResource, UntagResource, ListTagsForResource, GetTable · fields: tags,
    tagKeys
  - repro: TagResource({Env:a}); TagResource({env:b}); ListTagsForResource; UntagResource(['ENV']);
    UntagResource(['Env']); on t2 {env:a,team:b}: UntagResource(['ENV']) then (['env'])
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-155](tablebucket.md#s3t-tablebucket-155), [S3T-TABLEBUCKET-154](tablebucket.md#s3t-tablebucket-154), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLE-114](#s3t-table-114),
    [S3T-TABLE-127](#s3t-table-127), [S3T-TABLE-128](#s3t-table-128), [S3T-TABLE-117](#s3t-table-117) · hypotheses: H-T-089, H-T-202 · evidence:
    table/tags/table-tag-semantics
  - notes: H-T-089 confirmed. H-T-202 confirmed (predicted: 204 no-op for 'ENV', exact removal for 'env'
    leaving {team:b}).

- <a id="s3t-table-118"></a>**S3T-TABLE-118** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Tag/UntagResource with an 'aws:'-prefixed key (any case) -> BadRequestException 400; mixed untag ['aws:x','env'] rejected whole, env kept**
  TagResource({'aws:foo':x}) -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for
  system use."; {'AWS:foo':x} -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for
  system use."; {'awsfoo':x} (no colon) -> HTTP 200. ListTagsForResource shows keys starting with aws: (any
  case): []. UntagResource(['aws:cloudformation:stack-name']) -> BadRequestException (HTTP 400) "Tags
  beginning with 'aws:' are reserved for system use.". On a table with {env:a}: UntagResource(['aws:x','env'])
  -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for system use.", map afterwards
  {'team': 'b', 'k': 'v2', 'env': 'a', ' k ': 'v', 'e': '', 'w': ' spaced value '}.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, UntagResource, ListTagsForResource · fields:
    tags, tagKeys
  - repro: TagResource({'aws:foo':'x'}); TagResource({'AWS:foo':'x'});
    UntagResource(['aws:cloudformation:stack-name']); UntagResource(['aws:x','env'])
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-152](tablebucket.md#s3t-tablebucket-152), [S3T-TABLEBUCKET-025](tablebucket.md#s3t-tablebucket-025), [S3T-TABLE-119](#s3t-table-119), [S3T-TABLE-117](#s3t-table-117), [S3T-TABLE-126](#s3t-table-126),
    [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-125](#s3t-table-125) · hypotheses: H-T-091, H-T-206 · evidence: table/tags/table-tag-semantics
  - notes: H-T-091 confirmed (predicted BadRequestException 400 for aws: keys, no system tags returned).
    H-T-206 (contrarian: untag of aws: keys is a permissive 204) refuted: BadRequestException (HTTP 400) "Tags
    beginning with 'aws:' are reserved for system use.". Messages: tag "Tags beginning with 'aws:' are...
  - full notes: [details/S3T-TABLE-118.md](details/S3T-TABLE-118.md)

## Delete semantics

- <a id="s3t-table-231"></a>**S3T-TABLE-231** `delete-semantics` · impact medium · partially handled · verified 2026-10-13
  **S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s; plain s3tables DeleteTable removes it (204)**
  s3 DeleteBucketMetadataConfiguration -> HTTP 204, but GetTable(journal tableArn) kept returning 200 for the
  whole 300 s watch (5 s polls) and again 440 s after the configuration was created (ListTables in the
  'b_<bucket>' namespace of the aws-s3 table bucket still listed 'journal'). Customer credentials could then
  remove the AWS-managed table and its container directly: DeleteTable(aws-s3 bucket ARN, namespace,
  'journal') -> HTTP 204 without versionToken, DeleteNamespace -> 204, DeleteTableBucket(aws-s3) -> 204;
  ListTableBuckets(type=aws) afterwards -> empty.
  - ACK: pre-delete-cleanup, docs-only · ops: DeleteTable, DeleteNamespace, DeleteTableBucket, GetTable,
    ListTables · fields: type, managedByService
  - repro: s3 CreateBucketMetadataConfiguration(journal) -> wait ACTIVE -> DeleteBucketMetadataConfiguration
    -> GetTable every 5 s for 300 s -> DeleteTable/DeleteNamespace/DeleteTableBucket on aws-s3
  - measurements: journal_still_present_after_s=440, watch_window_s=300, n=1
  - handling: partially handled via `generator.yaml:45-46; generator.yaml:60-65; apis/v1alpha1/types.go:40; apis/v1alpha1/types.go:83` - see Handling gaps
  - related: [S3T-TABLE-313](table-schema.md#s3t-table-313), [S3T-TABLE-089](table-metadata-location.md#s3t-table-089) · evidence: table/creative/expiration-aws-managed-journal-table
  - notes: Side observation from the record-expiration creative probe, one table. Whether the service would
    have garbage-collected the journal table later is unknown (the lab deleted it). Relevant to any controller
    that creates metadata configurations or counts tables in the account: AWS-managed tables are not...
  - full notes: [details/S3T-TABLE-231.md](details/S3T-TABLE-231.md)

## Quotas and rate limits

- <a id="s3t-table-185"></a>**S3T-TABLE-185** `quota-limit` · impact high · handled · verified 2026-10-13
  **Table tag limit is 50 user tags cumulative, enforced atomically (BadRequestException 400, nothing applied); updating a key at 50 is allowed**
  TagResource with 50 new keys on a table that already had 1 tag (51 cumulative) -> BadRequestException (HTTP 400)
  'The provided tags are invalid. Number of user tags must not exceed 50' (1 keys afterwards); then 49 new
  keys (-> 50 cumulative) -> HTTP 200, 50 keys. With 50 keys on the table: TagResource({new key}) ->
  BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user tags must not exceed 50', 50
  keys afterwards; TagResource({existing key: new value}) -> HTTP 200, value afterwards 'changed'; one
  TagResource mixing an existing-key update with a new key -> BadRequestException (HTTP 400) 'The provided
  tags are invalid. Number of user tags must not exceed 50', 50 keys afterwards, existing key value 'v'.
  TagResource with 51 keys on an untagged table -> BadRequestException (HTTP 400) 'The provided tags are
  invalid. Number of user tags must not exceed 50', 0 keys afterwards. CreateTable with 51 tags ->
  BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user tags must not exceed 50';
  GetTable on that name -> NotFoundException (HTTP 404); ListTables(prefix) found 0 table(s). UntagResource
  with 49 keys in one call -> HTTP 204, 0 non-run keys left.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, UntagResource, ListTagsForResource, CreateTable,
    GetTable · fields: tags, tagKeys
  - repro: CreateTable(1 tag); TagResource(50 keys); TagResource(1 new); TagResource(1 existing);
    TagResource(51 keys) on an untagged table; CreateTable(51 tags)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032),
    [S3T-TABLE-143](#s3t-table-143), [S3T-TABLEBUCKET-148](tablebucket.md#s3t-tablebucket-148), [S3T-TABLEBUCKET-149](tablebucket.md#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](tablebucket.md#s3t-tablebucket-150), [S3T-TABLEBUCKET-031](tablebucket.md#s3t-tablebucket-031),
    [S3T-TABLEBUCKET-025](tablebucket.md#s3t-tablebucket-025), [S3T-TABLE-119](#s3t-table-119), [S3T-TABLE-117](#s3t-table-117), [S3T-TABLE-126](#s3t-table-126), [S3T-TABLE-118](#s3t-table-118) · hypotheses: H-T-087,
    H-T-088 · evidence: table/limits/tag-count-and-read-bursts
  - notes: H-T-087 confirmed: with 50 tags (the harness 'ackq-run' tag counts as a user tag)
    TagResource({new}) -> BadRequestException 400 'The provided tags are invalid. Number of user tags must not
    exceed 50' and nothing applied; TagResource({existing: new value}) -> 200 and the value changed; a single
    call...
  - full notes: [details/S3T-TABLE-185.md](details/S3T-TABLE-185.md)

- <a id="s3t-table-188"></a>**S3T-TABLE-188** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **20 concurrent CreateTable (distinct names, one namespace) all 200 but serialized (wall 9.7 s, latency 0.3-9.7 s); 20 DeleteTable at 4 TPS ok**
  20 CreateTable calls for distinct names in one namespace fired concurrently (20 threads, wall 9.69s) ->
  {"200": 20}. ListTables afterwards listed 20 of the 20 names, none of the rejected names exists. Deleting
  the created tables sequentially at 4 TPS with the no-retry client -> {"204": 20}.
  - ACK: requeue, e2e-timing · ops: CreateTable, ListTables, DeleteTable · fields: name
  - repro: 20 threads x CreateTable(distinct names); retry a rejected one every 250 ms; ListTables;
    DeleteTable each at 4 TPS
  - measurements: burst_wall_s=9.69, create_latency_ms_min_max=[331, 9679], throttle_refill_s=null,
    delete_latency_ms_min_max=[140, 195], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-218](#s3t-table-218), [S3T-TABLE-217](#s3t-table-217), [S3T-TABLE-186](#s3t-table-186) · hypotheses: H-T-045 · evidence:
    table/limits/tag-count-and-read-bursts
  - notes: H-T-045 create half confirmed (no 429/409 for 20 parallel CreateTable) but qualified: the service
    appears to serialize table creation within a namespace/bucket - 20 concurrent creates completed one after
    another (latencies 331-9679 ms, wall 9.69 s) instead of in parallel, so a controller creating...
  - full notes: [details/S3T-TABLE-188.md](details/S3T-TABLE-188.md)

- <a id="s3t-table-218"></a>**S3T-TABLE-218** `quota-limit` · impact high · handled · verified 2026-10-13
  **DeleteTable back-to-back at ~6.5/s: 68 x 204 then 52 x TooManyRequestsException 429 (n=120); 120 CreateTable back-to-back at ~3/s: no 429**
  120 CreateTable (format=ICEBERG, no metadata) issued one after another with no pause across 2 namespaces ->
  {"200": 120} in 39.5s (3.0/s; latency [278, 395] ms). 120 DeleteTable likewise -> {"204": 68,
  "TooManyRequestsException/429": 52} in 10.4s (11.5/s; latency [2, 203] ms) messages
  {"TooManyRequestsException": "Too Many Requests"}.
  - ACK: e2e-timing, requeue · ops: CreateTable, DeleteTable
  - repro: 120 x CreateTable back-to-back; 120 x DeleteTable back-to-back; no-retry client
  - measurements: create_wall_s=39.53, create_latency_ms_min_max=[278, 395], delete_wall_s=10.41,
    delete_latency_ms_min_max=[2, 203], n=240
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-165](tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018),
    [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLE-188](#s3t-table-188), [S3T-TABLE-217](#s3t-table-217), [S3T-TABLE-186](#s3t-table-186) · hypotheses: H-T-071,
    H-N-043 · evidence: table/limits/default-page-120-tables
  - notes: Rate observation (H-T-071 setup, H-N-043 context): 120 CreateTable one after another (278-395 ms
    each, 3.0/s) -> 120 x 200, no 429; 120 DeleteTable one after another (130-200 ms each, ~6.5/s) -> the
    first 67 returned 204 and calls 68-120 all returned TooManyRequestsException 429 'Too Many Requests'...
  - full notes: [details/S3T-TABLE-218.md](details/S3T-TABLE-218.md)

## Cross-region

- <a id="s3t-table-244"></a>**S3T-TABLE-244** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable/RenameTable/DeleteTable/Tag/UntagResource with the other region's ARNs -> 400 region mismatch; nothing changed in either region**
  us-west-2 endpoint with us-east-1 identifiers: TagResource, UntagResource, CreateTable, RenameTable,
  DeleteTable -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'. us-east-1 endpoint with us-west-2 identifiers: TagResource, UntagResource, CreateTable,
  RenameTable, DeleteTable -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.'. Verification via each region's own endpoint afterwards: tables in the namespace
  ['ackq_5d5312_tbl'] / ['ackq_5d5312_tbl'] (no new table, no rename), table tags {'ackq-run': '5d5312',
  'side': 'us-west-2'} / {'ackq-run': '5d5312', 'side': 'us-east-1'} (no foreign tag, 'side' not removed),
  GetTable(tableArn) HTTP 200 / HTTP 200.
  - ACK: terminal_codes, references · ops: CreateTable, RenameTable, DeleteTable, TagResource, UntagResource ·
    fields: tableBucketARN, resourceArn, namespace, name
  - repro: Same-named triple in both regions; from each endpoint CreateTable(new name)/RenameTable/DeleteTable
    with the other region's bucket ARN and Tag/Untag with the other region's tableArn;
    ListTables/ListTags/GetTable in each region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-143](#s3t-table-143), [S3T-TABLE-097](#s3t-table-097), [S3T-TABLE-267](#s3t-table-267) · hypotheses: H-T-022, H-T-075 · evidence:
    table/cross-region/foreign-region-arn-reads-writes
  - notes: H-T-022 confirmed (CreateTable on the us-west-2 endpoint with a real us-east-1 bucket ARN ->
    BadRequestException, no remote create, no local create); H-T-075 write part confirmed
    (TagResource/UntagResource -> 400, no tag landed). State in both regions unchanged: True.
  - full notes: [details/S3T-TABLE-244.md](details/S3T-TABLE-244.md)

## Handling gaps (bugs to file)

- [S3T-TABLE-035](#s3t-table-035) - Raw CreateTable with an extra 'type':'aws' member -> HTTP 400 BadRequestException "Creating
  tables of this type are not allowed."... (partially handled)
  H-T-033 (type cannot be requested): confirmed (rejected). The raw request bypasses the harness evidence log;
  result.yaml raw_create_type_aws has the exact body.
  - handling_ref: `generator.yaml:45-46; generator.yaml:60-65`
- [S3T-TABLE-119](#s3t-table-119) - Table tag bounds: 129-char key / 257-char value -> ValidationException 400 (unmodeled); '|',
  '#', emoji key or '☃' value -> BadRequest 400 (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven: TagResource with a tag key >128
  chars or value >256 chars returns the unmodeled ValidationException (HTTP 400), which is not in
  terminal_codes=[BadRequestException], so a Table whose spec.tags breaks the length bounds is retried as
  recoverable and never reaches a terminal condition (same for UntagResource.tagKeys >128 chars, 126, and a
  malformed tableArn/resourceArn, 267).
  - handling_ref: `generator.yaml:99-104; generator.yaml:170-175`
- [S3T-TABLE-120](#s3t-table-120) - TagResource over-long key/value and empty key return wire code ValidationException (HTTP
  400), which the s3tables model does not declare (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven: the service returns the
  unmodeled wire code ValidationException (HTTP 400) for permanent spec errors - TagResource on a table ARN
  with a key >128 chars, a value >256 chars or an empty key ([S3T-TABLE-120](#s3t-table-120)) and PutTablePolicy with a document >20480
  bytes ([S3T-TABLEPOLICY-001](table-policy.md#s3t-tablepolicy-001)) - which terminal_codes=[BadRequestException] never matches, so a Table whose
  spec.tags break the length bounds (and any future TablePolicy with an oversize document) is requeued as
  recoverable forever instead of reaching a Terminal condition; add ValidationException to terminal_codes
  (botocore stops only the empty key client-side, -120, so the over-long values reach the server) and/or
  enforce the bounds in the CRD.
  - handling_ref: `generator.yaml:99-104; generator.yaml:170-175`
- [S3T-TABLE-231](#s3t-table-231) - S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s;
  plain s3tables DeleteTable removes it (204) (partially handled)
  Side observation from the record-expiration creative probe, one table. Whether the service would have
  garbage-collected the journal table later is unknown (the lab deleted it). Relevant to any controller that
  creates metadata configurations or counts tables in the account: AWS-managed tables are not removed
  synchronously and the aws-s3 bucket appears in ListTableBuckets(type=aws).
  - handling_ref: `generator.yaml:45-46; generator.yaml:60-65; apis/v1alpha1/types.go:40;
    apis/v1alpha1/types.go:83`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-048](#s3t-table-048) | RenameTable to a name that already exists in the target namespace -> ConflictException; same name in another namespace is allowed | rename_collision_latency_ms=[55, 11] | 2 |
| [S3T-TABLE-068](#s3t-table-068) | RenameTable visibility at 150 ms polling: both-200 overlap never, both-404 gap never, ARN read always 200 across 5 renames | rename_latency_ms=[153, 146, 178, 147, 150], new_visible_rel_rename_end_s=[-0.045, -0.042, -0.071, -0.042, -0.046], old_gone_rel_rename_end_s=[-0.047, -0.041, -0.071, -0.041, -0.046], both_200_overlap_s=[0, 0, 0, 0, 0], both_404_gap_s=[0, 0, 0, 0, 0], poll_interval_s=0.15 | 5 |
| [S3T-TABLE-069](#s3t-table-069) | RenameTable immediately followed by the reverse RenameTable: 3/3 second calls -> 204 (no settling wait needed) | first_latency_ms=[142, 133, 144], second_latency_ms=[134, 167, 140] | 3 |
| [S3T-TABLE-070](#s3t-table-070) | Combined RenameTable(newNamespaceName+newName): intermediate triples never resolve at 200 ms polling (2 trials) | rename_both_latency_ms=[200, 131], poll_interval_s=0.2 | 2 |
| [S3T-TABLE-071](#s3t-table-071) | ListTables reflects RenameTable on poll #0/0/0 (first poll right after the call); list modifiedAt equals GetTable modifiedAt on poll #0/0/0 | first_consistent_poll_index=[0, 0, 0], first_modifiedAt_equal_poll_index=[0, 0, 0], poll_interval_s=0.25 | 3 |
| [S3T-TABLE-088](#s3t-table-088) | CreateTable is synchronous: GetTable by composite key and by tableArn -> 200 on the first call in 20/20 and 20/20 trials; no status field | after_create_composite_first_call_ms=[32, 62, 33, 53, 33, 51, 32, 50, 37, 46, 27, 46, 39, 46, 33, 51, 34, 43, 40, 42], after_create_arn_first_call_ms=[41, 54, 45, 46, 86, 42, 44, 43, 52, 40, 38, 39, 50, 40, 43, 43, 45, 35, 53, 34], create_latency_ms=[481, 395, 348, 330, 319, 332, 332, 293, 305, 357, 367, 359, 322, 315, 433, 340, 309, 395, 364, 407], not_visible_windows=[] | 20 |
| [S3T-TABLE-096](#s3t-table-096) | ListTables is read-after-write for CreateTable (10/10 first-poll) and DeleteTable (10/10 first-poll) at 200 ms polling | create_polls_to_visible=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], delete_polls_to_absent=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], create_latency_ms=[345, 354, 356, 327, 402, 344, 569, 428, 344, 349], delete_latency_ms=[137, 146, 156, 147, 141, 141, 143, 127, 159, 137], poll_interval_s=0.2 | 10 |
| [S3T-TABLE-128](#s3t-table-128) | ListTagsForResource is read-after-write for TagResource/UntagResource on table and bucket ARNs (20/20 first-poll each) | table_tag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], table_untag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], bucket_tag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], bucket_untag_polls=[1, 1, 1, 1, 1, 1, 1, 1, 1, 1], table_tag_latency_ms=[100, 108, 103, 108, 105, 99, 99, 128, 97, 105], bucket_tag_latency_ms=[47, 44, 52, 52, 60, 46, 50, 50, 47, 46], poll_interval_s=0.1 | 20 |
| [S3T-TABLE-186](#s3t-table-186) | 10 concurrent TagResource on one table -> 8-9/10 InternalServerErrorException 500 (no 409; 3 rounds), 3-6 s wall; concurrent Untag 9/10 500 | tag_race_wall_s=[6.18, 3.15, 3.16], untag_race_wall_s=4.14, untag_sequential_wall_s=3.08, untag_sequential_latency_ms_min_max=[88, 133] | 70 |
| [S3T-TABLE-188](#s3t-table-188) | 20 concurrent CreateTable (distinct names, one namespace) all 200 but serialized (wall 9.7 s, latency 0.3-9.7 s); 20 DeleteTable at 4 TPS ok | burst_wall_s=9.69, create_latency_ms_min_max=[331, 9679], throttle_refill_s=null, delete_latency_ms_min_max=[140, 195] | 20 |
| [S3T-TABLE-217](#s3t-table-217) | ListTables without maxTables on a bucket with 120 tables returns 120 items with NO continuationToken; maxTables=100 returns 100 + token | create_120_wall_s=39.53, create_latency_ms_min_max=[278, 395] | 120 |
| [S3T-TABLE-218](#s3t-table-218) | DeleteTable back-to-back at ~6.5/s: 68 x 204 then 52 x TooManyRequestsException 429 (n=120); 120 CreateTable back-to-back at ~3/s: no 429 | create_wall_s=39.53, create_latency_ms_min_max=[278, 395], delete_wall_s=10.41, delete_latency_ms_min_max=[2, 203] | 240 |
| [S3T-TABLE-231](#s3t-table-231) | S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s; plain s3tables DeleteTable removes it (204) | journal_still_present_after_s=440, watch_window_s=300 | 1 |
| [S3T-TABLE-305](#s3t-table-305) | ListTables tie order (same name in 4 namespaces) is stable but opaque (not namespace-name/creation order); token binds filters not maxTables | tables=8, walks=4 | 1 |
| [S3T-TABLE-317](#s3t-table-317) | ListTables entries on fresh customer tables omit namespaceId, tableBucketId and managedByService although the model declares them | n_tables=4 | 4 |
| [S3T-TABLE-324](#s3t-table-324) | UntagResource on any nonexistent table ARN (deleted or never-created) -> InternalServerErrorException 500 for 12 min; Tag/ListTags -> 404 | untag_500_latency_ms.min=314, untag_500_latency_ms.max=388, untag_500_latency_ms.n=23, late_samples_after_delete_s=[36, 181, 181], bucket_deleted_samples_age_min=[11.7, 8.3] | 23 |
| [S3T-TABLE-326](#s3t-table-326) | GetTable NotFound message: well-formed missing namespace -> 'table does not exist'; upper-cased/trailing-space -> 'namespace does not exist' | trials_per_cell=3, cells=9 | 27 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table.md` (s3tables) - unverified, thin or stale items worth a future probe; one line each.

- [S3T-TABLE-188](#s3t-table-188) / -218: the 429 refill time after a DeleteTable burst is unmeasured (`throttle_refill_s` null); -218's
  title says ~6.5/s while its behavior records 11.5/s, so the sustained rate at which TooManyRequestsException starts is
  not pinned down.
- [S3T-TABLE-070](#s3t-table-070): combined move+rename atomicity (no intermediate triple) rests on n=2 at 200 ms polling; a finer poll or
  more trials would firm up rule 15.
- [S3T-TABLE-048](#s3t-table-048): the ConflictException collision latency range (11-55 ms) is n=2.
- [S3T-TABLE-050](#s3t-table-050): no measurements; whether a controller that calls RenameTable only on a Spec.Name delta can still hit the
  same-name 409 (e.g. after an out-of-band rename makes `latest.Name` stale) is not probed.
- [S3T-TABLE-113](#s3t-table-113) / -127: no finding checks whether a spec declaring `tags: {}` against the controller's empty-map-to-nil
  read mapping yields a spurious delta (`compare.nil_equals_zero_value`); handling stays "partial".
- [S3T-TABLE-324](#s3t-table-324): the UntagResource 500 on a nonexistent table was observed out to 11.7 min after the delete; no upper
  bound (does it ever become 404?) is known.
- [S3T-TABLE-186](#s3t-table-186): the concurrency rounds ran with 2-3 other probes active, so the 3.15-6.18 s burst wall may include
  cross-probe load; a quiet-account rerun would isolate the per-table serialization.
- [S3T-TABLE-119](#s3t-table-119): the rejected non-ASCII value was 'café ☃'; the letters alone were not isolated, so whether non-ASCII
  letters are accepted in a tag VALUE (they are in a key) is unverified.
- [S3T-TABLE-032](#s3t-table-032) (partial, low): modifiedBy is absent right after create; when it appears (after RenameTable,
  UpdateTableMetadataLocation, a data-plane commit?) was not probed in this part.
- [S3T-TABLE-035](#s3t-table-035) (partial, low): a raw CreateTable with 'type':'aws' -> 400 "Creating tables of this type are not
  allowed."; the CRD exposes Type as read-only status only, so no spec path reaches this error - confirm nothing else
  can set it.
- [S3T-TABLE-217](#s3t-table-217): ListTables with maxTables omitted returned all 120 tables without a token; the real default page size
  (somewhere above 120) is unknown.
- [S3T-TABLE-068](#s3t-table-068): rename visibility (new triple 200 / old triple 404 completing before the call returns) is n=5 at 150 ms
  polling in one region; not repeated in us-east-1.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-029"></a>**S3T-TABLE-029** | identity | low | confirmed | namespaceId/tableBucketId are UUIDs shared by sibling tables and equal to GetNamespace.namespaceId / GetTableBucket.tableBucketId | [S3T-TABLE-022](#s3t-table-022), [S3T-TABLE-006](table-metadata-location.md#s3t-table-006), [S3T-TABLE-008](#s3t-table-008) | - |
| <a id="s3t-table-032"></a>**S3T-TABLE-032** | response-fidelity | low | confirmed | Right after CreateTable: modifiedAt == createdAt; createdBy and ownerAccountId are the caller's 12-digit account id; modifiedBy is ABSENT | [S3T-TABLE-002](table-metadata-location.md#s3t-table-002), [S3T-TABLE-004](table-policy.md#s3t-table-004) | - |
| <a id="s3t-table-035"></a>**S3T-TABLE-035** | request-validation | low | confirmed | Raw CreateTable with an extra 'type':'aws' member -> HTTP 400 BadRequestException "Creating tables of this type are not allowed."... | [S3T-TABLE-047](#s3t-table-047), [S3T-TABLE-079](#s3t-table-079), [S3T-TABLE-078](#s3t-table-078), [S3T-TABLE-080](#s3t-table-080) | - |
| <a id="s3t-table-053"></a>**S3T-TABLE-053** | delete-semantics | low | confirmed | RenameTable on a just-deleted table -> NotFoundException at t+0 and t+2 s; after DeleteTableBucket -> 404 'destination namespace' message | [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-027](#s3t-table-027), [S3T-TABLE-267](#s3t-table-267), [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-080](#s3t-table-080), [S3T-TABLE-094](#s3t-table-094), [S3T-TABLE-049](#s3t-table-049) | - |
| <a id="s3t-table-091"></a>**S3T-TABLE-091** | response-fidelity | low | confirmed (hypothesis refuted) | ListTables orders by table NAME ascending across namespaces (not namespace-then-name); maxTables=1 walk matches; final page has no token | [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-TABLE-293](#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLE-168](#s3t-table-168), [S3T-TABLE-092](#s3t-table-092), [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-305](#s3t-table-305), [S3T-TABLE-095](#s3t-table-095), [S3T-TABLE-217](#s3t-table-217) | - |
| <a id="s3t-table-092"></a>**S3T-TABLE-092** | request-validation | low | confirmed | ListTables maxTables=0 -> SDK ParamValidationError / server HTTP 400 BadRequestException; maxTables=1001 (over model max) -> HTTP 200 | [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-TABLE-293](#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLE-168](#s3t-table-168), [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](#s3t-table-091) | [S3T-TABLE-168](#s3t-table-168) |
| <a id="s3t-table-119"></a>**S3T-TABLE-119** | request-validation | low | confirmed | Table tag bounds: 129-char key / 257-char value -> ValidationException 400 (unmodeled); '\|', '#', emoji key or '☃' value -> BadRequest 400 | [S3T-TABLE-117](#s3t-table-117), [S3T-TABLE-126](#s3t-table-126), [S3T-TABLE-118](#s3t-table-118), [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-125](#s3t-table-125) | - |
| <a id="s3t-table-263"></a>**S3T-TABLE-263** | other | low | confirmed | Doc claim C027 TRUE: Table tag limit is also 50: 51 keys on a table ARN -> BadRequestException 400, 50 (cumulative) -> 200 | [S3T-TABLE-185](#s3t-table-185), [S3T-TABLE-264](#s3t-table-264), [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032), [S3T-TABLE-143](#s3t-table-143) | - |
| <a id="s3t-table-264"></a>**S3T-TABLE-264** | other | low | confirmed | Doc claim C033 TRUE: TagResource/ListTagsForResource/UntagResource accept a table ARN (.../bucket/<b>/table/<uuid>) like a bucket ARN | [S3T-TABLE-113](#s3t-table-113), [S3T-TABLE-114](#s3t-table-114), [S3T-TABLE-125](#s3t-table-125), [S3T-TABLEBUCKET-032](tablebucket.md#s3t-tablebucket-032), [S3T-TABLE-143](#s3t-table-143), [S3T-TABLE-263](#s3t-table-263), [S3T-TABLE-185](#s3t-table-185) | - |
| <a id="s3t-table-267"></a>**S3T-TABLE-267** | identity | low | confirmed | Identifiers are not normalized - upper-cased uuid/prefix/segment or stray whitespace in tableArn/tableBucketARN -> 400, in names -> 404 | [S3T-TABLE-027](#s3t-table-027), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLE-024](#s3t-table-024), [S3T-TABLE-025](#s3t-table-025), [S3T-TABLE-082](#s3t-table-082), [S3T-TABLE-053](#s3t-table-053), [S3T-TABLE-130](#s3t-table-130), [S3T-TABLE-143](#s3t-table-143), [S3T-TABLE-244](#s3t-table-244), [S3T-TABLE-097](#s3t-table-097), [S3T-TABLE-326](#s3t-table-326) | - |
| <a id="s3t-table-270"></a>**S3T-TABLE-270** | other | low | confirmed | Doc claim C078 TRUE: GetTable with no identifiers -> 400; tableArn + composite key -> BadRequestException, not resolved by ARN | [S3T-TABLE-024](#s3t-table-024), [S3T-TABLE-025](#s3t-table-025), [S3T-TABLE-269](table-metadata-location.md#s3t-table-269), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-001](table-metadata-location.md#s3t-table-001) | - |
| <a id="s3t-table-271"></a>**S3T-TABLE-271** | other | low | confirmed | Doc claim C079 TRUE: RenameTable(newName) keeps tableARN, tableBucketId, namespaceId; GetTable(old name) -> NotFoundException | [S3T-TABLE-002](table-metadata-location.md#s3t-table-002), [S3T-TABLE-175](table-replication.md#s3t-table-175), [S3T-TABLE-005](table-metadata-location.md#s3t-table-005), [S3T-TABLE-272](#s3t-table-272), [S3T-TABLE-273](#s3t-table-273), [S3T-TABLE-274](#s3t-table-274), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](#s3t-table-276), [S3T-TABLE-050](#s3t-table-050), [S3T-TABLE-292](table-metadata-location.md#s3t-table-292), [S3T-TABLE-277](table-metadata-location.md#s3t-table-277), [S3T-TABLE-282](table-metadata-location.md#s3t-table-282), [S3T-TABLE-291](#s3t-table-291), [S3T-TABLE-123](table-metadata-location.md#s3t-table-123) | - |
| <a id="s3t-table-272"></a>**S3T-TABLE-272** | other | low | confirmed | Doc claim C080 TRUE: RenameTable(newNamespaceName=existing ns) moves the table; ListTables(old ns) empty; old namespace still exists | [S3T-TABLE-006](table-metadata-location.md#s3t-table-006), [S3T-TABLE-071](#s3t-table-071), [S3T-NAMESPACE-027](namespace.md#s3t-namespace-027), [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-273](#s3t-table-273), [S3T-TABLE-274](#s3t-table-274), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](#s3t-table-276), [S3T-TABLE-050](#s3t-table-050) | - |
| <a id="s3t-table-273"></a>**S3T-TABLE-273** | other | low | confirmed | Doc claim C081 TRUE: RenameTable(newNamespaceName=nonexistent) -> NotFoundException; no namespace is created | [S3T-TABLE-049](#s3t-table-049), [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-272](#s3t-table-272), [S3T-TABLE-274](#s3t-table-274), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](#s3t-table-276), [S3T-TABLE-050](#s3t-table-050) | - |
| <a id="s3t-table-274"></a>**S3T-TABLE-274** | other | low | confirmed | Doc claim C082 TRUE: RenameTable(newName=<existing table in the target namespace>) -> ConflictException; both tables unchanged | [S3T-TABLE-048](#s3t-table-048), [S3T-TABLE-050](#s3t-table-050), [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-272](#s3t-table-272), [S3T-TABLE-273](#s3t-table-273), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-276](#s3t-table-276), [S3T-NAMESPACE-036](namespace.md#s3t-namespace-036), [S3T-NAMESPACE-021](namespace.md#s3t-namespace-021), [S3T-TABLE-064](table-metadata-location.md#s3t-table-064) | - |
| <a id="s3t-table-276"></a>**S3T-TABLE-276** | other | low | confirmed | Doc claim C084 TRUE: RenameTable with neither newName nor newNamespaceName -> BadRequestException, not a successful no-op | [S3T-TABLE-050](#s3t-table-050), [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-272](#s3t-table-272), [S3T-TABLE-273](#s3t-table-273), [S3T-TABLE-274](#s3t-table-274), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-NAMESPACE-036](namespace.md#s3t-namespace-036), [S3T-NAMESPACE-021](namespace.md#s3t-namespace-021), [S3T-TABLE-064](table-metadata-location.md#s3t-table-064) | - |
| <a id="s3t-table-291"></a>**S3T-TABLE-291** | other | low | confirmed | Doc claim C109 TRUE: After DeleteTable: GetTable by name and by tableArn -> 404, ListTables drops it, 2nd DeleteTable -> NotFoundException | [S3T-TABLE-066](table-metadata-location.md#s3t-table-066), [S3T-TABLE-089](table-metadata-location.md#s3t-table-089), [S3T-TABLE-096](#s3t-table-096), [S3T-TABLE-122](table-metadata-location.md#s3t-table-122), [S3T-TABLE-292](table-metadata-location.md#s3t-table-292), [S3T-TABLE-277](table-metadata-location.md#s3t-table-277), [S3T-TABLE-271](#s3t-table-271), [S3T-TABLE-275](table-metadata-location.md#s3t-table-275), [S3T-TABLE-282](table-metadata-location.md#s3t-table-282), [S3T-TABLE-123](table-metadata-location.md#s3t-table-123), [S3T-TABLE-283](table-schema.md#s3t-table-283), [S3T-TABLE-015](table-schema.md#s3t-table-015), [S3T-TABLE-088](#s3t-table-088) | - |
| <a id="s3t-table-293"></a>**S3T-TABLE-293** | other | low | confirmed | Doc claim C039 TRUE: ListTables maxTables=1 -> 1 table + continuationToken; following it returns the rest (name order), last page no token | [S3T-TABLE-168](#s3t-table-168), [S3T-TABLE-091](#s3t-table-091), [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLE-092](#s3t-table-092), [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069) | - |
| <a id="s3t-table-305"></a>**S3T-TABLE-305** | response-fidelity | low | confirmed | ListTables tie order (same name in 4 namespaces) is stable but opaque (not namespace-name/creation order); token binds filters not maxTables | [S3T-TABLE-091](#s3t-table-091), [S3T-TABLE-095](#s3t-table-095), [S3T-TABLE-168](#s3t-table-168), [S3T-TABLE-217](#s3t-table-217) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
