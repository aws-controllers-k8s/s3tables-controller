<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table metadata and schema (CreateTable metadata.iceberg)
_CreateTable metadata.iceberg: write-only through the S3 Tables API (no read path), schema={} -> 500, field ids renumbered and types normalized, server-added properties, required round-trip, format rules, what metadata.json carries._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-schema.md`; every claim cites the finding ids that carry the full evidence
below.

### Orientation

This document covers the create-only `metadata.iceberg` block of a Table - `schema` (v1) or `schemaV2`, `partitionSpec`,
`writeOrder`, `properties` - what CreateTable accepts and rejects, and what it commits to the metadata.json at
`metadataLocation`. The mental model shared with `table-metadata-location.md`: `metadata.iceberg` is write-only through
the S3 Tables API, `metadataLocation` is engine-managed state that CreateTable-with-schema sets synchronously and every
later commit moves, and `versionToken` is a deterministic function of `metadataLocation` (the token and location rules
live in the sibling document). Here the rule is that what is committed is not what was sent: GetTable,
GetTableMetadataLocation and ListTables echo none of the schema, so the only read-back is S3 GetObject on
`metadataLocation` ([S3T-TABLE-018](#s3t-table-018), -313, -322); field ids are renumbered 1..n, an omitted `required` becomes false, type
strings are normalized, spec/order/schema ids are forced and a property is injected ([S3T-TABLE-043](#s3t-table-043), -040, -314, -016,
-312, -038, -289, -045, -322, -017). Every verdict is returned on the CreateTable call itself: 200 with a committed table,
BadRequestException 400 'The specified metadata is not valid' for the rejected shapes, and InternalServerErrorException
500 for the one shape (`schema={}`) that never reaches the 400 ([S3T-TABLE-036](#s3t-table-036), -133, -041). A table created with a schema
shows `metadataLocation` = `<warehouseLocation>/metadata/00000-<uuid>.metadata.json` on the first GetTable, with the
CreateTable token ([S3T-TABLE-015](#s3t-table-015), -033, -283).

### Rules

1. Guard the empty-schema shapes before the wire: `schema={}` (fields omitted) -> InternalServerErrorException 500 'An
   internal error occurred. Try again.' in 4/4 with no table created - a retry loop, never the terminal 400 - while
   `schema.fields=[]` -> 200 and a committed empty-schema table (fields=[], last-column-id 0) ([S3T-TABLE-133](#s3t-table-133), -036).
2. The other empty shapes are the generic 400 'The specified metadata is not valid': `iceberg={}`, `metadata={}` (a
   client-side ParamValidationError while SDK validation is on), properties-only without a schema, `schemaV2={type:struct}`
   without fields; `schema` plus `schemaV2` -> 400 'Both 'schema' and 'schemaV2' were specified...' while schemaV2 alone
   is accepted ([S3T-TABLE-036](#s3t-table-036), -311, -133, -041).
3. Never diff `spec.metadata.iceberg` against the API: the raw bodies of GetTable, GetTableMetadataLocation and ListTables
   carry 0 of 19 probed tokens (field names, partition name, property marker, transform/direction/null-order words);
   GetTable exposes only `type='customer'` and `format='ICEBERG'`; the schema lives only in metadata.json
   ([S3T-TABLE-018](#s3t-table-018), -322, -313, -312).
4. Expect the committed metadata.json to differ from the request: field ids renumbered 1..n (v1 ids 5,9 -> 1,2; schemaV2
   1,5,42 -> 1,2,3; 0 and -1 accepted; only duplicates 5,5 / 1,1 -> 400), omitted `required` -> false, 'INT'/'String'
   lower-cased and 'decimal(10,2)' -> 'decimal(10, 2)', `specId`/`orderId`/`schemaId` ignored (spec-id 0, order-id 1,
   schema-id 0), partition `fieldId` forced to 1000.., and `write.parquet.compression-codec=zstd` added to properties
   ([S3T-TABLE-040](#s3t-table-040), -043, -314, -016, -312, -288, -038, -289, -074, -045, -286, -287, -322, -017).
5. v1 field types: the 14 Iceberg primitives plus the case variants 'INT'/'String' are accepted; 'integer', 'varchar',
   'bigint', 'decimal', 'fixed', '' and ' int ' -> 400; nested strings 'struct<a:int>', 'list<string>', 'map<string,int>'
   -> 400 - nested types need schemaV2 with a JSON type document ([S3T-TABLE-038](#s3t-table-038), -039, -046).
6. schemaV2 is stricter than v1: `type` must be exactly 'struct' (omitted, 'STRUCT', 'record' -> 400); every field must
   carry `required` (omitted -> 400, no default); `identifierFieldIds` must name required fields ([2] on an optional
   field or [9] unknown -> 400, [1] accepted and written); 'varchar' -> 400 ([S3T-TABLE-042](#s3t-table-042), -046, -044).
7. Field names are case-sensitive and must be unique: fields 'a','a' -> 400; 'a','A' -> accepted ([S3T-TABLE-037](#s3t-table-037)).
8. partitionSpec: `sourceId` must be a schema field id (99, 0 or past-last 3 -> 400, no dangling partition field) but the
   transform is not grammar-checked - 'bucket' without arity, 'days' and '' are written verbatim, 'DAY' lower-cased,
   only day(long) -> 400; `fieldId` (0, 5, 2) and `specId` (3) are ignored; `{}` and `fields=[]` pass as an
   unpartitioned spec (spec-id 0) ([S3T-TABLE-072](#s3t-table-072), -290, -073, -074, -284, -133, -286).
9. writeOrder: `direction` and `nullOrder` have no defaults (omitted -> 400; 'ASC' lower-cased; 'NULLS_FIRST' -> 400);
   `orderId` is ignored (omitted/2/7/-1 -> order-id 1) except orderId=0 with fields -> 400; an unknown `sourceId` -> 400;
   the empty shapes `{fields:[]}`, `{orderId:0|1, fields:[]}`, `{}` and `{orderId:1}` -> 200 as the unsorted order 0
   ([S3T-TABLE-076](#s3t-table-076), -075, -289, -077, -285, -133, -287).
10. properties: any map passes - empty key, empty value, format-version/owner keys, 200 entries, a 5 KiB value, {} -> 200;
    keys are written verbatim (case and whitespace kept) and `write.parquet.compression-codec=zstd` is added
    ([S3T-TABLE-194](#s3t-table-194), -017, -322).
11. CreateTable-with-schema is synchronous: `metadataLocation` = `<warehouseLocation>/metadata/00000-<uuid>.metadata.json`
    is present on the first GetTable, GetTableMetadataLocation returns the same location and the CreateTable token, and
    modifiedAt == createdAt; without a schema the key is absent (not an empty string) ([S3T-TABLE-015](#s3t-table-015), -033, -283).
12. Defaults in the committed file for a v1 schema: format-version 2, current-schema-id 0, ids [1, 2, 3], required
    [false, true, false], partition-specs [{spec-id 0, fields []}], default-spec-id 0, default-sort-order-id 0, 0 snapshots
    ([S3T-TABLE-016](#s3t-table-016), -312).

### Timing ranges

- CreateTable with a 3-field schema + properties: 200 in 375 ms; `metadataLocation` present on the first GetTable poll at
  0.42 s, 0 polls before present, location unchanged over the following 2 s (n=1, [S3T-TABLE-015](#s3t-table-015)); the immediate GetTable
  of -033 and the first read of -283 (0.42 s) also had it (n=1 each) - single-run, no range yet.
- Without a schema the `metadataLocation` key is absent at 0 s and at 5 s (n=1, [S3T-TABLE-283](#s3t-table-283)).
- No other duration is measured in this document: every validation verdict (-036, -133, -284, -285 and the matrices behind
  rules 5-9) is the CreateTable response itself, and the round-trip findings (-312, -313, -314, -322) read metadata.json
  once after the create. UpdateTableMetadataLocation, DeleteTable and RenameTable ranges are in
  `table-metadata-location.md`.

### Handling gaps

Today `Metadata` is `is_immutable` and sent only at CreateTable (generator.yaml:43-46), yet delta.go compares desired
against latest `Spec.Metadata.Iceberg` field by field (delta.go:70-131); `SchemaField.Required` is generated as
`elem.Required = *iter.Required` (sdk.go:363-364); `go_tag json:"type,omitempty"` exposes `type` on Table and SchemaField
(generator.yaml:45-46, 60-65); BadRequestException is the only terminal code (generator.yaml:99-104). One finding is
suspect-bug and three are partial (two HANDLING-GAP-marked plus the low-impact -314); none is fully handled.

- [S3T-TABLE-133](#s3t-table-133) (suspect-bug, not yet handled): no CRD guard on `schema.fields`. Evidence: `schema={}` on the wire -> 500
  in 4/4 with no table created (n=3 fresh + 1), so the reconciler requeues forever; `schema.fields=[]` -> 200 and an
  empty-schema table reported Synced (-036). The finding's fix: minItems=1 on `schema.fields` or exactly one of
  `schema`/`schemaV2`, or a pre-build guard; today the SDK's client-side required-member check is the only guard.
- [S3T-TABLE-312](#s3t-table-312) (partial, not yet handled): `Required` is passed through to CreateTable ([GT-S3T-037](service.md#gt-s3t-037) (controller hooks catalog entry)) but nothing verifies
  it. Evidence: required=true -> true, required=false and omitted -> false in metadata.json (n=1); no s3tables operation
  returns the schema, so verification needs s3:GetObject on the warehouse bucket.
- [S3T-TABLE-313](#s3t-table-313) (partial, not yet handled): the `type` go_tag exposes `type` on Table and SchemaField, but no API output
  carries field types. Evidence: GetTable `type='customer'`, `format='ICEBERG'`, ListTables agrees; the model walk finds
  no schema-bearing output shape; field type strings live only in metadata.json, so `spec.metadata.iceberg.schema` can
  only be diffed against metadata.json.
- [S3T-TABLE-314](#s3t-table-314) (partial, low impact, not yet handled): a user-set `SchemaField.id` is not honored. Evidence: ids 5/7/9 ->
  1/2/3 and 'decimal(10,2)' -> 'decimal(10, 2)' in metadata.json (n=1); since no API output returns the schema this never
  produces a delta - a documentation gap, not a reconcile defect.

### Where next

- `table-metadata-location.md` carries the other half of the shared model: `metadataLocation` absent on a no-schema table
  and added by the first commit (its -014, -123), token = f(full location) with the rollback and CAS semantics (-134,
  -302, -108, -192), the prefix and '.metadata.json' checks (-105, -107), the 409 messages, the DeleteTable and RenameTable
  series; rule 11 here is the CreateTable-with-schema side of its -014/-315 "metadataLocation only once the table has
  metadata".
- `table.md` (catch-all): identity by ARN vs triple, rename, 404 message shapes (FU2 "GetTable 404 message depends on name
  shape"), tags, concurrency. `table-encryption-storage.md`: the GetTable key set (-308), the creation-time snapshot of
  bucket defaults, and the `ServerSideEncryption=AES256` seen on the metadata.json GetObject of -312.
- `table-policy.md`: the IAM actions CreateTable-with-schema and the sub-reads need (its description: PutTableData for
  schemas) and the AccessDeniedException on UpdateTableMetadataLocation (-132). `table-maintenance.md`: sort/z-order
  compaction needs a metadata location (its -233 correction) and a metadata-only commit as the job trigger.
- Long catalogues render in `details/<id>.md` rather than inline: the v1 type-string matrix ([S3T-TABLE-038](#s3t-table-038)), the partition
  transform matrix (-073), the writeOrder field and orderId matrices (-076, -075), field-id renumbering (-043), schemaV2
  validation (-042, -044, -046), the empty-shape matrices (-036, -133, -284, -285) and the round-trip metadata.json dumps
  (-016, -312, -322). The TRUE doc-claim confirmations (-283, -286, -287, -288, -290) restate rules 4, 8, 9 and 11 and
  belong in the appendix; the FALSE ones (-284, -285, -289) are rules 8, 9 and 4.
- Open experiments: -312/-313/-314 come from the follow-up round-trip probe (FU-08/FU-10, one table each); -036/-133
  depend on botocore validation being bypassed, so whether the Go SDK v2 tagged-union encoding can emit `schema={}` or
  `metadata={}` at all is untested (-311 saw only the client-side ParamValidationError). n=1 and single-run items are in
  `table-schema.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 37 (high 10 / medium 16 / low 11); duplicates folded into the appendix: 1
