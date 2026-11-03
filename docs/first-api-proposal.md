# Proposal: first Fleet CRDs and allocation control loop

Status: proposed for review. This refines the agreed public API group, CRD
serving approach, consumer-Project binding placement, and Milo quota decision.
It does not enable allocation or physical effects.

Implementation note: the API slice in this repository installs the three public
CRDs and Go types only. Allocation remains disabled. Before the allocation slice,
Service Catalog must support `claimingResources` without generating an automatic
`ClaimCreationPolicy`, as described below.

## Recommendation

Build the first public APIs as standard CRDs in `fleet.datumapis.com/v1alpha1`.
Serve `MachineClass`, `Machine`, and `MachineClaim` in the owning Milo Project
control planes. Add a service-only, namespaced `MachineBinding` in
`internal.fleet.datumapis.com/v1alpha1` beside each bound `MachineClaim`. Keep
the binding and its cleanup obligations in that consumer Project. Use one
conditional reservation field on the provider `Machine` to exclude competing
consumer Projects; that pointer is not a second binding.

Milo's Project purger deletes namespaced resources and waits for their
finalizers, then reclaims the Project control plane. It deliberately omits
cluster-scoped resources. A namespaced binding with a release finalizer can
therefore hold normal Project deletion until the Machine is cleaned and released.
This is a stronger reason for namespaced binding placement than a generic
preference for putting related objects together.

## CRD contract

| Kind | Scope inside its Project | Writer | First-release intent |
| --- | --- | --- | --- |
| `MachineClass` | Cluster | Provider operator for definitions; Fleet publisher for consumer references | Immutable `definition` or one-hop `source {project,name}`; sanitized resolution status. |
| `Machine` | Namespaced in provider Project | Fleet/provider integration | One canonical allocatable physical unit, eligibility status, and one protected reservation pointer. |
| `MachineClaim` | Namespaced in consumer Project | Authorized consumer | Local class and location request; status reports pending reason or opaque allocation handle. |
| `MachineBinding` | Same namespace and Project as `MachineClaim` | Fleet only | Admitted contract, allocation phase and epoch, chosen Machine, and release obligations. No public read access. |

Keep the first `MachineClaim.spec` limited to class and location constraints.
Power, networking, installation, and console fields should enter the served API
when their controllers and authorization exist. The current Go contract package
is exploratory; it is not a served schema. `Machine` must not duplicate the
future `Device` hardware inventory. It describes allocatable supply and its
readiness, while inventory remains the physical source of truth.
`MachineClass` remains cluster-scoped inside a Project control plane as the
Fleet proposal specifies. It does not block Project deletion; verify that a
recreated Project receives a fresh control plane and cannot inherit classes
from its predecessor.

Use structural OpenAPI schemas and CEL for object-local rules: the class
definition/source union, required typed references, enums and bounds,
immutability of class source and Machine identity, and immutable Claim class and
placement after creation. Use admission and controllers for cross-object
resolution, source-use authorization, canonical physical identity, and
reservation transitions. Enable `/status` subresources and use
`observedGeneration` plus stable conditions. Candidate first conditions are
`Resolved` on MachineClass, `Available` on Machine, and `Accepted`, `QuotaReady`,
`Bound`, and `Releasing` on MachineClaim. Claim status must not expose the
provider Machine, private endpoints, or the full binding.

## Identity without redundant UID fields

Public specs use names and typed references; Kubernetes already gives each
object `metadata.uid`. Do not copy Milo `ResourceClaim` UIDs or add UIDs to
customer-authored specs. Give the binding a deterministic name derived from the
Claim's native `metadata.uid` (for example a sufficiently long hash) so a new
Claim with the same namespace/name cannot recover an older binding. The binding
may have a native same-namespace owner reference to the Claim. This uses
Kubernetes identity rather than duplicating it into a Fleet field.

The provider Machine reservation stores the consumer Project's immutable UID
and routable address, namespace, binding name, and fencing epoch. This single
cross-Project identity field prevents a reused Project name from reaching a
new control plane. The unique binding name identifies the exact Claim
incarnation; the epoch rejects stale workers within one binding. No Milo quota
claim UID or separate Claim/Machine UID field is needed in this pointer. This
Project UID is the one proposed exception to avoiding copied UIDs: if the rule
also forbids it, Milo must provide an equivalent immutable routing identity.
Reconciliation must compare it to the live binding before any physical effect.
An inaccessible or mismatched binding means `Quarantined`, never `Available`.
Project and Claim name reuse need explicit conformance tests.

