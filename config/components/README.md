# Kustomize components

The available components are:

- `namespace`: creates `fleet-system`
- `api-registration`: registers the aggregated API after resources are ready
- `cert-manager-ca`: provisions a namespaced serving certificate CA
- `observability`: adds a Prometheus `ServiceMonitor`
- `tracing`: mounts the API server tracing configuration

Include a component from an overlay, for example:

```yaml
components:
  - ../../components/cert-manager-ca
```
