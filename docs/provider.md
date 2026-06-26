# Provider Specific Functionality

While k8ify tries to be as provider agnostic as possible, some functionality depends on how k8s is set up and which operators are available.


## Encrypted Volume Scheme

How encrypted volumes have to be set up depends on the target k8s setup. Therefore k8ify allows you to specify which encryption scheme to use via the `x-targetCfg` option.

Example:
```
x-targetCfg:
  encryptedVolumeScheme: appuio-cloudscale
```

### appuio-cloudscale

Usage of encrypted volumes on APPUiO Cloudscale is documented at [LUKS Encrypted Volumes](https://hub.syn.tools/csi-cloudscale/index.html), but with k8ify you should not need to worry about this.

Example:
```
volumes:
  mongodb-data:
    labels:
      k8ify.storageClass: ssd-encrypted # or bulk-encrypted
```

The storage classes "ssd-encrypted" or "bulk-encrypted" trigger the LUKS encryption support.

k8ify expects you to specify the LUKS encryption keys via environment variables. These environment variables must have specific names, which depend on the names of the volumes and environments. The easiest way to set this up is to run k8ify without these variables set, and it will print error messages containing the names of variables it expects.

LUKS encryption keys should have an entropy of at least 512 bits. Suitable keys can be generated e.g. via `pwgen -s 100 1`. Be sure to restrict the visibility of environment variables containing LUKS keys appropriately.

## Plain LoadBalancer Scheme

To expose k8s Services of type LoadBalancer directly for non HTTP traffic (without an Openshift route or an ingress), some providers need additional manifests
or configuration.
This can be used to expose a database.

### appuio-cloudscale

On APPUiO Cloudscale, you need a `CiliumNetworkPolicy` to expose a plain k8s Service of type LoadBalancer. Refer to [Change to LoadBalancer on APPUiO](https://github.com/appuio/appuio-cloud-community/discussions/60) for
details.

To enable this, set `x-targetCfg.exposePlainLoadBalancerScheme` to
`appuio-cloudscale`:

```yaml
x-targetCfg:
  exposePlainLoadBalancerScheme: appuio-cloudscale
```

For a full example, see [compose.yml: expose plain on appuio](/tests/golden/expose-plain-appuio/compose.yml).

## Security Context Defaults

Operators can enforce pod-level hardening defaults for all services in a
compose file through `x-targetCfg.securityContext`. It accepts the **pod-level**
subset only (the same set as `k8ify.podSecurityContext.*`). Container-level
keys (`capabilities`, `privileged`, `allowPrivilegeEscalation`,
`readOnlyRootFilesystem`) are **not** cluster-wide — containers vary too much,
and the pod level is where the operator hardening story lives. Placing a
container-level key under `x-targetCfg.securityContext` is a hard validation
error.

Defaults are **field-level merged** with per-service `k8ify.podSecurityContext.*`
labels — a service label overrides the cluster default per field, so a service
that only sets `fsGroup` still benefits from a cluster-wide
`fsGroupChangePolicy: OnRootMismatch`. k8ify itself ships no built-in
hard-coded defaults that change existing output, so enabling this is opt-in for
operators.

The recommended default for services mounting volumes with many files is
`fsGroupChangePolicy: OnRootMismatch` (a pure reduction in kubelet relabeling
work, effective when `fsGroup` is also set — services set their own `fsGroup`
per deployment).

```yaml
x-targetCfg:
  securityContext:
    fsGroupChangePolicy: OnRootMismatch
    seccompProfile:
      type: RuntimeDefault
    runAsNonRoot: true
```

Per-service labels override these defaults field by field, and
`k8ify.converter` services ignore them (they short-circuit before pod
construction).