## Binding and Milo quota sequence

1. Add a Fleet release finalizer to the Claim, persist it, and reread the
   Claim before any Binding or quota work. Resolve its local class to one
   provider definition and check the consumer Project's current entitlement and
   source-use permission. Choose a physically eligible candidate without
   reserving it. A Claim waiting for supply has no Binding and consumes no quota.
2. Create or recover its deterministic `MachineBinding` with a Fleet release
   finalizer. Record the admitted public contract, source revision, candidate,
   and a new fencing epoch. Put the finalizer in the initial CREATE. The
   binding remains `Preparing`.
3. Create or recover a deterministically named Milo `ResourceClaim` in the same
   consumer Project for one `fleet.datumapis.com/machines` unit. Set its
   `consumerRef` to that Project and `resourceRef` to the Binding; register
   `MachineBinding` as an allowed claiming kind. Add a Fleet finalizer on claim
   CREATE so Project purge cannot free quota during physical cleanup. Wait for
   `Granted=True` for the current generation, checking the consumer, target,
   amount, and resource type on every recovery. Do not store its UID. A denied
   claim must be deleted and its deletion observed before retry. Quota denial
   or Milo unavailability leaves the customer Claim pending and Machine free.
4. Recheck authorization, quota, and Machine eligibility. Conditionally update
   the provider Machine's protected reservation using `resourceVersion` and an
   empty-reservation precondition. Only one competing consumer can win. A losing
   attempt releases its unneeded quota claim and binding or retries a different
   eligible Machine under a bounded hold. Advance the epoch for each candidate
   attempt; never switch candidates until the prior reservation is proven
   absent or its matching pointer has been cleared.
5. Mark the Binding active and publish a safe summary in Claim status. If a
   crash occurs between Project API writes, reconcile the existing Binding,
   Milo claim, and provider Machine pointer. Never interpret a timeout as a
   failed write without reading back the actual state.

Use **explicit** controller-managed Milo ResourceClaims for this first release,
as Compute does. Milo `ClaimCreationPolicy` is an alternative for a fixed cost
per Binding, but Service Catalog currently generates a `generateName` quota
claim. Milo admission creates the claim before the triggering resource is
persisted; retries can create temporary extra claims if persistence fails.
Explicit deterministic naming makes Fleet's cross-Project recovery easier to
reason about. Do not configure an automatic claim policy for the same Binding,
or one allocation could consume quota twice. Today's Service Catalog derives a
ResourceRegistration's allowed claiming kinds from `metricRules`; omitting a
rule substitutes `Project` rather than `MachineBinding`, while adding the rule
also creates the automatic policy. **The allocation slice therefore depends on
a small Service Catalog change**: let a quota limit declare allowed claiming
kinds without an automatic metric rule. Add an optional
`claimingResources: [{apiGroup, kind}]` field to a quota limit, union it with
selectors from metric rules, and deduplicate the result. Fleet declares a
Project machine limit with `MachineBinding` as its claiming kind and no metric
rule. Service Catalog continues to own the registration and entitlement grants;
Fleet owns individual ResourceClaims. Test that this configuration produces a
ResourceRegistration and default grants, but no ClaimCreationPolicy, while
existing metric-rule behavior remains unchanged.

## Release and deletion

The Claim, Binding, and Milo quota claim carry Fleet finalizers before provider
reservation. Release stops new work, fences executors by binding epoch,
revokes access, detaches networking, and completes cleaning. Then Fleet
conditionally clears and observes the matching provider Machine reservation,
removes the quota claim's Fleet finalizer, deletes and observes deletion of the
quota claim, removes the Binding finalizer, and finally removes the Claim
finalizer. If the provider reservation cannot be proven clear, keep quota held.
The provider Machine has its own protection finalizer while eligible or
reserved so it cannot disappear mid-allocation.

