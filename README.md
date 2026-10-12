# Fleet

Fleet is an incubating Datum service for managing machine fleets and
their lifecycle through Kubernetes-style APIs.

The repository currently contains the service and deployment scaffold. The API
group `fleet.datumapis.com` and its initial `v1alpha1` version are reserved, but
the server intentionally publishes no resources until the first API contract is
accepted and implemented. This avoids turning template examples into a public
contract.

## Repository layout

- `cmd/fleet`: service entry point and command-line options
- `internal/apiserver`: Kubernetes aggregated API server setup
- `pkg/apis/fleet`: API group registration
- `pkg/contracts/v1alpha1`: proposed allocation and installation contracts,
  with validation tests; these types are not served yet
- `config`: Kustomize base and optional deployment components
- `.github/workflows`: tests and release artifact publishing

Read the [proposal investigation](docs/proposal-investigation.md) for the full
service design and open decisions, and the [first contract slice](docs/allocation-contract-slice.md)
for the exact scope of the initial Go types.

Fleet uses [Milo quota](docs/quota-integration.md) for machine allocations;
the Fleet binding record stores the granted Milo ResourceClaim UID.

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
