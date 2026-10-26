<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-011: CreateNamespace list: 2 elements, [x,x] or a string -> 400 'Multipart namespaces are not supported'; [], [''], [None] -> 400 invalid name
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-NAMESPACE-004](../namespace.md#s3t-namespace-004), [S3T-NAMESPACE-012](../namespace.md#s3t-namespace-012), [S3T-NAMESPACE-013](../namespace.md#s3t-namespace-013), [S3T-NAMESPACE-014](../namespace.md#s3t-namespace-014), [S3T-NAMESPACE-015](../namespace.md#s3t-namespace-015),
    [S3T-NAMESPACE-016](../namespace.md#s3t-namespace-016) · hypotheses: H-N-005, H-N-006, H-N-007, H-N-008 · evidence:
    namespace/weird-inputs/naming-edge-cases

## Notes

H-N-005 confirmed / H-N-006 refuted (two-element list rejected, nothing created). H-N-007 qualified: the model
has min=1,max=1 on the list, so botocore blocks [] client-side (ParamValidationError) while the server answers
BadRequestException. H-N-008 confirmed ([x,x] -> BadRequestException; [''] -> BadRequestException).

Full original title: CreateNamespace namespace list: two elements -> BadRequestException; [] ->
BadRequestException (server) / ParamValidationError (botocore); [''] -> BadRequestException
