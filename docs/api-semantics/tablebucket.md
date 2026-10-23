<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# TableBucket (lifecycle, identity, errors, tags, limits)
_CreateTableBucket/GetTableBucket/DeleteTableBucket/ListTableBuckets: identity and ARN construction, the ~10 min post-delete name reservation, not-empty delete as 400, tags, throttling, account limits; catch-all for TableBucket findings not covered above._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `tablebucket.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the TableBucket core surface: CreateTableBucket, GetTableBucket, DeleteTableBucket,
ListTableBuckets and TagResource/UntagResource/ListTagsForResource on a bucket ARN, plus naming and tag validation,
throttling and account limits. There is no state machine: CreateTableBucket is synchronous, returns only `{arn}`, and
GetTableBucket succeeds on the first poll with the same six members every time (`arn`, `createdAt`, `name`,
`ownerAccountId`, `tableBucketId`, `type`) and no status enum ([S3T-TABLEBUCKET-003](#s3t-tablebucket-003), -004, -121); the ARN is
deterministic (`arn:aws:s3tables:<region>:<account>:bucket/<name>`) and can be built before the call (-003, -167).
DeleteTableBucket is synchronous too (204, Get 404 on the first poll, -014) but the name stays in a server-side
"transitional state" for ~10 min, invisible to Get and List and reported only as a ConflictException on a same-name
Create (-115, -158, -184). The bucket body has no mutable member: Get returns byte-identical keys before and after
sub-resource Puts and after child namespaces come and go (-060, -002); encryption, storage class, policy, maintenance
and replication live behind their own Get APIs, documented in the sibling documents under "Where next". Tags are the
only write against the bucket ARN covered here: TagResource is an additive upsert, removal is UntagResource, and both
read back consistently at +0 ms (-027, -028, -124).

### Rules

1. Never treat a same-name CreateTableBucket 409 after a delete as "exists" or as terminal: after DeleteTableBucket
   (204, Get 404 at once) the same name returns ConflictException 409 "transitional state because of a previous
   deletion attempt" on every retry until ~570-595 s and succeeds at 602-609 s; retries do not extend the window and
   bucket content does not lengthen it; the ARN comes back identical but `tableBucketId` is new and the bucket reads
   fresh defaults (policy 404, default maintenance, only Create tags) - requeue, then re-apply every sub-resource
   ([S3T-TABLEBUCKET-115](#s3t-tablebucket-115), -158, -184).
2. CreateTableBucket ConflictException has two meanings with one code and HTTP status: "already exists, and you own
   it" (bucket present, Get 200) vs "transitional state" (Get/Delete/Tag 404, List(prefix) empty); only the message or
   a GetTableBucket on the constructed ARN tells them apart: 409 + Get 200 => adopt, 409 + Get 404 => requeue
   ([S3T-TABLEBUCKET-159](#s3t-tablebucket-159), -160, -017).
3. DeleteTableBucket on a bucket holding a table, or only an empty namespace, returns BadRequestException 400 "The
   bucket that you tried to delete is not empty." (never ConflictException) on every retry; the identical call returned
   204 0.549 s after DeleteNamespace; the rejected delete leaves no mark on the bucket - requeue the parent's delete
   until the children are gone instead of going Terminal ([S3T-TABLEBUCKET-019](#s3t-tablebucket-019), -132, -070).
4. Serialize tag writes per bucket: 10 concurrent TagResource calls on one bucket produced 8-9/10
   InternalServerErrorException 500 (never 409) and some 500-failed calls still applied their tag; the same calls
   retried sequentially all returned 200 - treat the 500 as retryable and re-read tags before deciding
   ([S3T-TABLEBUCKET-070](#s3t-tablebucket-070)).
5. TagResource is an additive upsert (existing keys kept, same key overwritten, new key added); removal is only via
   UntagResource; UntagResource of absent keys -> 204, `tagKeys=[]` -> BadRequestException 400, `TagResource({})` ->
   200 no-op; ListTags on a never-tagged bucket -> 200 with an empty map, returns only user tags with no pagination
   token; bucket and table tags are independent ([S3T-TABLEBUCKET-027](#s3t-tablebucket-027), -028, -007, -156, -032, -151, -153, -154, -155).
6. Tag length violations on TagResource return the NON-modeled ValidationException 400 (129-char key, 257-char value,
   empty key) and the call is all-or-nothing, while the same tags on CreateTableBucket return BadRequestException 400
   and create no bucket; keys are case-sensitive and `aws:`/`AWS:` keys are BadRequestException on both paths
   ([S3T-TABLEBUCKET-030](#s3t-tablebucket-030), -025, -029, -152).
7. The 50-tag limit is cumulative per bucket and each call is atomic: 30 + 25 -> 400 with none applied, +20 -> 200 at
   50, a 51st key -> 400 with the 50 unchanged, updating a key at 50 -> 200; 50 keys in one call -> 200, 51 -> 400 with
   0 tags ([S3T-TABLEBUCKET-031](#s3t-tablebucket-031), -148, -149, -150).
8. Only NotFoundException means "gone", and it is what a second DeleteTableBucket (and Get/ListTags/Tag/Untag) on a
   deleted ARN returns - 404, not 204; every malformed identifier form (plain name, wrong service, trailing slash or
   space, other region or partition, empty region or account) -> BadRequestException 400 before any lookup, a
   foreign-account ARN -> ForbiddenException 403, and a wrong-region ARN -> 400 "Region does not match" even when the
   name exists nowhere ([S3T-TABLEBUCKET-015](#s3t-tablebucket-015), -162, -038, -127, -164).
9. CreateTableBucket enforces the documented naming page exactly (3-63 chars, lowercase, digits, hyphens; every
   reserved prefix incl. any `aws` prefix and every reserved suffix -> BadRequestException 400); GetTableBucket does
   not apply the `aws` prefix rule and tolerates `_` (404 instead of 400), so a name rejected by Create can still read
   as a plain 404 ([S3T-TABLEBUCKET-022](#s3t-tablebucket-022)).
10. The same name can exist in two regions: both Creates 200, ARNs differ only in the region segment, distinct
    `tableBucketId`s, List and tags per region; the other region accepts the name inside this region's post-delete
    window ([S3T-TABLEBUCKET-125](#s3t-tablebucket-125), -184).
11. Every 4xx body is restJson `{"message"}` with the code only in `x-amzn-ErrorType`; 400 messages name the field
    class (name, SSE algorithm, storage class, KMS key ARN, tags) not the rule, and the 404 message does not say which
    ARN segment was wrong - match on the code, use the message only to classify ([S3T-TABLEBUCKET-026](#s3t-tablebucket-026)).
12. Tag ops on a table ARN whose uuid does not exist: ListTags and Tag -> NotFoundException 404, but UntagResource ->
    InternalServerErrorException 500 in 3/3 tries (also under a missing bucket), so an untag-then-tag sync on a
    vanished table surfaces as a 500, not a 404 ([S3T-TABLEBUCKET-033](#s3t-tablebucket-033), -043).
13. ListTableBuckets: `prefix` is a plain string prefix, results sort lexically by name (not createdAt) and are stable
    across `maxBuckets=1` pages; a bad or stale `continuationToken` -> BadRequestException 400, a token reused after a
    delete skips the deleted item (200); `maxBuckets` 0/-1 -> 400 but 1001 -> 200; the last page omits
    `continuationToken` (key absent); invalid prefixes -> 200 empty ([S3T-TABLEBUCKET-040](#s3t-tablebucket-040), -041, -069, -168, -169).
14. Throttling is TooManyRequestsException 429, not 503: a same-name Create storm at ~8/s was throttled in runs of 4-6
    after 5-9 accepted (~5 TPS estimate, shared account); bursts of 5 then 4 concurrent distinct-name Creates and 60
    sequential Creates up to 62 buckets all 200 with no quota error; no read throttle was ever observed
    ([S3T-TABLEBUCKET-018](#s3t-tablebucket-018), -068, -071, -179, -165).
15. Do not poll for eventual consistency on this surface: after Create, Get, List(prefix) and ListTags were all 200 on
    poll #1 in 5/5 with no flapping; after Delete, Get was 404 and List dropped the bucket on the first poll; tag
    read-after-write was consistent at +0 ms in 10/10 ([S3T-TABLEBUCKET-121](#s3t-tablebucket-121), -014, -161, -124, -157).

### Timing ranges

- CreateTableBucket latency: 1403-1803 ms, n=60 ([S3T-TABLEBUCKET-179](#s3t-tablebucket-179)); 1470-1580 ms, n=5 (-003); 1454-1678 ms, n=5
  (-121); 1454-2049 ms for 10 calls released in 5 concurrent same-name pairs (-017); 1253-1911 ms, n=2 across two
  regions (-125). `createdAt` minus the local pre-call time: 1.42-1.51 s, n=5 (-004).
- First successful GetTableBucket after Create: 0.03-0.04 s, n=5 (-004); Get, List(prefix) and ListTagsForResource
  all first-OK at 0 ms offset in 5/5 trials, 5 lockstep polls each with no flapping (-121).
- Post-delete visibility: Get NotFoundException at 0.02 s and List(prefix) gone at 0.009 s (n=1, -014); <= 0.02 s on
  the first poll (n=2, -161) - six Gets after the first 404 never returned 200 again (-014).
- Post-delete name reservation (same-name Create): last "transitional state" 409 at 570.4-594.6 s, first 200 at
  602.1-609.4 s after the delete, 6 buckets across -115 (n=1, retries every 5-10 s), -158 (n=3, retries every 15 s,
  one name first retried only at +420 s) and -184 (n=2, retries every 30 s). A reconciler needs a >= 10 min requeue
  budget for a delete-and-recreate of the same name.
- Same name created in the other region inside the window: 0.97-1.25 s after the delete, n=2 (-184); 2.54 s, n=1
  (-159).
- Delete after the last child: only one retry measured, 0.549 s after DeleteNamespace returned, and it was 204
  (n=1, -019) - no range exists, do not derive a wait from it.
- TagResource 44-66 ms and UntagResource 44-65 ms, n=10 each, with the +0 ms and +500 ms ListTagsForResource reads
  consistent in 10/10 both times (-124); under 10-way concurrency TagResource latency spread to 29-631 ms with a burst
  wall of 0.18-0.63 s, n=20 calls in 2 bursts, and 8-9/10 calls per burst were 500s (-070).
- GetTableBucket latency 5-33 ms over 30 back-to-back calls at ~155 req/s with 0 throttled (n=30, -165).
- CreateTableBucket throttle: 500 same-name calls at ~100 ms spacing over 60 s -> 186 TooManyRequestsException 429 in
  cycles of 5-9 accepted then 4-6 throttled (~5 TPS estimate; one run, shared account; -018).

### Handling gaps

- [S3T-TABLEBUCKET-019](#s3t-tablebucket-019) (suspect-bug, not yet handled): today `terminal_codes: [BadRequestException]` on TableBucket
  (generator.yaml:267-268) and no delete hook, so DeleteTableBucket on a bucket that still holds a namespace or table
  sets ACK.Terminal on the first attempt. Evidence: the 400 "not empty" is deterministic while children exist (3/3
  retries), never ConflictException, the identical call returned 204 0.549 s after DeleteNamespace (n=1), and the
  rejected delete leaves no mark (-132). Fix direction: requeue this message, or hold the finalizer until the child
  CRs are gone.
- [S3T-TABLEBUCKET-030](#s3t-tablebucket-030) (suspect-bug, not yet handled): today `syncTags` (hooks.go:411-471) sends the desired keys
  and values unchecked via UntagResource then TagResource, the CRD has no tag length bounds and ValidationException is
  not in `terminal_codes` (generator.yaml:267-268). Evidence: a 129-char key, 257-char value or empty key on
  TagResource -> ValidationException 400 (all-or-nothing), while the same tags on CreateTableBucket ->
  BadRequestException 400 (-025, Terminal); a `spec.tags` change after create that breaks the bounds requeues forever.
- [S3T-TABLEBUCKET-159](#s3t-tablebucket-159) (suspect-bug, not yet handled): today `is_arn_primary_key: true` (generator.yaml:192), so the
  read path has no identifier until the first successful Create patches the status ARN, and ConflictException is not
  in `terminal_codes`, so a 409 on Create is retried as recoverable. Evidence: a pre-existing name -> 409 "already
  exists, and you own it" on every retry while Get on the constructed ARN is 200; a name deleted < ~10 min ago -> 409
  "transitional state" for 602-609 s while Get is 404 (-115, -158, -184). A CR for a pre-existing bucket, or one whose
  status patch was lost after a successful Create, requeues on Create indefinitely and never adopts. Fix direction:
  derive the ARN (-003) in the read path, or on 409 Get the derived ARN and adopt on 200 / requeue on 404.
- [S3T-TABLEBUCKET-004](#s3t-tablebucket-004) (partial): today `CreatedAt`, `Type` and `TableBucketID` are `from: GetTableBucket` read-only
  fields (generator.yaml:196-212) filled by the post-create read; no requeue-after-create exists beyond the Synced
  nudge when Policy or MaintenanceConfiguration is set (generator.yaml:270-275). Evidence: Create is synchronous (first
  Get OK at 0.03-0.04 s, n=5) with no status enum, so this matches; the partial gap is the `ownerAccountId` body
  member, which has no `from:` field and is not read back from the bucket.
- [S3T-TABLEBUCKET-122](#s3t-tablebucket-122) (partial): today the read path is GetTableBucket by the status ARN only; TableBucket declares
  no `list_operation`/`match_fields` and no custom find (generator.yaml:191-285). Evidence: the
  ListTableBuckets(prefix=name) summary equals GetTableBucket field by field in 5/5 (raw `createdAt` identical), so a
  name-based find or adoption from one List call is possible without a Get; it would also resolve the -159 case.

### Where next

- Sibling documents for the bucket's sub-resources, all read through their own Get APIs and all reset to defaults by a
  delete-and-recreate (-158): `tablebucket-encryption-storage.md` (SSE-S3/SSE-KMS, storage class, the
  ValidationException-vs-BadRequestException split), `tablebucket-policy.md` (resource policy, the two 403 codes, the
  first-sync deletion hazard), `tablebucket-maintenance-metrics.md` (maintenance merge-patch, metrics configuration),
  `tablebucket-replication.md` (bucket-level replication rules, versionToken locking).
- `namespace.md` carries the child side of rule 3 (DeleteNamespace "not empty" 400, strong consistency, the ~16 req/s
  per account-region limiter); `table.md` carries the table-ARN tag behavior referenced by rules 6 and 12
  (ValidationException on length, UntagResource 500 on vanished tables).
- Long catalogues render in `details/<id>.md` rather than inline: the identifier-form matrix for Get/ListTags/Delete
  ([S3T-TABLEBUCKET-038](#s3t-tablebucket-038)), the ARN-form matrix for the tag ops (-033), the naming accept/reject matrix (-022), the 400
  message classes (-026), the continuation-token and prefix matrix (-041), the post-window re-created defaults (-158).
- Open experiments (STATUS.md): the GetTableBucket read-throttle cell (~200 req/s) is out of reach under the lab's
  burst cap, so no read throttle code has been observed (-165, UNTESTABLE); no AWS-managed `aws`-type bucket exists in
  the account, so `type=aws` listings are empty and the `aws-s3` pseudo-bucket is a 404 (-130, -042, -166); the ~5 TPS
  CreateTableBucket limit is an estimate from one run on a shared account (-018). The post-delete window was
  re-verified independently in Phase 8 (-159 CONFIRMED, precision recorded as -184). One leftover lab bucket cannot be
  deleted by any principal because of a Deny bucket policy; that path belongs to `tablebucket-policy.md`
  ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007)), not here.
<!-- preserved:end -->

## At a glance

- canonical findings: 63 (high 18 / medium 16 / low 29); duplicates folded into the appendix: 2
- handling: handled 19 · partial 3 · tracked 0 · unhandled 38 · suspect-bug 3 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 3 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: other 18, identity 8, delete-semantics 6, error-code 6, tag-semantics 6, quota-limit 5,
  request-validation 4, idempotency 3, eventual-consistency 2, response-fidelity 2, async-state-machine 1,
  cross-region 1, read-gap 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateNamespace | create | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketStorageClass | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListTableBuckets | list | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

- <a id="s3t-tablebucket-004"></a>**S3T-TABLEBUCKET-004** `async-state-machine` · impact high · partially handled · verified 2026-10-13
  **CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms after Create, no status/state field exists**
  For 5 fresh buckets GetTableBucket issued immediately after CreateTableBucket returned (poll interval 50 ms)
  returned HTTP 200 on attempt sequence ["['OK']"]; the body always had exactly ['arn', 'createdAt', 'name',
  'ownerAccountId', 'tableBucketId', 'type'] (no status/state enum). type=['customer'], ownerAccountId ==
  caller account in 5/5, tableBucketId is a UUID in 5/5, createdAt minus the local pre-call time was [1.5,
  1.51, 1.43, 1.42, 1.47] s.
  - ACK: none, hook:sdk_create_post_set_output · ops: CreateTableBucket, GetTableBucket · fields: createdAt,
    ownerAccountId, tableBucketId, type
  - repro: t0=now; CreateTableBucket(name); GetTableBucket(arn) every 50 ms until 200; inspect body
  - measurements: first_get_ok_s=[0.04, 0.03, 0.03, 0.03, 0.03], createdAt_minus_t0_s=[1.504, 1.509, 1.432,
    1.419, 1.468], n=5
  - handling: partially handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304` - see Handling gaps
  - related: [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-002](#s3t-tablebucket-002),
    [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-042](#s3t-tablebucket-042) · hypotheses: H-B-002 · evidence:
    tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-002: no CREATING/DELETING phase was observable; the controller needs no requeue after
    create and can fill Status (ownerAccountId, createdAt, tableBucketId, type) from one GetTableBucket right
    after Create.
  - full notes: [details/S3T-TABLEBUCKET-004.md](details/S3T-TABLEBUCKET-004.md)

## Field matrix

C = accepted by the create input (CreateTableBucket), U = by the update input, R = present in the read output.

| leaf | C | U | R | type |
| --- | --- | --- | --- | --- |
| arn | - | - | x | string |
| createdAt | - | - | x | timestamp |
| encryptionConfiguration | x | - | - | struct:EncryptionConfiguration |
| kmsKeyArn | x | - | - | string |
| name | x | - | x | string |
| ownerAccountId | - | - | x | string |
| sseAlgorithm | x | - | - | enum:SSEAlgorithm |
| storageClass | x | - | - | enum:StorageClass |
| storageClassConfiguration | x | - | - | struct:StorageClassConfiguration |
| tableBucketId | - | - | x | string |
| tags | x | - | - | map<string,string> |
| type | - | - | x | enum:TableBucketType |

## Identity and lookup

- <a id="s3t-tablebucket-003"></a>**S3T-TABLEBUCKET-003** `identity` · impact high · handled · verified 2026-10-15, re-verified
  **CreateTableBucket returns only {arn}; ARN is arn:aws:s3tables:<region>:<account>:bucket/<name>, constructible before the call**
  In 5/5 creates the CreateTableBucket response body (HTTP 200) had exactly the keys ['arn'] and the arn
  equalled the string built in advance from (region, caller account, name) with resource segment
  'bucket/<name>' (no suffix, no trailing slash). GetTableBucket(tableBucketARN=<constructed arn>) returned
  HTTP 200 with keys ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'], and its arn was
  byte-identical to the Create response arn in 5/5 runs.
  - ACK: is_arn_primary_key, custom_find · ops: CreateTableBucket, GetTableBucket · fields: arn, name,
    tableBucketARN
  - repro: expected='arn:aws:s3tables:<region>:<account>:bucket/<name>'; CreateTableBucket(name) -> compare
    arn; GetTableBucket(expected)
  - measurements: create_latency_ms=[1580, 1556, 1477, 1470, 1514], n=5
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98`
  - related: [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-022](#s3t-tablebucket-022),
    [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033), [S3T-TABLEBUCKET-160](#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](#s3t-tablebucket-167),
    [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) · hypotheses: H-B-001, H-B-066 · evidence:
    tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-001 (deterministic ARN) and H-B-066 (Create output carries only arn; Status fields
    other than arn must come from an immediate GetTableBucket). The same arn string was accepted verbatim as
    tableBucketARN by GetTableBucket and DeleteTableBucket (see later steps for resourceArn on the tag...
  - full notes: [details/S3T-TABLEBUCKET-003.md](details/S3T-TABLEBUCKET-003.md)

- <a id="s3t-tablebucket-033"></a>**S3T-TABLEBUCKET-033** `identity` · impact high · handled · verified 2026-10-13
  **Tag ops on a table ARN with an unknown uuid: ListTags/Tag -> NotFoundException 404 but UntagResource -> InternalServerErrorException 500**
  ListTagsForResource / TagResource({a:b}) / UntagResource([a]) per ARN form: s3-general-purpose-bucket ->
  BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The
  specified resourceArn is not valid.' | BadRequestException HTTP 400 'The specified resourceArn is not
  valid.'; namespace-path -> BadRequestException HTTP 400 'The specified resourceArn is not valid.' |
  BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The
  specified resourceArn is not valid.'; glue-table -> BadRequestException HTTP 400 'The specified resourceArn
  is not valid.' | BadRequestException HTTP 400 'The specified resourceArn is not valid.' |
  BadRequestException HTTP 400 'The specified resourceArn is not valid.'; table-arn-missing-uuid ->
  NotFoundException HTTP 404 'The specified table does not exist.' | NotFoundException HTTP 404 'The specified
  table does not exist.' | InternalServerErrorException HTTP 500 'An internal error occurred. Try again.';
  table-arn-real -> HTTP 200 {"tags": {"t": "1"}} | HTTP 200 | HTTP 204 ; bucket-arn-missing ->
  NotFoundException HTTP 404 'The specified bucket does not exist.' | NotFoundException HTTP 404 'The
  specified bucket does not exist.' | NotFoundException HTTP 404 'The specified bucket does not exist.';
  plain-name -> BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException
  HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP 400 'The specified resourceArn
  is not valid.'; bucket-arn-trailing-slash -> BadRequestException HTTP 400 'The specified resourceArn is not
  valid.' | BadRequestException HTTP 400 'The specified resourceArn is not valid.' | BadRequestException HTTP
  400 'The specified resourceArn is not valid.'.
  - ACK: terminal_codes, exceptions.404, tags.custom-sync · ops: ListTagsForResource, TagResource,
    UntagResource · fields: resourceArn
  - repro: ListTagsForResource/TagResource/UntagResource with an S3 bucket ARN, a namespace path, a Glue table
    ARN, a table ARN with an unknown uuid, a missing bucket ARN, a plain name
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-022](#s3t-tablebucket-022), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](#s3t-tablebucket-039), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003),
    [S3T-TABLEBUCKET-043](#s3t-tablebucket-043), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015) · hypotheses: H-B-048 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: Confirms H-B-048 for the wrong-type forms (S3 bucket ARN, namespace path, Glue table ARN, plain
    name, trailing slash -> BadRequestException 400 'The specified resourceArn is not valid.') and for a
    missing bucket ARN (404). SURPRISE: UntagResource on a well-formed table ARN whose uuid does not exist...
  - full notes: [details/S3T-TABLEBUCKET-033.md](details/S3T-TABLEBUCKET-033.md)

- <a id="s3t-tablebucket-038"></a>**S3T-TABLEBUCKET-038** `identity` · impact high · handled · verified 2026-10-13
  **GetTableBucket/Delete/ListTags: missing bucket -> NotFoundException/404...**
  Per identifier form (code/HTTP): valid-missing -> Get NotFoundException/404 | ListTags NotFoundException/404
  | Delete NotFoundException/404; plain-name -> Get BadRequestException/400 | ListTags BadRequestException/400
  | Delete BadRequestException/400; wrong-service-s3 -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; table-arn-under-bucket -> Get
  BadRequestException/400 | ListTags NotFoundException/404 | Delete BadRequestException/400; uppercase-name ->
  Get BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400;
  foreign-account -> Get ForbiddenException/403 | ListTags ForbiddenException/403 | Delete
  ForbiddenException/403; trailing-slash -> Get BadRequestException/400 | ListTags BadRequestException/400 |
  Delete BadRequestException/400; trailing-space -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; leading-space -> Get BadRequestException/400 |
  ListTags BadRequestException/400 | Delete BadRequestException/400; uppercase-arn-prefix -> Get
  BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400; other-region ->
  Get BadRequestException/400 | ListTags BadRequestException/400 | Delete BadRequestException/400;
  other-partition -> Get BadRequestException/400 | ListTags BadRequestException/400 | Delete
  BadRequestException/400; empty-region -> Get BadRequestException/400 | ListTags BadRequestException/400 |
  Delete BadRequestException/400; empty-account -> Get BadRequestException/400 | ListTags
  BadRequestException/400 | Delete BadRequestException/400; underscore-name -> Get NotFoundException/404 |
  ListTags NotFoundException/404 | Delete NotFoundException/404; aws-s3-pseudo-bucket -> Get
  NotFoundException/404 | ListTags NotFoundException/404 | Delete NotFoundException/404. The real bucket was
  untouched (GetTableBucket -> HTTP 200). Messages seen: ['The provided Partition does not match the Partition
  of the resource.', 'The provided Region does not match the Region of the resource.', 'The specified bucket
  does not exist.', "The specified bucket name isn't valid. Specify a different bucket name, and then try
  again.", 'The specified resourceArn is not valid.', 'The specified table bucket ARN is not valid.', 'The
  specified table does not exist.', 'Unauthorized'].
  - ACK: exceptions.404, terminal_codes, is_arn_primary_key · ops: GetTableBucket, DeleteTableBucket,
    ListTagsForResource · fields: tableBucketARN, resourceArn
  - repro: GetTableBucket/ListTagsForResource/DeleteTableBucket with each identifier form
    (parameter_validation=False)
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; generator.yaml:99-104; generator.yaml:170-175; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](table-policy.md#s3t-table-109), [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLE-026](table-policy.md#s3t-table-026),
    [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178),
    [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLE-081](table.md#s3t-table-081), [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163),
    [S3T-TABLEBUCKET-164](#s3t-tablebucket-164), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029) · hypotheses: H-B-034, H-B-035, H-B-067 ·
    evidence: tablebucket/error-taxonomy/arn-forms-get-delete-tags
  - notes: H-B-034: confirmed. H-B-035 (foreign account): ForbiddenException/403. H-B-067: trailing slash
    BadRequestException/400, trailing space BadRequestException/400, leading space BadRequestException/400,
    uppercase 'ARN:AWS:S3TABLES' BadRequestException/400, uppercase name BadRequestException/400. Other...
  - full notes: [details/S3T-TABLEBUCKET-038.md](details/S3T-TABLEBUCKET-038.md)

- <a id="s3t-tablebucket-040"></a>**S3T-TABLEBUCKET-040** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTableBuckets prefix is a plain string prefix; results sorted lexically by name (not createdAt), stable across maxBuckets=1 pages**
  Buckets created 1 s apart in the order ordz, ord-c, ord-a. ListTableBuckets(prefix=<run>-ord) x3 ->
  [['ackq-<run>-ord-a', 'ackq-<run>-ord-c', 'ackq-<run>-ordz'], ['ackq-<run>-ord-a', 'ackq-<run>-ord-c',
  'ackq-<run>-ordz'], ['ackq-<run>-ord-a', 'ackq-<run>-ord-c', 'ackq-<run>-ordz']]. Paging maxBuckets=1 ->
  pages [['ackq-<run>-ord-a'], ['ackq-<run>-ord-c'], ['ackq-<run>-ordz']]. ListTableBuckets(prefix=<no match>)
  -> HTTP 200 raw body {"tableBuckets":[]}.
  - ACK: custom_find, list_operation.match_fields · ops: ListTableBuckets · fields: prefix, maxBuckets,
    continuationToken, tableBuckets
  - repro: Create ordz, ord-c, ord-a -> ListTableBuckets(prefix=ord) -> page with maxBuckets=1
  - measurements: list_repeats=3, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-041](#s3t-tablebucket-041), [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125),
    [S3T-TABLEBUCKET-168](#s3t-tablebucket-168), [S3T-TABLEBUCKET-169](#s3t-tablebucket-169) · hypotheses: H-B-053, H-B-051 · evidence:
    tablebucket/error-taxonomy/arn-forms-get-delete-tags
  - notes: H-B-053 confirmed (3 repeats, paging with maxBuckets=1 gave ord-a, ord-c, ordz although creation
    order was ordz, ord-c, ord-a). H-B-051 prefix half confirmed: prefix=<run>-ord returned <run>-ordz too, so
    lookup-by-name must still filter for equality.

- <a id="s3t-tablebucket-047"></a>**S3T-TABLEBUCKET-047** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Deleted bucket name blocked with ConflictException for 604.7 s after DeleteTableBucket (n=2); re-created bucket: empty tags, default config**
  2 samples: CreateTableBucket(name, tags={ghost: 1}, INTELLIGENT_TIERING) -> DeleteTableBucket (HTTP 204) ->
  CreateTableBucket(name) every 5 s (15 s after 2 min). Attempts before success returned ['ConflictException']
  with message(s) ['The bucket is in a transitional state because of a previous deletion attempt. Try again
  later.']; the first HTTP 200 came [604.7, 604.7] s after the delete. Meanwhile
  GetTableBucket/ListTableBuckets(prefix) reported [('NotFoundException', '0')]. The re-created bucket: same
  ARN [True, True], new tableBucketId [True, True], tags [{}, {}], storage class ['STANDARD', 'STANDARD'],
  encryption [{'sseAlgorithm': 'AES256'}, {'sseAlgorithm': 'AES256'}].
  - ACK: requeue, custom_create, e2e-timing, is_arn_primary_key · ops: CreateTableBucket, DeleteTableBucket,
    GetTableBucket, ListTableBuckets, ListTagsForResource, GetTableBucketStorageClass · fields: name,
    tableBucketId, tags, storageClassConfiguration
  - repro: CreateTableBucket(name, tags, INTELLIGENT_TIERING) -> DeleteTableBucket -> CreateTableBucket(name)
    every 5 s until 200 -> Get/ListTags/GetStorageClass
  - measurements: blocked_window_s=[604.7, 604.7], attempts=[57, 57], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-131](#s3t-tablebucket-131) · hypotheses: H-B-027, H-B-028 · evidence:
    tablebucket/identity/recreate-after-delete-window
  - notes: Refutes H-B-027 (there is a post-delete name-reservation window; see blocked_window_s). H-B-028:
    refuted - the second life starts with empty tags and default storage class. Follow-up to the 60 s
    observation in tablebucket/idempotency/dup-create-delete-recreate.
  - full notes: [details/S3T-TABLEBUCKET-047.md](details/S3T-TABLEBUCKET-047.md)

## Idempotency

- <a id="s3t-tablebucket-015"></a>**S3T-TABLEBUCKET-015** `idempotency` · impact high · handled · verified 2026-10-15, re-verified
  **Second DeleteTableBucket on a deleted ARN -> NotFoundException (HTTP 404); Get/ListTags/Tag/Untag on the deleted ARN -> NotFoundException**
  On the ARN of a just-deleted bucket: DeleteTableBucket -> NotFoundException HTTP 404 'The specified bucket
  does not exist.' (x-amzn-ErrorType
  'NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', raw body {"message":"The
  specified bucket does not exist."}); GetTableBucket -> NotFoundException HTTP 404 'The specified bucket does
  not exist.'; ListTagsForResource -> NotFoundException HTTP 404 'The specified bucket does not exist.';
  TagResource -> NotFoundException HTTP 404 'The specified bucket does not exist.'; UntagResource ->
  NotFoundException HTTP 404 'The specified bucket does not exist.'. DeleteTableBucket on a well-formed ARN
  that never existed -> NotFoundException HTTP 404 'The specified bucket does not exist.'; GetTableBucket on
  it -> NotFoundException HTTP 404 'The specified bucket does not exist.'.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucket, GetTableBucket, ListTagsForResource,
    TagResource, UntagResource
  - repro: CreateTableBucket -> DeleteTableBucket -> DeleteTableBucket again;
    GetTableBucket/ListTagsForResource/TagResource/UntagResource on the ARN
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-016](#s3t-tablebucket-016), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159),
    [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033), [S3T-TABLEBUCKET-043](#s3t-tablebucket-043), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-161](#s3t-tablebucket-161),
    [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-101](tablebucket-replication.md#s3t-tablebucket-101) · hypotheses: H-B-020, H-B-021 · evidence:
    tablebucket/idempotency/dup-create-delete-recreate
  - notes: Confirms H-B-020 and refutes the contrarian H-B-021 (second delete is not an idempotent 204). Same
    code for the never-existed ARN: True.

- <a id="s3t-tablebucket-131"></a>**S3T-TABLEBUCKET-131** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **A rejected CreateTableBucket (400) or CreateTable (400/500) does NOT reserve the name: the immediate same-name retry -> 200 (6/6)**
  CreateTableBucket failing with kms-ghost-key -> BadRequestException (HTTP 400) 'The specified KMS key ARN
  does not exist.'; 51-tags -> BadRequestException (HTTP 400) 'The provided tags are invalid. Number of user
  tags must not exceed 50'; bad-storage-class -> BadRequestException (HTTP 400) 'The specified storage class
  is not valid.'. A plain CreateTableBucket of the SAME name 1581-1610 ms later -> {'kms-ghost-key': 'HTTP
  200', '51-tags': 'HTTP 200', 'bad-storage-class': 'HTTP 200'}. CreateTable failing with schema-empty-500 ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'; partition-bad-source-400
  -> BadRequestException (HTTP 400) 'The specified metadata is not valid'; kms-ghost-key-400 ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN does not exist.'. A plain CreateTable of the SAME
  name right after -> {'schema-empty-500': 'HTTP 200', 'partition-bad-source-400': 'HTTP 200',
  'kms-ghost-key-400': 'HTTP 200'}; three more schema={} attempts -> ['InternalServerErrorException/500',
  'InternalServerErrorException/500', 'InternalServerErrorException/500']; no ghost entries in ListTables.
  - ACK: requeue, terminal_codes · ops: CreateTableBucket, CreateTable, GetTableBucket, GetTable · fields:
    name, encryptionConfiguration, tags, metadata
  - repro: CreateTableBucket(name, aws:kms ghost key | 51 tags | GLACIER) -> 400; CreateTableBucket(name) at
    once. CreateTable(name, schema={} | bad partition sourceId | ghost key) -> 500/400; CreateTable(name) at
    once.
  - measurements: bucket_retry_gap_ms=[1581, 1609, 1610], table_retry_gap_ms=[346, 387, 403], n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLE-133](table-schema.md#s3t-table-133), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-047](#s3t-tablebucket-047),
    [S3T-TABLEBUCKET-105](tablebucket-encryption-storage.md#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](tablebucket-encryption-storage.md#s3t-tablebucket-128), [S3T-TABLEBUCKET-107](tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) · evidence:
    tablebucket/creative/failed-create-name-reservation
  - notes: Extends [S3T-TABLEBUCKET-115](#s3t-tablebucket-115) (10-min 'transitional state' after a DELETE) with the create-failure
    side: unlike a deletion, a rejected create leaves no name reservation, so a controller may retry the same
    name on the next reconcile without a backoff; extends [S3T-TABLE-133](table-schema.md#s3t-table-133) (schema={} -> 500) with the...
  - full notes: [details/S3T-TABLEBUCKET-131.md](details/S3T-TABLEBUCKET-131.md)

## Errors

- <a id="s3t-tablebucket-026"></a>**S3T-TABLEBUCKET-026** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-core 4xx bodies are restJson {"message"} with the code only in x-amzn-ErrorType; 400 messages name the field class, not the rule**
  Across 63 rejected calls in this probe every error body was {"message": "..."} (no __type member) with
  header x-amzn-ErrorType '<Code>:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'.
  BadRequestException (HTTP 400) messages identify the field class: 'The specified bucket name isn't
  valid...', 'The specified SSE algorithm is invalid.', 'The specified storage class is not valid.', "The
  specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.", 'The provided tags are invalid.
  Number of user tags must not exceed 50' / 'Tag key is empty or exceeds max length' / 'Tag value exceeds max
  length', "Tags beginning with 'aws:' are reserved for system use.", 'The specified bucket type filter value
  isn't valid...'. NotFoundException (HTTP 404): 'The specified bucket does not exist.' for bucket ARNs; the
  message does not say which ARN segment was wrong.
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, GetTableBucket, ListTableBuckets
  - repro: Any rejected CreateTableBucket / GetTableBucket; inspect raw HTTP body and x-amzn-ErrorType header
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLEBUCKET-024](tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLEBUCKET-087](tablebucket-encryption-storage.md#s3t-tablebucket-087) · hypotheses:
    H-B-068 · evidence: tablebucket/weird-inputs/create-names-enums-empty-structs
  - notes: Qualifies H-B-068: shape and modeled HTTP statuses confirmed; messages are field-class specific
    (name / SSE algorithm / storage class / KMS key ARN / tags) but do not name the violated rule for bucket
    names. 404 messages are generic as predicted. 403/409/429 bodies are recorded in the idempotency...
  - full notes: [details/S3T-TABLEBUCKET-026.md](details/S3T-TABLEBUCKET-026.md)

- <a id="s3t-tablebucket-030"></a>**S3T-TABLEBUCKET-030** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **TagResource length violations return the NON-modeled ValidationException (HTTP 400), unlike CreateTableBucket's BadRequestException**
  TagResource({'k'*128: 'v'*256}) -> HTTP 200. TagResource({ok: 1, 'k'*129: v}) -> ValidationException HTTP
  400 "1 validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy
  constraint: [Member must have length less than or equal to 128, Member must have length greater than or
  equal to 1]"; 'ok' present afterwards: False. TagResource({ok2: 1, k: 'v'*257}) -> ValidationException HTTP
  400 "1 validation error detected: Value at 'tags' failed to satisfy constraint: Map value must satisfy
  constraint: [Member must have length less than or equal to 256, Member must have length greater than or
  equal to 0]"; 'ok2' present afterwards: False.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, ListTagsForResource · fields: tags
  - repro: TagResource with a 128/256 pair, then {ok, 129-char key}, then {ok2, 257-char value};
    ListTagsForResource after each
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-025](#s3t-tablebucket-025), [S3T-TABLEBUCKET-029](#s3t-tablebucket-029), [S3T-TABLEBUCKET-031](#s3t-tablebucket-031) · hypotheses: H-B-045 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: Confirms H-B-045 for the limits (128/256 accepted, 129/257 rejected, all-or-nothing per call) but
    QUALIFIES the code: TagResource (and an empty key) fail with code 'ValidationException' and the framework
    message '1 validation error detected: Value at 'tags' failed to satisfy constraint: Map keys...
  - full notes: [details/S3T-TABLEBUCKET-030.md](details/S3T-TABLEBUCKET-030.md)

- <a id="s3t-tablebucket-043"></a>**S3T-TABLEBUCKET-043** `error-code` · impact medium · handled · verified 2026-10-13
  **UntagResource on a table ARN with an unknown table uuid returns InternalServerErrorException 500 in 3/3 tries (ListTags/Tag -> 404)**
  Under an existing bucket, for three table ARNs with non-existent uuids: uuid 00000000...: ListTags
  NotFoundException HTTP 404 'The specified table does not exist.' | Tag NotFoundException HTTP 404 'The
  specified table does not exist.' | Untag InternalServerErrorException HTTP 500 'An internal error occurred.
  Try again.'; uuid b56fa2f5...: ListTags NotFoundException HTTP 404 'The specified table does not exist.' |
  Tag NotFoundException HTTP 404 'The specified table does not exist.' | Untag InternalServerErrorException
  HTTP 500 'An internal error occurred. Try again.'; uuid 7a6d9922...: ListTags NotFoundException HTTP 404
  'The specified table does not exist.' | Tag NotFoundException HTTP 404 'The specified table does not exist.'
  | Untag InternalServerErrorException HTTP 500 'An internal error occurred. Try again.'. Under a missing
  bucket: UntagResource -> InternalServerErrorException HTTP 500 'An internal error occurred. Try again.',
  ListTagsForResource -> NotFoundException HTTP 404 'The specified table does not exist.'.
  - ACK: terminal_codes, requeue, tags.custom-sync · ops: UntagResource, TagResource, ListTagsForResource ·
    fields: resourceArn, tagKeys
  - repro: UntagResource(resourceArn=<bucket arn>/table/<random uuid>, tagKeys=[a]) vs
    TagResource/ListTagsForResource on the same ARN
  - measurements: tries=3, http_500=3, n=3
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-033](#s3t-tablebucket-033), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015) · hypotheses: H-B-048 · evidence:
    tablebucket/error-taxonomy/arn-forms-get-delete-tags
  - notes: Qualifies H-B-048: the 'missing table ARN -> 404' rule holds for ListTagsForResource and
    TagResource but UntagResource surfaces a 5xx (service bug: a controller retrying 5xx would loop until the
    table ARN is corrected). First seen in tablebucket/tags/upsert-limits-chars-scope.

- <a id="s3t-tablebucket-070"></a>**S3T-TABLEBUCKET-070** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **10 concurrent TagResource on one bucket -> 8-9/10 InternalServerErrorException 500 (never 409); some 500-failed calls still applied the tag**
  Two bursts of 10 concurrent TagResource calls (one distinct key each) on one table bucket: run 1
  {"InternalServerErrorException/500": 9, "200": 1} (wall 0.63s), run 2 {"InternalServerErrorException/500":
  8, "200": 2} (wall 0.18s); ListTagsForResource afterwards showed 4/1 and 4/2 of the successfully-tagged
  keys. Rejected calls retried sequentially: [('k00', 'InternalServerErrorException', 'HTTP 200'), ('k01',
  'InternalServerErrorException', 'HTTP 200'), ('k02', 'InternalServerErrorException', 'HTTP 200'), ('k03',
  'InternalServerErrorException', 'HTTP 200'), ('k04', 'InternalServerErrorException', 'HTTP 200'), ('k05',
  'InternalServerErrorException', 'HTTP 200'), ('k07', 'InternalServerErrorException', 'HTTP 200'), ('k08',
  'InternalServerErrorException', 'HTTP 200'), ('k09', 'InternalServerErrorException', 'HTTP 200')]. Duplicate
  CreateTableBucket(existing name) -> ConflictException (HTTP 409) 'The bucket that you tried to create
  already exists, and you own it.'. DeleteTableBucket(bucket with one namespace) -> BadRequestException (HTTP 400)
  'The bucket that you tried to delete is not empty.'.
  - ACK: tags.custom-sync, requeue, one-per-reconcile · ops: TagResource, CreateTableBucket,
    DeleteTableBucket, ListTagsForResource · fields: tags
  - repro: CreateTableBucket; 10 threads x TagResource(arn, {k_i: v}); ListTagsForResource;
    CreateTableBucket(same name); CreateNamespace; DeleteTableBucket
  - measurements: tag_burst_wall_s=[0.63, 0.18], tag_latency_ms_min_max=[[50,631], [29,178]], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-159](#s3t-tablebucket-159), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKET-019](#s3t-tablebucket-019), [S3T-TABLEBUCKET-044](#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](#s3t-tablebucket-132), [S3T-NAMESPACE-024](namespace.md#s3t-namespace-024), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033),
    [S3T-TABLEBUCKET-043](#s3t-tablebucket-043), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015) · hypotheses: H-B-017 · evidence:
    tablebucket/limits/bucket-quota-create-burst
  - notes: H-B-017 refuted for the concurrent-write half: concurrent writes to one bucket's tags surface as
    InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' and never as
    ConflictException. Run 1: 1 call returned 200 but ListTagsForResource showed 4 keys (k00,k04,k05 were
    applied...
  - full notes: [details/S3T-TABLEBUCKET-070.md](details/S3T-TABLEBUCKET-070.md)

- <a id="s3t-tablebucket-127"></a>**S3T-TABLEBUCKET-127** `error-code` · impact medium · handled · verified 2026-10-14
  **GetTableBucket ARN variants -> other-region or bogus region 400 region mismatch before lookup; same-region nonexistent 404; empty region 400**
  On the us-west-2 endpoint: ARN with region us-east-1 and a bucket name that exists nowhere ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.'; same-region
  nonexistent name -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; ARN region
  'xx-fake-9' -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'; ARN with an empty region segment -> BadRequestException (HTTP 400) 'The specified table bucket
  ARN is not valid.'. The region check runs before the existence lookup: a wrong-region ARN never yields
  NotFoundException.
  - ACK: terminal_codes, exceptions.404 · ops: GetTableBucket · fields: tableBucketARN
  - repro: GetTableBucket with the four ARN variants on the us-west-2 endpoint
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115),
    [S3T-TABLEBUCKET-016](#s3t-tablebucket-016), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159), [S3T-TABLEBUCKET-022](#s3t-tablebucket-022), [S3T-TABLEBUCKET-039](#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033),
    [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125), [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163), [S3T-TABLEBUCKET-164](#s3t-tablebucket-164),
    [S3T-TABLEBUCKET-038](#s3t-tablebucket-038), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029), [S3T-SERVICE-011](service.md#s3t-service-011) · hypotheses: H-B-062,
    H-S-070 · evidence: tablebucket/cross-region/same-name-two-regions-arn-routing
  - notes: Supports H-B-062/H-S-070: the 400 is a pure ARN-region-vs-endpoint check, independent of whether
    the bucket exists in the other region.
  - full notes: [details/S3T-TABLEBUCKET-127.md](details/S3T-TABLEBUCKET-127.md)

- <a id="s3t-tablebucket-159"></a>**S3T-TABLEBUCKET-159** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-15, re-verified
  **CreateTableBucket ConflictException has two meanings (exists vs post-delete transitional); only the message and a Get tell them apart**
  CreateTableBucket for an existing name -> ConflictException HTTP 409 'The bucket that you tried to create
  already exists, and you own it.'. CreateTableBucket for a name deleted < ~10 min ago -> ConflictException
  HTTP 409 ["The bucket is in a transitional state because of a previous deletion attempt. Try again later."],
  while in that window GetTableBucket -> NotFoundException; DeleteTableBucket -> NotFoundException;
  TagResource -> NotFoundException; ListTableBuckets(prefix) -> 0 item(s), and the same name in the other
  region -> HTTP 200 (created, distinct ARN region) at +2.54 s after the delete. Same code, same HTTP status;
  the stored resource exists only in the first case.
  - ACK: terminal_codes, requeue, custom_create · ops: CreateTableBucket, GetTableBucket, DeleteTableBucket,
    TagResource, ListTableBuckets · fields: name
  - repro: CreateTableBucket(existing) ; DeleteTableBucket -> CreateTableBucket(same) at +15..+600 s ;
    Get/Delete/Tag/List on the ARN inside the window
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-016](#s3t-tablebucket-016), [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015),
    [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125) ·
    evidence: tablebucket/creative/delete-recreate-window
  - notes: extends [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013)/-016 with the side-by-side of both 409 messages and what every other
    bucket op returns inside the window. A controller mapping ConflictException -> 'already exists, adopt'
    would loop on GetTableBucket 404; mapping it -> terminal would never create. Rule: 409 + Get 200 =>...
  - full notes: [details/S3T-TABLEBUCKET-159.md](details/S3T-TABLEBUCKET-159.md)

## Request validation

- <a id="s3t-tablebucket-022"></a>**S3T-TABLEBUCKET-022** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTableBucket enforces the documented naming page exactly: all reserved affixes (incl. any 'aws' prefix) -> BadRequestException 400**
  Client validation off. Accepted (HTTP 200, then deleted): 3-char name, 63-char name, consecutive hyphens
  ('a--b'). Rejected with BadRequestException HTTP 400 "The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.": 2 chars, 64 chars, uppercase, underscore, period, leading
  hyphen, trailing hyphen, IP-like 192.168.0.1, prefixes xn--, sthree-, amzn-s3-demo-, aws-, aws<name> (no
  hyphen) and awsome-, suffixes -s3alias, --ol-s3, --x-s3, --table-s3. GetTableBucket on an ARN built from
  each rejected name: BadRequestException 400 (same name message) for 2/64 chars, uppercase, period,
  leading/trailing hyphen, IP-like, xn--/sthree-/amzn-s3-demo- prefixes and all four suffixes;
  NotFoundException 404 'The specified bucket does not exist.' for the underscore name and for
  aws-/aws<name>/awsome- (GetTableBucket does not apply the aws-prefix rule and tolerates '_').
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, GetTableBucket · fields: name
  - repro: CreateTableBucket(name=<candidate>) for each candidate with parameter_validation=False; Delete any
    200; Get constructed ARN for each 400
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003),
    [S3T-SERVICE-011](service.md#s3t-service-011) · hypotheses: H-B-030, H-B-031 · evidence:
    tablebucket/weird-inputs/create-names-enums-empty-structs
  - notes: Confirms H-B-030 for every cell (the model pattern [0-9a-z-]{3,63} is looser than the live rule).
    Refutes H-B-031: the 'aws' prefix rule is an exact string prefix ('awsome-...' and 'aws<name>' are
    rejected like 'aws-...'), and xn--/-s3alias/--table-s3 are enforced, so the live service matches the...
  - full notes: [details/S3T-TABLEBUCKET-022.md](details/S3T-TABLEBUCKET-022.md)

- <a id="s3t-tablebucket-025"></a>**S3T-TABLEBUCKET-025** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTableBucket with 51 tags / aws: key / 129-char key / 257-char value -> BadRequestException 400, no bucket; 50 tags, 128/256 accepted**
  51 tags -> BadRequestException/400 'The provided tags are invalid. Number of user tags must not exceed 50';
  {aws:x, ok} -> BadRequestException/400 "Tags beginning with 'aws:' are reserved for system use."; 129-char
  key -> BadRequestException/400 'The provided tags are invalid. Tag key is empty or exceeds max length';
  257-char value -> BadRequestException/400 'The provided tags are invalid. Tag value exceeds max length'. For
  each rejected call GetTableBucket on the constructed ARN -> ["NotFoundException HTTP 404 'The specified
  bucket does not exist.'"] and ListTableBuckets(prefix) -> ['0'] items. Boundary: one 128-char key with a
  256-char value -> HTTP 200 (deleted again); exactly 50 tags -> HTTP 200 (deleted again).
  - ACK: custom_create, terminal_codes · ops: CreateTableBucket, GetTableBucket, ListTableBuckets · fields:
    tags
  - repro: CreateTableBucket(name, tags=<bad set>) -> GetTableBucket(constructed arn) ->
    ListTableBuckets(prefix=name)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-030](#s3t-tablebucket-030), [S3T-TABLEBUCKET-029](#s3t-tablebucket-029), [S3T-TABLEBUCKET-031](#s3t-tablebucket-031), [S3T-TABLEBUCKET-148](#s3t-tablebucket-148),
    [S3T-TABLEBUCKET-149](#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](#s3t-tablebucket-150), [S3T-TABLE-185](table.md#s3t-table-185), [S3T-TABLEBUCKET-152](#s3t-tablebucket-152), [S3T-TABLE-118](table.md#s3t-table-118) · hypotheses:
    H-B-038, H-B-045 · evidence: tablebucket/weird-inputs/create-names-enums-empty-structs
  - notes: H-B-038: confirmed - create is atomic w.r.t. tags. Also covers the create-time half of H-B-045
    (128/256 accepted, 129/257 rejected).
  - full notes: [details/S3T-TABLEBUCKET-025.md](details/S3T-TABLEBUCKET-025.md)

- <a id="s3t-tablebucket-041"></a>**S3T-TABLEBUCKET-041** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTableBuckets: bad/stale token -> BadRequestException 400; token reused after a delete skips (200); invalid prefixes -> 200 empty**
  continuationToken='abc' -> BadRequestException HTTP 400 'The continuation token is no longer valid.'; 2049
  chars -> BadRequestException HTTP 400 'The continuation token is no longer valid.'; page-1 token
  (maxBuckets=1) re-sent after DeleteTableBucket of the next bucket -> HTTP 200 ['ackq-<run>-ordz']
  next_token=False; the same token with a different prefix -> n/a. prefix validation: {"Ackq": "HTTP 200 (0
  items)", "ackq_b": "HTTP 200 (0 items)", "ackq.b": "HTTP 200 (0 items)", "aaaaaaaaaa..(64)": "HTTP 200 (0
  items)", "aaaaaaaaaa..(63)": "HTTP 200 (0 items)", "<empty>": "HTTP 200 (7 items)", "ackq-": "HTTP 200 (7
  items)", "ackq-0a78c..(16)": "HTTP 200 (1 items)"}. maxBuckets=0 -> BadRequestException HTTP 400 'The
  specified limit is not valid. A limit should be a positive number'; maxBuckets=1001 -> HTTP 200.
  - ACK: custom_find, terminal_codes · ops: ListTableBuckets · fields: continuationToken, prefix, maxBuckets
  - repro: ListTableBuckets with bad tokens / prefixes / maxBuckets values (parameter_validation=False)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125) · hypotheses:
    H-B-051, H-B-052 · evidence: tablebucket/error-taxonomy/arn-forms-get-delete-tags
  - notes: H-B-051 token half: confirmed for garbage and over-long tokens (BadRequestException 'The
    continuation token is no longer valid.'), but a valid page token re-sent after the next bucket was deleted
    returned HTTP 200 with the following bucket (no error, no duplicate). H-B-052 REFUTED: prefix='Ackq',...
  - full notes: [details/S3T-TABLEBUCKET-041.md](details/S3T-TABLEBUCKET-041.md)

- <a id="s3t-tablebucket-069"></a>**S3T-TABLEBUCKET-069** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTableBuckets maxBuckets=0/-1 -> BadRequestException 400, maxBuckets=1001 -> 200 (not rejected); last page omits continuationToken**
  ListTableBuckets(maxBuckets=0) -> BadRequestException (HTTP 400) 'The specified limit is not valid. A limit
  should be a positive number'; (maxBuckets=1001) -> HTTP 200, 11 items; (maxBuckets=-1) ->
  BadRequestException (HTTP 400) 'The specified limit is not valid. A limit should be a positive number';
  (maxBuckets=1000) -> HTTP 200, 11 items. With 9 matching buckets, maxBuckets=1 + prefix returned 1 item per
  page with continuationToken present on 8 of 9 pages (token length [608]), the final page OMITS
  continuationToken entirely (key absent, not an empty string), 0 duplicate(s) across pages. The default page
  (maxBuckets omitted) returned 11 buckets with no continuationToken.
  - ACK: custom_find, list_operation.match_fields · ops: ListTableBuckets · fields: maxBuckets,
    continuationToken, prefix
  - repro: ListTableBuckets(maxBuckets=0|1001|-1) with client validation off; ListTableBuckets(prefix,
    maxBuckets=1) following tokens
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-041](#s3t-tablebucket-041), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125),
    [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-TABLE-293](table.md#s3t-table-293), [S3T-SERVICE-009](service.md#s3t-service-009), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLE-168](table.md#s3t-table-168), [S3T-TABLE-092](table.md#s3t-table-092),
    [S3T-TABLE-091](table.md#s3t-table-091), [S3T-TABLEBUCKET-168](#s3t-tablebucket-168), [S3T-TABLEBUCKET-169](#s3t-tablebucket-169) · hypotheses: H-B-050 · evidence:
    tablebucket/limits/bucket-quota-create-burst
  - notes: H-B-050 qualified: 0 rejected as predicted (400 'The specified limit is not valid. A limit should
    be a positive number'), token present/absent semantics as predicted (absent on the last page, never an
    empty string), BUT maxBuckets=1001 was accepted with HTTP 200 (not a 400) - the documented max of...
  - full notes: [details/S3T-TABLEBUCKET-069.md](details/S3T-TABLEBUCKET-069.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-tablebucket-060"></a>**S3T-TABLEBUCKET-060** `read-gap` · impact medium · handled · verified 2026-10-13
  **GetTableBucket raw JSON keys are identical before and the same after setting 5 sub-resources (type=customer)**
  GetTableBucket on the fresh bucket returned raw JSON keys ['arn', 'createdAt', 'name', 'ownerAccountId',
  'tableBucketId', 'type'] (type=customer). After PutTableBucketEncryption(AES256), PutTableBucketPolicy,
  PutTableBucketMetricsConfiguration, PutTableBucketStorageClass(INTELLIGENT_TIERING) and
  PutTableBucketMaintenanceConfiguration(disabled 2/2) (outcomes {'PutTableBucketEncryption': 'HTTP 200',
  'PutTableBucketPolicy': 'HTTP 200', 'PutTableBucketMetricsConfiguration': 'HTTP 204',
  'PutTableBucketStorageClass': 'HTTP 200', 'PutTableBucketMaintenanceConfiguration': 'HTTP 204'})
  GetTableBucket returned keys ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']: none
  of the sub-resources is reflected. Replication was not set.
  - ACK: custom_find, hook:sdk_read_one_post_set_output · ops: GetTableBucket · fields: arn, name,
    ownerAccountId, createdAt, tableBucketId, type
  - repro: GetTableBucket (raw) -> set encryption/policy/metrics/storage class/maintenance -> GetTableBucket
    (raw) -> diff keys
  - handling: handled via `templates/hooks/table_bucket/sdk_read_one_post_set_output.go.tpl:6-12; pkg/resource/table_bucket/hooks.go:51-171`
  - related: [S3T-TABLEBUCKET-117](tablebucket-encryption-storage.md#s3t-tablebucket-117), [S3T-TABLEBUCKET-116](tablebucket-replication.md#s3t-tablebucket-116) · hypotheses: H-S-068, H-S-082 · evidence:
    tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-068 confirmed (5 of 6 sub-resources set; replication not tried). H-S-082 (static): the six Put
    ops are modeled [create] with read-only read shapes; live here the maintenance Put merges into the stored
    value (see the partial-Put finding) and the other Puts replace a single value, consistent with...
  - full notes: [details/S3T-TABLEBUCKET-060.md](details/S3T-TABLEBUCKET-060.md)

## Response fidelity and consistency

- <a id="s3t-tablebucket-121"></a>**S3T-TABLEBUCKET-121** `eventual-consistency` · impact medium · handled · verified 2026-10-14
  **Hypothesis of a transient 404 after CreateTableBucket refuted: Get, List(prefix) and ListTags all 200 on poll #1 in 5/5, no flapping** (hypothesis refuted; behavior confirmed)
  5 fresh buckets (CreateTableBucket latency 1454-1678 ms, tags={seed:1}); right after Create returned,
  GetTableBucket, ListTableBuckets(prefix=<name>) and ListTagsForResource(arn) were polled in lockstep every
  100 ms until all agreed for 5 polls. First-success offsets since Create returned (ms): Get [0, 0, 0, 0, 0],
  List(prefix) [0, 0, 0, 0, 0] (bucket present), ListTags [0, 0, 0, 0, 0] (seed tag present). Codes seen
  before the first success: none. Successes followed by a failure (flapping): none.
  - ACK: none · ops: CreateTableBucket, GetTableBucket, ListTableBuckets, ListTagsForResource · fields:
    tableBucketARN, prefix, tags
  - repro: CreateTableBucket(tags) -> poll GetTableBucket + ListTableBuckets(prefix) + ListTagsForResource
    every 100 ms for 5 agreeing polls; x5
  - measurements: first_ok_ms_get=[0, 0, 0, 0, 0], first_ok_ms_list=[0, 0, 0, 0, 0], first_ok_ms_tags=[0, 0,
    0, 0, 0], create_latency_ms=[1678, 1454, 1540, 1458, 1538], polls_per_trial=[5, 5, 5, 5, 5],
    poll_interval_s=0.1, n=5
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122),
    [S3T-TABLEBUCKET-156](#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](#s3t-tablebucket-157), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028),
    [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-160](#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](#s3t-tablebucket-167), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) · hypotheses: H-B-003, H-B-046 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5
  - notes: H-B-003 (contrarian: transient 404 within 1-2 s after Create) refuted for GetTableBucket and
    extended to ListTableBuckets(prefix) and ListTagsForResource: every read path was correct on poll #1.
    H-B-046 create-time-tags part confirmed: tags passed to CreateTableBucket were present on the first...
  - full notes: [details/S3T-TABLEBUCKET-121.md](details/S3T-TABLEBUCKET-121.md)

- <a id="s3t-tablebucket-122"></a>**S3T-TABLEBUCKET-122** `response-fidelity` · impact medium · partially handled · verified 2026-10-14
  **ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6 fields, raw createdAt strings identical)**
  On the first ListTableBuckets(prefix=<name>) after CreateTableBucket (the bucket present on poll #1 in 5/5),
  the TableBucketSummary had fields ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']
  and GetTableBucket had ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']; parsed
  values equal on every field in 5/5 trials. Raw HTTP bodies: createdAt string identical between List and Get
  in 5/5 (sample precision '2026-10-14T00:43:19.929801410Z', 9 fractional digits).
  - ACK: list_operation.match_fields, custom_find · ops: ListTableBuckets, GetTableBucket · fields: arn, name,
    ownerAccountId, createdAt, tableBucketId, type, prefix
  - repro: CreateTableBucket -> ListTableBuckets(prefix=name) at +0 ms -> GetTableBucket(arn) -> compare
    summary to body (parsed and raw JSON)
  - measurements: trials_equal=5, raw_createdAt_equal=5, n=5
  - handling: partially handled via `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14; generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304` - see Handling gaps
  - related: [S3T-TABLEBUCKET-010](tablebucket-encryption-storage.md#s3t-tablebucket-010), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121),
    [S3T-TABLEBUCKET-002](#s3t-tablebucket-002), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-042](#s3t-tablebucket-042), [S3T-TABLEBUCKET-160](#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](#s3t-tablebucket-167),
    [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) · hypotheses: H-B-004 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5
  - notes: H-B-004 confirmed: a controller can adopt/de-duplicate by name from one ListTableBuckets(prefix)
    call without a follow-up Get; no field is dropped or reformatted in the summary.
  - full notes: [details/S3T-TABLEBUCKET-122.md](details/S3T-TABLEBUCKET-122.md)

- <a id="s3t-tablebucket-124"></a>**S3T-TABLEBUCKET-124** `eventual-consistency` · impact medium · handled · verified 2026-10-14
  **ListTagsForResource at +0 ms after TagResource/UntagResource on a table bucket: 10/10 and 10/10 consistent; +500 ms 10/10 and 10/10**
  One bucket, 10 iterations: TagResource({k_i: i}) then ListTagsForResource immediately and again at +500 ms;
  UntagResource([k_i]) then the same two reads. Updated map present on the +0 ms read in 10/10 (tag) and 10/10
  (untag); on the +500 ms read in 10/10 and 10/10. Codes: tag ['HTTP 200'], untag ['HTTP 204']. TagResource
  latency 44-66 ms, UntagResource 44-65 ms.
  - ACK: none, tags.custom-sync · ops: TagResource, UntagResource, ListTagsForResource · fields: tags, tagKeys
  - repro: loop 10x: TagResource({k:i}); ListTags @0 and @+500 ms; UntagResource([k]); ListTags @0 and @+500
    ms
  - measurements: tag_read0_consistent=10, tag_read500_consistent=10, untag_read0_consistent=10,
    untag_read500_consistent=10, tag_latency_ms=[52, 59, 66, 61, 44, 53, 54, 54, 49, 52],
    untag_latency_ms=[55, 50, 51, 46, 65, 51, 52, 44, 49, 47], n=10
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLEBUCKET-027](#s3t-tablebucket-027), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-032](#s3t-tablebucket-032)
    · hypotheses: H-B-046 · evidence: tablebucket/consistency-windows/create-read-paths-tags-n5
  - notes: H-B-046 confirmed (brief n=10, read-at-0 and read-at-500 ms, plus create-time tags in the
    read-first finding); consistent with [S3T-TABLE-128](table.md#s3t-table-128) (20/20 at 100 ms). Not run to 100 trials: the +0 ms
    result in 20+ independent cells leaves no window to measure.

## Tags

- <a id="s3t-tablebucket-027"></a>**S3T-TABLEBUCKET-027** `tag-semantics` · impact high · handled · verified 2026-10-13
  **TagResource is an additive upsert (HTTP 200): existing keys kept, same key overwritten, new key added; removal only via UntagResource**
  Bucket created with tags {a:1, b:2, ackq-run}; TagResource({b:9, c:3}) -> HTTP 200 (raw status 200, body
  '{}'); ListTagsForResource immediately -> {"a": "1", "ackq-run": "47dd85", "b": "9", "c": "3"}, at +2 s ->
  {"a": "1", "ackq-run": "47dd85", "b": "9", "c": "3"}. UntagResource([a]) -> HTTP 204; ListTagsForResource ->
  {"ackq-run": "47dd85", "b": "9", "c": "3"}.
  - ACK: tags.custom-sync · ops: TagResource, UntagResource, ListTagsForResource · fields: tags, tagKeys,
    resourceArn
  - repro: CreateTableBucket(tags={a:1,b:2}) -> TagResource({b:9,c:3}) -> ListTagsForResource ->
    UntagResource([a]) -> ListTagsForResource
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-124](#s3t-tablebucket-124), [S3T-TABLEBUCKET-032](#s3t-tablebucket-032),
    [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLEBUCKET-151](#s3t-tablebucket-151), [S3T-TABLE-114](table.md#s3t-table-114) · hypotheses: H-B-040 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: Confirms H-B-040 (visible on the first read, no lag seen).

- <a id="s3t-tablebucket-028"></a>**S3T-TABLEBUCKET-028** `tag-semantics` · impact medium · handled · verified 2026-10-13
  **UntagResource absent keys -> 204, empty tagKeys -> BadRequestException 400; TagResource {} -> 200; all tags removed -> ListTags {"tags":{}}**
  UntagResource([zzz]) (key absent) -> HTTP 204; UntagResource([b, zzz]) -> HTTP 204 and ListTagsForResource
  -> {"ackq-run": "47dd85", "c": "3"} (b removed). UntagResource(tagKeys=[]) -> BadRequestException HTTP 400
  'The provided tags are invalid.'; TagResource(tags={}) -> HTTP 200; TagResource re-sent with an identical
  existing pair -> HTTP 200. After UntagResource of every remaining key (-> HTTP 204) ListTagsForResource ->
  HTTP 200 with raw body {"tags":{}}; repeating that UntagResource -> HTTP 204.
  - ACK: tags.custom-sync, compare.nil_equals_zero_value · ops: UntagResource, TagResource,
    ListTagsForResource · fields: tags, tagKeys
  - repro: UntagResource([absent]); UntagResource([]); TagResource({}); UntagResource(all keys) ->
    ListTagsForResource raw body
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-027](#s3t-tablebucket-027), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-124](#s3t-tablebucket-124), [S3T-TABLEBUCKET-032](#s3t-tablebucket-032),
    [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLEBUCKET-153](#s3t-tablebucket-153), [S3T-TABLEBUCKET-154](#s3t-tablebucket-154), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023),
    [S3T-TABLEBUCKET-155](#s3t-tablebucket-155), [S3T-TABLE-116](table.md#s3t-table-116), [S3T-TABLEBUCKET-156](#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](#s3t-tablebucket-157), [S3T-TABLE-113](table.md#s3t-table-113),
    [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008) · hypotheses: H-B-043, H-B-071 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: H-B-043: confirmed (lenient for absent keys: HTTP 204; empty tagKeys: BadRequestException HTTP 400
    'The provided tags are invalid.'; empty map on the Tag side: HTTP 200). Also covers the
    TagResource-re-sent-unchanged cell of H-B-071 (HTTP 200).
  - full notes: [details/S3T-TABLEBUCKET-028.md](details/S3T-TABLEBUCKET-028.md)

- <a id="s3t-tablebucket-029"></a>**S3T-TABLEBUCKET-029** `tag-semantics` · impact medium · handled · verified 2026-10-13
  **Tag keys case-sensitive (Env/env coexist); empty/unicode values round-trip; aws:/AWS: key -> BadRequestException; '' key -> ValidationExc**
  TagResource({"Env": "a", "env": "b", "empty": "", "équipe": "日本", "k with spaces": "v+-=._:/@"}) -> HTTP
  200; ListTagsForResource raw body {"tags":{"équipe":"日本","k with
  spaces":"v+-=._:/@","Env":"a","env":"b","ackq-run":"47dd85","empty":""}} (all values identical to what was
  sent). TagResource({aws:x: 1}) -> BadRequestException HTTP 400 "Tags beginning with 'aws:' are reserved for
  system use."; ({AWS:x: 1}) -> BadRequestException HTTP 400 "Tags beginning with 'aws:' are reserved for
  system use."; ({'': v}) -> ValidationException HTTP 400 "1 validation error detected: Value at 'tags' failed
  to satisfy constraint: Map keys must satisfy constraint: [Member must have length less than or equal to 128,
  Member must have length greater than or equal to 1]"; tags afterwards {"Env": "a", "ackq-run": "47dd85",
  "empty": "", "env": "b", "k with spaces": "v+-=._:/@", "équipe": "日本"}.
  - ACK: tags.custom-sync · ops: TagResource, ListTagsForResource · fields: tags
  - repro: TagResource({Env, env, empty:'', unicode, spaces}) -> ListTagsForResource; TagResource({aws:x}),
    ({AWS:x}), ({'': v})
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-025](#s3t-tablebucket-025), [S3T-TABLEBUCKET-030](#s3t-tablebucket-030), [S3T-TABLEBUCKET-031](#s3t-tablebucket-031) · hypotheses: H-B-044 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: Confirms H-B-044. 'AWS:x' (uppercase) rejected; the AWS-wide reserved prefix is matched
    case-insensitively.
  - full notes: [details/S3T-TABLEBUCKET-029.md](details/S3T-TABLEBUCKET-029.md)

- <a id="s3t-tablebucket-153"></a>**S3T-TABLEBUCKET-153** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C030 FALSE: TagResource with an empty tags map -> HTTP 200 (no-op) on a bucket and on a table, not BadRequestException**
  TagResource(bucket ARN, tags={}) -> HTTP 200 ([S3T-TABLEBUCKET-028](#s3t-tablebucket-028)); TagResource(table ARN, tags={}) -> HTTP
  200 ([S3T-TABLE-125](table.md#s3t-table-125)); CreateTableBucket(tags={}) -> HTTP 200 as well ([S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023)). Only
  UntagResource(tagKeys=[]) is rejected (BadRequestException 400 'The provided tags are invalid.').
  - ACK: tags.custom-sync, docs-only · ops: TagResource · fields: tags
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023); no new
    API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLEBUCKET-154](#s3t-tablebucket-154) · evidence:
    service/static/doc-claims-1
  - notes: VERDICT: FALSE - the empty map is accepted with 200 on both resource kinds; the API does not
    require 'one or more' tags on TagResource.

## Delete semantics

- <a id="s3t-tablebucket-014"></a>**S3T-TABLEBUCKET-014** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucket -> 204 empty body; GetTableBucket NotFoundException from the first poll (0.02 s); List(prefix) drops it by 0.009 s**
  DeleteTableBucket on an empty bucket -> HTTP 204, body ''. GetTableBucket polled every 50 ms from 0 ms after
  the Delete returned: codes ['NotFoundException'] (first NotFoundException at 0.02 s); 5 further Gets ->
  ['NotFoundException'] (no 200 after a 404). ListTableBuckets(prefix=name) polled every 200 ms: presence
  timeline [(0.009, False)]. ListTagsForResource on the deleted ARN -> NotFoundException HTTP 404 'The
  specified bucket does not exist.'.
  - ACK: none · ops: DeleteTableBucket, GetTableBucket, ListTableBuckets, ListTagsForResource
  - repro: CreateTableBucket -> DeleteTableBucket -> GetTableBucket every 50 ms + ListTableBuckets(prefix)
    every 200 ms
  - measurements: first_404_s=0.02, list_gone_s=0.009, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-015](#s3t-tablebucket-015), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-016](#s3t-tablebucket-016), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159),
    [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-161](#s3t-tablebucket-161), [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-101](tablebucket-replication.md#s3t-tablebucket-101) · hypotheses: H-B-019 ·
    evidence: tablebucket/idempotency/dup-create-delete-recreate
  - notes: Qualifies H-B-019 (single bucket; the round-trip probe adds two more post-delete Get timelines).
    Full original title: DeleteTableBucket returns HTTP 204 with an empty body; GetTableBucket is
    NotFoundException on the first poll and ListTableBuckets(prefix) drops it within 0.009 s

- <a id="s3t-tablebucket-019"></a>**S3T-TABLEBUCKET-019** `delete-semantics` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with only an empty namespace -> same code**
  Bucket with one namespace and one ICEBERG table: DeleteTableBucket -> BadRequestException HTTP 400 'The
  bucket that you tried to delete is not empty.' (x-amzn-ErrorType
  'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', raw body {"message":"The
  bucket that you tried to delete is not empty."}); GetTableBucket afterwards -> HTTP 200 and GetTable -> HTTP
  200 (unchanged); three retries 2 s apart -> ["BadRequestException HTTP 400 'The bucket that you tried to
  delete is not empty.'", "BadRequestException HTTP 400 'The bucket that you tried to delete is not empty.'",
  "BadRequestException HTTP 400 'The bucket that you tried to delete is not empty.'"]. After DeleteTable (HTTP 204)
  with the empty namespace still present: DeleteTableBucket -> BadRequestException HTTP 400 'The bucket that
  you tried to delete is not empty.'. After DeleteNamespace (204) DeleteTableBucket -> HTTP 204 issued 0.549 s
  later.
  - ACK: terminal_codes, requeue, pre-delete-cleanup · ops: DeleteTableBucket, DeleteTable, DeleteNamespace,
    GetTableBucket
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(ICEBERG) -> DeleteTableBucket (x4) ->
    DeleteTable -> DeleteTableBucket -> DeleteNamespace -> DeleteTableBucket
  - measurements: retries=3, namespace_delete_to_bucket_delete_s=0.549, n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-044](#s3t-tablebucket-044), [S3T-TABLEBUCKET-132](#s3t-tablebucket-132), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-NAMESPACE-024](namespace.md#s3t-namespace-024),
    [S3T-NAMESPACE-037](namespace.md#s3t-namespace-037), [S3T-NAMESPACE-039](namespace.md#s3t-namespace-039), [S3T-NAMESPACE-027](namespace.md#s3t-namespace-027), [S3T-NAMESPACE-028](namespace.md#s3t-namespace-028), [S3T-NAMESPACE-021](namespace.md#s3t-namespace-021) ·
    hypotheses: H-B-022, H-B-023, H-B-024, H-B-025 · evidence:
    tablebucket/dependencies/delete-with-children-and-configs
  - notes: H-B-022 refuted (code BadRequestException); H-B-023 confirmed; H-B-024 confirmed; H-B-025 refuted.
    Compare the message with the 400 for an uppercase name: BadRequestException HTTP 400 "The specified bucket
    name isn't valid. Specify a different bucket name, and then try again.".
  - full notes: [details/S3T-TABLEBUCKET-019.md](details/S3T-TABLEBUCKET-019.md)

- <a id="s3t-tablebucket-115"></a>**S3T-TABLEBUCKET-115** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket with a just-deleted name -> ConflictException 409 'transitional state ... previous deletion attempt' ~10 min (OK +606 s)**
  After DeleteTableBucket returned HTTP 204 (GetTableBucket -> NotFoundException on the first poll),
  CreateTableBucket with the same name returned ConflictException (HTTP 409) 'The bucket is in a transitional
  state because of a previous deletion attempt. Try again later.' on all 44 attempts between +0.03 s and
  +594.6 s after the delete (every 5 s for the first minute, then every 10 s; a gap of ~4 min while the probe
  was restarted) and succeeded (HTTP 200) at +606.2 s. The name stays reserved in a server-side 'transitional
  state' that is invisible to GetTableBucket (404) and to the sub-resource Gets (404 'The specified bucket
  does not exist.').
  - ACK: requeue, deletable.when · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket · fields: name
  - repro: DeleteTableBucket -> GetTableBucket until 404 -> CreateTableBucket(same name) every 5-10 s until it
    succeeds
  - measurements: recreate_409_last_seen_s=594.6, recreate_success_s=606.2, attempts=45, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-016](#s3t-tablebucket-016), [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159),
    [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013),
    [S3T-TABLEBUCKETREPLICATION-029](tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-030](tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005) · hypotheses:
    H-S-071, H-S-022, H-S-099 · evidence: tablebucket/identity/subresources-across-recreate
  - notes: Qualifies H-S-071: no DELETING state is observable on the read path, but the create path exposes
    one through ConflictException 409 for ~10 minutes - a controller re-creating a deleted TableBucket (or an
    e2e test reusing names) must retry this 409 for at least that long; the window may be longer for...
  - full notes: [details/S3T-TABLEBUCKET-115.md](details/S3T-TABLEBUCKET-115.md)

- <a id="s3t-tablebucket-158"></a>**S3T-TABLEBUCKET-158** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Delete->404->CreateTableBucket every 15 s: name blocked ~609 s whether the bucket was empty, populated or not retried until +420 s**
  Controller loop (DeleteTableBucket 204, GetTableBucket 404 at once, CreateTableBucket(same name) every 15
  s): first 200 at +609.4 s for a bucket created and deleted within one second, +608.2 s for a bucket that had
  held a namespace, 2 tables, tags, policy, maintenance, storage-class and encryption settings (children
  deleted just before), and +608.8 s for a name whose first retry was only sent at +420 s (so 409 retries do
  not extend the window and content does not lengthen it). Every rejected retry was ConflictException HTTP 409
  with the single message ["The bucket is in a transitional state because of a previous deletion attempt. Try
  again later."] (n_409 per name: {"empty": 40, "full": 40, "late": 12}). After the window the ARN is
  identical and tableBucketId is new (changed); the re-created 'full' name reads fresh defaults (policy
  NotFoundException, maintenance {"icebergUnreferencedFileRemoval": {"status": "enabled", "settings":
  {"icebergUnreferencedFileRemoval": {"unreferencedDays": 3, "nonCurrentDays": 10}}}}, tags {"ackq-run":
  "5d7947"}).
  - ACK: requeue, terminal_codes, custom_create · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket ·
    fields: name, tableBucketId
  - repro: CreateTableBucket x3 -> (populate one) -> DeleteTableBucket x3 -> loop: CreateTableBucket every 15
    s until 200 (one name waits 420 s first)
  - measurements: recreate_ok_at_s.empty=609.4, recreate_ok_at_s.full=608.2,
    recreate_ok_at_s.late_first_retry_420s=608.8, retry_interval_s=15, n_409_per_name.empty=40,
    n_409_per_name.full=40, n_409_per_name.late=12, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-047](#s3t-tablebucket-047), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-116](tablebucket-replication.md#s3t-tablebucket-116), [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001),
    [S3T-TABLEBUCKET-053](tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-020](tablebucket-encryption-storage.md#s3t-tablebucket-020),
    [S3T-TABLEBUCKET-079](tablebucket-maintenance-metrics.md#s3t-tablebucket-079) · evidence: tablebucket/creative/delete-recreate-window
  - notes: extends [S3T-TABLEBUCKET-047](#s3t-tablebucket-047)/-115 with the controller-shaped loop and the three window variants
    (content, no-retry, retry); extends -116 with the fresh sub-resource reads after re-creation. Decision
    rule: a create that returns 409 while GetTableBucket is 404 is 'wait', not 'exists' and not terminal;...
  - full notes: [details/S3T-TABLEBUCKET-158.md](details/S3T-TABLEBUCKET-158.md)

- <a id="s3t-tablebucket-184"></a>**S3T-TABLEBUCKET-184** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Post-delete name reservation measured: same-name CreateTableBucket -> 409 transitional until 570/571, succeeds at 602/603 s (n=2)**
  Two buckets deleted (HTTP 204/204). CreateTableBucket of the same name at +2, +15, +30 s then every 30 s:
  ConflictException (HTTP 409) 'The bucket is in a transitional state because of a previous deletion attempt.
  Try again later.'; the first success came at 602.1/602.9 s after the delete (last 409 at 570.4/570.7 s);
  re-created ARN identical to the deleted one: True/True. Inside the window GetTableBucket ->
  NotFoundException / NotFoundException, DeleteTableBucket -> NotFoundException, TagResource ->
  NotFoundException, ListTableBuckets(prefix) -> 0/0 item(s); the same name in us-east-1 -> HTTP 200 at
  +1.25/+0.97 s (ARN region differs: True). CreateTableBucket of an EXISTING name -> ConflictException (HTTP 409)
  'The bucket that you tried to create already exists, and you own it.' (2/2).
  - ACK: requeue, e2e-timing · ops: CreateTableBucket, DeleteTableBucket, GetTableBucket, TagResource,
    ListTableBuckets · fields: name, arn
  - repro: CreateTableBucket(existing) x2; DeleteTableBucket(b1,b2) -> CreateTableBucket(same name) at
    +2/+15/+30 s then every 30 s (<= 12 min); Get/Delete/Tag/List(prefix) inside the window;
    CreateTableBucket(same name) in the other region right after the delete
  - measurements: recreate_success_at_s=[602.1, 602.9], last_transitional_409_at_s=[570.4, 570.7], n=2,
    other_region_create_at_s=[1.25, 0.97], attempts_per_bucket=[22, 22]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-159](#s3t-tablebucket-159) · evidence: tablebucket/creative/reverify-set-d
  - notes: Reverify set D for [S3T-TABLEBUCKET-159](#s3t-tablebucket-159): bounds the '< ~10 min' transitional window with n=2 instead
    of a single sample; the two ConflictException meanings are told apart only by the message (and a Get ->
    404 for the transitional case).

## Quotas and rate limits

- <a id="s3t-tablebucket-018"></a>**S3T-TABLEBUCKET-018** `quota-limit` · impact medium · handled · verified 2026-10-13
  **CreateTableBucket same-name 409 storm at ~8 calls/s is throttled: TooManyRequestsException 429 in runs of 4-6 after 5-9 accepted (~5 TPS)**
  500 CreateTableBucket calls for one name at ~100 ms spacing (~8/s) over 60 s: 314 reached the service (409
  ConflictException) and 186 were throttled with TooManyRequestsException HTTP 429 'Too Many Requests', in a
  regular pattern of 5-9 accepted then 4-6 throttled, consistent with a ~5 TPS token bucket for
  CreateTableBucket. Other probers shared the account, so the exact limit is an estimate.
  - ACK: requeue, terminal_codes · ops: CreateTableBucket
  - repro: CreateTableBucket(same name) every 100 ms for 60 s; count 429s
  - measurements: calls=500, throttled_429=186, window_s=60, approx_accept_rate_per_s=5, n=1
  - handling: handled via `test/e2e/tests/test_table_bucket.py:48-69; test/e2e/tests/test_table_bucket.py:225-233; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-068](#s3t-tablebucket-068), [S3T-TABLEBUCKET-179](#s3t-tablebucket-179), [S3T-TABLEBUCKET-071](#s3t-tablebucket-071), [S3T-TABLEBUCKET-165](#s3t-tablebucket-165),
    [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094),
    [S3T-SERVICE-004](service.md#s3t-service-004), [S3T-SERVICE-008](service.md#s3t-service-008), [S3T-SERVICE-018](service.md#s3t-service-018) · evidence:
    tablebucket/idempotency/dup-create-delete-recreate
  - notes: Observed incidentally; TooManyRequestsException is a modeled error on every bucket operation and
    must be retried, not treated as terminal.
  - full notes: [details/S3T-TABLEBUCKET-018.md](details/S3T-TABLEBUCKET-018.md)

- <a id="s3t-tablebucket-031"></a>**S3T-TABLEBUCKET-031** `quota-limit` · impact medium · handled · verified 2026-10-13
  **50-tag limit is cumulative per bucket: 30+25 -> BadRequestException 400 (none applied); +20 -> 200 (50); 51st -> 400; update at 50 -> 200**
  Bucket created with 30 tags (count 30). TagResource(25 new keys) -> BadRequestException HTTP 400 'The
  provided tags are invalid. Number of user tags must not exceed 50', count afterwards 30. TagResource(20 new
  keys) -> HTTP 200, count 50. TagResource({k51}) -> BadRequestException HTTP 400 'The provided tags are
  invalid. Number of user tags must not exceed 50', count 50. At 50 tags TagResource({k01: updated}) -> HTTP
  200 and TagResource({k02: updated, k52: x}) -> BadRequestException HTTP 400 'The provided tags are invalid.
  Number of user tags must not exceed 50'; final count 50, k01='updated', k02='v', k52 present=False.
  - ACK: tags.custom-sync, terminal_codes · ops: CreateTableBucket, TagResource, ListTagsForResource · fields:
    tags
  - repro: CreateTableBucket(30 tags) -> TagResource(25 new) -> TagResource(20 new) -> TagResource(1 new) ->
    TagResource(existing key) -> TagResource(existing+new)
  - measurements: max_tags_observed=50, n=1
  - handling: handled via `pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-025](#s3t-tablebucket-025), [S3T-TABLEBUCKET-030](#s3t-tablebucket-030), [S3T-TABLEBUCKET-029](#s3t-tablebucket-029), [S3T-TABLEBUCKET-148](#s3t-tablebucket-148),
    [S3T-TABLEBUCKET-149](#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](#s3t-tablebucket-150), [S3T-TABLE-185](table.md#s3t-table-185) · hypotheses: H-B-041, H-B-042 · evidence:
    tablebucket/tags/upsert-limits-chars-scope
  - notes: H-B-041 confirmed; H-B-042 refuted. The limit counts keys on the resource, so the controller's tag
    sync must account for out-of-band tags.
  - full notes: [details/S3T-TABLEBUCKET-031.md](details/S3T-TABLEBUCKET-031.md)

## Cross-region

- <a id="s3t-tablebucket-125"></a>**S3T-TABLEBUCKET-125** `cross-region` · impact high · handled · verified 2026-10-14
  **Same bucket name in us-west-2 and us-east-1 -> both Create 200, ARNs differ only in region, distinct tableBucketIds, List/tags per region**
  CreateTableBucket(name=X) in us-west-2 then the same name in us-east-1 -> both HTTP 200 (latency 1911 / 1253
  ms); the ARNs differ only in ARN segment(s) [3] (index 3 = region); GetTableBucket per region: same
  tableBucketId = False. ListTableBuckets(prefix=X) on the us-west-2 endpoint -> [('ackq-5443b8-xr',
  'us-west-2')]; on the us-east-1 endpoint -> [('ackq-5443b8-xr', 'us-east-1')] (each region lists only its
  own bucket). Tags are independent: us-west-2 tags {'ackq-run': '5443b8', 'side': 'A'}, us-east-1 tags
  {'ackq-run': '5443b8', 'side': 'B'}.
  - ACK: is_arn_primary_key, list_operation.match_fields · ops: CreateTableBucket, GetTableBucket,
    ListTableBuckets, ListTagsForResource · fields: name, arn, tableBucketId
  - repro: CreateTableBucket(X) in us-west-2; CreateTableBucket(X) in us-east-1; GetTableBucket each;
    ListTableBuckets(prefix=X) each; ListTagsForResource each
  - measurements: create_latency_ms=[1911, 1253], n=2
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155`
  - related: [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-041](#s3t-tablebucket-041), [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130),
    [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159) · hypotheses: H-B-062 · evidence:
    tablebucket/cross-region/same-name-two-regions-arn-routing
  - notes: H-B-062 (per-account AND per-region name scope) confirmed for the create/identity part; the routing
    part is in the companion finding. A controller must pin each CR to one region: the same spec name in two
    regions is two unrelated buckets.
  - full notes: [details/S3T-TABLEBUCKET-125.md](details/S3T-TABLEBUCKET-125.md)

## Handling gaps (bugs to file)

- [S3T-TABLEBUCKET-004](#s3t-tablebucket-004) - CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms
  after Create, no status/state field exists (partially handled)
  Confirms H-B-002: no CREATING/DELETING phase was observable; the controller needs no requeue after create
  and can fill Status (ownerAccountId, createdAt, tableBucketId, type) from one GetTableBucket right after
  Create.
  - handling_ref: `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14;
    generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118;
    pkg/resource/table_bucket/sdk.go:303-304`
- [S3T-TABLEBUCKET-019](#s3t-tablebucket-019) - DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with
  only an empty namespace -> same code (suspected bug)
  Suspected controller bug confirmed by evidence: DeleteTableBucket/DeleteNamespace on a non-empty parent
  return BadRequestException 400 ('The bucket|namespace that you tried to delete is not empty.'), never
  ConflictException (-019, -044, -132), so with BadRequestException in terminal_codes a TableBucket or
  Namespace CR deleted before its children takes the Terminal branch instead of being retried, although a
  retry succeeds 0.5 s after the last child is gone (-019); treat this 400 (match the message) as retryable or
  hold the parent's finalizer until the child CRs are gone.
  - handling_ref: `test/e2e/conftest.py:102-106; test/e2e/conftest.py:141-142; generator.yaml:99-104;
    generator.yaml:170-175`
- [S3T-TABLEBUCKET-030](#s3t-tablebucket-030) - TagResource length violations return the NON-modeled ValidationException (HTTP 400),
  unlike CreateTableBucket's BadRequestException (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven: TagResource rejects a key >128
  chars, an empty key or a value >256 chars with the non-modeled code ValidationException 400 (-030, -029),
  which is not in terminal_codes, so a TableBucket whose spec.tags gain such a key/value after creation
  requeues on the tag sync forever instead of going Terminal, while the same tags at create are
  BadRequestException and Terminal (-025); add ValidationException to terminal_codes or enforce tag key/value
  lengths in the CRD.
  - handling_ref: `generator.yaml:99-104; generator.yaml:170-175`
- [S3T-TABLEBUCKET-122](#s3t-tablebucket-122) - ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6
  fields, raw createdAt strings identical) (partially handled)
  H-B-004 confirmed: a controller can adopt/de-duplicate by name from one ListTableBuckets(prefix) call
  without a follow-up Get; no field is dropped or reformatted in the summary.
  - handling_ref: `generator.yaml:179-187; templates/hooks/namespace/sdk_read_one_post_set_output.go.tpl:1-14;
    generator.yaml:45-46; generator.yaml:60-65; test/e2e/tests/test_table_bucket.py:115-118;
    pkg/resource/table_bucket/sdk.go:303-304`
- [S3T-TABLEBUCKET-130](#s3t-tablebucket-130) - ListTableBuckets(type=aws) is empty in us-west-2 and us-east-1; own bucket
  type=customer in Get and List; aws-type bucket not reachable (partially handled)
  H-N-039: aws-type half UNTESTED - not reachable (no AWS-managed table bucket exists in the account and no
  feature was enabled to create one). Observed half: GetTableBucket.type and TableBucketSummary.type are
  'customer' for a customer-created bucket, the type filter partitions the listing, and an invalid filter
  value is rejected server-side with BadRequestException 400. Single run, both regions.
  - handling_ref: `generator.yaml:45-46; generator.yaml:60-65`
- [S3T-TABLEBUCKET-159](#s3t-tablebucket-159) - CreateTableBucket ConflictException has two meanings (exists vs post-delete
  transitional); only the message and a Get tell them apart (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven: while the status ARN is unset
  the read path reports NotFound and the reconciler creates, but CreateTableBucket for a name that already
  exists is ConflictException 409 'already exists, and you own it' (-017, -159) and for a name deleted <10 min
  ago 409 'transitional state' (-115), both recoverable, so a CR for a pre-existing bucket (or one re-created
  after a controller restart before the status patch, or deleted and re-created within 10 min) requeues on
  Create indefinitely (or ~10 min) and never reaches Get; derive the deterministic ARN
  arn:<partition>:s3tables:<region>:<account>:bucket/<name> (-003) in the read path, or on this 409 Get the
  derived ARN and adopt on 200 / requeue on 404.
  - handling_ref: `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLEBUCKET-003](#s3t-tablebucket-003) | CreateTableBucket returns only {arn}; ARN is arn:aws:s3tables:<region>:<account>:bucket/<name>, constructible before the call | create_latency_ms=[1580, 1556, 1477, 1470, 1514] | 5 |
| [S3T-TABLEBUCKET-004](#s3t-tablebucket-004) | CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms after Create, no status/state field exists | first_get_ok_s=[0.04, 0.03, 0.03, 0.03, 0.03], createdAt_minus_t0_s=[1.504, 1.509, 1.432, 1.419, 1.468] | 5 |
| [S3T-TABLEBUCKET-014](#s3t-tablebucket-014) | DeleteTableBucket -> 204 empty body; GetTableBucket NotFoundException from the first poll (0.02 s); List(prefix) drops it by 0.009 s | first_404_s=0.02, list_gone_s=0.009 | 1 |
| [S3T-TABLEBUCKET-017](#s3t-tablebucket-017) | Two concurrent CreateTableBucket calls for one name: outcome pattern ConflictException+OK in 5/5 rounds; List shows one bucket | rounds=5, create_latency_ms=[2049, 1499, 1628, 2039, 2046, 1463, 1454, 2038, 1482, 2037] | 5 |
| [S3T-TABLEBUCKET-018](#s3t-tablebucket-018) | CreateTableBucket same-name 409 storm at ~8 calls/s is throttled: TooManyRequestsException 429 in runs of 4-6 after 5-9 accepted (~5 TPS) | calls=500, throttled_429=186, window_s=60, approx_accept_rate_per_s=5 | 1 |
| [S3T-TABLEBUCKET-019](#s3t-tablebucket-019) | DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with only an empty namespace -> same code | retries=3, namespace_delete_to_bucket_delete_s=0.549 | 1 |
| [S3T-TABLEBUCKET-031](#s3t-tablebucket-031) | 50-tag limit is cumulative per bucket: 30+25 -> BadRequestException 400 (none applied); +20 -> 200 (50); 51st -> 400; update at 50 -> 200 | max_tags_observed=50 | 1 |
| [S3T-TABLEBUCKET-040](#s3t-tablebucket-040) | ListTableBuckets prefix is a plain string prefix; results sorted lexically by name (not createdAt), stable across maxBuckets=1 pages | list_repeats=3 | 3 |
| [S3T-TABLEBUCKET-043](#s3t-tablebucket-043) | UntagResource on a table ARN with an unknown table uuid returns InternalServerErrorException 500 in 3/3 tries (ListTags/Tag -> 404) | tries=3, http_500=3 | 3 |
| [S3T-TABLEBUCKET-047](#s3t-tablebucket-047) | Deleted bucket name blocked with ConflictException for 604.7 s after DeleteTableBucket (n=2); re-created bucket: empty tags, default config | blocked_window_s=[604.7, 604.7], attempts=[57, 57] | 2 |
| [S3T-TABLEBUCKET-068](#s3t-tablebucket-068) | Burst of 5 then 4 concurrent CreateTableBucket (distinct names, 2 existing): 9/9 succeeded, no 429/409 | burst5_wall_s=1.89, burst4_wall_s=1.75, burst5_latency_ms_min_max=[1502, 1884], burst4_latency_ms_min_max=[1520, 1744] | 9 |
| [S3T-TABLEBUCKET-070](#s3t-tablebucket-070) | 10 concurrent TagResource on one bucket -> 8-9/10 InternalServerErrorException 500 (never 409); some 500-failed calls still applied the tag | tag_burst_wall_s=[0.63, 0.18], tag_latency_ms_min_max=[[50,631], [29,178]] | 20 |
| [S3T-TABLEBUCKET-071](#s3t-tablebucket-071) | Hypothesis of a 10-bucket default quota refuted: 11 sequential CreateTableBucket calls up to 22 buckets in us-west-2 all 200, no quota error | create_latency_ms=[1625, 1536, 1495, 1529, 1460, 1437, 1453, 1465, 1556, 1575, 1661] | 11 |
| [S3T-TABLEBUCKET-115](#s3t-tablebucket-115) | CreateTableBucket with a just-deleted name -> ConflictException 409 'transitional state ... previous deletion attempt' ~10 min (OK +606 s) | recreate_409_last_seen_s=594.6, recreate_success_s=606.2, attempts=45 | 1 |
| [S3T-TABLEBUCKET-121](#s3t-tablebucket-121) | Hypothesis of a transient 404 after CreateTableBucket refuted: Get, List(prefix) and ListTags all 200 on poll #1 in 5/5, no flapping | first_ok_ms_get=[0, 0, 0, 0, 0], first_ok_ms_list=[0, 0, 0, 0, 0], first_ok_ms_tags=[0, 0, 0, 0, 0], create_latency_ms=[1678, 1454, 1540, 1458, 1538], polls_per_trial=[5, 5, 5, 5, 5], poll_interval_s=0.1 | 5 |
| [S3T-TABLEBUCKET-122](#s3t-tablebucket-122) | ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6 fields, raw createdAt strings identical) | trials_equal=5, raw_createdAt_equal=5 | 5 |
| [S3T-TABLEBUCKET-124](#s3t-tablebucket-124) | ListTagsForResource at +0 ms after TagResource/UntagResource on a table bucket: 10/10 and 10/10 consistent; +500 ms 10/10 and 10/10 | tag_read0_consistent=10, tag_read500_consistent=10, untag_read0_consistent=10, untag_read500_consistent=10, tag_latency_ms=[52, 59, 66, 61, 44, 53, 54, 54, 49, 52], untag_latency_ms=[55, 50, 51, 46, 65, 51, 52, 44, 49, 47] | 10 |
| [S3T-TABLEBUCKET-125](#s3t-tablebucket-125) | Same bucket name in us-west-2 and us-east-1 -> both Create 200, ARNs differ only in region, distinct tableBucketIds, List/tags per region | create_latency_ms=[1911, 1253] | 2 |
| [S3T-TABLEBUCKET-131](#s3t-tablebucket-131) | A rejected CreateTableBucket (400) or CreateTable (400/500) does NOT reserve the name: the immediate same-name retry -> 200 (6/6) | bucket_retry_gap_ms=[1581, 1609, 1610], table_retry_gap_ms=[346, 387, 403] | 6 |
| [S3T-TABLEBUCKET-158](#s3t-tablebucket-158) | Delete->404->CreateTableBucket every 15 s: name blocked ~609 s whether the bucket was empty, populated or not retried until +420 s | recreate_ok_at_s.empty=609.4, recreate_ok_at_s.full=608.2, recreate_ok_at_s.late_first_retry_420s=608.8, retry_interval_s=15, n_409_per_name.empty=40, n_409_per_name.full=40, n_409_per_name.late=12 | 3 |
| [S3T-TABLEBUCKET-161](#s3t-tablebucket-161) | Doc claim C107 TRUE: Right after DeleteTableBucket, GetTableBucket -> 404 and ListTableBuckets drops it on the first poll (<=0.02 s, n=2) | first_404_s=[0.02, 0.009], list_gone_s=[0.009] | 2 |
| [S3T-TABLEBUCKET-165](#s3t-tablebucket-165) | Doc claim C113 UNTESTABLE: A GetTableBucket burst (~200 req/s) eventually yields TooManyRequestsException 429, not 503 SlowDown | burst_calls=30, burst_wall_s=[0.193], achieved_rps=[155.4], throttled_429=0, latency_ms_min_max=[5, 33], other_probes_running=3 | 1 |
| [S3T-TABLEBUCKET-179](#s3t-tablebucket-179) | 60 sequential CreateTableBucket calls (0.5 s pacing, 1.4-1.8 s latency) all 200: no quota error up to 62 buckets in us-west-2 (one account) | creates_ok=60, create_latency_ms_min_max=[1403, 1803] | 60 |
| [S3T-TABLEBUCKET-184](#s3t-tablebucket-184) | Post-delete name reservation measured: same-name CreateTableBucket -> 409 transitional until 570/571, succeeds at 602/603 s (n=2) | recreate_success_at_s=[602.1, 602.9], last_transitional_409_at_s=[570.4, 570.7], other_region_create_at_s=[1.25, 0.97], attempts_per_bucket=[22, 22] | 2 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `tablebucket.md` (s3tables): unverified, single-run or untestable items worth a future probe.
One line each, with the finding id that raised it.

- [S3T-TABLEBUCKET-165](#s3t-tablebucket-165): GetTableBucket read throttling never observed (30 calls at ~155 req/s, 0 x 429); the ~200 req/s
  cell needs a lab burst cap above 30 calls / 5 s to test whether reads throttle as 429 or something else.
- [S3T-TABLEBUCKET-018](#s3t-tablebucket-018): the ~5 TPS CreateTableBucket token bucket is an estimate from one 60 s run while other probers
  shared the account; re-run alone to pin the accept rate and burst size.
- [S3T-TABLEBUCKET-019](#s3t-tablebucket-019): the "not empty" 400 -> 204 transition was timed once (0.549 s after DeleteNamespace, n=1); a
  tight poll after the last child delete would give a range for the parent's requeue.
- [S3T-TABLEBUCKET-014](#s3t-tablebucket-014): post-delete visibility (Get 404 at 0.02 s, List gone at 0.009 s) is n=1 here, n=2 in -161; more
  trials would confirm there is never a late 200 after the first 404.
- [S3T-TABLEBUCKET-130](#s3t-tablebucket-130): no AWS-managed `aws`-type table bucket exists in the account, so GetTableBucket, ListTableBuckets
  and CreateNamespace against one (and whether the controller could adopt it) are unexercised.
- [S3T-TABLEBUCKET-159](#s3t-tablebucket-159): the "409 + Get 404 => transitional" rule rests on one creative run plus the Phase 8 re-verify
  (-184, n=2); the cross-region create inside the window is n=1 in -159 and n=2 in -184.
- [S3T-TABLEBUCKET-017](#s3t-tablebucket-017): two concurrent same-name Creates always split ConflictException + OK (5/5); not probed with more
  than two writers or across regions.
- [S3T-TABLEBUCKET-071](#s3t-tablebucket-071) / -179: no account quota was reached at 62 buckets in one region; the actual limit and its error
  code are unknown.
- [S3T-TABLEBUCKET-070](#s3t-tablebucket-070): the 500s under concurrent TagResource were observed in 2 bursts on one bucket; whether the same
  happens for concurrent UntagResource or mixed Tag/Untag is untested.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-tablebucket-002"></a>**S3T-TABLEBUCKET-002** | response-fidelity | low | confirmed | GetTableBucket body byte-identical after creating and deleting 6 namespaces: six fixed members, no namespace count or modifiedAt | [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130), [S3T-TABLEBUCKET-042](#s3t-tablebucket-042) | - |
| <a id="s3t-tablebucket-007"></a>**S3T-TABLEBUCKET-007** | tag-semantics | low | confirmed | ListTagsForResource on a bucket created with one user tag returns only that tag: no system tags, no pagination token | [S3T-TABLEBUCKET-027](#s3t-tablebucket-027), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLEBUCKET-124](#s3t-tablebucket-124), [S3T-TABLEBUCKET-032](#s3t-tablebucket-032), [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLEBUCKET-156](#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](#s3t-tablebucket-157), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008) | - |
| <a id="s3t-tablebucket-016"></a>**S3T-TABLEBUCKET-016** | identity | high | confirmed | CreateTableBucket with a just-deleted name is rejected with ConflictException 'transitional state ... previous deletion attempt' for >60 s | [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-159](#s3t-tablebucket-159), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127) | [S3T-TABLEBUCKET-115](#s3t-tablebucket-115) |
| <a id="s3t-tablebucket-017"></a>**S3T-TABLEBUCKET-017** | idempotency | low | confirmed | Two concurrent CreateTableBucket calls for one name: outcome pattern ConflictException+OK in 5/5 rounds; List shows one bucket | [S3T-TABLEBUCKET-159](#s3t-tablebucket-159), [S3T-TABLEBUCKET-115](#s3t-tablebucket-115), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-160](#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](#s3t-tablebucket-167), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) | - |
| <a id="s3t-tablebucket-032"></a>**S3T-TABLEBUCKET-032** | tag-semantics | low | confirmed | Bucket and table tags are independent: TagResource(table ARN) invisible on the bucket ARN; UntagResource(bucket) never touches table tags | [S3T-TABLEBUCKET-027](#s3t-tablebucket-027), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-124](#s3t-tablebucket-124), [S3T-TABLE-128](table.md#s3t-table-128), [S3T-TABLE-263](table.md#s3t-table-263), [S3T-TABLE-264](table.md#s3t-table-264), [S3T-TABLE-185](table.md#s3t-table-185), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-114](table.md#s3t-table-114), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLE-143](table.md#s3t-table-143) | - |
| <a id="s3t-tablebucket-039"></a>**S3T-TABLEBUCKET-039** | identity | low | confirmed | Raw GET /buckets/<arn>: single-encoded ARN -> 200, double-encoded -> 400, unencoded (literal ':' '/') -> 404 | [S3T-TABLEBUCKET-022](#s3t-tablebucket-022), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-033](#s3t-tablebucket-033), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-SERVICE-001](service.md#s3t-service-001), [S3T-SERVICE-023](service.md#s3t-service-023) | - |
| <a id="s3t-tablebucket-042"></a>**S3T-TABLEBUCKET-042** | identity | low | confirmed | No S3 Metadata in region: GetTableBucket(...:bucket/aws-s3) -> NotFoundException 404; ListTableBuckets(type=aws) -> 0, list only customer | [S3T-TABLEBUCKET-002](#s3t-tablebucket-002), [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-130](#s3t-tablebucket-130) | - |
| <a id="s3t-tablebucket-044"></a>**S3T-TABLEBUCKET-044** | delete-semantics | high | confirmed | DeleteTableBucket with a table inside -> BadRequestException 400 'not empty'; only an empty namespace left -> same 400; then empty -> 204 | [S3T-TABLEBUCKET-019](#s3t-tablebucket-019), [S3T-TABLEBUCKET-132](#s3t-tablebucket-132), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070), [S3T-NAMESPACE-024](namespace.md#s3t-namespace-024), [S3T-NAMESPACE-037](namespace.md#s3t-namespace-037) | [S3T-TABLEBUCKET-019](#s3t-tablebucket-019) |
| <a id="s3t-tablebucket-068"></a>**S3T-TABLEBUCKET-068** | quota-limit | low | confirmed | Burst of 5 then 4 concurrent CreateTableBucket (distinct names, 2 existing): 9/9 succeeded, no 429/409 | [S3T-TABLEBUCKET-018](#s3t-tablebucket-018), [S3T-TABLEBUCKET-179](#s3t-tablebucket-179), [S3T-TABLEBUCKET-071](#s3t-tablebucket-071) | - |
| <a id="s3t-tablebucket-071"></a>**S3T-TABLEBUCKET-071** | quota-limit | low | confirmed (hypothesis refuted) | Hypothesis of a 10-bucket default quota refuted: 11 sequential CreateTableBucket calls up to 22 buckets in us-west-2 all 200, no quota error | [S3T-TABLEBUCKET-018](#s3t-tablebucket-018), [S3T-TABLEBUCKET-068](#s3t-tablebucket-068), [S3T-TABLEBUCKET-179](#s3t-tablebucket-179) | - |
| <a id="s3t-tablebucket-130"></a>**S3T-TABLEBUCKET-130** | identity | low | confirmed | ListTableBuckets(type=aws) is empty in us-west-2 and us-east-1; own bucket type=customer in Get and List; aws-type bucket not reachable | [S3T-TABLEBUCKET-002](#s3t-tablebucket-002), [S3T-TABLEBUCKET-004](#s3t-tablebucket-004), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-042](#s3t-tablebucket-042), [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-041](#s3t-tablebucket-041), [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-125](#s3t-tablebucket-125), [S3T-TABLEBUCKET-166](#s3t-tablebucket-166) | - |
| <a id="s3t-tablebucket-132"></a>**S3T-TABLEBUCKET-132** | delete-semantics | low | confirmed | Rejected DeleteTableBucket/DeleteNamespace (400 'not empty') leave no mark: CreateNamespace/CreateTable in the parent right after -> 200 | [S3T-TABLEBUCKET-019](#s3t-tablebucket-019), [S3T-NAMESPACE-024](namespace.md#s3t-namespace-024), [S3T-TABLEBUCKET-044](#s3t-tablebucket-044), [S3T-TABLEBUCKET-070](#s3t-tablebucket-070) | - |
| <a id="s3t-tablebucket-148"></a>**S3T-TABLEBUCKET-148** | other | low | confirmed | Doc claim C024 TRUE: TagResource with 50 keys in one call on an untagged bucket -> HTTP 200; ListTagsForResource returns 50 keys | [S3T-TABLEBUCKET-031](#s3t-tablebucket-031), [S3T-TABLE-185](table.md#s3t-table-185), [S3T-TABLEBUCKET-025](#s3t-tablebucket-025), [S3T-TABLEBUCKET-149](#s3t-tablebucket-149), [S3T-TABLEBUCKET-150](#s3t-tablebucket-150) | - |
| <a id="s3t-tablebucket-149"></a>**S3T-TABLEBUCKET-149** | other | low | confirmed | Doc claim C025 TRUE: TagResource with 51 keys in one call on an untagged bucket -> BadRequestException 400; 0 tags afterwards (atomic) | [S3T-TABLEBUCKET-031](#s3t-tablebucket-031), [S3T-TABLE-185](table.md#s3t-table-185), [S3T-TABLEBUCKET-025](#s3t-tablebucket-025), [S3T-TABLEBUCKET-148](#s3t-tablebucket-148), [S3T-TABLEBUCKET-150](#s3t-tablebucket-150) | - |
| <a id="s3t-tablebucket-150"></a>**S3T-TABLEBUCKET-150** | other | low | confirmed | Doc claim C026 TRUE: Adding a 51st key to a bucket carrying 50 tags -> BadRequestException 400; the 50 are unchanged (cumulative limit) | [S3T-TABLEBUCKET-031](#s3t-tablebucket-031), [S3T-TABLE-185](table.md#s3t-table-185), [S3T-TABLEBUCKET-148](#s3t-tablebucket-148), [S3T-TABLEBUCKET-149](#s3t-tablebucket-149), [S3T-TABLEBUCKET-025](#s3t-tablebucket-025) | - |
| <a id="s3t-tablebucket-151"></a>**S3T-TABLEBUCKET-151** | other | low | confirmed | Doc claim C028 TRUE: TagResource is an upsert: {k1: v2} on {k1: v1, k2: w} -> {k1: v2, k2: w} (other keys retained, not replace-all) | [S3T-TABLEBUCKET-027](#s3t-tablebucket-027), [S3T-TABLE-114](table.md#s3t-table-114) | - |
| <a id="s3t-tablebucket-152"></a>**S3T-TABLEBUCKET-152** | other | low | confirmed | Doc claim C029 TRUE: TagResource with an 'aws:' key -> BadRequestException 400 on a bucket (live) and on a table ([S3T-TABLE-118](table.md#s3t-table-118)) | [S3T-TABLE-118](table.md#s3t-table-118), [S3T-TABLEBUCKET-025](#s3t-tablebucket-025) | - |
| <a id="s3t-tablebucket-154"></a>**S3T-TABLEBUCKET-154** | other | low | confirmed | Doc claim C031 TRUE: UntagResource with an empty tagKeys list -> BadRequestException 400 'The provided tags are invalid.' (bucket and table) | [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLEBUCKET-153](#s3t-tablebucket-153), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLEBUCKET-155](#s3t-tablebucket-155), [S3T-TABLE-116](table.md#s3t-table-116) | - |
| <a id="s3t-tablebucket-155"></a>**S3T-TABLEBUCKET-155** | other | low | confirmed | Doc claim C032 TRUE: UntagResource naming an absent key -> HTTP 204 (no NotFoundException); present keys in the same call are removed | [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLE-125](table.md#s3t-table-125), [S3T-TABLE-116](table.md#s3t-table-116), [S3T-TABLEBUCKET-154](#s3t-tablebucket-154) | - |
| <a id="s3t-tablebucket-156"></a>**S3T-TABLEBUCKET-156** | other | low | confirmed | Doc claim C034 TRUE: ListTagsForResource on a never-tagged table bucket (live) or table -> HTTP 200 with an empty tags map, not 404 | [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028), [S3T-TABLEBUCKET-157](#s3t-tablebucket-157), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008) | - |
| <a id="s3t-tablebucket-157"></a>**S3T-TABLEBUCKET-157** | other | low | confirmed | Doc claim C035 TRUE: Tags from CreateTableBucket.tags / CreateTable.tags are listed by ListTagsForResource right after the create returns | [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLEBUCKET-007](#s3t-tablebucket-007), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-156](#s3t-tablebucket-156), [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-028](#s3t-tablebucket-028) | - |
| <a id="s3t-tablebucket-160"></a>**S3T-TABLEBUCKET-160** | other | low | confirmed | Doc claim C105 TRUE: Second CreateTableBucket with an existing name -> ConflictException 409; identical retries keep returning 409 | [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017), [S3T-TABLEBUCKET-167](#s3t-tablebucket-167), [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) | - |
| <a id="s3t-tablebucket-161"></a>**S3T-TABLEBUCKET-161** | other | low | confirmed | Doc claim C107 TRUE: Right after DeleteTableBucket, GetTableBucket -> 404 and ListTableBuckets drops it on the first poll (<=0.02 s, n=2) | [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-101](tablebucket-replication.md#s3t-tablebucket-101), [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-015](#s3t-tablebucket-015) | - |
| <a id="s3t-tablebucket-162"></a>**S3T-TABLEBUCKET-162** | other | low | confirmed | Doc claim C108 TRUE: Second DeleteTableBucket on an already deleted bucket -> NotFoundException 404, not 204 | [S3T-TABLEBUCKET-015](#s3t-tablebucket-015), [S3T-TABLEBUCKET-161](#s3t-tablebucket-161), [S3T-TABLEBUCKET-014](#s3t-tablebucket-014), [S3T-TABLEBUCKET-101](tablebucket-replication.md#s3t-tablebucket-101), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163), [S3T-TABLEBUCKET-164](#s3t-tablebucket-164), [S3T-TABLEBUCKET-038](#s3t-tablebucket-038), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029) | - |
| <a id="s3t-tablebucket-164"></a>**S3T-TABLEBUCKET-164** | other | low | confirmed | Doc claim C112 TRUE: GetTableBucket with a malformed ARN (wrong partition/service/resource format) -> BadRequestException 400, not 404 | [S3T-TABLEBUCKET-038](#s3t-tablebucket-038), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-127](#s3t-tablebucket-127), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029), [S3T-TABLEBUCKET-162](#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163) | - |
| <a id="s3t-tablebucket-165"></a>**S3T-TABLEBUCKET-165** | other | low | confirmed | Doc claim C113 UNTESTABLE: A GetTableBucket burst (~200 req/s) eventually yields TooManyRequestsException 429, not 503 SlowDown | [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](#s3t-tablebucket-018), [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094) | - |
| <a id="s3t-tablebucket-166"></a>**S3T-TABLEBUCKET-166** | other | low | confirmed | Doc claim C114 TRUE: GetTableBucket.type='customer' for created buckets; ListTableBuckets type=aws excludes them, type=customer lists them | [S3T-TABLEBUCKET-130](#s3t-tablebucket-130) | - |
| <a id="s3t-tablebucket-167"></a>**S3T-TABLEBUCKET-167** | other | low | confirmed | Doc claim C115 TRUE: CreateTableBucket.arn == GetTableBucket.arn == ListTableBuckets arn; usable at once in CreateNamespace | [S3T-TABLEBUCKET-003](#s3t-tablebucket-003), [S3T-TABLEBUCKET-122](#s3t-tablebucket-122), [S3T-TABLEBUCKET-121](#s3t-tablebucket-121), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019), [S3T-TABLEBUCKET-160](#s3t-tablebucket-160), [S3T-TABLEBUCKET-013](tablebucket-encryption-storage.md#s3t-tablebucket-013), [S3T-TABLEBUCKET-017](#s3t-tablebucket-017) | - |
| <a id="s3t-tablebucket-168"></a>**S3T-TABLEBUCKET-168** | other | low | confirmed | Doc claim C036 TRUE: ListTableBuckets maxBuckets=1 -> exactly 1 bucket + continuationToken; following the token yields the next bucket | [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-169](#s3t-tablebucket-169) | - |
| <a id="s3t-tablebucket-169"></a>**S3T-TABLEBUCKET-169** | other | low | confirmed | Doc claim C037 TRUE: ListTableBuckets never returns more than maxBuckets per page; the last page omits continuationToken (key absent) | [S3T-TABLEBUCKET-069](#s3t-tablebucket-069), [S3T-TABLEBUCKET-040](#s3t-tablebucket-040), [S3T-TABLEBUCKET-168](#s3t-tablebucket-168) | - |
| <a id="s3t-tablebucket-179"></a>**S3T-TABLEBUCKET-179** | quota-limit | low | confirmed | 60 sequential CreateTableBucket calls (0.5 s pacing, 1.4-1.8 s latency) all 200: no quota error up to 62 buckets in us-west-2 (one account) | [S3T-TABLEBUCKET-018](#s3t-tablebucket-018), [S3T-TABLEBUCKET-068](#s3t-tablebucket-068), [S3T-TABLEBUCKET-071](#s3t-tablebucket-071) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
