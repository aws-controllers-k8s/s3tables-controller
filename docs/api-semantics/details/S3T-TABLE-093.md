<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-093: ListTables.prefix: literal, case-sensitive table-name prefix across namespaces, ANDed with namespace; '*' matches none; '' -> 200 unfiltered
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-TABLE-094](../table.md#s3t-table-094), [S3T-TABLE-097](../table.md#s3t-table-097) · hypotheses: H-T-072 · evidence:
    table/weird-inputs/list-tables-bounds

## Notes

H-T-072 qualified: all semantics clauses confirmed (literal, case-sensitive, name only - a namespace name or
'ns.name' as prefix matches nothing - across namespaces, ANDed with the namespace filter, '*' is literal and
matches nothing), but the empty-prefix clause is refuted: the SDK rejects '' client-side (min length 1) while
the server answers HTTP 200 with the unfiltered list; a prefix over 255 chars -> BadRequestException 400 'The
specified search prefix is not valid. A search prefix can not exceed 255 characters.'

Full original title: ListTables.prefix: literal, case-sensitive, table-name only, across namespaces, ANDed
with namespace; '*' literal; '' -> 200 unfiltered (not 400)
