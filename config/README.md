# Deployment configuration

The supported first API artifact is the standard CRD Kustomization in `crd`.
It installs `MachineClass`, `Machine`, and `MachineClaim` in
`fleet.datumapis.com/v1alpha1`:

```bash
kubectl apply -k config/crd
```

There is deliberately no `APIService` for this group/version. The legacy
aggregated-server deployment under `base` is retained only while the service
entry point is replaced by a controller manager; it is not part of the public
API serving path.

## Development overlay

Render the development deployment with:

```bash
kubectl kustomize config/overlays/dev
```

The overlay creates `fleet-system`, deploys
`ghcr.io/datum-labs/fleet`, configures a cert-manager CA, and adds the
observability and tracing components. It intentionally does not install the
removed `APIService` component. Install the public CRDs separately with the
command above.

The development overlay still exercises the legacy process and its operational
components. Installing the CRDs is independent of that process.

## Components

- `namespace`: creates `fleet-system`
- `crd`: public Fleet CRDs
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