- handling: handled 7 · partial 3 · tracked 0 · unhandled 26 · suspect-bug 1 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 1 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: request-validation 13, normalization 12, other 5, read-gap 2, response-fidelity 2,
  server-default 2, error-code 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Errors

- <a id="s3t-table-133"></a>**S3T-TABLE-133** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created; partitionSpec={} and writeOrder={orderId:1} -> 200**
  metadata.iceberg.schema={} (the SDK would reject this client-side; sent with validation disabled) ->
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' /
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' /
  InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.'. schemaV2={type:struct}
  without fields -> BadRequestException (HTTP 400) 'The specified metadata is not valid'. partitionSpec={} ->
  HTTP 200. writeOrder={orderId:1} without fields -> HTTP 200. No table was created by any 5xx/4xx attempt
  (nothing registered).
  - ACK: terminal_codes, requeue · ops: CreateTable · fields: metadata.iceberg.schema,
    metadata.iceberg.schemaV2.fields, metadata.iceberg.partitionSpec, metadata.iceberg.writeOrder.fields
  - repro: CreateTable(format=ICEBERG, metadata={iceberg:{schema:{}}}) with botocore
    parameter_validation=False, three times
  - measurements: attempts=3, http_500=3, n=3
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLE-036](#s3t-table-036), [S3T-TABLE-311](#s3t-table-311), [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-072](#s3t-table-072), [S3T-TABLE-073](#s3t-table-073),
    [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-077](#s3t-table-077), [S3T-TABLE-018](#s3t-table-018), [S3T-TABLE-129](table-metadata-location.md#s3t-table-129), [S3T-TABLE-193](table-metadata-location.md#s3t-table-193) ·
    hypotheses: H-T-107 · evidence: table/creative/token-and-suffix-followup
  - notes: Follow-up to the schema-variants probe: 3 of 3 fresh attempts (plus the earlier one) returned
    InternalServerErrorException (HTTP 500) 'An internal error occurred. Try again.' and no table was created;
    schemaV2={type:struct} without fields -> BadRequestException 400 'The specified metadata is not...
  - full notes: [details/S3T-TABLE-133.md](details/S3T-TABLE-133.md)

## Request validation

- <a id="s3t-table-036"></a>**S3T-TABLE-036** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable schema.fields=[] -> 200 (empty-schema table committed); schema={} -> 500; iceberg={}, metadata={}, properties-only -> 400**
  With SDK validation disabled: metadata.iceberg.schema.fields=[] -> HTTP 200 accepted (metadata.json
  fields=[], last-column-id=0). schema={} -> InternalServerErrorException (HTTP 500) 'An internal error
  occurred. Try again.'. iceberg={} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  metadata={} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  iceberg={properties:{...}} without any schema -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid'.
  - ACK: terminal_codes, hook:sdk_create_pre_build_request · ops: CreateTable · fields: metadata,
    metadata.iceberg, metadata.iceberg.schema, metadata.iceberg.schema.fields, metadata.iceberg.properties
  - repro: CreateTable(format=ICEBERG) with each of the five metadata shapes above
    (parameter_validation=False); GetTable afterwards
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-311](#s3t-table-311), [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-129](table-metadata-location.md#s3t-table-129), [S3T-TABLE-193](table-metadata-location.md#s3t-table-193) ·
    hypotheses: H-T-107, H-T-113 · evidence: table/weird-inputs/schema-variants
  - notes: H-T-107 REFUTED: fields=[] -> HTTP 200 and a committed metadata.json with schemas[0].fields=[] and
    last-column-id=0 (table then deleted); schema={} (fields omitted on the wire) -> HTTP 500
    InternalServerErrorException 'An internal error occurred. Try again.' with no table created - a
    controller...
  - full notes: [details/S3T-TABLE-036.md](details/S3T-TABLE-036.md)

- <a id="s3t-table-037"></a>**S3T-TABLE-037** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable v1 schema with two fields named 'a' -> BadRequestException 400; 'a' and 'A' -> accepted**
  fields=[{a,int},{a,string}] -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  fields=[{a,int},{A,string}] -> HTTP 200 accepted (metadata.json fields=[{'id': 1, 'name': 'a', 'type':
  'int', 'required': False}, {'id': 2, 'name': 'A', 'type': 'string', 'required': False}], last-column-id=2).
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schema.fields.name
  - repro: CreateTable with the two field lists above; GetObject metadata.json for the accepted one
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-040](#s3t-table-040), [S3T-TABLE-314](#s3t-table-314), [S3T-TABLE-313](#s3t-table-313) · hypotheses: H-T-108 ·
    evidence: table/weird-inputs/schema-variants
  - notes: H-T-108 confirmed. One attempt each.

- <a id="s3t-table-039"></a>**S3T-TABLE-039** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **v1 schema nested type strings: struct<a:int> -> BadRequestException 400, list<string> -> BadRequestException 400...**
  type 'struct<a:int>' -> BadRequestException (HTTP 400) 'The specified metadata is not valid'; type
  'list<string>' -> BadRequestException (HTTP 400) 'The specified metadata is not valid'; type
  'map<string,int>' -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schema.fields.type,
    metadata.iceberg.schemaV2
  - repro: CreateTable v1 schema with one field of each nested type string
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-037](#s3t-table-037), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-040](#s3t-table-040), [S3T-TABLE-314](#s3t-table-314), [S3T-TABLE-313](#s3t-table-313), [S3T-TABLE-041](#s3t-table-041),
    [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-043](#s3t-table-043), [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-046](#s3t-table-046) · hypotheses: H-T-110 ·
    evidence: table/weird-inputs/schema-variants
  - notes: H-T-110 confirmed. One attempt each. Full original title: v1 schema nested type strings:
    struct<a:int> -> BadRequestException 400, list<string> -> BadRequestException 400, map<string,int> ->
    BadRequestException 400

- <a id="s3t-table-041"></a>**S3T-TABLE-041** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable with BOTH schema and schemaV2 -> BadRequestException 400; schemaV2 alone -> accepted**
  schema=[{a,int}] + schemaV2{struct,[{1,b,string}]} -> BadRequestException (HTTP 400) 'Both 'schema' and
  'schemaV2' were specified. Use either 'schema' or 'schemaV2' and retry the create table command.'. schemaV2
  only -> HTTP 200 accepted (metadata.json fields=[{'id': 1, 'name': 'b', 'type': 'string', 'required':
  False}], last-column-id=1). iceberg={} -> BadRequestException 400; properties-only -> BadRequestException
  400 (see the empty-structures finding).
  - ACK: terminal_codes, hook:sdk_create_pre_build_request · ops: CreateTable · fields:
    metadata.iceberg.schema, metadata.iceberg.schemaV2
  - repro: CreateTable with schema and schemaV2 both set; GetObject metadata.json if accepted
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-036](#s3t-table-036), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-311](#s3t-table-311), [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-043](#s3t-table-043),
    [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-046](#s3t-table-046), [S3T-TABLE-039](#s3t-table-039) · hypotheses: H-T-113, H-T-114 · evidence:
    table/weird-inputs/schema-variants
  - notes: H-T-113 confirmed (both-set rejected); H-T-114 refuted. One attempt each.

- <a id="s3t-table-044"></a>**S3T-TABLE-044** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **schemaV2.identifierFieldIds: required field [1] -> accepted; optional field [2] or unknown [9] -> BadRequestException 400**
  fields [{1,a,int,required},{2,b,string,optional}]: identifierFieldIds=[1] -> HTTP 200 accepted
  (metadata.json fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': True}, {'id': 2, 'name': 'b',
  'type': 'string', 'required': False}], last-column-id=2) (identifier-field-ids written: [1]). [2] ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'. [9] -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schemaV2.identifierFieldIds
  - repro: CreateTable schemaV2 with identifierFieldIds [1], [2], [9]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-043](#s3t-table-043), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-046](#s3t-table-046), [S3T-TABLE-039](#s3t-table-039) ·
    hypotheses: H-T-117 · evidence: table/weird-inputs/schema-variants
  - notes: H-T-117 confirmed. One attempt each. Full original title: schemaV2.identifierFieldIds: required
    field [1] -> accepted; optional field [2] -> BadRequestException 400; unknown [9] -> BadRequestException
    400

- <a id="s3t-table-046"></a>**S3T-TABLE-046** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **schemaV2 field type as a JSON document (nested struct + list) -> accepted; 'varchar' -> 400; field without 'required' -> 400 (no default)**
  schemaV2 fields [{1,s,{type:struct,fields:[{2,x,int}]}},{3,l,{type:list,element-id:4,element:string}}] ->
  HTTP 200 accepted (metadata.json fields=[{'id': 1, 'name': 's', 'type': {'type': 'struct', 'fields': [{'id':
  3, 'name': 'x', 'required': False, 'type': 'int'}]}, 'required': False}, {'id': 2, 'name': 'l', 'type':
  {'type': 'list', 'element-id': 4, 'element': 'string', 'element-required': False}, 'required': False}],
  last-column-id=4). schemaV2 field type 'varchar' -> BadRequestException (HTTP 400) 'The specified metadata
  is not valid'. schemaV2 field without 'required' (model: required) -> BadRequestException (HTTP 400) 'The
  specified metadata is not valid'.
  - ACK: docs-only, is_document · ops: CreateTable · fields: metadata.iceberg.schemaV2.fields.type,
    metadata.iceberg.schemaV2.fields.required
  - repro: CreateTable schemaV2 with a struct-typed and a list-typed field (Document type); GetObject
    metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-043](#s3t-table-043), [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-039](#s3t-table-039) ·
    hypotheses: H-T-110, H-T-111 · evidence: table/weird-inputs/schema-variants
  - notes: H-T-110 complement: nested struct/list types are accepted through schemaV2 (Document-typed 'type')
    and written as Iceberg nested types (ids reassigned); schemaV2 field type 'varchar' -> 400; schemaV2 field
    without 'required' -> 400 (not defaulted). H-T-111 qualified: in schemaV2 'required' must be...
  - full notes: [details/S3T-TABLE-046.md](details/S3T-TABLE-046.md)

- <a id="s3t-table-072"></a>**S3T-TABLE-072** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec sourceId: unknown 99, 0 or past-last 3 -> 400 'metadata is not valid'; id 1 accepted on schemaV2 and server-numbered v1 schema**
  schemaV2 {1 ts timestamp, 2 id long} + partitionSpec [{sourceId:99, fieldId:1000, name:p, transform:day}] ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'. sourceId:1 -> HTTP 200 accepted ->
  partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id": 1,
  "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. v1 schema [{ts,timestamp},{id,long}]
  (server-assigned ids) + sourceId:1 -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields":
  [{"name": "ts_day", "transform": "day", "source-id": 1, "field-id": 1000}]}] default-spec-id=0
  last-partition-id=1000. sourceId:0 -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  sourceId:3 -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.sourceId, metadata.iceberg.schema
  - repro: CreateTable with the partitionSpec variants above; GetObject metadata.json for accepted ones
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-290](#s3t-table-290), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-289](#s3t-table-289), [S3T-TABLE-073](#s3t-table-073), [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-133](#s3t-table-133) ·
    hypotheses: H-T-119, H-T-120 · evidence: table/weird-inputs/partition-sort-variants
  - notes: H-T-119 confirmed (unknown sourceId 99 -> 400 'The specified metadata is not valid' - the message
    does NOT name the source id; sourceId 1 -> 200, partition-specs written with source-id 1, field-id 1000).
    H-T-120 confirmed: with the v1 schema sourceId 1 maps to the first (server-numbered) field;...
  - full notes: [details/S3T-TABLE-072.md](details/S3T-TABLE-072.md)

- <a id="s3t-table-077"></a>**S3T-TABLE-077** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Empty writeOrder ({fields:[]}, {orderId:0|1, fields:[]}, {}) -> 200, written as the unsorted order (order-id 0, default-sort-order-id 0)**
  writeOrder={fields:[]} -> HTTP 200 accepted -> sort-orders=[{"order-id": 0, "fields": []}]
  default-sort-order-id=0. {orderId:0, fields:[]} -> HTTP 200 accepted -> sort-orders=[{"order-id": 0,
  "fields": []}] default-sort-order-id=0. {orderId:1, fields:[]} -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 0, "fields": []}] default-sort-order-id=0. writeOrder={} -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 0, "fields": []}] default-sort-order-id=0. (schema.fields=[] is covered by the
  schema-variants probe.)
  - ACK: terminal_codes, hook:sdk_create_pre_build_request · ops: CreateTable · fields:
    metadata.iceberg.writeOrder, metadata.iceberg.writeOrder.fields
  - repro: CreateTable with the empty writeOrder shapes above (parameter_validation=False); GetObject
    metadata.json if accepted
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-018](#s3t-table-018) · hypotheses: H-T-196 · evidence:
    table/weird-inputs/partition-sort-variants
  - notes: H-T-196 QUALIFIED: the accept part holds (no 400, unlike nothing - note schema.fields=[] is also
    accepted per the schema-variants probe), but the written sort order is {order-id 0, fields []} with
    default-sort-order-id 0 for all four shapes, including orderId=1 - i.e. an empty writeOrder is...
  - full notes: [details/S3T-TABLE-077.md](details/S3T-TABLE-077.md)

- <a id="s3t-table-284"></a>**S3T-TABLE-284** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C092 FALSE: CreateTable with partitionSpec present but fields omitted/empty -> BadRequestException (all accepted)**
  CreateTable(format=ICEBERG, metadata.iceberg.schema {id int required, ts timestamp} + partitionSpec
  variant), fields-less shapes sent with botocore parameter validation OFF (the SDK alone rejects a missing
  required 'fields' client-side): partitionSpec {} -> HTTP 200 (table created; metadata.json
  partition-specs=[{"spec-id": 0, "fields": []}] default-spec-id=0 last-partition-id=999; deleted: HTTP 204);
  partitionSpec {"specId": 0} -> HTTP 200 (table created; metadata.json partition-specs=[{"spec-id": 0,
  "fields": []}] default-spec-id=0 last-partition-id=999; deleted: HTTP 204); partitionSpec {"fields": []} ->
  HTTP 200 (table created; metadata.json partition-specs=[{"spec-id": 0, "fields": []}] default-spec-id=0
  last-partition-id=999; deleted: HTTP 204).
  - ACK: docs-only, compare.nil_equals_zero_value · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields, metadata.iceberg.partitionSpec.specId
  - repro: CreateTable with partitionSpec {} / {specId: 0} (parameter_validation=False) / {fields: []} ->
    GetTableMetadataLocation -> s3 GetObject metadata.json -> DeleteTable
  - measurements: variants=3, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289),
    [S3T-TABLE-075](#s3t-table-075) · evidence: service/static/doc-claims-3
  - notes: VERDICT: FALSE - partitionSpec {}, {specId: 0} and {fields: []} are all ACCEPTED (200) and written
    as an empty spec-id 0.

- <a id="s3t-table-285"></a>**S3T-TABLE-285** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C093 FALSE: CreateTable with writeOrder present but fields omitted/empty -> BadRequestException (all 4 shapes accepted, order 0)**
  writeOrder={fields: []}, {orderId: 0, fields: []}, {orderId: 1, fields: []} and {} (fields omitted on the
  wire) -> HTTP 200 accepted; the metadata.json then has sort-orders=[{'order-id': 0, 'fields': []}] and
  default-sort-order-id=0, i.e. an empty writeOrder is equivalent to omitting it. No BadRequestException for
  any of the four shapes.
  - ACK: docs-only, compare.nil_equals_zero_value · ops: CreateTable · fields:
    metadata.iceberg.writeOrder.fields
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-077](#s3t-table-077), [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289),
    [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-075](#s3t-table-075) · evidence: service/static/doc-claims-3
  - notes: VERDICT: FALSE - All four fields-less writeOrder shapes accepted (200) and written as the unsorted
    order.

- <a id="s3t-table-311"></a>**S3T-TABLE-311** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable metadata={} -> client-side ParamValidationError (validation on) / BadRequestException 400 'metadata is not valid' (off)**
  With botocore parameter validation ON, CreateTable(metadata={}) -> client-side ParamValidationError:
  Parameter validation failed: Must set one of the following keys for tagged unionstructure metadata:
  iceberg.. With validation OFF the same request -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid' (second attempt: BadRequestException (HTTP 400)); wire body {"name": "ackq_8d6c9c_empty_off1",
  "format": "ICEBERG", "metadata": {}}. metadata={iceberg:{}}: validation on -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'; off -> BadRequestException (HTTP 400).
  metadata.iceberg.schema.fields=[] -> HTTP 200. A field with type 'bogus' -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'. Unknown union key {notIceberg:{schema}} (validation off) ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'; iceberg plus an unknown sibling key ->
  HTTP 200; schema and schemaV2 both set -> BadRequestException (HTTP 400) 'Both 'schema' and 'schemaV2' were
  specified. Use either 'schema' or 'schemaV2' and retry the create table command.'; JSON null metadata on the
  wire -> HTTP 200.
  - ACK: terminal_codes, input_wrapper_field_path · ops: CreateTable · fields: metadata, metadata.iceberg,
    metadata.iceberg.schema.fields, metadata.iceberg.schema.fields.type
  - repro: CreateTable(format=ICEBERG, metadata={}) with botocore parameter_validation True then False; then
    the variants listed in behavior
  - measurements: n_empty_validation_off=2, n_per_other_cell=1, n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-036](#s3t-table-036), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-015](#s3t-table-015), [S3T-TABLE-034](table-metadata-location.md#s3t-table-034),
    [S3T-TABLE-019](table-metadata-location.md#s3t-table-019), [S3T-TABLE-123](table-metadata-location.md#s3t-table-123), [S3T-TABLE-316](table-metadata-location.md#s3t-table-316) · evidence: table/weird-inputs/fu-03-metadata-empty-union
  - notes: FU-03 items 2-3. The Go SDK v2 tagged-union (interface) encoding cannot be observed from Python;
    what IS observable is whether the server accepts an empty union object and the wire code it returns. For
    every accepted cell GetTable.metadataLocation present (= a schema was applied):...
  - full notes: [details/S3T-TABLE-311.md](details/S3T-TABLE-311.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-016"></a>**S3T-TABLE-016** `server-default` · impact medium · handled · verified 2026-10-13
  **v1 schema: metadata.json fields get ids [1, 2, 3], required [False, True, False]...**
  For CreateTable metadata.iceberg.schema.fields=[{fld_alpha,int}(required
  omitted),{fld_bravo,string,required=true},{fld_charlie,long,required=false}] the metadata.json at
  metadataLocation has format-version=2, current-schema-id=0, schemas=[{'schema-id': 0, 'type': 'struct',
  'n_fields': 3}], last-column-id=3, partition-specs=[{'spec-id': 0, 'fields': []}] (default-spec-id=0,
  last-partition-id=999), sort-orders=[{'order-id': 0, 'fields': []}] (default-sort-order-id=0),
  current-snapshot-id=-1, snapshots=[], identifier-field-ids=None. Written fields: [{'id': 1, 'name':
  'fld_alpha', 'required': False, 'type': 'int'}, {'id': 2, 'name': 'fld_bravo', 'required': True, 'type':
  'string'}, {'id': 3, 'name': 'fld_charlie', 'required': False, 'type': 'long'}].
  - ACK: docs-only, is_immutable · ops: CreateTable · fields: metadata.iceberg.schema.fields,
    metadata.iceberg.schema.fields.required, metadata.iceberg.writeOrder.orderId
  - repro: CreateTable with the 3-field v1 schema above -> GetTable.metadataLocation -> S3 GetObject ->
    inspect schemas/sort-orders/partition-specs
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; 5c1bda2; 007c152`
  - related: [S3T-TABLE-199](table-maintenance.md#s3t-table-199), [S3T-TABLE-203](table-maintenance.md#s3t-table-203), [S3T-TABLE-204](table-maintenance.md#s3t-table-204), [S3T-TABLE-211](table-maintenance.md#s3t-table-211) · hypotheses: H-T-111, H-T-112,
    H-T-195 · evidence: table/round-trip/metadata-create-readback
  - notes: H-T-111 confirmed (required omitted -> False; required=true -> True; no extra identifier field:
    identifier-field-ids=None). H-T-112 confirmed (ids [1, 2, 3], schema-id 0, format-version 2). H-T-195
    confirmed (writeOrder omitted -> default-sort-order-id=0, sort-orders=[{'order-id': 0, 'fields':...
  - full notes: [details/S3T-TABLE-016.md](details/S3T-TABLE-016.md)

- <a id="s3t-table-018"></a>**S3T-TABLE-018** `read-gap` · impact high · handled · verified 2026-10-15, re-verified
  **metadata.iceberg (schema/partitionSpec/writeOrder/properties) is write-only: GetTable, GetTableMetadataLocation, ListTables echo none of it**
  CreateTable with schema fields evt_ts_qx/evt_id_qx, partitionSpec(day on evt_ts_qx), writeOrder(identity asc
  nulls-first on evt_id_qx) and a marker property -> HTTP 200. The concatenated raw bodies of GetTable +
  GetTableMetadataLocation + ListTables contain none of the field names, partition name, property marker,
  transform/direction/null-order words or any orderId/sort/partition key (the only token hit, 'schema', is the
  name of a sibling table 'ackq_<run>_schema' in the ListTables body). The metadata.json at metadataLocation
  contains them: {'evt_ts_qx': True, 'evt_id_qx': True, 'evt_ts_day_qx': True, 'ackq_prop_marker': True,
  'order-id': True, 'sort-order': True, 'partition': True, 'schema': True, 'properties': True, 'identity':
  True, 'nulls-first': True, 'timestamp': True} (schemas=[{'type': 'struct', 'schema-id': 0, 'fields': [{'id':
  1, 'name': 'evt_ts_qx', 'required': False, 'type': 'timestamp'}, {'id': 2, 'name': 'evt_id_qx', 'required':
  False, 'type': 'long'}]}]; partition-specs=[{'spec-id': 0, 'fields': [{'name': 'evt_ts_day_qx', 'transform':
  'day', 'source-id': 1, 'field-id': 1000}]}]; sort-orders=[{'order-id': 1, 'fields': [{'transform':
  'identity', 'source-id': 2, 'direction': 'asc', 'null-order': 'nulls-first'}]}]; default-sort-order-id=1).
  - ACK: compare.is_ignored+delta_pre_compare, is_immutable, docs-only · ops: CreateTable, GetTable,
    GetTableMetadataLocation, ListTables · fields: metadata.iceberg.schema, metadata.iceberg.partitionSpec,
    metadata.iceberg.writeOrder, metadata.iceberg.properties
  - repro: CreateTable with schema+partitionSpec+writeOrder+properties -> dump
    GetTable/GetTableMetadataLocation/ListTables -> grep field names -> S3 GetObject metadataLocation
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-310](#s3t-table-310), [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-077](#s3t-table-077), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-017](#s3t-table-017),
    [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-313](#s3t-table-313), [S3T-TABLE-090](table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-028](table-metadata-location.md#s3t-table-028) · hypotheses:
    H-T-103, H-T-192 · evidence: table/round-trip/metadata-create-readback
  - notes: H-T-103 confirmed. H-T-192 confirmed: writeOrder with orderId omitted (SDK validation bypassed; the
    model marks orderId required) -> HTTP 200; metadata.json default-sort-order-id=1, sort-orders
    order-ids=[1]; no control-plane read exposes orderId. n=1 table. Side observation: S3 ListObjectsV2 on
    the...
  - full notes: [details/S3T-TABLE-018.md](details/S3T-TABLE-018.md)

- <a id="s3t-table-038"></a>**S3T-TABLE-038** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **v1 SchemaField.type: 14 Iceberg primitives + 'INT'/'String' accepted (lower-cased; decimal(10,2) re-spaced); integer/varchar/bigint/'' 400**
  CreateTable with a one-field v1 schema per type string. Accepted (and written to metadata.json as): {"int":
  [["c1", "int"]], "long": [["c1", "long"]], "string": [["c1", "string"]], "boolean": [["c1", "boolean"]],
  "timestamp": [["c1", "timestamp"]], "timestamptz": [["c1", "timestamptz"]], "decimal(10,2)": [["c1",
  "decimal(10, 2)"]], "fixed[16]": [["c1", "fixed[16]"]], "date": [["c1", "date"]], "time": [["c1", "time"]],
  "uuid": [["c1", "uuid"]], "binary": [["c1", "binary"]], "float": [["c1", "float"]], "double": [["c1",
  "double"]], "INT": [["c1", "int"]], "String": [["c1", "string"]]}. Rejected: {"integer":
  "BadRequestException 400 'The specified metadata is not valid'", "varchar": "BadRequestException 400 'The
  specified metadata is not valid'", "": "BadRequestException 400 'The specified metadata is not valid'", "
  int ": "BadRequestException 400 'The specified metadata is not valid'", "decimal": "BadRequestException 400
  'The specified metadata is not valid'", "fixed": "BadRequestException 400 'The specified metadata is not
  valid'", "bigint": "BadRequestException 400 'The specified metadata is not valid'"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schema.fields.type
  - repro: Loop CreateTable over the type strings above (one field each); GetObject metadata.json of accepted
    ones; DeleteTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-037](#s3t-table-037), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-040](#s3t-table-040), [S3T-TABLE-314](#s3t-table-314), [S3T-TABLE-313](#s3t-table-313), [S3T-TABLE-073](#s3t-table-073),
    [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-042](#s3t-table-042) · hypotheses: H-T-109 · evidence: table/weird-inputs/schema-variants
  - notes: H-T-109 QUALIFIED: all 14 Iceberg primitives accepted; 'integer','varchar','bigint','' and ' int '
    (whitespace not trimmed), bare 'decimal'/'fixed' rejected with BadRequestException 400; but upper/mixed
    case 'INT' and 'String' are ACCEPTED and written lower-cased ('int','string'), and...
  - full notes: [details/S3T-TABLE-038.md](details/S3T-TABLE-038.md)

- <a id="s3t-table-043"></a>**S3T-TABLE-043** `normalization` · impact high · handled · verified 2026-10-13
  **schemaV2 field ids are RENUMBERED 1..n (sparse 1,5,42 -> 1,2,3; 0 and -1 accepted); only duplicate ids rejected (400)**
  ids {1,1} -> BadRequestException (HTTP 400) 'The specified metadata is not valid'. ids {0,1} -> HTTP 200
  accepted (metadata.json fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name':
  'b', 'type': 'string', 'required': False}], last-column-id=2). ids {-1,1} -> HTTP 200 accepted
  (metadata.json fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name': 'b',
  'type': 'string', 'required': False}], last-column-id=2). ids {1,5,42} -> HTTP 200 accepted (metadata.json
  fields=[{'id': 1, 'name': 'a', 'type': 'int', 'required': False}, {'id': 2, 'name': 'b', 'type': 'string',
  'required': False}, {'id': 3, 'name': 'c', 'type': 'long', 'required': False}], last-column-id=3).
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schemaV2.fields.id
  - repro: CreateTable schemaV2 with the four id sets; GetObject metadata.json for accepted ones
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; 5c1bda2; 007c152`
  - related: [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-046](#s3t-table-046), [S3T-TABLE-039](#s3t-table-039) ·
    hypotheses: H-T-116 · evidence: table/weird-inputs/schema-variants
  - notes: H-T-116 REFUTED in its pass-through part: ids {0,1} and {-1,1} are accepted (written 1,2), sparse
    {1,5,42} accepted but written as 1,2,3 with last-column-id=3; duplicates {1,1} -> 400 'The specified
    metadata is not valid'. Nested ids are also reassigned (struct child sent id 2 -> written 3; list...
  - full notes: [details/S3T-TABLE-043.md](details/S3T-TABLE-043.md)

- <a id="s3t-table-073"></a>**S3T-TABLE-073** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec transform not grammar-checked: 'bucket' (no arity), 'days', '' written verbatim, 'DAY' lower-cased; only day(long) -> 400**
  On schemaV2 {1 ts timestamp, 2 id long}, one partition field per attempt. Accepted (partition-specs as
  written): {"identity(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "identity", "source-id":
  1, "field-id": 1000}]}], "year(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "year",
  "source-id": 1, "field-id": 1000}]}], "month(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform":
  "month", "source-id": 1, "field-id": 1000}]}], "day(ts)": [{"spec-id": 0, "fields": [{"name": "pcol",
  "transform": "day", "source-id": 1, "field-id": 1000}]}], "hour(ts)": [{"spec-id": 0, "fields": [{"name":
  "pcol", "transform": "hour", "source-id": 1, "field-id": 1000}]}], "bucket\[16\](id)": [{"spec-id": 0,
  "fields": [{"name": "pcol", "transform": "bucket[16]", "source-id": 2, "field-id": 1000}]}],
  "truncate\[4\](id)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "truncate[4]", "source-id": 2,
  "field-id": 1000}]}], "void(ts)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform": "void",
  "source-id": 1, "field-id": 1000}]}], "bucket(id)": [{"spec-id": 0, "fields": [{"name": "pcol", "transform":
  "bucket", "source-id": 2, "field-id": 1000}]}], "days(ts)": [{"spec-id": 0, "fields": [{"name": "pcol",
  "transform": "days", "source-id": 1, "field-id": 1000}]}], "DAY(ts)": [{"spec-id": 0, "fields": [{"name":
  "pcol", "transform": "day", "source-id": 1, "field-id": 1000}]}], "(ts)": [{"spec-id": 0, "fields":
  [{"name": "pcol", "transform": "", "source-id": 1, "field-id": 1000}]}]}. Rejected: {"day(id)":
  "BadRequestException 400 'The specified metadata is not valid'"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.transform
  - repro: Loop CreateTable over the transforms above; GetObject metadata.json; DeleteTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-072](#s3t-table-072), [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-042](#s3t-table-042) ·
    hypotheses: H-T-121 · evidence: table/weird-inputs/partition-sort-variants
  - notes: H-T-121 REFUTED for the grammar part: identity/year/month/day/hour/bucket[16]/truncate[4]/void
    accepted as expected, but 'bucket' (no arity), 'days' and the empty string '' were ALSO accepted and
    persisted verbatim in partition-specs (an Iceberg engine cannot parse 'bucket' or ''), and 'DAY' was...
  - full notes: [details/S3T-TABLE-073.md](details/S3T-TABLE-073.md)

- <a id="s3t-table-074"></a>**S3T-TABLE-074** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **partitionSpec fieldId (0, 5, 2) and specId (3) IGNORED (field-id 1000,1001..., spec-id 0); fields=[] accepted as an unpartitioned spec**
  fieldId omitted -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day",
  "transform": "day", "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=0
  -> HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day",
  "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=5 -> HTTP 200
  accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id":
  1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. fieldId=2 (equals a schema field id) ->
  HTTP 200 accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day",
  "source-id": 1, "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. specId=3 -> HTTP 200 accepted
  -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id": 1,
  "field-id": 1000}]}] default-spec-id=0 last-partition-id=1000. Two fields without fieldId -> HTTP 200
  accepted -> partition-specs=[{"spec-id": 0, "fields": [{"name": "ts_day", "transform": "day", "source-id":
  1, "field-id": 1000}, {"name": "id_bucket", "transform": "bucket[8]", "source-id": 2, "field-id": 1001}]}]
  default-spec-id=0 last-partition-id=1001. partitionSpec.fields=[] -> HTTP 200 accepted ->
  partition-specs=[{"spec-id": 0, "fields": []}] default-spec-id=0 last-partition-id=999.
  - ACK: docs-only, late_initialize · ops: CreateTable · fields:
    metadata.iceberg.partitionSpec.fields.fieldId, metadata.iceberg.partitionSpec.specId
  - repro: CreateTable with the partitionSpec variants above; GetObject metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289),
    [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-072](#s3t-table-072), [S3T-TABLE-073](#s3t-table-073), [S3T-TABLE-133](#s3t-table-133) · hypotheses: H-T-122 · evidence:
    table/weird-inputs/partition-sort-variants
  - notes: H-T-122 QUALIFIED/REFUTED: fieldId omitted -> 1000 (confirmed), but explicit fieldId 0, 5 and 2 are
    all rewritten to 1000 (not written verbatim), specId=3 is written as spec-id 0 / default-spec-id 0 (not
    3); two fields without fieldId -> 1000, 1001 with last-partition-id 1001;...
  - full notes: [details/S3T-TABLE-074.md](details/S3T-TABLE-074.md)

- <a id="s3t-table-075"></a>**S3T-TABLE-075** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **writeOrder.orderId IGNORED (omitted/2/7/-1 -> order-id 1) except orderId=0 with fields -> 400; unknown sourceId -> 400**
  schemaV2 {1 ts timestamp, 2 id long}; sort field identity/asc/nulls-first on id unless noted. sourceId:9 ->
  BadRequestException (HTTP 400) 'The specified metadata is not valid'. orderId omitted -> HTTP 200 accepted
  -> sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=0 (non-empty fields) -> BadRequestException
  (HTTP 400) 'The specified metadata is not valid'. orderId=2 -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=7 -> HTTP 200 accepted ->
  sort-orders=[{"order-id": 1, "fields": [{"transform": "identity", "source-id": 2, "direction": "asc",
  "null-order": "nulls-first"}]}] default-sort-order-id=1. orderId=-1 -> HTTP 200 accepted. Two sort fields
  (day(ts) asc nulls-first, id desc nulls-last) -> HTTP 200 accepted -> sort-orders=[{"order-id": 1, "fields":
  [{"transform": "day", "source-id": 1, "direction": "asc", "null-order": "nulls-first"}, {"transform":
  "identity", "source-id": 2, "direction": "desc", "null-order": "nulls-last"}]}] default-sort-order-id=1.
  - ACK: terminal_codes, docs-only, compare.nil_equals_zero_value · ops: CreateTable · fields:
    metadata.iceberg.writeOrder.orderId, metadata.iceberg.writeOrder.fields.sourceId
  - repro: CreateTable with the writeOrder variants above (parameter_validation=False for the omitted
    orderId); GetObject metadata.json
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289),
    [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-076](#s3t-table-076), [S3T-TABLE-077](#s3t-table-077), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-018](#s3t-table-018) · hypotheses: H-T-123, H-T-193,
    H-T-194 · evidence: table/weird-inputs/partition-sort-variants
  - notes: H-T-123 QUALIFIED: unknown sourceId 9 -> 400 'The specified metadata is not valid' (confirmed);
    omitted orderId -> order-id 1 / default-sort-order-id 1 (confirmed); but orderId=2 is written as order-id
    1 (not 2). H-T-193 REFUTED: orderId=7 -> accepted and written as order-id 1 (not honored);...
  - full notes: [details/S3T-TABLE-075.md](details/S3T-TABLE-075.md)

- <a id="s3t-table-076"></a>**S3T-TABLE-076** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **writeOrder field: direction/nullOrder omitted -> 400 (no default); 'ASC' lower-cased; 'NULLS_FIRST' -> 400; transform 'days' kept verbatim**
  With SDK validation disabled: direction omitted -> BadRequestException (HTTP 400) 'The specified metadata is
  not valid'. nullOrder omitted -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  direction='ASC' -> HTTP 200 accepted -> sort-orders=[{"order-id": 1, "fields": [{"transform": "identity",
  "source-id": 2, "direction": "asc", "null-order": "nulls-first"}]}] default-sort-order-id=1.
  nullOrder='NULLS_FIRST' -> BadRequestException (HTTP 400) 'The specified metadata is not valid'.
  transform='days' -> HTTP 200 accepted.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.writeOrder.fields.direction,
    metadata.iceberg.writeOrder.fields.nullOrder, metadata.iceberg.writeOrder.fields.transform
  - repro: CreateTable writeOrder with each variant (parameter_validation=False)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-077](#s3t-table-077), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-018](#s3t-table-018), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-073](#s3t-table-073),
    [S3T-TABLE-042](#s3t-table-042) · hypotheses: H-T-124 · evidence: table/weird-inputs/partition-sort-variants
  - notes: H-T-124 QUALIFIED: omitting direction or nullOrder -> BadRequestException 400 'The specified
    metadata is not valid' (confirmed: no defaulting); 'NULLS_FIRST' -> 400 (confirmed); but 'ASC' is ACCEPTED
    and normalized to 'asc' in metadata.json (refutes the upper-case part for direction). Sort-field...
  - full notes: [details/S3T-TABLE-076.md](details/S3T-TABLE-076.md)

- <a id="s3t-table-289"></a>**S3T-TABLE-289** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C097 FALSE: Explicit partitionSpec.specId=3 / writeOrder.orderId=7 recorded as-is (both ignored: spec-id 0, order-id 1 forced)**
  partitionSpec.specId=3 -> HTTP 200 but metadata.json has partition-specs[0].spec-id 0 and default-spec-id 0;
  writeOrder.orderId=7 (and 2, -1) -> HTTP 200 but sort-orders[0].order-id 1 and default-sort-order-id 1. Both
  ids are ignored (not overridable); the only effect of orderId is orderId=0 with non-empty fields ->
  BadRequestException 400 'The specified metadata is not valid'. partitionSpec fieldId values are likewise
  rewritten to 1000, 1001, ...
  - ACK: docs-only · ops: CreateTable · fields: metadata.iceberg.partitionSpec.specId,
    metadata.iceberg.writeOrder.orderId
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287),
    [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-290](#s3t-table-290), [S3T-TABLE-072](#s3t-table-072) · evidence: service/static/doc-claims-3
  - notes: VERDICT: FALSE - The documented defaults are forced, not overridable: 3 -> 0 and 7 -> 1 in the
    written metadata.

- <a id="s3t-table-312"></a>**S3T-TABLE-312** `normalization` · impact medium · partially handled · verified 2026-10-14
  **Schema 'required' round-trips into metadata.json: required=true -> True, required=false -> False, omitted -> False**
  CreateTable(format=ICEBERG,
  metadata.iceberg.schema.fields=[{id:int,required:true},{label:string,required:false},{ts:long (required
  omitted)}]) -> HTTP 200. GetTable.metadataLocation =
  s3://<table-uuid-prefix><opaque>--table-s3/metadata/00000-<uuid>.metadata.json; S3 GetObject on it -> HTTP
  200 (809 bytes, ServerSideEncryption=AES256). The Iceberg metadata (format-version 2, current-schema-id 0, 1
  schema(s), v1 'schema' member present: False) lists fields ['id', 'label', 'ts']: 'required' per field =
  {"id": true, "label": false, "ts": false}, 'type' per field = {"id": "int", "label": "string", "ts":
  "long"}, field ids = {"id": 1, "label": 2, "ts": 3}. The metadata has 0 snapshots,
  location=<warehouseLocation>, properties keys ['write.parquet.compression-codec'].
  - ACK: none, is_immutable · ops: CreateTable, GetTable, GetTableMetadataLocation · fields:
    metadata.iceberg.schema.fields[].required, metadata.iceberg.schema.fields[].type, metadataLocation
  - repro: CreateTable with the three fields -> GetTable -> parse metadataLocation ->
    s3:GetObject(bucket=<warehouse bucket>, key=metadata/00000-<uuid>.metadata.json) ->
    schemas[current].fields[].required
  - handling: partially handled via `5c1bda2; 007c152` - see Handling gaps
  - related: [S3T-TABLE-318](table-maintenance.md#s3t-table-318), [S3T-TABLE-319](table-encryption-storage.md#s3t-table-319) · evidence: table/round-trip/fu-08-schema-required-metadata-json
  - notes: FU-08. required=true is preserved as a non-null column (true), required=false and an omitted
    required both become False in the metadata; the SDK-side pointer/omission handling itself is not
    observable from the API (only its effect in metadata.json is). The schema is never returned by any
    s3tables...
  - full notes: [details/S3T-TABLE-312.md](details/S3T-TABLE-312.md)

- <a id="s3t-table-313"></a>**S3T-TABLE-313** `read-gap` · impact medium · partially handled · verified 2026-10-14
  **Table 'type' is 'customer' in GetTable and ListTables (agree); schema field types live only in metadata.json, in no API output**
  GetTable wire JSON members: ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name',
  'namespace', 'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken',
  'warehouseLocation'] (type='customer', format='ICEBERG'); ListTables summary wire members: ['createdAt',
  'modifiedAt', 'name', 'namespace', 'tableARN', 'type'] (type='customer', format present: False); Get and
  List agree on type: True. No s3tables output shape carries schema/fields (model walk: only
  GetTable.format/type and ListTables.tables[].type exist). The field type strings sent in CreateTable ({"id":
  "int", "label": "string", "ts": "long"}) appear in the Iceberg metadata.json as {"id": "int", "label":
  "string", "ts": "long"}; explicit field ids {"a": 5, "b": 7, "c": 9} sent for the second table came back as
  {"a": 1, "b": 2, "c": 3} with types {"a": "int", "b": "decimal(10, 2)", "c": "timestamp"}.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: GetTable, ListTables, CreateTable · fields:
    type, format, metadata.iceberg.schema.fields[].type, metadata.iceberg.schema.fields[].id
  - repro: CreateTable(ICEBERG, schema) -> GetTable + ListTables(namespace) raw JSON; model walk for
    schema-bearing outputs; s3:GetObject(metadataLocation)
  - handling: partially handled via `generator.yaml:45-46; generator.yaml:60-65` - see Handling gaps
  - related: [S3T-TABLE-037](#s3t-table-037), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-040](#s3t-table-040), [S3T-TABLE-314](#s3t-table-314), [S3T-TABLE-018](#s3t-table-018),
    [S3T-TABLE-310](#s3t-table-310), [S3T-TABLE-090](table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-028](table-metadata-location.md#s3t-table-028), [S3T-TABLE-231](table.md#s3t-table-231), [S3T-TABLE-089](table-metadata-location.md#s3t-table-089) ·
    evidence: table/round-trip/fu-08-schema-required-metadata-json
  - notes: FU-10. The table type member is literally named 'type' with the value shown (TableType enum
    customer|aws); a spec.metadata.iceberg.schema can never be diffed against the API, only against
    metadata.json. Creative extra (second table): explicit SchemaField.id values 5/7/9 were IGNORED and
    renumbered...
  - full notes: [details/S3T-TABLE-313.md](details/S3T-TABLE-313.md)

- <a id="s3t-table-322"></a>**S3T-TABLE-322** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable metadata.iceberg.properties: metadata.json keeps the user property and adds write.parquet.compression-codec=zstd**
  CreateTable(format=ICEBERG, metadata.iceberg with schema [evt_ts_qx timestamp, evt_id_qx long],
  partitionSpec day(evt_ts_qx) as evt_ts_day_qx, writeOrder orderId=1 identity(evt_id_qx) asc nulls-first,
  properties {ackq_prop_marker: ackq_prop_value_qx}) -> HTTP 200. S3 GetObject of
  GetTableMetadataLocation.metadataLocation (s3://<warehouse>--table-s3/metadata/00000-<uuid>.metadata.json)
  -> HTTP 200; metadata.json format-version=2, properties={"ackq_prop_marker": "ackq_prop_value_qx",
  "write.parquet.compression-codec": "zstd"}, schemas=[{"type": "struct", "schema-id": 0, "fields": [{"id": 1,
  "name": "evt_ts_qx", "required": false, "type": "timestamp"}, {"id": 2, "name": "evt_id_qx", "required":
  false, "type": "long"}]}], partition-specs=[{"spec-id": 0, "fields": [{"name": "evt_ts_day_qx", "transform":
  "day", "source-id": 1, "field-id": 1000}]}], sort-orders=[{"order-id": 1, "fields": [{"transform":
  "identity", "source-id": 2, "direction": "asc", "null-order": "nulls-first"}]}], default-sort-order-id=1.
  The raw bodies of GetTable, GetTableMetadataLocation and ListTables contain 0 of 19 probed tokens (field
  names, partition name, property marker/value, transform/direction/null-order words, 'schema', 'properties').
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: CreateTable, GetTableMetadataLocation,
    GetTable, ListTables · fields: metadata.iceberg.properties, metadata.iceberg.schema,
    metadata.iceberg.partitionSpec, metadata.iceberg.writeOrder, metadataLocation
  - repro: CreateTable(schema+partitionSpec+writeOrder+properties) -> raw
    GetTable/GetTableMetadataLocation/ListTables bodies -> s3 GetObject metadataLocation
  - measurements: api_body_token_hits=0, tokens_probed=19, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-018](#s3t-table-018) · evidence: table/creative/reverify-set-a
  - notes: Re-verification of [S3T-TABLE-018](#s3t-table-018) (confirmed: zero API-body hits, no sibling-name false positive
    this time). Precision gain: the only way to read metadata.iceberg back is the metadata.json in S3, and
    that document is not byte-equal to the request (server adds Iceberg defaults such as the compression...
  - full notes: [details/S3T-TABLE-322.md](details/S3T-TABLE-322.md)

## Response fidelity and consistency

- <a id="s3t-table-015"></a>**S3T-TABLE-015** `response-fidelity` · impact high · handled · verified 2026-10-13
  **CreateTable with schema: metadataLocation already present on the first GetTable (0.42 s after create); versionToken equals create response**
  CreateTable(format=ICEBERG, metadata.iceberg.schema with 3 fields + properties) returned HTTP 200 in 375 ms
  with tableARN+versionToken. GetTable polled every 250 ms: metadataLocation was present on the very first
  poll (poll values [(0.416, '<loc>')]). Its shape is
  '<warehouseLocation>/metadata/00000-<uuid>.metadata.json' (under the warehouseLocation). versionToken from
  CreateTable == GetTable.versionToken: True; == GetTableMetadataLocation.versionToken: True;
  GetTableMetadataLocation returns the same metadataLocation/warehouseLocation as GetTable: True. The location
  did not change over the following 2 s; modifiedAt == createdAt: True.
  - ACK: hook:sdk_create_post_set_output, requeue · ops: CreateTable, GetTable, GetTableMetadataLocation ·
    fields: metadata, metadataLocation, versionToken, warehouseLocation
  - repro: CreateTable(format=ICEBERG, metadata.iceberg.schema.fields=[int,string required,long]) -> GetTable
    every 250 ms until metadataLocation present -> GetTableMetadataLocation
  - measurements: create_latency_ms=[375], first_get_table_with_metadataLocation_s=[0.416],
    polls_before_present=[0], n=1
  - handling: handled via `generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-292](table-metadata-location.md#s3t-table-292), [S3T-TABLE-283](#s3t-table-283), [S3T-TABLE-291](table.md#s3t-table-291), [S3T-TABLE-088](table.md#s3t-table-088), [S3T-TABLE-066](table-metadata-location.md#s3t-table-066), [S3T-TABLE-034](table-metadata-location.md#s3t-table-034),
    [S3T-TABLE-019](table-metadata-location.md#s3t-table-019), [S3T-TABLE-123](table-metadata-location.md#s3t-table-123), [S3T-TABLE-311](#s3t-table-311), [S3T-TABLE-316](table-metadata-location.md#s3t-table-316), [S3T-TABLE-101](table-metadata-location.md#s3t-table-101), [S3T-TABLE-102](table-metadata-location.md#s3t-table-102), [S3T-TABLE-104](table-metadata-location.md#s3t-table-104) ·
    hypotheses: H-T-102, H-T-104 · evidence: table/round-trip/metadata-create-readback
  - notes: H-T-102 confirmed. H-T-104 (asynchronous first commit) refuted: the first GetTable after
    CreateTable returned already carried metadataLocation (n=1 table; polling at 250 ms).

- <a id="s3t-table-033"></a>**S3T-TABLE-033** `response-fidelity` · impact medium · handled · verified 2026-10-13
  **CreateTable with iceberg.schema: GetTable metadataLocation = <warehouseLocation>/metadata/00000-<uuid>.metadata.json; token == Create's**
  CreateTable(format=ICEBERG, metadata={iceberg:{schema:{fields:[id int required, v string]}}}) -> HTTP 200;
  immediate GetTable -> HTTP 200 with metadataLocation present(str), starting with warehouseLocation + '/':
  True, relative path pattern 'metadata/00000-<uuid>.metadata.json' (matches 00000-<uuid>.metadata.json:
  True); create versionToken == GetTable versionToken: True; GetTableMetadataLocation agrees on
  metadataLocation: True and token: True; type=customer, managedByService="<absent>".
  - ACK: is_read_only, late_initialize · ops: CreateTable, GetTable, GetTableMetadataLocation · fields:
    metadata, metadataLocation, versionToken, warehouseLocation
  - repro: CreateTable with metadata.iceberg.schema; GetTable(tableArn); GetTableMetadataLocation
  - handling: handled via `generator.yaml:76-83; generator.yaml:105-110`
  - related: [S3T-TABLE-014](table-metadata-location.md#s3t-table-014), [S3T-TABLE-031](table-metadata-location.md#s3t-table-031), [S3T-TABLE-309](table-maintenance.md#s3t-table-309), [S3T-TABLE-315](table-metadata-location.md#s3t-table-315) · hypotheses: H-T-030 · evidence:
    table/round-trip/fresh-table-fields
  - notes: H-T-030 confirmed: schema-at-create yields a committed initial metadata file in one synchronous
    call; no follow-up UpdateTableMetadataLocation is needed.
  - full notes: [details/S3T-TABLE-033.md](details/S3T-TABLE-033.md)

## Handling gaps (bugs to file)

- [S3T-TABLE-133](#s3t-table-133) - CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created;
  partitionSpec={} and writeOrder={orderId:1} -> 200 (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven. A Table CR whose
  metadata.iceberg.schema has no fields never reaches the terminal BadRequestException the controller expects:
  schema={} on the wire -> InternalServerErrorException 500 reproducibly (4/4, no table created;
  [S3T-TABLE-133](#s3t-table-133)), so the reconciler requeues forever, and schema.fields=[] -> HTTP 200 creating an
  empty-schema table reported as Synced ([S3T-TABLE-036](#s3t-table-036)); add CRD validation (minItems=1 on schema.fields, or
  require exactly one of schema/schemaV2) or a pre-build guard.
  - handling_ref: `test/e2e/tests/test_table.py:175-193; test/e2e/resources/table.yaml`
- [S3T-TABLE-312](#s3t-table-312) - Schema 'required' round-trips into metadata.json: required=true -> True, required=false ->
  False, omitted -> False (partially handled)
  FU-08. required=true is preserved as a non-null column (true), required=false and an omitted required both
  become False in the metadata; the SDK-side pointer/omission handling itself is not observable from the API
  (only its effect in metadata.json is). The schema is never returned by any s3tables operation (GetTable
  keys: ['createdAt', 'createdBy', 'format', 'metadataLocation', 'modifiedAt', 'name', 'namespace',
  'namespaceId', 'ownerAccountId', 'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']),
  so the read back needs s3:GetObject on the warehouse bucket (allowed for the account's caller here).
  - handling_ref: `5c1bda2; 007c152`
- [S3T-TABLE-313](#s3t-table-313) - Table 'type' is 'customer' in GetTable and ListTables (agree); schema field types live only
  in metadata.json, in no API output (partially handled)
  FU-10. The table type member is literally named 'type' with the value shown (TableType enum customer|aws); a
  spec.metadata.iceberg.schema can never be diffed against the API, only against metadata.json. Creative extra
  (second table): explicit SchemaField.id values 5/7/9 were IGNORED and renumbered 1/2/3 in metadata.json, and
  the type string 'decimal(10,2)' was normalized to 'decimal(10, 2)' (space inserted); 'timestamp' and
  primitives were kept verbatim. ListTables summaries carried only createdAt, modifiedAt, name, namespace,
  tableARN, type on the wire (no namespaceId/tableBucketId/managedByService although the model has them).
  - handling_ref: `generator.yaml:45-46; generator.yaml:60-65`
- [S3T-TABLE-314](#s3t-table-314) - CreateTable ignores explicit schema field ids (renumbered 1..n) and normalizes
  'decimal(10,2)' to 'decimal(10, 2)' in metadata.json (partially handled)
  FU-08/FU-10 creative extra, single table. Since no API output returns the schema, this never produces a
  delta; it matters only for documentation (a user-set id is not honored) and for anyone comparing spec
  against metadata.json.
  - handling_ref: `generator.yaml:45-46; generator.yaml:60-65`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-015](#s3t-table-015) | CreateTable with schema: metadataLocation already present on the first GetTable (0.42 s after create); versionToken equals create response | create_latency_ms=[375], first_get_table_with_metadataLocation_s=[0.416], polls_before_present=[0] | 1 |
| [S3T-TABLE-133](#s3t-table-133) | CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created; partitionSpec={} and writeOrder={orderId:1} -> 200 | attempts=3, http_500=3 | 3 |
| [S3T-TABLE-284](#s3t-table-284) | Doc claim C092 FALSE: CreateTable with partitionSpec present but fields omitted/empty -> BadRequestException (all accepted) | variants=3 | 3 |
| [S3T-TABLE-311](#s3t-table-311) | CreateTable metadata={} -> client-side ParamValidationError (validation on) / BadRequestException 400 'metadata is not valid' (off) | n_empty_validation_off=2, n_per_other_cell=1 | 4 |
| [S3T-TABLE-322](#s3t-table-322) | CreateTable metadata.iceberg.properties: metadata.json keeps the user property and adds write.parquet.compression-codec=zstd | api_body_token_hits=0, tokens_probed=19 | 1 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-schema.md` - unverified, n=1 or single-run items worth a future probe; one line each.

- [S3T-TABLE-015](#s3t-table-015) / -033 / -283: `metadataLocation` present on the first GetTable poll (0.42 s) is n=1 per finding; run 10+
  schema-at-create tables to get a range and confirm it is never absent on the first read.
- [S3T-TABLE-133](#s3t-table-133) / -036: the `schema={}` -> 500 is n=3 (+1) from one day with botocore validation bypassed; re-check after a
  service change, and test whether the Go SDK v2 tagged-union encoding can emit `schema={}` or `metadata={}` at all
  ([S3T-TABLE-311](#s3t-table-311) saw only the client-side ParamValidationError with validation on).
- [S3T-TABLE-312](#s3t-table-312) / -313 / -314: one table each (FU-08/FU-10); the v1 `required` round-trip and the id renumbering were not
  repeated on schemaV2 beyond the single attempts of -043/-044/-046.
- [S3T-TABLE-046](#s3t-table-046): in the one accepted nested schemaV2 the struct child got id 3 and the sibling list id 2 (sent 2 and 3) -
  the renumbering order for nested fields and deeper nesting is unprobed.
- [S3T-TABLE-073](#s3t-table-073) / -076 / -038: 'days', 'bucket' (no arity) and '' transforms are written verbatim and 'INT'/'String'
  lower-cased (one attempt each); whether Iceberg readers or the maintenance engine later reject such metadata is open.
- [S3T-TABLE-037](#s3t-table-037): 'a' and 'A' accepted as distinct fields (one attempt); how engines treat case-only distinct names is open.
- [S3T-TABLE-322](#s3t-table-322) / -017 / -194: `write.parquet.compression-codec=zstd` is injected (n=1); whether a user-supplied value for
  that key, or the accepted `format-version`/`owner` keys of -194, survive into metadata.json is untested.
- [S3T-TABLE-284](#s3t-table-284) / -285 / -077: empty partitionSpec/writeOrder shapes commit as spec-id 0 / order-id 0 (n=3 and 4 shapes,
  one day); whether the controller's generated CreateTable input can emit these shapes (nil vs empty struct) is unprobed.
- [S3T-TABLE-041](#s3t-table-041): schemaV2-alone acceptance is n=1; the Go SDK encoding of a Table CR carrying both `schema` and `schemaV2`
  (which the API rejects with its own 400 message) has not been exercised through a reconcile.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-017"></a>**S3T-TABLE-017** | normalization | low | confirmed | metadata.iceberg.properties written to metadata.json verbatim (key case/whitespace kept) plus injected write.parquet.compression-codec=zstd | [S3T-TABLE-194](#s3t-table-194), [S3T-TABLE-018](#s3t-table-018) | - |
| <a id="s3t-table-040"></a>**S3T-TABLE-040** | normalization | low | confirmed | v1 SchemaField.id (ids 5,9 sent) -> accepted; written ids [1, 2]; duplicate ids 5,5 -> BadRequestException 400 | [S3T-TABLE-037](#s3t-table-037), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-314](#s3t-table-314), [S3T-TABLE-313](#s3t-table-313) | - |
| <a id="s3t-table-042"></a>**S3T-TABLE-042** | request-validation | low | confirmed | schemaV2.type: only 'struct' accepted; omitted, 'STRUCT' or 'record' -> BadRequestException 400 (no default, case-sensitive) | [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-043](#s3t-table-043), [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-045](#s3t-table-045), [S3T-TABLE-046](#s3t-table-046), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-073](#s3t-table-073), [S3T-TABLE-076](#s3t-table-076) | - |
| <a id="s3t-table-045"></a>**S3T-TABLE-045** | normalization | low | confirmed | schemaV2.schemaId is ignored: 7 and -1 both accepted and written as schema-id 0 / current-schema-id 0 | [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-042](#s3t-table-042), [S3T-TABLE-043](#s3t-table-043), [S3T-TABLE-044](#s3t-table-044), [S3T-TABLE-046](#s3t-table-046), [S3T-TABLE-039](#s3t-table-039) | - |
| <a id="s3t-table-194"></a>**S3T-TABLE-194** | request-validation | low | confirmed | CreateTable accepts any metadata.iceberg.properties: empty key, empty value, format-version/owner keys, 200 entries, 5 KiB value, {} -> 200 | [S3T-TABLE-036](#s3t-table-036), [S3T-TABLE-133](#s3t-table-133), [S3T-TABLE-311](#s3t-table-311), [S3T-TABLE-041](#s3t-table-041), [S3T-TABLE-017](#s3t-table-017), [S3T-TABLE-018](#s3t-table-018) | - |
| <a id="s3t-table-283"></a>**S3T-TABLE-283** | other | low | confirmed | Doc claim C091 TRUE: No metadata: GetTableMetadataLocation has warehouseLocation+versionToken, no metadataLocation; with schema it is set | [S3T-TABLE-014](table-metadata-location.md#s3t-table-014), [S3T-TABLE-033](#s3t-table-033), [S3T-TABLE-015](#s3t-table-015), [S3T-TABLE-292](table-metadata-location.md#s3t-table-292), [S3T-TABLE-291](table.md#s3t-table-291), [S3T-TABLE-088](table.md#s3t-table-088), [S3T-TABLE-066](table-metadata-location.md#s3t-table-066) | - |
| <a id="s3t-table-286"></a>**S3T-TABLE-286** | other | low | confirmed | Doc claim C094 TRUE: partitionSpec with fields and no specId -> metadata default-spec-id 0 and partition-specs[0].spec-id 0 | [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-072](#s3t-table-072), [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289), [S3T-TABLE-075](#s3t-table-075) | - |
| <a id="s3t-table-287"></a>**S3T-TABLE-287** | other | low | confirmed | Doc claim C095 TRUE: writeOrder with fields and no orderId -> metadata default-sort-order-id 1 and sort-orders[].order-id 1 | [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-288](#s3t-table-288), [S3T-TABLE-289](#s3t-table-289), [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-290](#s3t-table-290), [S3T-TABLE-072](#s3t-table-072) | - |
| <a id="s3t-table-288"></a>**S3T-TABLE-288** | other | low | confirmed | Doc claim C096 TRUE: A schema field that omits `required` is written to metadata.json with required=false | [S3T-TABLE-016](#s3t-table-016), [S3T-TABLE-284](#s3t-table-284), [S3T-TABLE-285](#s3t-table-285), [S3T-TABLE-286](#s3t-table-286), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-289](#s3t-table-289), [S3T-TABLE-074](#s3t-table-074), [S3T-TABLE-075](#s3t-table-075) | - |
| <a id="s3t-table-290"></a>**S3T-TABLE-290** | other | low | confirmed | Doc claim C098 TRUE: partitionSpec sourceId that is not a schema field id -> BadRequestException 400; no dangling partition field | [S3T-TABLE-072](#s3t-table-072), [S3T-TABLE-075](#s3t-table-075), [S3T-TABLE-287](#s3t-table-287), [S3T-TABLE-289](#s3t-table-289) | - |
| <a id="s3t-table-310"></a>**S3T-TABLE-310** | read-gap | high | confirmed | CreateTable metadata (iceberg schema) is write-only: GetTable/ListTables bodies have no metadata, schema or iceberg member | [S3T-TABLE-018](#s3t-table-018), [S3T-TABLE-313](#s3t-table-313), [S3T-TABLE-090](table-metadata-location.md#s3t-table-090), [S3T-TABLE-023](table-metadata-location.md#s3t-table-023), [S3T-TABLE-113](table.md#s3t-table-113), [S3T-TABLE-028](table-metadata-location.md#s3t-table-028) | [S3T-TABLE-018](#s3t-table-018) |
| <a id="s3t-table-314"></a>**S3T-TABLE-314** | normalization | low | confirmed | CreateTable ignores explicit schema field ids (renumbered 1..n) and normalizes 'decimal(10,2)' to 'decimal(10, 2)' in metadata.json | [S3T-TABLE-037](#s3t-table-037), [S3T-TABLE-038](#s3t-table-038), [S3T-TABLE-039](#s3t-table-039), [S3T-TABLE-040](#s3t-table-040), [S3T-TABLE-313](#s3t-table-313) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
