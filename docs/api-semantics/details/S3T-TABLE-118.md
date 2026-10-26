<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-118: Tag/UntagResource with an 'aws:'-prefixed key (any case) -> BadRequestException 400; mixed untag ['aws:x','env'] rejected whole, env kept
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-118"></a>**S3T-TABLE-118** `tag-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Tag/UntagResource with an 'aws:'-prefixed key (any case) -> BadRequestException 400; mixed untag ['aws:x','env'] rejected whole, env kept**
  TagResource({'aws:foo':x}) -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for
  system use."; {'AWS:foo':x} -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for
  system use."; {'awsfoo':x} (no colon) -> HTTP 200. ListTagsForResource shows keys starting with aws: (any
  case): []. UntagResource(['aws:cloudformation:stack-name']) -> BadRequestException (HTTP 400) "Tags
  beginning with 'aws:' are reserved for system use.". On a table with {env:a}: UntagResource(['aws:x','env'])
  -> BadRequestException (HTTP 400) "Tags beginning with 'aws:' are reserved for system use.", map afterwards
  {'team': 'b', 'k': 'v2', 'env': 'a', ' k ': 'v', 'e': '', 'w': ' spaced value '}.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, UntagResource, ListTagsForResource · fields:
    tags, tagKeys
  - repro: TagResource({'aws:foo':'x'}); TagResource({'AWS:foo':'x'});
    UntagResource(['aws:cloudformation:stack-name']); UntagResource(['aws:x','env'])
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-152](../tablebucket.md#s3t-tablebucket-152), [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025), [S3T-TABLE-119](../table.md#s3t-table-119), [S3T-TABLE-117](../table.md#s3t-table-117), [S3T-TABLE-126](../table.md#s3t-table-126),
    [S3T-TABLE-185](../table.md#s3t-table-185), [S3T-TABLE-125](../table.md#s3t-table-125) · hypotheses: H-T-091, H-T-206 · evidence: table/tags/table-tag-semantics

## Notes

H-T-091 confirmed (predicted BadRequestException 400 for aws: keys, no system tags returned). H-T-206
(contrarian: untag of aws: keys is a permissive 204) refuted: BadRequestException (HTTP 400) "Tags beginning
with 'aws:' are reserved for system use.". Messages: tag "Tags beginning with 'aws:' are reserved for system
use."; untag "Tags beginning with 'aws:' are reserved for system use.".

Full original title: TagResource with an 'aws:' key -> BadRequestException; UntagResource(['aws:...']) ->
BadRequestException; mixed ['aws:x','env'] -> BadRequestException (env removed: False)
