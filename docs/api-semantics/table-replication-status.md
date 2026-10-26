<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table replication status and replica tables (GetTableReplicationStatus, service-managed replicas)
_The asynchronous side of table replication: GetTableReplicationStatus entries per destination state (pending/completed/failed), the ~60 s evaluation after every Put, failed destinations recovering only via a token-refreshing re-Put, replica tables as service-managed resources (type, managedByService, orphans after DeleteTableReplication), same-name collisions, dead destinations._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-replication-status.md`; every claim cites the finding ids that carry the full evidence
below. The configuration API (Put/Get/Delete, token rules, validation, caller permissions) is the sibling `table-replication.md`.

### Orientation

This document covers what happens after PutTableReplication returns 200: the GetTableReplicationStatus body (one entry
per destination with replicationStatus pending/completed/failed and, per state, destinationTableArn,
lastSuccessfulReplicatedUpdate or failureMessage), the service-created replica tables and the leftovers
([S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), -033, [S3T-TABLE-136](#s3t-table-136), -223). The controller implements no replication resource (scope verdict in
`table-replication.md`), so every finding is input for a would-be TableReplication reconciler's status and delete paths.
The model is asynchronous and Put-driven: destinations is [] for ~60 s after the first Put, then 'pending' and, on an empty
table, 'completed' ~1 s later; every Put - a replacement or a byte-identical re-Put alike - schedules one evaluation ~61 s
later and the status shows the previous view until then; nothing else re-evaluates an untouched configuration
([S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), -019, -043, -045, -047, -048). Permission, destination and name problems are accepted with 200
and surface only as 'failed' at +60.6-72.5 s; a 'failed' destination never recovers on its own and only a re-Put with the
current token (which rotates it) re-evaluates it ([S3T-TABLEREPLICATION-002](#s3t-tablereplication-002), -008, -025, -028, -029, -032, -043).
'completed' is not terminal: a source commit flips it back to 'pending' for 403-428 s (-006, -031). The replica is a
service-managed table in the destination bucket (type 'customer', managedByService='replication.s3tables.amazonaws.com',
managedTableInformation.replicationInformation.sourceTableARN; the source carries neither) that outlives
DeleteTableReplication and DeleteTable(source) as an orphan - a reconciler owns a cleanup it did not create
([S3T-TABLE-136](#s3t-table-136), -138, -195, -196, -224, -225, -268, -329).

### Rules

1. Never judge a Put from the first status read: for ~61 s after any Put the body is the previous view (stale 'failed' with
   the old failureMessage after a corrective re-Put, old destination 'completed' after a replacement, stale 'completed' of
   a dead destination); a controller reading right after its Put would wrongly report failure or Put again; 'completed' is
   the only state that proves the new Put was evaluated ([S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), -027, -031, -042, -043, -047, -048).
2. Deleting a replica under an active configuration is destructive: DeleteTable(replica) -> 204, status stays 'completed',
   the replica is not re-created within 300 s, and the next source commit turns the entry 'failed: The destination table has
   been deleted during replication.'; only a re-Put re-created it ([S3T-TABLE-138](#s3t-table-138), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027)).
3. Orphans: after DeleteTableReplication (Status -> 404 at once) the replica keeps type 'customer', managedByService and
   managedTableInformation.replicationInformation.sourceTableARN on every sample, same- and cross-region; commit -> 400
   'Updates to service managed tables are not allowed.', PutTableMaintenanceConfiguration and PutTableReplication on it ->
   400. DeleteTable(source) -> 204 in every state (no-destinations, pending, completed, failed) and its replicas persist,
   still naming the deleted source; DeleteTable(replica) -> 204 (7 replicas), then DeleteNamespace and DeleteTableBucket ->
   204; sourceTableARN does not indicate an active replication ([S3T-TABLE-139](#s3t-table-139), -195, -196, -225, -307, -329,
   [S3T-TABLEREPLICATION-035](#s3t-tablereplication-035)).
4. A re-Put over a leftover replica does not re-attach it: 200, 'pending', then 'failed: A table with an identical name
   already exists at the destination.' with a destinationTableArn differing from the old replica; a user table with the
   source's namespace/name fails the same way and is left untouched (versionToken, modifiedAt unchanged, managedByService None);
   a pre-existing destination namespace is reused; delete the leftover first ([S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), -025).
5. 'failed' never self-heals: restoring the trust policy, adding the destination to the role policy or deleting the colliding
   table changed nothing in 150-320 s; a re-Put of the byte-identical configuration with the current token -> 200 and the
   failed destination went 'pending' -> 'completed' while a completed sibling destination never flipped - the fix is a
   token-refreshing re-Put, not waiting ([S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), -027, -029, -043).
6. A Put 200 is not a success signal (doc claim C071 FALSE): a nonexistent role, a role trusting only ec2.amazonaws.com, a
   role allowed on one of two destinations and a well-formed ARN of a nonexistent bucket are all accepted; 'failed' follows
   with the generic 'Insufficient permissions given to successfully complete replication.' (naming neither role, action nor
   bucket) or 'The specified destination bucket does not exist.'; failed entries carry no destinationTableArn and no
   lastSuccessfulReplicatedUpdate; destinations are independent - B1 completed while B2 stayed 'failed'
   ([S3T-TABLEREPLICATION-002](#s3t-tablereplication-002), -005, -008, -028, -032, -033).
7. 'completed' is not terminal and 'pending' is not an error: a source commit flips the destination to 'pending' while
   lastSuccessfulReplicatedUpdate.metadataLocation still names the previous file (two commits applied in order); a schema
   table stayed 'pending' >= 300 s with the replica already listable; destinationTableArn is absent on 'failed' and on a
   first-backfill 'pending' but present on a re-Put 'pending' and at 'completed' - per entry, optional in every
   non-completed state ([S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), -019, -028, -029, -031, -045).
8. Dead destinations are reported only when something triggers an evaluation: a deleted destination bucket read 'failed: The
   specified destination bucket does not exist.' within 5 s only for configs that got a 10x re-Put storm ~60 s before (3/3)
   and stayed 'completed' for 180 s otherwise (3/3); a Put naming [dead, new] -> 200 (6/6) and, after the stale view, both
   'failed' and 'completed'; re-pointing to a new bucket alone leaves no dead entry; Put accepts a deleted or nonexistent
   destination ([S3T-TABLEREPLICATION-040](#s3t-tablereplication-040), -042, -048).
9. Bucket-level and table-level combine by precedence, not union: under a bucket-level rule only, GetTableReplication(table)
   -> 404 while GetTableReplicationStatus -> 200 with the bucket destination (completed in ~2 s); a table-level Put (no
   token) -> 200 and the status went 'D:completed|E:pending' then 'E:completed' only; naming the bucket-level destination in
   a table-level Put -> 'D:failed' (identical name) while E completed; after DeleteTableReplication the stale table view
   lasts ~60 s, then the bucket view; after DeleteTableBucketReplication both -> 404 ([S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), -022).
10. The replica is a partial, destination-shaped copy: same namespace/name as the source at creation; equal createdBy,
    format, name, namespace, ownerAccountId, type; own createdAt, modifiedAt, namespaceId, tableARN, tableBucketId,
    versionToken, warehouseLocation; storage class and encryption follow the destination bucket (INTELLIGENT_TIERING vs
    STANDARD and the reverse; AES256 / aws:kms); snapshot management copied, compaction forced disabled 512/auto; tags {},
    policy 404; replication Gets on the replica -> 400 'Replication operations are not allowed on service managed tables.'
    ([S3T-TABLE-136](#s3t-table-136), -223, -268, -299, [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030)).
11. Replica writes split: RenameTable -> 204 (managedByService kept, source status still 'completed' with the same
    destinationTableArn), PutTablePolicy -> 200, TagResource -> 200 / UntagResource -> 204 (source tags not copied);
    PutTableMaintenanceConfiguration, UpdateTableMetadataLocation and PutTableReplication(replica -> S) -> 400,
    PutTableRecordExpirationConfiguration -> 405 ([S3T-TABLE-224](#s3t-table-224)).
12. Source changes do not propagate: RenameTable (name, then namespace) leaves the replica at the old namespace/name with
    status 'completed', an identical re-Put is a no-op, no new replica or namespace appears; tags {} and policy 404 on the
    replica 5 s after the source writes; later source maintenance and policy changes did not reach the replica (the source
    TagResource in that step got a 500); after a rename the replica is found only through
    destinations[].destinationTableArn ([S3T-TABLE-300](#s3t-table-300), -306, [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041)).
13. Two unrelated 'failed's render here: the replication entry's lowercase 'failed' with failureMessage, and the maintenance
    job status 'Failed' (model casing, lastRunTimestamp + failureMessage 'Insufficient access to perform table
    maintenance...') reached by each job type minutes after a metadata-only commit under a revoked KMS key grant and kept
    through Disabled and PendingDeletion, states in which only CreateTable in the bucket fails (400, same text)
    ([S3T-TABLE-236](table-maintenance.md#s3t-table-236), -237, -238, -239, -240, -241, rendered in `table-maintenance.md` and `table-record-expiration.md`); no
    replication finding involves a KMS key state.

### Timing ranges

- destinations [] after the first Put until the first entry: 60.6-72.5 s ([S3T-TABLEREPLICATION-005](#s3t-tablereplication-005) n=3, -019 n=1, -025 n=3,
  -027 n=3, -030 n=2, -002, -008, -028, -045 n=1 each; [S3T-TABLE-195](#s3t-table-195) tc 66.25 s).
- Empty table, Put -> 'completed': 62-72 s ([S3T-TABLEREPLICATION-031](#s3t-tablereplication-031) summary n=12 across 7 probes; -019, -021, -047 n=1
  each; -025, -027, -030 n=2 each; -041, [S3T-TABLE-299](#s3t-table-299), [S3T-TABLE-306](#s3t-table-306) n=1 each); one gap observation by 138.2 s
  ([S3T-TABLEREPLICATION-005](#s3t-tablereplication-005)); replica listable 62.2-75.0 s after the Put (-019, -005, -027).
- Schema table, Put -> 'completed': 436-521 s ([S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), -028, -029 n=1 each), the 'pending' entry at
  61.9-66 s first; still 'pending' at 300 s (-045 n=1); source commit -> 'pending' -> 'completed': 403-428 s (-006 n=2).
- 'failed' after an accepted Put: 60.6-72.5 s ([S3T-TABLEREPLICATION-002](#s3t-tablereplication-002), -005, -008, -020, -022, -025, -027, -028, -043; n=1
  each); role cases 60.6-65.7 s, name collisions 62.0-72.5 s.
- Replacement Put on a completed config: new entry 'pending' 61.0-64.1 s, 'completed' 62.0-65.2 s, old entry gone 61.1-65.2
  s ([S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) n=6 at Put offsets 0.9-70 s after the last change; -042, -021 n=1 each; -017's +1.06 s, n=1,
  did not reproduce); under a bucket-level rule: 'pending' 1.1-1.12 s, 'completed' 2.13-2.16 s (-021, -022 n=1 each).
- Corrective re-Put on a 'failed' destination -> 'completed': 61.2-75.0 s on empty tables ([S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), -027,
  -043 n=1 each; -027 replica re-created at 75.0 s), 520.83 s on a schema table (-029 n=1); stale 'failed' for 61.17 s after
  the re-Put (-043 n=1); no-self-heal watches 150-320 s (-008, -029, -043 n=1 each); dead destination 'failed' at 4.4-4.6 s
  after a storm (-048 n=3), unchanged 180 s without (-048 n=3; -040 182.7 s n=1), dead+new Put: both entries at 63.2-80.0 s
  (-048 n=6); deleted replica not re-created within 300 s ([S3T-TABLE-138](#s3t-table-138) n=1).
- Source RenameTable: replica unchanged over 122.3-244.6 s windows ([S3T-TABLE-306](#s3t-table-306) n=1, three windows) and 150 s
  ([S3T-TABLEREPLICATION-041](#s3t-tablereplication-041) n=1); source maintenance/policy changes absent on the replica over 184.3 s ([S3T-TABLE-300](#s3t-table-300), 36
  samples); orphan replica still service-managed with sourceTableARN: 0-479.3 s ([S3T-TABLE-329](#s3t-table-329) n=12, 3 replicas), +0/+90 s
  after DeleteTable(source) ([S3T-TABLE-196](#s3t-table-196)), 2 min ([S3T-TABLE-225](#s3t-table-225) n=1), 874 s/30 samples ([S3T-TABLE-307](#s3t-table-307) n=1).
- KMS maintenance-job 'Failed' (rule 13, renders in `table-maintenance.md`): 5.0 / 6.0 / 8.1 min after the commit ([S3T-TABLE-236](table-maintenance.md#s3t-table-236),
  -237, -238 n=1 each); soaks 20 / 15 / 10 min ([S3T-TABLE-239](table-record-expiration.md#s3t-table-239) n=21, -240 n=17, -241 n=10, in `table-record-expiration.md`).

### Handling gaps

The controller implements Table, Namespace and TableBucket only and `generator.yaml` carries no replication resource,
operation or field, so no finding here compares against existing controller code and none is marked `suspect-bug`. A future
TableReplication implementation must handle at minimum, with the evidence here: the ~61 s stale window after every Put and
'completed' as the only proof of evaluation ([S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), -047); asynchronous 'failed' after an accepted Put
with generic messages and per-destination independence (-002, -005, -008, -028, -032); recovery only through a
token-refreshing re-Put, never by waiting (-008, -027, -029, -043); 'completed' flipping to 'pending' on commits and to
'failed' on a deleted replica or bucket, dead destinations reported only after an evaluation (-006, -031, -040, -048,
[S3T-TABLE-138](#s3t-table-138)); leftover service-managed replicas after DeleteTableReplication and DeleteTable(source), their read-only
writes and the identical-name failure on re-Put ([S3T-TABLE-139](#s3t-table-139), -195, -196, -224, -225, -329, [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020),
-025); bucket-level precedence and the ~60 s fallback view (-021, -022); replica identity and inheritance that no source
update propagates ([S3T-TABLE-136](#s3t-table-136), -223, -299, -300, -306, [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), -041).

### Where next

- `table-replication.md` - the write side this status follows: configuration vs table versionToken and rotation on every
  Put, token rules, request validation, iam:PassRole, Delete -> 204 with both Gets 404 at once ([S3T-TABLEREPLICATION-004](table-replication.md#s3t-tablereplication-004),
  -015, -018, -044, -046 render there). `tablebucket-replication.md` - the bucket-level twin; read the precedence cells
  ([S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), -022) and the bad-trust role finding (-008) against it; its Phase 8 correction that a
  bucket-level replica deleted at the destination IS re-created in 31-47 s ([S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032), there) has no
  table-level counterpart ([S3T-TABLE-138](#s3t-table-138) saw none in 300 s).
- `table.md` - DeleteTable/RenameTable/ListTables behind [S3T-TABLE-138](#s3t-table-138), -139, -195, -196, -306 ([S3T-TABLE-231](table.md#s3t-table-231) and -317
  render there); `table-metadata-location.md` - the commit that re-pends a destination ([S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), -031;
  [S3T-TABLE-031](table-metadata-location.md#s3t-table-031) and -315 render there); `table-maintenance.md` / `table-encryption-storage.md` - what the replica inherits
  ([S3T-TABLE-136](#s3t-table-136), -299, -300, [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030)) and the KMS key-state findings of rule 13 ([S3T-TABLE-236](table-maintenance.md#s3t-table-236) to -238;
  -239 to -241 render in `table-record-expiration.md`); `table-policy.md` - PutTablePolicy on a replica ([S3T-TABLE-224](#s3t-table-224));
  `table-record-expiration.md` - the 405 on service-managed tables ([S3T-TABLE-224](#s3t-table-224), -225; [S3T-TABLE-221](table-record-expiration.md#s3t-table-221) renders there).
- Long catalogues worth their `details/<id>.md` page: [S3T-TABLE-224](#s3t-table-224) (replica write matrix), [S3T-TABLE-223](#s3t-table-223) (field-by-field
  replica/source diff), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022) (bucket/table interplay cells), -048 (six-table dead-destination storm),
  -025/-027 (collision and recovery transitions), -047 (six replacement offsets), [S3T-TABLE-239](table-record-expiration.md#s3t-table-239)/-240/-241 (KMS phase tables).
- Open experiments (STATUS Phase 8, F-followups-2): FU2 settled the replacement delay as a fixed ~61 s per Put
  ([S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) canonical; -017 superseded), replicationInformation nested under managedTableInformation
  ([S3T-TABLE-329](#s3t-table-329) over -307) and dead destinations 'failed' only after a re-Put storm ([S3T-TABLEREPLICATION-048](#s3t-tablereplication-048) over -040,
  -042); self-heal after re-creating a missing destination bucket (-005 notes) and recovery of a deleted replica by a new
  Put ([S3T-TABLE-138](#s3t-table-138) notes) were not tested. See `table-replication-status.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 39 (high 23 / medium 14 / low 2); duplicates folded into the appendix: 2
