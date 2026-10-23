<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-015: Error catalogue BadRequestException (400): 57 message classes - 45 PERMANENT-SPEC, 5 WAIT-FOR-STATE, 4 AMBIGUOUS, 1 AUTHZ, 1 QUOTA, 1 RETRY
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-015"></a>**S3T-SERVICE-015** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Error catalogue BadRequestException (400): 57 message classes - 45 PERMANENT-SPEC, 5 WAIT-FOR-STATE, 4 AMBIGUOUS, 1 AUTHZ, 1 QUOTA, 1 RETRY**
  BadRequestException: declared on 49/49 operations; 988 occurrences, 60 message classes in the seed + 3 new.
  57 message classes; classification counts {'AMBIGUOUS': 4, 'AUTHZ': 1, 'PERMANENT-QUOTA': 1,
  'PERMANENT-SPEC': 45, 'RETRY': 1, 'WAIT-FOR-STATE': 5}. Mostly PERMANENT-SPEC, but 5 classes are
  state-dependent (not-empty deletes, missing metadata/sort order, disabled KMS key), 1 is RETRY (continuation
  token), 1 is a KMS key-policy AUTHZ problem and 4 are AMBIGUOUS (fresh-principal propagation, catch-all
  'Invalid input', identifier message for an empty configuration, invisible foreign key).
  - ACK: terminal_codes, docs-only · ops: CreateNamespace, CreateTable, CreateTableBucket, DeleteNamespace,
    DeleteTableBucket, DeleteTableBucketReplication, DeleteTableReplication, GetNamespace, GetTable,
    GetTableBucket, GetTableReplication, ListNamespaces, ListTableBuckets, ListTables, ListTagsForResource,
    PutTableBucketEncryption, PutTableBucketMetricsConfiguration, PutTableBucketPolicy,
    PutTableBucketReplication, PutTableMaintenanceConfiguration, PutTablePolicy,
    PutTableRecordExpirationConfiguration, PutTableReplication, RenameTable, TagResource, UntagResource,
    UpdateTableMetadataLocation
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-SERVICE-007](../service.md#s3t-service-007), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-021](../service.md#s3t-service-021),
    [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-016](../service.md#s3t-service-016), [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-022](../service.md#s3t-service-022) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue

## Notes

Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the request is
right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion cooldown, or
the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC = the request is wrong or the
desired state already holds (non-idempotent deletes), terminal until the spec changes; PERMANENT-QUOTA = a
limit; AUTHZ = permissions (IAM/resource/boundary/session policy, KMS key policy, cross-account); AMBIGUOUS =
the same code+message has different causes with different remediations (named in the rule). Sources:
seed-error-messages.md (105 structured probes, 391 rows; finding ids cited per row) plus the creative-round
probes service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
service/creative/iam-authz-variants, service/creative/throttle-scope ([S3T-SERVICE-001](../service.md#s3t-service-001)..013). Messages are
quoted as recorded; <ARN>, <N>, <T> are normalised. Controllers should branch on the wire code and HTTP status
and treat the message as diagnostic text only - several texts echo caller input or differ by a curly
apostrophe.

\## BadRequestException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| The provided Region does not match the Region of the resource. (n=159) | every op taking an ARN (33 ops
seen) | 400 | PERMANENT-SPEC | The ARN's region differs from the endpoint; checked BEFORE authz and lookup
(ghost names in an other-region ARN also 400, [S3T-TABLE-245](../table-metadata-location.md#s3t-table-245); under an IAM deny still 400 not 403,
[S3T-SERVICE-011](../service.md#s3t-service-011)). Controller: use a client for the ARN's region; never retry unchanged. |
| The provided Partition does not match the Partition of the resource. (n=4) | DeleteTableBucket,
GetTableBucket, ListTagsForResource, PutTableBucketMetricsConfiguration | 400 | PERMANENT-SPEC |
aws-cn/aws-us-gov ARN on a commercial endpoint; terminal. |
| The specified bucket name isn't valid. Specify a different bucket name, and then try again. (n=125) |
CreateTableBucket (name rules, missing name, name=null, empty body), GetTableBucket and 18 more ops when the
ARN's bucket segment violates the name rules | 400 | PERMANENT-SPEC | Name rule violations incl. all reserved
affixes ([S3T-TABLEBUCKET-022](../tablebucket.md#s3t-tablebucket-022)) and a missing/null name ([S3T-SERVICE-002](../service.md#s3t-service-002)). On non-Create ops it means the ARN is
malformed. Note the curly apostrophe in some variants of this text (namespace message uses U+2019) - do not
match on the message. |
| The specified table bucket ARN is not valid. (n=82) | GetTableBucket, GetNamespace, CreateNamespace,
CreateTable, DeleteTableBucket, ListNamespaces + 10 more; PutTableBucketReplication with no tableBucketARN |
400 | PERMANENT-SPEC | Not an ARN / bare name / tableBucketId / a TABLE ARN in the bucket slot
([S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003)/029, [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLE-180](../table-maintenance.md#s3t-table-180)). Also returned for a missing tableBucketARN query
parameter on PutTableBucketReplication ([S3T-SERVICE-007](../service.md#s3t-service-007) variants). |
| The specified resourceArn is not valid. (n=43) | ListTagsForResource, TagResource, UntagResource | 400 |
PERMANENT-SPEC | Tag ops accept only table-bucket and table ARNs; a synthesized namespace ARN is rejected
(namespaces are not taggable, [S3T-NAMESPACE-007](../namespace.md#s3t-namespace-007)). |
| Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name)
together (n=39) | GetTable, PutTableReplication, PutTableBucketReplication | 400 | AMBIGUOUS | (a) GetTable
with no identifier, a partial triple, or tableArn AND the triple even for the same table ([S3T-TABLE-024](../table.md#s3t-table-024)/025)
-> spec; (b) MISLEADING: PutTableReplication/PutTableBucketReplication with a valid identifier but
configuration={} or rules=[] return this identifier message ([S3T-SERVICE-007](../service.md#s3t-service-007) variants) -> the real problem is
the empty configuration. |
| The specified table ARN is not valid. (n=36) | GetTable, GetTableRecordExpiration*,
GetTableReplication(+Status), PutTableRecordExpirationConfiguration, PutTableReplication | 400 |
PERMANENT-SPEC | tableArn-keyed ops given a bucket ARN, a name-based pseudo-ARN or 'ns/table' ([S3T-TABLE-173](../table-replication.md#s3t-table-173)).
Checked before authz ([S3T-TABLE-190](../table-record-expiration.md#s3t-table-190)). |
| The specified metadata is not valid (n=35) | CreateTable | 400 | PERMANENT-SPEC | iceberg={}, metadata={},
properties without schema, schemaV2 without fields, duplicate field names/ids, bad nested type strings,
unknown partition sourceId ([S3T-TABLE-036](../table-schema.md#s3t-table-036)/037/039/040/042/072). schema={} is NOT this 400 but a 500 (see
InternalServerErrorException). |
| Both 'schema' and 'schemaV2' were specified. Use either 'schema' or 'schemaV2' and retry the create table
command. (n=1) | CreateTable | 400 | PERMANENT-SPEC | [S3T-TABLE-041](../table-schema.md#s3t-table-041); mutually exclusive fields. |
| The specified open table format is not valid (n=15) | CreateTable (format other than exactly 'ICEBERG', or
format missing / empty body) | 400 | PERMANENT-SPEC | Case-sensitive enum ([S3T-TABLE-047](../table.md#s3t-table-047)/079); the
missing-member case comes from the handler, not the framework ([S3T-SERVICE-002](../service.md#s3t-service-002)). |
| The specified table name is not valid. (n=22) | CreateTable, RenameTable | 400 | PERMANENT-SPEC | Name rules
[0-9a-z_], leading '_' rejected, 256+ chars ([S3T-TABLE-078](../table.md#s3t-table-078), [S3T-TABLE-052](../table.md#s3t-table-052)). |
| The specified namespace name isn't valid. Specify a different namespace name, and then try again. (n=22) |
CreateNamespace | 400 | PERMANENT-SPEC | Rules in [S3T-NAMESPACE-012](../namespace.md#s3t-namespace-012)..016; also the empty-body case. Text uses
U+2019. |
| Multipart namespaces are not supported. (n=3) | CreateNamespace(namespace=[a, b]) | 400 | PERMANENT-SPEC |
namespace is list<string> but max 1 element ([S3T-NAMESPACE-004](../namespace.md#s3t-namespace-004)/011) - codegen hazard. |
| Neither a new namespace name nor a new table name is specified. (n=1) | RenameTable | 400 | PERMANENT-SPEC |
RenameTable with neither field / empty body ([S3T-TABLE-050](../table.md#s3t-table-050), [S3T-SERVICE-002](../service.md#s3t-service-002)). A no-op rename is a 409, see
ConflictException. |
| The specified storage class is not valid. (n=11) | CreateTable, CreateTableBucket | 400 | PERMANENT-SPEC |
Enum typo at create ([S3T-TABLE-012](../table-encryption-storage.md#s3t-table-012)); the same typo on PutTableBucketStorageClass is a ValidationException
instead. |
| The specified SSE algorithm is invalid. (n=29) | CreateTable, PutTableBucketEncryption, CreateTableBucket |
400 | PERMANENT-SPEC | Case-sensitive 'AES256'/'aws:kms' ([S3T-TABLE-087](../table-encryption-storage.md#s3t-table-087), [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024)); the enum is
enforced by the handler, not the framework. |
| An encryption configuration is not specified (n=live) | PutTableBucketEncryption with body {} | 400 |
PERMANENT-SPEC | [S3T-SERVICE-002](../service.md#s3t-service-002) (new class, handler-side required check). |
| KMS key ARN detected. Please specify 'aws:kms' as the encryption algorithm. (n=7) | CreateTable,
PutTableBucketEncryption, CreateTableBucket | 400 | PERMANENT-SPEC | AES256 + kmsKeyArn
([S3T-TABLEBUCKET-034](../tablebucket-encryption-storage.md#s3t-tablebucket-034)/102). |
| The specified KMS key ARN is invalid. Only key ARN format
'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported. (n=28) | CreateTable,
PutTableBucketEncryption, CreateTableBucket | 400 | PERMANENT-SPEC | Bare key id, alias, empty string, aws:kms
without kmsKeyArn ([S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084)/085, [S3T-TABLEBUCKET-103](../tablebucket-encryption-storage.md#s3t-tablebucket-103)). |
| Cross region KMS keys are not supported. (n=8) | CreateTable, CreateTableBucket, PutTableBucketEncryption |
400 | PERMANENT-SPEC | Key ARN of another region ([S3T-TABLE-247](../table-policy.md#s3t-table-247), [S3T-TABLEBUCKET-105](../tablebucket-encryption-storage.md#s3t-tablebucket-105)/128). |
| The specified KMS key ARN does not exist. (n=10) | CreateTable, CreateTableBucket, PutTableBucketEncryption
| 400 | AMBIGUOUS | (a) typo / deleted key -> terminal; (b) a key the caller cannot see (another account's key
without a grant) is reported with the same text ([S3T-TABLE-086](../table-policy.md#s3t-table-086), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035)) -> looks like absence but
is cross-account access; (c) a key created in the same reconcile: KMS keys are usable immediately, no window
measured here. |
| The specified KMS key ARN is disabled. (n=7) | CreateTable, PutTableBucketEncryption | 400 | WAIT-FOR-STATE
| Same text for a Disabled AND a PendingDeletion key ([S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107)) -> external KMS
state; requeue (the user may re-enable / cancel deletion); existing tables keep working for control-plane ops
([S3T-TABLE-240](../table-record-expiration.md#s3t-table-240)/241). |
| Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
exists, is active, and has a resource policy granting access to the S3 service principal
'maintenance.s3tables.amazonaws.com'. (n=15) | CreateTable (explicit aws:kms or inheriting a bucket default
whose key lacks the grant) | 400 | AUTHZ | An authorization problem (KMS key policy) behind a 400: the key
must grant maintenance.s3tables.amazonaws.com with the documented conditions ([S3T-TABLE-054](../table-maintenance.md#s3t-table-054)/056/239/246).
PutTableBucketEncryption with the same key is accepted (200) and the failure only appears at CreateTable
([S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119)) -> a Table reconciler must surface it as a prerequisite condition; after
the key policy is fixed, retry (KMS policy propagation is seconds). |
| Public policies cannot be applied to S3 table buckets and tables. (n=36) | PutTablePolicy,
PutTableBucketPolicy | 400 | PERMANENT-SPEC | Allow with Principal '*' / {AWS:'*'} / a list containing '*' and
no Condition ([S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089)); Deny '*' is accepted. |
| Invalid principal in policy (n=13) | PutTablePolicy, PutTableBucketPolicy | 400 | AMBIGUOUS | (a)
nonexistent / deleted role or user ARN, Federated principals ([S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001)) ->
terminal; (b) TRANSIENT for a principal created seconds earlier: 7.2 / 7.9 / 7.8 s after iam:CreateRole the
Put flips from this 400 to 200 (n=3, [S3T-SERVICE-010](../service.md#s3t-service-010)) -> a controller that creates the Role and the policy in
one pass must requeue on this message for ~10 s before giving up. Once stored, deleting the role rewrites the
principal to its AROA id ([S3T-TABLE-165](../table-policy.md#s3t-table-165)). |
| Invalid service principal in policy (n=1) | PutTablePolicy | 400 | PERMANENT-SPEC | Non-amazonaws Service
principal ([S3T-TABLE-164](../table-policy.md#s3t-table-164)). |
| Policy contains a statement with no principals (n=2) | PutTableBucketPolicy, PutTablePolicy | 400 |
PERMANENT-SPEC | Resource policies need Principal ([S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088)). |
| The specified policy is malformed (n=11) | PutTablePolicy, PutTableBucketPolicy | 400 | PERMANENT-SPEC | Bad
JSON, unknown top-level keys, Statement [], Version 2030, whitespace-only document ([S3T-TABLE-162](../table-policy.md#s3t-table-162),
[S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088)/090). An EMPTY string is a ValidationException instead. |
| Invalid Action (n=4) | PutTableBucketPolicy, PutTablePolicy | 400 | PERMANENT-SPEC | Foreign action
namespace (s3:GetObject); unknown s3tables:* action names are accepted ([S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLEBUCKET-110](../tablebucket-policy.md#s3t-tablebucket-110)). |
| The specified compaction file size is invalid. Minimum value: 64 MB, maximum value: <N> MB. (n=30) |
PutTableMaintenanceConfiguration | 400 | PERMANENT-SPEC | targetFileSizeMB bounds 64..512 ([S3T-TABLE-226](../table-maintenance.md#s3t-table-226)). |
| The specified minSnapshotsToKeep is invalid. Value must be a positive integer. (n=9) |
PutTableMaintenanceConfiguration | 400 | PERMANENT-SPEC | 0 / -1 ([S3T-TABLE-227](../table-maintenance.md#s3t-table-227)). |
| The specified maxSnapshotAgeHours is invalid. Value must be a positive integer. (n=3) |
PutTableMaintenanceConfiguration | 400 | PERMANENT-SPEC | 0 / -1 ([S3T-TABLE-227](../table-maintenance.md#s3t-table-227)). |
| The compaction strategy specified is not valid. (n=9) | PutTableMaintenanceConfiguration | 400 |
PERMANENT-SPEC | Exact lowercase 'auto'/'binpack'/'sort'/'z-order' ([S3T-TABLE-230](../table-maintenance.md#s3t-table-230)). |
| The specified status is invalid. Value must be either 'enabled' or 'disabled.' (n=7) |
PutTableMaintenanceConfiguration | 400 | PERMANENT-SPEC | Table-level status enum is handler-validated
(BadRequest) while the BUCKET-level one is a ValidationException ([S3T-TABLE-230](../table-maintenance.md#s3t-table-230) vs [S3T-TABLEBUCKET-074](../tablebucket-maintenance-metrics.md#s3t-tablebucket-074)). |
| Invalid table maintenance type: <value> (n=22, six variants incl. icebergUnreferencedFileRemoval,
IcebergCompaction, ICEBERGCOMPACTION, iceberg-compaction, foo, icebergCompaction%20) |
PutTableMaintenanceConfiguration | 400 | PERMANENT-SPEC | The path segment must be exactly icebergCompaction
or icebergSnapshotManagement; icebergUnreferencedFileRemoval is bucket-level only ([S3T-TABLE-160](../table-maintenance.md#s3t-table-160)/229). The
message echoes the caller's value - do not match on it. |
| The specified maintenance type is not valid for service managed tables. (n=1) |
PutTableMaintenanceConfiguration on a replica | 400 | PERMANENT-SPEC | Replica (service-managed) tables are
read-only for maintenance ([S3T-TABLE-224](../table-replication-status.md#s3t-table-224)). |
| The table metadata location is not specified. (n=6) | PutTableMaintenanceConfiguration(strategy
sort/z-order) | 400 | WAIT-FOR-STATE | Needs a first metadata commit ([S3T-TABLE-204](../table-maintenance.md#s3t-table-204)); binpack/auto/size work
on an empty table. Requeue until the table has metadataLocation. |
| The specified compaction strategy requires sort order in the table metadata. Please refer to the S3 Tables
documentation for steps on configuring sort order. (n=2) | PutTableMaintenanceConfiguration(strategy
sort/z-order) | 400 | WAIT-FOR-STATE | Depends on the Iceberg metadata written by the data plane
([S3T-TABLE-232](../table-maintenance.md#s3t-table-232)); a controller cannot fix it - surface as a prerequisite condition and requeue slowly. |
| The table maintenance configuration value is not specified. (n=live) | PutTableMaintenanceConfiguration with
body {} | 400 | PERMANENT-SPEC | [S3T-SERVICE-002](../service.md#s3t-service-002) (new class): required 'value' is handler-checked here,
framework-checked (ValidationException) at bucket level. |
| Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints on request
parameters. (n=10 + live) | PutTableMaintenanceConfiguration, PutTableRecordExpirationConfiguration; ANY op
given malformed JSON, a wrong JSON type (string for list, int for string, list for map), 'null'/'[]' bodies or
a non-integer query parameter (maxBuckets=abc) | 400 | AMBIGUOUS | Catch-all of the deserializer
([S3T-SERVICE-002](../service.md#s3t-service-002)) and of some handler checks ([S3T-TABLE-226](../table-maintenance.md#s3t-table-226): targetFileSizeMB as string/bool); the message
never names the field. Always a request-shape problem (terminal) - but a generated controller cannot tell
which field from the text. |
| The namespace that you tried to delete is not empty. (n=12) | DeleteNamespace | 400 | WAIT-FOR-STATE |
Children (tables) must be deleted first ([S3T-NAMESPACE-024](../namespace.md#s3t-namespace-024)); DeleteNamespace at +0 ms after the last
DeleteTable succeeds ([S3T-NAMESPACE-027](../namespace.md#s3t-namespace-027)) -> finalizer ordering, requeue. |
| The bucket that you tried to delete is not empty. (n=11) | DeleteTableBucket | 400 | WAIT-FOR-STATE | Any
namespace (even empty) or table, including service-created REPLICA tables, blocks the delete
([S3T-TABLEBUCKET-019](../tablebucket.md#s3t-tablebucket-019)/044/118); policy/encryption/tags/maintenance do not ([S3T-TABLEBUCKET-020](../tablebucket-encryption-storage.md#s3t-tablebucket-020)). Finalizer:
delete children (or wait for their reconcilers) then retry. |
| Replication operations are not allowed on service managed tables. (n=17) |
Get/Put/DeleteTableReplication(+Status) on a replica | 400 | PERMANENT-SPEC | Scope: replica tables
(managedByService=replication...) cannot be sources ([S3T-TABLE-137](../table-replication-status.md#s3t-table-137)/224). |
| Updates to service managed tables are not allowed. (n=6) | UpdateTableMetadataLocation on a replica | 400 |
PERMANENT-SPEC | Replica tables are read-only ([S3T-TABLE-137](../table-replication-status.md#s3t-table-137)); stays so for >=2 min after
DeleteTableReplication ([S3T-TABLE-225](../table-replication-status.md#s3t-table-225)). |
| The specified metadata location is not valid. (n=15) | UpdateTableMetadataLocation | 400 | PERMANENT-SPEC |
Prefix must be the table's warehouseLocation (plain string prefix, [S3T-TABLE-105](../table-metadata-location.md#s3t-table-105)), scheme 's3://' lowercase
([S3T-TABLE-106](../table-metadata-location.md#s3t-table-106)), must contain '.metadata.json' ([S3T-TABLE-135](../table-metadata-location.md#s3t-table-135)). A NON-EXISTENT object under the warehouse is
accepted (200) - no 400 protects against dangling pointers ([S3T-TABLE-100](../table-metadata-location.md#s3t-table-100)/192). |
| A version token is not specified. (n=13) | DeleteTableBucketReplication, PutTableBucketReplication,
PutTableReplication, DeleteTableReplication | 400 | PERMANENT-SPEC | versionToken is mandatory for the
replication Deletes and for a Put when a config exists ([S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006)). Read
Get*Replication.versionToken first; it ROTATES on every Put, even an unchanged one
([S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002)). |
| The continuation token is no longer valid. (n=11) | ListNamespaces, ListTables, ListTableBuckets | 400 |
RETRY | Garbage, another bucket's or an expired token ([S3T-NAMESPACE-009](../namespace.md#s3t-namespace-009), [S3T-TABLEBUCKET-041](../tablebucket.md#s3t-tablebucket-041)) -> restart the
listing from the first page (tokens survive concurrent create/delete, [S3T-NAMESPACE-010](../namespace.md#s3t-namespace-010), [S3T-TABLE-095](../table.md#s3t-table-095)). |
| The specified limit is not valid. A limit should be a positive number (n=7) | ListTableBuckets,
ListNamespaces, ListTables | 400 | PERMANENT-SPEC | max*=0 or negative; values above the model max (1001) are
ACCEPTED ([S3T-NAMESPACE-032](../namespace.md#s3t-namespace-032), [S3T-TABLE-168](../table.md#s3t-table-168), [S3T-TABLEBUCKET-069](../tablebucket.md#s3t-tablebucket-069)). |
| The specified search prefix is not valid. A search prefix can not exceed <N> characters. (n=3) |
ListNamespaces, ListTables | 400 | PERMANENT-SPEC | Over-long prefix; invalid characters in a prefix return
200 with an empty list instead ([S3T-TABLEBUCKET-041](../tablebucket.md#s3t-tablebucket-041)). |
| The specified bucket type filter value isn't valid. Specify a valid value, and then try again. (n=4) |
ListTableBuckets(type=...) | 400 | PERMANENT-SPEC | Enum 'customer'/'aws' ([S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130)). |
| The provided tags are invalid. Number of user tags must not exceed 50 (n=9) | TagResource, CreateTable,
CreateTableBucket | 400 | PERMANENT-QUOTA | 50 user tags per bucket/table, cumulative and enforced atomically -
none of the batch is applied ([S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLEBUCKET-031](../tablebucket.md#s3t-tablebucket-031)); at Create nothing is created
([S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025)). Terminal until tags are removed. |
| Tags beginning with 'aws:' are reserved for system use. (n=7) | TagResource, UntagResource,
CreateTableBucket | 400 | PERMANENT-SPEC | Also on UNTAG of an aws: key ([S3T-TABLE-118](../table.md#s3t-table-118), [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025)).
|
| The provided tags are invalid. Tag key can only contain alphanumeric characters, spaces, and _.:/=+@- (n=3)
| TagResource | 400 | PERMANENT-SPEC | [S3T-TABLE-119](../table.md#s3t-table-119). |
| The provided tags are invalid. Tag value can only contain alphanumeric characters, spaces, and _.:/=+@-
(n=1) | TagResource | 400 | PERMANENT-SPEC | [S3T-TABLE-119](../table.md#s3t-table-119). |
| The provided tags are invalid. Tag key is empty or exceeds max length (n=1) | CreateTableBucket | 400 |
PERMANENT-SPEC | At Create the length rules are handler-checked (BadRequest); on TagResource the SAME
violation is a ValidationException ([S3T-TABLEBUCKET-030](../tablebucket.md#s3t-tablebucket-030)). |
| The provided tags are invalid. Tag value exceeds max length (n=1) | CreateTableBucket | 400 | PERMANENT-SPEC
| As above. |
| The provided tags are invalid. (n=2 + live) | UntagResource(tagKeys=[]), TagResource with body {} (tags
missing) or tags=[] | 400 | PERMANENT-SPEC | Empty key list / missing map ([S3T-TABLE-125](../table.md#s3t-table-125), [S3T-TABLEBUCKET-028](../tablebucket.md#s3t-tablebucket-028),
[S3T-SERVICE-002](../service.md#s3t-service-002)); TagResource with tags={} is 200. |

Classification counts: {'AMBIGUOUS': 4, 'AUTHZ': 1, 'PERMANENT-QUOTA': 1, 'PERMANENT-SPEC': 45, 'RETRY': 1,
'WAIT-FOR-STATE': 5}.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. BadRequestException is not
uniformly terminal: 'The specified KMS key ARN is disabled.' (Disabled or PendingDeletion key, re-enable heals
it), 'The table metadata location is not specified.' and the sort-order message (heal after a data-plane
commit), 'Insufficient access to perform table maintenance' (KMS key policy) and 'The continuation token is no
longer valid.' (restart the listing) all heal without a spec change, so a blanket terminal_codes:
[BadRequestException] strands CreateTable/PutTableBucketEncryption/maintenance updates in ACK.Terminal;
classify by message or narrow the terminal list for those paths.
