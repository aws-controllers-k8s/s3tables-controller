<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Namespace
_CreateNamespace/GetNamespace/DeleteNamespace/ListNamespaces: minimal Create echo, Get-only identifiers, naming rules, strong consistency, not-empty delete as 400, the ~16 req/s per account-region limiter, no tags or ARN._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `namespace.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the Namespace surface: CreateNamespace, GetNamespace, DeleteNamespace and ListNamespaces, plus
ListTables(namespace), the parent `tableBucketARN` error forms and the authorization codes seen through these calls.
There is no state machine and no status member: CreateNamespace is synchronous and echoes only `{tableBucketARN,
namespace:[name]}`, while the five read-only members (`createdAt`, `createdBy`, `namespaceId`, `ownerAccountId`,
`tableBucketId`) appear only on GetNamespace and in ListNamespaces entries - no ARN, no tags, no state
([S3T-NAMESPACE-002](#s3t-namespace-002), -003, -005, -007).
Identity is the pair (tableBucketARN, name): the same name in two buckets yields two `namespaceId`s, and neither the
`namespaceId` nor a synthesized ARN is accepted as the `namespace` argument (-006). The `namespace` member is a
one-element list in the Create input/output and in the Get/List output but a scalar string in the Get/Delete/ListTables
input (-004). Both writes are synchronous and strongly consistent: Get and List reflected Create and Delete at +0 ms in
20/20 trials, and ListNamespaces entries carry the same six members and values as GetNamespace in name-ascending order
(-018, -005). The one asynchronous-looking failure is the delete of a non-empty namespace, which is a deterministic
BadRequestException 400 until the last table is gone and 204 at +0 ms afterwards (-024, -027).

### Rules

1. Never let DeleteNamespace on a namespace that still holds a table go Terminal: the call returns BadRequestException
   400 "The namespace that you tried to delete is not empty." on 6/6 attempts over 10 s, nothing cascades (GetTable and
   GetNamespace stay 200), the code is never ConflictException, and the identical call returned 204 at +0 ms after
   DeleteTable (20/20) or after RenameTable moved the last table out (10/10) - requeue until the children are gone
   ([S3T-NAMESPACE-024](#s3t-namespace-024), -037, -027).
2. Parallel finalizers hit that 400 by design: DeleteTable||DeleteNamespace fired at the same instant -> DeleteNamespace
   400 "not empty" in 6/8 and the immediate retry 204 in 6/6; DeleteNamespace||DeleteTableBucket -> the bucket delete
   400 in 6/8, retry 204 in 6/6; 0 namespaces and 0 orphan tables remained ([S3T-NAMESPACE-039](#s3t-namespace-039)).
3. CreateTable||DeleteNamespace released together ended as CreateTable NotFoundException | DeleteNamespace 204 in 20/20
   with 0 trials where both succeeded and 0 orphan tables, and CreateTable or RenameTable into a missing namespace ->
   NotFoundException with no implicit namespace created; a fresh bucket has no `default` namespace either
   ([S3T-NAMESPACE-028](#s3t-namespace-028), -026, -001).
4. One BadRequestException 400 code carries unrelated meanings: every name rule violation (one shared message "The
   specified namespace name isn't valid..."), multipart/duplicate/empty/null `namespace` lists, a malformed or
   wrong-region parent ARN, a `tableBucketId` passed as `tableBucketARN`, an over-long List prefix, and the "not empty"
   delete - classify by message before treating it as terminal ([S3T-NAMESPACE-011](#s3t-namespace-011), -012, -013, -014, -015, -016, -029,
   -030, -003, -008, -024).
5. NotFoundException also has two meanings: GetNamespace/CreateNamespace/DeleteNamespace on a well-formed ARN of a
   missing bucket -> 404 "The specified bucket does not exist.", on a missing namespace -> 404 "The specified namespace
   does not exist."; neither message names the namespace, and a second DeleteNamespace (or a never-existing name) is
   404, not 204 ([S3T-NAMESPACE-029](#s3t-namespace-029), -022).
6. CreateNamespace has no idempotent 200: a repeat with the same name -> ConflictException 409 "A namespace with an
   identical name already exists in the bucket." on 6/6 over 10 s with the existing namespace untouched (same
   `namespaceId`, `createdAt`), and 5 concurrent creates of one fresh name yield exactly one 200 and four 409 in 6/6
   trials - resolve a 409 by reading back (tableBucketARN, name) ([S3T-NAMESPACE-021](#s3t-namespace-021), -034, -036).
7. A changed `namespaceId` under an unchanged (bucket, name) means the namespace was deleted and re-created: the
   immediate CreateNamespace after DeleteNamespace returned 200 in 5/5 with no cool-down, a new `namespaceId`, a newer
   `createdAt` and the same `tableBucketId` ([S3T-NAMESPACE-023](#s3t-namespace-023)).
8. Parent ARN forms decide the code before any lookup: plain bucket name, general-purpose S3 ARN, uppercased bucket or
   table-ARN form -> BadRequestException 400; well-formed ARN of a missing bucket -> NotFoundException 404; another
   account's ARN -> ForbiddenException 403 "Unauthorized"; a bucket's ARN through a client in another region, or the
   region swapped inside the ARN -> 400 "The provided Region does not match the Region of the resource." and nothing is
   created cross-region ([S3T-NAMESPACE-029](#s3t-namespace-029), -030).
9. An explicit Deny (bucket policy or IAM session policy) returns AccessDeniedException 403 with the IAM-style message
   on all six ops - GetNamespace, ListNamespaces, CreateNamespace, DeleteNamespace, GetTableBucket, ListTables - although
   CreateNamespace, DeleteNamespace and ListTables do not declare that code in the model; ForbiddenException appears only
   for the foreign-account ARN ([S3T-NAMESPACE-031](#s3t-namespace-031), -029).
10. Namespaces are not taggable: TagResource/ListTagsForResource on `bucket/<b>/namespace/<ns>` or `.../namespaces/<ns>`
    -> BadRequestException 400 "The specified resourceArn is not valid.", and bucket tags never appear in GetNamespace or
    ListNamespaces ([S3T-NAMESPACE-007](#s3t-namespace-007)).
11. No read path returns the parent ARN: Get and List entries carry `tableBucketId` only, and `tableBucketId` is not
    accepted as `tableBucketARN` (400); `createdBy` is the 12-digit account id even when the caller is an assumed role
    ([S3T-NAMESPACE-003](#s3t-namespace-003), -002).
12. Naming accepted by CreateNamespace: `[0-9a-z_]`, 1-255 chars, trailing `_`, `__` inside, a leading digit; rejected:
    256 chars, a leading `_`, `_` alone, uppercase, hyphen, period, space, slash, non-ASCII, and the tested reserved
    prefixes `aws`, `awsfoo`, `aws_data`, `Aws_data`, while `aws` later in the name (`<p>_my_aws_data`) is accepted
    ([S3T-NAMESPACE-014](#s3t-namespace-014), -012, -013, -016, -015).
13. ListNamespaces: `prefix` is literal, case-sensitive and left-anchored (`''` and invalid-character prefixes -> 200
    empty, >255 chars -> 400); `maxNamespaces` 0/-1 -> 400 but 1001 -> 200; the last page omits `continuationToken`;
    a garbage, empty or other-bucket token -> 400 "The continuation token is no longer valid."; the token is a live
    name-ordered cursor (names created after it appear, deleted ones vanish; n=1); ListTables(namespace=<missing>) -> 404
    while an empty namespace -> 200 `{"tables": []}` ([S3T-NAMESPACE-008](#s3t-namespace-008), -032, -038, -009, -010, -025).
14. Do not poll for read-after-write: GetNamespace 200 and ListNamespaces present at +0 ms in 20/20 after Create,
    GetNamespace 404 and List absent at +0 ms in 20/20 after Delete; CreateNamespace at +0 ms after CreateTableBucket
    and CreateTable at +0 ms after CreateNamespace were 200 in 20/20; reads concurrent with a delete+recreate of one name
    saw only 200/NotFoundException, never ConflictException ([S3T-NAMESPACE-018](#s3t-namespace-018), -019, -020).
15. CreateNamespace is throttled by a per account-region limiter of ~16 back-to-back calls (16, sometimes 17; 11-22
    across 15 drains) shared by all table buckets of the region and not counted against DeleteNamespace (which has its
    own: 16/30 and 16/31 deletes before 429), CreateTable or TagResource: 16/30 and 14-16/20 sequential creates
    succeeded before TooManyRequestsException 429 "Too Many Requests" with no Retry-After header; retrying every 250 ms
    succeeded after 0.54 s; 13 creates at 4/s -> 12/13 ok in 3/3 and 106 paced at <= 5/s -> 0 x 429; 30 concurrent Gets
    were all 200 twice ([S3T-NAMESPACE-033](#s3t-namespace-033), -035, -040).

### Timing ranges

- CreateNamespace latency: 30-74 ms, n=20 ([S3T-NAMESPACE-018](#s3t-namespace-018)); 49-88 ms, n=20 when fired at +0 ms after
  CreateTableBucket (-019); 29-51 ms, n=5 for the immediate re-create of a just-deleted name (-023); the duplicate-name
  409 answers in 10-21 ms, n=6 (-021); inside back-to-back bursts 2-37 ms over 30 calls (-033) and 2-40 ms over 20 calls
  (-035).
- GetNamespace after Create: 10-13 ms, n=20 (-018); 30 concurrent Gets on one namespace 11-156 ms then 10-78 ms in two
  bursts (wall 0.16 s and 0.09 s, n=60 calls, all 200) (-033). ListNamespaces: 9-49 ms, n=20 (-018).
- DeleteNamespace latency: 23-29 ms, n=20 (-018); 2-52 ms over 30 back-to-back calls (-033); the "not empty"
  BadRequestException answers in 11-59 ms, n=6 (-024).
- Read-after-write lag: 0 ms in every direction - Get 200 and List present in 20/20 at +0 ms after Create, Get 404 and
  List absent in 20/20 at +0 ms after Delete (lag fields recorded as not measurable, -018); `createdAt` minus the client
  clock -0.023 to -0.019 s, n=20 (-018); re-created `createdAt` is 0.068-0.094 s newer than the deleted one, n=5 (-023).
- Parent/child chain at +0 ms, n=20 (-019): CreateTableBucket 1444-2931 ms, CreateNamespace 49-88 ms, CreateTable
  310-440 ms, DeleteTable 139-187 ms, DeleteTableBucket 445-523 ms. Before an immediately-successful DeleteNamespace:
  DeleteTable 130-167 ms, n=20, or RenameTable 155-202 ms, n=10 (-027).
- Delete of a non-empty namespace: 400 on 6/6 attempts spread over 10 s (2 s apart, -024); the retry after the child
  delete returned was 204 at +0 ms in 30/30 sequential trials (-027) and 6/6 immediate retries in the parallel-finalizer
  trials (-039) - there is no settle window to wait on.
- Throttle window: first 429 at index 14-16 of a back-to-back CreateNamespace run and at index 16 of a DeleteNamespace
  run (-035), 16/30 for both ops at ~30 calls/s with wall 0.51-0.57 s for 30 calls (-033); accepted-before-429 across 15
  drains 11-22, mode 16 (-040). Refill: a throttled create and a throttled delete each succeeded 0.54 s later when
  retried every 250 ms (one series each, -035; not a range); after a 5 s pause 16/20 creates passed (-035).
- Same-name create race: 5 concurrent calls complete in 0.11-0.15 s wall, n=6 trials (-034). Delete+recreate under
  concurrent reads: 120 reads over 10 trials, codes 200/NotFoundException only (-020).
- Not a range: assume-role success 9.1 s after CreateRole, n=1 (-031) - an e2e budget hint only.

### Handling gaps

- [S3T-NAMESPACE-024](#s3t-namespace-024) (suspect-bug, not yet handled): today Namespace lists `terminal_codes: [BadRequestException]`
  (generator.yaml:174-175), `sdkDelete` is a plain DeleteNamespace with no pre/post hook and returns the AWS error as-is
  (pkg/resource/namespace/sdk.go:220-238), and `onError` converts it into the Terminal condition and `ackerr.Terminal`
  (pkg/resource/namespace/manager.go:358-378). Evidence: a Namespace holding a table answers BadRequestException 400 "The
  namespace that you tried to delete is not empty." deterministically (6/6 over 10 s), never ConflictException, and
  nothing cascades (-024, -037); DeleteTable||DeleteNamespace fired together hit that 400 in 6/8 trials with the
  immediate retry 204 in 6/6 (-039); after the last DeleteTable or RenameTable the identical call was 204 at +0 ms in
  30/30 (-027). So a Namespace CR deleted together with its Table CRs goes Terminal on the first attempt while the child
  finalizer is still in flight. Fix direction: requeue this message (match code + message), or hold the finalizer until
  the child Table CRs are gone.
- No other finding of this stem is marked suspect-bug. The high-impact `unhandled` items are the limiter (-035, -040):
  TooManyRequestsException is not in `terminal_codes` (generator.yaml:174-175) and no hook paces CreateNamespace, so the
  429 is retried as a recoverable error; the evidence shows the 429 only past ~16 back-to-back creates per
  account-region (0 x 429 at <= 5 calls/s, 12/13 at 4/s right after a drain, refill 0.54 s, no Retry-After), so it
  concerns fan-outs of many Namespace CRs in one region, not a single CR. Strong consistency (-018) is `unhandled`
  because nothing needs handling.

### Where next

- `tablebucket.md` carries the parent side: DeleteTableBucket "not empty" 400 with an empty namespace inside and the
  identical retry 204 (-039 set 2, -037), the parent-ARN forms that fail before any namespace lookup (-029), and the
  bucket's post-delete name reservation described in that document.
- `table.md` carries the child side: CreateTable/RenameTable into a missing namespace (-026), the CreateTable vs
  DeleteNamespace race (-028), ListTables by namespace (-025), duplicate CreateTable 409 (-036), and the table-ARN tag
  behavior that namespaces lack (-007).
- `tablebucket-policy.md` and `table-policy.md` carry the AccessDeniedException vs ForbiddenException split and the
  Deny propagation windows; this document only records that all six namespace-path ops return the IAM-style
  AccessDeniedException under an explicit Deny (-031, -029).
- Long catalogues render in `details/<id>.md` rather than inline: the parent-ARN form matrix for four ops (-029), the
  naming accept/reject matrices (-011, -012, -013, -014, -015, -016), the prefix matrix (-008), the continuation-token
  matrix (-009), the limiter-sharing patterns per op/bucket/region (-040), the wire error-code matrix (-017).
- Open experiments (STATUS.md): none of the three Phase 8 reverify corrections touched this stem, and no namespace
  probe is running or blocked. Items worth a future probe are listed in `namespace.open-questions.md`: the live-cursor
  token (-010, n=1, confidence medium), the concurrent-reads trial (-020, confidence medium), the single-series 0.54 s
  refill and the "fixed window rather than token bucket" reading (-035, -040), and the n=1 IAM propagation hint (-031).
  The lab's one undeletable leftover (bucket `ackq-9ba36e-polnorm` with namespace `ackq_9ba36e_ns` and table
  `ackq_9ba36e_t`, blocked by a Deny policy for every principal) is a `tablebucket-policy.md` matter
  ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007)), not a namespace behavior.
<!-- preserved:end -->

## At a glance

- canonical findings: 40 (high 14 / medium 18 / low 8); duplicates folded into the appendix: 1
- handling: handled 20 · partial 0 · tracked 0 · unhandled 19 · suspect-bug 1 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: request-validation 8, error-code 5, eventual-consistency 4, response-fidelity 4,
  delete-semantics 3, idempotency 3, other 3, quota-limit 3, identity 2, cross-region 1, prerequisite 1,
  read-gap 1, shape-mismatch 1, tag-semantics 1

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
| ListNamespaces | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

No status-like enums in the model for these resources.

No high/medium state-machine findings.

## Field matrix

C = accepted by the create input (CreateNamespace), U = by the update input, R = present in the read output.

| leaf | C | U | R | type |
| --- | --- | --- | --- | --- |
| createdAt | - | - | x | timestamp |
| createdBy | - | - | x | string |
| namespace | x | - | x | list<string> |
| namespaceId | - | - | x | string |
| ownerAccountId | - | - | x | string |
| tableBucketARN | x | - | - | string |
| tableBucketId | - | - | x | string |

## Identity and lookup

- <a id="s3t-namespace-006"></a>**S3T-NAMESPACE-006** `identity` · impact medium · handled · verified 2026-10-13
  **Namespace identity is (tableBucketARN, name): same name in 2 buckets -> 2 namespaceIds; namespaceId or ARN as namespace -> NotFoundException**
  CreateNamespace(same name) in two buckets -> HTTP 200 and HTTP 200; namespaceIds are 36-char uuid and
  36-char uuid (different). GetNamespace(namespace=<namespaceId>) -> NotFoundException (HTTP 404) 'The
  specified namespace does not exist.'; DeleteNamespace(namespace=<namespaceId>) -> NotFoundException (HTTP 404)
  'The specified namespace does not exist.'; GetNamespace(namespace='arn:...:bucket/<b>/namespace/<name>') ->
  NotFoundException (HTTP 404) 'The specified namespace does not exist.'.
  - ACK: is_arn_primary_key, custom_find, references · ops: CreateNamespace, GetNamespace, DeleteNamespace ·
    fields: namespace, namespaceId, tableBucketARN
  - repro: Two buckets; CreateNamespace([same]) in each; GetNamespace both; GetNamespace/DeleteNamespace with
    the namespaceId string and a synthesized ARN as namespace
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-007](#s3t-namespace-007), [S3T-NAMESPACE-022](#s3t-namespace-022) · hypotheses: H-N-027
    · evidence: namespace/round-trip/create-get-list-shape
  - notes: H-N-027 confirmed: no ARN form exists for a namespace; namespaceId is not accepted as a lookup key
    (NotFoundException); the two namespaces have distinct uuid namespaceIds.
  - full notes: [details/S3T-NAMESPACE-006.md](details/S3T-NAMESPACE-006.md)

- <a id="s3t-namespace-023"></a>**S3T-NAMESPACE-023** `identity` · impact medium · handled · verified 2026-10-13
  **CreateNamespace at +0 ms after DeleteNamespace of the same name -> 200 in 5/5; new namespaceId, newer createdAt, same tableBucketId**
  5 delete->recreate rounds on one name: the immediate CreateNamespace returned ['HTTP 200'] (latency {'min':
  29, 'max': 51, 'n': 5} ms, no 409/cool-down); each recreated namespace had a different namespaceId (5/5), a
  newer createdAt (5/5, delta {'min': 0.068, 'max': 0.094, 'n': 5} s) and the same tableBucketId (5/5).
  - ACK: custom_find, none · ops: DeleteNamespace, CreateNamespace, GetNamespace · fields: namespaceId,
    createdAt, tableBucketId
  - repro: CreateNamespace -> GetNamespace -> 5x (DeleteNamespace -> CreateNamespace immediately ->
    GetNamespace)
  - measurements: recreate_latency_ms.min=29, recreate_latency_ms.max=51, recreate_latency_ms.n=5, n=5,
    immediate_200_rounds=5
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-NAMESPACE-006](#s3t-namespace-006), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-007](#s3t-namespace-007), [S3T-NAMESPACE-022](#s3t-namespace-022), [S3T-NAMESPACE-021](#s3t-namespace-021),
    [S3T-NAMESPACE-034](#s3t-namespace-034), [S3T-NAMESPACE-018](#s3t-namespace-018), [S3T-NAMESPACE-019](#s3t-namespace-019), [S3T-NAMESPACE-020](#s3t-namespace-020), [S3T-NAMESPACE-027](#s3t-namespace-027),
    [S3T-NAMESPACE-010](#s3t-namespace-010) · hypotheses: H-N-014 · evidence: namespace/idempotency/duplicate-create-double-delete
  - notes: H-N-014 confirmed: namespaceId is not stable across delete/recreate, so (tableBucketARN, name) is
    the only durable key.

## Idempotency

- <a id="s3t-namespace-021"></a>**S3T-NAMESPACE-021** `idempotency` · impact high · handled · verified 2026-10-13
  **Second CreateNamespace with the same name -> ConflictException (HTTP 409) on 6/6 attempts over 10 s; existing namespace untouched**
  CreateNamespace(name) -> HTTP 200; repeating the identical call 6 times at 2 s intervals returned
  ConflictException with message(s) ['A namespace with an identical name already exists in the bucket.'] every
  time (latency {'min': 10, 'max': 21, 'n': 6} ms). GetNamespace before and after shows an identical body
  (namespaceId, createdAt unchanged: True). There is no idempotent 200.
  - ACK: custom_create, terminal_codes, custom_find · ops: CreateNamespace, GetNamespace · fields: namespace
  - repro: CreateNamespace([n]) -> GetNamespace -> CreateNamespace([n]) x6 at 2 s intervals -> GetNamespace
  - measurements: duplicate_create_latency_ms.min=10, duplicate_create_latency_ms.max=21,
    duplicate_create_latency_ms.n=6, attempts=6, n=6
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-NAMESPACE-036](#s3t-namespace-036), [S3T-TABLE-274](table.md#s3t-table-274), [S3T-TABLE-276](table.md#s3t-table-276), [S3T-TABLE-064](table-metadata-location.md#s3t-table-064), [S3T-TABLE-050](table.md#s3t-table-050), [S3T-NAMESPACE-034](#s3t-namespace-034),
    [S3T-NAMESPACE-022](#s3t-namespace-022), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-024](#s3t-namespace-024), [S3T-NAMESPACE-039](#s3t-namespace-039), [S3T-NAMESPACE-027](#s3t-namespace-027),
    [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019) · hypotheses: H-N-009, H-N-010 · evidence:
    namespace/idempotency/duplicate-create-double-delete
  - notes: H-N-009 confirmed. H-N-010 duplicate-create half confirmed: the 409 is deterministic, not
    transient, despite the modeled shape text 'You can retry the request'; the delete-non-empty 409 message is
    compared in dependencies/delete-with-tables.

- <a id="s3t-namespace-022"></a>**S3T-NAMESPACE-022** `idempotency` · impact high · handled · verified 2026-10-13
  **DeleteNamespace of an already-deleted or never-existing namespace -> NotFoundException (HTTP 404); message does not name the namespace**
  CreateNamespace -> DeleteNamespace -> HTTP 204; DeleteNamespace again at +0 ms -> NotFoundException (HTTP 404)
  'The specified namespace does not exist.'; 2 s later -> NotFoundException (HTTP 404) 'The specified
  namespace does not exist.'; DeleteNamespace of a name that never existed -> NotFoundException (HTTP 404)
  'The specified namespace does not exist.'. The 404 message is identical to GetNamespace's and does not
  include the namespace name.
  - ACK: exceptions.404, custom_delete · ops: DeleteNamespace · fields: namespace
  - repro: CreateNamespace -> DeleteNamespace -> DeleteNamespace (same name) -> DeleteNamespace(never_existed)
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131; pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-006](#s3t-namespace-006), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-007](#s3t-namespace-007), [S3T-NAMESPACE-021](#s3t-namespace-021),
    [S3T-NAMESPACE-034](#s3t-namespace-034) · hypotheses: H-N-012, H-N-013 · evidence:
    namespace/idempotency/duplicate-create-double-delete
  - notes: H-N-012 confirmed; H-N-013 (contrarian: idempotent 204) refuted. A finalizer must treat
    NotFoundException from DeleteNamespace as success.

- <a id="s3t-namespace-034"></a>**S3T-NAMESPACE-034** `idempotency` · impact medium · handled · verified 2026-10-13
  **5 concurrent CreateNamespace for the same fresh name x6 trials: {"200": 6, "ConflictException/409": 24}; exactly one 200 per trial**
  Six trials, each firing 5 CreateNamespace calls for one fresh name concurrently (5 threads, wall
  0.11-0.15s): aggregate codes {"200": 6, "ConflictException/409": 24} (per trial: [{'200': 1,
  'ConflictException/409': 4}, {'ConflictException/409': 4, '200': 1}, {'ConflictException/409': 4, '200': 1},
  {'ConflictException/409': 4, '200': 1}, {'200': 1, 'ConflictException/409': 4}, {'200': 1,
  'ConflictException/409': 4}]). Error messages: {"ConflictException": "A namespace with an identical name
  already exists in the bucket."}. After each trial GetNamespace -> ['HTTP 200'] and ListNamespaces(prefix)
  counted [1] entry.
  - ACK: custom_create, requeue · ops: CreateNamespace, GetNamespace, ListNamespaces · fields: namespace
  - repro: 5 threads x CreateNamespace(same fresh name) at once; GetNamespace; ListNamespaces(prefix); repeat
    6 times
  - measurements: race_wall_s=[0.12, 0.11, 0.14, 0.14, 0.15, 0.12], n=6
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-021](#s3t-namespace-021), [S3T-NAMESPACE-022](#s3t-namespace-022), [S3T-NAMESPACE-023](#s3t-namespace-023) · hypotheses: H-N-044 · evidence:
    namespace/limits/pagination-and-bursts
  - notes: H-N-044 confirmed (predicted exactly one 200 + four ConflictException per trial, never 500/400,
    single namespaceId).

## Errors

- <a id="s3t-namespace-017"></a>**S3T-NAMESPACE-017** `error-code` · impact medium · handled · verified 2026-10-13
  **Wire error codes match the modeled shape names (BadRequestException, ConflictException, NotFoundException)...**
  Across 3 distinct error codes seen in this probe, botocore Error.Code vs wire: BadRequestException -> HTTP
  400, header 'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body
  {"message": "Multipart namespaces are not supported."}; ConflictException -> HTTP 409, header
  'ConflictException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body {"message": "A
  namespace with an identical name already exists in the bucket."}; NotFoundException -> HTTP 404, header
  'NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body {"message": "The
  specified namespace does not exist."}. No NoSuchBucket/NoSuchKey style codes appeared.
  - ACK: terminal_codes, exceptions.404 · ops: CreateNamespace, GetNamespace, DeleteNamespace, ListNamespaces
  - repro: Force BadRequestException (bad name), ConflictException (duplicate create), NotFoundException
    (missing namespace); read the raw HTTP response
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-029](#s3t-namespace-029), [S3T-NAMESPACE-030](#s3t-namespace-030), [S3T-NAMESPACE-031](#s3t-namespace-031), [S3T-NAMESPACE-003](#s3t-namespace-003) · hypotheses: H-N-030
    · evidence: namespace/weird-inputs/naming-edge-cases
  - notes: H-N-030 confirmed: the HTTP header x-amzn-ErrorType is
    '<ShapeName>:http://internal.amazon.com/coral/...' and botocore strips the suffix; the JSON body carries
    only 'message'.
  - full notes: [details/S3T-NAMESPACE-017.md](details/S3T-NAMESPACE-017.md)

- <a id="s3t-namespace-025"></a>**S3T-NAMESPACE-025** `error-code` · impact medium · handled · verified 2026-10-13
  **ListTables(namespace=<missing>) -> NotFoundException; ListTables(namespace=<existing, empty>) -> HTTP 200 {"tables": []}...**
  Bucket with tables ns1/t1 and ns_b/t2 plus an empty namespace: ListTables(namespace=ns1) -> HTTP 200 with 1
  table(s) [('ackq_54dbb7_with_tbl', 't1')] (summary members ['createdAt', 'modifiedAt', 'name', 'namespace',
  'tableARN', 'type']; namespace is a one-element list). ListTables() -> 2 tables across both namespaces.
  ListTables(namespace=<does_not_exist>) -> NotFoundException (HTTP 404) 'The specified namespace does not
  exist.'. ListTables(namespace=<empty namespace>) -> HTTP 200 with raw body '{"tables":[]}' (member present,
  empty). ListTables(prefix=<lab>_t) -> HTTP 200 2 tables (prefix matches the table name across namespaces).
  - ACK: custom_find, exceptions.404, pre-delete-cleanup · ops: ListTables · fields: namespace, tables, prefix
  - repro: CreateNamespace x3, CreateTable in two of them -> ListTables(namespace=...) for an existing, an
    empty and a missing namespace; ListTables() ; ListTables(prefix)
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131`
  - related: [S3T-NAMESPACE-002](#s3t-namespace-002), [S3T-NAMESPACE-041](#s3t-namespace-041), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-004](#s3t-namespace-004),
    [S3T-NAMESPACE-001](#s3t-namespace-001), [S3T-NAMESPACE-026](#s3t-namespace-026), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-029](#s3t-namespace-029) · hypotheses: H-N-035, H-N-036
    · evidence: namespace/dependencies/delete-with-tables
  - notes: H-N-035 confirmed; H-N-036 (contrarian: 200 []) refuted. An 'is the namespace empty' precheck gets
    a 404 for a missing namespace and 200 [] for an empty one.
  - full notes: [details/S3T-NAMESPACE-025.md](details/S3T-NAMESPACE-025.md)

- <a id="s3t-namespace-029"></a>**S3T-NAMESPACE-029** `error-code` · impact high · handled · verified 2026-10-13
  **Parent tableBucketARN: malformed/wrong region -> BadRequestException; missing bucket -> NotFoundException; other account -> Forbidden 403**
  Per ARN form, outcomes of GetNamespace / ListNamespaces / CreateNamespace / GetTableBucket: [bucket name,
  not an ARN] GetNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  GetTableBucket=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'; [well-formed
  ARN, bucket does not exist] GetNamespace=NotFoundException (HTTP 404) 'The specified bucket does not exist.'
  | ListNamespaces=NotFoundException (HTTP 404) 'The specified bucket does not exist.' |
  CreateNamespace=NotFoundException (HTTP 404) 'The specified bucket does not exist.' |
  GetTableBucket=NotFoundException (HTTP 404) 'The specified bucket does not exist.'; [real ARN with region
  swapped] GetNamespace=BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.' | ListNamespaces=BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.' | CreateNamespace=BadRequestException (HTTP 400) 'The provided Region does not match the
  Region of the resource.' | GetTableBucket=BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; [real ARN with bucket name uppercased] GetNamespace=BadRequestException (HTTP 400)
  'The specified bucket name isn't valid. Specify a different bucket name, and then try again.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different
  bucket name, and then try again.' | CreateNamespace=BadRequestException (HTTP 400) 'The specified bucket
  name isn't valid. Specify a different bucket name, and then try again.' | GetTableBucket=BadRequestException
  (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and then try again.';
  [well-formed ARN, other account (doc sample id)] GetNamespace=ForbiddenException (HTTP 403) 'Unauthorized' |
  ListNamespaces=ForbiddenException (HTTP 403) 'Unauthorized' | CreateNamespace=ForbiddenException (HTTP 403)
  'Unauthorized' | GetTableBucket=ForbiddenException (HTTP 403) 'Unauthorized'; [general-purpose S3 bucket
  ARN] GetNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  GetTableBucket=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'; [table ARN
  form under the bucket] GetNamespace=BadRequestException (HTTP 400) 'The specified bucket name isn't valid.
  Specify a different bucket name, and then try again.' | ListNamespaces=BadRequestException (HTTP 400) 'The
  specified bucket name isn't valid. Specify a different bucket name, and then try again.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different
  bucket name, and then try again.' | GetTableBucket=BadRequestException (HTTP 400) 'The specified bucket name
  isn't valid. Specify a different bucket name, and then try again.'. GetNamespace(real bucket, missing
  namespace) -> NotFoundException (HTTP 404) 'The specified namespace does not exist.'; GetNamespace(missing
  bucket, missing namespace) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  CreateNamespace(missing bucket) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  DeleteNamespace(missing bucket) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: terminal_codes, exceptions.404, references · ops: GetNamespace, ListNamespaces, CreateNamespace,
    DeleteNamespace, GetTableBucket · fields: tableBucketARN
  - repro: GetNamespace/ListNamespaces/CreateNamespace/GetTableBucket with each tableBucketARN form in
    behavior
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-162](tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163), [S3T-TABLEBUCKET-164](tablebucket.md#s3t-tablebucket-164), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-127](tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-NAMESPACE-017](#s3t-namespace-017), [S3T-NAMESPACE-030](#s3t-namespace-030), [S3T-NAMESPACE-031](#s3t-namespace-031),
    [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-001](#s3t-namespace-001), [S3T-NAMESPACE-025](#s3t-namespace-025), [S3T-NAMESPACE-026](#s3t-namespace-026), [S3T-NAMESPACE-028](#s3t-namespace-028) ·
    hypotheses: H-N-028, H-N-029 · evidence: namespace/error-taxonomy/parent-arn-iam-deny-cross-region
  - notes: H-N-028 qualified (see per-form codes). H-N-029 confirmed: missing-namespace and missing-bucket are
    distinguishable by message ('The specified namespace does not exist.' vs 'The specified bucket does not
    exist.').
  - full notes: [details/S3T-NAMESPACE-029.md](details/S3T-NAMESPACE-029.md)

- <a id="s3t-namespace-031"></a>**S3T-NAMESPACE-031** `error-code` · impact high · handled · verified 2026-10-13
  **Explicit deny (bucket or IAM session policy) -> AccessDeniedException 403 on all 6 namespace ops; foreign-account ARN -> ForbiddenException**
  With a table bucket policy Deny (Principal '*') on the six actions, and separately with an IAM session
  policy Deny while assuming a role that allows s3tables:*, GetNamespace, ListNamespaces, CreateNamespace,
  DeleteNamespace, GetTableBucket and ListTables ALL returned AccessDeniedException (HTTP 403,
  x-amzn-ErrorType 'AccessDeniedException') with the IAM-style message 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<role>/<session> is not authorized to perform: s3tables:<Action> on
  resource: <bucket ARN> with an explicit deny in a resource-based policy' (IAM path: '... with an explicit
  deny in an identity-based policy'). CreateNamespace, DeleteNamespace and ListTables do not declare
  AccessDeniedException in the model. ForbiddenException (HTTP 403, message 'Unauthorized') was returned only
  for a well-formed ARN of a bucket in another account (see the parent-ARN finding). A non-denied op on the
  same session (GetTable of a missing table) returned NotFoundException, proving the credentials worked.
  - ACK: terminal_codes, requeue · ops: GetNamespace, ListNamespaces, CreateNamespace, DeleteNamespace,
    GetTableBucket, ListTables
  - repro: PutTableBucketPolicy(Deny the six actions for '*') on an otherwise empty bucket; and AssumeRole
    with a session policy denying them; call each op
  - measurements: assume_role_success_after_create_role_s=9.1, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-017](#s3t-namespace-017), [S3T-NAMESPACE-029](#s3t-namespace-029), [S3T-NAMESPACE-030](#s3t-namespace-030), [S3T-NAMESPACE-003](#s3t-namespace-003) · hypotheses: H-N-041,
    H-N-042 · evidence: namespace/error-taxonomy/parent-arn-iam-deny-cross-region
  - notes: H-N-041 refuted (ForbiddenException was never returned for an IAM/resource-policy deny;
    AccessDeniedException was returned on every op, including the three that do not model it). H-N-042 refuted
    (reads and writes return the same code). New role: AssumeRole succeeded 9.1 s after CreateRole; the
    first...
  - full notes: [details/S3T-NAMESPACE-031.md](details/S3T-NAMESPACE-031.md)

## Request validation

- <a id="s3t-namespace-011"></a>**S3T-NAMESPACE-011** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateNamespace list: 2 elements, [x,x] or a string -> 400 'Multipart namespaces are not supported'; [], [''], [None] -> 400 invalid name**
  two elements [<p>_m_a,<p>_m_b] -> BadRequestException (HTTP 400) 'Multipart namespaces are not supported.' |
  empty list [] (server) -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify
  a different namespace name, and then try again.' | empty list [] (botocore validation on) -> client-side
  ParamValidationError | duplicate elements [x,x] -> BadRequestException (HTTP 400) 'Multipart namespaces are
  not supported.' | one empty string [''] -> BadRequestException (HTTP 400) 'The specified namespace name
  isn’t valid. Specify a different namespace name, and then try again.' | null element [None] ->
  BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different namespace
  name, and then try again.' | string instead of list -> BadRequestException (HTTP 400) 'Multipart namespaces
  are not supported.'
  - ACK: custom_create, terminal_codes · ops: CreateNamespace · fields: namespace
  - repro: CreateNamespace(arn, namespace=<each list in behavior>) with botocore parameter_validation=False
    (and on, where noted)
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-014](#s3t-namespace-014), [S3T-NAMESPACE-015](#s3t-namespace-015),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-005, H-N-006, H-N-007, H-N-008 · evidence:
    namespace/weird-inputs/naming-edge-cases
  - notes: H-N-005 confirmed / H-N-006 refuted (two-element list rejected, nothing created). H-N-007
    qualified: the model has min=1,max=1 on the list, so botocore blocks [] client-side (ParamValidationError)
    while the server answers BadRequestException. H-N-008 confirmed ([x,x] -> BadRequestException; [''] ->...
  - full notes: [details/S3T-NAMESPACE-011.md](details/S3T-NAMESPACE-011.md)

- <a id="s3t-namespace-012"></a>**S3T-NAMESPACE-012** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace: trailing '_' -> 200, leading '_' -> BadRequestException, '__' inside -> 200, leading digit -> 200**
  trailing underscore <p>_ -> HTTP 200; GetNamespace {'ackq_595cd6_ns_': "HTTP 200
  stored=['ackq_595cd6_ns_']"}; delete HTTP 204 | double trailing <p>__ -> HTTP 200; GetNamespace
  {'ackq_595cd6_ns__': "HTTP 200 stored=['ackq_595cd6_ns__']"}; delete HTTP 204 | leading underscore _<p> ->
  BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different namespace
  name, and then try again.' | double underscore inside <p>__x -> HTTP 200; GetNamespace {'ackq_595cd6_ns__x':
  "HTTP 200 stored=['ackq_595cd6_ns__x']"}; delete HTTP 204 | underscore only '_' -> BadRequestException (HTTP 400)
  'The specified namespace name isn’t valid. Specify a different namespace name, and then try again.' |
  leading digit 7<p> -> HTTP 200; GetNamespace {'7ackq_595cd6_ns': "HTTP 200 stored=['7ackq_595cd6_ns']"};
  delete HTTP 204
  - ACK: custom_create, terminal_codes · ops: CreateNamespace, GetNamespace, DeleteNamespace · fields:
    namespace
  - repro: CreateNamespace with names ending/starting with '_' (see behavior); GetNamespace + DeleteNamespace
    each accepted one
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-014](#s3t-namespace-014), [S3T-NAMESPACE-015](#s3t-namespace-015),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-021 · evidence: namespace/weird-inputs/naming-edge-cases
  - notes: H-N-021 confirmed (trailing '_' accepted: True; leading '_' rejected: True).

- <a id="s3t-namespace-013"></a>**S3T-NAMESPACE-013** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace with uppercase/hyphen/period/space/slash: all rejected BadRequestException...**
  uppercase <p>_MyNs -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a
  different namespace name, and then try again.' | hyphen <p>-ns -> BadRequestException (HTTP 400) 'The
  specified namespace name isn’t valid. Specify a different namespace name, and then try again.' | period
  <p>.ns -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different
  namespace name, and then try again.' | space '<p> ns' -> BadRequestException (HTTP 400) 'The specified
  namespace name isn’t valid. Specify a different namespace name, and then try again.' | slash <p>/ns ->
  BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different namespace
  name, and then try again.' | trailing space '<p>_t ' -> BadRequestException (HTTP 400) 'The specified
  namespace name isn’t valid. Specify a different namespace name, and then try again.' | leading space '
  <p>_s' -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different
  namespace name, and then try again.'. ListNamespaces(prefix=<p>) afterwards: [].
  - ACK: custom_create, terminal_codes · ops: CreateNamespace, GetNamespace, ListNamespaces · fields:
    namespace
  - repro: CreateNamespace with each name in behavior (parameter_validation=False); ListNamespaces(prefix)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-014](#s3t-namespace-014), [S3T-NAMESPACE-015](#s3t-namespace-015),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-022, H-N-023 · evidence: namespace/weird-inputs/naming-edge-cases
  - notes: H-N-022 confirmed; H-N-023 refuted (uppercase name rejected). Full original title: CreateNamespace
    with uppercase/hyphen/period/space/slash: all rejected BadRequestException; no lowercased or stripped
    variant appears in ListNamespaces

- <a id="s3t-namespace-014"></a>**S3T-NAMESPACE-014** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace name length: 255 chars -> 200 (echoed intact), 256 chars -> BadRequestException; single char 'a'/'7' -> 200/200**
  255 chars -> HTTP 200; GetNamespace {'<255 chars>': "HTTP 200 stored=['<255 chars>']"}; delete HTTP 204 |
  256 chars -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different
  namespace name, and then try again.' | 1 char 'a' -> HTTP 200; GetNamespace {'a': "HTTP 200 stored=['a']"};
  delete HTTP 204 | 1 char '7' -> HTTP 200; GetNamespace {'7': "HTTP 200 stored=['7']"}; delete HTTP 204
  - ACK: custom_create, terminal_codes · ops: CreateNamespace, GetNamespace · fields: namespace
  - repro: CreateNamespace with a 255- and a 256-character [a-z0-9_] name and with 'a' and '7'; GetNamespace
    each accepted one
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-015](#s3t-namespace-015),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-024 · evidence: namespace/weird-inputs/naming-edge-cases
  - notes: H-N-024 confirmed.

- <a id="s3t-namespace-015"></a>**S3T-NAMESPACE-015** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace reserved prefix: 'aws', 'awsfoo', 'aws_data', 'Aws_data' -> BadRequestException; 'aws' later in the name -> 200**
  aws_data -> BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different
  namespace name, and then try again.' | awsfoo -> BadRequestException (HTTP 400) 'The specified namespace
  name isn’t valid. Specify a different namespace name, and then try again.' | aws -> BadRequestException
  (HTTP 400) 'The specified namespace name isn’t valid. Specify a different namespace name, and then try
  again.' | <p>_my_aws_data -> HTTP 200; GetNamespace {'ackq_595cd6_ns_my_aws_data': "HTTP 200
  stored=['ackq_595cd6_ns_my_aws_data']"}; delete HTTP 204 | <p>_aw_s -> HTTP 200; GetNamespace
  {'ackq_595cd6_ns_aw_s': "HTTP 200 stored=['ackq_595cd6_ns_aw_s']"}; delete HTTP 204 | Aws_data (capital) ->
  BadRequestException (HTTP 400) 'The specified namespace name isn’t valid. Specify a different namespace
  name, and then try again.'
  - ACK: custom_create, terminal_codes · ops: CreateNamespace · fields: namespace
  - repro: CreateNamespace with 'aws_data', 'awsfoo', 'aws', '<lab>_my_aws_data'
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-014](#s3t-namespace-014),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-025 · evidence: namespace/weird-inputs/naming-edge-cases
  - notes: H-N-025 confirmed (aws_data rejected: True, awsfoo rejected: True, my_aws_data accepted: True).
    Full original title: CreateNamespace reserved prefix: 'aws_data' -> BadRequestException, 'awsfoo' ->
    BadRequestException, 'aws' -> BadRequestException; name containing 'aws' later -> 200

- <a id="s3t-namespace-032"></a>**S3T-NAMESPACE-032** `request-validation` · impact medium · handled · verified 2026-10-13
  **ListNamespaces maxNamespaces 0/-1 -> BadRequestException, 1001 -> 200 (not rejected); =1 pages 3 namespaces in 3 pages, last page no token**
  Bucket with exactly 3 namespaces, client validation off: maxNamespaces=0 -> BadRequestException (HTTP 400)
  'The specified limit is not valid. A limit should be a positive number'; maxNamespaces=1 -> HTTP 200, 1
  items, token present; maxNamespaces=2 -> HTTP 200, 2 items, token present; maxNamespaces=3 -> HTTP 200, 3
  items, token absent; maxNamespaces=1000 -> HTTP 200, 3 items, token absent; maxNamespaces=1001 -> HTTP 200,
  3 items, token absent; maxNamespaces=-1 -> BadRequestException (HTTP 400) 'The specified limit is not valid.
  A limit should be a positive number'. Following the token at maxNamespaces=1 took 3 pages, returned
  ['ackq_f745ff_nsa', 'ackq_f745ff_nsb', 'ackq_f745ff_nsc'] (0 duplicates); the final page omits
  continuationToken.
  - ACK: custom_find · ops: ListNamespaces · fields: maxNamespaces, continuationToken
  - repro: CreateTableBucket; 3 x CreateNamespace; ListNamespaces(maxNamespaces=0|1|2|3|1000|1001|-1); follow
    tokens at 1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-038](#s3t-namespace-038), [S3T-TABLE-293](table.md#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-TABLE-168](table.md#s3t-table-168), [S3T-TABLE-092](table.md#s3t-table-092),
    [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](table.md#s3t-table-091), [S3T-NAMESPACE-008](#s3t-namespace-008), [S3T-NAMESPACE-009](#s3t-namespace-009), [S3T-NAMESPACE-010](#s3t-namespace-010),
    [S3T-NAMESPACE-005](#s3t-namespace-005) · hypotheses: H-N-031 · evidence: namespace/limits/pagination-and-bursts
  - notes: H-N-031 qualified: maxNamespaces=0 and -1 -> BadRequestException 400 'The specified limit is not
    valid. A limit should be a positive number' as predicted, =1 returns one entry plus a token, the last page
    omits the token (key absent), BUT maxNamespaces=1001 is accepted with HTTP 200 (not rejected, so...
  - full notes: [details/S3T-NAMESPACE-032.md](details/S3T-NAMESPACE-032.md)

## Update granularity and ordering

- <a id="s3t-namespace-026"></a>**S3T-NAMESPACE-026** `prerequisite` · impact high · handled · verified 2026-10-13
  **CreateTable or RenameTable into a missing namespace -> NotFoundException, no implicit namespace is created; the table stays put**
  CreateTable(namespace=<missing>, format=ICEBERG) -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.'; ListNamespaces(prefix) afterwards does not contain the namespace. RenameTable(ns1/t1,
  newNamespaceName=<missing>) -> NotFoundException (HTTP 404) 'The specified destination namespace does not
  exist.'; GetTable(ns1/t1) -> HTTP 200 and the namespace still does not exist. RenameTable to an existing
  namespace -> HTTP 204; GetTable at the new place -> HTTP 200 (tableARN unchanged).
  - ACK: references, requeue, exceptions.404 · ops: CreateTable, RenameTable, ListNamespaces, GetTable ·
    fields: namespace, newNamespaceName
  - repro: CreateTable(namespace=<never created>) -> ListNamespaces(prefix) ->
    RenameTable(newNamespaceName=<never created>) -> GetTable -> RenameTable(newNamespaceName=<existing>)
  - handling: handled via `generator.yaml:128-136; pkg/resource/namespace/sdk.go:208-217; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-001](#s3t-namespace-001), [S3T-NAMESPACE-025](#s3t-namespace-025), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-029](#s3t-namespace-029) · hypotheses: H-N-046,
    H-N-047 · evidence: namespace/dependencies/delete-with-tables
  - notes: H-N-046 confirmed; H-N-047 (contrarian: implicit creation) refuted. The CreateTable 404 message is
    'The specified namespace does not exist.', the RenameTable one 'The specified destination namespace does
    not exist.'.
  - full notes: [details/S3T-NAMESPACE-026.md](details/S3T-NAMESPACE-026.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-namespace-003"></a>**S3T-NAMESPACE-003** `read-gap` · impact medium · handled · verified 2026-10-13
  **No namespace read path returns the parent tableBucketARN; tableBucketId as tableBucketARN -> BadRequestException**
  GetNamespace and ListNamespaces entries carry tableBucketId (36-char uuid) but no member containing 'arn:'.
  GetTableBucket(tableBucketARN=<tableBucketId>) -> BadRequestException (HTTP 400) 'The specified table bucket
  ARN is not valid.'; GetNamespace(tableBucketARN=<tableBucketId>, namespace) -> BadRequestException (HTTP 400)
  'The specified table bucket ARN is not valid.'. GetNamespace.tableBucketId == GetTableBucket.tableBucketId,
  ownerAccountId equal; createdBy is 12-digit account id (== ownerAccountId) although the caller is an assumed
  IAM role.
  - ACK: is_arn_primary_key, references, custom_find · ops: GetNamespace, ListNamespaces, GetTableBucket ·
    fields: tableBucketId, tableBucketARN, createdBy, ownerAccountId
  - repro: CreateNamespace -> GetNamespace -> GetTableBucket(tableBucketARN=<tableBucketId value>)
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304`
  - related: [S3T-NAMESPACE-002](#s3t-namespace-002), [S3T-NAMESPACE-041](#s3t-namespace-041), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-025](#s3t-namespace-025),
    [S3T-NAMESPACE-006](#s3t-namespace-006), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-007](#s3t-namespace-007), [S3T-NAMESPACE-022](#s3t-namespace-022), [S3T-NAMESPACE-017](#s3t-namespace-017),
    [S3T-NAMESPACE-029](#s3t-namespace-029), [S3T-NAMESPACE-030](#s3t-namespace-030), [S3T-NAMESPACE-031](#s3t-namespace-031) · hypotheses: H-N-051, H-N-038 · evidence:
    namespace/round-trip/create-get-list-shape
  - notes: H-N-051 confirmed. H-N-038 confirmed (createdBy format: 12-digit account id; only one principal
    available, so the assumed-role-vs-account-id question has one data point).

- <a id="s3t-namespace-004"></a>**S3T-NAMESPACE-004** `shape-mismatch` · impact high · handled · verified 2026-10-13
  **namespace is list<string> (min 1, max 1) in CreateNamespace in/out and Get/List output but a scalar string in Get/Delete/ListTables input**
  botocore model (s3tables, api version 2018-05-10): CreateNamespace.input=list<string> {"max": 1, "min": 1};
  CreateNamespace.output=list<string> {}; GetNamespace.input=string {"max": 255, "min": 1, "pattern":
  "[0-9a-z_]*"}; GetNamespace.output=list<string> {}; DeleteNamespace.input=string {"max": 255, "min": 1,
  "pattern": "[0-9a-z_]*"}; ListTables.input=string {"max": 255, "min": 1, "pattern": "[0-9a-z_]*"};
  ListNamespaces.output.namespaces[].namespace=list<string> {}. Live calls agree: CreateNamespace accepted
  namespace=[name] and echoed a list; GetNamespace required namespace=<string> and returned namespace=[name].
  - ACK: custom_find, custom_delete, hook:sdk_read_one_pre_build_request, hook:sdk_delete_pre_build_request ·
    ops: CreateNamespace, GetNamespace, DeleteNamespace, ListNamespaces, ListTables · fields: namespace
  - repro: Inspect botocore service model for s3tables; CreateNamespace([name]) then GetNamespace(name)
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131`
  - related: [S3T-NAMESPACE-002](#s3t-namespace-002), [S3T-NAMESPACE-041](#s3t-namespace-041), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-025](#s3t-namespace-025),
    [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-014](#s3t-namespace-014), [S3T-NAMESPACE-015](#s3t-namespace-015),
    [S3T-NAMESPACE-016](#s3t-namespace-016) · hypotheses: H-N-052 · evidence: namespace/round-trip/create-get-list-shape
  - notes: H-N-052 confirmed from the model plus one live round trip. The CreateNamespace list has min=1,max=1
    in the model, so botocore blocks [] client-side but passes two-element lists to the server (see
    weird-inputs/naming-edge-cases).

## Response fidelity and consistency

- <a id="s3t-namespace-002"></a>**S3T-NAMESPACE-002** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **CreateNamespace 200 body = {tableBucketARN, namespace:[name]} only; GetNamespace adds 5 read-only members, no ARN/tags/status**
  CreateNamespace(tableBucketARN, namespace=[name]) returned HTTP 200 with raw body
  {"namespace":["ackq_86ff1a_ns_a"],"tableBucketARN":"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-86ff1a-b1"}:
  members ['namespace', 'tableBucketARN']; tableBucketARN byte-identical to the input and namespace is a
  one-element list. GetNamespace at +12 ms returned HTTP 200 with members ['createdAt', 'createdBy',
  'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId'] (namespace=list[str], createdAt=timestamp,
  createdBy=12-digit account id, ownerAccountId=12-digit account id, namespaceId=36-char uuid,
  tableBucketId=36-char uuid); no member contains an ARN, no tags, no status/state.
  - ACK: hook:sdk_create_post_set_output, is_read_only · ops: CreateNamespace, GetNamespace · fields:
    namespace, tableBucketARN, createdAt, createdBy, ownerAccountId, namespaceId, tableBucketId
  - repro: CreateTableBucket -> CreateNamespace(arn, [name]) -> GetNamespace(arn, name) immediately
  - measurements: get_after_create_latency_ms=12, created_at_minus_client_clock_s=0.014, n=1
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131; pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:188-189; pkg/resource/namespace/manager.go:284-291; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98`
  - related: [S3T-NAMESPACE-041](#s3t-namespace-041), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-025](#s3t-namespace-025) ·
    hypotheses: H-N-001, H-N-002 · evidence: namespace/round-trip/create-get-list-shape
  - notes: H-N-001 confirmed (create echo identical; GetNamespace members exactly {namespace, createdAt,
    createdBy, ownerAccountId, namespaceId, tableBucketId}: True). H-N-002 createdAt-vs-client-clock part:
    createdAt is 0.01 s from the client clock (single sample; the 20-trial part is in...
  - full notes: [details/S3T-NAMESPACE-002.md](details/S3T-NAMESPACE-002.md)

- <a id="s3t-namespace-005"></a>**S3T-NAMESPACE-005** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **ListNamespaces entries carry the same 6 members and values as GetNamespace; order is name-ascending**
  With 5 namespaces created in the order ['ns_a', 'ns_b', 'ns_a2', 'ns_c', 'ns_d'], ListNamespaces returned
  them as ['ns_a', 'ns_a2', 'ns_b', 'ns_c', 'ns_d'] (sorted ascending by name). Each entry has members
  ['createdAt', 'createdBy', 'namespace', 'namespaceId', 'ownerAccountId', 'tableBucketId']; comparing every
  member with GetNamespace of the same name gave 0 mismatches.
  - ACK: custom_find, list_operation.match_fields · ops: ListNamespaces, GetNamespace · fields: namespaces,
    namespace, createdAt, createdBy, ownerAccountId, namespaceId, tableBucketId
  - repro: Create ns_a, ns_b, ns_a2, ns_c, ns_d in that order -> ListNamespaces -> GetNamespace each ->
    compare members
  - handling: handled via `generator.yaml:188-189; pkg/resource/namespace/manager.go:284-291; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-NAMESPACE-002](#s3t-namespace-002), [S3T-NAMESPACE-041](#s3t-namespace-041), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-025](#s3t-namespace-025),
    [S3T-NAMESPACE-008](#s3t-namespace-008), [S3T-NAMESPACE-009](#s3t-namespace-009), [S3T-NAMESPACE-010](#s3t-namespace-010), [S3T-NAMESPACE-032](#s3t-namespace-032) · hypotheses: H-N-034 ·
    evidence: namespace/round-trip/create-get-list-shape
  - notes: H-N-034 confirmed: members identical; ordering ascending by name (5 names, single list call).

- <a id="s3t-namespace-008"></a>**S3T-NAMESPACE-008** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListNamespaces prefix: literal, case-sensitive, left-anchored; '' and invalid-char prefixes -> 200 (empty); >255 chars -> BadRequest 400**
  With namespaces <base>ns_a, <base>ns_b, <base>xns_a in the bucket: literal <base>ns_ -> HTTP 200 ['ns_a',
  'ns_b']; upper <BASE>NS_ -> HTTP 200 []; percent <base>ns% -> HTTP 200 []; full name <base>ns_a -> HTTP 200
  ['ns_a']; unanchored ns_ -> HTTP 200 []; unanchored s_ -> HTTP 200 []; empty '' -> HTTP 200 ['ns_a', 'ns_b',
  'xns_a']; hyphen/upper My-Ns -> HTTP 200 []; dot a.b -> HTTP 200 []; 300 x 'a' -> BadRequestException (HTTP 400)
  'The specified search prefix is not valid. A search prefix can not exceed 255 characters.'; 256 x 'a' ->
  BadRequestException (HTTP 400) 'The specified search prefix is not valid. A search prefix can not exceed 255
  characters.'; unicode e-acute -> HTTP 200 []; space 'a b' -> HTTP 200 []; slash a/b -> HTTP 200 [].
  - ACK: custom_find, list_operation.match_fields · ops: ListNamespaces · fields: prefix
  - repro: Create ns_a, ns_b, xns_a (lab prefix) -> ListNamespaces with each prefix listed in behavior
    (botocore parameter_validation=False for '')
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-009](#s3t-namespace-009), [S3T-NAMESPACE-010](#s3t-namespace-010), [S3T-NAMESPACE-032](#s3t-namespace-032), [S3T-NAMESPACE-005](#s3t-namespace-005) · hypotheses: H-N-033,
    H-N-050 · evidence: namespace/response-fidelity/list-prefix-pagination
  - notes: H-N-033 confirmed (literal prefix returned exactly {ns_a, ns_b}: True; NS_/ns%/unanchored returned
    empty: True; prefix='' -> HTTP 200). H-N-050 qualified/refuted (see behavior) (My-Ns -> HTTP 200; a.b ->
    HTTP 200; 300 chars -> BadRequestException (HTTP 400) 'The specified search prefix is not valid....
  - full notes: [details/S3T-NAMESPACE-008.md](details/S3T-NAMESPACE-008.md)

- <a id="s3t-namespace-018"></a>**S3T-NAMESPACE-018** `eventual-consistency` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **GetNamespace and ListNamespaces reflect CreateNamespace and DeleteNamespace at +0 ms in 20/20 trials (no read-after-write window)**
  20 trials: CreateNamespace -> GetNamespace at +0 ms returned 200 in 20/20 (latency 10-13 ms), ListNamespaces
  contained the name in 20/20 and ListNamespaces(prefix=<name>) returned exactly one entry in 20/20.
  DeleteNamespace returned HTTP 204 in 20/20 (latency 23-29 ms, raw body [('', None)]); GetNamespace at +0 ms
  after the 204 returned NotFoundException in 20/20 and ListNamespaces no longer contained it in 20/20.
  - ACK: none, e2e-timing · ops: CreateNamespace, GetNamespace, ListNamespaces, DeleteNamespace · fields:
    namespace, createdAt
  - repro: 20x: CreateNamespace -> GetNamespace, ListNamespaces, ListNamespaces(prefix) immediately ->
    DeleteNamespace -> GetNamespace, ListNamespaces immediately
  - measurements: n=20, create_latency_ms.min=30, create_latency_ms.max=74, create_latency_ms.n=20,
    get_after_create_latency_ms.min=10, get_after_create_latency_ms.max=13, get_after_create_latency_ms.n=20,
    list_latency_ms.min=9, list_latency_ms.max=49, list_latency_ms.n=20, delete_latency_ms.min=23,
    delete_latency_ms.max=29, delete_latency_ms.n=20, created_at_minus_client_clock_s.min=-0.023,
    created_at_minus_client_clock_s.max=-0.019, created_at_minus_client_clock_s.n=20,
    get_after_create_200_trials=20, list_after_create_present_trials=20, list_prefix_exact_one_trials=20,
    delete_204_trials=20, get_after_delete_404_trials=20, list_after_delete_absent_trials=20, get_lag_s=null,
    list_lag_s=null, get_gone_lag_s=null, list_gone_lag_s=null
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-019](#s3t-namespace-019), [S3T-NAMESPACE-020](#s3t-namespace-020), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-NAMESPACE-010](#s3t-namespace-010) ·
    hypotheses: H-N-002, H-N-003, H-N-004, H-N-011 · evidence:
    namespace/consistency-windows/create-delete-visibility
  - notes: H-N-002 confirmed (createdAt minus client clock {'min': -0.023, 'max': -0.019, 'n': 20} s). H-N-003
    confirmed. H-N-004 (contrarian: list lags) refuted. H-N-011 confirmed (delete is synchronous: 204,
    immediate 404).

- <a id="s3t-namespace-019"></a>**S3T-NAMESPACE-019** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace at +0 ms after CreateTableBucket -> 200 in 20/20; CreateTable at +0 ms after CreateNamespace -> 200 in 20/20**
  20 trials of CreateTableBucket (latency 1444-2931 ms) -> CreateNamespace immediately (200 in 20/20, latency
  49-88 ms) -> CreateTable(format=ICEBERG) immediately (200 in 20/20, latency 310-440 ms) -> DeleteTable,
  DeleteNamespace, DeleteTableBucket (bucket delete 445-523 ms, 204).
  - ACK: none, e2e-timing, references · ops: CreateTableBucket, CreateNamespace, CreateTable,
    DeleteTableBucket · fields: tableBucketARN, namespace
  - repro: 20x: CreateTableBucket -> CreateNamespace at +0 ms -> CreateTable at +0 ms -> delete table,
    namespace, bucket
  - measurements: n=20, create_bucket_latency_ms.min=1444, create_bucket_latency_ms.max=2931,
    create_bucket_latency_ms.n=20, create_namespace_latency_ms.min=49, create_namespace_latency_ms.max=88,
    create_namespace_latency_ms.n=20, create_table_latency_ms.min=310, create_table_latency_ms.max=440,
    create_table_latency_ms.n=20, delete_table_latency_ms.min=139, delete_table_latency_ms.max=187,
    delete_table_latency_ms.n=20, delete_bucket_latency_ms.min=445, delete_bucket_latency_ms.max=523,
    delete_bucket_latency_ms.n=20, namespace_at_zero_200_trials=20, table_at_zero_200_trials=20,
    create_ns_lag_s=null, create_table_lag_s=null
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-160](tablebucket.md#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](tablebucket.md#s3t-tablebucket-167), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](tablebucket.md#s3t-tablebucket-017),
    [S3T-TABLEBUCKET-003](tablebucket.md#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](tablebucket.md#s3t-tablebucket-122), [S3T-TABLEBUCKET-121](tablebucket.md#s3t-tablebucket-121), [S3T-NAMESPACE-018](#s3t-namespace-018), [S3T-NAMESPACE-020](#s3t-namespace-020),
    [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-NAMESPACE-010](#s3t-namespace-010) · hypotheses: H-N-045 · evidence:
    namespace/consistency-windows/create-delete-visibility
  - notes: H-N-045 confirmed (20 trials, sequential, one extra bucket alive at a time).

- <a id="s3t-namespace-027"></a>**S3T-NAMESPACE-027** `eventual-consistency` · impact medium · handled · verified 2026-10-13
  **DeleteNamespace at +0 ms after DeleteTable of its last table -> 204 in 20/20; after RenameTable moving the last table out -> 204 in 10/10**
  20 trials CreateNamespace -> CreateTable -> DeleteTable (latency {'min': 130, 'max': 167, 'n': 20} ms) ->
  DeleteNamespace immediately: first-call outcomes ['HTTP 204']. 10 trials with
  RenameTable(newNamespaceName=<other ns>) (latency {'min': 155, 'max': 202, 'n': 10} ms) instead of
  DeleteTable: first-call outcomes ['HTTP 204'].
  - ACK: requeue, none · ops: DeleteTable, RenameTable, DeleteNamespace · fields: namespace, newNamespaceName
  - repro: 20x: CreateNamespace -> CreateTable -> DeleteTable -> DeleteNamespace at +0 ms; 10x with
    RenameTable to another namespace instead of DeleteTable
  - measurements: n_delete_table_trials=20, n_rename_trials=10, n=30,
    delete_ns_immediate_204_after_delete_table=20, delete_ns_immediate_204_after_rename=10,
    time_to_204_s_after_delete_table=null, time_to_204_s_after_rename=null, delete_table_latency_ms.min=130,
    delete_table_latency_ms.max=167, delete_table_latency_ms.n=20, rename_table_latency_ms.min=155,
    rename_table_latency_ms.max=202, rename_table_latency_ms.n=10
  - handling: handled via `test/e2e/conftest.py:102-106; test/e2e/conftest.py:141-142; generator.yaml:128-136; pkg/resource/namespace/sdk.go:208-217`
  - related: [S3T-NAMESPACE-018](#s3t-namespace-018), [S3T-NAMESPACE-019](#s3t-namespace-019), [S3T-NAMESPACE-020](#s3t-namespace-020), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-010](#s3t-namespace-010),
    [S3T-NAMESPACE-024](#s3t-namespace-024), [S3T-NAMESPACE-039](#s3t-namespace-039), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-021](#s3t-namespace-021), [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019) ·
    hypotheses: H-N-017 · evidence: namespace/dependencies/delete-with-tables
  - notes: H-N-017 (contrarian: emptiness check lags) refuted.

## Tags

- <a id="s3t-namespace-007"></a>**S3T-NAMESPACE-007** `tag-semantics` · impact high · handled · verified 2026-10-15, re-verified
  **Namespaces are not taggable: TagResource on a synthesized namespace ARN -> BadRequestException; bucket tags never appear in namespace reads**
  tag bucket/<b>/namespace/<ns> -> BadRequestException (HTTP 400) 'The specified resourceArn is not valid.';
  listtags bucket/<b>/namespace/<ns> -> BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.'; tag bucket/<b>/namespaces/<ns> -> BadRequestException (HTTP 400) 'The specified resourceArn is not
  valid.'; listtags bucket/<b>/namespaces/<ns> -> BadRequestException (HTTP 400) 'The specified resourceArn is
  not valid.'. TagResource on the bucket ARN -> HTTP 200; afterwards GetNamespace and ListNamespaces(prefix)
  carry no tag member.
  - ACK: tags.ignore, scope:field-on-parent · ops: TagResource, ListTagsForResource, GetNamespace,
    ListNamespaces · fields: resourceArn, tags
  - repro: TagResource(resourceArn='<bucketARN>/namespace/<ns>', tags) and ListTagsForResource with the same
    ARN; TagResource(bucket ARN); GetNamespace
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:188-189; pkg/resource/namespace/manager.go:284-291`
  - related: [S3T-NAMESPACE-006](#s3t-namespace-006), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-022](#s3t-namespace-022) · hypotheses: H-N-040
    · evidence: namespace/round-trip/create-get-list-shape
  - notes: H-N-040 confirmed. Two synthesized ARN forms tried (…/namespace/<ns> and …/namespaces/<ns>); no
    model operation accepts a namespace identifier for tagging.

## Delete semantics

- <a id="s3t-namespace-024"></a>**S3T-NAMESPACE-024** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6 retries over 10 s; table and namespace untouched**
  DeleteNamespace on a namespace holding one ICEBERG table returned BadRequestException with message ['The
  namespace that you tried to delete is not empty.'] on 6/6 attempts at 2 s intervals (latency {'min': 11,
  'max': 59, 'n': 6} ms). Afterwards GetTable -> HTTP 200 and GetNamespace -> HTTP 200 (nothing cascaded). The
  duplicate-create rejection on the same namespace is ConflictException 'A namespace with an identical name
  already exists in the bucket.', so the two rejections are distinguishable by code and message.
  - ACK: requeue, pre-delete-cleanup, terminal_codes · ops: DeleteNamespace, CreateNamespace, GetTable,
    GetNamespace · fields: namespace
  - repro: CreateNamespace -> CreateTable(format=ICEBERG) -> DeleteNamespace x6 at 2 s -> GetTable,
    GetNamespace -> CreateNamespace(same name)
  - measurements: delete_nonempty_latency_ms.min=11, delete_nonempty_latency_ms.max=59,
    delete_nonempty_latency_ms.n=6, attempts=6, n=6
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019), [S3T-TABLEBUCKET-044](tablebucket.md#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](tablebucket.md#s3t-tablebucket-132), [S3T-TABLEBUCKET-070](tablebucket.md#s3t-tablebucket-070),
    [S3T-NAMESPACE-037](#s3t-namespace-037), [S3T-NAMESPACE-039](#s3t-namespace-039), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-021](#s3t-namespace-021) ·
    hypotheses: H-N-015, H-N-016, H-N-010 · evidence: namespace/dependencies/delete-with-tables
  - notes: H-N-015 refuted; H-N-016 (contrarian: BadRequestException) confirmed; H-N-010 qualified: the
    delete-non-empty rejection is BadRequestException (not 409 as hypothesized), it is deterministic (not
    transient) and its message differs from the duplicate-create ConflictException (['The namespace that
    you...
  - full notes: [details/S3T-NAMESPACE-024.md](details/S3T-NAMESPACE-024.md)

- <a id="s3t-namespace-028"></a>**S3T-NAMESPACE-028** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable vs DeleteNamespace fired concurrently (20 trials): CreateTable NotFoundException | DeleteNamespace 204 in 20/20; 0 orphan tables**
  20 trials with CreateTable(ns, t) and DeleteNamespace(ns) released from a 2-thread barrier: outcome pairs
  (CreateTable | DeleteNamespace) = {"NotFoundException | HTTP 204": 20}. Trials where both returned success: 0.
  Trials leaving a table listed under a namespace absent from ListNamespaces: 0. Messages: CreateTable 404
  "NotFoundException (HTTP 404) 'The specified namespace does not exist.'"; DeleteNamespace 409 None.
  - ACK: requeue, none · ops: CreateTable, DeleteNamespace, ListTables, ListNamespaces · fields: namespace
  - repro: 20x: CreateNamespace -> CreateTable and DeleteNamespace from two threads at once ->
    ListNamespaces(prefix), ListTables(prefix), GetTable
  - measurements: n=20, both_ok=0, orphans=0, outcome_pairs.NotFoundException | HTTP 204=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-001](#s3t-namespace-001), [S3T-NAMESPACE-025](#s3t-namespace-025), [S3T-NAMESPACE-026](#s3t-namespace-026), [S3T-NAMESPACE-029](#s3t-namespace-029), [S3T-NAMESPACE-024](#s3t-namespace-024),
    [S3T-NAMESPACE-039](#s3t-namespace-039), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-NAMESPACE-021](#s3t-namespace-021), [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019) · hypotheses: H-N-018 ·
    evidence: namespace/dependencies/delete-with-tables
  - notes: H-N-018 confirmed (no trial left a table in a deleted namespace: True). Full original title:
    CreateTable vs DeleteNamespace fired concurrently (20 trials): both succeeded in 0, orphan tables 0;
    outcomes NotFoundException | HTTP 204 x20

- <a id="s3t-namespace-039"></a>**S3T-NAMESPACE-039** `delete-semantics` · impact high · handled · verified 2026-10-14
  **Parallel finalizers: DeleteTable||DeleteNamespace -> 400 'not empty' 6/8; DeleteNamespace||DeleteTableBucket -> 400 6/8; +0 ms retry 204**
  Fired from two threads at the same instant, n=8 each. Set 1 (namespace holding one table): DeleteTable ->
  {"HTTP 204": 8}, DeleteNamespace -> {"BadRequestException/400": 6, "HTTP 204": 2} (messages ["The namespace
  that you tried to delete is not empty."]); the pair alone removed the namespace in 2/8 trials; after the
  retry below 0 namespaces and 0 orphan tables were left. Immediate controller retry of the failed
  DeleteNamespace -> {"HTTP 204": 6}. Set 2 (bucket holding one empty namespace): DeleteNamespace -> {"HTTP
  204": 8}, DeleteTableBucket -> {"HTTP 204": 2, "BadRequestException/400": 6} (messages ["The bucket that you
  tried to delete is not empty."]); bucket left in 0/8; immediate retry of the failed DeleteTableBucket ->
  {"HTTP 204": 6}. The parent 'not empty' rejection is BadRequestException 400 and clears on the very next
  call once the child delete has returned.
  - ACK: requeue, terminal_codes, pre-delete-cleanup · ops: DeleteTable, DeleteNamespace, DeleteTableBucket
  - repro: 8x (namespace+table): DeleteTable and DeleteNamespace from two threads; retry the failed parent at
    +0 ms; 8x (bucket+namespace): DeleteNamespace and DeleteTableBucket likewise
  - measurements: set1_parent_rejected=6, set2_parent_rejected=6, set1_namespace_left=0, set2_bucket_left=0,
    n=8
  - handling: handled via `test/e2e/conftest.py:102-106; test/e2e/conftest.py:141-142; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-024](#s3t-namespace-024), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-021](#s3t-namespace-021) ·
    evidence: namespace/creative/parallel-finalizers
  - notes: extends [S3T-NAMESPACE-027](#s3t-namespace-027) (sequential +0 ms -> 204) and -024/[S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019) (400 'not empty')
    with the concurrent case two controller workers produce. Decision rule: a 400 BadRequestException whose
    message says 'not empty' on DeleteNamespace/DeleteTableBucket is a requeue (the child finalizer...
  - full notes: [details/S3T-NAMESPACE-039.md](details/S3T-NAMESPACE-039.md)

## Quotas and rate limits

- <a id="s3t-namespace-033"></a>**S3T-NAMESPACE-033** `quota-limit` · impact high · handled · verified 2026-10-13
  **Back-to-back CreateNamespace or DeleteNamespace (~30/s): 16 of 30 succeed, rest 429 'Too Many Requests' (1 burst each); 30 parallel Gets 200**
  30 GetNamespace calls fired concurrently (20 threads) on one namespace: run 1 {"200": 30} (wall 0.16s), run
  2 after 20 s {"200": 30} (wall 0.09s). 30 CreateNamespace for distinct fresh names sent sequentially with no
  pause: {"200": 16, "TooManyRequestsException/429": 14} in 0.57s (latency [2, 37] ms); 30 DeleteNamespace
  likewise: {"204": 16, "TooManyRequestsException/429": 14} in 0.51s (latency [2, 52] ms). Error messages:
  {"TooManyRequestsException": "Too Many Requests"}; Retry-After headers: [].
  - ACK: requeue, e2e-timing · ops: GetNamespace, CreateNamespace, DeleteNamespace
  - repro: 20 threads x 30 GetNamespace; 30 sequential CreateNamespace(fresh); 30 sequential DeleteNamespace
  - measurements: get_burst_wall_s=[0.16, 0.09], get_latency_ms_min_max=[[11,156], [10,78]],
    create_30_wall_s=0.57, delete_30_wall_s=0.51, create_latency_ms_min_max=[2, 37],
    delete_latency_ms_min_max=[2, 52], n=120
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-165](tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018),
    [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-NAMESPACE-035](#s3t-namespace-035), [S3T-NAMESPACE-040](#s3t-namespace-040) · hypotheses: H-N-043 · evidence:
    namespace/limits/pagination-and-bursts
  - notes: H-N-043 qualified: the read half is refuted at the tried scale (2 x 30 concurrent GetNamespace, all
    200, wall 0.09-0.16 s; the 200-call cell is untested), the mutation half is confirmed and sharper than
    predicted: 30 sequential CreateNamespace (distinct names, ~33 ms each, ~30 TPS) returned 200 for...
  - full notes: [details/S3T-NAMESPACE-033.md](details/S3T-NAMESPACE-033.md)

- <a id="s3t-namespace-035"></a>**S3T-NAMESPACE-035** `quota-limit` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateNamespace back-to-back: 14/20 then 429 (2 setup calls used budget); retry every 250 ms succeeds after 0.54 s; after a 5 s pause 16/20**
  20 sequential CreateNamespace (distinct names, no pause, ~2 ms each) -> {"200": 14,
  "TooManyRequestsException/429": 6} in 0.45s, first TooManyRequestsException at index 14 (message 'Too Many
  Requests', Retry-After []). Retrying the first rejected name every 250 ms succeeded after 0.54 s (attempts
  {'TooManyRequestsException (HTTP 429)': 2, 'HTTP 200': 1}). After a 5 s pause, 20 more sequential creates ->
  {"200": 16, "TooManyRequestsException/429": 4} (first failure at index 16). After another 5 s, 31 sequential
  DeleteNamespace -> {"204": 16, "TooManyRequestsException/429": 15} in 0.44s (first failure at index 16); a
  throttled delete succeeded again after 0.54 s.
  - ACK: requeue, e2e-timing · ops: CreateNamespace, DeleteNamespace · fields: namespace
  - repro: 20 x CreateNamespace back-to-back; retry the first 429 every 250 ms; sleep 5; 20 more; sleep 5;
    DeleteNamespace all back-to-back
  - measurements: create_wall_s=0.45, create_latency_ms_min_max=[2, 40], first_429_index_create=14,
    refill_lag_s_create=0.54, ok_after_5s_pause=16, first_429_index_delete=16, refill_lag_s_delete=0.54, n=71
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-033](#s3t-namespace-033), [S3T-NAMESPACE-040](#s3t-namespace-040) · hypotheses: H-N-043 · evidence:
    table/limits/default-page-120-tables
  - notes: Follow-up to the Namespace 16-then-429 finding (H-N-043 mutation half, qualified further): the
    limiter behaves like a token bucket of capacity 16 per operation - 2 setup CreateNamespace calls a few
    seconds earlier left 14 tokens (14 of 20 succeeded, 429 from index 14), after a 5 s pause exactly 16...
  - full notes: [details/S3T-NAMESPACE-035.md](details/S3T-NAMESPACE-035.md)

- <a id="s3t-namespace-040"></a>**S3T-NAMESPACE-040** `quota-limit` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **CreateNamespace 429 after ~16 back-to-back is per account-region (all buckets), not shared with Delete/CreateTable/Tag; <= 4/s is safe**
  CreateNamespace back-to-back (~27-30 calls/s) in one table bucket: 16 (sometimes 17) x 200 then
  TooManyRequestsException HTTP 429 'Too Many Requests' within ~0.55 s; the 429 carries no Retry-After header
  (only x-amzn-ErrorType and Date). Immediately after the 429, 10 back-to-back calls alternating X with
  CreateNamespace(A) ('.'=200, 'T'=429), 3 clean runs: X=CreateNamespace in ANOTHER table bucket of the same
  region -> X TTTTT A TTTTT in 4 of 6 trials (both '.....' in the other 2, where A had recovered too): the
  limit is shared by all table buckets of the account-region. X=CreateNamespace in the other region -> X
  '.....' in 3/3 (A 'TTTT.' in one): separate limiter. X=DeleteNamespace -> X '.....' in 3/3 while A
  'TTTTT'/'TTTTT'/'TTT..': deletes are NOT counted against it. X=CreateTable -> X '...' in 3/3 (A 'TTT' once);
  X=TagResource(bucket) -> X '...' in 3/3 (A 'TTT' twice): separate. After a drain, 13 CreateNamespace at 4/s
  -> 12/13 ok in 3/3 runs (only the first call, inside the same window, was rejected); in an earlier run 90
  CreateNamespace paced at <= 5/s and 16 at 4/s gave 0 x 429. The recovery pattern (whole batches passing or
  failing) looks like a fixed window of ~16 per second rather than a smooth token bucket.
  - ACK: requeue, one-per-reconcile, terminal_codes · ops: CreateNamespace, DeleteNamespace, CreateTable,
    TagResource · fields: namespace
  - repro: CreateNamespace back-to-back in A until 429, then 10 back-to-back calls alternating <other op or
    bucket> with CreateNamespace(A); repeat per X after 12 s idle; drain then 4/s
  - measurements: drain_ok_before_429=[16, 16, 16, 16, 16, 17, 16, 16, 16, 16, 17, 16, 17, 22, 11],
    same_region_other_bucket_X_patterns=[TTTTT, TTTTT, TTTTT, TTTTT, ....., .....],
    other_region_X_patterns=[....., ....., .....], delete_namespace_X_patterns=[....., ....., .....],
    create_table_X_patterns=[..., ..., ...], tag_bucket_X_patterns=[..., ..., ...],
    refill_ok_of_13_at_4_per_s=[12, 12, 12], paced_le_5_per_s_ok=106, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-033](#s3t-namespace-033), [S3T-NAMESPACE-035](#s3t-namespace-035) · evidence: namespace/creative/throttle-key
  - notes: extends [S3T-NAMESPACE-033](#s3t-namespace-033)/-035 with the KEY of the limit and the pacing rule: the budget is per
    account-region and shared by every table bucket, so N parallel workers creating Namespace CRs in different
    buckets still compete; DeleteNamespace (contrary to the '-033' wording that bundles...
  - full notes: [details/S3T-NAMESPACE-040.md](details/S3T-NAMESPACE-040.md)

## Cross-region

- <a id="s3t-namespace-030"></a>**S3T-NAMESPACE-030** `cross-region` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **us-west-2 tableBucketARN via a us-east-1 client -> BadRequestException 'Region does not match' on all 5 ops tried; no cross-region create**
  With the bucket existing in us-west-2, calls through a us-east-1 client: GetNamespace -> BadRequestException
  (HTTP 400) 'The provided Region does not match the Region of the resource.'; ListNamespaces ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  CreateNamespace -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'; GetTableBucket -> BadRequestException (HTTP 400) 'The provided Region does not match the Region
  of the resource.'; ListTables -> BadRequestException (HTTP 400) 'The provided Region does not match the
  Region of the resource.'. ListNamespaces in us-west-2 afterwards shows 0 namespace(s) with the attempted
  name (no cross-region create).
  - ACK: terminal_codes, references · ops: GetNamespace, ListNamespaces, CreateNamespace, GetTableBucket,
    ListTables · fields: tableBucketARN
  - repro: CreateTableBucket in us-west-2 -> from a us-east-1 client
    GetNamespace/ListNamespaces/CreateNamespace with the us-west-2 ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-057](tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-126](tablebucket-encryption-storage.md#s3t-tablebucket-126), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-NAMESPACE-017](#s3t-namespace-017),
    [S3T-NAMESPACE-029](#s3t-namespace-029), [S3T-NAMESPACE-031](#s3t-namespace-031), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-TABLE-243](table-record-expiration.md#s3t-table-243),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-190](table-record-expiration.md#s3t-table-190) · hypotheses: H-N-048 · evidence:
    namespace/error-taxonomy/parent-arn-iam-deny-cross-region
  - notes: H-N-048 confirmed (expected BadRequestException with a region/ARN mismatch message; observed
    ['BadRequestException']). Full original title: A us-west-2 tableBucketARN sent to the us-east-1 endpoint
    -> BadRequestException on GetNamespace/ListNamespaces/CreateNamespace/GetTableBucket/ListTables

## Handling gaps (bugs to file)

- [S3T-NAMESPACE-024](#s3t-namespace-024) - DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6
  retries over 10 s; table and namespace untouched (suspected bug)
  Suspected controller bug confirmed by evidence: DeleteNamespace on a non-empty namespace returns
  BadRequestException 400 'The namespace that you tried to delete is not empty.' deterministically (6/6 over
  10 s; same code and pattern for DeleteTableBucket with a namespace inside, [S3T-NAMESPACE-039](#s3t-namespace-039)), so with
  terminal_codes: [BadRequestException] a Namespace or TableBucket CR deleted together with its children goes
  Terminal on the first delete attempt while the child finalizer is still in flight (6/8 concurrent trials in
  [S3T-NAMESPACE-039](#s3t-namespace-039)) and never retries, although the identical call returns 204 at +0 ms once the child delete
  has returned ([S3T-NAMESPACE-027](#s3t-namespace-027), -039); the 'not empty' BadRequestException must be requeued, not treated as
  terminal, and no delete path ever returned ConflictException.
  - handling_ref: `test/e2e/conftest.py:102-106; test/e2e/conftest.py:141-142; generator.yaml:99-104;
    generator.yaml:170-175`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-NAMESPACE-002](#s3t-namespace-002) | CreateNamespace 200 body = {tableBucketARN, namespace:[name]} only; GetNamespace adds 5 read-only members, no ARN/tags/status | get_after_create_latency_ms=12, created_at_minus_client_clock_s=0.014 | 1 |
| [S3T-NAMESPACE-018](#s3t-namespace-018) | GetNamespace and ListNamespaces reflect CreateNamespace and DeleteNamespace at +0 ms in 20/20 trials (no read-after-write window) | create_latency_ms.min=30, create_latency_ms.max=74, create_latency_ms.n=20, get_after_create_latency_ms.min=10, get_after_create_latency_ms.max=13, get_after_create_latency_ms.n=20, list_latency_ms.min=9, list_latency_ms.max=49, list_latency_ms.n=20, delete_latency_ms.min=23, delete_latency_ms.max=29, delete_latency_ms.n=20, created_at_minus_client_clock_s.min=-0.023, created_at_minus_client_clock_s.max=-0.019, created_at_minus_client_clock_s.n=20, get_after_create_200_trials=20, list_after_create_present_trials=20, list_prefix_exact_one_trials=20, delete_204_trials=20, get_after_delete_404_trials=20, list_after_delete_absent_trials=20, get_lag_s=null, list_lag_s=null, get_gone_lag_s=null, list_gone_lag_s=null | 20 |
| [S3T-NAMESPACE-019](#s3t-namespace-019) | CreateNamespace at +0 ms after CreateTableBucket -> 200 in 20/20; CreateTable at +0 ms after CreateNamespace -> 200 in 20/20 | create_bucket_latency_ms.min=1444, create_bucket_latency_ms.max=2931, create_bucket_latency_ms.n=20, create_namespace_latency_ms.min=49, create_namespace_latency_ms.max=88, create_namespace_latency_ms.n=20, create_table_latency_ms.min=310, create_table_latency_ms.max=440, create_table_latency_ms.n=20, delete_table_latency_ms.min=139, delete_table_latency_ms.max=187, delete_table_latency_ms.n=20, delete_bucket_latency_ms.min=445, delete_bucket_latency_ms.max=523, delete_bucket_latency_ms.n=20, namespace_at_zero_200_trials=20, table_at_zero_200_trials=20, create_ns_lag_s=null, create_table_lag_s=null | 20 |
| [S3T-NAMESPACE-020](#s3t-namespace-020) | Reads concurrent with DeleteNamespace+CreateNamespace of the same name: 120 calls, codes ['200', 'NotFoundException'], no ConflictException | read_calls=120, non_200_by_op.get.NotFoundException=24, non_200_by_op.listtables.NotFoundException=24 | 10 |
| [S3T-NAMESPACE-021](#s3t-namespace-021) | Second CreateNamespace with the same name -> ConflictException (HTTP 409) on 6/6 attempts over 10 s; existing namespace untouched | duplicate_create_latency_ms.min=10, duplicate_create_latency_ms.max=21, duplicate_create_latency_ms.n=6, attempts=6 | 6 |
| [S3T-NAMESPACE-023](#s3t-namespace-023) | CreateNamespace at +0 ms after DeleteNamespace of the same name -> 200 in 5/5; new namespaceId, newer createdAt, same tableBucketId | recreate_latency_ms.min=29, recreate_latency_ms.max=51, recreate_latency_ms.n=5, immediate_200_rounds=5 | 5 |
| [S3T-NAMESPACE-024](#s3t-namespace-024) | DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6 retries over 10 s; table and namespace untouched | delete_nonempty_latency_ms.min=11, delete_nonempty_latency_ms.max=59, delete_nonempty_latency_ms.n=6, attempts=6 | 6 |
| [S3T-NAMESPACE-027](#s3t-namespace-027) | DeleteNamespace at +0 ms after DeleteTable of its last table -> 204 in 20/20; after RenameTable moving the last table out -> 204 in 10/10 | n_delete_table_trials=20, n_rename_trials=10, delete_ns_immediate_204_after_delete_table=20, delete_ns_immediate_204_after_rename=10, time_to_204_s_after_delete_table=null, time_to_204_s_after_rename=null, delete_table_latency_ms.min=130, delete_table_latency_ms.max=167, delete_table_latency_ms.n=20, rename_table_latency_ms.min=155, rename_table_latency_ms.max=202, rename_table_latency_ms.n=10 | 30 |
| [S3T-NAMESPACE-028](#s3t-namespace-028) | CreateTable vs DeleteNamespace fired concurrently (20 trials): CreateTable NotFoundException \| DeleteNamespace 204 in 20/20; 0 orphan tables | both_ok=0, orphans=0, outcome_pairs.NotFoundException \| HTTP 204=20 | 20 |
| [S3T-NAMESPACE-031](#s3t-namespace-031) | Explicit deny (bucket or IAM session policy) -> AccessDeniedException 403 on all 6 namespace ops; foreign-account ARN -> ForbiddenException | assume_role_success_after_create_role_s=9.1 | 1 |
| [S3T-NAMESPACE-033](#s3t-namespace-033) | Back-to-back CreateNamespace or DeleteNamespace (~30/s): 16 of 30 succeed, rest 429 'Too Many Requests' (1 burst each); 30 parallel Gets 200 | get_burst_wall_s=[0.16, 0.09], get_latency_ms_min_max=[[11,156], [10,78]], create_30_wall_s=0.57, delete_30_wall_s=0.51, create_latency_ms_min_max=[2, 37], delete_latency_ms_min_max=[2, 52] | 120 |
| [S3T-NAMESPACE-034](#s3t-namespace-034) | 5 concurrent CreateNamespace for the same fresh name x6 trials: {"200": 6, "ConflictException/409": 24}; exactly one 200 per trial | race_wall_s=[0.12, 0.11, 0.14, 0.14, 0.15, 0.12] | 6 |
| [S3T-NAMESPACE-035](#s3t-namespace-035) | CreateNamespace back-to-back: 14/20 then 429 (2 setup calls used budget); retry every 250 ms succeeds after 0.54 s; after a 5 s pause 16/20 | create_wall_s=0.45, create_latency_ms_min_max=[2, 40], first_429_index_create=14, refill_lag_s_create=0.54, ok_after_5s_pause=16, first_429_index_delete=16, refill_lag_s_delete=0.54 | 71 |
| [S3T-NAMESPACE-039](#s3t-namespace-039) | Parallel finalizers: DeleteTable\|\|DeleteNamespace -> 400 'not empty' 6/8; DeleteNamespace\|\|DeleteTableBucket -> 400 6/8; +0 ms retry 204 | set1_parent_rejected=6, set2_parent_rejected=6, set1_namespace_left=0, set2_bucket_left=0 | 8 |
| [S3T-NAMESPACE-040](#s3t-namespace-040) | CreateNamespace 429 after ~16 back-to-back is per account-region (all buckets), not shared with Delete/CreateTable/Tag; <= 4/s is safe | drain_ok_before_429=[16, 16, 16, 16, 16, 17, 16, 16, 16, 16, 17, 16, 17, 22, 11], same_region_other_bucket_X_patterns=[TTTTT, TTTTT, TTTTT, TTTTT, ....., .....], other_region_X_patterns=[....., ....., .....], delete_namespace_X_patterns=[....., ....., .....], create_table_X_patterns=[..., ..., ...], tag_bucket_X_patterns=[..., ..., ...], refill_ok_of_13_at_4_per_s=[12, 12, 12], paced_le_5_per_s_ok=106 | 3 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `namespace.md`: unverified, single-trial or interpretive items worth a future probe. One line each.

- [S3T-NAMESPACE-010](#s3t-namespace-010): the ListNamespaces token as a live name-ordered cursor (created-after names appear, deleted ones
  vanish) rests on one run (n=1, confidence medium); re-run with several cursor positions and a create that sorts
  before the cursor.
- [S3T-NAMESPACE-020](#s3t-namespace-020): reads concurrent with a delete+recreate of one name saw only 200/NotFoundException over 120 calls
  (confidence medium); not yet tested whether a GetNamespace can return the OLD `namespaceId` after the recreate 200.
- [S3T-NAMESPACE-035](#s3t-namespace-035): the 0.54 s refill after a 429 was measured in one retry series for create and one for delete
  (`refill_lag_s_*` are single numbers); no range exists for the throttle recovery time.
- [S3T-NAMESPACE-040](#s3t-namespace-040): "fixed window of ~16 per second rather than a smooth token bucket" is the prober's reading of the
  batch pass/fail pattern (drains 11-22, mode 16); the window length itself was not measured independently of the
  250 ms retry cadence.
- [S3T-NAMESPACE-040](#s3t-namespace-040): the CreateNamespace limiter was shown to be separate from DeleteNamespace, CreateTable and
  TagResource in 3/3 runs each, but the other-bucket sharing pattern was `.....` in 2 of 6 trials (where the probe name
  had also recovered); a longer same-region two-bucket trial would tighten "shared by all table buckets".
- [S3T-NAMESPACE-031](#s3t-namespace-031): `assume_role_success_after_create_role_s` 9.1 s is n=1; the e2e role-propagation budget for Deny
  tests is unmeasured as a range.
- [S3T-NAMESPACE-002](#s3t-namespace-002): `get_after_create_latency_ms` 12 and `created_at_minus_client_clock_s` 0.014 are n=1 here; the
  n=20 values live in -018 (10-13 ms, -0.023 to -0.019 s) and should be the ones quoted.
- [S3T-NAMESPACE-024](#s3t-namespace-024) / -039: the "not empty" 400 was always cleared by the very next call once the child delete had
  returned; no trial measured a DeleteNamespace issued while DeleteTable was still in flight for a table with many
  snapshots or a KMS-encrypted table (the sibling `table.md` DeleteTable timings may differ).
- [S3T-NAMESPACE-015](#s3t-namespace-015): only `aws`, `awsfoo`, `aws_data`, `Aws_data` were tried for the reserved prefix; `amzn`/`amazon`
  prefixes and the `aws` substring directly after a leading digit were not probed.
- [S3T-NAMESPACE-029](#s3t-namespace-029): the other-account ForbiddenException case used the documentation's sample account id, not a real
  foreign account with a cross-account bucket policy allowing the caller.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-namespace-001"></a>**S3T-NAMESPACE-001** | response-fidelity | low | confirmed | Fresh table bucket: ListNamespaces -> HTTP 200, raw body 'namespaces': []; GetNamespace('default') -> NotFoundException | [S3T-NAMESPACE-025](#s3t-namespace-025), [S3T-NAMESPACE-026](#s3t-namespace-026), [S3T-NAMESPACE-028](#s3t-namespace-028), [S3T-NAMESPACE-029](#s3t-namespace-029) | - |
| <a id="s3t-namespace-009"></a>**S3T-NAMESPACE-009** | request-validation | low | confirmed | ListNamespaces token: garbage -> BadRequestException; another bucket's token -> BadRequestException... | [S3T-NAMESPACE-008](#s3t-namespace-008), [S3T-NAMESPACE-010](#s3t-namespace-010), [S3T-NAMESPACE-032](#s3t-namespace-032), [S3T-NAMESPACE-005](#s3t-namespace-005) | - |
| <a id="s3t-namespace-010"></a>**S3T-NAMESPACE-010** | eventual-consistency | low | confirmed | ListNamespaces token is a live name-ordered cursor: names created after the cursor appear, deleted ones vanish; 5 pages 200, no dups (n=1) | [S3T-NAMESPACE-008](#s3t-namespace-008), [S3T-NAMESPACE-009](#s3t-namespace-009), [S3T-NAMESPACE-032](#s3t-namespace-032), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-018](#s3t-namespace-018), [S3T-NAMESPACE-019](#s3t-namespace-019), [S3T-NAMESPACE-020](#s3t-namespace-020), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-027](#s3t-namespace-027) | - |
| <a id="s3t-namespace-016"></a>**S3T-NAMESPACE-016** | request-validation | low | confirmed | CreateNamespace with non-ASCII (e-acute, kanji, zero-width space, combining mark, fullwidth digit) -> BadRequestException, nothing created | [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-011](#s3t-namespace-011), [S3T-NAMESPACE-012](#s3t-namespace-012), [S3T-NAMESPACE-013](#s3t-namespace-013), [S3T-NAMESPACE-014](#s3t-namespace-014), [S3T-NAMESPACE-015](#s3t-namespace-015) | - |
| <a id="s3t-namespace-020"></a>**S3T-NAMESPACE-020** | error-code | low | confirmed | Reads concurrent with DeleteNamespace+CreateNamespace of the same name: 120 calls, codes ['200', 'NotFoundException'], no ConflictException | [S3T-NAMESPACE-018](#s3t-namespace-018), [S3T-NAMESPACE-019](#s3t-namespace-019), [S3T-NAMESPACE-023](#s3t-namespace-023), [S3T-NAMESPACE-027](#s3t-namespace-027), [S3T-NAMESPACE-010](#s3t-namespace-010) | - |
| <a id="s3t-namespace-036"></a>**S3T-NAMESPACE-036** | other | low | confirmed | Doc claim C106 TRUE: Duplicate CreateNamespace -> ConflictException 409; duplicate CreateTable (same namespace+name) -> ConflictException | [S3T-NAMESPACE-021](#s3t-namespace-021), [S3T-TABLE-064](table-metadata-location.md#s3t-table-064), [S3T-TABLE-274](table.md#s3t-table-274), [S3T-TABLE-276](table.md#s3t-table-276), [S3T-TABLE-050](table.md#s3t-table-050) | - |
| <a id="s3t-namespace-037"></a>**S3T-NAMESPACE-037** | other | low | confirmed | Doc claim C110 TRUE: DeleteNamespace with a table inside -> BadRequestException (not Conflict); DeleteTableBucket with a namespace -> same | [S3T-NAMESPACE-024](#s3t-namespace-024), [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019), [S3T-TABLEBUCKET-044](tablebucket.md#s3t-tablebucket-044) | - |
| <a id="s3t-namespace-038"></a>**S3T-NAMESPACE-038** | other | low | confirmed | Doc claim C038 TRUE: ListNamespaces maxNamespaces=1 -> 1 namespace + continuationToken; following it returns the rest, 0 duplicates | [S3T-NAMESPACE-032](#s3t-namespace-032), [S3T-NAMESPACE-010](#s3t-namespace-010), [S3T-TABLE-293](table.md#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-TABLE-168](table.md#s3t-table-168), [S3T-TABLE-092](table.md#s3t-table-092), [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](table.md#s3t-table-091) | - |
| <a id="s3t-namespace-041"></a>**S3T-NAMESPACE-041** | read-gap | medium | confirmed | CreateNamespace returns only {tableBucketARN, namespace}; namespaceId/tableBucketId/ownerAccountId/createdBy/createdAt only in Get/List | [S3T-NAMESPACE-002](#s3t-namespace-002), [S3T-NAMESPACE-005](#s3t-namespace-005), [S3T-NAMESPACE-003](#s3t-namespace-003), [S3T-NAMESPACE-004](#s3t-namespace-004), [S3T-NAMESPACE-025](#s3t-namespace-025) | [S3T-NAMESPACE-002](#s3t-namespace-002) |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