- handling: handled 0 · partial 0 · tracked 0 · unhandled 39 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 0 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: async-state-machine 10, delete-semantics 7, prerequisite 5, eventual-consistency 4,
  requested-vs-effective 3, immutable-field 2, other 2, read-gap 2, stale-response 2, idempotency 1,
  response-fidelity 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateNamespace | create | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketReplication | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableReplication | delete | tableArn, versionToken | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetNamespace | read | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketReplication | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMaintenanceConfiguration | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTablePolicy | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableRecordExpirationConfiguration | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableRecordExpirationJobStatus | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableReplication | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableReplicationStatus | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketEncryption | create | tableBucketARN, encryptionConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketReplication | create | tableBucketARN, configuration | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableMaintenanceConfiguration | create | tableBucketARN, namespace, name, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTablePolicy | create | tableBucketARN, namespace, name, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableRecordExpirationConfiguration | create | tableArn, value | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| PutTableReplication | create | tableArn, configuration | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

Also referenced by findings but not in the s3tables model (other services or annotated variants):
PutRolePolicy, UpdateAssumeRolePolicy

## State machine

- **JobStatus**: Not_Yet_Run, Successful, Failed, Disabled
- **ReplicationStatus**: pending, completed, failed (transitional: pending)
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)
- **TableRecordExpirationJobStatus**: NotYetRun, Successful, Failed, Disabled

