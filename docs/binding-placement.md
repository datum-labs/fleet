# Binding placement and exclusive reservation

## Decision

Keep Fleet's protected binding record in the same consumer Project control
plane as its `MachineClaim`. The binding is service-managed and is not a
customer-writable public resource. The Machine stays in its provider Project;
ownership does not move with the claim.

This changes the placement suggested by the draft Fleet proposal, which keeps
cleanup authority outside a terminating consumer Project. The project deletion
protocol below is therefore required before enabling allocation.

## Binding and reservation protocol

1. Create or recover a pending binding record in the consumer Project, tied to
   exactly one namespaced `MachineClaim`. The record contains the admitted
   contract, binding ID and epoch, and cleanup obligations. Admission and IAM
   permit only Fleet's service identity to write it.
2. Obtain Milo quota and recheck current source-use permission, entitlement,
   and physical eligibility.
3. Reserve the provider `Machine` using one conditional update against its
   current resource version. The Machine records an opaque reference to the
   consumer Project and binding. Competing consumers must conflict; the
   provider-side reservation is the exclusive-use gate. It is not a second
   binding or an authorization to perform physical effects.
4. Mark the consumer binding active only after the quota grant and provider
   reservation agree. A controller crash at any point is recovered from both
   records. If either side is unavailable or disagrees, permit no new physical
   effects and retain or quarantine the Machine.

The exact CRD scope within the consumer Project, admission rules, reference
shape, and conditional-update implementation remain to be specified. No
cross-project Kubernetes owner reference is used.

## Release and Project deletion

Release stops new effects, revokes access, completes cleanup, removes the Milo
quota claim, and only then clears the provider Machine reservation. The binding
is deleted last. Finalizers on the claim and binding keep them available while
release is incomplete.

Consumer Project finalization must wait for its active bindings to release.
If the Project control plane or binding is lost despite that guard, the provider
Machine remains reserved or quarantined. Fleet must not infer that the hardware
is free from the missing consumer record. An explicit recovery or operator
decision is needed before reuse; no automatic force-deletion path is defined.

## Required proof before allocation

- Two claims in different Projects race for one Machine; exactly one wins the
  provider reservation.
- Crash after binding creation, quota grant, Machine reservation, and active
  publication; reconciliation converges without double allocation or leaked
  quota.
- Claim deletion and consumer Project deletion wait for cleanup. Loss of the
  consumer API leaves the Machine unavailable for reuse.
- A stale worker cannot perform effects after binding epoch changes or release.

## Sources

- [Fleet cross-project allocation and Project deletion proposal](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/capacity-publication.md)
- [Kubernetes API optimistic concurrency](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Kubernetes finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
- [Kubernetes owner-reference scope](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
