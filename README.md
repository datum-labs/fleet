# Fleet

Fleet is an incubating Datum service for managing machine fleets and
their lifecycle through Kubernetes-style APIs.

The repository currently contains a non-serving API server scaffold and draft
contracts. [Standard CRDs are the recommended starting point](docs/api-serving-decision.md)
for MachineClass, Machine, and MachineClaim. The server scaffold publishes no
resources and is not the selected serving path for these first APIs.

## Repository layout

- `cmd/fleet`: service entry point and command-line options
- `internal/apiserver`: unused aggregated API server scaffold
- `pkg/apis/fleet`: API group registration
- `pkg/contracts/v1alpha1`: proposed allocation and installation contracts,
  with validation tests; these types are not served yet
- `config`: Kustomize base and optional deployment components
- `.github/workflows`: tests and release artifact publishing

Read the [proposal investigation](docs/proposal-investigation.md) for the full
service design and open decisions, and the [first contract slice](docs/allocation-contract-slice.md)
for the exact scope of the initial Go types.

Fleet uses [Milo quota](docs/quota-integration.md) for machine allocations.
The protected [binding record](docs/binding-placement.md) lives in the
consumer Project with its MachineClaim.

## Development

Requirements:

- Go 1.26 or later
- [Task](https://taskfile.dev) for the convenience commands
- Docker for building the container image

Run the local checks:

```bash
task build
task test
task vet
```

Build the development image:

```bash
task dev:build
```

The image is `ghcr.io/datum-labs/fleet`. See
[`config/README.md`](config/README.md) for the deployment manifests and their
current limitations.

## Status

This repository lives in `datum-labs` while the service and its API contracts
are being developed. It is not a production service yet.

## License

Fleet is licensed under the GNU Affero General Public License v3.0. See
[`LICENSE`](LICENSE).
