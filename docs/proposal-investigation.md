# Fleet management proposal investigation

## Status and scope

[datum-cloud/infra PR #5782](https://github.com/datum-cloud/infra/pull/5782),
“docs: propose Fleet infrastructure management service,” is an open **draft design
proposal**, not an implemented API or an approved production architecture. This
investigation describes the proposal at head commit
[`51781e2456367be82281e1443ca56d0686311133`](https://github.com/datum-cloud/infra/tree/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management).
Names, scopes, storage choices, operation surfaces, providers, and ownership remain
subject to review.

The proposal makes Fleet one service composed of independently authoritative
subsystems: physical inventory, machine allocation, bare-metal provisioning,
operating-system configuration, and cluster orchestration. Standalone consumers
and managed clusters use the same allocator and installation contract. Provider
adapters execute admitted work but do not become a second allocation authority.

Primary proposal sources:

- [Service overview and common contracts](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/README.md)
- [Service architecture and recovery model](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/architecture.md)
- [Physical inventory](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/inventory.md)
- [Machine allocation](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/bare-metal-hardware.md)
- [Capacity publication](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/capacity-publication.md)
- [Bare-metal provisioning](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/bare-metal-provisioning.md)
- [Operating-system configuration](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/os-configuration.md)
- [Cluster lifecycle](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/cluster-lifecycle.md)

## API inventory

The proposal uses Kubernetes-style APIs served through Milo. It proposes CRD-like
resources and reconciling controllers; it does not define a protobuf, gRPC, or
independent REST API. The `/prepare` operation mentioned for owned hardware and
the console/action surfaces are conceptual operation APIs whose exact form still
needs to be designed.

### Physical inventory

Proposed group: `inventory.fleet.datumapis.com/v1alpha1`.

| Kind | Authority and purpose |
| --- | --- |
| `Facility` | Provider-neutral physical premises linked to a geographic Location. |
| `Rack` | Facility membership, mounting capacity, and physical constraints. |
| `DeviceClaim` | Operator-approved expected installation, reservations, verification requirements, and one committed Device binding. |
| `Device` | Registrar-created canonical physical asset and observed typed components. |
| `Cable` | Sole authority for expected physical endpoints and pass-through mappings. |
| `PowerFeed` | Supply connection, capacity, and upstream dependency. |

Typed ports and components initially remain embedded entries with stable IDs.
`DeviceClaim` represents the approved plan; `Device` represents what was actually
registered. A `Machine` for owned hardware references the `Device` UID rather than
copying serial, rack, cable, or component facts.

### Shared capacity and allocation

Proposed group: `fleet.datumapis.com/v1alpha1`.

| Kind | Authority and purpose |
| --- | --- |
| `MachineClass` | Either a provider definition or a consumer-local one-hop `spec.source` reference. The two forms are mutually exclusive and immutable. |
| `Machine` | Provider-owned allocatable capacity, its typed provider reference, eligibility, reclaim policy, and for owned metal its Device reference. |
| `MachineClaim` | Consumer request for class, placement, power, networking, and optional initial installation; receives one immutable private binding. |

Definitions are project-scoped and select an explicit namespace of provider
supply. References expose only sanitized capabilities. The customer never names
a foreign `Machine`, `Device`, provider class, or management secret.

The binder commits unique keys for `(provider Project UID, Machine UID)` and
`(consumer Project UID, Claim UID)`, together with quota, authorization
preconditions, source revisions, an opaque binding ID, and a binding epoch.

### Owned bare-metal provider

Proposed group: `baremetal.fleet.datumapis.com/v1alpha1`.

| Kind | Authority and purpose |
| --- | --- |
| `BareMetalMachine` | One-to-one management extension for a Machine, including ObserveOnly/Managed policy and private BMC/console configuration. |
| `BareMetalMachineClass` | Qualified preparation, boot, installation-method, disk-selection, and cleanup requirements. |

Backend enrollment IDs, workflow IDs, templates, and progress remain private
adapter records. Enrollment starts in `ObserveOnly`; registration or discovery
does not authorize cleaning, power changes, boot inspection, or disk writes.

### Operating-system configuration

Proposed group: `fleet.datumapis.com/v1alpha1`, with typed adapters such as
`talos.fleet.datumapis.com`.

| Kind | Authority and purpose |
| --- | --- |
| `MachineImage` | Logical immutable OS release resolved by a typed provider to content-addressed artifacts. |
| `InterfaceClass` | Interface shape and address requirements; IPAM remains authoritative for live addresses. |
| `Installation` | Immutable initial-install or reinstall request bound to a MachineClaim, image, bootstrap input, and exact predecessor. |
| `OSConfigPolicy` | Managed-target configuration contribution composed into an immutable retained recipe. |

A reinstall creates another `Installation` that names the current predecessor UID.
The coordinator compares and commits that predecessor atomically. Policy
composition rejects disagreements instead of resolving them through arbitrary
priority. Changes are classified as Live, Reboot, Reinstall, or Unsupported.
Self-managed customer operating systems leave ongoing reconciliation after
handoff.

### Cluster orchestration

Proposed group: `fleet.datumapis.com/v1alpha1`.

| Kind | Authority and purpose |
| --- | --- |
| `Site` | Operational placement policy beneath a Location, selecting eligible Facilities. |
| `Cluster` | Managed-cluster intent and selected provider. |
| `MachinePool` | Desired capacity and admitted configuration for a role or pool. |
| `NodeClaim` | Durable logical node identity across machine replacement. |
| `Rollout` | Explicit admission and bounded progression of retained revisions. |

There is no Fleet `Node` kind. Kubernetes retains that resource. A `NodeClaim`
owns logical continuity while a `MachineClaim` owns one physical allocation;
shared identity becomes active only after the predecessor is fenced.

### Existing external authorities

Fleet integrates with, rather than absorbs, the Location service, networking and
IPAM, Galactic/routing, IAM, audit, telemetry, DNS, certificates, and service
addresses. Networking owns switch behavior, logical `NetworkAttachment` state,
and addresses. Fleet consumes pre-binding feasibility and binding-specific
attachment evidence.

Capacity publication should reuse the platform's existing IPClass
definition/reference and `ServiceConfiguration.spec.provisioning` patterns:

- [Provider IPClass definitions](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/apps/network-services-operator/platform-project/classes.yaml)
- [Service provisioning of references](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/apps/network-services-operator/components/control-plane-networking/service-catalog-provisioning-patch.yaml)
- [Project authorization limitation of Kubernetes RBAC](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/apps/vpc-system/control-plane/milo-ipam-rbac/rbac.yaml)

Fleet requires stronger allocation semantics than the current IPAM reference
implementation: authorization must be current at binding time and protected UID
provenance, rather than the copied reference or status, establishes identity.

## Architecture and authority boundaries

### Control-plane services

1. **Fleet API and state** authenticate intent, enforce field ownership and
   admission, and expose bounded status.
2. **Catalog publisher** installs local MachineClass references for projects with
   service entitlement and retains protected source provenance. It cannot reserve
   a Machine.
3. **Inventory registrar and verifier** commit exclusive physical reservations,
   bind approved claims to registered assets, and evaluate evidence and freshness.
4. **Machine coordinator** validates entitlement, source-use IAM, quota,
   compatibility, physical readiness, and network feasibility before making the
   only allocation decision.
5. **OS resolver and configuration service** pins exact images, bootstrap inputs,
   recipes, and native documents. Exactly one selected executor applies an
   admitted managed revision.
6. **Cluster orchestrator** creates shared MachineClaim and Installation requests,
   manages logical identity, and coordinates disruption and replacement.
7. **Infrastructure adapter** translates already-admitted work into provider or
   owned-hardware operations and returns normalized evidence. A backend may
   schedule workers but cannot choose, reassign, or free customer hardware.
8. **Console gateway** creates short-lived sessions scoped to principal, project,
   Claim UID, and binding epoch, and actively closes them on revocation.

Deployment may combine services, but it must preserve separate credentials,
field ownership, and authorization. Provisioning and recovery must run on
independently recoverable management capacity and remain available when a guest
or managed workload cluster fails.

### Durable state and recovery

Kubernetes objects and status carry public intent and observations. Status is
explicitly insufficient as the only record authorizing irreversible effects.
The implementation needs protected durable storage for:

- **Admission records:** exact provider and consumer Project UIDs, class/claim/
  machine UIDs, source and policy revisions, quota hold, public contract, binding
  ID and epoch, predecessor, and permitted effects.
- **Operation journal:** exact targets, artifacts and configuration digests,
  selected route/executor, idempotency key, backend operation correlation,
  cancellation boundary, progress, and completion evidence, persisted before
  external effects.
- **Artifact and revision store:** immutable content-addressed images, bootstrap
  inputs, compiled recipes, protected native documents, selections, and supported
  rollback prerequisites.
- **Inventory reservation and evidence state:** UID bindings, exclusive rack/
  termination/power reservations, evidence provenance, freshness, and invalidation.

If quota, policy, and reservation cannot commit in one transaction, durable
prepare/commit/abort records must provide equivalent conflict and recovery
guarantees. A timeout means the outcome is unknown; recovery queries the existing
backend operation and physical evidence before retrying. A Kubernetes Lease does
not fence a stale remote executor.

Release is a retained protocol: stop new effects, account for committed work,
revoke and close access, detach or fence networking, clean or restore according
to owner policy, and only then return capacity and quota. Unknown results retain
the reservation or quarantine the machine. Finalizers or equivalent records must
survive request and consumer-project deletion. Owned assets use Retain semantics.

### Security invariants

- Derive immutable Project UID from authenticated project API context. Project
  names, namespaces, labels, and copied status are not authorization boundaries.
- Separate class discovery, source-use permission, service entitlement, quota,
  and allocation. Revalidate all applicable policy at allocation commit.
- A controller's broad provider credentials never become customer permission.
  Use supported project-scoped IAM principals and PolicyBindings where Kubernetes
  RBAC cannot interpret Milo project identity.
- Resolve names to exact UIDs before admission. Reject source retargeting,
  reference chains/cycles, self-reference, forged provenance, name reuse, and
  late callbacks from an earlier binding epoch.
- Keep BMC/switch endpoints, credentials, supply selectors, full physical
  bindings, bootstrap secrets, rendered native configurations, and console bearer
  URLs out of consumer-visible status.
- Physical readiness is not permission to wipe. Every destructive operation pins
  the asset, binding, exact target, artifacts, permitted effect, and idempotency
  record before execution.
- Bootstrap authenticates the intended host. MAC or serial matching plus an
  unverified TLS connection is inadequate. Artifact ingestion must not permit
  arbitrary fetching from private management networks.
- Management, privileged provisioning, and consumer access remain separate trust
  boundaries. Old power, network, installation, console, and backend operations
  must fail after release or binding-epoch reassignment.

## Phased implementation plan

### Phase 0: settle contracts and ownership

- Decide group/version and project/operator/namespace scope for every kind.
- Specify schemas, defaults, validation, immutable fields, typed reference rules,
  field ownership, status conditions, finalizers, quotas, audit events, and
  operation/subresource shapes.
- Write storage and transaction ADRs for inventory reservations, cross-project
  bindings, quota, admission, journals, and revision retention.
- Define team and repository ownership for each subsystem and adapter.
- Generate clients, SDK types, API reference, examples, and conformance fixtures.

### Phase 1: observation-only APIs

- Install APIs, admission, protected stores, and status without enabling effects.
- Import existing Omni, servers.com, NetActuate, IPAM, inventory, and repository
  observations; retain contradictions and unknown evidence.
- Prove that backfill does not install, wipe, power, rename, or reallocate assets.

Existing migration sources include the current
[Omni fleet layout](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/provisioning/fleet/README.md)
and the existing provisioning templates. The older Infrastructure Platform
inventory and Fleet Operations proposals should be reconciled rather than left as
a second writable authority.

### Phase 2: qualified inventory lab

- Implement Facility, Rack, DeviceClaim, Device, Cable, and PowerFeed plus the
  trusted registrar and evidence evaluator.
- Qualify one rack and hardware family, including management, data, console where
  used, power, patch-panel, and breakout paths.
- Test conflicting reservations, wrong identity/port/rack, stale or missing
  evidence, duplicate registration, restart, name reuse, moves, and recovery of
  the reservation database.

### Phase 3: catalog and shared allocator

- Implement provider definitions and consumer references, the service publisher,
  protected provenance, source-use IAM, quota, and the shared binding transaction.
- Publish one provider class into two consumer projects.
- Race overlapping classes and projects for one Machine; require exactly one
  winner. Test revocation during commit, project/class recreation, exhaustion,
  delayed publication, controller failure, and project deletion.

### Phase 4: owned-metal adapter qualification

- Implement BareMetalMachine/Class, ObserveOnly adoption, explicit preparation,
  the operation journal, artifact delivery, and the console gateway.
- On disposable lab servers prove correct-host authentication, exact disk and
  component targeting, isolated provisioning, boot outcome, operation query and
  deduplication, safe cancellation, stale-executor exclusion, console revocation,
  cleanup, quarantine, backup/restore, and integration handover.
- Keep the execution backend pluggable and publish capability gaps explicitly.

### Phase 5: standalone general-purpose pilot

- Qualify a supported general-purpose OS other than Talos with no Cluster,
  NodeClaim, Kubernetes guest, or mandatory ongoing guest agent.
- Exercise claim, network feasibility and handoff, initial installation, power,
  console, reinstall with predecessor concurrency, release, cleanup, and reuse.
- Demonstrate cross-project exclusivity, quota enforcement, stale operation
  rejection after reassignment, and customer ownership of the OS after handoff.

### Phase 6: managed cluster adoption

- Implement Site, Cluster, MachinePool, NodeClaim, Rollout, and the Omni adapter.
- Import existing cluster, pool, node identity, and configuration without reset or
  automatic rename.
- Compare candidates with current outputs and transfer one field or resource at a
  time only after equivalent output is demonstrated.
- Pilot shared allocation, predecessor fencing, machine replacement, release,
  cluster readiness, and Karmada output.

### Phase 7: controlled updates and production readiness

- Enable managed OS composition and retained per-node revision admission.
- Prove policy conflict behavior, pause, partial-rollout repair, fresh-health and
  disruption gates, controller restart, rollback retention, and late-callback
  rejection.
- Exercise a second execution adapter against the conformance contract.
- Qualify backups, storage restoration, provider failure and handover, management
  plane recovery, physical maintenance, and cross-consumer reuse before widening
  facilities or providers.

## Open decisions

The following choices are unresolved in PR #5782 and should block the relevant
implementation stage rather than be inferred from the examples:

- Exact API groups, versions, namespacing, administrative scopes, operation
  endpoints/subresources, and compatibility strategy.
- Repository and team ownership for the API service, inventory, allocator, OS,
  cluster orchestration, publisher, operation coordinator, and each adapter.
- Durable database and transaction mechanism for physical reservations,
  cross-project allocation, policy/quota commit, effect admission, fencing, and
  failover.
- Representation, custody, retention, backup, and garbage collection of recipes,
  selections, native configuration, journals, evidence, and rollback artifacts.
- Trusted enrollment and identity-deduplication mechanisms by hardware family;
  evidence sources, freshness, invalidation, and technician-attestation rules.
- First supported rack, patching, breakout, power, and maintenance/move models.
- Provisioning backend selection and adapter conformance; `/prepare`, console,
  reset/reboot/rescue/media semantics; cancellation boundaries; remote-executor
  fencing.
- First qualified hardware, firmware, image, bootstrap, installation method, and
  cleaning combinations, including the non-Talos first-release OS and the
  assurance offered between tenants.
- Omni/Talos single-writer and final-verification contract, adopted Kubernetes
  node names, and durable Rollout revision representation.
- Deployment topology, replica/failure-domain targets, secret custody, management
  network isolation, audit retention, telemetry, and service-level objectives.
- Initial portal, CLI, SDK, and agent experiences and their measurable success
  criteria.

These decisions do not weaken the proposal's central invariants: one allocator,
separate subsystem authorities, explicit effect admission, exact identity and
content pinning, non-destructive adoption, and release completion before reuse.
