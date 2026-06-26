package converter

import (
	core "k8s.io/api/core/v1"

	"github.com/vshn/k8ify/pkg/ir"
)

// mergePodSecurityContextSpec combines a per-service spec (svc) and a
// cluster-wide default (def) with field-level precedence: a non-nil svc field
// wins over a non-nil def field. Both may be nil. The resulting
// *core.PodSecurityContext is nil when every field is zero.
func mergePodSecurityContextSpec(svc, def *ir.PodSecurityContextSpec) *core.PodSecurityContext {
	fsGroup := pickInt64Ptr(svc, def, func(s *ir.PodSecurityContextSpec) *int64 { return s.FSGroup })
	fsGroupChangePolicy := pickPolicy(svc, def)
	seLinuxOptions := pickSELinuxOptions(svc, def)
	supplementalGroups := pickSupplementalGroups(svc, def)
	seccompProfile := pickSeccompProfile(svc, def)
	runAsUser := pickInt64Ptr(svc, def, func(s *ir.PodSecurityContextSpec) *int64 { return s.RunAsUser })
	runAsGroup := pickInt64Ptr(svc, def, func(s *ir.PodSecurityContextSpec) *int64 { return s.RunAsGroup })
	runAsNonRoot := pickBoolPtr(svc, def, func(s *ir.PodSecurityContextSpec) *bool { return s.RunAsNonRoot })

	if fsGroup == nil && fsGroupChangePolicy == nil && seLinuxOptions == nil &&
		len(supplementalGroups) == 0 && seccompProfile == nil &&
		runAsUser == nil && runAsGroup == nil && runAsNonRoot == nil {
		return nil
	}

	result := &core.PodSecurityContext{
		FSGroup:             fsGroup,
		FSGroupChangePolicy: fsGroupChangePolicy,
		SELinuxOptions:      seLinuxOptions,
		SeccompProfile:      seccompProfile,
		RunAsUser:           runAsUser,
		RunAsGroup:          runAsGroup,
		RunAsNonRoot:        runAsNonRoot,
	}
	if len(supplementalGroups) > 0 {
		result.SupplementalGroups = supplementalGroups
	}
	return result
}

func pickInt64Ptr(svc, def *ir.PodSecurityContextSpec, get func(*ir.PodSecurityContextSpec) *int64) *int64 {
	if svc != nil {
		if v := get(svc); v != nil {
			return v
		}
	}
	if def != nil {
		if v := get(def); v != nil {
			return v
		}
	}
	return nil
}

func pickBoolPtr(svc, def *ir.PodSecurityContextSpec, get func(*ir.PodSecurityContextSpec) *bool) *bool {
	if svc != nil {
		if v := get(svc); v != nil {
			return v
		}
	}
	if def != nil {
		if v := get(def); v != nil {
			return v
		}
	}
	return nil
}

func pickPolicy(svc, def *ir.PodSecurityContextSpec) *core.PodFSGroupChangePolicy {
	if svc != nil && svc.FSGroupChangePolicy != nil {
		return svc.FSGroupChangePolicy
	}
	if def != nil && def.FSGroupChangePolicy != nil {
		return def.FSGroupChangePolicy
	}
	return nil
}

func pickSELinuxOptions(svc, def *ir.PodSecurityContextSpec) *core.SELinuxOptions {
	if svc != nil && svc.SELinuxOptions != nil {
		return svc.SELinuxOptions
	}
	if def != nil && def.SELinuxOptions != nil {
		return def.SELinuxOptions
	}
	return nil
}

func pickSeccompProfile(svc, def *ir.PodSecurityContextSpec) *core.SeccompProfile {
	if svc != nil && svc.SeccompProfile != nil {
		return svc.SeccompProfile
	}
	if def != nil && def.SeccompProfile != nil {
		return def.SeccompProfile
	}
	return nil
}

func pickSupplementalGroups(svc, def *ir.PodSecurityContextSpec) []int64 {
	if svc != nil && len(svc.SupplementalGroups) > 0 {
		return svc.SupplementalGroups
	}
	if def != nil && len(def.SupplementalGroups) > 0 {
		return def.SupplementalGroups
	}
	return nil
}

// buildPodSecurityContext builds the pod-level security context for a workload
// by field-merging the per-service labels with the cluster-wide TargetCfg
// default. Returns nil when nothing is configured (no output change).
func buildPodSecurityContext(workload *ir.ParentService, targetCfg ir.TargetCfg) *core.PodSecurityContext {
	def, _ := targetCfg.PodSecurityContextDefault()
	svc, _ := ir.PodSecurityContextSpecFromLabels(workload.Labels())
	return mergePodSecurityContextSpec(svc, def)
}

// buildContainerSecurityContext converts an intermediate spec into a
// *core.SecurityContext. nil spec or an all-zero spec returns nil.
func buildContainerSecurityContext(spec *ir.ContainerSecurityContextSpec) *core.SecurityContext {
	if spec == nil {
		return nil
	}
	if spec.RunAsUser == nil && spec.RunAsGroup == nil && spec.RunAsNonRoot == nil &&
		spec.ReadOnlyRootFilesystem == nil && spec.AllowPrivilegeEscalation == nil &&
		spec.Privileged == nil && spec.Capabilities == nil && spec.SeccompProfile == nil &&
		spec.SELinuxOptions == nil {
		return nil
	}
	result := &core.SecurityContext{
		RunAsUser:                spec.RunAsUser,
		RunAsGroup:               spec.RunAsGroup,
		RunAsNonRoot:             spec.RunAsNonRoot,
		ReadOnlyRootFilesystem:   spec.ReadOnlyRootFilesystem,
		AllowPrivilegeEscalation: spec.AllowPrivilegeEscalation,
		Privileged:               spec.Privileged,
		SeccompProfile:           spec.SeccompProfile,
		SELinuxOptions:           spec.SELinuxOptions,
	}
	if spec.Capabilities != nil && (len(spec.Capabilities.Add) > 0 || len(spec.Capabilities.Drop) > 0) {
		caps := &core.Capabilities{}
		if len(spec.Capabilities.Add) > 0 {
			caps.Add = spec.Capabilities.Add
		}
		if len(spec.Capabilities.Drop) > 0 {
			caps.Drop = spec.Capabilities.Drop
		}
		result.Capabilities = caps
	}
	return result
}

// buildContainerSecurityContextFromLabels parses the `k8ify.securityContext.*`
// labels (or a part's labels) into a *core.SecurityContext. Returns nil when
// no labels are set (no output change).
func buildContainerSecurityContextFromLabels(labels map[string]string) *core.SecurityContext {
	spec, _ := ir.ContainerSecurityContextSpecFromLabels(labels)
	return buildContainerSecurityContext(spec)
}