Milo Project deletion may send delete requests to the Claim, Binding, and quota
claim concurrently; reconciliation cannot rely on an API deletion order. It
keeps the Project control plane until namespaced finalizers finish. Fleet
should report the Binding as a deletion blocker while cleanup is in progress.
If the consumer control plane is lost outside normal finalization, keep the provider
Machine reserved or quarantined for explicit recovery; absence of a Binding is
not proof that access was revoked or hardware cleaned.

## Access, deployment, and phases

Register Fleet CRDs for Project discovery, and register Milo
`ProtectedResource` permissions for each kind, `/status`, and finalizers, all
parented to Project. Consumer roles may create/read their Claims. Provider
roles manage class definitions and observe their Machines. Grant Fleet's
workload identity explicit cross-Project Binding CRUD/finalizer and Machine
reservation-status permissions; omit Binding permissions from consumer and
provider roles. Admission must fail closed for requester-sensitive Binding
writes and finalizer removal. Source-use is a separate IAM `use`
permission checked in the source Project context, not implied by a published
class or Fleet's broad credentials.

Publish CRD, IAM, and service-catalog manifests as a versioned Kustomize
artifact, following Compute's Milo integration. Run a controller manager with
Milo-aware `multicluster-runtime` Project watches, preferably limited to
entitled Service Catalog consumers, and leader election. Use a dedicated
service account, restricted pod security, probes, and separate credentials for
discovery, Project, and quota operations. Retire the inherited non-serving
aggregated API server scaffold; do not register an `APIService` for the public
group/version.

1. **API slice:** CRDs, generated clients, schema/CEL tests, IAM, status and
   admission; no allocation or external effects.
2. **Observation slice:** class publication and Machine import/readiness from
   existing providers; report contradictions and pending reasons.
3. **Allocation slice:** consumer Binding, explicit Milo quota, provider
   reservation, release, and crash recovery. Land the Service Catalog quota
   registration change first. Keep provisioning disabled until the acceptance
   cases below pass.
4. **Effects slice:** network, power, installation, and console authorization
   and operation journals, each behind its own safety tests.

## Acceptance gates for allocation

- Two consumer Projects race through overlapping classes for one Machine;
  exactly one provider reservation commits.
- Reconcile the same Claim concurrently; one Binding and at most one active
  quota claim remain. Inject conflicts and crashes at every cross-Project step.
- Quota denial, unavailable Milo, and revoked source-use permission activate no
  Machine. Pending Claims consume no quota; quota stays held while a Binding
  is cleaning up, including during Project purge.
- Claim deletion, namespace deletion, and Project deletion retain finalizers
  until cleanup. Loss of the consumer control plane quarantines the provider
  Machine. Name reuse and stale callbacks cannot act on a later binding.
- Project recreation never sees the prior Project's cluster-scoped classes;
  Project purging a terminating ResourceClaim keeps its quota counted until
  Fleet clears the claim's finalizer after physical cleanup.
- A consumer cannot read, create, edit, delete, or remove a finalizer from the
  protected Binding; a provider cannot overwrite the reservation field.

## Sources and evidence

- [Fleet class publication and allocation proposal](https://github.com/datum-cloud/infra/blob/51781e2456367be82281e1443ca56d0686311133/docs/enhancements/fleet-management/capacity-publication.md)
- [Milo Project deletion controller](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/internal/controllers/resourcemanager/project_controller.go)
- [Milo namespaced Project purge](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/internal/controllers/projectpurge/purge.go)
- [Milo quota admission behavior](https://github.com/milo-os/milo/blob/718749b934aba630da4efb834b416aed4f5bfedb/pkg/quota/admission/plugin.go)
- [Service Catalog quota policy generation](https://github.com/milo-os/service-catalog/blob/8cb645e756ba6bd19eb275c0bcbf7321cfcffc81/internal/controller/quota_fanout.go)
- [Service Catalog entitlement quota grants](https://github.com/milo-os/service-catalog/blob/8cb645e756ba6bd19eb275c0bcbf7321cfcffc81/internal/controller/entitlement_quota_grants.go)
- [Compute's explicit ResourceClaim lifecycle](https://github.com/datum-cloud/compute/blob/d70ed7baf0ee81cb5ec46a272770c64311af5048/internal/controller/instance_controller.go)
- [Kubernetes CRD validation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [Kubernetes conditional updates](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Kubernetes owner-reference scope](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)
