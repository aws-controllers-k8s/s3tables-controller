<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-029: Tag keys case-sensitive (Env/env coexist); empty/unicode values round-trip; aws:/AWS: key -> BadRequestException; '' key -> ValidationExc
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

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
  - related: [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025), [S3T-TABLEBUCKET-030](../tablebucket.md#s3t-tablebucket-030), [S3T-TABLEBUCKET-031](../tablebucket.md#s3t-tablebucket-031) · hypotheses: H-B-044 · evidence:
    tablebucket/tags/upsert-limits-chars-scope

## Notes

Confirms H-B-044. 'AWS:x' (uppercase) rejected; the AWS-wide reserved prefix is matched case-insensitively.

Full original title: Tag keys are case-sensitive (Env/env coexist), empty values and unicode round-trip
byte-identical; aws:x -> BadRequestException 400, AWS:x -> BadRequestException 400, empty key ->
ValidationException 400
