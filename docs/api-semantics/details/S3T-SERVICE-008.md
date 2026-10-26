<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-008: CreateNamespace/DeleteNamespace 429 budget is shared across buckets (~16 per burst, account/principal-wide): A 12/12 OK, B 429 from 5th call
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-008"></a>**S3T-SERVICE-008** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateNamespace/DeleteNamespace 429 budget is shared across buckets (~16 per burst, account/principal-wide): A 12/12 OK, B 429 from 5th call**
  12 CreateNamespace into bucket A then 12 into bucket B back-to-back (wall 0.6 s): A codes ['200', '200',
  '200', '200', '200', '200', '200', '200', '200', '200', '200', '200'] (first 429 at index None); B codes
  ['200', '200', '200', '200', 'TooManyRequestsException', 'TooManyRequestsException',
  'TooManyRequestsException', 'TooManyRequestsException', '200', 'TooManyRequestsException',
  'TooManyRequestsException', 'TooManyRequestsException'] (first 429 at index 4). ListNamespaces afterwards: A
  has 12, B has 5; names whose create got 429 but exist: []; names whose create got 200 but are missing: [].
  Delete burst (wall 0.4 s): A ['204', '204', '204', '204', '204', '204', '204', '204', '204', '204', '204',
  '204'], B ['204', '204', '204', '204', 'TooManyRequestsException']; namespaces whose delete got 429 but are
  gone: [].
  - ACK: requeue, one-per-reconcile · ops: CreateNamespace, DeleteNamespace
  - repro: Two fresh table buckets; 12 CreateNamespace into A then 12 into B without pause; ListNamespaces
    both; repeat with DeleteNamespace
  - measurements: creates_ok_A=12, creates_ok_B=5, first_429_index_A=null, first_429_index_B=4, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-004](../service.md#s3t-service-004), [S3T-SERVICE-018](../service.md#s3t-service-018), [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018) · evidence: service/creative/throttle-scope

## Notes

Creative round, seed error-messages: fixes what the TooManyRequestsException RETRY row is keyed on: NOT the
bucket - bucket B was throttled once the combined count reached 16 (A 12 + B 4), and the delete burst behaved
the same (A 12 + B 4 then 429). A 429'd create/delete had no side effect (0 ghost namespaces, 0 deletes
applied), so retrying as-is is safe. A token-bucket refill is visible (B #8 -> 200 at 0.60 s between 429s).
Other explorers share the account, so the shared budget may have been partly consumed by them; n=1 burst each
for create and delete.

Full original title: CreateNamespace/DeleteNamespace 429 budget is shared across buckets (~16 per burst,
account/principal-wide): A 12/12 ok, B 429 from its 5th call
