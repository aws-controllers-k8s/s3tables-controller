<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-026: CreateTable or RenameTable into a missing namespace -> NotFoundException, no implicit namespace is created; the table stays put
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-026"></a>**S3T-NAMESPACE-026** `prerequisite` · impact high · handled · verified 2026-10-13
  **CreateTable or RenameTable into a missing namespace -> NotFoundException, no implicit namespace is created; the table stays put**
  CreateTable(namespace=<missing>, format=ICEBERG) -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.'; ListNamespaces(prefix) afterwards does not contain the namespace. RenameTable(ns1/t1,
  newNamespaceName=<missing>) -> NotFoundException (HTTP 404) 'The specified destination namespace does not
  exist.'; GetTable(ns1/t1) -> HTTP 200 and the namespace still does not exist. RenameTable to an existing
  namespace -> HTTP 204; GetTable at the new place -> HTTP 200 (tableARN unchanged).
  - ACK: references, requeue, exceptions.404 · ops: CreateTable, RenameTable, ListNamespaces, GetTable ·
    fields: namespace, newNamespaceName
  - repro: CreateTable(namespace=<never created>) -> ListNamespaces(prefix) ->
    RenameTable(newNamespaceName=<never created>) -> GetTable -> RenameTable(newNamespaceName=<existing>)
  - handling: handled via `generator.yaml:128-136; pkg/resource/namespace/sdk.go:208-217; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-001](../namespace.md#s3t-namespace-001), [S3T-NAMESPACE-025](../namespace.md#s3t-namespace-025), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029) · hypotheses: H-N-046,
    H-N-047 · evidence: namespace/dependencies/delete-with-tables

## Notes

H-N-046 confirmed; H-N-047 (contrarian: implicit creation) refuted. The CreateTable 404 message is 'The
specified namespace does not exist.', the RenameTable one 'The specified destination namespace does not
exist.'.

Full original title: CreateTable into a missing namespace -> NotFoundException, no implicit namespace;
RenameTable(newNamespaceName=<missing>) -> NotFoundException, table stays put
