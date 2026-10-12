# Architecture

The [Fleet proposal](https://github.com/datum-cloud/infra/pull/5782) defines
the service lifecycle and authority boundaries. The initial implementation
should serve MachineClass, Machine, and MachineClaim through
[standard CRDs](api-serving-decision.md), with a Fleet controller coordinating
Milo quota and exclusive allocation. The protected binding and operation store
must survive deletion of a consumer project and exclude stale workers; its
storage design remains to be proven before enabling physical effects.
