# Deployment configuration

This directory contains Kustomize manifests for the Fleet Manager aggregated
API server.

## Development overlay

Render the development deployment with:

```bash
kubectl kustomize config/overlays/dev
```

The overlay creates `fleet-system`, deploys
`ghcr.io/datum-labs/fleet-manager`, configures a cert-manager CA, and adds the
observability and tracing components. It intentionally does not install the
`APIService` component because Fleet Manager does not publish a resource API
yet.

Once the first `fleet.datumapis.com/v1alpha1` resources and their storage are
implemented, add `../../components/api-registration` to an overlay. Registering
it earlier would advertise an API group with no supported resources.

## Components

- `namespace`: creates `fleet-system`
- `api-registration`: registers `fleet.datumapis.com/v1alpha1`; deferred until
  resources exist
- `cert-manager-ca`: provisions the serving certificate CA
- `observability`: adds a Prometheus `ServiceMonitor`
- `tracing`: supplies the Kubernetes API server tracing configuration

The base expects the cert-manager CSI driver and an existing
`selfsigned-cluster-issuer`. Review certificate trust and replace the
development defaults before deploying outside an isolated environment.

## Troubleshooting

```bash
kubectl get pods -n fleet-system
kubectl logs -n fleet-system -l app=fleet-manager-apiserver --tail=50
kubectl get events -n fleet-system --sort-by=.lastTimestamp
```
