# Kustomize components

The available components are:

- `namespace`: creates `fleet-system`
- `cert-manager-ca`: provisions a namespaced serving certificate CA
- `observability`: adds a Prometheus `ServiceMonitor`
- `tracing`: mounts the API server tracing configuration

Include a component from an overlay, for example:

```yaml
components:
  - ../../components/cert-manager-ca
```

There is no API registration component. Standard CRDs own
`fleet.datumapis.com/v1alpha1`; registering an aggregated `APIService` for the
same group/version would conflict with them.
