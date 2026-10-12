# API serving approach for the first Fleet resources

## Recommendation

Use standard Kubernetes CRDs and controllers for the first public Fleet APIs:
`MachineClass`, `Machine`, and `MachineClaim`. Install the project-facing CRDs in
the Milo control planes where their owners create and read them. This follows
Milo's existing use of CRDs, including quota `ResourceClaim`.

The draft Fleet proposal requires Kubernetes-style objects, project isolation,
and reconciliation. It does not require an aggregated API server for these
resources. CRDs provide schema validation, status, watches, finalizers, and
admission webhooks. Fleet can add custom admission and controllers without
owning the public API server's storage or request handling.

An aggregated API may still be useful later for operations that need custom
request semantics or storage. If introduced, assign it a separate API group or
version; an `APIService` claims its whole group/version path and cannot serve a
subset of the CRDs in that same path.

## Ownership and coordination

- Consumer project control planes hold local `MachineClass` references and
  `MachineClaim` intent. Provider project control planes hold provider class
  definitions and `Machine` inventory.
- A Fleet controller resolves cross-project references and checks current
  permission and Milo quota before binding. It never treats a copied class or
  object name as proof of authorization.
- Fleet needs service-owned, durable allocation and operation records outside
  the consumer project's deletion domain. A protected CRD in an independent
  control plane may be sufficient for the initial unique Machine reservation if
  its create/update conflict behavior, retention, recovery, and effect fencing
  are proven. A transactional store remains an option if those guarantees cannot
  be met with CRDs. Public claim status alone is insufficient.
- Milo quota `ResourceClaim` remains Milo-owned. Fleet resolves and verifies it
  during reconciliation without storing its UID in the binding record.

## Repository impact

The current aggregated API server is a non-serving scaffold from the repository
template, not an architectural requirement. Do not register its `APIService` for
the first CRDs. Replace the deployment path with CRD manifests, admission, and
controllers as the resource schemas become accepted. Remove the unused server
scaffold at that point.

## Sources

- [Fleet service architecture](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/architecture.md)
- [Fleet allocation proposal](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/bare-metal-hardware.md)
- [Milo quota ResourceClaim CRD](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/config/crd/bases/quota/quota.miloapis.com_resourceclaims.yaml)
- [Kubernetes CRDs and aggregated APIs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Kubernetes aggregation path ownership](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)