- <a id="s3t-tablereplication-005"></a>**S3T-TABLEREPLICATION-005** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableReplicationStatus: destinations [] for ~60 s, then pending; empty table completed by +138 s, schema table at +436 s...**
  Right after PutTableReplication the status body is {sourceTableArn, destinations: []} (empty list, no state)
  and stays so for about a minute. Relative to each table's Put: t1 (CreateTable with an Iceberg schema)
  destinations[0] appeared as replicationStatus='pending' with destinationTableArn at +61.9 s (the replica
  table was listable in the destination bucket at about the same time, +65 s) and became 'completed' at +435.7
  s with lastSuccessfulReplicatedUpdate{metadataLocation (source file name), timestamp}. t3 (CreateTable
  WITHOUT metadata) was already 'completed' at +138 s (first observation after a gap; it was still [] at +63
  s). t2 (well-formed ARN of a destination bucket that does not exist): Put -> 200; destinations[0] became
  'failed' at +63.6 s with failureMessage 'The specified destination bucket does not exist.' and no
  destinationTableArn.
  - ACK: synced.when, requeue, e2e-timing · ops: GetTableReplicationStatus, PutTableReplication · fields:
    destinations[].replicationStatus, destinations[].destinationTableArn,
    destinations[].lastSuccessfulReplicatedUpdate, destinations[].failureMessage
  - repro: PutTableReplication on three tables (schema / no metadata / missing destination) ->
    GetTableReplicationStatus every 2 s for <= 10 min
  - measurements: status_populated_after_put_s=[61.9, 63.6], empty_table_completed_by_s=[138.2],
    schema_table_pending_to_completed_s=[435.7], missing_destination_failed_at_s=[63.6], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-033](#s3t-tablereplication-033), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028),
    [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-011](tablebucket-replication.md#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](tablebucket-replication.md#s3t-tablebucketreplication-010),
    [S3T-TABLEBUCKETREPLICATION-029](tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-030](tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031),
    [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043) ·
    hypotheses: H-U-134, H-U-135, H-S-058, H-S-062 · evidence: table/round-trip/repl-put-get
  - notes: H-U-134 (synchronous 404 for a missing destination) REFUTED: Put returned 200. H-U-135 confirmed:
    the failure is asynchronous, 'failed' with a message naming the missing bucket at ~64 s. H-S-058/H-S-062
    are covered at bucket level in tablebucket/sub-resources/repl-backfill-delete. Self-heal after...
  - full notes: [details/S3T-TABLEREPLICATION-005.md](details/S3T-TABLEREPLICATION-005.md)

- <a id="s3t-tablereplication-019"></a>**S3T-TABLEREPLICATION-019** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Empty table, table-level Put: status [] until +61.06s, pending until +62.09s, then 'completed'; replica listable in D at +62.2s (n=1)**
  PutTableReplication(empty table -> D) -> HTTP 200. GetTableReplicationStatus at 1 s: [(0.06,
  'no-destinations'), (61.06, 'pending'), (62.09, 'completed')]. The replica table (same namespace/name) was
  first listed by ListTables(D) at +62.2s (polled every ~5 s). Raw entry while pending: {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "pending"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}.
  Raw entry at terminal: {"destinations": [{"destinationTableArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}.
  Note the key spelling destinationTableBucketArn (status) vs destinationTableBucketARN (configuration).
  - ACK: synced.when, e2e-timing, requeue · ops: PutTableReplication, GetTableReplicationStatus, ListTables ·
    fields: destinations[].replicationStatus, destinations[].destinationTableArn,
    destinations[].destinationTableBucketArn, destinations[].lastSuccessfulReplicatedUpdate
  - repro: CreateTable(no metadata) -> PutTableReplication -> GetTableReplicationStatus @1 s + ListTables(D)
    @5 s until terminal
  - measurements: no_destinations_window_s=[61.06], pending_to_terminal_s=[1.03], put_to_terminal_s=[62.09],
    replica_listable_s=[62.2], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-033](#s3t-tablereplication-033),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031),
    [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-224](#s3t-table-224), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029) · hypotheses: H-U-151,
    H-U-152, H-U-161, H-U-162, H-U-197 · evidence: table/state-machine/repl-status-lifecycle
  - notes: H-U-151 qualified: no 404 window, but the body is {sourceTableArn, destinations: []} (no entry at
    all) for ~1 min before 'pending'. H-U-152 REFUTED (empty table reached 'completed' without any commit).
    H-U-161 confirmed / H-U-162 refuted: namespace and table are created by the service in the empty...
  - full notes: [details/S3T-TABLEREPLICATION-019.md](details/S3T-TABLEREPLICATION-019.md)

- <a id="s3t-tablereplication-027"></a>**S3T-TABLEREPLICATION-027** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Re-Put (current token) on a 'failed' collision dest after the fix: 'completed' +63.3s; re-Put after DeleteTable(replica) re-creates it +75s**
  First Puts: t1 (user table with the same name in D) [(0.1, 'no-destinations'), (72.5, 'failed')] -> failed
  (message 'A table with an identical name already exists at the destination.'); t2 [(1.4, 'no-destinations'),
  (61.0, 'pending'), (64.1, 'completed')]; t3 [(2.6, 'no-destinations'), (67.3, 'completed')]. Fixes:
  DeleteTable(user table) -> HTTP 204; DeleteTable(replica t2) -> HTTP 204; DeleteTable(replica t3) -> HTTP
  204; status right after the deletes: {'t1': 'failed', 't2': 'completed', 't3': 'completed'}. Re-Put
  (identical config, current token) on t1 -> HTTP 200 and t2 -> HTTP 200; t3 gets no Put. Status over the next
  4 min (1 call/s round-robin, ListTables(D) every 5 rounds): t1 [(0.1, 'failed'), (63.3, 'completed')]
  (listed in D at +63.4s, final message None); t2 [(1.2, 'completed'), (66.7, 'pending'), (68.7, 'completed')]
  (listed at +75.0s, final message None); t3 [(2.3, 'completed')] (listed at +Nones). Final GetTable(D, ns,
  <name>): {"t1": {"managedByService": "replication.s3tables.amazonaws.com", "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/6f26a844-8319-4864-b62c-f3c103842fd4"},
  "t2": {"managedByService": "replication.s3tables.amazonaws.com", "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/f8d4a971-f9a8-4523-bbf8-decb4a4e5457"},
  "t3": {"error": "NotFoundException"}}.
  - ACK: requeue, custom_update, synced.when, pre-delete-cleanup · ops: PutTableReplication,
    GetTableReplicationStatus, DeleteTable, ListTables, GetTable · fields: destinations[].replicationStatus,
    destinations[].failureMessage, destinations[].destinationTableArn
  - repro: Put x3 -> terminal; fix causes (delete user table / replicas); re-Put t1, t2 (token); poll 4 min
    ignoring the stale first entry
  - measurements: t1_fix_re_put_to_terminal_s=[63.3], t2_replica_recreated_s=[75.0], t1_first_failed_s=[72.5],
    n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028) ·
    hypotheses: H-U-155, H-U-156, H-U-167, H-U-176 · evidence: table/state-machine/repl-failed-recovery
  - notes: H-U-155 confirmed (a token-refreshing re-Put is the recovery path once the cause is gone). H-U-167:
    a deleted replica IS re-created by a re-Put; the control t3 (no Put) stayed 'completed'. Right after any
    Put the status still returns the PREVIOUS entry for ~60 s - a controller must not judge a...
  - full notes: [details/S3T-TABLEREPLICATION-027.md](details/S3T-TABLEREPLICATION-027.md)

- <a id="s3t-tablereplication-028"></a>**S3T-TABLEREPLICATION-028** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Two destinations, role allowed on B1 only (schema table): B2 'failed' at +66s, generic 'Insufficient permissions' message; B1 completed 475s**
  Role: trust replication.s3tables.amazonaws.com, inline policy Allow s3tables:* on the source bucket and B1
  (bucket ARN and ARN/*), nothing on B2. PutTableReplication(destinations [B1, B2]) -> HTTP 200 {versionToken,
  status 'success'} on the first attempt (5 s after CreateRole). GetTableReplicationStatus every 5 s: +0 s
  destinations [] (no entries); +61 s B1 'pending' only (B2 entry not yet present); +66 s B1 'pending' + B2
  'failed'; +475 s B1 'completed' + B2 'failed'. Final B1 entry: {replicationStatus 'completed',
  destinationTableBucketArn, destinationTableArn, lastSuccessfulReplicatedUpdate {metadataLocation (source
  file), timestamp}}. Final B2 entry: {replicationStatus 'failed', destinationTableBucketArn, failureMessage
  'Insufficient permissions given to successfully complete replication.'} - no destinationTableArn, no
  lastSuccessfulReplicatedUpdate. ListTables: the replica exists in B1, none in B2. The B2 message is the same
  generic text seen for a nonexistent role / bad trust policy ([S3T-TABLEREPLICATION-002](#s3t-tablereplication-002)/008): it names neither
  the denied action nor the destination bucket.
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: PutTableReplication,
    GetTableReplicationStatus, ListTables · fields: destinations[].replicationStatus,
    destinations[].failureMessage, destinations[].destinationTableArn,
    destinations[].destinationTableBucketArn, configuration.role
  - repro: CreateRole(policy on source+B1) -> CreateTable(schema) -> PutTableReplication([B1,B2]) ->
    GetTableReplicationStatus every 5 s
  - measurements: b1_completed_after_put_s=[475.15], b2_failed_after_put_s=[65.72], put_attempts=[1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), [S3T-TABLEREPLICATION-033](#s3t-tablereplication-033), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006),
    [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](table-replication.md#s3t-tablereplication-013), [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](table-policy.md#s3t-table-026),
    [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-224](#s3t-table-224)
    · hypotheses: H-U-154 · evidence: tablereplication/state-machine/per-destination-failure
  - notes: H-U-154 confirmed for the per-destination part (independent statuses, the allowed destination
    completed while the other stayed failed) and for the 30-min bound (failed at 66 s); REFUTED for 'message
    naming the denied action' (generic text). Members are present-or-absent per entry (failed entries...
  - full notes: [details/S3T-TABLEREPLICATION-028.md](details/S3T-TABLEREPLICATION-028.md)

- <a id="s3t-tablereplication-029"></a>**S3T-TABLEREPLICATION-029** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Fixing the role policy alone does NOT heal a 'failed' dest (180 s); identical re-Put (current token): B2 pending 66s, completed 521s; B1 ok**
  With B2 'failed' and B1 'completed': PutRolePolicy adding B2 to the same inline policy (IAM only, no
  s3tables call) -> GetTableReplicationStatus every 5 s for 180 s: unchanged 'B1 completed / B2 failed'. Then
  GetTableReplication -> current versionToken; PutTableReplication(byte-identical configuration, that token)
  -> HTTP 200 {new versionToken, status 'success'}; status right after the Put still 'B2 failed'; +66 s B2
  'pending' (the failureMessage gone, destinationTableArn now present); +521 s B2 'completed' with
  lastSuccessfulReplicatedUpdate pointing at the same source metadata file as B1. B1 stayed 'completed'
  throughout (never flipped to pending). A replica table then exists in B2.
  - ACK: requeue, custom_update, synced.when · ops: PutRolePolicy, GetTableReplicationStatus,
    GetTableReplication, PutTableReplication · fields: destinations[].replicationStatus,
    destinations[].failureMessage, versionToken
  - repro: B2 failed -> PutRolePolicy(+B2) -> watch 3 min -> GetTableReplication -> PutTableReplication(same
    config, token) -> watch
  - measurements: self_heal_watch_s=[180], re_put_to_b2_pending_s=[65.85], re_put_to_b2_completed_s=[520.83],
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](table-replication.md#s3t-tablereplication-039), [S3T-TABLEREPLICATION-044](table-replication.md#s3t-tablereplication-044),
    [S3T-TABLEREPLICATION-014](table-replication.md#s3t-tablereplication-014), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019) ·
    hypotheses: H-U-154 · evidence: tablereplication/state-machine/per-destination-failure
  - notes: Extends H-U-154 with the recovery path: a controller that only fixes the IAM prerequisite sees the
    destination stay 'failed' (at least 3 min); a token-refreshing re-Put of the unchanged configuration
    re-triggers the failed destination only, without disturbing the completed one. Mirrors...
  - full notes: [details/S3T-TABLEREPLICATION-029.md](details/S3T-TABLEREPLICATION-029.md)

- <a id="s3t-tablereplication-031"></a>**S3T-TABLEREPLICATION-031** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Cross-ref: a source commit flips the destination 'completed' -> 'pending' -> 'completed' (403-428 s); 'completed' is not a stable state**
  Not re-proved in shard F b/c (no data-plane commits were made here). Part a observed on one schema table:
  after UpdateTableMetadataLocation on the source the destination entry returned to 'pending' and became
  'completed' again after 403 s and 428 s for two consecutive commits ([S3T-TABLEREPLICATION-006](#s3t-tablereplication-006)). Shard b/c
  adds that on EMPTY tables the first backfill reaches 'completed' in 62-72 s after the Put (n=12 across 7
  probes) and that right after any Put the status still shows the previous entry for ~60 s.
  - ACK: synced.when, e2e-timing · ops: UpdateTableMetadataLocation, GetTableReplicationStatus · fields:
    destinations[].replicationStatus, destinations[].lastSuccessfulReplicatedUpdate
  - repro: see [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029),
    [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-224](#s3t-table-224) · hypotheses: H-U-153 · evidence:
    table/round-trip/repl-put-get, table/state-machine/repl-status-lifecycle
  - notes: H-U-153 confirmed by part a (cross-reference per the shard brief: same cell, not re-proved). A
    controller mapping 'completed' to ResourceSynced=True would flap on every commit; treat 'pending' as
    healthy-in-progress.

- <a id="s3t-tablereplication-033"></a>**S3T-TABLEREPLICATION-033** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C072 PARTLY: GetTableReplicationStatus: sourceTableArn at once, but destinations [] for ~61 s after the first Put; N entries later**
  Right after PutTableReplication the status body is {sourceTableArn: <source table ARN>, destinations: []}
  and stays empty for ~61 s. Entries then appear per destination (two-destination Put: B1 'pending' at +61 s,
  B2 'failed' at +66 s), each with replicationStatus and destinationTableBucketArn; destinationTableArn and
  lastSuccessfulReplicatedUpdate are present only once 'completed', failureMessage only when 'failed'. Key
  spelling differs from the configuration (destinationTableBucketArn vs destinationTableBucketARN).
  - ACK: synced.when, requeue · ops: GetTableReplicationStatus, PutTableReplication · fields: sourceTableArn,
    destinations[].replicationStatus
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028),
    [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017) · evidence: service/static/doc-claims-3
  - notes: VERDICT: PARTLY - sourceTableArn is present at once, but 'exactly N entries with a status' holds
    only after ~1 min (destinations is [] before).

- <a id="s3t-tablereplication-040"></a>**S3T-TABLEREPLICATION-040** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Destination bucket deleted under a completed config: 'failed: The specified destination bucket does not exist.' in 1 of 2 runs; re-Put 200**
  Replica table + namespace deleted in D, then DeleteTableBucket(D) -> HTTP 204 while the source's
  configuration still named D. GetTableReplicationStatus + GetTableReplication every 5 s for 183s: [{"t": 0.0,
  "status": "failed:The specified destination bucket does not exist.", "get_config": "HTTP 200"}]. Identical
  re-Put naming the deleted bucket (current token) -> HTTP 200; status right after 'failed:The specified
  destination bucket does not exist.', after 20 s 'failed:The specified destination bucket does not exist.'.
  PutTableReplication switching to a new bucket D2 -> HTTP 200; status right after 'failed:The specified
  destination bucket does not exist.'; watch: [{"elapsed_s": 0.03, "value": "failed:The specified destination
  bucket does not exist."}]; final entries [{"replicationStatus": "failed", "failureMessage": "The specified
  destination bucket does not exist."}].
  - ACK: requeue, terminal_codes, custom_update · ops: DeleteTableBucket, GetTableReplicationStatus,
    GetTableReplication, PutTableReplication · fields: destinations[].replicationStatus,
    destinations[].failureMessage, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: completed replication -> empty + DeleteTableBucket(destination) -> poll status 3 min -> re-Put same
    config -> Put new destination -> watch
  - measurements: lost_destination_window_s=182.7, status_change_at_s=null, heal_completed_at_s=0.03, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118),
    [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](table-replication.md#s3t-tablereplication-012), [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023), [S3T-TABLEREPLICATION-048](#s3t-tablereplication-048) ·
    evidence: tablereplication/creative/reput-storm-and-lost-destination
  - notes: Cross-service analogy (ElastiCache CacheSubnetGroupNotFound at call time; DynamoDB replica of a
    deleted table): S3 Tables accepts a configuration whose destination no longer exists (re-Put naming the
    deleted bucket -> HTTP 200, as does a Put to a brand-new bucket D2) and reports the problem only...
  - full notes: [details/S3T-TABLEREPLICATION-040.md](details/S3T-TABLEREPLICATION-040.md)

- <a id="s3t-tablereplication-041"></a>**S3T-TABLEREPLICATION-041** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Replicated source: RenameTable/namespace move leave replica at old ns/name 150 s, status 'completed'; tags/policy/maintenance absent at +5s**
  Replication completed after 66.1 s (same-region destination). A)
  TagResource/PutTablePolicy/PutTableMaintenanceConfiguration on the SOURCE -> HTTP 200 / HTTP 200 / HTTP 204;
  5 s later the replica reads tags {"tags": {}}, policy NotFoundException/404 'The specified table policy does
  not exist.', snapshot maintenance {"status": "enabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}} (status completed). B) RenameTable(source, newName)
  -> HTTP 204; destination/status over 150 s (samples deduplicated; only one distinct state seen): +0.1s:
  D={"ackq_5ad769_ns": ["ackq_5ad769_t"], "ackq_5ad769_ns2": "ERR:NotFoundException"} status=completed
  replica={"managedByService": "replication.s3tables.amazonaws.com", "name": "ackq_5ad769_t", "namespace":
  ["ackq_5ad769_ns"]}. C) RenameTable(source, newNamespaceName=ns2) with ns2 existing only in the source
  bucket -> HTTP 204; over 150 s (one distinct state): +0.1s: D={"ackq_5ad769_ns": ["ackq_5ad769_t"],
  "ackq_5ad769_ns2": "ERR:NotFoundException"} status=completed replica={"managedByService":
  "replication.s3tables.amazonaws.com", "name": "ackq_5ad769_t", "namespace": ["ackq_5ad769_ns"]}; namespaces
  in D at the end: ['ackq_5ad769_ns']. In both cases the replica table kept its old namespace/name and
  managedByService, no new replica or namespace appeared in D and the status never left 'completed' (no
  metadata commit was made on the source).
  - ACK: custom_update, requeue, synced.when, is_immutable · ops: RenameTable, TagResource, PutTablePolicy,
    PutTableMaintenanceConfiguration, GetTableReplicationStatus, ListTables · fields: name, namespace, tags,
    resourcePolicy, icebergSnapshotManagement
  - repro: replicated empty table -> tag/policy/maintenance on source -> read on replica; RenameTable(newName)
    -> watch D/status 150 s; RenameTable(newNamespaceName=ns2 absent in D) -> watch 150 s
  - measurements: completed_at_s=66.1, rename_watch_last_at_s=0.1, move_watch_last_at_s=0.1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028) · evidence:
    tablereplication/creative/source-updates-vs-replica
  - notes: extends [S3T-TABLE-224](#s3t-table-224)/-136 (what a replica accepts) with the SOURCE-side view a Table controller
    has: none of its ordinary updates (tags, policy, maintenance, rename, namespace move) reach the replica,
    and the replication status keeps reporting 'completed' although destination and source no longer...
  - full notes: [details/S3T-TABLEREPLICATION-041.md](details/S3T-TABLEREPLICATION-041.md)

- <a id="s3t-tablereplication-048"></a>**S3T-TABLEREPLICATION-048** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Deleted destination bucket: 'failed ... does not exist.' within 5 s only after a re-Put storm (3/3); otherwise 'completed' for 180 s (3/3)**
  Six empty tables, each with a completed table-level config to its own same-region bucket E_k. t0-t2: 10
  identical re-Puts (current token chain) -> {'t0': ['HTTP 200'], 't1': ['HTTP 200'], 't2': ['HTTP 200']},
  status right after the storm {'t0': 'E0=completed', 't1': 'E1=completed', 't2': 'E2=completed'}; ~60 s later
  all six destination buckets were emptied and deleted ({'t0': 'HTTP 204', 't1': 'HTTP 204', 't2': 'HTTP 204',
  't3': 'HTTP 204', 't4': 'HTTP 204', 't5': 'HTTP 204'}). First GetTableReplicationStatus after each delete:
  {'t0': 'E0=completed', 't1': 'E1=completed', 't2': 'E2=failed:The specified destination bucket does not
  exist.', 't3': 'E3=completed', 't4': 'E4=completed', 't5': 'E5=completed'}. Round-robin polling (<= 1
  call/s, each table every ~6 s) for 180 s, transitions relative to the delete: {"t0": [[4.4, "E0=failed:The
  specified destination bucket does not exist."]], "t1": [[4.6, "E1=failed:The specified destination bucket
  does not exist."]], "t2": [[4.5, "E2=failed:The specified destination bucket does not exist."]], "t3":
  [[4.7, "E3=completed"]], "t4": [[5.0, "E4=completed"]], "t5": [[5.3, "E5=completed"]]}. Then
  PutTableReplication [dead E_k, new N_k] -> {'t0': 'HTTP 200', 't1': 'HTTP 200', 't2': 'HTTP 200', 't3':
  'HTTP 200', 't4': 'HTTP 200', 't5': 'HTTP 200'}; over 150 s the dead destination was listed at all and was
  'failed'; new destination completed at {'t0': 63.2, 't1': 64.0, 't2': 71.1, 't3': 72.0, 't4': 66.7, 't5':
  80.0} s; final views {'t0': 'E0=failed:The specified destination bucket does not exist.|N0=completed', 't1':
  'E1=failed:The specified destination bucket does not exist.|N1=completed', 't2': 'E2=failed:The specified
  destination bucket does not exist.|N2=completed', 't3': 'E3=failed:The specified destination bucket does not
  exist.|N3=completed', 't4': 'E4=failed:The specified destination bucket does not exist.|N4=completed', 't5':
  'E5=failed:The specified destination bucket does not exist.|N5=completed'}.
  - ACK: requeue, synced.when, custom_update · ops: PutTableReplication, DeleteTableBucket,
    GetTableReplicationStatus · fields: destinations[].replicationStatus, destinations[].failureMessage,
    destinations[].destinationTableBucketArn
  - repro: 6 tables with completed table-level replication -> 10 re-Puts on 3 of them -> 60 s ->
    empty+DeleteTableBucket(destination) x6 -> poll status 180 s -> PutTableReplication [dead, new] x6 -> poll
    150 s
  - measurements: storm_first_failed_s=[4.4, 4.6, 4.5], no_storm_first_failed_s=[null, null, null],
    dead_poll_s=180, heal_poll_s=150, heal_new_completed_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0], n=6,
    heal_dead_failed_entry_appears_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-040](#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042) · evidence:
    tablereplication/creative/fu2-03-05-orphan-body-dead-destination
  - notes: FU2-05: both readings are right under different conditions, and the preceding re-Put storm is the
    deciding condition: for the 3 tables whose config received 10 identical re-Puts ~60 s before
    DeleteTableBucket(destination), GetTableReplicationStatus showed 'failed: The specified destination
    bucket...
  - full notes: [details/S3T-TABLEREPLICATION-048.md](details/S3T-TABLEREPLICATION-048.md)

## Idempotency

- <a id="s3t-tablereplication-020"></a>**S3T-TABLEREPLICATION-020** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Re-Put after DeleteTableReplication with the service-managed replica left behind -> 200 but 'failed: identical name already exists' at +62s** (hypothesis refuted; behavior confirmed)
  (a) With the leftover replica present: PutTableReplication(identical config, no token) -> HTTP 200; status
  transitions [(0.05, 'no-destinations'), (60.95, 'pending'), (61.99, 'failed')]; destinationTableArn DIFFERS
  from the pre-delete replica ARN; ListTables(D, ns) afterwards: ['ackq_f81240_t']. (b) Then
  DeleteTable(replica) -> HTTP 204; status right after: 'failed'; PutTableReplication(identical, current
  token) -> HTTP 200; the watch stopped at once on the STALE 'failed' entry (0.05 s), so whether the re-Put
  re-creates the replica was NOT observed here (re-run in table/state-machine/repl-failed-recovery); final
  status body {"destinations": [{"destinationTableBucketArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest", "failureMessage": "A table with an
  identical name already exists at the destination.", "replicationStatus": "failed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"};
  ListTables(D, ns) at the end: [].
  - ACK: custom_create, requeue, pre-delete-cleanup · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTable, ListTables · fields: destinations[].destinationTableArn, destinations[].replicationStatus,
    destinations[].failureMessage
  - repro: DeleteTableReplication -> Put identical -> watch; DeleteTable(replica) -> Put identical (token) ->
    watch 7 min + ListTables(D)
  - measurements: re_put_to_failed_s=[61.99], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEBUCKETREPLICATION-031](tablebucket-replication.md#s3t-tablebucketreplication-031), [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-224](#s3t-table-224),
    [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008) · hypotheses: H-U-176, H-U-166, H-U-167 · evidence:
    table/state-machine/repl-status-lifecycle
  - notes: H-U-176 REFUTED: the leftover service-managed replica is NOT re-attached; the new configuration
    fails with 'A table with an identical name already exists at the destination.' (the same failure as a
    user-table collision), so after a Delete+re-Put cycle the controller must delete the leftover replica...
  - full notes: [details/S3T-TABLEREPLICATION-020.md](details/S3T-TABLEREPLICATION-020.md)

## Update granularity and ordering

- <a id="s3t-tablereplication-002"></a>**S3T-TABLEREPLICATION-002** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Nonexistent role: PutTableReplication -> 200; status 'failed: Insufficient permissions given to successfully complete replication.' ~61 s** (hypothesis refuted; behavior confirmed)
  PutTableReplication(configuration.role = ARN of a role that does not exist, existing destination bucket) ->
  HTTP 200. GetTableReplicationStatus polled every 5 s: 0.05s -> no-destinations; 60.61s ->
  failed:Insufficient permissions given to successfully complete replication.; final destinations[0] =
  'failed:Insufficient permissions given to successfully complete replication.'.
  - ACK: requeue, synced.when, references · ops: PutTableReplication, GetTableReplicationStatus · fields:
    configuration.role, destinations[].replicationStatus, destinations[].failureMessage
  - repro: PutTableReplication(role=arn of a missing role) -> poll GetTableReplicationStatus
  - measurements: nonexistent_role_to_terminal_status_s=[60.61], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEBUCKETREPLICATION-028](tablebucket-replication.md#s3t-tablebucketreplication-028),
    [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009) · hypotheses: H-U-133, H-S-054 · evidence:
    service/error-taxonomy/repl-permissions
  - notes: H-U-133 second half (nonexistent role -> 400) REFUTED: accepted with 200; H-S-054 at table level
    likewise refuted. Role validation is asynchronous: the status goes pending (destinations empty at +0) ->
    failed at +60.6 s with the generic message 'Insufficient permissions given to successfully...
  - full notes: [details/S3T-TABLEREPLICATION-002.md](details/S3T-TABLEREPLICATION-002.md)

- <a id="s3t-tablereplication-008"></a>**S3T-TABLEREPLICATION-008** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Role trusting only ec2: PutTableReplication -> 200, 'failed' at ~61 s; no self-heal in 150 s after fixing trust; re-Put -> completed**
  With the replication role's trust policy changed to allow only ec2.amazonaws.com (UpdateAssumeRolePolicy ->
  HTTP 200, 30 s wait): PutTableReplication(new empty table -> D) -> HTTP 200; GetTableReplication -> HTTP 200.
  PutTableBucketReplication(B2, 10 tables -> D) with the same role -> HTTP 200. GetTableReplicationStatus(b8)
  every 5 s: 0.03s -> no-destinations; 60.66s -> failed:Insufficient permissions given to successfully
  complete replication.. Bucket-level: status(c1) -> no-destinations; new tables in D within the window: [].
  After restoring the trust policy (no new Put): status stayed 'failed' for 150 s; a re-Put of the identical
  configuration (current token) -> HTTP 200 and the status then completed after 65.88s.
  - ACK: requeue, synced.when, docs-only · ops: PutTableReplication, PutTableBucketReplication,
    GetTableReplicationStatus, UpdateAssumeRolePolicy · fields: configuration.role,
    destinations[].replicationStatus, destinations[].failureMessage
  - repro: UpdateAssumeRolePolicy(role, ec2 only) -> PutTableReplication / PutTableBucketReplication -> poll
    status -> restore trust -> poll -> re-Put -> poll
  - measurements: bad_trust_to_terminal_status_s=[60.66], self_heal_s=[], after_re_put_s=[65.88], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEBUCKETREPLICATION-028](tablebucket-replication.md#s3t-tablebucketreplication-028),
    [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013),
    [S3T-TABLEBUCKETREPLICATION-030](tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020) · hypotheses: H-U-131,
    H-U-132, H-S-053 · evidence: tablebucket/sub-resources/repl-backfill-delete
  - notes: H-U-131 confirmed: Put 200 (configuration stored), failure visible only as status 'failed' at ~61 s
    with the generic message 'Insufficient permissions given to successfully complete replication.' (the
    message names neither the role nor AssumeRole; the same text appears for a nonexistent role)....
  - full notes: [details/S3T-TABLEREPLICATION-008.md](details/S3T-TABLEREPLICATION-008.md)

- <a id="s3t-tablereplication-025"></a>**S3T-TABLEREPLICATION-025** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Collision with a user table in D: Put 200, 'D:failed' (identical name exists), user table untouched; same-region dest/existing ns: completed**
  (t1) A user-created table with the source's namespace/name already exists in D: PutTableReplication -> HTTP
  200; status transitions [(0.1, 'no-destinations'), (62.0, 'D:failed')]; terminal entry {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "failureMessage": "A table with an identical name already exists at the destination.", "replicationStatus":
  "failed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/58ee5109-3a05-4669-a8e6-a78158cdf347"}.
  GetTable(user table in D) after: versionToken unchanged, managedByService=None, modifiedAt unchanged.
  ListTables(D, ns) -> [{"arn_is_user_table": true, "name": "ackq_9d291a_t1"}, {"arn_is_user_table": false,
  "name": "ackq_9d291a_t3"}]. (t2) Destination E in the SAME region as the source: Put -> HTTP 200;
  transitions [(1.1, 'no-destinations'), (71.2, 'E:pending'), (72.3, 'E:completed')]; GetTable(E replica) ->
  {"managedByService": "replication.s3tables.amazonaws.com", "managedTableInformation":
  {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/21e11778-a08e-4f39-a7f6-f0dd8e00f285"}},
  "type": "customer"}. (t3) The destination namespace already existed (created by the user) but not the table:
  Put -> HTTP 200; transitions [(2.2, 'no-destinations'), (66.1, 'D:pending'), (68.1, 'D:completed')];
  GetTable(D replica) -> {"managedByService": "replication.s3tables.amazonaws.com", "managedTableInformation":
  {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-9d291a-cr-src/table/f102c252-f65c-4530-a177-ae241e61e940"}},
  "type": "customer"}.
  - ACK: synced.when, requeue, pre-delete-cleanup · ops: PutTableReplication, GetTableReplicationStatus,
    GetTable, ListTables, CreateTable, CreateNamespace · fields: destinations[].replicationStatus,
    destinations[].failureMessage, managedByService
  - repro: CreateNamespace+CreateTable(same names) in D -> Put t1->D; Put t2->E(same region); Put t3->D
    (namespace exists); poll 1 call/s round-robin
  - measurements: t1_terminal_s=[62.0], t2_terminal_s=[72.3], t3_terminal_s=[68.1], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028),
    [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-224](#s3t-table-224) · hypotheses: H-U-177, H-U-149, H-U-150, H-U-162 · evidence:
    table/cross-region/repl-same-region-collision
  - notes: H-U-177 confirmed (accepted, then 'failed' without touching the user table). H-U-149 confirmed /
    H-U-150 refuted (same-region table-level destination). H-U-162 refuted again: a pre-existing namespace is
    simply reused ('D:completed').
  - full notes: [details/S3T-TABLEREPLICATION-025.md](details/S3T-TABLEREPLICATION-025.md)

- <a id="s3t-tablereplication-030"></a>**S3T-TABLEREPLICATION-030** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **SSE-KMS source -> AES256 dest and AES256 source -> SSE-KMS dest: both Put 200, 'completed' ~62-63s; replica encryption follows dest bucket** (hypothesis refuted; behavior confirmed)
  Source table tk in a bucket with PutTableBucketEncryption(aws:kms, customer key; key policy grants the
  replication role and the S3 Tables service principals): GetTableEncryption(tk) -> {"kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6c06ac3e-922e-4619-97f6-71c7ce6ee948", "sseAlgorithm": "aws:kms"}.
  PutTableReplication(tk -> EA, EA has {"sseAlgorithm": "AES256"}, same region) -> HTTP 200; status
  transitions [(0.1, 'no-destinations'), (62.2, 'EA:completed')] (failure message None).
  GetTableEncryption(replica of tk in EA) -> HTTP 200 {"sseAlgorithm": "AES256"}; GetTable(replica)
  {"managedByService": "replication.s3tables.amazonaws.com", "type": "customer"}. Source table tp (AES256
  bucket) -> EK (bucket with aws:kms, same key, same region): Put -> HTTP 200; transitions [(1.1,
  'no-destinations'), (61.1, 'EK:pending'), (63.2, 'EK:completed')] (failure message None);
  GetTableEncryption(replica of tp in EK) -> HTTP 200 {"kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/6c06ac3e-922e-4619-97f6-71c7ce6ee948", "sseAlgorithm": "aws:kms"};
  GetTable(replica) {"managedByService": "replication.s3tables.amazonaws.com", "type": "customer"}.
  - ACK: requeue, synced.when, late_initialize · ops: PutTableReplication, GetTableReplicationStatus,
    GetTableEncryption, PutTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn, destinations[].replicationStatus, destinations[].failureMessage
  - repro: KMS key (role + service principals) -> SSE-KMS source bucket + table -> Put to an AES256 bucket;
    AES256 table -> Put to an SSE-KMS bucket; poll; GetTableEncryption(replicas)
  - measurements: tk_terminal_s=[62.2], tp_terminal_s=[63.2], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](#s3t-table-136), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLE-224](#s3t-table-224) ·
    hypotheses: H-U-192 · evidence: table/cross-region/repl-kms-parity
  - notes: H-U-192 REFUTED: no synchronous encryption-parity check (both Puts 200). Replica encryption follows
    the DESTINATION bucket setting (AES256 / aws:kms); empty tables, one key, n=1 per direction (run 2; in run
    1 the AES->KMS direction also completed, and the KMS->AES direction toward the shard's...
  - full notes: [details/S3T-TABLEREPLICATION-030.md](details/S3T-TABLEREPLICATION-030.md)

- <a id="s3t-tablereplication-032"></a>**S3T-TABLEREPLICATION-032** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C071 FALSE: PutTableReplication accepts roles lacking permission (nonexistent/wrong trust/1 dest) -> 200; 'failed' at +61-66 s**
  PutTableReplication is accepted with HTTP 200 {versionToken, status 'success'} for a role that does not
  exist, for a role whose trust policy allows only ec2.amazonaws.com, and for a role allowed on only one of
  two destinations. The permission failure surfaces asynchronously in GetTableReplicationStatus: destinations
  [] for ~60 s, then replicationStatus 'failed' with failureMessage 'Insufficient permissions given to
  successfully complete replication.' at +61-66 s (generic text that names neither the role nor the denied
  action). No synchronous error is returned.
  - ACK: requeue, synced.when · ops: PutTableReplication, GetTableReplicationStatus · fields:
    configuration.role
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-002](#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028),
    [S3T-TABLEBUCKETREPLICATION-021](tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-033](#s3t-tablereplication-033), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005),
    [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017) · evidence: service/static/doc-claims-3
  - notes: VERDICT: FALSE - Put returns 200 for every bad-permission role tried (n=3 roles); the failure is
    only visible ~1 min later in the status.

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-136"></a>**S3T-TABLE-136** `read-gap` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replica table: GetTable has managedByService='replication.s3tables.amazonaws.com' and replicationInformation.sourceTableARN...**
  The replica created by the service in the destination bucket has the same namespace and table name as the
  source (yes); GetTable(replica) -> HTTP 200 with type='customer',
  managedByService='replication.s3tables.amazonaws.com',
  managedTableInformation.replicationInformation.sourceTableARN equals the source table ARN; metadataLocation
  is under the replica's own warehouseLocation with the same file name as the source.
  GetTableStorageClass(replica)=INTELLIGENT_TIERING while the source table is STANDARD and the destination
  bucket default is INTELLIGENT_TIERING; GetTableEncryption(replica)=AES256;
  GetTableMaintenanceConfiguration(replica) -> HTTP 200. On the DESTINATION bucket GetTableBucketReplication
  -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  GetTableReplication(replica) -> BadRequestException (HTTP 400) 'Replication operations are not allowed on
  service managed tables.'; GetTableReplicationStatus(replica) -> BadRequestException (HTTP 400) 'Replication
  operations are not allowed on service managed tables.'.
  - ACK: is_read_only, late_initialize, custom_find · ops: GetTable, GetTableStorageClass, GetTableEncryption,
    GetTableBucketReplication, GetTableReplication, GetTableReplicationStatus · fields: managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, type,
    storageClassConfiguration.storageClass
  - repro: PutTableReplication -> wait completed -> GetTable(destinationTableArn) and sub-resource reads in
    the destination region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-031](table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-223](#s3t-table-223), [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](table-replication.md#s3t-table-177), [S3T-TABLE-299](#s3t-table-299),
    [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-225](#s3t-table-225), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-196](#s3t-table-196), [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041),
    [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028) · hypotheses: H-S-117, H-S-059 · evidence: table/round-trip/repl-put-get
  - notes: H-S-117 confirmed: the bucket-level read on the destination is NotFoundException; the replica
    itself exposes its source through managedTableInformation.replicationInformation. H-S-059 qualified (no
    KMS tested): cross-region replication works and the replica took storage class INTELLIGENT_TIERING...
  - full notes: [details/S3T-TABLE-136.md](details/S3T-TABLE-136.md)

- <a id="s3t-table-224"></a>**S3T-TABLE-224** `immutable-field` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replica writes: Rename -> 2xx; PutMaintenance -> BadRequestException; PutExpiration -> MethodNotAllowedException; Tag -> 2xx...**
  On the service-created replica while the source configuration is active. By bucket+namespace+name:
  PutTableMaintenanceConfiguration(icebergCompaction) -> BadRequestException (HTTP 400) 'The specified
  maintenance type is not valid for service managed tables.'; PutTablePolicy -> HTTP 200;
  UpdateTableMetadataLocation -> BadRequestException (HTTP 400) 'Updates to service managed tables are not
  allowed.'; RenameTable -> HTTP 204 - the replica's GetTable.name became 'ackq_f81240_t_renamed'
  (managedByService still 'replication.s3tables.amazonaws.com'), ListTables(D) then ['ackq_f81240_t_renamed'],
  GetTableReplicationStatus(source) {"destinations": [{"destinationTableArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"};
  RenameTable back -> HTTP 204 (name 'ackq_f81240_t'). By table ARN:
  PutTableRecordExpirationConfiguration(enabled, 5 days) -> MethodNotAllowedException (HTTP 405) 'Table record
  expiration configuration API calls to this table are not allowed'; TagResource -> HTTP 200
  (ListTagsForResource then {"replkey": "replval"}); UntagResource -> HTTP 204; PutTableReplication(replica ->
  S) -> BadRequestException (HTTP 400) 'Replication operations are not allowed on service managed tables.';
  GetTableReplication -> BadRequestException (HTTP 400) 'Replication operations are not allowed on service
  managed tables.'; GetTableReplicationStatus -> BadRequestException (HTTP 400) 'Replication operations are
  not allowed on service managed tables.'; GetTableRecordExpirationConfiguration -> MethodNotAllowedException
  (HTTP 405) 'Table record expiration configuration API calls to this table are not allowed' null;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405) 'Table record expiration
  configuration API calls to this table are not allowed' null. ListTagsForResource(replica) before any tag
  write -> {} (source tags: {"ackq-run": "f81240", "srckey": "srcval"}).
  - ACK: is_read_only, terminal_codes, tags.custom-sync · ops: RenameTable, PutTableMaintenanceConfiguration,
    PutTableRecordExpirationConfiguration, TagResource, UntagResource, UpdateTableMetadataLocation,
    PutTableReplication, PutTablePolicy, GetTableReplication, GetTableReplicationStatus,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: managedByService, tags,
    configuration.status, metadataLocation, name
  - repro: replica present -> each write op once (by-name ops first, rename last and renamed back); reads on
    the replica ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-137](#s3t-table-137), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-225](#s3t-table-225), [S3T-TABLE-299](#s3t-table-299), [S3T-TABLE-223](#s3t-table-223), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-300](#s3t-table-300), [S3T-TABLE-306](#s3t-table-306),
    [S3T-TABLE-196](#s3t-table-196), [S3T-TABLE-147](table-record-expiration.md#s3t-table-147), [S3T-TABLEBUCKETREPLICATION-031](tablebucket-replication.md#s3t-tablebucketreplication-031), [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032),
    [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028) · hypotheses: H-U-165,
    H-U-194, H-U-199, H-U-158, H-U-159, H-U-178 · evidence: table/state-machine/repl-status-lifecycle
  - notes: H-U-165 REFUTED in part (RenameTable on a replica succeeds): codes per op {'rename': '2xx',
    'maintenance': 'BadRequestException', 'expiration': 'MethodNotAllowedException', 'commit':
    'BadRequestException', 'policy': '2xx'}. H-U-194 confirmed (tags writable on the replica, source tags NOT
    copied)....
  - full notes: [details/S3T-TABLE-224.md](details/S3T-TABLE-224.md)

- <a id="s3t-table-299"></a>**S3T-TABLE-299** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Replica copies the source's snapshot-management config but forces compaction disabled (512/auto); no tags, no aws: tags, no policy**
  Source table (empty, explicit INTELLIGENT_TIERING) primed before PutTableReplication with tags {"team":
  "data", "env": "lab", "ackq-run": "97f95f"}, policy Sid SrcAllowGet (Resource = source ARN) and maintenance
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256,
  "strategy": "binpack"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5, "maxSnapshotAgeHours": 48}}}}. Replica (same
  namespace/name in the destination bucket, status 'completed' at +66s): ListTagsForResource -> {} (aws: keys:
  []); GetTablePolicy -> NotFoundException (HTTP 404) 'The specified table policy does not exist.';
  GetTableMaintenanceConfiguration -> {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5,
  "maxSnapshotAgeHours": 48}}}} (equal to source: False); GetTableStorageClass -> {"storageClass": "STANDARD"}
  (destination bucket default STANDARD, source IT); GetTableEncryption -> {"sseAlgorithm": "AES256"}. Replica
  GetTable: createdBy == source createdBy, modifiedBy=None,
  managedByService=replication.s3tables.amazonaws.com. Destination namespace GetNamespace: createdBy == the
  source namespace's, keys ['createdAt', 'createdBy', 'namespace', 'namespaceId', 'ownerAccountId',
  'tableBucketId'].
  - ACK: tags.custom-sync, scope:skip, hook:sdk_read_one_post_set_output · ops: PutTableReplication,
    ListTagsForResource, GetTablePolicy, GetTableMaintenanceConfiguration, GetTable, GetNamespace · fields:
    tags, resourcePolicy, configuration.icebergCompaction, configuration.icebergSnapshotManagement, createdBy,
    managedByService
  - repro: CreateTable -> TagResource + PutTablePolicy + PutTableMaintenanceConfiguration ->
    PutTableReplication(-> D) -> wait completed -> read the replica's tags/policy/maintenance
  - measurements: completed_at_s=65.68, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-223](#s3t-table-223), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-177](table-replication.md#s3t-table-177), [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-300](#s3t-table-300),
    [S3T-TABLE-306](#s3t-table-306) · evidence: table/creative/replica-settings-fidelity
  - notes: Cross-service analogy (DynamoDB restore drops tags/policy/TTL; keeps SSE/billing): the replica is a
    partial copy. Corrects/extends [S3T-TABLE-223](#s3t-table-223) (whose note says the maintenance config was 'echoed': only
    icebergSnapshotManagement is copied; icebergCompaction is forced to status=disabled with...
  - full notes: [details/S3T-TABLE-299.md](details/S3T-TABLE-299.md)

- <a id="s3t-table-300"></a>**S3T-TABLE-300** `immutable-field` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Replica settings are a replication-time snapshot: later source maintenance and policy changes did not reach the replica in 184 s**
  With the destination 'completed': source PutTableMaintenanceConfiguration compaction 384/binpack -> HTTP
  204, snapshots disabled 3/72 -> HTTP 204, TagResource {stage: v2} -> InternalServerErrorException (HTTP 500)
  'An internal error occurred. Try again.', PutTablePolicy (Sid SrcAllowGetV2) -> HTTP 200. Replica polled
  every 5 s (tags, policy, maintenance) for 184s: 36 samples, first change none; final replica vs pre-change
  snapshot differs in []; final replica maintenance {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5,
  "maxSnapshotAgeHours": 48}}}} vs source {"icebergCompaction": {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 384, "strategy": "binpack"}}}, "icebergSnapshotManagement":
  {"status": "disabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 3,
  "maxSnapshotAgeHours": 72}}}}. GetTableReplicationStatus during the window: [(0.2, 'completed')].
  - ACK: requeue, scope:skip · ops: PutTableMaintenanceConfiguration, TagResource, PutTablePolicy,
    GetTableMaintenanceConfiguration, ListTagsForResource, GetTablePolicy, GetTableReplicationStatus · fields:
    configuration, tags, resourcePolicy, destinations[].replicationStatus
  - repro: replica completed -> change source maintenance config + tags + policy -> poll the replica's reads
    every 5 s for 3 min
  - measurements: window_s=184.3, first_change_s=null, n=36
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006), [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-299](#s3t-table-299), [S3T-TABLE-306](#s3t-table-306), [S3T-SERVICE-007](service.md#s3t-service-007),
    [S3T-SERVICE-020](service.md#s3t-service-020), [S3T-TABLE-304](table-replication.md#s3t-table-304) · evidence: table/creative/replica-settings-fidelity
  - notes: Cross-service analogy (ElastiCache global datastore propagates parameter changes; DynamoDB replica
    settings are per-region): tells a controller whether managing the source's sub-resources implicitly
    manages the replica's. Config-only changes (no metadata commit) on an empty table; a commit-triggered...
  - full notes: [details/S3T-TABLE-300.md](details/S3T-TABLE-300.md)

- <a id="s3t-table-306"></a>**S3T-TABLE-306** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **RenameTable (name, then namespace) on a replicated source: replica keeps the OLD name/namespace for 490 s; status completed; re-Put no-op**
  Empty source table t replicated to D (completed at +66s; D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}).
  RenameTable(newName=t2) -> HTTP 204. Over 120 s polling D every 5 s: transitions [[0.1, "completed",
  {"ackq_7b3249_ns": ["ackq_7b3249_t"]}]]; final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]};
  GetTable(replica ARN).name == new name: False. Then identical PutTableReplication (current token) -> HTTP
  200, status right after 'completed'; over 245s: [[0.1, "completed", {"ackq_7b3249_ns": ["ackq_7b3249_t"]}]];
  final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}; final status entries [{"replicationStatus": "completed",
  "destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3249-rfr-dst/table/b3c876c2-2d35-4a11-b88c-a185d682f40e"}].
  Then RenameTable(newNamespaceName=ns2) -> HTTP 204; over 122s: [[0.1, "completed", {"ackq_7b3249_ns":
  ["ackq_7b3249_t"]}]]; final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}.
  - ACK: custom_update, pre-delete-cleanup, requeue · ops: RenameTable, PutTableReplication,
    GetTableReplicationStatus, ListTables, GetTable · fields: newName, newNamespaceName,
    destinations[].destinationTableArn, destinations[].replicationStatus
  - repro: PutTableReplication(empty table -> D) -> completed -> RenameTable(newName) -> poll D 120 s ->
    re-Put identical -> poll 240 s -> RenameTable(newNamespaceName) -> poll 120 s
  - measurements: initial_completed_s=65.79, r1_window_s=122.3, r2_window_s=244.6, r3_window_s=122.3, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-224](#s3t-table-224), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-299](#s3t-table-299), [S3T-TABLE-300](#s3t-table-300),
    [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118) · evidence: table/creative/replica-follows-rename
  - notes: Cross-service analogy (DynamoDB replicas are addressed by the same table name; SNS has no rename):
    the controller's only Update (RenameTable) interacts with a service-managed replica. Extends [S3T-TABLE-224](#s3t-table-224)
    (rename of the replica itself is allowed) and [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020) (re-Put with a leftover...
  - full notes: [details/S3T-TABLE-306.md](details/S3T-TABLE-306.md)

- <a id="s3t-tablereplication-021"></a>**S3T-TABLEREPLICATION-021** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table under bucket-level replication: GetTableReplication 404, Status 200; table-level Put (no token) -> 200, replaces bucket dest in status**
  PutTableBucketReplication(S -> D) -> HTTP 200; t1 (existing empty table) status transitions [(0.08,
  'no-destinations'), (1.12, 'D:pending'), (2.16, 'D:completed')]. Under bucket-level only:
  GetTableReplication(t1) -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; GetTableReplicationStatus(t1) -> HTTP 200. PutTableReplication(t1 -> E, no token) -> HTTP 200.
  GetTableReplication(t1) then -> 200 {"role": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>", "rules":
  [{"destinations": [{"destinationTableBucketARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same"}]}]}; GetTableBucketReplication(S)
  unchanged -> 200 dests ['D']. Status transitions after the table-level Put: [(0.03, 'D:completed'), (61.04,
  'D:completed|E:pending'), (62.07, 'E:completed')]. ListTables(E, ns) afterwards: ['ackq_04fb03_t1'];
  ListTables(D, ns): ['ackq_04fb03_t1'].
  - ACK: custom_find, synced.when, compare.is_ignored+delta_pre_compare · ops: PutTableBucketReplication,
    GetTableReplication, GetTableReplicationStatus, PutTableReplication, ListTables · fields:
    configuration.rules[].destinations, destinations[].destinationTableBucketArn,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication(S->D) -> status(t1) terminal -> GetTableReplication(t1) ->
    PutTableReplication(t1->E) -> poll status -> ListTables(E), ListTables(D)
  - measurements: bucket_level_t1_terminal_s=[2.16], table_level_e_terminal_s=[62.07], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008),
    [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-014](tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) · hypotheses:
    H-U-179, H-U-180, H-U-181 · evidence: table/sub-resources/repl-bucket-vs-table-config
  - notes: H-U-179 confirmed (GetTableReplication 404 under bucket-level only, GetTableReplicationStatus 200
    with the bucket-level destination). H-U-181 REFUTED (table-level Put accepted with HTTP 200, no 409).
    H-U-180 confirmed for the empty table: after the table-level Put [E] the status showed...
  - full notes: [details/S3T-TABLEREPLICATION-021.md](details/S3T-TABLEREPLICATION-021.md)

- <a id="s3t-tablereplication-022"></a>**S3T-TABLEREPLICATION-022** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table-level Put [D,E] where D holds the bucket-level replica -> D 'failed' (identical name exists); table-level delete -> bucket view**
  CreateTable t2 while S has bucket-level replication to D (and t1 has a table-level config to E): status(t2)
  transitions [(0.06, 'no-destinations'), (1.1, 'D:pending'), (2.13, 'D:completed')]; GetTableReplication(t2)
  -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'; ListTables(E) =
  ['ackq_04fb03_t1'] (t2 not sent to the table-level destination of t1). Table-level Put [D, E] on t1 (D also
  a bucket-level destination) -> HTTP 200; status transitions [(0.06, 'E:completed'), (61.17,
  'D:pending|E:completed'), (62.2, 'D:failed|E:completed')]; status body {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "failureMessage": "A table with an identical name already exists at the destination.", "replicationStatus":
  "failed"}, {"destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same/table/eb9e8561-0ee2-4507-a849-c3a42cf6d0b0",
  "destinationTableBucketArn": "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-src/table/0e1ad48c-783d-4b4a-aecc-4bebbd6e9103"}.
  DeleteTableReplication(t1) -> HTTP 204; status(t1) over 60 s: [(0.1, 'D:failed|E:completed'), (61.3,
  'D:failed')]; GetTableReplication(t1) -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'. DeleteTableBucketReplication(S) -> HTTP 204; status(t1)/status(t2) over 30
  s: [(0.1, 't1=ERR:NotFoundException t2=ERR:NotFoundException')]. Final ListTables: D ['ackq_04fb03_t1',
  'ackq_04fb03_t2'], E ['ackq_04fb03_t1'].
  - ACK: custom_delete, synced.when, pre-delete-cleanup · ops: CreateTable, GetTableReplicationStatus,
    PutTableReplication, DeleteTableReplication, DeleteTableBucketReplication · fields:
    destinations[].destinationTableBucketArn, destinations[].replicationStatus
  - repro: bucket-level S->D + table-level t1->E -> CreateTable t2 -> status; Put t1 [D,E] -> status;
    DeleteTableReplication(t1) -> status x60 s; DeleteTableBucketReplication -> status x30 s
  - measurements: new_table_terminal_s=[2.13], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-014](tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017),
    [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043), [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) · hypotheses: H-U-180 · evidence:
    table/sub-resources/repl-bucket-vs-table-config
  - notes: Interplay cells (n=1 each): (1) a table-level destination that is also the bucket-level destination
    does NOT re-attach to the bucket-level replica - it fails with 'A table with an identical name already
    exists at the destination.' while the other destination completes; (2) after...
  - full notes: [details/S3T-TABLEREPLICATION-022.md](details/S3T-TABLEREPLICATION-022.md)

## Response fidelity and consistency

- <a id="s3t-table-223"></a>**S3T-TABLE-223** `response-fidelity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTable(replica) vs GetTable(source): type='customer', managedByService='replication.s3tables.amazonaws.com'...**
  Replica GetTable: type='customer', managedByService='replication.s3tables.amazonaws.com',
  managedTableInformation={"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}},
  format='ICEBERG'; fields equal to the source: ['createdBy', 'format', 'name', 'namespace', 'ownerAccountId',
  'type']; fields that differ: ['createdAt', 'managedByService', 'managedTableInformation', 'modifiedAt',
  'namespaceId', 'tableARN', 'tableBucketId', 'versionToken', 'warehouseLocation']. Source GetTable has
  managedByService=None and managedTableInformation=None. metadataLocation: source None, replica None.
  ListTables(D) entry for the replica: {"createdAt": "2026-10-13 23:41:41.473835+00:00", "managedByService":
  "replication.s3tables.amazonaws.com", "modifiedAt": "2026-10-13 23:41:41.669341+00:00", "name":
  "ackq_f81240_t", "namespace": ["ackq_f81240_ns"], "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "type": "customer"}. Replica sub-resource reads: storage class {"storageClassConfiguration":
  {"storageClass": "STANDARD"}}; encryption {"encryptionConfiguration": {"sseAlgorithm": "AES256"}};
  maintenance {"configuration": {"icebergCompaction": {"settings": {"icebergCompaction": {"strategy": "auto",
  "targetFileSizeMB": 512}}, "status": "disabled"}, "icebergSnapshotManagement": {"settings":
  {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 48, "minSnapshotsToKeep": 2}}, "status": "enabled"}},
  "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8"};
  metadata location {"versionToken": "c52faa2fa3f8cd457bbe", "warehouseLocation":
  "s3://93cf68ab-77f5-416b-zgdhmbc3kh1cb9n9eq6a3aenwqidouse1b--table-s3"}.
  - ACK: is_read_only, custom_find, late_initialize · ops: GetTable, ListTables, GetTableStorageClass,
    GetTableEncryption, GetTableMaintenanceConfiguration · fields: type, managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, metadataLocation, warehouseLocation,
    versionToken
  - repro: replication terminal -> GetTable(source) and GetTable(destinationTableArn); diff all members
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-031](table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](table-replication.md#s3t-table-177), [S3T-TABLE-299](#s3t-table-299),
    [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-307](#s3t-table-307) · hypotheses: H-U-163, H-U-164, H-U-193 · evidence:
    table/state-machine/repl-status-lifecycle
  - notes: H-U-163 qualified: managedByService is the STRING 'replication.s3tables.amazonaws.com' (not a
    boolean) and managedTableInformation.replicationInformation.sourceTableARN names the source; type stays
    'customer'. H-U-164 refuted (type is in the enum). H-U-193 confirmed: own...
  - full notes: [details/S3T-TABLE-223.md](details/S3T-TABLE-223.md)

- <a id="s3t-tablereplication-006"></a>**S3T-TABLEREPLICATION-006** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **A source commit (UpdateTableMetadataLocation) reaches the cross-region replica after 403-428 s; status drops back to 'pending' meanwhile**
  UpdateTableMetadataLocation on the source (new metadata.json in the warehouse) -> 200.
  GetTableReplicationStatus then shows replicationStatus='pending' again while
  lastSuccessfulReplicatedUpdate.metadataLocation still names the previously replicated file; it switched to
  the new file and 'completed' after 403 s and 428 s for two consecutive commits (two commits queued up were
  applied in order). GetTable(replica).metadataLocation then ends with the same file name under the replica's
  own warehouse. For comparison, the initial backfill of this schema-only table took 436 s.
  - ACK: e2e-timing, synced.when · ops: UpdateTableMetadataLocation, GetTableReplicationStatus, GetTable ·
    fields: destinations[].lastSuccessfulReplicatedUpdate.metadataLocation, metadataLocation
  - repro: PutTableReplication -> completed -> UpdateTableMetadataLocation(source) -> poll
    GetTableReplicationStatus
  - measurements: commit_to_replica_s=[403.0, 428.0], initial_pending_to_completed_s=[435.7], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-029](tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029),
    [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031) · hypotheses: H-S-058 · evidence: table/round-trip/repl-put-get
  - notes: n=2 commits on one table. 'completed' is therefore not a stable state: every source commit flips
    the destination back to 'pending' for several minutes; a controller must not treat pending as an error and
    must not expect sub-minute propagation.

- <a id="s3t-tablereplication-017"></a>**S3T-TABLEREPLICATION-017** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Switch [D]->[E] (same region, empty table, n=1): Get reflects it at once; status showed E at +1.06s, dropped D and E 'completed' at +2.1s**
  With D's entry D:completed (63 s after the first Put), PutTableReplication [E] with the current token ->
  HTTP 200; GetTableReplication 181 ms later showed ['E']. GetTableReplicationStatus polled at 1 s:
  transitions [(0.02, 'D:completed'), (1.06, 'D:completed|E:pending'), (2.1, 'E:completed')] (D entry gone at
  +2.1s, E first seen at +1.06s, E completed at +2.1s; final view 'E:completed').
  - ACK: synced.when, custom_update · ops: PutTableReplication, GetTableReplication, GetTableReplicationStatus
    · fields: destinations[].replicationStatus, destinations[].destinationTableBucketArn
  - repro: config [D] completed -> Put [E] (token) -> Get -> poll Status at 1 s for <= 7 min
  - measurements: status_drops_old_dest_s=[2.1], status_shows_new_dest_s=[1.06], new_dest_completed_s=[2.1],
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-033](#s3t-tablereplication-033), [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028), [S3T-TABLEREPLICATION-032](#s3t-tablereplication-032), [S3T-TABLEBUCKETREPLICATION-022](tablebucket-replication.md#s3t-tablebucketreplication-022),
    [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-011](tablebucket-replication.md#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) ·
    hypotheses: H-U-160 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-160 second half confirmed (hypothesis: old set visible up to 60 s). Empty table, same-region
    new destination; n=1.
  - full notes: [details/S3T-TABLEREPLICATION-017.md](details/S3T-TABLEREPLICATION-017.md)

- <a id="s3t-tablereplication-042"></a>**S3T-TABLEREPLICATION-042** `stale-response` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Re-pointing after a lost destination: Put [D2] -> D2 pending at +61 s, completed at +66 s, dead D entry gone; Put [dead D, D2] lists only D2**
  After DeleteTableBucket(D) (replica + namespace removed first) the status read 'DST=completed'.
  CreateTableBucket(D2) then PutTableReplication([D2], current token) -> HTTP 200; status right after:
  'DST=completed'; transitions while watching (5 s): [[0.02, "DST=completed"], [60.68,
  "DST=completed;DST2=pending"], [65.73, "DST2=completed"]]; final 'DST2=completed'; D2 contents
  {"ackq_62668e_ns": ["ackq_62668e_t"]}. Then PutTableReplication([D, D2]) with D deleted -> HTTP 200; status
  right after 'DST2=completed', after 30 s 'DST2=completed'; raw entries [{"replicationStatus": "completed",
  "destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-62668e-ldh-dst2/table/f4862249-85b2-4af0-9642-c6881a826dc1",
  "dest": "DST2"}].
  - ACK: requeue, custom_update, terminal_codes · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTableBucket, ListTables · fields: destinations[].replicationStatus, destinations[].failureMessage,
    destinations[].destinationTableBucketArn
  - repro: completed replication -> delete destination bucket -> PutTableReplication([new bucket]) -> watch
    per-destination status -> PutTableReplication([dead, new])
  - measurements: d2_completed_s=65.73, heal_window_s=360, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-040](#s3t-tablereplication-040), [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027),
    [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-012](table-replication.md#s3t-tablereplication-012), [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEREPLICATION-047](#s3t-tablereplication-047), [S3T-TABLEREPLICATION-048](#s3t-tablereplication-048) · evidence:
    tablereplication/creative/lost-destination-heal
  - notes: Follow-up to [S3T-TABLEREPLICATION-040](#s3t-tablereplication-040) (status 'failed: The specified destination bucket does not
    exist.' at once after the destination is deleted; Put accepted). Cross-service analogy
    (ElastiCache/DynamoDB: a deleted dependency must be re-pointed by the controller): whether re-pointing is
    enough and...
  - full notes: [details/S3T-TABLEREPLICATION-042.md](details/S3T-TABLEREPLICATION-042.md)

- <a id="s3t-tablereplication-043"></a>**S3T-TABLEREPLICATION-043** `stale-response` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Corrective re-Put on a 'failed' collision dest: stale 'failed'+old message 61 s, then 'completed' directly; no self-heal without Put (300 s)**
  Destination collision (customer table with the replica's name): PutTableReplication -> HTTP 200 status
  'success'; GetTableReplicationStatus at +59 ms -> []; then []@0.03s -> failed+msg@61.15s. Deleting the
  colliding table alone: status stayed 'failed' with the old message for 320 s (a 300 s watch in an earlier
  run of this probe plus 20 s here) - no self-heal. Re-Put of the unchanged configuration with the current
  token -> HTTP 200 (token rotated: True); status right after (+0.06 s) -> failed+msg; 1 s polling:
  failed+msg@0.02s -> completed@61.17s. A controller that reads the status right after its re-Put sees the
  stale 'failed' with the old failureMessage and would wrongly report failure (or re-Put again).
  - ACK: requeue, synced.when, custom_update · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTable · fields: destinations.replicationStatus, destinations.failureMessage, versionToken
  - repro: Create table in D with the source's ns/name -> PutTableReplication -> watch failed ->
    DeleteTable(D) -> wait 20 s -> PutTableReplication(current token) -> poll status 1/s
  - measurements: failed_to_first_at_s=61.15, stale_failed_after_reput_s=61.17, completed_after_reput_s=61.17,
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025), [S3T-TABLEREPLICATION-026](#s3t-tablereplication-026), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027),
    [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029), [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028),
    [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008) · evidence: tablereplication/creative/controller-loop
  - notes: extends [S3T-TABLEREPLICATION-026](#s3t-tablereplication-026)/-027 with what the status shows BETWEEN the corrective Put and
    'completed'. Decision rule: after any Put, ignore 'failed' entries for at least the measured stale window
    (or until the entry changes) before setting a failure condition; 'completed' is the only state...
  - full notes: [details/S3T-TABLEREPLICATION-043.md](details/S3T-TABLEREPLICATION-043.md)

- <a id="s3t-tablereplication-045"></a>**S3T-TABLEREPLICATION-045** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **GetTableReplicationStatus right after PutTableReplication -> 200 with destinations [] for ~60 s...**
  PutTableReplication -> HTTP 200 status 'success'; GetTableReplicationStatus at +0 s -> HTTP 200
  {sourceTableArn, destinations: []} (no destination entry and no error). Polled every 5 s, the destination
  entry appeared with replicationStatus 'pending' (destinationTableBucketArn and destinationTableArn present,
  no lastSuccessfulReplicatedUpdate) at 60.7 s and was still 'pending' at 300 s, while ListTables on the
  destination bucket already listed the replica (same namespace and table name as the source, type
  'customer').
  - ACK: requeue, hook:sdk_read_one_post_set_output · ops: PutTableReplication, GetTableReplicationStatus,
    ListTables · fields: destinations, destinations.replicationStatus, destinations.destinationTableArn
  - repro: CreateTable(schema) -> PutTableReplication -> GetTableReplicationStatus immediately, then every 5 s
    for 300 s; ListTables(destination)
  - measurements: status_destinations_empty_for_s=60.7, still_pending_at_s=300, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003) · evidence: tablereplication/creative/reverify-set-e-repl
  - notes: Precision gain while re-verifying [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003) (whose Put/Get clauses all reproduced):
    a controller reading status right after Put sees an empty destinations list, not pending; single run,
    timings are one sample.
  - full notes: [details/S3T-TABLEREPLICATION-045.md](details/S3T-TABLEREPLICATION-045.md)

- <a id="s3t-tablereplication-047"></a>**S3T-TABLEREPLICATION-047** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Replacing the destination of a completed table-level config: new entry 'pending' +61-64 s after EVERY Put, completed ~1 s later (n=6)**
  Empty ICEBERG table, same-region destination buckets, GetTableReplicationStatus polled at 1 s. First Put
  [D0]: D0 entry at +61.06 s, completed at +62.1 s. Replacement Puts (current versionToken, HTTP 200 each):
  Put at +5 s after last change (67 s after first Put): D1 entry at +61.2 s (pending), completed at 62.24 s,
  old gone at 62.24 s; Put at +20 s after last change (149 s after first Put): D2 entry at +61.0 s (pending),
  completed at 62.04 s, old gone at 62.04 s; Put at +40 s after last change (251 s after first Put): D3 entry
  at +61.9 s (pending), completed at 62.94 s, old gone at 61.91 s; Put at +55 s after last change (369 s after
  first Put): D4 entry at +61.9 s (pending), completed at 62.96 s, old gone at 61.93 s; Put at +70 s after
  last change (502 s after first Put): D5 entry at +61.1 s (pending), completed at 62.15 s, old gone at 61.12
  s; Put at +1 s after last change (565 s after first Put): D6 entry at +64.1 s (pending), completed at 65.15
  s, old gone at 65.15 s. The old entry disappears when the new one reaches 'completed' in 3/6 cells.
  - ACK: synced.when, requeue, e2e-timing · ops: PutTableReplication, GetTableReplicationStatus,
    GetTableReplication · fields: destinations[].replicationStatus, destinations[].destinationTableBucketArn,
    configuration.rules[].destinations[]
  - repro: PutTableReplication [D0] on an empty table -> poll status 1 s to completed -> at +5/20/40/55/70/1 s
    after the last status change PutTableReplication [Dk] (fresh bucket, current token) -> poll status 1 s;
    record first Dk entry, Dk completed, old entry gone
  - measurements: first_put_entry_s=61.06, first_put_completed_s=62.1, offsets_after_last_change_s=[4.9,
    19.97, 40.12, 55.08, 70.08, 0.87], new_entry_after_put_s=[61.21, 61.0, 61.91, 61.93, 61.12, 64.12],
    completed_after_put_s=[62.24, 62.04, 62.94, 62.96, 62.15, 65.15], new_entry_since_first_put_s=[128.21,
    210.21, 313.28, 431.32, 563.56, 629.58], n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022) · evidence: tablereplication/state-machine/fu2-04-put-offset-tick
  - notes: FU2-04: 042/021/022 hold; 017's +1.06 s did not reproduce. The delay is a fixed ~61 s per Put, not
    a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last observed status change (67-565 s
    after the first Put) all showed the new destination as 'pending' 61.0-64.1 s later and...
  - full notes: [details/S3T-TABLEREPLICATION-047.md](details/S3T-TABLEREPLICATION-047.md)

## Delete semantics

- <a id="s3t-table-138"></a>**S3T-TABLE-138** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable(replica) with source config active -> 204; status stays 'completed' until the next source commit turns it 'failed'; no re-create**
  DeleteTable on the service-created replica (destination region) while the source table's replication
  configuration is active -> HTTP 204 (replicas are deletable). GetTableReplicationStatus(source) still said
  'completed' right after and 60 s later; the replica did not come back on its own. After a new
  UpdateTableMetadataLocation commit on the source, ListTables(destination) polled every 5 s for 300 s never
  showed the replica again, and the status was then 'failed' with failureMessage 'The destination table has
  been deleted during replication.' (destinationTableArn and lastSuccessfulReplicatedUpdate still present).
  - ACK: pre-delete-cleanup, custom_delete · ops: DeleteTable, GetTableReplicationStatus,
    UpdateTableMetadataLocation · fields: destinations[].replicationStatus, destinations[].failureMessage
  - repro: replica completed -> DeleteTable(replica) -> status; sleep 60; commit on source -> poll
    ListTables(destination) 5 min
  - measurements: replica_not_recreated_within_s=[300], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](#s3t-table-139), [S3T-TABLE-196](#s3t-table-196), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEBUCKET-012](tablebucket-replication.md#s3t-tablebucket-012), [S3T-TABLE-306](#s3t-table-306),
    [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032),
    [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020), [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022), [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043) ·
    hypotheses: H-S-060 · evidence: table/round-trip/repl-put-get
  - notes: H-S-060 second half qualified: the replica is deletable, is NOT re-created on its own, and a later
    source commit does NOT re-create it either - it moves the status to a terminal 'failed'. Recovery
    presumably needs a new Put (not tested). Deleting replicas while replication is configured is therefore...
  - full notes: [details/S3T-TABLE-138.md](details/S3T-TABLE-138.md)

- <a id="s3t-table-139"></a>**S3T-TABLE-139** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Destination cleanup: DeleteTable on 7 service-created replicas -> 204 each; DeleteNamespace -> 204; DeleteTableBucket(destination) -> 204**
  Tearing the destination bucket down after the source configurations were deleted: per replica DeleteTable
  results [{"table": "ackq_19dd08_b1", "ok": true, "code": null, "msg": ""}, {"table": "ackq_19dd08_b2", "ok":
  true, "code": null, "msg": ""}, {"table": "ackq_19dd08_b3", "ok": true, "code": null, "msg": ""}, {"table":
  "ackq_19dd08_b4", "ok": true, "code": null, "msg": ""}, {"table": "ackq_19dd08_b5", "ok": true, "code":
  null, "msg": ""}, {"table": "ackq_19dd08_b6", "ok": true, "code": null, "msg": ""}, {"table":
  "ackq_19dd08_b8", "ok": true, "code": null, "msg": ""}]; DeleteNamespace results [{"namespace":
  "ackq_19dd08_ns", "ok": true, "code": null}]; DeleteTableBucket -> HTTP 204.
  - ACK: pre-delete-cleanup · ops: DeleteTable, DeleteNamespace, DeleteTableBucket
  - repro: replicas present -> DeleteTable each -> DeleteNamespace -> DeleteTableBucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](#s3t-table-138), [S3T-TABLE-196](#s3t-table-196), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEBUCKET-012](tablebucket-replication.md#s3t-tablebucket-012) · hypotheses: H-S-060,
    H-S-061 · evidence: tablebucket/sub-resources/repl-backfill-delete
  - notes: Replicas are deletable by the destination owner; the destination bucket can be deleted once they
    are gone (recorded for e2e cleanup).

- <a id="s3t-table-195"></a>**S3T-TABLE-195** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable(source) with a replication config: at no-destinations -> 204, pending -> 204, completed -> 204, failed -> 204** (hypothesis refuted; behavior confirmed)
  Empty source tables each with their own PutTableReplication: DeleteTable(source) issued while
  GetTableReplicationStatus showed 'no-destinations' -> HTTP 204; while 'pending' -> HTTP 204; while
  'completed' -> HTTP 204; while 'failed' (nonexistent destination bucket) -> HTTP 204. After a successful
  DeleteTable(source): GetTableReplication(old table ARN) -> NotFoundException (HTTP 404) 'The specified table
  does not exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: pre-delete-cleanup, custom_delete · ops: DeleteTable, DeleteTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: destinations[].replicationStatus
  - repro: Put on 4 empty tables -> DeleteTable(source) at each observed status value -> reads on the deleted
    ARN
  - measurements: td_put_to_terminal_s=[6.3], te_put_to_failed_s=[0.06], tc_put_to_first_entry_s=[66.25], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-225](#s3t-table-225), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-196](#s3t-table-196) · hypotheses: H-U-168, H-U-169 · evidence:
    table/dependencies/repl-source-delete
  - notes: H-U-168 REFUTED: DeleteTable(source) succeeds with an active configuration in every state tried;
    H-U-169 confirmed in part (see the leftover finding for the replica's fate). States at delete time: {'ta':
    'no-destinations', 'tc': 'pending', 'te': 'failed', 'td': 'completed'}.

- <a id="s3t-table-196"></a>**S3T-TABLE-196** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable(source) the replicas persist (+0/+90 s) as read-only service-managed tables: commit -> 400; DeleteTable(replica) -> 204**
  ListTables(D, ns) right after the last source delete: ['tc', 'td']; 90 s later: ['tc', 'td'] (ta = deleted
  at no-destinations, tb = config deleted at +2 s, tc = deleted at 'pending', td = deleted at completed, te =
  nonexistent destination). Per leftover replica: {"tc": {"commit": "BadRequestException (HTTP 400) 'Updates
  to service managed tables are not allowed.'", "delete": "HTTP 204", "managedByService":
  "replication.s3tables.amazonaws.com", "mti_names_deleted_source": true}, "td": {"commit":
  "BadRequestException (HTTP 400) 'Updates to service managed tables are not allowed.'", "delete": "HTTP 204",
  "managedByService": "replication.s3tables.amazonaws.com", "mti_names_deleted_source": true}}. ListTables
  after deleting them: [].
  - ACK: pre-delete-cleanup, is_read_only · ops: ListTables, GetTable, UpdateTableMetadataLocation,
    DeleteTable · fields: managedByService, managedTableInformation
  - repro: after the deletes: ListTables(D) at +0 and +90 s ->
    GetTable/UpdateTableMetadataLocation/DeleteTable on each leftover
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](#s3t-table-139), [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-225](#s3t-table-225),
    [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-195](#s3t-table-195), [S3T-TABLE-138](#s3t-table-138), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEBUCKET-012](tablebucket-replication.md#s3t-tablebucket-012),
    [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLE-329](#s3t-table-329) · hypotheses: H-U-169, H-U-170, H-U-198 · evidence:
    table/dependencies/repl-source-delete
  - notes: H-U-169 orphan half confirmed: the replica of a deleted source remains service-managed and
    read-only but IS deletable (HTTP 204). H-U-198 confirmed: DeleteTableReplication 2 s after the Put -> HTTP
    204, status -> NotFoundException, replica of tb never appeared. H-U-170 (replica deletable after the...
  - full notes: [details/S3T-TABLE-196.md](details/S3T-TABLE-196.md)

- <a id="s3t-table-225"></a>**S3T-TABLE-225** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableReplication the replica stays managedByService='replication.s3tables.amazonaws.com' for 2 min...**
  DeleteTableReplication(current token) -> HTTP 204; GetTableReplicationStatus(source) right after ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'. GetTable(replica)
  every 10 s for 2 min: [(0.1, 'replication.s3tables.amazonaws.com|mti')]. Then
  UpdateTableMetadataLocation(replica) -> BadRequestException (HTTP 400) 'Updates to service managed tables
  are not allowed.'; PutTableRecordExpirationConfiguration(replica) -> MethodNotAllowedException (HTTP 405)
  'Table record expiration configuration API calls to this table are not allowed'.
  - ACK: pre-delete-cleanup, is_read_only, docs-only · ops: DeleteTableReplication, GetTable,
    UpdateTableMetadataLocation, PutTableRecordExpirationConfiguration, GetTableReplicationStatus · fields:
    managedByService, managedTableInformation
  - repro: completed replication -> DeleteTableReplication -> GetTable(replica) @10 s x 2 min -> write ops on
    the replica
  - measurements: replica_managed_observed_s=[0.1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-024](tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](tablebucket-replication.md#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-224](#s3t-table-224), [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-196](#s3t-table-196), [S3T-TABLE-195](#s3t-table-195), [S3T-TABLE-147](table-record-expiration.md#s3t-table-147),
    [S3T-TABLE-329](#s3t-table-329) · hypotheses: H-U-170, H-U-171, H-U-172 · evidence:
    table/state-machine/repl-status-lifecycle
  - notes: H-U-170 confirmed at table level (replica remains, still service-managed, writes rejected;
    DeleteTable(replica) succeeded later in this probe). H-U-171 REFUTED (no promotion within 2 min). H-U-172
    qualified: status -> 404 immediately; the no-propagation half was shown at bucket level...
  - full notes: [details/S3T-TABLE-225.md](details/S3T-TABLE-225.md)

- <a id="s3t-table-307"></a>**S3T-TABLE-307** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Orphaned replica after DeleteTableReplication: still managedByService for 874 s (30 samples); commit/maintenance 400...**
  DeleteTableReplication(current token) -> HTTP 204. GetTable(replica ARN) every 30 s for 874 s:
  managedByService=replication.s3tables.amazonaws.com, replicationInformation present=False, type=customer on
  every sample. At the end: UpdateTableMetadataLocation(replica) -> BadRequestException/400 'Updates to
  service managed tables are not allowed.'; PutTableMaintenanceConfiguration -> BadRequestException/400 'The
  specified maintenance type is not valid for service managed tables.'; PutTableReplication with the replica
  as source -> BadRequestException/400 'Replication operations are not allowed on service managed tables.'.
  - ACK: is_read_only, custom_delete, pre-delete-cleanup · ops: DeleteTableReplication, GetTable,
    UpdateTableMetadataLocation, PutTableMaintenanceConfiguration, PutTableReplication · fields:
    managedByService, replicationInformation
  - repro: table replication completed -> DeleteTableReplication -> GetTable(replica) every 30 s for 15 min ->
    write attempts on the replica
  - measurements: watched_s=874, samples=30, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-225](#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLE-299](#s3t-table-299), [S3T-TABLE-223](#s3t-table-223), [S3T-TABLE-224](#s3t-table-224),
    [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-196](#s3t-table-196), [S3T-TABLE-195](#s3t-table-195), [S3T-TABLE-329](#s3t-table-329) · evidence:
    tablereplication/creative/controller-loop
  - notes: extends [S3T-TABLE-225](#s3t-table-225) (>= 2 min) to a 15 min watch. Decision rule for a Table CR adopting a
    replica: treat managedByService != '' as read-only for metadata/maintenance; the only converging path to a
    writable table is DeleteTable + CreateTable (new ARN). Single run.
  - full notes: [details/S3T-TABLE-307.md](details/S3T-TABLE-307.md)

- <a id="s3t-table-329"></a>**S3T-TABLE-329** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Orphan replica after DeleteTableReplication keeps managedTableInformation.replicationInformation.sourceTableARN (0-479 s, same+cross region)**
  Completed table-level replicas (empty ICEBERG sources): baseline GetTable(replica) raw
  managedTableInformation = {"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/e5e8835c-55bd-4fe2-b696-2097ecb5f0e8"}}.
  DeleteTableReplication(tS -> same-region DS) -> HTTP 204, DeleteTableReplication(tX -> cross-region DX) ->
  HTTP 204, DeleteTable(source tD, replica in DS) -> HTTP 204; GetTableReplicationStatus(tS) right after ->
  NotFoundException (HTTP 404). Raw GetTable(replica) at [0.0, 10.0, 60.1, 479.3] s after the deletes - tS:
  managedByService ['replication.s3tables.amazonaws.com'], managedTableInformation present [True], raw
  ['{"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/e5e8835c-55bd-4fe2-b696-2097ecb5f0e8"}}'],
  nested replicationInformation [True], top-level replicationInformation key [False], sourceTableARN == source
  [True]; tX: ['replication.s3tables.amazonaws.com'] / [True] / ['{"replicationInformation":
  {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/a0072454-62b8-4f9a-8fc5-42addb7954bc"}}']
  / [True] / [False] / [True]; tD (source deleted): ['replication.s3tables.amazonaws.com'] / [True] /
  ['{"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-566f65-fu2-0305-src/table/b366eebe-2b77-43ed-bfdb-2e5a574c2ecd"}}']
  / [True] / [False] / [True].
  - ACK: is_read_only, pre-delete-cleanup, custom_find · ops: DeleteTableReplication, DeleteTable, GetTable,
    GetTableReplicationStatus · fields: managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, type
  - repro: table-level replication to a same-region and a cross-region bucket -> completed ->
    DeleteTableReplication (and DeleteTable(source) for a third table) -> raw GetTable(replica) at
    0/10/60/300+ s
  - measurements: sample_offsets_s=[0.0, 10.0, 60.1, 479.3], replicas=3, n=12
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-225](#s3t-table-225), [S3T-TABLE-307](#s3t-table-307), [S3T-TABLE-196](#s3t-table-196) · evidence:
    tablereplication/creative/fu2-03-05-orphan-body-dead-destination
  - notes: FU2-03: 225 holds and 307's 'replicationInformation present=False' is a key-path artifact, not a
    different behavior: in the GetTable body replicationInformation is NESTED under managedTableInformation
    (there is no top-level replicationInformation key, so a top-level check is always False). The...
  - full notes: [details/S3T-TABLE-329.md](details/S3T-TABLE-329.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-138](#s3t-table-138) | DeleteTable(replica) with source config active -> 204; status stays 'completed' until the next source commit turns it 'failed'; no re-create | replica_not_recreated_within_s=[300] | 1 |
| [S3T-TABLE-195](#s3t-table-195) | DeleteTable(source) with a replication config: at no-destinations -> 204, pending -> 204, completed -> 204, failed -> 204 | td_put_to_terminal_s=[6.3], te_put_to_failed_s=[0.06], tc_put_to_first_entry_s=[66.25] | 4 |
| [S3T-TABLE-225](#s3t-table-225) | After DeleteTableReplication the replica stays managedByService='replication.s3tables.amazonaws.com' for 2 min... | replica_managed_observed_s=[0.1] | 1 |
| [S3T-TABLE-299](#s3t-table-299) | Replica copies the source's snapshot-management config but forces compaction disabled (512/auto); no tags, no aws: tags, no policy | completed_at_s=65.68 | 1 |
| [S3T-TABLE-300](#s3t-table-300) | Replica settings are a replication-time snapshot: later source maintenance and policy changes did not reach the replica in 184 s | window_s=184.3, first_change_s=null | 36 |
| [S3T-TABLE-306](#s3t-table-306) | RenameTable (name, then namespace) on a replicated source: replica keeps the OLD name/namespace for 490 s; status completed; re-Put no-op | initial_completed_s=65.79, r1_window_s=122.3, r2_window_s=244.6, r3_window_s=122.3 | 1 |
| [S3T-TABLE-307](#s3t-table-307) | Orphaned replica after DeleteTableReplication: still managedByService for 874 s (30 samples); commit/maintenance 400... | watched_s=874, samples=30 | 1 |
| [S3T-TABLE-329](#s3t-table-329) | Orphan replica after DeleteTableReplication keeps managedTableInformation.replicationInformation.sourceTableARN (0-479 s, same+cross region) | sample_offsets_s=[0.0, 10.0, 60.1, 479.3], replicas=3 | 12 |
| [S3T-TABLEREPLICATION-002](#s3t-tablereplication-002) | Nonexistent role: PutTableReplication -> 200; status 'failed: Insufficient permissions given to successfully complete replication.' ~61 s | nonexistent_role_to_terminal_status_s=[60.61] | 1 |
| [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005) | GetTableReplicationStatus: destinations [] for ~60 s, then pending; empty table completed by +138 s, schema table at +436 s... | status_populated_after_put_s=[61.9, 63.6], empty_table_completed_by_s=[138.2], schema_table_pending_to_completed_s=[435.7], missing_destination_failed_at_s=[63.6] | 3 |
| [S3T-TABLEREPLICATION-006](#s3t-tablereplication-006) | A source commit (UpdateTableMetadataLocation) reaches the cross-region replica after 403-428 s; status drops back to 'pending' meanwhile | commit_to_replica_s=[403.0, 428.0], initial_pending_to_completed_s=[435.7] | 2 |
| [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008) | Role trusting only ec2: PutTableReplication -> 200, 'failed' at ~61 s; no self-heal in 150 s after fixing trust; re-Put -> completed | bad_trust_to_terminal_status_s=[60.66], self_heal_s=[], after_re_put_s=[65.88] | 1 |
| [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017) | Switch [D]->[E] (same region, empty table, n=1): Get reflects it at once; status showed E at +1.06s, dropped D and E 'completed' at +2.1s | status_drops_old_dest_s=[2.1], status_shows_new_dest_s=[1.06], new_dest_completed_s=[2.1] | 1 |
| [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019) | Empty table, table-level Put: status [] until +61.06s, pending until +62.09s, then 'completed'; replica listable in D at +62.2s (n=1) | no_destinations_window_s=[61.06], pending_to_terminal_s=[1.03], put_to_terminal_s=[62.09], replica_listable_s=[62.2] | 1 |
| [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020) | Re-Put after DeleteTableReplication with the service-managed replica left behind -> 200 but 'failed: identical name already exists' at +62s | re_put_to_failed_s=[61.99] | 1 |
| [S3T-TABLEREPLICATION-021](#s3t-tablereplication-021) | Table under bucket-level replication: GetTableReplication 404, Status 200; table-level Put (no token) -> 200, replaces bucket dest in status | bucket_level_t1_terminal_s=[2.16], table_level_e_terminal_s=[62.07] | 1 |
| [S3T-TABLEREPLICATION-022](#s3t-tablereplication-022) | Table-level Put [D,E] where D holds the bucket-level replica -> D 'failed' (identical name exists); table-level delete -> bucket view | new_table_terminal_s=[2.13] | 1 |
| [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025) | Collision with a user table in D: Put 200, 'D:failed' (identical name exists), user table untouched; same-region dest/existing ns: completed | t1_terminal_s=[62.0], t2_terminal_s=[72.3], t3_terminal_s=[68.1] | 3 |
| [S3T-TABLEREPLICATION-027](#s3t-tablereplication-027) | Re-Put (current token) on a 'failed' collision dest after the fix: 'completed' +63.3s; re-Put after DeleteTable(replica) re-creates it +75s | t1_fix_re_put_to_terminal_s=[63.3], t2_replica_recreated_s=[75.0], t1_first_failed_s=[72.5] | 3 |
| [S3T-TABLEREPLICATION-028](#s3t-tablereplication-028) | Two destinations, role allowed on B1 only (schema table): B2 'failed' at +66s, generic 'Insufficient permissions' message; B1 completed 475s | b1_completed_after_put_s=[475.15], b2_failed_after_put_s=[65.72], put_attempts=[1] | 1 |
| [S3T-TABLEREPLICATION-029](#s3t-tablereplication-029) | Fixing the role policy alone does NOT heal a 'failed' dest (180 s); identical re-Put (current token): B2 pending 66s, completed 521s; B1 ok | self_heal_watch_s=[180], re_put_to_b2_pending_s=[65.85], re_put_to_b2_completed_s=[520.83] | 1 |
| [S3T-TABLEREPLICATION-030](#s3t-tablereplication-030) | SSE-KMS source -> AES256 dest and AES256 source -> SSE-KMS dest: both Put 200, 'completed' ~62-63s; replica encryption follows dest bucket | tk_terminal_s=[62.2], tp_terminal_s=[63.2] | 2 |
| [S3T-TABLEREPLICATION-040](#s3t-tablereplication-040) | Destination bucket deleted under a completed config: 'failed: The specified destination bucket does not exist.' in 1 of 2 runs; re-Put 200 | lost_destination_window_s=182.7, status_change_at_s=null, heal_completed_at_s=0.03 | 1 |
| [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041) | Replicated source: RenameTable/namespace move leave replica at old ns/name 150 s, status 'completed'; tags/policy/maintenance absent at +5s | completed_at_s=66.1, rename_watch_last_at_s=0.1, move_watch_last_at_s=0.1 | 1 |
| [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042) | Re-pointing after a lost destination: Put [D2] -> D2 pending at +61 s, completed at +66 s, dead D entry gone; Put [dead D, D2] lists only D2 | d2_completed_s=65.73, heal_window_s=360 | 1 |
| [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043) | Corrective re-Put on a 'failed' collision dest: stale 'failed'+old message 61 s, then 'completed' directly; no self-heal without Put (300 s) | failed_to_first_at_s=61.15, stale_failed_after_reput_s=61.17, completed_after_reput_s=61.17 | 1 |
| [S3T-TABLEREPLICATION-045](#s3t-tablereplication-045) | GetTableReplicationStatus right after PutTableReplication -> 200 with destinations [] for ~60 s... | status_destinations_empty_for_s=60.7, still_pending_at_s=300 | 1 |
| [S3T-TABLEREPLICATION-047](#s3t-tablereplication-047) | Replacing the destination of a completed table-level config: new entry 'pending' +61-64 s after EVERY Put, completed ~1 s later (n=6) | first_put_entry_s=61.06, first_put_completed_s=62.1, offsets_after_last_change_s=[4.9, 19.97, 40.12, 55.08, 70.08, 0.87], new_entry_after_put_s=[61.21, 61.0, 61.91, 61.93, 61.12, 64.12], completed_after_put_s=[62.24, 62.04, 62.94, 62.96, 62.15, 65.15], new_entry_since_first_put_s=[128.21, 210.21, 313.28, 431.32, 563.56, 629.58] | 6 |
| [S3T-TABLEREPLICATION-048](#s3t-tablereplication-048) | Deleted destination bucket: 'failed ... does not exist.' within 5 s only after a re-Put storm (3/3); otherwise 'completed' for 180 s (3/3) | storm_first_failed_s=[4.4, 4.6, 4.5], no_storm_first_failed_s=[null, null, null], dead_poll_s=180, heal_poll_s=150, heal_new_completed_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0], heal_dead_failed_entry_appears_s=[63.2, 64.0, 71.1, 72.0, 66.7, 80.0] | 6 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-replication-status.md` - unverified, stale or single-run items worth a future probe (one line
each). Configuration-API items (token forms, validation, PassRole, the one-off 500) are in `table-replication.open-questions.md`.

- [S3T-TABLE-138](#s3t-table-138) (n=1): a table-level replica deleted under an active config was not re-created in 300 s, while the
  bucket-level reading [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032) (in `tablebucket-replication.md`) was corrected to "re-created in
  31-47 s" - re-run the table-level cell with a longer watch, with and without a source commit; recovery by a new Put after
  the 'failed: The destination table has been deleted during replication.' entry was not tested.
- [S3T-TABLEREPLICATION-005](#s3t-tablereplication-005) (notes): self-heal after creating the missing destination bucket later was not tested; -040/-048
  show no self-heal for a bucket deleted afterwards - the create-later direction is still open.
- [S3T-TABLEREPLICATION-040](#s3t-tablereplication-040) (confidence low, 1 of 2 runs): 'failed: The specified destination bucket does not exist.' at t=0
  without a storm; -048 (n=6) attributes the fast 'failed' to the 10x re-Put storm - keep -048 canonical.
- [S3T-TABLEREPLICATION-042](#s3t-tablereplication-042) vs -048: a Put naming a dead and a new bucket listed only the new one in a 30 s watch (-042) but
  the dead one 'failed' from +63-80 s in a 150 s watch (-048) - the ~60 s stale view explains it; -048 is canonical.
- [S3T-TABLEREPLICATION-017](#s3t-tablereplication-017) (confidence low): the +1.06 s appearance of a replacement destination did not reproduce in FU2-04;
  -047 (n=6, 61-64 s) is canonical - -017 should stay superseded in the rendered entry.
- [S3T-TABLE-307](#s3t-table-307) (confidence low): 'replicationInformation present=False' on the orphan replica is a key-path artifact
  superseded by [S3T-TABLE-329](#s3t-table-329) (nested under managedTableInformation, n=12); only its 874 s persistence watch remains useful.
- [S3T-TABLEREPLICATION-031](#s3t-tablereplication-031) / -006: the commit-driven 'completed' -> 'pending' -> 'completed' flip (403-428 s) rests on n=2
  commits on one cross-region schema table; not re-proved on same-region or empty tables.
- [S3T-TABLEREPLICATION-019](#s3t-tablereplication-019) vs -029: destinationTableArn absent on the first-backfill 'pending' (-019) but present on a re-Put
  'pending' (-029), n=1 each - the rendered entries should not call it a terminal marker; one more first-backfill watch at
  1 s would settle whether it ever appears before 'completed'.
- [S3T-TABLE-195](#s3t-table-195) (n=4): td reached terminal 6.3 s after its Put and te read 'failed' 0.06 s after its Put - far below the
  60.6-72.5 s window every other probe saw; the measurement anchors in that probe are unexplained.
- [S3T-TABLE-306](#s3t-table-306) / [S3T-TABLEREPLICATION-041](#s3t-tablereplication-041) (n=1 each): rename/namespace move never reached the replica in 122-490 s and no
  source commit was made; whether a later commit renames, re-creates or 'fails' the replica is untested.
- [S3T-TABLE-300](#s3t-table-300) (n=36 polls): the tag cell of the live-sync test is untested because the source TagResource returned 500;
  whether a successful later source tag write reaches the replica was not observed ([S3T-TABLE-299](#s3t-table-299) shows tags {} at creation).
- [S3T-TABLE-223](#s3t-table-223) vs -299: 223's note says the maintenance config was 'echoed'; 299 shows only snapshot management copied and
  compaction forced disabled 512/auto - 299 is canonical; 223's own data fits it.
- [S3T-TABLEREPLICATION-020](#s3t-tablereplication-020) (b): whether a re-Put after DeleteTable(replica) re-creates the replica was not observed there
  (stale 'failed'); -027 saw re-creation at +75.0 s (n=1) - a second run would settle it.
- [S3T-TABLEREPLICATION-025](#s3t-tablereplication-025) / -027: the colliding user table was 'untouched' at n=1 each; whether a 'failed' collision entry
  ever changes if the user table is later renamed (rather than deleted) was not tried.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-137"></a>**S3T-TABLE-137** | immutable-field | high | confirmed | Replica table writes: UpdateTableMetadataLocation -> BadRequestException; PutTablePolicy -> 200 accepted... | - | [S3T-TABLE-224](#s3t-table-224) |
| <a id="s3t-table-268"></a>**S3T-TABLE-268** | other | low | confirmed | Doc claim C076 TRUE: GetTable(replica) has managedByService and replicationInformation.sourceTableARN; the source table has neither | [S3T-TABLE-136](#s3t-table-136), [S3T-TABLE-223](#s3t-table-223), [S3T-TABLE-177](table-replication.md#s3t-table-177) | - |
| <a id="s3t-tablereplication-026"></a>**S3T-TABLEREPLICATION-026** | async-state-machine | high | confirmed | Failed destination (name collision): after removing the cause the status stayed 'failed' for 120 s with no Put... | [S3T-TABLEREPLICATION-008](#s3t-tablereplication-008) | [S3T-TABLEREPLICATION-043](#s3t-tablereplication-043) |
| <a id="s3t-tablereplication-035"></a>**S3T-TABLEREPLICATION-035** | other | low | confirmed | Doc claim C074 TRUE: DeleteTableReplication(current token) -> 204; GetTableReplication -> NotFoundException; the replica table remains | [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-024](tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](tablebucket-replication.md#s3t-tablebucketreplication-025), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
