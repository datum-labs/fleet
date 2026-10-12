# Architecture

The [Fleet proposal](https://github.com/datum-cloud/infra/pull/5782) defines
the service lifecycle and authority boundaries. The initial implementation
should serve MachineClass, Machine, and MachineClaim through
[standard CRDs](api-serving-decision.md), with a Fleet controller coordinating
Milo quota and exclusive allocation. The [binding record](binding-placement.md)
lives in the consumer Project with the MachineClaim. A provider Machine
reservation prevents concurrent use, and consumer Project finalization must
wait for release. Stale-worker exclusion and recovery must be proven before
enabling physical effects.
