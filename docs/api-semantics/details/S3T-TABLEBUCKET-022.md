<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-022: CreateTableBucket enforces the documented naming page exactly: all reserved affixes (incl. any 'aws' prefix) -> BadRequestException 400
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-022"></a>**S3T-TABLEBUCKET-022** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTableBucket enforces the documented naming page exactly: all reserved affixes (incl. any 'aws' prefix) -> BadRequestException 400**
  Client validation off. Accepted (HTTP 200, then deleted): 3-char name, 63-char name, consecutive hyphens
  ('a--b'). Rejected with BadRequestException HTTP 400 "The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.": 2 chars, 64 chars, uppercase, underscore, period, leading
  hyphen, trailing hyphen, IP-like 192.168.0.1, prefixes xn--, sthree-, amzn-s3-demo-, aws-, aws<name> (no
  hyphen) and awsome-, suffixes -s3alias, --ol-s3, --x-s3, --table-s3. GetTableBucket on an ARN built from
  each rejected name: BadRequestException 400 (same name message) for 2/64 chars, uppercase, period,
  leading/trailing hyphen, IP-like, xn--/sthree-/amzn-s3-demo- prefixes and all four suffixes;
  NotFoundException 404 'The specified bucket does not exist.' for the underscore name and for
  aws-/aws<name>/awsome- (GetTableBucket does not apply the aws-prefix rule and tolerates '_').
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, GetTableBucket · fields: name
  - repro: CreateTableBucket(name=<candidate>) for each candidate with parameter_validation=False; Delete any
    200; Get constructed ARN for each 400
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](../tablebucket.md#s3t-tablebucket-033), [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003),
    [S3T-SERVICE-011](../service.md#s3t-service-011) · hypotheses: H-B-030, H-B-031 · evidence:
    tablebucket/weird-inputs/create-names-enums-empty-structs

## Notes

Confirms H-B-030 for every cell (the model pattern [0-9a-z-]{3,63} is looser than the live rule). Refutes
H-B-031: the 'aws' prefix rule is an exact string prefix ('awsome-...' and 'aws<name>' are rejected like
'aws-...'), and xn--/-s3alias/--table-s3 are enforced, so the live service matches the naming page. One
generic message for all name failures: the controller cannot tell the user which rule failed. GetTableBucket
applies only the ARN pattern, not the naming rules (404 for underscore/aws-prefixed names).
