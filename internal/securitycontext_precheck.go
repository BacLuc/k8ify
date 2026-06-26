package internal

import "fmt"

// PartPodSecurityContextError indicates that a k8ify.partOf (part) service
// declared a parent-only k8ify.podSecurityContext.* label.
type PartPodSecurityContextError struct {
	Part string
}

func (e *PartPodSecurityContextError) Error() string {
	return fmt.Sprintf("service %q is a k8ify.partOf part; k8ify.podSecurityContext.* labels are only allowed on parent services", e.Part)
}

// TargetCfgContainerLevelKeyError indicates that a container-level-only key was
// configured under the pod-level x-targetCfg.securityContext default.
type TargetCfgContainerLevelKeyError struct {
	Key string
}

func (e *TargetCfgContainerLevelKeyError) Error() string {
	return fmt.Sprintf("x-targetCfg.securityContext.%s is a container-level key and is not allowed in the pod-level cluster-wide default", e.Key)
}
