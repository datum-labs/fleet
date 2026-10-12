# Milo quota integration for machine allocation

## Decision

Fleet uses Milo's `quota.miloapis.com` system as the quota authority. Fleet does
not maintain a separate quota counter, grant ledger, or policy engine. The
binding record does not duplicate the Milo `ResourceClaim` UID.

The initial quota resource should be a count of allocated machines, consumed by
the customer Project: one granted unit per bound `MachineClaim`. Register a
resource type such as `fleet.datumapis.com/machines` with `MachineClaim` as an
allowed claiming resource. Class-specific or CPU/memory quota can be added only
when there is a product requirement and a migration plan for existing grants.

## Recommended binding sequence

1. Validate the consumer's authenticated Project, class provenance and current
   source-use permission, location, network feasibility, and candidate Machine.
2. Create or recover a deterministically named Milo quota `ResourceClaim` in
   the consumer Project's quota control plane. Set `resourceRef` to the Fleet
   `MachineClaim`. Derive the name from the Fleet claim's immutable identity so
   a reused object name cannot inherit an earlier quota claim. Only the Fleet
   service identity may manage this Milo claim.
3. Resolve that Milo claim by its derived name and wait for `Granted=True` for
   its current generation. Verify its `resourceRef` and requested resources on
   every reconciliation. `Pending` or an unavailable quota service leaves the
   Fleet claim pending; `QuotaExceeded` is reported distinctly from
   `InsufficientCapacity`. No Machine is activated before a grant.
4. Revalidate candidate, policy, and grant, then commit Fleet's unique Machine
   and Claim reservations. If a competing allocation wins, release the unused
   quota claim. A crash between Milo and Fleet writes is recovered by
   reconciling the two records; they are not one database transaction.
5. Reconcile both records while the allocation is active. Missing or changed
   quota authority must block new effects and trigger an explicit recovery or
   release path, never silently permit a new consumer to use the Machine.

This sequence claims quota at allocation time, not when a customer first creates
a potentially long-pending `MachineClaim`. Milo's `ClaimCreationPolicy` can
create quota claims at API admission, but using it for the first Fleet release
would charge requests that have not received a Machine. Fleet should instead
create the Milo `ResourceClaim` as part of binding and treat its status as the
gate. The exact claim name, quota namespace, IAM role, and resource type must be
validated against Milo before serving the API.

## Release and recovery

The Fleet binder retains its Machine reservation while it revokes access,
detaches or fences networking, and completes the owner's cleanup policy. Once
the Machine is safe to reuse, Fleet deletes the Milo quota `ResourceClaim` and
records its deletion before it marks the binding released. If Milo is
unavailable, release stays retryable; Fleet derives the quota claim name from
its retained claim identity and consumer Project identity. Project deletion
needs a service-owned cleanup route and a retained project identity tombstone
so quota reconciliation does not depend on a surviving consumer `MachineClaim`
or owner reference.

This order deliberately avoids freeing physical capacity while an earlier
consumer can still access it. Milo quota usage and Fleet's physical reservation
may temporarily diverge during failure or cleanup; the reconciler must detect,
repair, and alert on that divergence. The Fleet binding store remains the
authority for exclusive Machine ownership, while Milo remains the quota
authority.

## Validation before allocation is enabled

- Register the resource type, Project consumer, and `MachineClaim` claiming kind;
  verify grants are provisioned in two consumer projects.
- Race claims from two projects for one eligible Machine. At most one Fleet
  binding succeeds, and the losing Milo quota claim is removed.
- Fail between Milo grant and Fleet commit, and between Fleet commit and status
  publication. Restart must recover without double binding or leaking quota.
- Deny quota, revoke source-use permission, delete a consumer project, and make
  Milo temporarily unavailable. No ungranted or stale claim may activate a
  Machine; cleanup must remain visible and retryable.

## Sources

- [Milo quota architecture](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/docs/architecture/quota.md)
- [Milo ResourceClaim API](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/pkg/apis/quota/v1alpha1/resourceclaim_types.go)
- [Milo ClaimCreationPolicy API](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/pkg/apis/quota/v1alpha1/claimcreationpolicy_types.go)
- [Compute's project quota integration](https://github.com/datum-cloud/compute/blob/d70ed7baf0ee81cb5ec46a272770c64311af5048/docs/enhancements/quota-enforcement/README.md)
- [Fleet allocation proposal](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/bare-metal-hardware.md)
