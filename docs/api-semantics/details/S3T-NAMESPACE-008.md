<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-008: ListNamespaces prefix: literal, case-sensitive, left-anchored; '' and invalid-char prefixes -> 200 (empty); >255 chars -> BadRequest 400
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-NAMESPACE-009](../namespace.md#s3t-namespace-009), [S3T-NAMESPACE-010](../namespace.md#s3t-namespace-010), [S3T-NAMESPACE-032](../namespace.md#s3t-namespace-032), [S3T-NAMESPACE-005](../namespace.md#s3t-namespace-005) · hypotheses: H-N-033,
    H-N-050 · evidence: namespace/response-fidelity/list-prefix-pagination

## Notes

H-N-033 confirmed (literal prefix returned exactly {ns_a, ns_b}: True; NS_/ns%/unanchored returned empty:
True; prefix='' -> HTTP 200). H-N-050 qualified/refuted (see behavior) (My-Ns -> HTTP 200; a.b -> HTTP 200;
300 chars -> BadRequestException (HTTP 400) 'The specified search prefix is not valid. A search prefix can not
exceed 255 characters.'; 256 chars -> BadRequestException (HTTP 400) 'The specified search prefix is not
valid. A search prefix can not exceed 255 characters.'; e-acute -> HTTP 200). botocore's client-side
validation rejects prefix='' (min length 1) unless disabled.

Full original title: ListNamespaces prefix is a literal, case-sensitive, left-anchored match; prefix='' ->
HTTP 200; invalid-char/300-char prefixes -> BadRequestException/HTTP 200
