# Fleet API contract slice

This initial contract package turns the settled lifecycle invariants in
`datum-cloud/infra#5782` at commit `51781e2456367be82281e1443ca56d0686311133`
into reviewable Go contracts at `pkg/contracts/v1alpha1`. These types are not
served yet; CRD schemas, admission, and controllers are required before a client
can create these resources. See [API serving approach](api-serving-decision.md).
The [first API proposal](first-api-proposal.md) recommends narrowing the served
Claim schema and using native Kubernetes identity instead of some duplicated
UID fields in these exploratory types.

## Scope

The slice covers:

- mutually exclusive provider-definition and consumer-reference MachineClass modes;
- immutable one-hop source identity and explicitly namespaced provider supply;
- local MachineClaim inputs, immutable binding constraints, and mutable power intent;
- protected BindingRecord identity in the consumer Project using Project, class,
  claim, and Machine UIDs rather than names;
- immutable Installation input and UID-fenced reinstall predecessors;
- conditions correlated to generation, binding epoch, and resolved input digest;
- protected InstallationPlan authority persisted before external effects; and
- release obligations that keep the binding and Milo quota claim until cleanup succeeds.

The public resources contain local name references only. Foreign provider identity,
the Machine UID, authorization evidence, and cleanup obligations are kept in
protected service records. Fleet resolves its Milo quota claim when reconciling
the binding. A source project name is only a lookup address;
it never proves permission.

## Deliberately unresolved

The draft proposal does not yet decide several implementation details. This module
does not silently choose them:

- Kubernetes ObjectMeta, generated deep-copy code, CRD schemas, admission, and
  controller wiring are omitted. The structs are dependency-free contract models
  pending implementation of the public CRDs.
- The authoritative Project API reference shape and cross-API-server resolver are
  open. `SourceReference.Project` therefore remains the proposal's address string,
  while the protected record requires the resolved immutable Project UID.
- The same protected binding model supports both cross-project allocations and a
  definition serving an authorized claim in its own project.
- MachineClass definition fields are a minimal hardware/access subset. Interface,
  placement, availability summaries, complete capability negotiation, and CEL
  schema rules remain to be specified.
- MachineClaim Site/Facility constraints, network object API groups, status
  visibility, opaque public allocation handle, and condition vocabulary remain
  design work.
- Installation disk/interface targets, cancellation boundaries, executor protocol,
  and durable current-incarnation compare-and-swap storage are not selected here.
- Finalizer names and force-deletion policy require the API server's project
  deletion and service cleanup design. The contract instead makes the durable
  release gate directly testable.

## Review and test

Run:

```sh
go test ./pkg/contracts/v1alpha1
```

The tests exercise mode exclusivity, source retarget rejection, bound-claim
immutability, allowed power-state changes, predecessor UID fencing, project
UID identity, and release gating.
