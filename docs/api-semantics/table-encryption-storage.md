<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table encryption and storage class
_Create-only encryptionConfiguration/storageClassConfiguration, creation-time snapshot of bucket defaults, GetTableEncryption/GetTableStorageClass read paths, KMS key-state and permission effects on CreateTable._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-encryption-storage.md`; every claim cites the finding ids that carry the
full evidence below.

### Orientation

This document covers the two create-only blocks of a Table - `encryptionConfiguration` (`sseAlgorithm` `AES256` or
`aws:kms` plus `kmsKeyArn`) and `storageClassConfiguration` (`STANDARD` or `INTELLIGENT_TIERING`) - their only read
paths GetTableEncryption and GetTableStorageClass, and what the KMS key's form and state do to CreateTable. There is no
state machine and no update path: CreateTable is synchronous and answers with only `tableARN` and `versionToken`;
GetTable (by triple or `tableArn`) and ListTables never carry either block; the model has no UpdateTable and no
PutTableStorageClass; RenameTable is the one mutation of the triple and both values follow the table under its new name
([S3T-TABLE-308](#s3t-table-308), -030, -161, -319). Both sub-reads take only the exact `tableBucketARN`+`namespace`+`name` triple and are
read-after-write consistent with CreateTable: an explicit value echoes byte-for-byte and an omitted block never reads as
absent but as the value fixed at creation - a snapshot of the bucket default: it inherits the bucket's `aws:kms`(K) or
`INTELLIGENT_TIERING`, an explicit value wins over the default, and earlier tables keep their value through later bucket
Puts ([S3T-TABLE-062](#s3t-table-062), -061, -059, -010, -011, -057, -020, -266, -259, -261). On the KMS side only the exact full ARN of an
Enabled key passes - form errors, `aws:kms` without a key, `AES256` with a key, a Disabled or PendingDeletion key are all
BadRequestException 400 with nothing created - while deleting a table, namespace or bucket never needs a usable key
([S3T-TABLE-084](#s3t-table-084), -085, -248, -242, -249).

### Rules

1. Never look for either block in GetTable or ListTables: GetTable by triple and by `tableArn` return the identical key
   set without them (nor metadata/schema or tags); the only reads are GetTableEncryption, GetTableStorageClass and
   ListTagsForResource, one call each ([S3T-TABLE-308](#s3t-table-308)).
2. Key the sub-reads by the exact current triple: a table ARN in `tableBucketARN`, a trailing slash or an uppercased
   bucket name -> BadRequestException 400; an uppercased name or the namespace id -> NotFoundException 404; after
   RenameTable the old triple -> 404, the new triple carries the same config, and GetTable(pre-rename `tableArn`) -> 200
   with the new name and namespace ([S3T-TABLE-062](#s3t-table-062), -319).
3. Treat an omitted block as "inherit the bucket default at creation", never as "unset": without `encryptionConfiguration`
   in an `aws:kms`(K) bucket the table reads `aws:kms` with `kmsKeyArn` == K byte-for-byte; without
   `storageClassConfiguration` in an `INTELLIGENT_TIERING` bucket it reads `INTELLIGENT_TIERING`; only a bucket with no
   bucket-level Put yields `AES256` (kmsKeyArn absent) and `STANDARD` ([S3T-TABLE-057](#s3t-table-057), -020, -260, -010, -011, -059).
4. Do not expect either value to change after CreateTable: no UpdateTable and no PutTableStorageClass exist (only
   GetTableBucketStorageClass, GetTableStorageClass, PutTableBucketStorageClass); tables created earlier kept their class
   across two bucket-default changes; a same-name re-create after DeleteTable (204) read the new `{STANDARD + AES256}`
   with no cached old values ([S3T-TABLE-030](#s3t-table-030), -161, -261, -063).
5. An explicit value wins over the bucket default: explicit `STANDARD` under an `INTELLIGENT_TIERING` default reads
   `STANDARD` (and the reverse), explicit `AES256` under an `aws:kms` default reads `AES256` with `kmsKeyArn` absent
   ([S3T-TABLE-266](#s3t-table-266), -262, -057, -259).
6. Send the enum strings exactly: nine `sseAlgorithm` variants ('aes256', 'AES-256', 'AWS:KMS', 'aws-kms', 'aws:kms ',
   'kms', 'aws:kms:dsse', ...) -> 400 'The specified SSE algorithm is invalid.' (the enum is not named); seven
   `storageClass` variants ('standard', 'INTELLIGENT-TIERING', 'GLACIER', 'ONEZONE_IA', '', ...) -> 400 'The specified
   storage class is not valid.'; `{}` for either block, `{kmsKeyArn only}` or `{sseAlgorithm: ''}` -> botocore client-side
   rejection, on the wire 400; the same `GLACIER` on PutTableBucketStorageClass is ValidationException instead
   ([S3T-TABLE-087](#s3t-table-087), -012, -013, -020).
7. `aws:kms` needs the exact full key ARN and `AES256` must carry none: `aws:kms` without `kmsKeyArn` or with '' ->
   400 'The specified KMS key ARN is invalid. Only key ARN format ... is supported.'; `AES256` plus a `kmsKeyArn`
   (existing or not) -> 400 'KMS key ARN detected. Please specify 'aws:kms' ...'; bare id, alias name, alias ARN or
   uppercased ARN -> 400 'is invalid'; trailing space -> 400 'does not exist.'; only the full key ARN -> 200, echoed
   byte-identical by GetTableEncryption ([S3T-TABLE-084](#s3t-table-084), -258, -085, -255).
8. A Disabled or PendingDeletion key is a 400, not a 404, and the states are indistinguishable: CreateTable(`aws:kms` K)
   -> 400 'The specified KMS key ARN is disabled.' with the same text for both, again on a second attempt 10 s later;
   after EnableKey the same CreateTable -> 200 on its first attempt ([S3T-TABLE-248](#s3t-table-248)).
9. An inherited `aws:kms` default is validated at CreateTable: without `encryptionConfiguration` in a bucket whose key
   lacks the `maintenance.s3tables.amazonaws.com` grant -> 400 'Insufficient access to perform table maintenance ...',
   GetTable -> 404 (nothing created); an AES256 or unconfigured bucket -> 200 ([S3T-TABLE-257](#s3t-table-257)).
10. 'The provided Region does not match the Region of the resource.' (400) is an ARN-region-vs-endpoint mismatch, not an
    invalid spec and not a 404: returned for GetTable (both forms), GetTableBucket, ListNamespaces, GetTableEncryption,
    ListTagsForResource, TagResource and DeleteTable with a foreign-region ARN; the region-rewritten ARN on its matching
    endpoint -> NotFoundException 404; a non-UUID table id -> 400 'The specified table ARN is not valid.' ([S3T-TABLE-252](#s3t-table-252)).
11. Deletion never needs a usable key: with the key Disabled or PendingDeletion GetTable, GetTableMetadataLocation and
    GetTableEncryption still succeed (the latter returns the dead key's ARN), DeleteTable -> 204, GetTable -> 404;
    DeleteNamespace and DeleteTableBucket -> 204 under a PendingDeletion key ([S3T-TABLE-249](#s3t-table-249), -242).
12. DeleteTable has no `tableArn` form (it requires `tableBucketARN`+`namespace`+`name`) while GetTable accepts input
    `tableArn` and returns `tableARN`; keep the triple resolvable at delete time - GetTable(`tableArn`) provides it
    ([S3T-TABLE-030](#s3t-table-030)).
13. Read back immediately: both sub-reads returned exactly the requested values on 30/30 polls at +0/+1/+2 s after
    CreateTable; after DeleteTable all three reads were NotFoundException at once, 0 stale hits in 33 polls over 30 s
    ([S3T-TABLE-061](#s3t-table-061), -063).

### Timing ranges

- CreateTable: 311-447 ms with explicit blocks and no KMS (n=4, [S3T-TABLE-059](#s3t-table-059)); 427-501 ms for `{INTELLIGENT_TIERING +
  aws:kms}` including the KMS pre-check (n=10, -061); in the key-state series the successful creates took 531-652 ms
  (n=2) and the 'is disabled.' 400s returned in 18-36 ms (n=4, -248).
- Sub-reads: GetTableEncryption/GetTableStorageClass 5-39 ms over 60 calls on fresh KMS tables (n=60, -061);
  GetTableEncryption 7-36 ms (n=2, -010; re-verification n=6: first call 24-34 ms, second 4-11 ms); GetTableStorageClass
  6-7 ms (n=2, -011).
- Read-after-write: 30/30 polls at +0/+1/+2 s on 10 tables, 0 NotFoundException and 0 default-valued responses (n=10,
  -061); the immediate read was already final in -010 and -011 (n=2 each) and in all 4 combinations of -059 (n=4).
- DeleteTable: 130-166 ms (n=3 across -063 and -249); the three reads were NotFoundException on the first poll and for
  33 polls over 30 s, 0 stale hits (n=1 table, -063).
- RenameTable across namespaces: old triple 404 and new triple 200 on the first poll after the 204, the four sub-resource
  reads under the old triple 404 at the same moment (n=1, single run, -319; -062).
- Bucket default to new tables: the bucket Get showed the new class on the first read after the Put (n=1, -020); tables
  created 0.5-7 s after a bucket Put inherited it (-260; measured series = the encryption Put/Delete of
  [S3T-TABLEBUCKET-073](tablebucket-encryption-storage.md#s3t-tablebucket-073), 0.49-6.68 s, n=10); a pre-existing table kept its class for the 120 s / 66-poll window after the
  flip, one table each (-261, -262; measurements in [S3T-TABLE-021](#s3t-table-021), `table-maintenance.md`).
- KMS key state: Disabled/Enabled/PendingDeletion visible in DescribeKey 0.01-0.02 s after the KMS call (n=3 changes);
  CreateTable saw the new state on its first attempt at +5 s and gave the same code 10 s later (n=1 per state, -248);
  offsets below 5 s were not probed here (bucket-side offsets +5/+30/+120 s, n=30, in [S3T-TABLEBUCKET-185](tablebucket-encryption-storage.md#s3t-tablebucket-185)).
- Foreign-region ARN: the 'Region does not match' 400 returned in 21-31 ms (n=2, -252).

### Handling gaps

- [S3T-TABLE-021](#s3t-table-021) (partial, [GT-S3T-034](service.md#gt-s3t-034) (controller hooks catalog entry)/-038): `storageClassConfiguration` is `is_immutable: true` on Table (generator.yaml) and
  nothing calls GetTableStorageClass, so the effective class is never read back. Evidence: a table that inherited
  INTELLIGENT_TIERING at create still read INTELLIGENT_TIERING on 66 polls over 120 s after PutTableBucketStorageClass(STANDARD)
  (bucket read STANDARD after 0.01 s, n=2); a no-config table created after the flip reads STANDARD. Immutability is handled; the
  creation-time snapshot stays invisible to a spec that omitted the field - not yet handled.
- [S3T-TABLE-058](#s3t-table-058) (partial, [GT-S3T-034](service.md#gt-s3t-034) (controller hooks catalog entry)/-038): same mechanism for `encryptionConfiguration` (no GetTableEncryption read). Evidence:
  the table inheriting the bucket's aws:kms key was unchanged on 66 polls over 120 s after DeleteTableBucketEncryption, and a
  no-config AES256 table stayed AES256 on 60 polls over 60 s after the bucket was re-keyed (n=126 polls). Not yet handled.

Today `EncryptionConfiguration` and `StorageClassConfiguration` are `is_immutable` and sent only at CreateTable
(generator.yaml:40-50); the read path is GetTable plus ListTagsForResource (generator.yaml:111-118; hooks.go:35-67) and
never calls GetTableEncryption or GetTableStorageClass; `customUpdateTable` handles only Name, MetadataLocation and Tags
(hooks.go:69-139); BadRequestException is the only terminal code (generator.yaml:99-104). No finding here is suspect-bug;
the five `partial` ones are one gap seen from five sides - the effective values are never read back - none handled yet.

- [S3T-TABLE-010](#s3t-table-010) (partial, not yet handled): the effective encryption is never read. Evidence: omitted block, bucket with
  none -> HTTP 200 `{sseAlgorithm: AES256}`, `kmsKeyArn` absent, immediate and later (n=2; 6/6 on re-verification). The
  spec stays nil while AWS reports AES256 - or, in an `aws:kms` bucket, the inherited key (-057); no late-initialize.
- [S3T-TABLE-011](#s3t-table-011) (partial, not yet handled): GetTableStorageClass is never called. Evidence: omitted -> HTTP 200
  `{storageClass: STANDARD}` (n=2); a ghost name -> 404 'The specified table does not exist.' identically on GetTable and
  both sub-reads. An inherited `INTELLIGENT_TIERING` (-020) is invisible in the CR.
- [S3T-TABLE-059](#s3t-table-059) (partial, not yet handled): both blocks are sent as declared and nothing confirms the round-trip.
  Evidence: 4/4 explicit combinations echoed byte-for-byte, unset reads AES256/STANDARD (n=4). Immutability prevents an
  update attempt, but the Spec is never compared with what AWS holds; the CR shows desired, not observed, configuration.
- [S3T-TABLE-062](#s3t-table-062) (partial, not yet handled): no sub-read exists to key, and generator.yaml states the GetTable key two ways
  (18-21: keys off the stable `tableARN`; 111-113: `tableArn` cleared so the read uses the triplet); RenameTable sends
  `newName` only (hooks.go:126-131). Evidence: the sub-reads accept only the exact triple (table ARN -> 400; uppercased
  name or namespace id -> 404); after RenameTable old triple -> 404, new triple keeps the config, GetTable(`tableArn`) ->
  200 with the new name - so a future sub-read must use the post-rename triple from GetTable(`tableArn`).
- [S3T-TABLE-161](#s3t-table-161) (partial, not yet handled): `StorageClassConfiguration` is `is_immutable` and `customUpdateTable` has no
  storage-class branch (hooks.go:69-112) - correct, the model has no PutTableStorageClass and no UpdateTable. Evidence: t1
  and t2 kept their class across two bucket-default changes (n=2). The gap is the read side: a bucket-default change
  between two creates is not reflected on either table's CR.

### Where next

- `tablebucket-encryption-storage.md` is the bucket side of rules 3-6 and 8-9: Put/Get/DeleteTableBucketEncryption and
  Put/GetTableBucketStorageClass, the ValidationException-vs-BadRequestException split, key-policy grant and key-state
  prerequisites ([S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), -185), inheritance timing (-073, -084, -120), the superseded "HTTP 200 for a
  just-disabled key" reading (-037, confidence low).
- `table-maintenance.md` carries the creation-time-snapshot findings that landed there via its `snapshot` include
  ([S3T-TABLE-021](#s3t-table-021), -058, -318), the 'Insufficient access to perform table maintenance' 400 behind rule 9 (-054, -055) and
  the Disabled-key maintenance series (-253, -301, -246). `table-policy.md` has the caller side of SSE-KMS
  (kms:DescribeKey - -086, -140, -328) and the settled contradiction of -057 with -141 (FU2-07: with kms:* denied,
  CreateTable `{AES256}` and no-config -> 200; -328 canonical); `table-record-expiration.md` the revoked-grant follow-up
  -239; `table.md` identity, 404 shapes, RenameTable in general and tags; `table-metadata-schema.md` `versionToken`
  (unchanged by the rename in -319) and the write-only `metadata.iceberg` that -308 shows missing from every read.
- Long catalogues render in `details/<id>.md` rather than inline: the `kmsKeyArn` form matrix ([S3T-TABLE-085](#s3t-table-085)), the nine
  `sseAlgorithm` variants (-087), the seven `storageClass` variants (-012), the `{}` shape matrix (-013), the combination
  matrix of -084, the identifier-form matrix of the sub-reads (-062), the 7-operation region-mismatch matrix (-252), the
  GetTable key set and sub-read map (-308), the rename follow-through table (-319) and the key-state series (-248).
- Open experiments (STATUS.md, Phase 8): FU2 settled "disabled KMS key -> 400 always" (-248 agrees; -037 superseded) and
  "CreateTable aws:kms needs caller kms:DescribeKey" (table-policy.md). n=1 and single-run items worth a re-probe are in
  `table-encryption-storage.open-questions.md`. The 8 KMS keys used by -242, -248 and -249 are PendingDeletion and expire
  2026-10-20/21, so any re-run of the key-state findings needs fresh keys; -030 and -161 are model-derived (botocore
  2018-05-10 / s3tables v1.16.4; botocore 1.42.97) and should be re-checked after an SDK bump.
<!-- preserved:end -->

## At a glance

- canonical findings: 31 (high 10 / medium 12 / low 9); duplicates folded into the appendix: 1
- handling: handled 10 · partial 0 · tracked 0 · unhandled 21 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: other 7, request-validation 5, server-default 5, delete-semantics 3, immutable-field 3, identity
  2, codegen-artifact 1, cross-region 1, error-code 1, eventual-consistency 1, read-gap 1, response-fidelity 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketEncryption | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableBucketStorageClass | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListNamespaces | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketEncryption | create | tableBucketARN, encryptionConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketStorageClass | create | tableBucketARN, storageClassConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

Also referenced by findings but not in the s3tables model (other services or annotated variants):
ScheduleKeyDeletion

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-table-062"></a>**S3T-TABLE-062** `identity` · impact medium · handled · verified 2026-10-13
  **GetTableEncryption/GetTableStorageClass take only the exact bucketARN+namespace+name triple: table ARN -> 400; old name after Rename -> 404**
  On the {IT + aws:kms} table: tableBucketARN=<table ARN> -> enc BadRequestException 400 / sc
  BadRequestException 400; tableBucketARN with trailing slash -> enc BadRequestException 400 / sc
  BadRequestException 400; tableBucketARN bucket name UPPERCASED -> enc BadRequestException 400 / sc
  BadRequestException 400; name UPPERCASED -> enc NotFoundException 404 / sc NotFoundException 404; namespace
  as the namespace id -> enc NotFoundException 404 / sc NotFoundException 404. RenameTable(newName) -> HTTP
  204; old triple -> enc NotFoundException 404 / sc NotFoundException 404; new triple -> enc {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"} / sc
  {"storageClass": "INTELLIGENT_TIERING"}; GetTable(tableArn=<ARN from before the rename>) -> HTTP 200,
  name=ackq_5fda45_it_kms_renamed.
  - ACK: is_arn_primary_key, custom_find, references · ops: GetTableEncryption, GetTableStorageClass,
    RenameTable, GetTable · fields: tableBucketARN, namespace, name
  - repro: GetTableEncryption/GetTableStorageClass with each identifier form, then RenameTable and re-read
    with old and new names
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-011](#s3t-table-011), [S3T-TABLE-063](#s3t-table-063), [S3T-TABLE-252](#s3t-table-252), [S3T-TABLEBUCKET-085](tablebucket-encryption-storage.md#s3t-tablebucket-085) · hypotheses: H-T-180 ·
    evidence: table/round-trip/explicit-encryption-storage
  - notes: H-T-180 confirmed (table ARN in tableBucketARN gives BadRequestException, not NotFoundException;
    rename: old NotFoundException / new 200).
  - full notes: [details/S3T-TABLE-062.md](details/S3T-TABLE-062.md)

- <a id="s3t-table-319"></a>**S3T-TABLE-319** `identity` · impact high · handled · verified 2026-10-14
  **RenameTable across namespaces (HTTP 204): tableARN unchanged, old triple -> NotFoundException 404, encryption/storage class follow**
  Table created with encryptionConfiguration={sseAlgorithm:aws:kms, kmsKeyArn:<key>},
  storageClass=INTELLIGENT_TIERING and a 2-field schema; RenameTable(namespace=ns1, name=t,
  newNamespaceName=ns2, newName=t2, no versionToken) -> HTTP 204 (empty body). Then GetTable(ns1,t) ->
  NotFoundException (HTTP 404) (poll codes right after the call: ['ERR:NotFoundException'], 404 after 0.01 s);
  GetTable(ns2,t2) -> HTTP 200 (poll codes ['OK'], 200 after 0.01 s); GetTable(tableArn=<pre-rename ARN>) ->
  HTTP 200 with namespace=['<ns2>'] name=<t2>; tableARN identical to the Create ARN in both reads: True.
  ListTables(ns1) -> [], ListTables(ns2) -> ['<t2>']. Under the new triple GetTableEncryption -> HTTP 200
  {"kmsKeyArn_eq_sent_key": true, "sseAlgorithm": "aws:kms"} (same as before rename: True),
  GetTableStorageClass -> HTTP 200 {"storageClass": "INTELLIGENT_TIERING"} (same: True),
  GetTableMetadataLocation -> HTTP 200, GetTableMaintenanceConfiguration -> HTTP 200; under the old triple the
  four sub-resource reads -> ['NotFoundException']. versionToken changed by the rename: False (identical token
  before/after); modifiedAt changed: True; modifiedBy was ABSENT from GetTable on the fresh table and PRESENT
  (== caller account) after the rename; ListTagsForResource(tableArn) after rename -> HTTP 200 with the create
  tag intact: True.
  - ACK: is_arn_primary_key, custom_update, is_immutable · ops: RenameTable, GetTable, ListTables,
    GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation, ListTagsForResource · fields:
    namespace, name, newNamespaceName, newName, tableARN, versionToken, encryptionConfiguration,
    storageClassConfiguration
  - repro: CreateTable(ICEBERG, aws:kms key, INTELLIGENT_TIERING) in ns1 -> RenameTable(newNamespaceName=ns2,
    newName=t2) -> GetTable by old triple, new triple, tableArn; GetTableEncryption/GetTableStorageClass under
    both triples; ListTables(ns1), ListTables(ns2)
  - measurements: old_triple_404_after_s=0.01, new_triple_200_after_s=0.01, n=1
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-012](table-maintenance.md#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-265](table-maintenance.md#s3t-table-265), [S3T-TABLE-318](table-maintenance.md#s3t-table-318), [S3T-TABLE-312](table-schema.md#s3t-table-312) · evidence:
    table/sub-resources/fu-07-rename-kms-table
  - notes: FU-07. A controller keyed by (bucket, namespace, name) loses the table on an out-of-band rename,
    while a controller keyed by tableARN keeps finding it (namespace/name are then the drift). RenameTable has
    no output body; the new state must be read back. CreateTable with aws:kms succeeded (key policy:...
  - full notes: [details/S3T-TABLE-319.md](details/S3T-TABLE-319.md)

## Errors

- <a id="s3t-table-248"></a>**S3T-TABLE-248** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms K) with K Disabled or PendingDeletion -> BadRequestException 400 'KMS key ARN is disabled.' (same text both)**
  K1 = SYMMETRIC_DEFAULT with root + maintenance-principal policy; CreateTable(aws:kms K1) while Enabled ->
  HTTP 200. DisableKey -> Disabled visible after 0.0 s, +5 s: CreateTable(aws:kms K1) -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is disabled.' (x-amzn-ErrorType
  BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/; second attempt 10 s later:
  same code). PutTableBucketEncryption(aws:kms K1) while Disabled -> BadRequestException (HTTP 400) 'The
  specified KMS key ARN is disabled.'. EnableKey -> Enabled after 0.0 s, +5 s: CreateTable -> HTTP 200
  (attempts 1). ScheduleKeyDeletion(7 d) -> PendingDeletion after 0.0 s, +5 s: CreateTable(aws:kms K1) ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is disabled.' (second attempt: same code). Message
  texts: Disabled='The specified KMS key ARN is disabled.'; PendingDeletion='The specified KMS key ARN is
  disabled.'.
  - ACK: terminal_codes, requeue · ops: CreateTable, PutTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateKey(symmetric, maintenance principal) -> DisableKey -> CreateTable(aws:kms K); EnableKey ->
    CreateTable; ScheduleKeyDeletion(7) -> CreateTable
  - measurements: kms_state_visible_in_describe_key_s.disabled=0.01,
    kms_state_visible_in_describe_key_s.enabled=0.02,
    kms_state_visible_in_describe_key_s.pending_deletion=0.01, create_table_latency_ms.enabled=652,
    create_table_latency_ms.disabled=[30, 30], create_table_latency_ms.pending_deletion=[18, 36],
    create_table_latency_ms.reenabled=531, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-253](table-maintenance.md#s3t-table-253), [S3T-TABLE-301](table-maintenance.md#s3t-table-301), [S3T-TABLE-254](table-maintenance.md#s3t-table-254), [S3T-TABLE-246](table-maintenance.md#s3t-table-246), [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLE-239](table-record-expiration.md#s3t-table-239),
    [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-242](#s3t-table-242) · hypotheses: H-T-167 · evidence:
    table/error-taxonomy/kms-key-states-types
  - notes: H-T-167 confirmed: BadRequestException 400 (not NotFoundException) for both Disabled and
    PendingDeletion; qualified in that the message does not embed the KMS exception name and is identical for
    both states ('The specified KMS key ARN is disabled.'), so PendingDeletion is indistinguishable from...
  - full notes: [details/S3T-TABLE-248.md](details/S3T-TABLE-248.md)

## Request validation

- <a id="s3t-table-084"></a>**S3T-TABLE-084** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable: aws:kms without kmsKeyArn or with '' -> BadRequestException 400; AES256 with a kmsKeyArn -> BadRequestException 400**
  encryptionConfiguration={aws:kms} (no kmsKeyArn) -> BadRequestException (HTTP 400) 'The specified KMS key
  ARN is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'.
  {aws:kms, kmsKeyArn=''} -> botocore rejects client-side (Parameter validation failed: Invalid length for
  parameter encryptionConfiguration.kmsKeyArn, value: 0, valid min length: 1); sent on the wire with
  validation disabled -> BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN
  format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'. {AES256, kmsKeyArn=<valid
  key K>} -> BadRequestException (HTTP 400) 'KMS key ARN detected. Please specify 'aws:kms' as the encryption
  algorithm.'. {AES256, kmsKeyArn=<well-formed non-existent key>} -> BadRequestException (HTTP 400) 'KMS key
  ARN detected. Please specify 'aws:kms' as the encryption algorithm.'.
  - ACK: terminal_codes · ops: CreateTable, GetTableEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(ICEBERG, encryptionConfiguration=<each combination>) with distinct names;
    GetTable/GetTableEncryption afterwards
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-102](tablebucket-encryption-storage.md#s3t-tablebucket-102), [S3T-TABLEBUCKET-103](tablebucket-encryption-storage.md#s3t-tablebucket-103), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-085](#s3t-table-085),
    [S3T-TABLEBUCKET-135](tablebucket-encryption-storage.md#s3t-tablebucket-135), [S3T-TABLE-258](#s3t-table-258), [S3T-TABLEBUCKET-136](tablebucket-encryption-storage.md#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](tablebucket-encryption-storage.md#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](tablebucket-encryption-storage.md#s3t-tablebucket-138),
    [S3T-TABLEBUCKET-034](tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLE-012](#s3t-table-012), [S3T-TABLE-013](#s3t-table-013), [S3T-TABLE-087](#s3t-table-087) · hypotheses: H-T-156, H-T-157 ·
    evidence: table/weird-inputs/encryption-storage-enums
  - notes: H-T-156 confirmed (no AWS-managed key substituted: true). H-T-157 confirmed (rejected). Full
    original title: CreateTable aws:kms without kmsKeyArn -> BadRequestException 400; kmsKeyArn='' ->
    BadRequestException 400; AES256 + kmsKeyArn -> BadRequestException 400

- <a id="s3t-table-085"></a>**S3T-TABLE-085** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable kmsKeyArn: only the exact full key ARN accepted; bare id, alias name/ARN, uppercased or trailing-space ARN -> 400**
  bare key id -> BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; alias name -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; alias ARN -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; full key ARN -> HTTP 200,
  GetTableEncryption {"sseAlgorithm": "aws:kms", "kmsKeyArn": "<K ARN>"}; full key ARN + trailing space ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN does not exist.'; full key ARN uppercased ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; key ARN with 'kms' service
  uppercased -> BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'.
  - ACK: terminal_codes, references · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms, kmsKeyArn=<form>) for each form, distinct names; GetTableEncryption on success
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-102](tablebucket-encryption-storage.md#s3t-tablebucket-102), [S3T-TABLEBUCKET-103](tablebucket-encryption-storage.md#s3t-tablebucket-103), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-084](#s3t-table-084),
    [S3T-TABLEBUCKET-105](tablebucket-encryption-storage.md#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](tablebucket-encryption-storage.md#s3t-tablebucket-128), [S3T-TABLEBUCKET-107](tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLEBUCKET-131](tablebucket.md#s3t-tablebucket-131), [S3T-TABLEBUCKET-135](tablebucket-encryption-storage.md#s3t-tablebucket-135),
    [S3T-TABLE-258](#s3t-table-258), [S3T-TABLEBUCKET-136](tablebucket-encryption-storage.md#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](tablebucket-encryption-storage.md#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](tablebucket-encryption-storage.md#s3t-tablebucket-138), [S3T-TABLEBUCKET-034](tablebucket-encryption-storage.md#s3t-tablebucket-034),
    [S3T-TABLE-012](#s3t-table-012), [S3T-TABLE-013](#s3t-table-013), [S3T-TABLE-087](#s3t-table-087) · hypotheses: H-T-158, H-T-159 · evidence:
    table/weird-inputs/encryption-storage-enums
  - notes: H-T-158 confirmed (only the full key ARN is accepted). H-T-159 (alias ARN accepted and resolved to
    the key ARN) refuted (alias ARN rejected).
  - full notes: [details/S3T-TABLE-085.md](details/S3T-TABLE-085.md)

- <a id="s3t-table-087"></a>**S3T-TABLE-087** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable sseAlgorithm is case-sensitive: 'aes256', 'Aes256', 'AES-256', 'Aws:Kms', 'AWS:KMS', 'aws-kms', 'kms', 'aws:kms:dsse' -> 400**
  'aes256' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'Aes256' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'AES-256' -> BadRequestException
  (HTTP 400) 'The specified SSE algorithm is invalid.'; 'Aws:Kms' -> BadRequestException (HTTP 400) 'The
  specified SSE algorithm is invalid.'; 'AWS:KMS' -> BadRequestException (HTTP 400) 'The specified SSE
  algorithm is invalid.'; 'aws-kms' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is
  invalid.'; 'aws:kms ' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'kms' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'aws:kms:dsse' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'. Distinct messages: ['The specified
  SSE algorithm is invalid.']. (Exact 'AES256' and 'aws:kms' succeed in the sibling probes.)
  - ACK: terminal_codes · ops: CreateTable · fields: encryptionConfiguration.sseAlgorithm
  - repro: CreateTable(encryptionConfiguration={sseAlgorithm: <variant>[, kmsKeyArn: K]}) per variant,
    distinct names
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-024](tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-013](#s3t-table-013), [S3T-TABLE-012](#s3t-table-012),
    [S3T-TABLE-084](#s3t-table-084), [S3T-TABLE-085](#s3t-table-085) · hypotheses: H-T-181 · evidence: table/weird-inputs/encryption-storage-enums
  - notes: H-T-181 confirmed. The message does not name the enum values.
  - full notes: [details/S3T-TABLE-087.md](details/S3T-TABLE-087.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-010"></a>**S3T-TABLE-010** `server-default` · impact high · handled · verified 2026-10-15, re-verified
  **GetTableEncryption on a table created without encryptionConfiguration (bucket has none) -> HTTP 200 sseAlgorithm=AES256, kmsKeyArn absent**
  CreateTable(format=ICEBERG, no encryptionConfiguration) in a table bucket with no PutTableBucketEncryption.
  GetTableEncryption called immediately after CreateTable returned -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "AES256"}}; again after other reads -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "AES256"}}. GetTableBucketEncryption on the same bucket -> HTTP 200
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}. GetTable on the same triple -> HTTP 200.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: CreateTable, GetTableEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket -> CreateNamespace -> CreateTable(ICEBERG, no encryptionConfiguration) ->
    GetTableEncryption
  - measurements: get_table_encryption_latency_ms=[36, 7], n=2
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLE-255](#s3t-table-255), [S3T-TABLE-259](#s3t-table-259), [S3T-TABLE-257](#s3t-table-257), [S3T-TABLE-057](#s3t-table-057), [S3T-TABLEBUCKET-120](tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLE-011](#s3t-table-011),
    [S3T-TABLE-059](#s3t-table-059), [S3T-TABLE-061](#s3t-table-061), [S3T-TABLEBUCKET-005](tablebucket-encryption-storage.md#s3t-tablebucket-005), [S3T-TABLEBUCKET-083](tablebucket-encryption-storage.md#s3t-tablebucket-083) · hypotheses: H-T-150, H-T-151 ·
    evidence: table/round-trip/encryption-storage-defaults
  - notes: H-T-150 confirmed; H-T-151 (NotFoundException from the sub-read) refuted. kmsKeyArn member in the
    response: absent.

- <a id="s3t-table-011"></a>**S3T-TABLE-011** `server-default` · impact high · handled · verified 2026-10-13
  **GetTableStorageClass on a table created without storageClassConfiguration (bucket has none) -> HTTP 200 storageClass=STANDARD**
  CreateTable(format=ICEBERG, no storageClassConfiguration) in a bucket with no PutTableBucketStorageClass.
  GetTableStorageClass immediately after CreateTable -> HTTP 200 {"storageClassConfiguration":
  {"storageClass": "STANDARD"}}; later -> HTTP 200 {"storageClassConfiguration": {"storageClass":
  "STANDARD"}}. GetTableBucketStorageClass on the bucket -> HTTP 200 {"storageClassConfiguration":
  {"storageClass": "STANDARD"}}. Ghost table name: GetTable -> NotFoundException (HTTP 404) 'The specified
  table does not exist.'; GetTableEncryption -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTableStorageClass -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare, exceptions.404 · ops: CreateTable,
    GetTableStorageClass, GetTableBucketStorageClass, GetTable, GetTableEncryption · fields:
    storageClassConfiguration.storageClass
  - repro: CreateTable(ICEBERG, no storageClassConfiguration) -> GetTableStorageClass; same sub-reads on a
    non-existent table name
  - measurements: get_table_storage_class_latency_ms=[7, 6], n=2
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLE-010](#s3t-table-010), [S3T-TABLE-059](#s3t-table-059), [S3T-TABLE-061](#s3t-table-061), [S3T-TABLEBUCKET-005](tablebucket-encryption-storage.md#s3t-tablebucket-005), [S3T-TABLEBUCKET-083](tablebucket-encryption-storage.md#s3t-tablebucket-083),
    [S3T-TABLE-062](#s3t-table-062), [S3T-TABLE-063](#s3t-table-063), [S3T-TABLE-252](#s3t-table-252) · hypotheses: H-T-172, H-T-173 · evidence:
    table/round-trip/encryption-storage-defaults
  - notes: H-T-172 confirmed; H-T-173 refuted. The three not-found messages for a ghost table name are
    identical.

- <a id="s3t-table-020"></a>**S3T-TABLE-020** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default INTELLIGENT_TIERING: CreateTable without storageClassConfiguration inherits it; explicit STANDARD reads STANDARD**
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200 (GetTableBucketStorageClass showed it after 0.01
  s; fresh bucket read before the put: HTTP 200 {"storageClassConfiguration": {"storageClass": "STANDARD"}};
  put with storageClass=GLACIER -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'storageClassConfiguration.storageClass' failed to satisfy constraint: Member must satisfy enum value set:
  [INTELLIGENT_TIERING, STANDARD]'). Then CreateTable(no storageClassConfiguration) -> HTTP 200,
  GetTableStorageClass -> INTELLIGENT_TIERING; CreateTable(explicit INTELLIGENT_TIERING) -> HTTP 200, read
  INTELLIGENT_TIERING; CreateTable(explicit STANDARD) -> HTTP 200, read STANDARD.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketStorageClass,
    GetTableBucketStorageClass, CreateTable, GetTableStorageClass · fields:
    storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass(IT) -> CreateTable(no config), CreateTable(explicit IT),
    CreateTable(explicit STANDARD) -> GetTableStorageClass on each
  - measurements: bucket_storage_class_visible_after_put_s=[0.01], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-087](tablebucket-encryption-storage.md#s3t-tablebucket-087), [S3T-TABLEBUCKET-024](tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-012](#s3t-table-012), [S3T-TABLE-013](#s3t-table-013),
    [S3T-TABLEBUCKET-145](tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](#s3t-table-260), [S3T-TABLE-261](#s3t-table-261), [S3T-TABLE-262](#s3t-table-262), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLE-021](#s3t-table-021),
    [S3T-TABLE-266](#s3t-table-266), [S3T-TABLE-161](#s3t-table-161), [S3T-TABLEBUCKET-072](tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-073](tablebucket-encryption-storage.md#s3t-tablebucket-073) · hypotheses: H-T-174, H-T-177 ·
    evidence: table/sub-resources/storage-class-inheritance
  - notes: H-T-174 first half (inheritance at create) confirmed; H-T-177 (explicit STANDARD under an IT bucket
    default) confirmed. No DeleteTableBucketStorageClass operation exists in the API, so 'deleting the bucket
    config' cannot be tested.
  - full notes: [details/S3T-TABLE-020.md](details/S3T-TABLE-020.md)

- <a id="s3t-table-021"></a>**S3T-TABLE-021** `immutable-field` · impact high · handled · verified 2026-10-15, re-verified
  **Table storage class is a creation-time snapshot: inheriting table still INTELLIGENT_TIERING 66 polls/120 s after bucket flips to STANDARD**
  With tables t1 (no config, read INTELLIGENT_TIERING at create), t3 (explicit IT), t4 (explicit STANDARD)
  under a bucket default of INTELLIGENT_TIERING, PutTableBucketStorageClass(STANDARD) -> HTTP 200 (bucket read
  showed STANDARD after 0.01 s). GetTableStorageClass polled: t1 every 1 s for 60 s then every 10 s to 120 s
  -> values ['INTELLIGENT_TIERING']; t3 -> ['INTELLIGENT_TIERING']; t4 -> ['STANDARD']. A table created after
  the flip without config (t2) reads STANDARD.
  - ACK: compare.is_ignored+delta_pre_compare, is_immutable, late_initialize · ops:
    PutTableBucketStorageClass, GetTableStorageClass, CreateTable · fields:
    storageClassConfiguration.storageClass
  - repro: Bucket IT -> CreateTable t1 (no config) -> PutTableBucketStorageClass(STANDARD) -> poll
    GetTableStorageClass(t1) for 120 s -> CreateTable t2 (no config)
  - measurements: poll_window_s=120, t1_polls=66, bucket_storage_class_visible_after_put_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLEBUCKET-145](tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](#s3t-table-260), [S3T-TABLE-261](#s3t-table-261), [S3T-TABLE-262](#s3t-table-262), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084),
    [S3T-TABLE-020](#s3t-table-020), [S3T-TABLE-057](#s3t-table-057), [S3T-TABLE-058](#s3t-table-058) · hypotheses: H-T-174, H-T-175 · evidence:
    table/sub-resources/storage-class-inheritance
  - notes: H-T-174 second half (snapshot >= 2 min) confirmed; H-T-175 (live tracking within 60 s) refuted. The
    bucket default is read live by GetTableBucketStorageClass, the table read is not.
  - full notes: [details/S3T-TABLE-021.md](details/S3T-TABLE-021.md)

- <a id="s3t-table-057"></a>**S3T-TABLE-057** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default aws:kms(K): CreateTable without encryptionConfiguration reads aws:kms with kmsKeyArn == K byte-for-byte...**
  Key K grants the maintenance principal (no condition). PutTableBucketEncryption(aws:kms, K) -> HTTP 200;
  GetTableBucketEncryption -> {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}. CreateTable(no
  encryptionConfiguration) -> HTTP 200; GetTableEncryption immediately -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}} (kmsKeyArn equals the key ARN
  returned by CreateKey). CreateTable(encryptionConfiguration={AES256}) under the same bucket -> HTTP 200;
  GetTableEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketEncryption, CreateTable,
    GetTableEncryption · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: PutTableBucketEncryption(aws:kms, K with maintenance grant) -> CreateTable(no config) and
    CreateTable({AES256}) -> GetTableEncryption on both
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-255](#s3t-table-255), [S3T-TABLE-259](#s3t-table-259), [S3T-TABLE-257](#s3t-table-257), [S3T-TABLEBUCKET-120](tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLE-010](#s3t-table-010), [S3T-TABLE-054](table-maintenance.md#s3t-table-054),
    [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLE-056](table-maintenance.md#s3t-table-056), [S3T-TABLE-141](table-policy.md#s3t-table-141), [S3T-TABLE-021](#s3t-table-021), [S3T-TABLE-058](#s3t-table-058), [S3T-TABLE-328](table-policy.md#s3t-table-328) · hypotheses:
    H-T-152, H-T-162 · evidence: table/dependencies/kms-maintenance-principal
  - notes: H-T-152 confirmed; H-T-162 confirmed.
  - full notes: [details/S3T-TABLE-057.md](details/S3T-TABLE-057.md)

- <a id="s3t-table-058"></a>**S3T-TABLE-058** `immutable-field` · impact high · handled · verified 2026-10-13
  **Table encryption is a creation-time snapshot: inheriting KMS table unchanged 66 polls/120 s after DeleteTableBucketEncryption...**
  t_inherit (created without config under bucket default aws:kms K) read
  'aws:kms|arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a'.
  DeleteTableBucketEncryption -> HTTP 204; GetTableBucketEncryption right after -> HTTP 200
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}. GetTableEncryption(t_inherit) polled 1 s x 60 then
  10 s to 120 s: values ['aws:kms|arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a'].
  A no-config table created after the delete reads {"sseAlgorithm": "AES256"}.
  PutTableBucketEncryption(aws:kms K) again -> HTTP 200; that AES256 no-config table polled 1 s x 60:
  ['AES256|<absent>']; the explicit-AES256 table sampled every 10 s: ['AES256|<absent>'].
  - ACK: is_immutable, compare.is_ignored+delta_pre_compare, late_initialize · ops:
    DeleteTableBucketEncryption, PutTableBucketEncryption, GetTableEncryption, GetTableBucketEncryption ·
    fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: Bucket default aws:kms K -> CreateTable(no config) -> DeleteTableBucketEncryption -> poll
    GetTableEncryption 120 s -> CreateTable(no config) -> PutTableBucketEncryption(K) -> poll 60 s
  - measurements: poll_window_after_delete_s=120, poll_window_after_put_s=60, polls_inherit=66,
    polls_plain=60, n=126
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLE-021](#s3t-table-021), [S3T-TABLE-057](#s3t-table-057) · hypotheses: H-T-153, H-T-154 · evidence:
    table/dependencies/kms-maintenance-principal
  - notes: H-T-153 (snapshot) confirmed; H-T-154 (live inheritance within seconds) refuted. Only one KMS key
    is available in this shard, so the 'switch to K2' cell was replaced by delete-default and re-put-default.
  - full notes: [details/S3T-TABLE-058.md](details/S3T-TABLE-058.md)

- <a id="s3t-table-161"></a>**S3T-TABLE-161** `immutable-field` · impact medium · handled · verified 2026-10-13
  **No PutTableStorageClass in the model (only Get*/PutTableBucketStorageClass): a table's storage class is fixed at CreateTable**
  The SDK model (botocore 1.42.97) lists only ['GetTableBucketStorageClass', 'GetTableStorageClass',
  'PutTableBucketStorageClass']; CreateTable accepts storageClassConfiguration and there is no UpdateTable. In
  this probe tables t1 (created before Put INTELLIGENT_TIERING) and t2 (created after) kept their class across
  two bucket-default changes (STANDARD / INTELLIGENT_TIERING).
  - ACK: is_immutable, docs-only · ops: CreateTable, GetTableStorageClass · fields:
    storageClassConfiguration.storageClass
  - repro: Inspect the service model; change the bucket default twice and re-read the storage class of
    existing tables
  - measurements: n=2
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLEBUCKET-009](tablebucket-encryption-storage.md#s3t-tablebucket-009), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-083](tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLEBUCKET-085](tablebucket-encryption-storage.md#s3t-tablebucket-085),
    [S3T-TABLE-020](#s3t-table-020), [S3T-TABLE-266](#s3t-table-266), [S3T-TABLEBUCKET-072](tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-073](tablebucket-encryption-storage.md#s3t-tablebucket-073) · hypotheses: H-S-042 · evidence:
    tablebucket/sub-resources/storage-class-default-transitions
  - notes: H-S-042 confirmed (operation absent in the model; existing tables did not change in the
    experiment). Full original title: No PutTableStorageClass in the s3tables model
    (GetTableBucketStorageClass, GetTableStorageClass, PutTableBucketStorageClass): a table's storage class is
    fixed at CreateTable

- <a id="s3t-table-266"></a>**S3T-TABLE-266** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Explicit CreateTable storageClass=STANDARD wins over the bucket default INTELLIGENT_TIERING (reads STANDARD); omitted inherits the default**
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200.
  CreateTable(storageClassConfiguration={storageClass: STANDARD}) -> 200; GetTableStorageClass ->
  {"storageClass": "STANDARD"}. CreateTable without storageClassConfiguration -> GetTableStorageClass ->
  {"storageClass": "INTELLIGENT_TIERING"}. After PutTableBucketStorageClass(STANDARD): CreateTable(explicit
  INTELLIGENT_TIERING) -> {"storageClass": "INTELLIGENT_TIERING"}.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketStorageClass, CreateTable,
    GetTableStorageClass · fields: storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass(INTELLIGENT_TIERING) -> CreateTable(storageClassConfiguration STANDARD)
    -> GetTableStorageClass
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-020](#s3t-table-020), [S3T-TABLE-059](#s3t-table-059), [S3T-TABLE-161](#s3t-table-161), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-072](tablebucket-encryption-storage.md#s3t-tablebucket-072),
    [S3T-TABLEBUCKET-073](tablebucket-encryption-storage.md#s3t-tablebucket-073) · evidence: table/creative/bucket-maintenance-propagation
  - notes: Cross-service analogy (DynamoDB TableClass default / ElastiCache parameter defaults): whether an
    explicit per-table value is honored when it equals the service default but differs from the parent
    default. Extends [S3T-TABLE-020](#s3t-table-020) and [S3T-TABLE-059](#s3t-table-059) (explicit values on a bucket WITHOUT a conflicting...
  - full notes: [details/S3T-TABLE-266.md](details/S3T-TABLE-266.md)

- <a id="s3t-table-308"></a>**S3T-TABLE-308** `read-gap` · impact high · handled · verified 2026-10-14
  **GetTable (by triple or tableArn) omits encryptionConfiguration, storageClassConfiguration, metadata/schema and tags; each needs its own read**
  CreateTable(ICEBERG, metadata.iceberg.schema with 2 fields, encryptionConfiguration={aws:kms,K},
  storageClassConfiguration=INTELLIGENT_TIERING, 3 tags) -> HTTP 200, body keys ['tableARN', 'versionToken'].
  GetTable by (tableBucketARN, namespace, name) -> HTTP 200 with exactly the raw JSON keys ['createdAt',
  'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name', 'namespace', 'namespaceId',
  'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']; GetTable by
  tableArn -> HTTP 200 with keys ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt',
  'name', 'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] (identical key set: True; differing values: []). None of ['encryptionConfiguration',
  'storageClassConfiguration', 'metadata', 'schema', 'tags'] appears in either body nor in the ListTables
  entry (keys ['createdAt', 'modifiedAt', 'name', 'namespace', 'tableARN', 'type']). The inputs are readable
  only via: GetTableEncryption -> HTTP 200 {"kmsKeyArn": "<the key ARN sent>", "sseAlgorithm": "aws:kms"};
  GetTableStorageClass -> HTTP 200 {"storageClass": "INTELLIGENT_TIERING"}; ListTagsForResource(tableArn) ->
  HTTP 200 same 3 keys/values as sent; GetTableMetadataLocation -> HTTP 200 keys ['metadataLocation',
  'versionToken', 'warehouseLocation'] (metadataLocation present
  (s3://<table-bucket-uuid>--table-s3/metadata/<file>.metadata.json)); the schema itself is never returned by
  any API (botocore model 2018-05-10: operations whose OUTPUT carries a
  schema/schemaV2/partitionSpec/sortOrder/writeOrder/properties/fields member: none; INPUTs carrying them:
  ['CreateTable']).
  - ACK: hook:sdk_read_one_post_set_output, custom_find, tags.custom-sync, is_immutable · ops: CreateTable,
    GetTable, ListTables, GetTableEncryption, GetTableStorageClass, ListTagsForResource,
    GetTableMetadataLocation · fields: metadata, metadata.iceberg.schema, encryptionConfiguration,
    storageClassConfiguration, tags, metadataLocation
  - repro: CreateTable(format=ICEBERG, metadata.iceberg.schema, encryptionConfiguration={aws:kms,K},
    storageClassConfiguration, tags) -> GetTable(triple), GetTable(tableArn), ListTables, GetTableEncryption,
    GetTableStorageClass, ListTagsForResource(tableArn), GetTableMetadataLocation
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181`
  - related: [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-010](tablebucket-encryption-storage.md#s3t-tablebucket-010), [S3T-TABLE-060](#s3t-table-060) · evidence:
    table/round-trip/fu-06-full-create-read-paths
  - notes: FU-06. Single table, one run (the key set is deterministic from the model: GetTable output members
    are ['name', 'type', 'tableARN', 'namespace', 'namespaceId', 'versionToken', 'metadataLocation',
    'warehouseLocation', 'createdAt', 'createdBy', 'managedByService', 'modifiedAt', 'modifiedBy',...
  - full notes: [details/S3T-TABLE-308.md](details/S3T-TABLE-308.md)

## Response fidelity and consistency

- <a id="s3t-table-059"></a>**S3T-TABLE-059** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **Explicit storageClassConfiguration and encryptionConfiguration at CreateTable round-trip byte-for-byte through the two sub-reads (4/4)**
  Fresh bucket without bucket-level defaults. {"storageClassConfiguration": {"storageClass":
  "INTELLIGENT_TIERING"}} -> HTTP 200, GetTableEncryption {"sseAlgorithm": "AES256"}, GetTableStorageClass
  {"storageClass": "INTELLIGENT_TIERING"}; {"storageClassConfiguration": {"storageClass": "STANDARD"}} -> HTTP
  200, GetTableEncryption {"sseAlgorithm": "AES256"}, GetTableStorageClass {"storageClass": "STANDARD"};
  {"encryptionConfiguration": {"kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a", "sseAlgorithm": "aws:kms"},
  "storageClassConfiguration": {"storageClass": "INTELLIGENT_TIERING"}} -> HTTP 200, GetTableEncryption
  {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}, GetTableStorageClass
  {"storageClass": "INTELLIGENT_TIERING"}; {"encryptionConfiguration": {"sseAlgorithm": "AES256"},
  "storageClassConfiguration": {"storageClass": "STANDARD"}} -> HTTP 200, GetTableEncryption {"sseAlgorithm":
  "AES256"}, GetTableStorageClass {"storageClass": "STANDARD"}. Unset encryption reads AES256 and unset
  storage class reads STANDARD (server defaults).
  - ACK: compare.is_ignored+delta_pre_compare, late_initialize · ops: CreateTable, GetTableEncryption,
    GetTableStorageClass · fields: storageClassConfiguration.storageClass,
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTable with {IT}, {STANDARD}, {IT+aws:kms K}, {STANDARD+AES256} -> GetTableEncryption +
    GetTableStorageClass on each
  - measurements: create_table_latency_ms=[396, 311, 447, 343], n=4
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLEBUCKET-050](tablebucket-encryption-storage.md#s3t-tablebucket-050), [S3T-TABLEBUCKET-052](tablebucket-encryption-storage.md#s3t-tablebucket-052), [S3T-TABLEBUCKET-083](tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLEBUCKET-103](tablebucket-encryption-storage.md#s3t-tablebucket-103),
    [S3T-TABLE-010](#s3t-table-010), [S3T-TABLE-011](#s3t-table-011), [S3T-TABLE-061](#s3t-table-061) · hypotheses: H-T-176 · evidence:
    table/round-trip/explicit-encryption-storage
  - notes: H-T-176 confirmed. Full original title: Explicit storageClassConfiguration and
    encryptionConfiguration at CreateTable round-trip byte-for-byte through the two sub-reads (4/4
    combinations)

- <a id="s3t-table-061"></a>**S3T-TABLE-061** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableEncryption/GetTableStorageClass read-after-write consistent with CreateTable: 30/30 polls at +0/+1/+2 s on 10 {IT+aws:kms} tables**
  For each of 10 tables (CreateTable {INTELLIGENT_TIERING + aws:kms K}, 10 succeeded) both sub-reads were
  called immediately after CreateTable returned and again at +1 s and +2 s. Every poll returned HTTP 200 with
  exactly the requested values; no NotFoundException and no default-valued (AES256/STANDARD) response.
  CreateTable latency 427-501 ms (KMS pre-check included).
  - ACK: none · ops: CreateTable, GetTableEncryption, GetTableStorageClass · fields: encryptionConfiguration,
    storageClassConfiguration
  - repro: 10x CreateTable({IT + aws:kms K}) -> GetTableEncryption + GetTableStorageClass at +0, +1, +2 s
  - measurements: create_table_kms_latency_ms=[430, 475, 439, 493, 483, 476, 438, 501, 427, 442],
    sub_read_latency_ms=[29, 13, 8, 6, 9, 5, 23, 6, 9, 5, 8, 6, 32, 6, 8, 5, 8, 5, 39, 6, 8, 6, 9, 10, 33, 6,
    8, 6, 9, 6, 26, 8, 9, 6, 8, 5, 29, 6, 8, 6, 8, 6, 27, 6, 9, 6, 9, 6, 29, 6, 10, 7, 9, 7, 30, 6, 9, 6, 8,
    5], polls=30, n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-123](tablebucket-encryption-storage.md#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](tablebucket-encryption-storage.md#s3t-tablebucket-073), [S3T-TABLEBUCKET-008](tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](tablebucket-encryption-storage.md#s3t-tablebucket-009),
    [S3T-TABLEBUCKET-037](tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLE-010](#s3t-table-010), [S3T-TABLE-011](#s3t-table-011), [S3T-TABLE-059](#s3t-table-059) · hypotheses: H-T-186 · evidence:
    table/round-trip/explicit-encryption-storage
  - notes: H-T-186 confirmed (polling at 1 s rather than 100 ms to respect the shard's sub-read rate limit).
    Full original title: GetTableEncryption/GetTableStorageClass are read-after-write consistent with
    CreateTable: 30/30 immediate+1s+2s polls on 10 fresh {IT+aws:kms} tables returned the final config

## Delete semantics

- <a id="s3t-table-063"></a>**S3T-TABLE-063** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable (204) GetTable, GetTableEncryption, GetTableStorageClass -> NotFoundException at once; 0 stale hits in 33 polls over 30 s**
  Table {IT + aws:kms K} (pre-delete reads {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"} / {"storageClass":
  "INTELLIGENT_TIERING"}). DeleteTable -> HTTP 204. Polled right after and then round-robin 1/s for 30 s:
  GetTable -> ['NotFoundException:The specified table does not exist.']; GetTableEncryption ->
  ['NotFoundException:The specified table does not exist.']; GetTableStorageClass -> ['NotFoundException:The
  specified table does not exist.']. CreateTable with the same name and {STANDARD + AES256} -> HTTP 200;
  GetTableEncryption -> {"sseAlgorithm": "AES256"}; GetTableStorageClass -> {"storageClass": "STANDARD"} (no
  cached old values).
  - ACK: exceptions.404 · ops: DeleteTable, GetTable, GetTableEncryption, GetTableStorageClass, CreateTable ·
    fields: encryptionConfiguration, storageClassConfiguration
  - repro: CreateTable({IT + aws:kms}) -> DeleteTable -> poll the three reads 30 s -> CreateTable same name
    {STANDARD + AES256} -> re-read
  - measurements: post_delete_polls=33, stale_hits=0, window_s=30, delete_table_latency_ms=[130], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-011](#s3t-table-011), [S3T-TABLE-062](#s3t-table-062), [S3T-TABLE-252](#s3t-table-252), [S3T-TABLE-249](#s3t-table-249) · hypotheses: H-T-179 · evidence:
    table/round-trip/explicit-encryption-storage
  - notes: H-T-179 confirmed: the three 404 messages are identical ('The specified table does not exist.').
  - full notes: [details/S3T-TABLE-063.md](details/S3T-TABLE-063.md)

- <a id="s3t-table-242"></a>**S3T-TABLE-242** `delete-semantics` · impact high · handled · verified 2026-10-14
  **KMS key PendingDeletion (after Disabled): DeleteTable, DeleteNamespace, DeleteTableBucket -> 204; deletion does not need a usable key (n=1)**
  With the bucket's KMS key in state 'PendingDeletion' (ScheduleKeyDeletion 7 days, after the maintenance
  grant had been revoked and the key disabled): DeleteTable -> HTTP 204; GetTable afterwards ->
  NotFoundException 404; DeleteNamespace -> HTTP 204; DeleteTableBucket -> HTTP 204. Final key state
  'PendingDeletion'.
  - ACK: deletable.when, pre-delete-cleanup, terminal_codes · ops: DeleteTable, DeleteNamespace,
    DeleteTableBucket, ScheduleKeyDeletion · fields: encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS bucket + table -> DisableKey -> ScheduleKeyDeletion(7) -> DeleteTable -> DeleteNamespace ->
    DeleteTableBucket
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-235](table-maintenance.md#s3t-table-235), [S3T-TABLE-236](table-maintenance.md#s3t-table-236), [S3T-TABLE-237](table-maintenance.md#s3t-table-237), [S3T-TABLE-238](table-maintenance.md#s3t-table-238),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](table-maintenance.md#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-248](#s3t-table-248) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: Deletion-window behavior for a controller's finalizer path: a table/bucket whose KMS key is
    disabled or pending deletion can be deleted without restoring the key; GetTable afterwards -> 404. The key
    itself stays PendingDeletion for 7 days (lab leftover, registered). n=1 table + 1 extra table created...
  - full notes: [details/S3T-TABLE-242.md](details/S3T-TABLE-242.md)

- <a id="s3t-table-249"></a>**S3T-TABLE-249** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **DeleteTable on an SSE-KMS table -> HTTP 204 with the key Disabled and PendingDeletion; GetTableEncryption still returns the key ARN**
  t1 created with aws:kms K1 while Enabled. With K1 Disabled: GetTableEncryption(t1) -> {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/6d730d5b-b3f2-41da-816f-c4549508aca6"};
  GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; DeleteTable -> HTTP 204; GetTable afterwards ->
  NotFoundException 404. t3 created after re-enabling; with K1 PendingDeletion: GetTableEncryption(t3) ->
  {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6d730d5b-b3f2-41da-816f-c4549508aca6"}; GetTable -> HTTP 200;
  DeleteTable -> HTTP 204; GetTable afterwards -> NotFoundException 404.
  - ACK: none · ops: DeleteTable, GetTableEncryption, GetTable, GetTableMetadataLocation · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms K) -> DisableKey(K) -> GetTableEncryption, GetTable, DeleteTable; second table
    -> ScheduleKeyDeletion(K) -> GetTableEncryption, DeleteTable
  - measurements: delete_table_latency_ms=[166, 147], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-037](tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLE-063](#s3t-table-063) · hypotheses: H-T-184 · evidence:
    table/error-taxonomy/kms-key-states-types
  - notes: H-T-184 confirmed: DeleteTable -> HTTP 204 with the key Disabled (t1) and with the key
    PendingDeletion (t3); GetTable, GetTableMetadataLocation and GetTableEncryption all still succeed and
    GetTableEncryption returns the dead key's ARN. No finalizer-stuck scenario from a dead key at the
    control-plane...
  - full notes: [details/S3T-TABLE-249.md](details/S3T-TABLE-249.md)

## Cross-region

- <a id="s3t-table-252"></a>**S3T-TABLE-252** `cross-region` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **ARN of another region on 7 ops (GetTable, GetTableBucket, GetTableEncryption, TagResource, DeleteTable, ...) -> 400 'Region does not match'**
  Table created in us-east-1. On the us-west-2 endpoint every call carrying the us-east-1 ARN fails with
  BadRequestException HTTP 400 'The provided Region does not match the Region of the resource.'
  (x-amzn-ErrorType BadRequestException:...s3tablebuckets/): GetTable(tableArn),
  GetTable(tableBucketARN+namespace+name), GetTableBucket, ListNamespaces, GetTableEncryption,
  ListTagsForResource(table and bucket ARN), TagResource, DeleteTable - no cross-region 200 and nothing
  changed at home (table exists, tags unchanged). The check is ARN-region-vs-endpoint, not existence: on the
  HOME endpoint (us-east-1) the same ARN with the region segment rewritten to us-west-2 or to 'xx-fake-1' also
  -> 400 'The provided Region does not match...'. The region-rewritten ARN sent to the endpoint that matches
  it (us-west-2, where no such table exists) -> NotFoundException 404 'The specified table does not exist.'
  (by tableArn and by bucketARN+names). A tableArn whose table id is not a UUID -> BadRequestException 400
  'The specified table ARN is not valid.'. ListTableBuckets(prefix=<us-east-1 bucket name>) on us-west-2 ->
  200 with 0 buckets (listing is regional).
  - ACK: is_arn_primary_key, terminal_codes · ops: GetTable, GetTableBucket, ListTagsForResource, TagResource,
    DeleteTable, GetTableEncryption, ListNamespaces · fields: tableArn, tableBucketARN, resourceArn
  - repro: CreateTableBucket/Namespace/Table in us-east-1 -> GetTable(tableArn) and the other reads on the
    us-west-2 endpoint; rewrite the ARN region segment and repeat
  - measurements: get_table_foreign_latency_ms=[31, 21], get_table_home_latency_ms=97, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-011](#s3t-table-011), [S3T-TABLE-062](#s3t-table-062), [S3T-TABLE-063](#s3t-table-063) · hypotheses: H-T-007 · evidence:
    table/identity/foreign-region-arn-lookup
  - notes: H-T-007 confirmed: BadRequestException 400 whose message names the Region, not NotFoundException
    and not a cross-region success; the same code/message applies to GetTableBucket, ListNamespaces,
    GetTableEncryption, ListTagsForResource, TagResource and DeleteTable. A controller resolving by ARN
    must...
  - full notes: [details/S3T-TABLE-252.md](details/S3T-TABLE-252.md)

## Codegen notes

- <a id="s3t-table-030"></a>**S3T-TABLE-030** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Model: GetTable input 'tableArn' vs output 'tableARN'; DeleteTable requires bucket ARN+namespace+name; no UpdateTable operation**
  botocore s3tables model (api version 2018-05-10): GetTable input members ['tableBucketARN', 'namespace',
  'name', 'tableArn'], output ARN member ['tableARN']; DeleteTable input members ['tableBucketARN',
  'namespace', 'name', 'versionToken'] with required ['tableBucketARN', 'namespace', 'name'] (no tableArn
  form); 'UpdateTable' in operation names: False; table-level mutation operations present:
  ['PutTableMaintenanceConfiguration', 'PutTablePolicy', 'PutTableRecordExpirationConfiguration',
  'PutTableReplication', 'RenameTable', 'UpdateTableMetadataLocation']. Verified live: GetTable(tableArn=...)
  -> HTTP 200.
  - ACK: renames, is_immutable, custom_delete · ops: GetTable, DeleteTable, CreateTable · fields: tableArn,
    tableARN, format, metadata, encryptionConfiguration, storageClassConfiguration
  - repro: Inspect botocore service model for s3tables; GetTable(tableArn=<arn>)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-173](table-replication.md#s3t-table-173), [S3T-TABLE-175](table-replication.md#s3t-table-175) · hypotheses: H-T-018, H-T-048 · evidence:
    table/identity/get-table-identifier-forms
  - notes: H-T-018 confirmed from the model (TableArn vs TableARN will be distinct generated Go fields;
    ReadOne must copy status ARN into tableArn or use the composite key). H-T-048 confirmed: DeleteTable has
    no tableArn member, so bucket ARN, namespace and name must stay resolvable at delete time...
  - full notes: [details/S3T-TABLE-030.md](details/S3T-TABLE-030.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-010](#s3t-table-010) | GetTableEncryption on a table created without encryptionConfiguration (bucket has none) -> HTTP 200 sseAlgorithm=AES256, kmsKeyArn absent | get_table_encryption_latency_ms=[36, 7] | 2 |
| [S3T-TABLE-011](#s3t-table-011) | GetTableStorageClass on a table created without storageClassConfiguration (bucket has none) -> HTTP 200 storageClass=STANDARD | get_table_storage_class_latency_ms=[7, 6] | 2 |
| [S3T-TABLE-020](#s3t-table-020) | Bucket default INTELLIGENT_TIERING: CreateTable without storageClassConfiguration inherits it; explicit STANDARD reads STANDARD | bucket_storage_class_visible_after_put_s=[0.01] | 1 |
| [S3T-TABLE-021](#s3t-table-021) | Table storage class is a creation-time snapshot: inheriting table still INTELLIGENT_TIERING 66 polls/120 s after bucket flips to STANDARD | poll_window_s=120, t1_polls=66, bucket_storage_class_visible_after_put_s=[0.01, 0.01] | 2 |
| [S3T-TABLE-058](#s3t-table-058) | Table encryption is a creation-time snapshot: inheriting KMS table unchanged 66 polls/120 s after DeleteTableBucketEncryption... | poll_window_after_delete_s=120, poll_window_after_put_s=60, polls_inherit=66, polls_plain=60 | 126 |
| [S3T-TABLE-059](#s3t-table-059) | Explicit storageClassConfiguration and encryptionConfiguration at CreateTable round-trip byte-for-byte through the two sub-reads (4/4) | create_table_latency_ms=[396, 311, 447, 343] | 4 |
| [S3T-TABLE-061](#s3t-table-061) | GetTableEncryption/GetTableStorageClass read-after-write consistent with CreateTable: 30/30 polls at +0/+1/+2 s on 10 {IT+aws:kms} tables | create_table_kms_latency_ms=[430, 475, 439, 493, 483, 476, 438, 501, 427, 442], sub_read_latency_ms=[29, 13, 8, 6, 9, 5, 23, 6, 9, 5, 8, 6, 32, 6, 8, 5, 8, 5, 39, 6, 8, 6, 9, 10, 33, 6, 8, 6, 9, 6, 26, 8, 9, 6, 8, 5, 29, 6, 8, 6, 8, 6, 27, 6, 9, 6, 9, 6, 29, 6, 10, 7, 9, 7, 30, 6, 9, 6, 8, 5], polls=30 | 10 |
| [S3T-TABLE-063](#s3t-table-063) | After DeleteTable (204) GetTable, GetTableEncryption, GetTableStorageClass -> NotFoundException at once; 0 stale hits in 33 polls over 30 s | post_delete_polls=33, stale_hits=0, window_s=30, delete_table_latency_ms=[130] | 1 |
| [S3T-TABLE-248](#s3t-table-248) | CreateTable(aws:kms K) with K Disabled or PendingDeletion -> BadRequestException 400 'KMS key ARN is disabled.' (same text both) | kms_state_visible_in_describe_key_s.disabled=0.01, kms_state_visible_in_describe_key_s.enabled=0.02, kms_state_visible_in_describe_key_s.pending_deletion=0.01, create_table_latency_ms.enabled=652, create_table_latency_ms.disabled=[30, 30], create_table_latency_ms.pending_deletion=[18, 36], create_table_latency_ms.reenabled=531 | 1 |
| [S3T-TABLE-249](#s3t-table-249) | DeleteTable on an SSE-KMS table -> HTTP 204 with the key Disabled and PendingDeletion; GetTableEncryption still returns the key ARN | delete_table_latency_ms=[166, 147] | 2 |
| [S3T-TABLE-252](#s3t-table-252) | ARN of another region on 7 ops (GetTable, GetTableBucket, GetTableEncryption, TagResource, DeleteTable, ...) -> 400 'Region does not match' | get_table_foreign_latency_ms=[31, 21], get_table_home_latency_ms=97 | 2 |
| [S3T-TABLE-319](#s3t-table-319) | RenameTable across namespaces (HTTP 204): tableARN unchanged, old triple -> NotFoundException 404, encryption/storage class follow | old_triple_404_after_s=0.01, new_triple_200_after_s=0.01 | 1 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-encryption-storage.md` - unverified or stale items worth a future probe, one line each.

- [S3T-TABLE-013](#s3t-table-013): the client-side half (`{}`, `{kmsKeyArn only}`) was tested with botocore ParamValidationError only; the Go SDK v2 validators were not executed, so what the controller's CreateTable call rejects before the wire is unverified.
- [S3T-TABLE-248](#s3t-table-248): n=1 per key state and offsets below +5 s after DisableKey/EnableKey/ScheduleKeyDeletion were not probed at CreateTable; Disabled and PendingDeletion share the message text, so no probe has shown a way to tell them apart from the 400.
- [S3T-TABLE-242](#s3t-table-242) / -249: n=1 per state; deletion with a key already past its PendingDeletion window (the lab keys expire 2026-10-20/21) was not probed.
- [S3T-TABLE-020](#s3t-table-020): `bucket_storage_class_visible_after_put_s` is n=1 here (n=2 in [S3T-TABLE-021](#s3t-table-021)); no DeleteTableBucketStorageClass exists, so what an omitted block inherits after "un-setting" the bucket class cannot be tested.
- [S3T-TABLE-319](#s3t-table-319): single run; RenameTable across namespaces was sent without `versionToken`, and the controller only sends `newName` (hooks.go:126-131) - a namespace change via the controller path is untested.
- [S3T-TABLE-062](#s3t-table-062): identifier-form matrix on one table; GetTableStorageClass with a foreign-region ARN was not among the operations of [S3T-TABLE-252](#s3t-table-252).
- [S3T-TABLE-252](#s3t-table-252): single run with 2 GetTable samples; the home-endpoint latency is a single number (not quoted).
- [S3T-TABLE-057](#s3t-table-057): contradiction with [S3T-TABLE-141](table-policy.md#s3t-table-141) settled by FU2-07 in favour of both readings under different conditions ([S3T-TABLE-328](table-policy.md#s3t-table-328) canonical, table-policy.md); -057 itself has no measurements block.
- [S3T-TABLE-063](#s3t-table-063): n=1 table for the 33-poll post-delete window; the same-name re-create reading the new config is one observation.
- [S3T-TABLE-011](#s3t-table-011): not independently re-verified ([S3T-TABLE-010](#s3t-table-010) was, 6/6); the "ghost table" identical-404 observation rests on one probe.
- [S3T-TABLE-030](#s3t-table-030) / -161: model-derived (botocore 2018-05-10, s3tables v1.16.4 / botocore 1.42.97); re-check for a PutTableStorageClass or UpdateTable after the next SDK bump.
- [S3T-TABLE-257](#s3t-table-257): the 'ungranted key' cell is a doc-claim verdict built on [S3T-TABLE-055](table-maintenance.md#s3t-table-055) / [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119) / [S3T-TABLE-239](table-record-expiration.md#s3t-table-239) rendered in other docs; no probe here created a table under an ungranted bucket default with an explicit `aws:kms` of a different, granted key.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-012"></a>**S3T-TABLE-012** | request-validation | low | confirmed | CreateTable storageClass 'standard', 'intelligent_tiering', 'INTELLIGENT-TIERING', 'GLACIER', 'ONEZONE_IA', '' -> BadRequestException 400 | [S3T-TABLEBUCKET-087](tablebucket-encryption-storage.md#s3t-tablebucket-087), [S3T-TABLEBUCKET-024](tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-013](#s3t-table-013), [S3T-TABLE-020](#s3t-table-020), [S3T-TABLE-084](#s3t-table-084), [S3T-TABLE-085](#s3t-table-085), [S3T-TABLE-087](#s3t-table-087) | - |
| <a id="s3t-table-013"></a>**S3T-TABLE-013** | request-validation | low | confirmed | CreateTable encryptionConfiguration={} / storageClassConfiguration={}: botocore rejects client-side; on the wire -> BadRequestException 400 | [S3T-TABLEBUCKET-024](tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLEBUCKET-023](tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-087](#s3t-table-087), [S3T-TABLEBUCKET-087](tablebucket-encryption-storage.md#s3t-tablebucket-087), [S3T-TABLE-012](#s3t-table-012), [S3T-TABLE-020](#s3t-table-020), [S3T-TABLE-084](#s3t-table-084), [S3T-TABLE-085](#s3t-table-085) | - |
| <a id="s3t-table-060"></a>**S3T-TABLE-060** | read-gap | high | confirmed | GetTable, ListTables and GetTableMetadataLocation expose no encryption or storage-class fields even when both were set explicitly at create | - | [S3T-TABLE-308](#s3t-table-308) |
| <a id="s3t-table-255"></a>**S3T-TABLE-255** | other | low | confirmed | Doc claim C013 TRUE: CreateTable{aws:kms, granted key} -> 200; GetTableEncryption returns aws:kms and the kmsKeyArn byte-identical | [S3T-TABLE-059](#s3t-table-059), [S3T-TABLE-057](#s3t-table-057), [S3T-TABLE-085](#s3t-table-085), [S3T-TABLE-259](#s3t-table-259), [S3T-TABLE-257](#s3t-table-257), [S3T-TABLEBUCKET-120](tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLE-010](#s3t-table-010) | - |
| <a id="s3t-table-257"></a>**S3T-TABLE-257** | other | low | confirmed | Doc claim C015 TRUE: CreateTable w/o encryptionConfiguration in an aws:kms bucket (ungranted key) -> 400 as C014; AES256 bucket -> 200 | [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-239](table-record-expiration.md#s3t-table-239), [S3T-TABLE-010](#s3t-table-010), [S3T-TABLEBUCKET-139](tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](tablebucket-encryption-storage.md#s3t-tablebucket-142), [S3T-TABLE-256](table-maintenance.md#s3t-table-256), [S3T-TABLEBUCKET-134](tablebucket-encryption-storage.md#s3t-tablebucket-134), [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLE-255](#s3t-table-255), [S3T-TABLE-259](#s3t-table-259), [S3T-TABLE-057](#s3t-table-057), [S3T-TABLEBUCKET-120](tablebucket-encryption-storage.md#s3t-tablebucket-120) | - |
| <a id="s3t-table-258"></a>**S3T-TABLE-258** | other | low | confirmed | Doc claim C016 TRUE: CreateTable{aws:kms} without kmsKeyArn -> BadRequestException 400 (message: invalid KMS key ARN format) | [S3T-TABLE-084](#s3t-table-084), [S3T-TABLEBUCKET-135](tablebucket-encryption-storage.md#s3t-tablebucket-135), [S3T-TABLEBUCKET-136](tablebucket-encryption-storage.md#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](tablebucket-encryption-storage.md#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](tablebucket-encryption-storage.md#s3t-tablebucket-138), [S3T-TABLEBUCKET-034](tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLE-085](#s3t-table-085) | - |
| <a id="s3t-table-259"></a>**S3T-TABLE-259** | other | low | confirmed | Doc claim C017 TRUE: CreateTable{AES256} in an aws:kms-default bucket -> 200; GetTableEncryption returns AES256 (table overrides bucket) | [S3T-TABLE-057](#s3t-table-057), [S3T-TABLEBUCKET-120](tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLE-255](#s3t-table-255), [S3T-TABLE-257](#s3t-table-257), [S3T-TABLE-010](#s3t-table-010) | - |
| <a id="s3t-table-260"></a>**S3T-TABLE-260** | other | low | confirmed | Doc claim C019 TRUE: PutTableBucketStorageClass{INTELLIGENT_TIERING}: a new table without storageClassConfiguration reads IT | [S3T-TABLE-020](#s3t-table-020), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-145](tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-261](#s3t-table-261), [S3T-TABLE-262](#s3t-table-262), [S3T-TABLE-021](#s3t-table-021) | - |
| <a id="s3t-table-261"></a>**S3T-TABLE-261** | other | low | confirmed | Doc claim C020 TRUE: A table created before PutTableBucketStorageClass{INTELLIGENT_TIERING} still reads STANDARD afterwards | [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLE-021](#s3t-table-021), [S3T-TABLEBUCKET-145](tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](#s3t-table-260), [S3T-TABLE-262](#s3t-table-262), [S3T-TABLE-020](#s3t-table-020) | - |
| <a id="s3t-table-262"></a>**S3T-TABLE-262** | other | low | confirmed | Doc claim C022 TRUE: CreateTable with explicit storageClassConfiguration{STANDARD} in an INTELLIGENT_TIERING-default bucket reads STANDARD | [S3T-TABLE-021](#s3t-table-021), [S3T-TABLE-020](#s3t-table-020), [S3T-TABLEBUCKET-145](tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](#s3t-table-260), [S3T-TABLE-261](#s3t-table-261), [S3T-TABLEBUCKET-084](tablebucket-encryption-storage.md#s3t-tablebucket-084) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
