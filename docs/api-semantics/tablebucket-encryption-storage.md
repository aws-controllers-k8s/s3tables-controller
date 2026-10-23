<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# TableBucket encryption (SSE-S3 / SSE-KMS) and storage class
_Put/Get/DeleteTableBucketEncryption and Put/GetTableBucketStorageClass: defaults, KMS key policy and key-state prerequisites, caller permissions, inheritance by tables, the ValidationException-vs-BadRequestException split._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `tablebucket-encryption-storage.md`; every claim cites the finding ids that carry the
full evidence below.

### Orientation

This document covers the two bucket-level configuration pairs that live outside CreateTableBucket/GetTableBucket:
Put/Get/DeleteTableBucketEncryption (`sseAlgorithm` `AES256` or `aws:kms` plus `kmsKeyArn`) and
Put/GetTableBucketStorageClass (`STANDARD` or `INTELLIGENT_TIERING`). Both may be set at CreateTableBucket but are never
echoed by GetTableBucket or ListTableBuckets; each is read back only through its own Get, which on a fresh bucket returns
HTTP 200 with a server default (AES256 / STANDARD) rather than 404, byte-identical to the body after an explicit Put of
the same value ([S3T-TABLEBUCKET-008](#s3t-tablebucket-008), -147, -005, -083, -001). There is no state machine: a Put is synchronous and visible
on the first Get (0.01 s, 20/20 immediate reads), each noun is a whole-value replace where the latest Put wins, an
unchanged re-send is 200, and the only "unset" path is DeleteTableBucketEncryption (204, then the AES256 default body;
storage class has no Delete) (-037, -140, -141, -143, -009, -145, -146, -058, -108, -051). The bucket value is a default
snapshotted at CreateTable time: existing tables keep their encryption and storage class after a bucket Put or Delete,
only tables created afterwards inherit the new value (-084, -120, -072, -073). SSE-KMS has two prerequisite layers
checked at different moments: key form, region, existence and key state are validated at Create/Put time
(BadRequestException 400, nothing stored), while the key policy's grant to `maintenance.s3tables.amazonaws.com` is
checked only when a table is created (-107, -185, -119, -180, -139, -142).

### Rules

1. Never send a `storageClass` outside the closed, case-sensitive enum from the update path: PutTableBucketStorageClass
   rejects `standard`, `Standard`, `STANDARD_IA`, `GLACIER`, `INTELLIGENT-TIERING`, `''` and `{}` with wire code
   ValidationException 400 (7/7), a shape not in the API model, while CreateTableBucket rejects the same values with
   BadRequestException 400; `sseAlgorithm` is likewise exactly `AES256` or `aws:kms` (`AWS_KMS`, `aws_kms`, `AWS:KMS`,
   `Aes256`, `aes256` -> BadRequestException 400) ([S3T-TABLEBUCKET-087](#s3t-tablebucket-087), -024, -104, -137, -023).
2. Omit `kmsKeyArn` entirely for AES256 and send a full key ARN for aws:kms: aws:kms without a key or with `''`, and
   AES256 with any key (even `''`), are BadRequestException 400 at both CreateTableBucket and PutTableBucketEncryption,
   and the stored value is unchanged after a rejected Put (4/4) (-102, -023, -034, -037, -135, -136).
3. Do not treat DeleteTableBucketEncryption as idempotent: it is 204 only when a configuration exists; on a
   never-configured bucket or on a second Delete it is NotFoundException 404 'The specified encryption configuration does
   not exist.' (20/20 second deletes), and the Get afterwards is the AES256 default body, never 404 (-049, -108, -037,
   -051, -052, -144).
4. A bucket Put or Delete does not rewrite existing tables: a table created before keeps aws:kms+key (unchanged 10 s
   later) or its storage class, tables created 0.49-6.68 s after the mutation inherit the new default (0 lagging of 10),
   and GetTableBucketStorageClass returns the bucket default even when a table in it uses INTELLIGENT_TIERING (-120,
   -072, -073, -084, -085).
5. Expect SSE-KMS to succeed at the bucket and fail at CreateTable: a key whose policy lacks a statement for
   `maintenance.s3tables.amazonaws.com` is accepted by CreateTableBucket and PutTableBucketEncryption (200, Get echoes
   the key) and every CreateTable in that bucket is BadRequestException 400 'Insufficient access to perform table
   maintenance...' (with or without metadata) until the key policy grants the principal (-119, -180, -139, -142).
6. KMS key validation at Create/Put is BadRequestException 400 for every class but one: nonexistent key ('does not
   exist', not 404), key in another region ('Cross region KMS keys are not supported.', checked before existence),
   Disabled or PendingDeletion key ('The specified KMS key ARN is disabled.', 30/30 cells), bare key id / alias name /
   alias ARN / upper-cased region or prefix; another account's key is ForbiddenException 403 naming kms:DescribeKey; no
   bucket is created and nothing is stored (-035, -107, -185, -105, -128, -103, -034, -138).
7. A Disabled key is a recoverable condition: PutTableBucketEncryption and CreateTable stay 400 at +3/+30/+120/+300 s
   after DisableKey (warm key and 600 s idle key alike) and succeed on the first poll after EnableKey, while
   GetTableBucketEncryption returns 200 with the disabled key ARN throughout; CreateTable's message is the same
   'Insufficient access to perform table maintenance' text as a missing grant (-182, -183, -185).
8. Two 403 codes: the caller's own kms:DescribeKey denied -> ForbiddenException 403 on PutTableBucketEncryption(aws:kms)
   (CreateTable needs kms:GenerateDataKey and does not call DescribeKey); an s3tables identity-policy or bucket-policy
   Deny -> AccessDeniedException 403; GetTableBucketEncryption needs no KMS access (200 under Deny kms:*) (-181, -045,
   -106).
9. CreateTableBucket with tags, encryptionConfiguration or storageClassConfiguration needs the matching sub-resource
   action: a role allowed only Create/Get/DeleteTableBucket gets 200 for a plain Create but AccessDeniedException 403
   naming s3tables:TagResource, s3tables:PutTableBucketEncryption (even for AES256) or s3tables:PutTableBucketStorageClass
   (even for STANDARD), tags checked first, and no bucket is created (-129).
10. A rejected CreateTableBucket (400) leaves no bucket and reserves nothing: Get on the constructed ARN is 404, List is
    empty, an immediate same-name retry is 200 (6/6); a 409 on re-create means the bucket exists (unchanged whether or
    not encryption/tags match) or was deleted less than ~605 s ago (-023, -035, -131, -013, -047).
11. Identifier and region errors on the encryption/storage-class Gets are BadRequestException 400 (bare name, upper-case
    name, trailing slash, table-style suffix, a table ARN as tableBucketARN, an other-region bucket ARN on this
    endpoint); only a well-formed ARN of a missing or deleted bucket is NotFoundException 404 (-066, -085, -126, -020).
12. "Unset" and "explicitly default" are indistinguishable on read: Put(STANDARD) or Put(AES256) on a fresh bucket and
    DeleteTableBucketEncryption after Put(aws:kms) or Put(AES256) all return exactly the never-configured body;
    `kmsKeyArn` is absent from the JSON after AES256 and echoed byte-identical after aws:kms (-083, -005, -051, -052,
    -050, -133, -134).
13. These writes have no side effects on the rest of the bucket: `createdAt` and `tableBucketId` are byte-identical
    across 4 reads over 2 min spanning tag/encryption/storage-class Puts, tags are unchanged after every Put and after
    DeleteTableBucketEncryption, both may be set together at Create, and none of them (nor a policy or maintenance
    configuration) blocks DeleteTableBucket (204) (-010, -114, -117, -020).
14. No settle wait is needed between CreateTableBucket and these Puts: TagResource at +0 ms then PutTableBucketEncryption
    at +69-142 ms (5/5) and PutTableBucketStorageClass at <1 ms (12/12) succeeded on the first attempt; read-after-write
    was consistent 20/20 for both nouns and 0/60 reads were stale under alternating Puts (-123, -112, -108, -059).
15. Concurrent writes to these two nouns did not 500: 20 concurrent PutTableBucketStorageClass(STANDARD) -> 20 x 200;
    5 Put + 5 Delete TableBucketEncryption interleaved -> 5 x 200, 3 x 204, 2 x 404 (a Delete landing on no
    configuration); by contrast 10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (-082).

### Timing ranges

- Put-to-visible on the own Get: 0.01-0.01 s for PutTableBucketEncryption (n=2, [S3T-TABLEBUCKET-037](#s3t-tablebucket-037)) and for
  PutTableBucketStorageClass (n=2, -009); the bucket Get reflected a storage-class Put on the first read in 2/2 (-084).
- Read-after-write: the immediate Get matched the mutation 20/20 for storage class and 20/20 for encryption, the
  immediate second DeleteTableBucketEncryption was 404 20/20 (n=60 trials, -108); 10 alternating storage-class Puts with
  Gets at +0/+0.3/+1 s -> 0/30 stale, Get latency 7-15 ms (n=30, -059). Put latency 13-53 ms for storage class (n=20,
  -108); under concurrency 10-178 ms for encryption Put/Delete (n=10) and 16-226 ms for storage class (n=20), one burst
  each with walls of 0.18 s and 0.24 s (-082).
- First write after CreateTableBucket: PutTableBucketStorageClass sent 0-0 ms after the create response succeeded at
  564-655 ms from the create call (12/12, create latency 1435-1940 ms, n=12, -112); PutTableBucketEncryption sent
  69-142 ms after the create response succeeded 5/5 (create latency 1561-1823 ms, n=5, -123).
- Inheritance by new tables: tables created 0.49-6.68 s after a bucket Put(aws:kms) or DeleteTableBucketEncryption
  carried the new default 10/10 (n=10, -073).
- Disabled / PendingDeletion key: PutTableBucketEncryption(aws:kms) and CreateTableBucket(aws:kms) were 400 at
  +5/+30/+120 s after DisableKey and +5/+60 s after ScheduleKeyDeletion (3 trials per offset, 15/15 Put and 15/15 Create
  cells, n=30, -185); offsets below ~6 s were not probed. PutTableBucketEncryption and CreateTable were 400 at
  +3/+30/+120/+300 s in a warm-key series and in a 600 s idle-key series (n=1 series each, -182, -183).
- Recovery after EnableKey: PutTableBucketEncryption accepted 0.04-0.04 s after the call (polled every 2 s; n=2 series,
  one each in -182 and -183); no caching of the disabled state was observed.
- Key-policy grant propagation: CreateTable in an SSE-KMS bucket succeeded on its first attempt +5 s after PutKeyPolicy
  added the maintenance principal (2/2, -180) and after a 30 s wait (1 attempt, -120); nothing shorter than 5 s was
  probed, so budget 5-30 s before the first CreateTable retry.
- Caller KMS permission changes: the ForbiddenException/200 outcomes were stable 12-22 s after the identity-policy write
  (2 Puts per variant, -181).
- Post-delete same-name re-create (rule 10): ConflictException 'transitional state' for 604.7-604.7 s after
  DeleteTableBucket, 57 attempts each, n=2 (-047); the re-created bucket had default encryption/storage class and no tags.

### Handling gaps

- [S3T-TABLEBUCKET-087](#s3t-tablebucket-087) (suspect-bug, not yet handled): today `StorageClass` is a plain `*string` with no enum marker
  (apis/v1alpha1/types.go:163-164), `syncStorageClass` sends the raw string and returns the error unchanged
  (hooks.go:508-537), `terminal_codes` lists only BadRequestException (generator.yaml:267-268; sdk.go:400-402) and the
  read hook refills the live `STANDARD` so the delta never closes (hooks.go:98-105). Evidence: 7/7 invalid values ->
  ValidationException 400 on Put, the same value on CreateTableBucket -> BadRequestException 400 (-024). A misspelled
  `storageClass` set after create is re-Put on every reconcile forever; at create it goes Terminal. Fix direction from
  the finding: an enum marker on `storageClass`, or ValidationException in the TableBucket terminal codes.
- [S3T-TABLEBUCKET-023](#s3t-tablebucket-023) (partial, not yet handled): today the create path passes `encryptionConfiguration` and
  `storageClassConfiguration` as declared; `KMSKeyARN`, `SSEAlgorithm` and `StorageClass` are optional plain strings
  (types.go:32-36, 163-164), so an empty `encryptionConfiguration: {}` block or `aws:kms` without a key is admitted and
  sent; BadRequestException is Terminal. Evidence: `{}` for either block, `aws:kms` without `kmsKeyArn` and with `''` ->
  BadRequestException 400, no bucket; `tags={}` -> 200. The Terminal is correct, but the shapes pass admission.
- [S3T-TABLEBUCKET-035](#s3t-tablebucket-035) (partial, not yet handled): today `KMSKeyARN` is a literal string or a reference to a kms Key's
  `Status.ACKResourceMetadata.ARN` (generator.yaml:221-225) with no existence, region or account check. Evidence:
  nonexistent key -> 400 'does not exist' (not 404) and other-region key -> 400 'Cross region KMS keys are not
  supported.' (Terminal today, no bucket created); another account's key -> ForbiddenException 403 naming
  kms:DescribeKey, which is not in the terminal codes and is requeued as recoverable.
- [S3T-TABLEBUCKET-102](#s3t-tablebucket-102) (partial, not yet handled): today `syncEncryption` copies `KmsKeyArn` and `SseAlgorithm` from the
  spec unchanged (hooks.go:488-495), so `sseAlgorithm: AES256` with a `kmsKeyARN` (or `kmsKeyARN: ""`) and `aws:kms`
  without a key are sent as-is. Evidence: all four combinations -> BadRequestException 400 at Put, stored configuration
  unchanged (4/4). The finding's instruction - omit the member, never send `''` - is not implemented; today the result
  is Terminal with the live configuration unchanged.
- [S3T-TABLEBUCKET-107](#s3t-tablebucket-107) (partial, not yet handled): today nothing inspects the referenced key's `KeyState` before Put; the
  reference resolver only requires the kms Key CR to be Synced (references.go:174-201), and BadRequestException is
  Terminal. Evidence: nonexistent, cross-region, Disabled and PendingDeletion keys -> 400 at Put, previous value kept
  (n=1; key state settled 30/30 in -185). A Disabled key puts the TableBucket into Terminal although the identical Put
  succeeds on the first poll after EnableKey (-182, -183); no recoverable classification for 'is disabled.' exists.
- [S3T-TABLEBUCKET-103](#s3t-tablebucket-103) (partial, not yet handled): today `KMSKeyARN` carries no pattern or CEL rule (types.go:33), so a
  bare key id, alias name or alias ARN is admitted and sent. Evidence: five non-ARN forms -> BadRequestException 400 at
  Put, only the full key ARN -> 200, echoed byte-identical (n=6). A literal alias in the spec becomes a Terminal 400 at
  update rather than an admission error.
- [S3T-TABLEBUCKET-105](#s3t-tablebucket-105) (partial, not yet handled): today the kms Key reference is resolved by ARN with no comparison of
  the key's region against the bucket's (generator.yaml:221-225). Evidence: us-east-1 bucket with a us-west-2 key ->
  BadRequestException 400 'Cross region KMS keys are not supported.' at Put, Get stays AES256 (n=1); the same message is
  returned for a nonexistent key in another region (region checked before existence). A reference to a Key in another
  region is a Terminal 400 at update.

### Where next

- `tablebucket.md` is the catch-all for the bucket itself: ARN construction, the ~10 min post-delete name reservation
  behind rule 10 (-047), the "not empty" delete 400, tag validation and limits. `tablebucket-policy.md` carries the
  resource-policy side of rule 8 (AccessDeniedException vs ForbiddenException, the first-sync deletion hazard, principal
  propagation windows, -045) and the policy half of the concurrency burst (-082). `tablebucket-maintenance-metrics.md`
  carries the maintenance halves of -082, -059 and -001; `tablebucket-replication.md` the bucket-level replication rules.
- `table-encryption-storage.md` is the table side of rules 4, 5 and 7: create-only `encryptionConfiguration` /
  `storageClassConfiguration`, the creation-time snapshot of the bucket defaults, GetTableEncryption/GetTableStorageClass,
  KMS key-state and permission effects on CreateTable; `table-maintenance.md` carries the maintenance-principal
  prerequisite at table level and `table-policy.md` the per-operation IAM actions (kms:DescribeKey for aws:kms).
- Long catalogues render in `details/<id>.md` rather than inline: the identifier-form matrix for the three bucket Gets
  ([S3T-TABLEBUCKET-066](#s3t-tablebucket-066)), the 11-operation region-mismatch matrix (-126), the wire-code tuple table (-109), the deleted-ARN
  operation matrix (-020), the key-state offset table (-185) and the two disabled-key series (-182, -183), the caller
  KMS-permission variants V1-V3 (-181), the kmsKeyArn form matrices for Create and Put (-034, -103), the KMS 400 message
  classes (-035, -107), the concurrency burst table (-082), and the sub-resource actions a Create needs (-129).
- Open experiments and stale readings (STATUS.md, Phase 8): follow-up FU2-01 settled "disabled KMS key -> 400 always"
  (-185 canonical; -037's HTTP 200 for a key disabled ~5 s earlier is a superseded reading at confidence low) and
  "sub-resource Gets each need their own IAM action". Not probed: offsets below ~6 s after DisableKey (-185), an
  other-account key at PutTableBucketEncryption (Create side only, -035), a key-policy grant propagation shorter than
  5 s (-180, -120). The 8 KMS keys used here are PendingDeletion and expire 2026-10-20/21, so a re-run of -185 or -107
  needs fresh keys. The undeletable lab bucket belongs to `tablebucket-policy.md` ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007)), not here.
<!-- preserved:end -->

## At a glance

- canonical findings: 61 (high 18 / medium 25 / low 18); duplicates folded into the appendix: 5
- handling: handled 14 · partial 0 · tracked 0 · unhandled 46 · suspect-bug 1 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: other 13, request-validation 8, error-code 7, eventual-consistency 5, sub-resource-api 5,
  delete-semantics 4, prerequisite 4, idempotency 3, server-default 3, response-fidelity 2, codegen-artifact
  1, cross-region 1, identity 1, normalization 1, read-gap 1, shape-mismatch 1, tag-semantics 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketEncryption | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketPolicy | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableBucketMaintenanceConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketPolicy | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketStorageClass | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMaintenanceConfiguration | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListNamespaces | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTableBuckets | list | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketEncryption | create | tableBucketARN, encryptionConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketMaintenanceConfiguration | create | tableBucketARN, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketMetricsConfiguration | create | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketPolicy | create | tableBucketARN, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketStorageClass | create | tableBucketARN, storageClassConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-tablebucket-066"></a>**S3T-TABLEBUCKET-066** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **tableBucketARN forms: bare name -> BadRequestException 400, nonexistent bucket -> NotFoundException 404...**
  With a policy set on the bucket, 7 identifier forms were sent as tableBucketARN to three Get operations:
  bare_name -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; nonexistent_bucket -> GetTableBucket NotFoundException 404
  / GetTableBucketPolicy NotFoundException 404 / GetTableBucketEncryption NotFoundException 404;
  uppercase_name -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; trailing_slash -> GetTableBucket BadRequestException 400 /
  GetTableBucketPolicy BadRequestException 400 / GetTableBucketEncryption BadRequestException 400;
  uppercase_prefix -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; table_style_suffix -> GetTableBucket BadRequestException
  400 / GetTableBucketPolicy BadRequestException 400 / GetTableBucketEncryption BadRequestException 400;
  wrong_region -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400. Messages: bare name 'The specified table bucket ARN is not
  valid.'; upper-case name 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; trailing slash 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; wrong region 'The provided Region does not match the Region of the resource.'.
  - ACK: is_arn_primary_key, exceptions.404, terminal_codes · ops: GetTableBucket, GetTableBucketPolicy,
    GetTableBucketEncryption · fields: tableBucketARN
  - repro: Send each identifier form as tableBucketARN to GetTableBucket, GetTableBucketPolicy,
    GetTableBucketEncryption; record code/HTTP/message
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065),
    [S3T-TABLEBUCKET-109](#s3t-tablebucket-109), [S3T-TABLE-081](table.md#s3t-table-081), [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-162](tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](tablebucket-policy.md#s3t-tablebucket-163),
    [S3T-TABLEBUCKET-164](tablebucket.md#s3t-tablebucket-164), [S3T-TABLEBUCKET-127](tablebucket.md#s3t-tablebucket-127), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029) · hypotheses: H-S-069 · evidence:
    tablebucket/sub-resources/policy-get-put-delete
  - notes: H-S-069 confirmed: bare name BadRequestException 400, nonexistent NotFoundException 404, upper-case
    name BadRequestException 400, trailing slash BadRequestException 400, upper-case prefix
    BadRequestException 400 (expected 400/404/400/400/400).
  - full notes: [details/S3T-TABLEBUCKET-066.md](details/S3T-TABLEBUCKET-066.md)

## Idempotency

- <a id="s3t-tablebucket-013"></a>**S3T-TABLEBUCKET-013** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Second CreateTableBucket for an existing name -> ConflictException 409 whether or not encryption/tags match; existing bucket unchanged**
  Identical re-create (same tags) -> ConflictException HTTP 409 'The bucket that you tried to create already
  exists, and you own it.'; re-create with encryptionConfiguration=AES256 and different tags ->
  ConflictException HTTP 409 'The bucket that you tried to create already exists, and you own it.'; re-create
  with name only -> ConflictException HTTP 409 'The bucket that you tried to create already exists, and you
  own it.'. Afterwards GetTableBucket tableBucketId/createdAt, ListTagsForResource ({"a": "1", "ackq-run":
  "8b5b08"}) and GetTableBucketEncryption ({"sseAlgorithm": "AES256"}) were unchanged. Raw error body:
  {"message":"The bucket that you tried to create already exists, and you own it."}.
  - ACK: custom_create, terminal_codes · ops: CreateTableBucket · fields: name, tags, encryptionConfiguration
  - repro: CreateTableBucket(name, tags) twice; third time with different encryption/tags; compare
    Get/ListTags/GetEncryption before and after
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-159](tablebucket.md#s3t-tablebucket-159), [S3T-TABLEBUCKET-017](tablebucket.md#s3t-tablebucket-017), [S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115), [S3T-TABLEBUCKET-070](tablebucket.md#s3t-tablebucket-070),
    [S3T-TABLEBUCKET-047](tablebucket.md#s3t-tablebucket-047), [S3T-TABLEBUCKET-131](tablebucket.md#s3t-tablebucket-131), [S3T-TABLEBUCKET-160](tablebucket.md#s3t-tablebucket-160), [S3T-TABLEBUCKET-167](tablebucket.md#s3t-tablebucket-167), [S3T-TABLEBUCKET-003](tablebucket.md#s3t-tablebucket-003),
    [S3T-TABLEBUCKET-122](tablebucket.md#s3t-tablebucket-122), [S3T-TABLEBUCKET-121](tablebucket.md#s3t-tablebucket-121), [S3T-NAMESPACE-019](namespace.md#s3t-namespace-019) · hypotheses: H-B-016 · evidence:
    tablebucket/idempotency/dup-create-delete-recreate
  - notes: Confirms H-B-016. Three duplicate forms tried (identical, different config, bare). Full original
    title: Second CreateTableBucket for an existing name returns ConflictException (HTTP 409) whether or not
    the configuration matches; existing bucket unchanged

- <a id="s3t-tablebucket-049"></a>**S3T-TABLEBUCKET-049** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketEncryption is NOT idempotent - NotFoundException 404 when never set and on a second Delete; 204 only when a config exists** (hypothesis refuted; behavior confirmed)
  On one plain bucket: DeleteTableBucketEncryption never set -> NotFoundException (HTTP 404) 'The specified
  encryption configuration does not exist.'; Get right after -> '200
  {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. PutTableBucketEncryption(AES256) -> HTTP 200; Get ->
  '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. Delete -> HTTP 204; the first Get right after
  the Delete returned '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (equal to the
  never-configured body); 4 more Gets at 1 s intervals all identical. Second Delete -> NotFoundException (HTTP 404)
  'The specified encryption configuration does not exist.'.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketEncryption, PutTableBucketEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration
  - repro: Fresh bucket: Delete; Put(AES256); Get; Delete; Get immediately + 4 polls at 1 s; Delete again
  - measurements: get_converged_on_first_read=true, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-005](#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-051](#s3t-tablebucket-051), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052),
    [S3T-TABLEBUCKET-072](#s3t-tablebucket-072) · hypotheses: H-S-003 · evidence: tablebucket/sub-resources/encryption-get-put-delete
  - notes: H-S-003 REFUTED on the never-set/idempotency part - DeleteTableBucketEncryption returns
    NotFoundException 404 'The specified encryption configuration does not exist.' when nothing was ever set
    and again on a repeated Delete (Delete is NOT idempotent); the convergence part held - the first Get
    after...
  - full notes: [details/S3T-TABLEBUCKET-049.md](details/S3T-TABLEBUCKET-049.md)

## Errors

- <a id="s3t-tablebucket-035"></a>**S3T-TABLEBUCKET-035** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket with aws:kms and a nonexistent key ARN -> BadRequestException 400 (not 404); other-region key -> 400...**
  Nonexistent key in the caller account/region -> BadRequestException HTTP 400 'The specified KMS key ARN does
  not exist.', GetTableBucket on the constructed bucket ARN -> NotFoundException HTTP 404 'The specified
  bucket does not exist.'; key ARN in us-east-1 -> BadRequestException HTTP 400 'Cross region KMS keys are not
  supported.'; key ARN of another account -> ForbiddenException HTTP 403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on this
  resource because the resource does not exist in this Region, no resource-based policies allow access, or a
  resource-based policy explicitly denies access'. No bucket was created in any case.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTableBucket, GetTableBucket · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(aws:kms,
    kmsKeyArn=arn:aws:kms:<region>:<account>:key/00000000-0000-4000-8000-000000000000) and variants
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLEBUCKET-181](#s3t-tablebucket-181), [S3T-TABLEBUCKET-106](tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](table-policy.md#s3t-table-086),
    [S3T-TABLEBUCKET-185](#s3t-tablebucket-185), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107) · hypotheses: H-B-014 · evidence:
    tablebucket/sub-resources/kms-encryption-forms
  - notes: REFUTES H-B-014 (a): a well-formed nonexistent key ARN yields BadRequestException 'The specified
    KMS key ARN does not exist.', not NotFoundException; so 404 from CreateTableBucket was not reachable
    through KMS. Other-region key: BadRequestException 'Cross region KMS keys are not supported.'. Another...
  - full notes: [details/S3T-TABLEBUCKET-035.md](details/S3T-TABLEBUCKET-035.md)

- <a id="s3t-tablebucket-082"></a>**S3T-TABLEBUCKET-082** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (no 409), last writer wins; 10 maintenance Puts 204 but ~6.5 s**
  Fired concurrently on one table bucket: PutTableBucketPolicy x10 -> {"200": 7,
  "InternalServerErrorException/500": 3} (wall 6.29s); PutTableBucketMaintenanceConfiguration x10 -> {"204":
  10} (wall 6.46s); Put/DeleteTableBucketEncryption 5+5 -> {"200": 5, "NotFoundException/404": 2, "204": 3}
  (wall 0.18s); PutTableBucketStorageClass(STANDARD) x20 -> {"200": 20} (wall 0.24s). Error messages:
  {"PutTableBucketPolicy x10": {"InternalServerErrorException": "An internal error occurred. Try again."},
  "Put/DeleteTableBucketEncryption 5+5": {"NotFoundException": "The specified encryption configuration does
  not exist."}}. After the policy race GetTableBucketPolicy showed Sid P07 (which is one of the 7 successful
  puts); rejected puts retried sequentially -> [('P03', 'InternalServerErrorException', 'HTTP 200'), ('P06',
  'InternalServerErrorException', 'HTTP 200'), ('P09', 'InternalServerErrorException', 'HTTP 200')]. After the
  maintenance race unreferencedDays=6 (which is one of the 10 successful puts). After the encryption
  put/delete race GetTableBucketEncryption -> HTTP 200 {"sseAlgorithm": "AES256"}.
  - ACK: requeue, one-per-reconcile · ops: PutTableBucketPolicy, GetTableBucketPolicy,
    PutTableBucketMaintenanceConfiguration, PutTableBucketEncryption, DeleteTableBucketEncryption,
    PutTableBucketStorageClass · fields: resourcePolicy, configuration, encryptionConfiguration,
    storageClassConfiguration
  - repro: 10 threads x PutTableBucketPolicy(distinct Sid); 10 x
    PutTableBucketMaintenanceConfiguration(unreferencedDays 1..10); 5 Put + 5 Delete TableBucketEncryption; 20
    x PutTableBucketStorageClass(STANDARD); Get* after each
  - measurements: wall_s.policy=6.29, wall_s.maintenance=6.46, wall_s.encryption=0.18,
    wall_s.storageclass=0.24, n=50, latency_ms_min_max.policy=[118, 6281],
    latency_ms_min_max.maintenance=[134, 6452], latency_ms_min_max.encryption=[10, 178],
    latency_ms_min_max.storageclass=[16, 226]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-078](tablebucket-maintenance-metrics.md#s3t-tablebucket-078), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093) · hypotheses: H-S-074, H-S-073 · evidence:
    tablebucket/limits/read-write-bursts-policy-size
  - notes: H-S-074 refuted on the code: concurrent PutTableBucketPolicy produced InternalServerErrorException
    500 'An internal error occurred. Try again.' (3/10), never ConflictException; the 7 successful puts and
    the stored policy (Sid P07, one of the successes) are consistent with last-writer-wins and the 3...
  - full notes: [details/S3T-TABLEBUCKET-082.md](details/S3T-TABLEBUCKET-082.md)

- <a id="s3t-tablebucket-087"></a>**S3T-TABLEBUCKET-087** `error-code` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **PutTableBucketStorageClass enum/null violations return wire code ValidationException (HTTP 400), a shape NOT in the API model**
  Seven invalid storageClassConfiguration values ('standard', 'Standard', 'STANDARD_IA', 'GLACIER',
  'INTELLIGENT-TIERING', '' and {}) were all rejected with HTTP 400 and error code 'ValidationException',
  message "1 validation error detected: Value at 'storageClassConfiguration.storageClass' failed to satisfy
  constraint: Member must satisfy enum value set: [INTELLIGENT_TIERING, STANDARD]" (for {}: '... Member must
  not be null'). The model declares only BadRequestException for this operation; a malformed ARN on the same
  operation family returns BadRequestException 400 instead.
  - ACK: terminal_codes · ops: PutTableBucketStorageClass · fields: storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass with storageClass 'standard' (SDK parameter validation disabled) -> read
    the error code
  - measurements: n=7
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-086](#s3t-tablebucket-086), [S3T-TABLEBUCKET-026](tablebucket.md#s3t-tablebucket-026), [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLEBUCKET-024](#s3t-tablebucket-024),
    [S3T-TABLEBUCKET-104](#s3t-tablebucket-104), [S3T-TABLE-012](table-encryption-storage.md#s3t-table-012), [S3T-TABLE-013](table-encryption-storage.md#s3t-table-013), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020) · hypotheses: H-S-040, H-S-072 ·
    evidence: tablebucket/sub-resources/storage-class-default-transitions
  - notes: H-S-072 qualified: the wire code for enum validation failures is ValidationException, not one of
    the seven declared shapes; terminal-code lists keyed on BadRequestException would miss it. Seen 7/7 times
    on one bucket.
  - full notes: [details/S3T-TABLEBUCKET-087.md](details/S3T-TABLEBUCKET-087.md)

- <a id="s3t-tablebucket-109"></a>**S3T-TABLEBUCKET-109** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **s3tables wire codes: BadRequestException=400, ConflictException=409, NotFoundException=404, ValidationException=400...**
  45 failed calls in this probe produced 4 distinct tuples: code=BadRequestException http=400
  header-prefix=BadRequestException
  header='BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['get_table_bucket_policy', 'put_table_bucket_policy']; code=ConflictException http=409
  header-prefix=ConflictException
  header='ConflictException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['create_table_bucket']; code=NotFoundException http=404 header-prefix=NotFoundException
  header='NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/' body.__type=None
  ops=['delete_table_bucket_encryption', 'get_table_bucket_policy']; code=ValidationException http=400
  header-prefix=ValidationException
  header='ValidationException:http://internal.amazon.com/coral/com.amazon.coral.validate/' body.__type=None
  ops=['put_table_bucket_storage_class']. The SDK error code is the part of x-amzn-ErrorType before the first
  ':'; the JSON body carries 'message' and no '__type'.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTableBucket, DeleteTableBucketEncryption,
    GetTableBucketPolicy, PutTableBucketPolicy, PutTableBucketStorageClass
  - repro: Force a bad enum (client validation off), a malformed ARN, a missing policy, malformed policy JSON
    and a duplicate CreateTableBucket; read raw headers
  - measurements: n=45
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-067](#s3t-tablebucket-067), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-083](table-policy.md#s3t-table-083), [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLE-109](table-policy.md#s3t-table-109), [S3T-TABLE-120](table.md#s3t-table-120) ·
    hypotheses: H-S-072 · evidence: tablebucket/consistency-windows/subresource-read-after-write
  - notes: H-S-072 qualified: the status mapping held for the shapes reached (400/404/409 as modeled) and the
    code is the bare shape name after stripping the ':http://internal.amazon.com/coral/...' suffix, but
    ValidationException (400) is returned for enum/length violations and is not in the model;...
  - full notes: [details/S3T-TABLEBUCKET-109.md](details/S3T-TABLEBUCKET-109.md)

- <a id="s3t-tablebucket-181"></a>**S3T-TABLEBUCKET-181** `error-code` · impact medium · handled · verified 2026-10-14
  **PutTableBucketEncryption(aws:kms) needs caller kms:DescribeKey, CreateTable needs kms:GenerateDataKey; denied -> ForbiddenException 403**
  Assumed role, inline policy only, bucket already SSE-KMS with the same key. V1 (Allow s3tables:*; Deny
  kms:DescribeKey; no kms Allow): PutTableBucketEncryption -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' /
  ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' (12 s
  and 22 s after the policy write); CreateTable(schema) -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:GenerateDataKey on
  resource: arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a because no
  identity-base'; GetTableBucketEncryption -> HTTP 200. V2 (Allow s3tables:*, Allow kms:*, Deny
  kms:DescribeKey): Put -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ' /
  ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/b927a53d-c11d-439b-8665-831260b0bf8a with an explicit deny in an ';
  CreateTable -> HTTP 200; Get -> HTTP 200. V3 (Allow s3tables:*, Allow kms:*): Put -> HTTP 200 / HTTP 200;
  CreateTable -> HTTP 200; Get -> HTTP 200.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, CreateTable, GetTableBucketEncryption ·
    fields: encryptionConfiguration.kmsKeyArn, encryptionConfiguration.sseAlgorithm
  - repro: iam CreateRole(trust: account) + PutRolePolicy(variant) -> sts AssumeRole ->
    PutTableBucketEncryption(aws:kms, key) x2, CreateTable(schema), GetTableBucketEncryption
  - measurements: n_put_per_variant=2, iam_propagation_wait_s=12, n=2
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](table-policy.md#s3t-table-109), [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLE-131](table-policy.md#s3t-table-131), [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLEBUCKET-035](#s3t-tablebucket-035), [S3T-TABLEBUCKET-106](tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](table-policy.md#s3t-table-086)
    · evidence: tablebucket/sub-resources/fu-01-kms-key-policy-disable
  - notes: FU-01 part 2. The key policy at this point: root + maintenance.s3tables.amazonaws.com. Caller-side
    KMS permission checks happen on the caller's identity (the controller's role), so the controller's IAM
    policy needs whatever V1->V3 shows. Two attempts per variant guard against IAM propagation; see...
  - full notes: [details/S3T-TABLEBUCKET-181.md](details/S3T-TABLEBUCKET-181.md)

- <a id="s3t-tablebucket-182"></a>**S3T-TABLEBUCKET-182** `error-code` · impact high · handled · verified 2026-10-14
  **Disabled KMS key (used seconds before): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; EnableKey recovers**
  Key used by CreateTable(schema) seconds before (-> HTTP 200), then kms DisableKey (KeyState Disabled).
  Series: +3 s: Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200;
  +30 s: Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +120 s:
  Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +300 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200. Error messages
  seen: ["BadRequestException: 'Insufficient access to perform table maintenance. Ensure that the key used to
  encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.'", "BadRequestException: 'The specified KMS key ARN is
  disabled.'"]. After EnableKey, PutTableBucketEncryption succeeded after 0.04 s (polled every 2 s; codes seen
  ['OK']); CreateTable(schema) right after -> HTTP 200.
  - ACK: terminal_codes, requeue · ops: PutTableBucketEncryption, CreateTable, GetTableBucketEncryption ·
    fields: encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(schema) in an SSE-KMS bucket -> kms DisableKey -> PutTableBucketEncryption(aws:kms,
    key)/CreateTable(schema)/GetTableBucketEncryption at +3, +30, +120, +300 s -> EnableKey -> poll Put
  - measurements: offsets_s=[3, 30, 120, 300], recovery_after_enable_s=0.04, n=1
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-180](#s3t-tablebucket-180), [S3T-TABLEBUCKET-183](#s3t-tablebucket-183) · evidence:
    tablebucket/sub-resources/fu-01-kms-key-policy-disable
  - notes: FU-01 part 3 (recently used key). One series; whether the service caches key state would show as a
    success at +3 s turning into a failure later - see the series for what was observed. Messages differ per
    operation: PutTableBucketEncryption -> 'The specified KMS key ARN is disabled.' (specific),...
  - full notes: [details/S3T-TABLEBUCKET-182.md](details/S3T-TABLEBUCKET-182.md)

- <a id="s3t-tablebucket-183"></a>**S3T-TABLEBUCKET-183** `error-code` · impact medium · handled · verified 2026-10-14
  **Disabled KMS key (idle 10 min): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; same as warm-key series**
  No call touched the key or the bucket for 600 s, then kms DisableKey (KeyState Disabled). Series: +3 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +30 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +120 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +300 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200. Error messages
  seen: ["BadRequestException: 'Insufficient access to perform table maintenance. Ensure that the key used to
  encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.'", "BadRequestException: 'The specified KMS key ARN is
  disabled.'"]. After EnableKey, PutTableBucketEncryption succeeded after 0.04 s (codes seen ['OK']);
  CreateTable(schema) right after -> HTTP 200. Compared with the recently-used series (BadRequestException)
  the outcomes are identical.
  - ACK: terminal_codes, requeue · ops: PutTableBucketEncryption, CreateTable, GetTableBucketEncryption ·
    fields: encryptionConfiguration.kmsKeyArn
  - repro: idle 600 s -> kms DisableKey -> PutTableBucketEncryption(aws:kms,
    key)/CreateTable(schema)/GetTableBucketEncryption at +3, +30, +120, +300 s -> EnableKey -> poll Put
  - measurements: idle_s=600, offsets_s=[3, 30, 120, 300], recovery_after_enable_s=0.04, n=1
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-180](#s3t-tablebucket-180), [S3T-TABLEBUCKET-182](#s3t-tablebucket-182) · evidence:
    tablebucket/sub-resources/fu-01-kms-key-policy-disable
  - notes: FU-01 part 4 (idle key). One series. Messages differ per operation: PutTableBucketEncryption ->
    'The specified KMS key ARN is disabled.' (specific), while CreateTable -> 'Insufficient access to perform
    table maintenance. Ensure that the key ... still exists, is active, and has a resource policy...
  - full notes: [details/S3T-TABLEBUCKET-183.md](details/S3T-TABLEBUCKET-183.md)

## Request validation

- <a id="s3t-tablebucket-023"></a>**S3T-TABLEBUCKET-023** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket: tags={} -> 200; encryptionConfiguration={}, storageClassConfiguration={} or aws:kms without kmsKeyArn -> 400**
  tags={} -> HTTP 200 (deleted again); encryptionConfiguration={} -> BadRequestException/400 'The specified
  SSE algorithm is invalid.'; storageClassConfiguration={} -> BadRequestException/400 'The specified storage
  class is not valid.'; encryptionConfiguration={sseAlgorithm: aws:kms} (no key) -> BadRequestException/400
  "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.", GetTableBucket on its constructed
  ARN -> NotFoundException HTTP 404 'The specified bucket does not exist.', ListTableBuckets(prefix) -> 0
  items; aws:kms with kmsKeyArn='' -> BadRequestException/400 "The specified KMS key ARN is invalid. Only key
  ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.".
  - ACK: custom_create, terminal_codes · ops: CreateTableBucket · fields: tags, encryptionConfiguration,
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn, storageClassConfiguration
  - repro: CreateTableBucket with tags={} / encryptionConfiguration={} / storageClassConfiguration={} /
    {sseAlgorithm: aws:kms} (parameter_validation=False)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-026](tablebucket.md#s3t-tablebucket-026), [S3T-TABLEBUCKET-024](#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](#s3t-tablebucket-104), [S3T-TABLEBUCKET-087](#s3t-tablebucket-087),
    [S3T-TABLE-087](table-encryption-storage.md#s3t-table-087), [S3T-TABLE-013](table-encryption-storage.md#s3t-table-013), [S3T-TABLE-012](table-encryption-storage.md#s3t-table-012), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020), [S3T-TABLEBUCKET-102](#s3t-tablebucket-102), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103),
    [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085), [S3T-TABLEBUCKET-153](tablebucket.md#s3t-tablebucket-153), [S3T-TABLEBUCKET-154](tablebucket.md#s3t-tablebucket-154), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028), [S3T-TABLE-125](table.md#s3t-table-125)
    · hypotheses: H-B-032, H-B-011 · evidence: tablebucket/weird-inputs/create-names-enums-empty-structs
  - notes: H-B-032: confirmed. H-B-011: confirmed (message names kmsKeyArn: True). Full original title:
    CreateTableBucket: tags={} accepted; encryptionConfiguration={} BadRequestException/400;
    storageClassConfiguration={} BadRequestException/400; aws:kms without kmsKeyArn BadRequestException/400

- <a id="s3t-tablebucket-034"></a>**S3T-TABLEBUCKET-034** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket encryption: AES256+kmsKeyArn -> BadRequestException 400...**
  With one symmetric CMK (default key policy): {AES256, kmsKeyArn=<key ARN>} -> BadRequestException HTTP 400
  "KMS key ARN detected. Please specify 'aws:kms' as the encryption algorithm."; {aws:kms, kmsKeyArn=<bare key
  id>} -> BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn=<alias ARN>} ->
  BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn='alias/...'} ->
  BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn=<key ARN>} ->
  HTTP 200 and GetTableBucketEncryption returned {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/12aa39a2-b0c3-475c-aa11-bb7fd4033450"}.
  - ACK: terminal_codes, references, docs-only · ops: CreateTableBucket, GetTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket with each encryptionConfiguration form; GetTableBucketEncryption on any 200
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-035](#s3t-tablebucket-035), [S3T-TABLEBUCKET-181](#s3t-tablebucket-181), [S3T-TABLEBUCKET-106](tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](table-policy.md#s3t-table-086),
    [S3T-TABLEBUCKET-135](#s3t-tablebucket-135), [S3T-TABLE-258](table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](#s3t-tablebucket-138),
    [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) · hypotheses: H-B-012, H-B-013 · evidence:
    tablebucket/sub-resources/kms-encryption-forms
  - notes: H-B-012: confirmed (rejected). H-B-013: confirmed (key id rejected, alias ARN rejected, key ARN
    echo identical).
  - full notes: [details/S3T-TABLEBUCKET-034.md](details/S3T-TABLEBUCKET-034.md)

- <a id="s3t-tablebucket-102"></a>**S3T-TABLEBUCKET-102** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption: aws:kms without kmsKeyArn or with '' -> 400; AES256 with any kmsKeyArn (even '') -> 400; stored value unchanged**
  Baseline aws:kms+key stored. kms_no_key -> BadRequestException 400 ('The specified KMS key ARN is invalid.
  Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'), Get after ->
  aws:kms+key; aes_with_key -> BadRequestException 400 ('KMS key ARN detected. Please specify 'aws:kms' as the
  encryption algorithm.'), Get after -> aws:kms+key; kms_empty_key -> BadRequestException 400 ('The specified
  KMS key ARN is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is
  supported.'), Get after -> aws:kms+key; aes_empty_key -> BadRequestException 400 ('KMS key ARN detected.
  Please specify 'aws:kms' as the encryption algorithm.'), Get after -> aws:kms+key. Get after each rejected
  call still returned the baseline (True).
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: Put aws:kms+key; then Put {aws:kms}, {AES256,key}, {aws:kms,''}, {AES256,''} with SDK validation
    off; Get after each
  - measurements: n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-103](#s3t-tablebucket-103), [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) · hypotheses: H-S-004 ·
    evidence: tablebucket/weird-inputs/encryption-put-validation
  - notes: H-S-004 confirmed: aws:kms without kmsKeyArn and aws:kms with '' -> 400 'The specified KMS key ARN
    is invalid. Only key ARN format ... is supported.'; AES256 with a key ARN and AES256 with '' -> 400 "KMS
    key ARN detected. Please specify 'aws:kms' as the encryption algorithm." Note: an empty string...
  - full notes: [details/S3T-TABLEBUCKET-102.md](details/S3T-TABLEBUCKET-102.md)

- <a id="s3t-tablebucket-105"></a>**S3T-TABLEBUCKET-105** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption on a us-east-1 bucket with a us-west-2 KMS key ARN -> BadRequestException 400**
  Bucket in us-east-1, key in us-west-2: Put(aws:kms, key) -> BadRequestException (HTTP 400) 'Cross region KMS
  keys are not supported.'. Get -> AES256.
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, CreateTable · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket in the secondary region; Put aws:kms with a primary-region key ARN; if accepted
    CreateNamespace+CreateTable
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-128](#s3t-tablebucket-128), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107), [S3T-TABLEBUCKET-131](tablebucket.md#s3t-tablebucket-131), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) · hypotheses:
    H-S-006 · evidence: tablebucket/weird-inputs/encryption-put-validation
  - notes: H-S-006 confirmed: rejected at Put time with BadRequestException 400 'Cross region KMS keys are not
    supported.' (bucket us-east-1, key us-west-2); the same message is returned for a nonexistent key ARN in
    another region, i.e. the region check precedes the existence check.

- <a id="s3t-tablebucket-107"></a>**S3T-TABLEBUCKET-107** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption validates the KMS key at Put time: nonexistent, Disabled and PendingDeletion keys -> BadRequestException 400** (hypothesis refuted; behavior confirmed)
  Well-formed ARN of a key that does not exist (same region) -> BadRequestException 400 ('The specified KMS
  key ARN does not exist.'); same in us-east-1 -> BadRequestException 400 ('Cross region KMS keys are not
  supported.'). Key disabled (DisableKey -> HTTP 200): Put -> BadRequestException 400 ('The specified KMS key
  ARN is disabled.'), Get -> AES256. Key scheduled for deletion (ScheduleKeyDeletion -> HTTP 200): Put ->
  BadRequestException 400 ('The specified KMS key ARN is disabled.'), Get -> AES256.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Put with a random key UUID ARN; DisableKey -> Put; ScheduleKeyDeletion(7) -> Put; Get after each
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-105](#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](#s3t-tablebucket-128), [S3T-TABLEBUCKET-131](tablebucket.md#s3t-tablebucket-131), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085),
    [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLE-249](table-encryption-storage.md#s3t-table-249), [S3T-TABLEBUCKET-185](#s3t-tablebucket-185), [S3T-TABLEBUCKET-035](#s3t-tablebucket-035) · hypotheses: H-S-008 ·
    evidence: tablebucket/weird-inputs/encryption-put-validation
  - notes: H-S-008 REFUTED (hypothesis: no key-state validation at Put): a nonexistent key -> 400 'The
    specified KMS key ARN does not exist.'; a Disabled key -> 400 'The specified KMS key ARN is disabled.'; a
    PendingDeletion key -> 400 with the SAME 'is disabled' message (KMS reports KeyState=PendingDeletion)....
  - full notes: [details/S3T-TABLEBUCKET-107.md](details/S3T-TABLEBUCKET-107.md)

- <a id="s3t-tablebucket-185"></a>**S3T-TABLEBUCKET-185** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Disabled or PendingDeletion KMS key: PutTableBucketEncryption and CreateTableBucket(aws:kms) -> 400 'is disabled.' at +5/+30/+120 s (30/30)**
  Enabled-key control (n=3): Put aws:kms K -> ['HTTP 200'], CreateTableBucket(aws:kms K) -> ['HTTP 200'].
  DisableKey -> HTTP 200; then per offset {KMS KeyState, Put outcomes, Create outcomes}: {"5": {"at_s": 6.4,
  "create": ["BadRequestException (HTTP 400)"], "key_state": "Disabled", "put": ["BadRequestException (HTTP
  400)"]}, "30": {"at_s": 31.5, "create": ["BadRequestException (HTTP 400)"], "key_state": "Disabled", "put":
  ["BadRequestException (HTTP 400)"]}, "120": {"at_s": 121.5, "create": ["BadRequestException (HTTP 400)"],
  "key_state": "Disabled", "put": ["BadRequestException (HTTP 400)"]}}. ScheduleKeyDeletion(7 d) -> HTTP 200;
  per offset: {"5": {"at_s": 6.4, "create": ["BadRequestException (HTTP 400)"], "key_state":
  "PendingDeletion", "put": ["BadRequestException (HTTP 400)"]}, "60": {"at_s": 61.6, "create":
  ["BadRequestException (HTTP 400)"], "key_state": "PendingDeletion", "put": ["BadRequestException (HTTP
  400)"]}}. Rejection messages: ["BadRequestException (HTTP 400) 'The specified KMS key ARN is disabled.'"].
  GetTableBucketEncryption after a rejected Put kept the AES256 reset value: True.
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, CreateTableBucket,
    GetTableBucketEncryption · fields: encryptionConfiguration.kmsKeyArn
  - repro: CreateKey K -> bucket B (AES256) -> DisableKey -> at +5/+30/+120 s: 3 x [Put AES256; Put aws:kms K;
    Get] and 3 x CreateTableBucket(aws:kms K) -> ScheduleKeyDeletion -> +5/+60 s same
  - measurements: disabled_offsets_s=[5, 30, 120], pending_offsets_s=[5, 60], trials_per_offset=3,
    put_cells=15, create_cells=15, n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107), [S3T-TABLEBUCKET-035](#s3t-tablebucket-035) · evidence:
    tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny
  - notes: FU2-01: 107 holds. At +5 (actual 6.4), +30 and +120 s after DisableKey and at +5 and +60 s after
    ScheduleKeyDeletion(7 d), every PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K)
    was rejected with BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put...
  - full notes: [details/S3T-TABLEBUCKET-185.md](details/S3T-TABLEBUCKET-185.md)

## Update granularity and ordering

- <a id="s3t-tablebucket-119"></a>**S3T-TABLEBUCKET-119** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTableBucket(aws:kms) 200, then CreateTable BadRequestException**
  Customer key whose policy grants only the account root (no maintenance.s3tables.amazonaws.com statement).
  PutTableBucketEncryption(aws:kms, key) -> HTTP 200; GetTableBucketEncryption -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}. CreateTableBucket
  with the same encryptionConfiguration -> HTTP 200; its GetTableBucketEncryption -> {'sseAlgorithm':
  'aws:kms', 'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}
  (byte-identical to the post-create Put body). CreateTable(format=ICEBERG) in the Put bucket ->
  BadRequestException (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used
  to encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazona'.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, CreateTableBucket,
    GetTableBucketEncryption, CreateTable · fields: encryptionConfiguration.kmsKeyArn,
    encryptionConfiguration.sseAlgorithm
  - repro: CreateKey(policy: root only) -> PutTableBucketEncryption(aws:kms) -> Get ->
    CreateTableBucket(encryptionConfiguration aws:kms) -> Get -> CreateTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-120](#s3t-tablebucket-120), [S3T-TABLEBUCKET-180](#s3t-tablebucket-180), [S3T-TABLEBUCKET-139](#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](#s3t-tablebucket-142),
    [S3T-TABLE-256](table-maintenance.md#s3t-table-256), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLE-248](table-encryption-storage.md#s3t-table-248),
    [S3T-TABLE-246](table-maintenance.md#s3t-table-246), [S3T-TABLE-301](table-maintenance.md#s3t-table-301), [S3T-TABLE-254](table-maintenance.md#s3t-table-254) · hypotheses: H-S-007, H-S-083, H-S-012 · evidence:
    tablebucket/dependencies/kms-key-policy-gate
  - notes: H-S-083 (the Put itself validates the key policy): refuted. H-S-007 (Put succeeds, only CreateTable
    fails with 403 naming the maintenance principal/KMS): qualified: Put HTTP 200, CreateTable
    BadRequestException/400. H-S-012 (create-time and Put bodies identical; create-time KMS unvalidated):...
  - full notes: [details/S3T-TABLEBUCKET-119.md](details/S3T-TABLEBUCKET-119.md)

- <a id="s3t-tablebucket-139"></a>**S3T-TABLEBUCKET-139** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C007 FALSE: CreateTableBucket{aws:kms, key WITHOUT maintenance grant} -> HTTP 200; the check fires at CreateTable (400)**
  CreateTableBucket(aws:kms, CMK with the default key policy only - no statement for
  maintenance.s3tables.amazonaws.com) -> HTTP 200; GetTableBucketEncryption -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': <that key>} ([S3T-TABLEBUCKET-036](#s3t-tablebucket-036)/119). The grant is checked when a table is created:
  CreateTable(format=ICEBERG, no encryptionConfiguration) in such a bucket -> BadRequestException (HTTP 400)
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.' and GetTable afterwards -> NotFoundException 404 ([S3T-TABLE-055](table-maintenance.md#s3t-table-055)).
  - ACK: requeue, docs-only · ops: CreateTableBucket, GetTableBucketEncryption, CreateTable · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-036](#s3t-tablebucket-036), [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-054](table-maintenance.md#s3t-table-054),
    [S3T-TABLE-055](table-maintenance.md#s3t-table-055); no new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-036](#s3t-tablebucket-036), [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-142](#s3t-tablebucket-142),
    [S3T-TABLE-256](table-maintenance.md#s3t-table-256), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134) · evidence: service/static/doc-claims-1
  - notes: VERDICT: FALSE - the doc-fetch alternative is what happens: bucket create accepts an ungranted key;
    the first CreateTable in the bucket fails with 400.

- <a id="s3t-tablebucket-142"></a>**S3T-TABLEBUCKET-142** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C010 FALSE: PutTableBucketEncryption{aws:kms, key WITHOUT maintenance grant} -> HTTP 200; only the next CreateTable fails (400)**
  PutTableBucketEncryption(aws:kms, key whose policy grants only the account root) -> HTTP 200;
  GetTableBucketEncryption -> {'sseAlgorithm': 'aws:kms', 'kmsKeyArn': <that key>}.
  CreateTable(format=ICEBERG) in the bucket -> BadRequestException (HTTP 400) 'Insufficient access to perform
  table maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.'
  ([S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-055](table-maintenance.md#s3t-table-055)). The Put DOES validate key existence/state (nonexistent, Disabled,
  PendingDeletion keys -> BadRequestException 400, [S3T-TABLEBUCKET-107](#s3t-tablebucket-107)) but not the key policy.
  - ACK: requeue, docs-only · ops: PutTableBucketEncryption, CreateTable · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107); no new
    API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107), [S3T-TABLEBUCKET-139](#s3t-tablebucket-139), [S3T-TABLE-256](table-maintenance.md#s3t-table-256),
    [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLE-054](table-maintenance.md#s3t-table-054) · evidence: service/static/doc-claims-1
  - notes: VERDICT: FALSE - the recorded alternative holds: the Put is accepted and the failure surfaces at
    the first CreateTable in the bucket.

- <a id="s3t-tablebucket-180"></a>**S3T-TABLEBUCKET-180** `prerequisite` · impact high · handled · verified 2026-10-14
  **SSE-KMS bucket with a default-policy key: CreateTableBucket 200 but CreateTable 400 until the key policy grants the maintenance principal**
  KMS CreateKey without a Policy (default key policy: only the account root statement) then
  CreateTableBucket(encryptionConfiguration={aws:kms, that key}) -> HTTP 200; GetTableBucketEncryption -> HTTP 200.
  In that bucket CreateTable(format=ICEBERG, no metadata) -> BadRequestException (HTTP 400) 'Insufficient
  access to perform table maintenance. Ensure that the key used to encrypt the table still exists, is active,
  and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.'; CreateTable(metadata.iceberg.schema 2 fields) -> BadRequestException
  (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table
  still exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.'. After PutKeyPolicy adding Principal Service
  maintenance.s3tables.amazonaws.com with kms:GenerateDataKey+kms:Decrypt (Condition aws:SourceAccount) (+5
  s): CreateTable(no metadata) -> HTTP 200; CreateTable(schema) -> HTTP 200.
  - ACK: terminal_codes, docs-only, references · ops: CreateTableBucket, CreateTable, GetTableBucketEncryption
    · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn, metadata.iceberg.schema
  - repro: kms CreateKey (no Policy) -> CreateTableBucket(aws:kms, key) -> CreateNamespace -> CreateTable x2
    -> PutKeyPolicy(+maintenance principal) -> CreateTable x2
  - measurements: n_create_table_before=2, n_create_table_after=2, n=2
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-036](#s3t-tablebucket-036), [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLEBUCKET-120](#s3t-tablebucket-120), [S3T-TABLEBUCKET-182](#s3t-tablebucket-182),
    [S3T-TABLEBUCKET-183](#s3t-tablebucket-183) · evidence: tablebucket/sub-resources/fu-01-kms-key-policy-disable
  - notes: FU-01 part 1. The caller is an account admin, so caller-side KMS permissions are not the limiting
    factor here; the only difference between the two halves is the maintenance.s3tables.amazonaws.com
    statement in the key policy. The grant changed the CreateTable outcome. Both CreateTable variants (with...
  - full notes: [details/S3T-TABLEBUCKET-180.md](details/S3T-TABLEBUCKET-180.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-tablebucket-001"></a>**S3T-TABLEBUCKET-001** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket/table: the 6 encryption/maintenance/storage-class/metadata Gets return 200 server defaults, none 404 (metrics/policy not read)**
  On a table bucket and an ICEBERG table created without encryption, storage-class or maintenance settings:
  GetTableBucketEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "AES256"}};
  GetTableBucketMaintenanceConfiguration -> HTTP 200 {"configuration": {"icebergUnreferencedFileRemoval":
  {"status": "enabled", "settings": {"icebergUnreferencedFileRemoval": {"unreferencedDays": 3,
  "nonCurrentDays": 10}}}}}; GetTableEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm":
  "AES256"}}; GetTableStorageClass -> HTTP 200 {"storageClassConfiguration": {"storageClass": "STANDARD"}};
  GetTableMetadataLocation -> HTTP 200 {"versionToken": "<token>", "warehouseLocation": "<s3-uri>"};
  GetTableMaintenanceConfiguration -> HTTP 200 {"configuration": {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}},
  "icebergSnapshotManagement": {"status": "enabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}}. No sub-resource read returned NotFoundException
  for an unset setting.
  - ACK: late_initialize, hook:sdk_read_one_post_set_output · ops: GetTableBucketEncryption,
    GetTableBucketMaintenanceConfiguration, GetTableEncryption, GetTableStorageClass,
    GetTableMetadataLocation, GetTableMaintenanceConfiguration · fields: encryptionConfiguration.sseAlgorithm,
    storageClassConfiguration.storageClass, configuration, metadataLocation
  - repro: CreateTableBucket(name only) -> CreateNamespace -> CreateTable(format=ICEBERG) -> each Get*
    sub-resource read
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-053](tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-173](tablebucket-maintenance-metrics.md#s3t-tablebucket-173), [S3T-TABLEBUCKET-171](tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-056](tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059) ·
    evidence: tablebucket/round-trip/smoke
  - notes: Through the API a controller cannot tell 'never set by the user' from 'set to the default' for
    these settings; a nil spec field compared against the Get* value would show a permanent delta unless
    late-initialized or ignored. GetTableMetadataLocation on a fresh table has...
  - full notes: [details/S3T-TABLEBUCKET-001.md](details/S3T-TABLEBUCKET-001.md)

- <a id="s3t-tablebucket-005"></a>**S3T-TABLEBUCKET-005** `server-default` · impact high · handled · verified 2026-10-13
  **GetTableBucketEncryption on a bucket created without encryptionConfiguration -> HTTP 200 {"sseAlgorithm": "AES256"}, no kmsKeyArn member**
  Bucket created with name+tags only: GetTableBucketEncryption at 0 s -> HTTP 200, raw JSON
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}; at +3 s -> HTTP 200, raw JSON
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}. After PutTableBucketEncryption(AES256) (-> HTTP 200)
  the read returned {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}; DeleteTableBucketEncryption ->
  HTTP 204 (raw status 204, body ''); GetTableBucketEncryption right after the delete -> HTTP 200
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}, at +2 s -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "AES256"}}; a second DeleteTableBucketEncryption -> NotFoundException (HTTP 404).
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare, exceptions.404 · ops:
    GetTableBucketEncryption, PutTableBucketEncryption, DeleteTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(name) -> GetTableBucketEncryption; PutTableBucketEncryption(AES256) -> Get;
    DeleteTableBucketEncryption -> Get (0 s, +2 s)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53`
  - related: [S3T-TABLEBUCKET-048](#s3t-tablebucket-048), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](#s3t-tablebucket-051),
    [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLEBUCKET-072](#s3t-tablebucket-072), [S3T-TABLE-010](table-encryption-storage.md#s3t-table-010), [S3T-TABLE-011](table-encryption-storage.md#s3t-table-011), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083) · hypotheses:
    H-B-006, H-B-007 · evidence: tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-006 (default materialized, kmsKeyArn member absent) and refutes H-B-007 (no
    NotFoundException for the unset default).
  - full notes: [details/S3T-TABLEBUCKET-005.md](details/S3T-TABLEBUCKET-005.md)

- <a id="s3t-tablebucket-008"></a>**S3T-TABLEBUCKET-008** `read-gap` · impact high · handled · verified 2026-10-15, re-verified
  **encryptionConfiguration, storageClassConfiguration and tags set at create are absent from GetTableBucket/ListTableBuckets (own Get each)**
  _Behavior summary; the full recorded observation is in the lab finding._
  CreateTableBucket(AES256, INTELLIGENT_TIERING, 3 tags): the GetTableBucket body and ListTableBuckets entry
  both contain exactly the keys [arn, createdAt, name, ownerAccountId, tableBucketId, type], with no
  encryptionConfiguration, storageClassConfiguration or tags. Immediately after create,
  GetTableBucketEncryption returns HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "AES256"}},
  GetTableBucketStorageClass returns HTTP 200 {"storageClassConfiguration": {"storageClass":
  "INTELLIGENT_TIERING"}}, and ListTagsForResource returns HTTP 200 with exactly the 3 tags sent, including
  their keys and values. All three read-backs match the sent values with no missing, changed or extra fields;
  all reads are identical at +5 s.
  - ACK: hook:sdk_read_one_post_set_output, custom_update, tags.custom-sync · ops: CreateTableBucket,
    GetTableBucket, ListTableBuckets, GetTableBucketEncryption, GetTableBucketStorageClass,
    ListTagsForResource · fields: encryptionConfiguration, storageClassConfiguration, tags
  - repro: CreateTableBucket(name, encryptionConfiguration, storageClassConfiguration, tags) ->
    GetTableBucket, ListTableBuckets(prefix), GetTableBucketEncryption, GetTableBucketStorageClass,
    ListTagsForResource
  - handling: handled via `templates/hooks/table_bucket/sdk_read_one_post_set_output.go.tpl:6-12; pkg/resource/table_bucket/hooks.go:51-171; pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; pkg/resource/table_bucket/hooks.go:107-121; pkg/resource/table_bucket/hooks.go:409-471`
  - related: [S3T-TABLEBUCKET-010](#s3t-tablebucket-010), [S3T-TABLE-308](table-encryption-storage.md#s3t-table-308), [S3T-TABLEBUCKET-123](#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073),
    [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLE-061](table-encryption-storage.md#s3t-table-061), [S3T-TABLEBUCKET-146](#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](#s3t-tablebucket-147),
    [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-156](tablebucket.md#s3t-tablebucket-156), [S3T-TABLEBUCKET-157](tablebucket.md#s3t-tablebucket-157),
    [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-127](table.md#s3t-table-127), [S3T-TABLEBUCKET-028](tablebucket.md#s3t-tablebucket-028), [S3T-TABLEBUCKET-121](tablebucket.md#s3t-tablebucket-121), [S3T-TABLEBUCKET-007](tablebucket.md#s3t-tablebucket-007) · hypotheses:
    H-B-005 · evidence: tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-005: ReadOne must fan out to three extra read calls; all three returned the values
    exactly as sent, immediately. Full original title: encryptionConfiguration, storageClassConfiguration and
    tags set at create are absent from GetTableBucket/ListTableBuckets; each needs its own read

- <a id="s3t-tablebucket-083"></a>**S3T-TABLEBUCKET-083** `server-default` · impact high · handled · verified 2026-10-13
  **GetTableBucketStorageClass on a fresh bucket -> 200 {"storageClass": "STANDARD"}; after an explicit Put(STANDARD) the body is byte-identical**
  CreateTableBucket(name, tags only) -> GetTableBucketStorageClass at +0 s: '200
  {"storageClassConfiguration":{"storageClass":"STANDARD"}}' and at +5 s: '200
  {"storageClassConfiguration":{"storageClass":"STANDARD"}}' (identical). PutTableBucketStorageClass(STANDARD)
  -> HTTP 200; Get -> '200 {"storageClassConfiguration":{"storageClass":"STANDARD"}}' (identical to the fresh
  body: no isDefault/source marker).
  - ACK: late_initialize, compare.nil_equals_zero_value · ops: GetTableBucketStorageClass,
    PutTableBucketStorageClass, CreateTableBucket · fields: storageClassConfiguration.storageClass
  - repro: CreateTableBucket -> GetTableBucketStorageClass (+0 s, +5 s, raw body) -> Put(STANDARD) -> Get (raw
    body)
  - measurements: n=2
  - handling: handled via `pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-006](#s3t-tablebucket-006), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103),
    [S3T-TABLE-059](table-encryption-storage.md#s3t-table-059), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-085](#s3t-tablebucket-085), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161),
    [S3T-TABLE-010](table-encryption-storage.md#s3t-table-010), [S3T-TABLE-011](table-encryption-storage.md#s3t-table-011), [S3T-TABLEBUCKET-005](#s3t-tablebucket-005) · hypotheses: H-S-036, H-S-037, H-S-137 · evidence:
    tablebucket/sub-resources/storage-class-default-transitions
  - notes: H-S-036 (200 STANDARD materialized) confirmed; H-S-037 (contrarian 404 until first Put) refuted;
    H-S-137 (explicit STANDARD indistinguishable from implicit) confirmed.
  - full notes: [details/S3T-TABLEBUCKET-083.md](details/S3T-TABLEBUCKET-083.md)

- <a id="s3t-tablebucket-085"></a>**S3T-TABLEBUCKET-085** `shape-mismatch` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketStorageClass returns the BUCKET default (STANDARD) even with an INTELLIGENT_TIERING table; table ARN as tableBucketARN -> 400**
  CreateTable with storageClassConfiguration INTELLIGENT_TIERING in a STANDARD-default bucket -> HTTP 200,
  GetTableStorageClass -> INTELLIGENT_TIERING. GetTableBucketStorageClass(bucket ARN) -> STANDARD.
  GetTableBucketStorageClass(tableBucketARN=<table ARN>) -> BadRequestException (HTTP 400) 'The specified
  bucket name isn't valid. Specify a different bucket name, and then try again.'.
  - ACK: docs-only, is_arn_primary_key · ops: GetTableBucketStorageClass, CreateTable, GetTableStorageClass ·
    fields: tableBucketARN, storageClassConfiguration.storageClass
  - repro: CreateTable(storageClass=INTELLIGENT_TIERING) in a STANDARD bucket;
    GetTableBucketStorageClass(bucket ARN); GetTableBucketStorageClass(table ARN)
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161), [S3T-TABLE-062](table-encryption-storage.md#s3t-table-062) ·
    hypotheses: H-S-039 · evidence: tablebucket/sub-resources/storage-class-default-transitions
  - notes: H-S-039 confirmed: the model doc ('storage class configuration for a specific table') is a
    copy-paste; the op returns the bucket default only.
  - full notes: [details/S3T-TABLEBUCKET-085.md](details/S3T-TABLEBUCKET-085.md)

- <a id="s3t-tablebucket-103"></a>**S3T-TABLEBUCKET-103** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption kmsKeyArn: bare key id, alias name, alias ARN, upper-case region/prefix -> 400; key ARN echoed byte-identical**
  key_id -> BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); alias_name -> BadRequestException
  400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); alias_arn -> BadRequestException
  400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); upper_region ->
  BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); upper_prefix ->
  BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'). The full key ARN -> HTTP 200 and
  Get returned the kmsKeyArn byte-identical to the sent string.
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Put aws:kms with kmsKeyArn = key id / alias name / alias ARN / ARN with upper-case region / ARN
    with upper-case prefix / key ARN; Get after each
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLE-059](table-encryption-storage.md#s3t-table-059),
    [S3T-TABLEBUCKET-102](#s3t-tablebucket-102), [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) · hypotheses: H-S-005 · evidence:
    tablebucket/weird-inputs/encryption-put-validation
  - notes: H-S-005 confirmed: bare key id, alias name, alias ARN, ARN with upper-cased region and ARN with
    upper-cased 'ARN:AWS:KMS' prefix are all rejected with BadRequestException 400 "The specified KMS key ARN
    is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is...
  - full notes: [details/S3T-TABLEBUCKET-103.md](details/S3T-TABLEBUCKET-103.md)

## Response fidelity and consistency

- <a id="s3t-tablebucket-010"></a>**S3T-TABLEBUCKET-010** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **GetTableBucket/ListTableBuckets createdAt and tableBucketId byte-identical across 4 reads over 2 min and tag/encryption/storage-class Puts**
  Raw createdAt strings seen: ['2026-10-13T22:19:39.390242430Z'] (format sample
  '2026-10-13T22:19:39.390242430Z', 9 fractional digits); tableBucketId values seen: 1 distinct (UUID v4
  lowercase: True); ownerAccountId is the 12-digit caller account: True; type=customer. Reads at t0, right
  after TagResource+PutTableBucketEncryption+PutTableBucketStorageClass (-> ['HTTP 200', 'HTTP 200', 'HTTP
  200']), +30 s and +120 s; Get and List agreed on every read.
  - ACK: none · ops: GetTableBucket, ListTableBuckets, TagResource, PutTableBucketEncryption,
    PutTableBucketStorageClass · fields: createdAt, tableBucketId, ownerAccountId, type
  - repro: CreateTableBucket -> GetTableBucket+ListTableBuckets(prefix) at 0 s, after 3 mutations, +30 s, +120
    s; compare raw JSON strings
  - measurements: reads=4, span_s=120.4, n=4
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; templates/hooks/table_bucket/sdk_read_one_post_set_output.go.tpl:6-12; pkg/resource/table_bucket/hooks.go:51-171`
  - related: [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLE-308](table-encryption-storage.md#s3t-table-308) · hypotheses: H-B-065 · evidence:
    tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-065 (one bucket, 4 read pairs). Full original title: GetTableBucket/ListTableBuckets
    createdAt and tableBucketId are byte-identical across 4 reads over 2 min and tag/encryption/storage-class
    mutations

- <a id="s3t-tablebucket-050"></a>**S3T-TABLEBUCKET-050** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **GetTableBucketEncryption echoes sseAlgorithm 'aws:kms' and the kmsKeyArn byte-identical; after Put(AES256) kmsKeyArn is absent from the JSON**
  PutTableBucketEncryption(aws:kms, <key ARN>) -> HTTP 200; raw Get body '200
  {"encryptionConfiguration":{"kmsKeyArn":"<kms-key-arn>","sseAlgorithm":"aws:kms"}}' (kmsKeyArn equals the
  sent ARN). PutTableBucketEncryption(AES256) on the same bucket -> HTTP 200; raw Get body '200
  {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (kmsKeyArn member absent, not null and not empty).
  - ACK: compare.nil_equals_zero_value, late_initialize · ops: PutTableBucketEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: Put(aws:kms,key) -> Get (raw HTTP body) -> Put(AES256) -> Get (raw HTTP body)
  - measurements: n=1
  - handling: handled via `pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103), [S3T-TABLE-059](table-encryption-storage.md#s3t-table-059),
    [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143),
    [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) · hypotheses: H-S-010 · evidence:
    tablebucket/sub-resources/encryption-get-put-delete
  - notes: H-S-010 confirmed. Raw JSON captured from the HTTP response, not the SDK struct.

- <a id="s3t-tablebucket-059"></a>**S3T-TABLEBUCKET-059** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance status and storage class: 10 alternating Puts each, Get at +0/+0.3/+1 s -> 0 / 0 stale reads** (hypothesis refuted; behavior confirmed)
  PutTableBucketMaintenanceConfiguration alternating status disabled/enabled 10 times, three Gets after each
  Put (immediately, +0.3 s, +1 s): 0 of 30 reads did not match the just-written status.
  PutTableBucketStorageClass alternating INTELLIGENT_TIERING/STANDARD 10 times with the same read schedule: 0
  of 30 reads stale.
  - ACK: none · ops: PutTableBucketMaintenanceConfiguration, GetTableBucketMaintenanceConfiguration,
    PutTableBucketStorageClass, GetTableBucketStorageClass · fields: value.status,
    storageClassConfiguration.storageClass
  - repro: Alternate the value 10x per noun; after each Put read at +0, +0.3 and +1 s; compare to the written
    value
  - measurements: maint_get_latency_ms_min_max=[23, 329], sc_get_latency_ms_min_max=[7, 15], n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-056](tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-058](#s3t-tablebucket-058), [S3T-TABLEBUCKET-112](#s3t-tablebucket-112), [S3T-TABLEBUCKET-146](#s3t-tablebucket-146),
    [S3T-TABLEBUCKET-147](#s3t-tablebucket-147), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008),
    [S3T-TABLEBUCKET-173](tablebucket-maintenance-metrics.md#s3t-tablebucket-173), [S3T-TABLEBUCKET-171](tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-053](tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-001](#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](tablebucket-maintenance-metrics.md#s3t-tablebucket-057) ·
    hypotheses: H-S-089 · evidence: tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-089 (stale reads for several seconds) refuted in 30 reads per noun from one client in one
    region; a stale read would also have to beat the Put's own latency, which was not seen.

- <a id="s3t-tablebucket-073"></a>**S3T-TABLEBUCKET-073** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Tables created 0.49-6.68 s after Put(aws:kms)/Delete of the bucket encryption all inherited the new default (n=5 per direction)**
  After Put(aws:kms): GetTableBucketEncryption right after -> aws:kms+key; 5 tables created at [0.53, 2.08,
  3.58, 5.15, 6.68] s got ['aws:kms+key']. After Delete: bucket Get -> AES256; 5 tables created at [0.49,
  1.88, 3.3, 4.7, 6.17] s got ['AES256']. Lagging tables (previous default): put direction 0, delete direction 0.
  - ACK: requeue, none · ops: PutTableBucketEncryption, DeleteTableBucketEncryption, CreateTable,
    GetTableEncryption · fields: encryptionConfiguration
  - repro: Put(aws:kms) -> CreateTable immediately and 1/s for 5 s -> GetTableEncryption each; same after
    Delete
  - measurements: first_table_created_after_mutation_s=[0.53, 0.49], lagging_tables=0, n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-123](#s3t-tablebucket-123), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037),
    [S3T-TABLE-061](table-encryption-storage.md#s3t-table-061), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020), [S3T-TABLE-266](table-encryption-storage.md#s3t-table-266), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-072](#s3t-tablebucket-072) ·
    hypotheses: H-S-134 · evidence: tablebucket/dependencies/encryption-tables-inherit
  - notes: H-S-134 (contrarian: new tables lag the configuration read) refuted with 5 tables per direction,
    the first created ~0.49 s after the mutation returned.

- <a id="s3t-tablebucket-108"></a>**S3T-TABLEBUCKET-108** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Read-after-write of bucket storage class / policy / encryption: 20/20, 20/20 (+20/20 delete), 20/20 immediate Gets reflected the mutation**
  One bucket, 20 alternating trials per pair, one Get issued immediately after each mutation (0.5 s pause
  between trials): PutTableBucketStorageClass(INTELLIGENT_TIERING/STANDARD) -> Get consistent 20/20;
  PutTableBucketPolicy(SidA/SidB) -> Get consistent 20/20, DeleteTableBucketPolicy -> Get 404 20/20;
  PutTableBucketEncryption(AES256) -> Get AES256 20/20, DeleteTableBucketEncryption 204 20/20 with an
  immediate second Delete returning NotFoundException 20/20. Mismatches: none. Throttles: 0. Put latency 13-53
  ms (storage class), 93-135 ms (policy).
  - ACK: none, requeue · ops: PutTableBucketStorageClass, GetTableBucketStorageClass, PutTableBucketPolicy,
    GetTableBucketPolicy, DeleteTableBucketPolicy, PutTableBucketEncryption, GetTableBucketEncryption,
    DeleteTableBucketEncryption · fields: storageClassConfiguration.storageClass, resourcePolicy,
    encryptionConfiguration
  - repro: For each pair alternate two values 20 times: Put v; Get immediately; (Delete; Get immediately)
  - measurements: storage_class_consistent=20, policy_put_consistent=20, policy_delete_consistent=20,
    encryption_put_consistent=20, encryption_second_delete_404=20, n=60, put_latency_ms_storage_class=[13,
    53], put_latency_ms_policy=[93, 135], get_latency_ms_policy=[29, 68]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-003](tablebucket-policy.md#s3t-tablebucketpolicy-003), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLEBUCKETPOLICY-005](tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKET-064](tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKETPOLICY-010](tablebucket-policy.md#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011),
    [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLE-156](table-policy.md#s3t-table-156), [S3T-TABLE-149](table-policy.md#s3t-table-149) · hypotheses: H-S-075 · evidence:
    tablebucket/consistency-windows/subresource-read-after-write
  - notes: H-S-075 confirmed for the three pairs of shard D1 (maintenance, metrics and replication pairs not
    tested here).

- <a id="s3t-tablebucket-112"></a>**S3T-TABLEBUCKET-112** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Sub-resource Puts sent <1 ms after CreateTableBucket returned: 48/48 first attempts succeeded on 12 fresh buckets (metrics/policy/maint/SC)** (hypothesis refuted; behavior confirmed)
  12 x CreateTableBucket (latency 1435-1940 ms) each immediately followed by
  PutTableBucketMetricsConfiguration, PutTableBucketPolicy, PutTableBucketMaintenanceConfiguration and
  PutTableBucketStorageClass (first Put sent 0-0 ms after the create response): first-attempt codes
  {'metrics': ['HTTP 204'], 'policy': ['HTTP 200'], 'maintenance': ['HTTP 204'], 'storage_class': ['HTTP
  200']}. The Get of each noun right after the Puts reflected the written value in all 12 buckets.
  - ACK: none · ops: CreateTableBucket, PutTableBucketMetricsConfiguration, PutTableBucketPolicy,
    PutTableBucketMaintenanceConfiguration, PutTableBucketStorageClass · fields: tableBucketARN
  - repro: loop: CreateTableBucket -> four sub-resource Puts with no delay -> four Gets -> DeleteTableBucket
  - measurements: first_put_offset_ms_min_max=[0, 0], put_success_at_ms_from_create.metrics=[70, 128],
    put_success_at_ms_from_create.policy=[175, 248], put_success_at_ms_from_create.maintenance=[550, 642],
    put_success_at_ms_from_create.storage_class=[564, 655], create_latency_ms_min_max=[1435, 1940], n=12
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-056](tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-058](#s3t-tablebucket-058) · hypotheses: H-S-094 · evidence:
    tablebucket/consistency-windows/create-then-subresource-put
  - notes: H-S-094 (contrarian: some fraction of sub-resource Puts within ~100 ms of create fail with 404)
    refuted in 12 buckets x 4 nouns (0 failures, 0 retries); the Get of each noun immediately after the Puts
    reflected the written value on every bucket. Sequential buckets, one at a time, single client,...
  - full notes: [details/S3T-TABLEBUCKET-112.md](details/S3T-TABLEBUCKET-112.md)

- <a id="s3t-tablebucket-123"></a>**S3T-TABLEBUCKET-123** `eventual-consistency` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **TagResource at +0 ms and PutTableBucketEncryption at +69-142 ms after CreateTableBucket -> 5/5 first attempts 200, no ConflictException** (hypothesis refuted; behavior confirmed)
  5 fresh buckets created without the probe tag (CreateTableBucket latency 1561-1823 ms); TagResource({t:1})
  was sent [0, 0, 0, 0, 0] ms after the create response and PutTableBucketEncryption(AES256) right after it
  ([93, 69, 142, 74, 76] ms): first-attempt outcomes ['HTTP 200'] and ['HTTP 200'] (messages none).
  ListTagsForResource right after showed the tag in 5/5 and GetTableBucketEncryption showed AES256 in 5/5.
  - ACK: none, tags.custom-sync · ops: CreateTableBucket, TagResource, PutTableBucketEncryption,
    ListTagsForResource, GetTableBucketEncryption · fields: tags, encryptionConfiguration.sseAlgorithm
  - repro: CreateTableBucket(no tags) -> immediately TagResource({t:1}) -> immediately
    PutTableBucketEncryption(AES256) -> ListTags + GetEncryption; x5
  - measurements: tag_offset_ms=[0, 0, 0, 0, 0], encryption_offset_ms=[93, 69, 142, 74, 76],
    tag_first_attempt_ok=5, encryption_first_attempt_ok=5, create_latency_ms=[1562, 1600, 1646, 1823, 1561],
    n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-112](#s3t-tablebucket-112), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009),
    [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLE-061](table-encryption-storage.md#s3t-table-061) · hypotheses: H-B-018 · evidence:
    tablebucket/consistency-windows/create-read-paths-tags-n5
  - notes: H-B-018 (409 'conflict with a previous write' window under 2 s for writes right after Create)
    refuted for TagResource and PutTableBucketEncryption (0 retries needed); extends [S3T-TABLEBUCKET-112](#s3t-tablebucket-112)
    (metrics/policy/maintenance/storage class) to the two remaining write paths. ACK's...
  - full notes: [details/S3T-TABLEBUCKET-123.md](details/S3T-TABLEBUCKET-123.md)

## Sub-resources

- <a id="s3t-tablebucket-009"></a>**S3T-TABLEBUCKET-009** `sub-resource-api` · impact medium · handled · verified 2026-10-13
  **storageClassConfiguration is mutable via PutTableBucketStorageClass (HTTP 200, visible after 0.01 s); no Delete op exists to unset it**
  Bucket created with storageClass=INTELLIGENT_TIERING read back {"storageClass": "INTELLIGENT_TIERING"}
  immediately. PutTableBucketStorageClass(STANDARD) -> HTTP 200 (raw body ''); GetTableBucketStorageClass
  showed STANDARD after 0.01 s (1 s polling, first poll immediately). Re-sending STANDARD unchanged -> HTTP 200.
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200, visible after 0.01 s. The botocore model
  (apiVersion 2018-05-10) has no DeleteTableBucketStorageClass (present=False) and no Update*/Rename* bucket
  operation ([]).
  - ACK: custom_update, is_immutable · ops: CreateTableBucket, PutTableBucketStorageClass,
    GetTableBucketStorageClass · fields: storageClassConfiguration.storageClass, name
  - repro: CreateTableBucket(storageClass=INTELLIGENT_TIERING) -> PutTableBucketStorageClass(STANDARD) -> poll
    Get -> Put(INTELLIGENT_TIERING) again
  - measurements: put_to_visible_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLEBUCKET-085](#s3t-tablebucket-085), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161),
    [S3T-TABLEBUCKET-123](#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLE-061](table-encryption-storage.md#s3t-table-061),
    [S3T-TABLEBUCKET-146](#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](#s3t-tablebucket-147), [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059) · hypotheses: H-B-009,
    H-B-029, H-B-071 · evidence: tablebucket/round-trip/create-identity-defaults
  - notes: Confirms H-B-009 (not create-only; both directions accepted; the way back to the default is Put
    STANDARD). Confirms the static H-B-029 (no UpdateTableBucket/rename in the model: name, type,
    ownerAccountId, tableBucketId have no mutation path). Partially covers H-B-071:
    PutTableBucketStorageClass...
  - full notes: [details/S3T-TABLEBUCKET-009.md](details/S3T-TABLEBUCKET-009.md)

- <a id="s3t-tablebucket-037"></a>**S3T-TABLEBUCKET-037** `sub-resource-api` · impact high · handled · verified 2026-10-13
  **PutTableBucketEncryption switches AES256<->aws:kms both ways (visible 0.01 s); unchanged re-send 200; Delete on aws:kms -> 204, reads AES256**
  Bucket created with AES256: Put AES256 (unchanged) -> HTTP 200; Put aws:kms -> HTTP 200; Put aws:kms
  (unchanged) -> HTTP 200; Put aws:kms without kmsKeyArn -> BadRequestException HTTP 400 "The specified KMS
  key ARN is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is
  supported."; Put AES256 (from aws:kms) -> HTTP 200; Put aws:kms again -> HTTP 200;
  DeleteTableBucketEncryption on aws:kms bucket -> HTTP 204; Get after delete -> {"sseAlgorithm": "AES256"};
  DeleteTableBucketEncryption again -> NotFoundException HTTP 404 'The specified encryption configuration does
  not exist.'; Put AES256 after delete -> HTTP 200; DeleteTableBucketEncryption on explicit AES256 -> HTTP
  204; Put aws:kms with DISABLED key -> HTTP 200. GetTableBucketEncryption after Put aws:kms returned
  {"kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/12aa39a2-b0c3-475c-aa11-bb7fd4033450", "sseAlgorithm":
  "aws:kms"} (kmsKeyArn echoed byte-identical).
  - ACK: custom_update, late_initialize, exceptions.404 · ops: PutTableBucketEncryption,
    GetTableBucketEncryption, DeleteTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(AES256) -> Put AES256 -> Put aws:kms -> Get -> Put AES256 -> Put aws:kms ->
    DeleteTableBucketEncryption -> Get -> Delete again -> Put AES256 -> Delete
  - measurements: put_to_visible_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522; generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-005](#s3t-tablebucket-005), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](#s3t-tablebucket-051), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052),
    [S3T-TABLEBUCKET-072](#s3t-tablebucket-072), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107), [S3T-TABLE-249](table-encryption-storage.md#s3t-table-249), [S3T-TABLEBUCKET-123](#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073),
    [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLE-061](table-encryption-storage.md#s3t-table-061), [S3T-TABLEBUCKET-185](#s3t-tablebucket-185), [S3T-TABLEBUCKET-035](#s3t-tablebucket-035) ·
    hypotheses: H-B-071, H-B-007, H-B-014 · evidence: tablebucket/sub-resources/kms-encryption-forms
  - notes: H-B-071: confirmed for every cell listed. DeleteTableBucketEncryption is the only 'unset' path and
    the read afterwards shows the AES256 default (never 404 for the setting itself), but a second Delete ->
    NotFoundException HTTP 404 'The specified encryption configuration does not exist.'. H-B-014 (b)...
  - full notes: [details/S3T-TABLEBUCKET-037.md](details/S3T-TABLEBUCKET-037.md)

- <a id="s3t-tablebucket-084"></a>**S3T-TABLEBUCKET-084** `sub-resource-api` · impact high · handled · verified 2026-10-13
  **PutTableBucketStorageClass STANDARD->INTELLIGENT_TIERING->STANDARD (HTTP 200, visible on the first Get); existing tables keep their class**
  Table t1 created under the default -> GetTableStorageClass STANDARD. Put(INTELLIGENT_TIERING) -> HTTP 200;
  bucket Get at +0 s -> INTELLIGENT_TIERING, +1 s -> INTELLIGENT_TIERING. CreateTable t2 ->
  GetTableStorageClass INTELLIGENT_TIERING; t1 re-read -> STANDARD. Put(STANDARD) -> HTTP 200; bucket Get ->
  STANDARD; CreateTable t3 -> STANDARD; t2 re-read -> INTELLIGENT_TIERING.
  - ACK: custom_update, docs-only · ops: PutTableBucketStorageClass, GetTableBucketStorageClass, CreateTable,
    GetTableStorageClass · fields: storageClassConfiguration.storageClass
  - repro: CreateTable t1; Put(INTELLIGENT_TIERING); Get +0/+1 s; CreateTable t2; read t1/t2; Put(STANDARD);
    Get; CreateTable t3; read t3/t2
  - measurements: bucket_get_reflects_put_on_first_read=true, n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLEBUCKET-085](#s3t-tablebucket-085), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161),
    [S3T-TABLEBUCKET-145](#s3t-tablebucket-145), [S3T-TABLE-260](table-encryption-storage.md#s3t-table-260), [S3T-TABLE-261](table-encryption-storage.md#s3t-table-261), [S3T-TABLE-262](table-encryption-storage.md#s3t-table-262), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020), [S3T-TABLE-021](table-encryption-storage.md#s3t-table-021),
    [S3T-TABLEBUCKET-146](#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](#s3t-tablebucket-147), [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008),
    [S3T-TABLE-266](table-encryption-storage.md#s3t-table-266), [S3T-TABLEBUCKET-072](#s3t-tablebucket-072), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073) · hypotheses: H-S-038, H-S-086 · evidence:
    tablebucket/sub-resources/storage-class-default-transitions
  - notes: H-S-038 confirmed. H-S-086 (contrarian: revert to STANDARD rejected) refuted: Put(STANDARD) after
    INTELLIGENT_TIERING returned HTTP 200.
  - full notes: [details/S3T-TABLEBUCKET-084.md](details/S3T-TABLEBUCKET-084.md)

- <a id="s3t-tablebucket-117"></a>**S3T-TABLEBUCKET-117** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket accepts storageClassConfiguration + encryptionConfiguration (HTTP 200)...**
  CreateTableBucket(name, tags, storageClassConfiguration=INTELLIGENT_TIERING, encryptionConfiguration=AES256)
  -> HTTP 200; GetTableBucketStorageClass -> {'storageClassConfiguration': {'storageClass':
  'INTELLIGENT_TIERING'}}, GetTableBucketEncryption -> {'encryptionConfiguration': {'sseAlgorithm':
  'AES256'}}. A plain bucket then PutTableBucketStorageClass(INTELLIGENT_TIERING) (HTTP 200) and
  PutTableBucketEncryption(AES256) (HTTP 200) reads {'storageClassConfiguration': {'storageClass':
  'INTELLIGENT_TIERING'}} / {'encryptionConfiguration': {'sseAlgorithm': 'AES256'}}. The CreateTableBucket
  input shape has members ['encryptionConfiguration', 'name', 'storageClassConfiguration', 'tags']: policy,
  maintenance, metrics and replication cannot be set at create.
  - ACK: custom_create, scope:field-on-parent · ops: CreateTableBucket, GetTableBucketStorageClass,
    GetTableBucketEncryption, PutTableBucketStorageClass, PutTableBucketEncryption · fields:
    storageClassConfiguration, encryptionConfiguration
  - repro: CreateTableBucket with both configs -> Gets; CreateTableBucket plain -> Puts -> Gets; diff; inspect
    the model
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-146](#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](#s3t-tablebucket-147), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009),
    [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-060](tablebucket.md#s3t-tablebucket-060), [S3T-TABLEBUCKET-116](tablebucket-replication.md#s3t-tablebucket-116) · hypotheses: H-S-041,
    H-S-012 · evidence: tablebucket/identity/subresources-across-recreate
  - notes: H-S-041 confirmed. H-S-012 qualified: only the AES256 cell was run (create-time and post-create
    bodies identical); the aws:kms-without-maintenance-access cell needs a KMS key and was not run in this
    probe.
  - full notes: [details/S3T-TABLEBUCKET-117.md](details/S3T-TABLEBUCKET-117.md)

- <a id="s3t-tablebucket-120"></a>**S3T-TABLEBUCKET-120** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket SSE-KMS applies only to tables created after the Put (t0 before: AES256; t1 after: aws:kms)...**
  After granting maintenance.s3tables.amazonaws.com kms:GenerateDataKey+kms:Decrypt on the key (30 s wait),
  CreateTable t1 -> HTTP 200 (attempts: [(0.6, 'HTTP 200')]). GetTableEncryption: t0 (created before
  PutTableBucketEncryption) -> {'sseAlgorithm': 'AES256'}; t1 (created after) -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}.
  DeleteTableBucketEncryption -> HTTP 204, GetTableBucketEncryption -> {'sseAlgorithm': 'AES256'}; CreateTable
  t2 -> HTTP 200, its encryption -> {'sseAlgorithm': 'AES256'}; t1 afterwards -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}.
  - ACK: e2e-timing, docs-only · ops: PutTableBucketEncryption, DeleteTableBucketEncryption, CreateTable,
    GetTableEncryption, GetTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable t0 -> PutTableBucketEncryption(aws:kms) -> grant maintenance principal -> CreateTable
    t1 -> GetTableEncryption t0,t1 -> DeleteTableBucketEncryption -> CreateTable t2 -> GetTableEncryption
    t1,t2
  - measurements: key_policy_wait_s=30, create_table_attempts_after_grant=1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLEBUCKET-180](#s3t-tablebucket-180), [S3T-TABLE-255](table-encryption-storage.md#s3t-table-255), [S3T-TABLE-259](table-encryption-storage.md#s3t-table-259), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257),
    [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057), [S3T-TABLE-010](table-encryption-storage.md#s3t-table-010) · hypotheses: H-S-009, H-S-007 · evidence:
    tablebucket/dependencies/kms-key-policy-gate
  - notes: H-S-009 confirmed. H-S-007 second half (CreateTable succeeds once the key policy grants the
    maintenance principal; the bucket kept the KMS configuration): confirmed.
  - full notes: [details/S3T-TABLEBUCKET-120.md](details/S3T-TABLEBUCKET-120.md)

## Delete semantics

- <a id="s3t-tablebucket-020"></a>**S3T-TABLEBUCKET-020** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket policy, maintenance configuration, explicit encryption and tags do not block DeleteTableBucket (HTTP 204)...**
  Bucket created with AES256 + 2 tags, then PutTableBucketPolicy (HTTP 200) and
  PutTableBucketMaintenanceConfiguration(icebergUnreferencedFileRemoval enabled) (HTTP 204); DeleteTableBucket
  -> HTTP 204 with body '' in one call. On the deleted ARN: {"DeleteTableBucketPolicy":
  "NotFoundException/404", "GetTableBucket": "NotFoundException/404", "GetTableBucketEncryption":
  "NotFoundException/404", "GetTableBucketMaintenanceConfiguration": "NotFoundException/404",
  "GetTableBucketPolicy": "NotFoundException/404", "GetTableBucketStorageClass": "NotFoundException/404",
  "ListNamespaces": "NotFoundException/404", "ListTagsForResource": "NotFoundException/404",
  "PutTableBucketEncryption": "NotFoundException/404"}. Distinct messages: ['The specified bucket does not
  exist.'].
  - ACK: none, tags.custom-sync · ops: DeleteTableBucket, PutTableBucketPolicy,
    PutTableBucketMaintenanceConfiguration, GetTableBucketPolicy, GetTableBucketEncryption,
    GetTableBucketMaintenanceConfiguration, ListTagsForResource
  - repro: CreateTableBucket(AES256, tags) -> PutTableBucketPolicy -> PutTableBucketMaintenanceConfiguration
    -> DeleteTableBucket -> each Get* on the ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-079](tablebucket-maintenance-metrics.md#s3t-tablebucket-079), [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158) · hypotheses: H-B-026 · evidence:
    tablebucket/dependencies/delete-with-children-and-configs
  - notes: Confirms H-B-026. Full original title: Bucket policy, maintenance configuration, explicit
    encryption and tags do not block DeleteTableBucket (HTTP 204); all sub-resource reads then
    NotFoundException

- <a id="s3t-tablebucket-051"></a>**S3T-TABLEBUCKET-051** `delete-semantics` · impact medium · handled · verified 2026-10-13
  **After Put(aws:kms) or Put(AES256) then DeleteTableBucketEncryption, Get returns exactly the never-configured body**
  Never-configured Get: '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. Put(aws:kms,key) ->
  Delete (HTTP 204) -> Get: '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. Put(AES256) -> Delete
  (HTTP 204) -> Get: '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'. No residual kmsKeyArn and no
  distinct deleted state in these 2 sequences.
  - ACK: custom_delete, compare.nil_equals_zero_value · ops: DeleteTableBucketEncryption,
    GetTableBucketEncryption, PutTableBucketEncryption · fields: encryptionConfiguration
  - repro: Fresh Get (raw) -> Put(aws:kms) -> Delete -> Get (raw) -> Put(AES256) -> Delete -> Get (raw);
    byte-compare status+body
  - measurements: n=2
  - handling: handled via `pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-005](#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052),
    [S3T-TABLEBUCKET-072](#s3t-tablebucket-072) · hypotheses: H-S-130 · evidence: tablebucket/sub-resources/encryption-get-put-delete
  - notes: H-S-130 confirmed (byte comparison of HTTP status + body, key ARN masked).

- <a id="s3t-tablebucket-052"></a>**S3T-TABLEBUCKET-052** `delete-semantics` · impact medium · handled · verified 2026-10-13
  **DeleteTableBucketEncryption on a bucket created WITH encryptionConfiguration (aws:kms) returns HTTP 204...**
  Bucket created with encryptionConfiguration aws:kms: Get -> '200
  {"encryptionConfiguration":{"kmsKeyArn":"<kms-key-arn>","sseAlgorithm":"aws:kms"}}'; Delete -> HTTP 204; Get
  -> '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (equals the never-configured body). Bucket
  created with explicit AES256: Get -> '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (identical
  to the never-configured body); Delete -> HTTP 204; Get -> '200
  {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'.
  - ACK: is_immutable, custom_delete · ops: CreateTableBucket, DeleteTableBucketEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration
  - repro: CreateTableBucket(encryptionConfiguration=aws:kms) -> Get -> Delete -> Get;
    CreateTableBucket(AES256) -> Get -> Delete -> Get
  - measurements: n=1
  - handling: handled via `pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-005](#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](#s3t-tablebucket-051),
    [S3T-TABLEBUCKET-072](#s3t-tablebucket-072), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103), [S3T-TABLE-059](table-encryption-storage.md#s3t-table-059) ·
    hypotheses: H-S-132 · evidence: tablebucket/sub-resources/encryption-get-put-delete
  - notes: H-S-132 (contrarian: create-time encryption stickier) refuted: create-time encryption is removable
    by Delete. Explicit AES256 at create is indistinguishable from the default in Get.
  - full notes: [details/S3T-TABLEBUCKET-052.md](details/S3T-TABLEBUCKET-052.md)

- <a id="s3t-tablebucket-072"></a>**S3T-TABLEBUCKET-072** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketEncryption with a KMS-encrypted table present returns HTTP 204; the table keeps aws:kms+key, a new table gets AES256**
  Put(aws:kms,key) on the bucket -> HTTP 200; CreateTable t1 -> HTTP 200, GetTableEncryption(t1) ->
  aws:kms+key. DeleteTableBucketEncryption -> HTTP 204; GetTableBucketEncryption -> AES256;
  GetTableEncryption(t1) right after -> aws:kms+key and 10 s later -> aws:kms+key (kmsKeyArn unchanged).
  CreateTable t2 10 s after the Delete -> HTTP 200; GetTableEncryption(t2) -> AES256.
  - ACK: docs-only, custom_delete · ops: DeleteTableBucketEncryption, GetTableEncryption, CreateTable ·
    fields: encryptionConfiguration
  - repro: Put(aws:kms) -> CreateTable t1 -> GetTableEncryption(t1) -> DeleteTableBucketEncryption ->
    GetTableEncryption(t1) -> CreateTable t2 -> GetTableEncryption(t2)
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-005](#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](#s3t-tablebucket-051),
    [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020), [S3T-TABLE-266](table-encryption-storage.md#s3t-table-266), [S3T-TABLE-161](table-encryption-storage.md#s3t-table-161), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-073](#s3t-tablebucket-073)
    · hypotheses: H-S-135 · evidence: tablebucket/dependencies/encryption-tables-inherit
  - notes: H-S-135 confirmed: Delete succeeded (204) with a KMS table present; the table's encryption did not
    change; bucket default and table setting are independent afterwards.

## Cross-region

- <a id="s3t-tablebucket-126"></a>**S3T-TABLEBUCKET-126** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Other-region bucket ARN on this endpoint -> 11 bucket ops (Get*/Put/Tag/Untag/ListNamespaces/Delete) all 400 region mismatch, both ways**
  With bucket X existing in BOTH regions under the same name: the us-west-2 endpoint given the us-east-1 ARN
  -> {'BadRequestException (HTTP 400) The provided Region does not match the Region of the resource.':
  ['GetTableBucket', 'GetTableBucketPolicy', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketStorageClass', 'PutTableBucketStorageClass',
  'ListNamespaces', 'ListTagsForResource', 'TagResource', 'UntagResource', 'DeleteTableBucket']}; the
  us-east-1 endpoint given the us-west-2 ARN -> {'BadRequestException (HTTP 400) The provided Region does not
  match the Region of the resource.': ['GetTableBucket', 'GetTableBucketPolicy', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketStorageClass', 'PutTableBucketStorageClass',
  'ListNamespaces', 'ListTagsForResource', 'TagResource', 'UntagResource', 'DeleteTableBucket']}. Afterwards
  the us-east-1 bucket still exists (HTTP 200) with tags {'ackq-run': '5443b8', 'side': 'B'} (no foreign tag
  landed, 'side' tag not removed); the us-west-2 bucket HTTP 200, tags {'ackq-run': '5443b8', 'side': 'A'},
  storage class {'storageClassConfiguration': {'storageClass': 'STANDARD'}}. Cross-region DeleteTableBucket
  deleted nothing (B deleted: False, A deleted: False).
  - ACK: terminal_codes, is_arn_primary_key · ops: DeleteTableBucket, GetTableBucket,
    GetTableBucketEncryption, GetTableBucketMaintenanceConfiguration, GetTableBucketPolicy,
    GetTableBucketStorageClass, ListNamespaces, ListTagsForResource, PutTableBucketStorageClass, TagResource,
    UntagResource · fields: tableBucketARN, resourceArn
  - repro: CreateTableBucket(X) in both regions; from each endpoint call the 11 bucket ops with the OTHER
    region's ARN; verify both buckets via their own region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLEBUCKET-057](tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](tablebucket-maintenance-metrics.md#s3t-tablebucket-100) · hypotheses: H-B-062,
    H-B-063, H-S-070 · evidence: tablebucket/cross-region/same-name-two-regions-arn-routing
  - notes: H-B-062 routing part confirmed, H-B-063 (contrarian: routes by ARN, cross-region delete works)
    refuted, H-S-070 confirmed (BadRequestException, not 404, no redirect) for GetTableBucketPolicy,
    PutTableBucketStorageClass, GetTableBucketEncryption/Maintenance/StorageClass, Tag/Untag/ListTags,...
  - full notes: [details/S3T-TABLEBUCKET-126.md](details/S3T-TABLEBUCKET-126.md)

## Handling gaps (bugs to file)

- [S3T-TABLEBUCKET-087](#s3t-tablebucket-087) - PutTableBucketStorageClass enum/null violations return wire code ValidationException
  (HTTP 400), a shape NOT in the API model (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven. A misspelled
  spec.storageClassConfiguration.storageClass on an existing TableBucket is never rejected by the CRD and is
  re-Put on every reconcile forever (ValidationException is not in terminal_codes), while the same value at
  create goes Terminal via BadRequestException - add an enum marker on storageClass or add ValidationException
  to the TableBucket terminal codes. code:
  config/crd/bases/s3tables.services.k8s.aws_tablebuckets.yaml:116-117 storageClass is a plain 'type: string'
  (no enum; apis/v1alpha1/types.go:164 carries no kubebuilder marker), so 'standard' passes admission;
  pkg/resource/table_bucket/hooks.go:519-536 syncStorageClass sends the raw string to
  PutTableBucketStorageClass and returns the error unchanged; pkg/resource/table_bucket/sdk.go:400-405
  terminalAWSError matches only BadRequestException, so the ValidationException is Recoverable and the
  reconciler requeues; hooks.go:98-105 reads STANDARD back into latest so delta.go keeps the 'standard' !=
  'STANDARD' delta alive each cycle; create path goes Terminal (generator.yaml:267-268). finding:
  [S3T-TABLEBUCKET-087](#s3t-tablebucket-087) (ValidationException 400 on PutTableBucketStorageClass, 7/7), [S3T-TABLEBUCKET-024](#s3t-tablebucket-024)
  (BadRequestException 400 for the same value on CreateTableBucket).
  - handling_ref: `pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLEBUCKET-009](#s3t-tablebucket-009) | storageClassConfiguration is mutable via PutTableBucketStorageClass (HTTP 200, visible after 0.01 s); no Delete op exists to unset it | put_to_visible_s=[0.01, 0.01] | 2 |
| [S3T-TABLEBUCKET-010](#s3t-tablebucket-010) | GetTableBucket/ListTableBuckets createdAt and tableBucketId byte-identical across 4 reads over 2 min and tag/encryption/storage-class Puts | reads=4, span_s=120.4 | 4 |
| [S3T-TABLEBUCKET-037](#s3t-tablebucket-037) | PutTableBucketEncryption switches AES256<->aws:kms both ways (visible 0.01 s); unchanged re-send 200; Delete on aws:kms -> 204, reads AES256 | put_to_visible_s=[0.01, 0.01] | 2 |
| [S3T-TABLEBUCKET-049](#s3t-tablebucket-049) | DeleteTableBucketEncryption is NOT idempotent - NotFoundException 404 when never set and on a second Delete; 204 only when a config exists | get_converged_on_first_read=true | 1 |
| [S3T-TABLEBUCKET-058](#s3t-tablebucket-058) | Re-sending an unchanged maintenance value 3x -> ['HTTP 204'] each; unchanged storage class 2x -> ['HTTP 200']; no ConflictException | maint_put_latency_ms=[374, 42, 303], sc_put_latency_ms=[14, 14] | 5 |
| [S3T-TABLEBUCKET-059](#s3t-tablebucket-059) | Maintenance status and storage class: 10 alternating Puts each, Get at +0/+0.3/+1 s -> 0 / 0 stale reads | maint_get_latency_ms_min_max=[23, 329], sc_get_latency_ms_min_max=[7, 15] | 30 |
| [S3T-TABLEBUCKET-073](#s3t-tablebucket-073) | Tables created 0.49-6.68 s after Put(aws:kms)/Delete of the bucket encryption all inherited the new default (n=5 per direction) | first_table_created_after_mutation_s=[0.53, 0.49], lagging_tables=0 | 10 |
| [S3T-TABLEBUCKET-082](#s3t-tablebucket-082) | 10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (no 409), last writer wins; 10 maintenance Puts 204 but ~6.5 s | wall_s.policy=6.29, wall_s.maintenance=6.46, wall_s.encryption=0.18, wall_s.storageclass=0.24, latency_ms_min_max.policy=[118, 6281], latency_ms_min_max.maintenance=[134, 6452], latency_ms_min_max.encryption=[10, 178], latency_ms_min_max.storageclass=[16, 226] | 50 |
| [S3T-TABLEBUCKET-084](#s3t-tablebucket-084) | PutTableBucketStorageClass STANDARD->INTELLIGENT_TIERING->STANDARD (HTTP 200, visible on the first Get); existing tables keep their class | bucket_get_reflects_put_on_first_read=true | 2 |
| [S3T-TABLEBUCKET-108](#s3t-tablebucket-108) | Read-after-write of bucket storage class / policy / encryption: 20/20, 20/20 (+20/20 delete), 20/20 immediate Gets reflected the mutation | storage_class_consistent=20, policy_put_consistent=20, policy_delete_consistent=20, encryption_put_consistent=20, encryption_second_delete_404=20, put_latency_ms_storage_class=[13, 53], put_latency_ms_policy=[93, 135], get_latency_ms_policy=[29, 68] | 60 |
| [S3T-TABLEBUCKET-112](#s3t-tablebucket-112) | Sub-resource Puts sent <1 ms after CreateTableBucket returned: 48/48 first attempts succeeded on 12 fresh buckets (metrics/policy/maint/SC) | first_put_offset_ms_min_max=[0, 0], put_success_at_ms_from_create.metrics=[70, 128], put_success_at_ms_from_create.policy=[175, 248], put_success_at_ms_from_create.maintenance=[550, 642], put_success_at_ms_from_create.storage_class=[564, 655], create_latency_ms_min_max=[1435, 1940] | 12 |
| [S3T-TABLEBUCKET-120](#s3t-tablebucket-120) | Bucket SSE-KMS applies only to tables created after the Put (t0 before: AES256; t1 after: aws:kms)... | key_policy_wait_s=30, create_table_attempts_after_grant=1 | 1 |
| [S3T-TABLEBUCKET-123](#s3t-tablebucket-123) | TagResource at +0 ms and PutTableBucketEncryption at +69-142 ms after CreateTableBucket -> 5/5 first attempts 200, no ConflictException | tag_offset_ms=[0, 0, 0, 0, 0], encryption_offset_ms=[93, 69, 142, 74, 76], tag_first_attempt_ok=5, encryption_first_attempt_ok=5, create_latency_ms=[1562, 1600, 1646, 1823, 1561] | 5 |
| [S3T-TABLEBUCKET-180](#s3t-tablebucket-180) | SSE-KMS bucket with a default-policy key: CreateTableBucket 200 but CreateTable 400 until the key policy grants the maintenance principal | n_create_table_before=2, n_create_table_after=2 | 2 |
| [S3T-TABLEBUCKET-181](#s3t-tablebucket-181) | PutTableBucketEncryption(aws:kms) needs caller kms:DescribeKey, CreateTable needs kms:GenerateDataKey; denied -> ForbiddenException 403 | n_put_per_variant=2, iam_propagation_wait_s=12 | 2 |
| [S3T-TABLEBUCKET-182](#s3t-tablebucket-182) | Disabled KMS key (used seconds before): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; EnableKey recovers | offsets_s=[3, 30, 120, 300], recovery_after_enable_s=0.04 | 1 |
| [S3T-TABLEBUCKET-183](#s3t-tablebucket-183) | Disabled KMS key (idle 10 min): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; same as warm-key series | idle_s=600, offsets_s=[3, 30, 120, 300], recovery_after_enable_s=0.04 | 1 |
| [S3T-TABLEBUCKET-185](#s3t-tablebucket-185) | Disabled or PendingDeletion KMS key: PutTableBucketEncryption and CreateTableBucket(aws:kms) -> 400 'is disabled.' at +5/+30/+120 s (30/30) | disabled_offsets_s=[5, 30, 120], pending_offsets_s=[5, 60], trials_per_offset=3, put_cells=15, create_cells=15 | 30 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `tablebucket-encryption-storage.md` - unverified or stale items worth a future probe, one line each.

- [S3T-TABLEBUCKET-185](#s3t-tablebucket-185): Put/CreateTableBucket(aws:kms) with a key disabled less than ~6 s earlier is untested (earliest
  cell was +6.4 s); if a KMS key-state propagation window exists it is shorter than that.
- [S3T-TABLEBUCKET-037](#s3t-tablebucket-037): the "Put aws:kms with a DISABLED key -> HTTP 200" reading is superseded (confidence low) by -185
  (30/30 -> 400); keep as evidence only, do not derive a rule from it.
- [S3T-TABLEBUCKET-107](#s3t-tablebucket-107): another account's key at PutTableBucketEncryption was not tested; only the Create side is known
  (ForbiddenException 403 naming kms:DescribeKey, -035).
- [S3T-TABLEBUCKET-180](#s3t-tablebucket-180), -120: the key-policy grant was observed to take effect at +5 s (2/2) and after a 30 s wait (1/1);
  nothing shorter than 5 s was probed, so the lower bound of the propagation window is unknown.
- [S3T-TABLEBUCKET-072](#s3t-tablebucket-072): an existing KMS-encrypted table kept aws:kms+key only up to +10 s after DeleteTableBucketEncryption;
  whether a longer horizon ever changes an existing table's encryption was not observed.
- [S3T-TABLEBUCKET-082](#s3t-tablebucket-082): the "no 500 under concurrency" result for storage class (20 x 200) and encryption Put/Delete
  (5+5) is one burst each; the policy race that did 500 is also a single run.
- [S3T-TABLEBUCKET-129](tablebucket-policy.md#s3t-tablebucket-129): the AccessDeniedException naming s3tables:PutTableBucketEncryption / PutTableBucketStorageClass at
  CreateTableBucket is n=1, as is "tags checked first".
- [S3T-TABLEBUCKET-106](tablebucket-policy.md#s3t-tablebucket-106): GetTableBucketEncryption succeeding under an explicit Deny kms:* is n=1.
- [S3T-TABLEBUCKET-105](#s3t-tablebucket-105), -128: cross-region rejection was probed with one bucket/key pair per direction (n=1 each); the
  message is shared with "nonexistent key in another region", so the region-before-existence ordering rests on -105's
  single sample.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-tablebucket-006"></a>**S3T-TABLEBUCKET-006** | server-default | medium | confirmed | GetTableBucketStorageClass on a bucket created without storageClassConfiguration returns HTTP 200 with {"storageClass": "STANDARD"} | - | [S3T-TABLEBUCKET-083](#s3t-tablebucket-083) |
| <a id="s3t-tablebucket-024"></a>**S3T-TABLEBUCKET-024** | request-validation | low | confirmed | Enums are closed and case-sensitive: sseAlgorithm aes256/AWS:KMS, storageClass standard/GLACIER, ListTableBuckets type=Customer/AWS -> 400 | [S3T-TABLEBUCKET-026](tablebucket.md#s3t-tablebucket-026), [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLEBUCKET-104](#s3t-tablebucket-104), [S3T-TABLEBUCKET-087](#s3t-tablebucket-087), [S3T-TABLE-087](table-encryption-storage.md#s3t-table-087), [S3T-TABLE-013](table-encryption-storage.md#s3t-table-013), [S3T-TABLE-012](table-encryption-storage.md#s3t-table-012), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020) | - |
| <a id="s3t-tablebucket-036"></a>**S3T-TABLEBUCKET-036** | prerequisite | medium | confirmed | CreateTableBucket with aws:kms succeeds (HTTP 200) with a CMK whose key policy grants nothing to maintenance.s3tables.amazonaws.com | - | [S3T-TABLEBUCKET-180](#s3t-tablebucket-180) |
| <a id="s3t-tablebucket-048"></a>**S3T-TABLEBUCKET-048** | server-default | high | confirmed | GetTableBucketEncryption on a bucket created without encryptionConfiguration -> 200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}} | [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050) | [S3T-TABLEBUCKET-005](#s3t-tablebucket-005) |
| <a id="s3t-tablebucket-058"></a>**S3T-TABLEBUCKET-058** | idempotency | low | confirmed | Re-sending an unchanged maintenance value 3x -> ['HTTP 204'] each; unchanged storage class 2x -> ['HTTP 200']; no ConflictException | [S3T-TABLEBUCKET-056](tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-112](#s3t-tablebucket-112) | - |
| <a id="s3t-tablebucket-067"></a>**S3T-TABLEBUCKET-067** | error-code | medium | confirmed | Wire codes on bucket/policy/encryption Gets: BadRequestException=400, NotFoundException=404; x-amzn-ErrorType is '<Shape>:<coral URI>' | [S3T-TABLEBUCKET-109](#s3t-tablebucket-109), [S3T-TABLE-083](table-policy.md#s3t-table-083), [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLE-109](table-policy.md#s3t-table-109), [S3T-TABLE-120](table.md#s3t-table-120) | [S3T-TABLEBUCKET-109](#s3t-tablebucket-109) |
| <a id="s3t-tablebucket-086"></a>**S3T-TABLEBUCKET-086** | request-validation | low | confirmed | PutTableBucketStorageClass: 'standard', 'Standard', 'STANDARD_IA', 'GLACIER', 'INTELLIGENT-TIERING', '' and {} -> ValidationException 400 | - | [S3T-TABLEBUCKET-087](#s3t-tablebucket-087) |
| <a id="s3t-tablebucket-104"></a>**S3T-TABLEBUCKET-104** | codegen-artifact | low | confirmed | PutTableBucketEncryption sseAlgorithm is the literal 'aws:kms': 'AWS_KMS', 'aws_kms', 'AWS:KMS', 'Aes256' -> BadRequestException 400 | [S3T-TABLEBUCKET-026](tablebucket.md#s3t-tablebucket-026), [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLEBUCKET-024](#s3t-tablebucket-024), [S3T-TABLEBUCKET-087](#s3t-tablebucket-087), [S3T-TABLE-087](table-encryption-storage.md#s3t-table-087), [S3T-TABLE-013](table-encryption-storage.md#s3t-table-013) | - |
| <a id="s3t-tablebucket-114"></a>**S3T-TABLEBUCKET-114** | tag-semantics | low | confirmed | Table bucket tags unchanged after encryption/policy/maintenance/storage-class/metrics Puts and after DeleteTableBucketEncryption | - | - |
| <a id="s3t-tablebucket-128"></a>**S3T-TABLEBUCKET-128** | request-validation | low | confirmed | CreateTableBucket aws:kms with a us-east-1 key ARN on us-west-2 -> BadRequestException 'Cross region KMS keys are not supported.', no bucket | [S3T-TABLEBUCKET-035](#s3t-tablebucket-035), [S3T-TABLEBUCKET-105](#s3t-tablebucket-105), [S3T-TABLEBUCKET-107](#s3t-tablebucket-107), [S3T-TABLEBUCKET-131](tablebucket.md#s3t-tablebucket-131), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) | - |
| <a id="s3t-tablebucket-133"></a>**S3T-TABLEBUCKET-133** | other | low | confirmed | Doc claim C001 TRUE: CreateTableBucket{AES256} -> 200; GetTableBucketEncryption returns sseAlgorithm=AES256, no kmsKeyArn | [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) | - |
| <a id="s3t-tablebucket-134"></a>**S3T-TABLEBUCKET-134** | other | low | confirmed | Doc claim C002 TRUE: CreateTableBucket{aws:kms, kmsKeyArn} -> 200; GetTableBucketEncryption echoes aws:kms + kmsKeyArn byte-identical | [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLEBUCKET-036](#s3t-tablebucket-036), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-139](#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](#s3t-tablebucket-142), [S3T-TABLE-256](table-maintenance.md#s3t-table-256), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-119](#s3t-tablebucket-119), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) | - |
| <a id="s3t-tablebucket-135"></a>**S3T-TABLEBUCKET-135** | other | low | confirmed | Doc claim C003 TRUE: CreateTableBucket{aws:kms} without kmsKeyArn -> BadRequestException 400 and no bucket exists afterwards | [S3T-TABLEBUCKET-023](#s3t-tablebucket-023), [S3T-TABLEBUCKET-102](#s3t-tablebucket-102), [S3T-TABLE-258](table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](#s3t-tablebucket-138), [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) | - |
| <a id="s3t-tablebucket-136"></a>**S3T-TABLEBUCKET-136** | other | low | confirmed | Doc claim C004 TRUE: CreateTableBucket{AES256 + kmsKeyArn} -> BadRequestException 400 'KMS key ARN detected...' (key not silently dropped) | [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLEBUCKET-102](#s3t-tablebucket-102), [S3T-TABLEBUCKET-135](#s3t-tablebucket-135), [S3T-TABLE-258](table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-137](#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](#s3t-tablebucket-138), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) | - |
| <a id="s3t-tablebucket-137"></a>**S3T-TABLEBUCKET-137** | other | low | confirmed | Doc claim C005 TRUE: sseAlgorithm accepts only AES256/aws:kms; any other string -> BadRequestException 400 'SSE algorithm is invalid' | [S3T-TABLEBUCKET-024](#s3t-tablebucket-024), [S3T-TABLE-087](table-encryption-storage.md#s3t-table-087), [S3T-TABLEBUCKET-135](#s3t-tablebucket-135), [S3T-TABLE-258](table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](#s3t-tablebucket-136), [S3T-TABLEBUCKET-138](#s3t-tablebucket-138), [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085) | - |
| <a id="s3t-tablebucket-138"></a>**S3T-TABLEBUCKET-138** | other | low | confirmed | Doc claim C006 TRUE: CreateTableBucket kmsKeyArn: bare key id / alias name / alias ARN -> BadRequestException 400; only a full key ARN | [S3T-TABLEBUCKET-034](#s3t-tablebucket-034), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103), [S3T-TABLE-085](table-encryption-storage.md#s3t-table-085), [S3T-TABLEBUCKET-135](#s3t-tablebucket-135), [S3T-TABLE-258](table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](#s3t-tablebucket-137), [S3T-TABLE-084](table-encryption-storage.md#s3t-table-084) | - |
| <a id="s3t-tablebucket-140"></a>**S3T-TABLEBUCKET-140** | other | low | confirmed | Doc claim C008 TRUE: PutTableBucketEncryption{AES256} on a bucket created without encryption -> 200; Get shows AES256 immediately | [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-108](#s3t-tablebucket-108), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048), [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050) | - |
| <a id="s3t-tablebucket-141"></a>**S3T-TABLEBUCKET-141** | other | low | confirmed | Doc claim C009 TRUE: PutTableBucketEncryption{aws:kms, kmsKeyArn} -> 200; Get immediately returns aws:kms and the kmsKeyArn byte-identical | [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-103](#s3t-tablebucket-103), [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) | - |
| <a id="s3t-tablebucket-143"></a>**S3T-TABLEBUCKET-143** | other | low | confirmed | Doc claim C011 TRUE: Second PutTableBucketEncryption with a different configuration (AES256 -> aws:kms) -> 200; Get returns only the latest | [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-144](#s3t-tablebucket-144), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) | - |
| <a id="s3t-tablebucket-144"></a>**S3T-TABLEBUCKET-144** | other | low | confirmed | Doc claim C012 TRUE: After DeleteTableBucketEncryption on an aws:kms bucket, Get returns the AES256 default body (200), no key ARN, not 404 | [S3T-TABLEBUCKET-051](#s3t-tablebucket-051), [S3T-TABLEBUCKET-037](#s3t-tablebucket-037), [S3T-TABLEBUCKET-052](#s3t-tablebucket-052), [S3T-TABLEBUCKET-049](#s3t-tablebucket-049), [S3T-TABLEBUCKET-133](#s3t-tablebucket-133), [S3T-TABLEBUCKET-134](#s3t-tablebucket-134), [S3T-TABLEBUCKET-140](#s3t-tablebucket-140), [S3T-TABLEBUCKET-141](#s3t-tablebucket-141), [S3T-TABLEBUCKET-143](#s3t-tablebucket-143), [S3T-TABLEBUCKET-050](#s3t-tablebucket-050), [S3T-TABLEBUCKET-048](#s3t-tablebucket-048) | - |
| <a id="s3t-tablebucket-145"></a>**S3T-TABLEBUCKET-145** | other | low | confirmed | Doc claim C018 TRUE: PutTableBucketStorageClass{INTELLIGENT_TIERING} -> 200; GetTableBucketStorageClass returns INTELLIGENT_TIERING | [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-083](#s3t-tablebucket-083), [S3T-TABLE-260](table-encryption-storage.md#s3t-table-260), [S3T-TABLE-261](table-encryption-storage.md#s3t-table-261), [S3T-TABLE-262](table-encryption-storage.md#s3t-table-262), [S3T-TABLE-020](table-encryption-storage.md#s3t-table-020), [S3T-TABLE-021](table-encryption-storage.md#s3t-table-021) | - |
| <a id="s3t-tablebucket-146"></a>**S3T-TABLEBUCKET-146** | other | low | confirmed | Doc claim C021 TRUE: PutTableBucketStorageClass twice (INTELLIGENT_TIERING then STANDARD) -> 200/200; Get returns the latest value | [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059), [S3T-TABLEBUCKET-147](#s3t-tablebucket-147), [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008) | - |
| <a id="s3t-tablebucket-147"></a>**S3T-TABLEBUCKET-147** | other | low | confirmed | Doc claim C023 TRUE: CreateTableBucket{storageClassConfiguration: INTELLIGENT_TIERING} -> 200; only GetTableBucketStorageClass shows it | [S3T-TABLEBUCKET-117](#s3t-tablebucket-117), [S3T-TABLEBUCKET-009](#s3t-tablebucket-009), [S3T-TABLEBUCKET-008](#s3t-tablebucket-008), [S3T-TABLEBUCKET-146](#s3t-tablebucket-146), [S3T-TABLEBUCKET-084](#s3t-tablebucket-084), [S3T-TABLEBUCKET-059](#s3t-tablebucket-059) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
