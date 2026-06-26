package ir

import (
	"testing"

	core "k8s.io/api/core/v1"

	assertions "github.com/stretchr/testify/assert"
	"github.com/vshn/k8ify/pkg/util"
)

func TestPodSecurityContextDefaultAbsent(t *testing.T) {
	assert := assertions.New(t)
	cfg := TargetCfg{}
	spec, errs := cfg.PodSecurityContextDefault()
	assert.Nil(spec)
	assert.Nil(errs)
}

func TestPodSecurityContextDefaultFull(t *testing.T) {
	trace := assertions.New(t)
	cfg := TargetCfg{
		"securityContext": map[string]interface{}{
			"fsGroup":             int64(1000),
			"fsGroupChangePolicy": "OnRootMismatch",
			"seLinuxOptions": map[string]interface{}{
				"level": "s0:c100,c200",
			},
			"supplementalGroups": "500,500,200",
			"seccompProfile": map[string]interface{}{
				"type": "RuntimeDefault",
			},
			"runAsUser":    int64(1000),
			"runAsGroup":   int64(2000),
			"runAsNonRoot": true,
		},
	}
	spec, errs := cfg.PodSecurityContextDefault()
	trace.Nil(errs)
	trace.NotNil(spec)
	trace.Equal(util.GetPointer(int64(1000)), spec.FSGroup)
	mismatch := core.FSGroupChangeOnRootMismatch
	trace.Equal(&mismatch, spec.FSGroupChangePolicy)
	trace.Equal(&core.SELinuxOptions{Level: "s0:c100,c200"}, spec.SELinuxOptions)
	trace.Equal([]int64{200, 500}, spec.SupplementalGroups)
	rd := core.SeccompProfileTypeRuntimeDefault
	trace.Equal(&core.SeccompProfile{Type: rd}, spec.SeccompProfile)
	trace.Equal(util.GetPointer(int64(1000)), spec.RunAsUser)
	trace.Equal(util.GetPointer(int64(2000)), spec.RunAsGroup)
	trace.Equal(util.GetPointer(true), spec.RunAsNonRoot)
}

func TestPodSecurityContextDefaultInvalidType(t *testing.T) {
	assert := assertions.New(t)
	cfg := TargetCfg{
		"securityContext": "not-a-map",
	}
	spec, errs := cfg.PodSecurityContextDefault()
	assert.Nil(spec)
	assert.NotEmpty(errs)
}

func TestPodSecurityContextSpecFromLabelsEmpty(t *testing.T) {
	assert := assertions.New(t)
	spec, errs := PodSecurityContextSpecFromLabels(map[string]string{})
	assert.Nil(spec)
	assert.Nil(errs)
}

func TestPodSecurityContextSpecFromLabelsFull(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{
		"k8ify.podSecurityContext.fsGroup":              "1000",
		"k8ify.podSecurityContext.fsGroupChangePolicy":  "OnRootMismatch",
		"k8ify.podSecurityContext.seLinuxOptions.level": "s0:c100,c200",
		"k8ify.podSecurityContext.runAsNonRoot":         "true",
	}
	spec, errs := PodSecurityContextSpecFromLabels(labels)
	assert.Nil(errs)
	assert.NotNil(spec)
	assert.Equal(util.GetPointer(int64(1000)), spec.FSGroup)
	mismatch := core.FSGroupChangeOnRootMismatch
	assert.Equal(&mismatch, spec.FSGroupChangePolicy)
	assert.Equal(&core.SELinuxOptions{Level: "s0:c100,c200"}, spec.SELinuxOptions)
	assert.Equal(util.GetPointer(true), spec.RunAsNonRoot)
}

func TestPodSecurityContextSpecFromLabelsInvalidEnum(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{
		"k8ify.podSecurityContext.fsGroupChangePolicy": "Never",
	}
	spec, errs := PodSecurityContextSpecFromLabels(labels)
	assert.Nil(spec)
	assert.NotEmpty(errs)
}

func TestContainerSecurityContextSpecFromLabelsEmpty(t *testing.T) {
	assert := assertions.New(t)
	spec, errs := ContainerSecurityContextSpecFromLabels(map[string]string{})
	assert.Nil(spec)
	assert.Nil(errs)
}

func TestContainerSecurityContextSpecFromLabelsCapabilitiesSorting(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{
		"k8ify.securityContext.capabilities.add":  "SYS_PTRACE,net_bind_service",
		"k8ify.securityContext.capabilities.drop": "ALL",
	}
	spec, errs := ContainerSecurityContextSpecFromLabels(labels)
	assert.Nil(errs)
	assert.NotNil(spec)
	assert.Equal([]core.Capability{"NET_BIND_SERVICE", "SYS_PTRACE"}, spec.Capabilities.Add)
	assert.Equal([]core.Capability{"ALL"}, spec.Capabilities.Drop)
}

func TestContainerSecurityContextSpecFromLabelsSeccompLocalhostWithoutProfile(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{
		"k8ify.securityContext.seccompProfile.type": "Localhost",
	}
	spec, errs := ContainerSecurityContextSpecFromLabels(labels)
	assert.Nil(spec)
	assert.NotEmpty(errs)
}
