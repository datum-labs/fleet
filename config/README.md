# Deployment configuration

This directory contains a non-serving Kustomize scaffold for an aggregated API
server. The first public Fleet APIs are expected to use standard CRDs; see the
[API serving decision](../docs/api-serving-decision.md). Do not deploy this
scaffold as the first Fleet API implementation.

## Development overlay

Render the development deployment with:

```bash
kubectl kustomize config/overlays/dev
```

The overlay creates `fleet-system`, deploys
`ghcr.io/datum-labs/fleet`, configures a cert-manager CA, and adds the
observability and tracing components. It intentionally does not install the
`APIService` component because Fleet does not publish a resource API
yet.

The `api-registration` component is retained from the scaffold and is not part
of the CRD deployment path. An APIService registration would claim the entire
`fleet.datumapis.com/v1alpha1` path, which the CRDs need to own.

## Components

- `namespace`: creates `fleet-system`
- `api-registration`: unused aggregated-server registration scaffold
- `cert-manager-ca`: provisions the serving certificate CA
- `observability`: adds a Prometheus `ServiceMonitor`
- `tracing`: supplies the Kubernetes API server tracing configuration

The base expects the cert-manager CSI driver and an existing
`selfsigned-cluster-issuer`. Review certificate trust and replace the
development defaults before deploying outside an isolated environment.

## Troubleshooting

```bash
kubectl get pods -n fleet-system
kubectl logs -n fleet-system -l app=fleet-apiserver --tail=50
kubectl get events -n fleet-system --sort-by=.lastTimestamp
```
