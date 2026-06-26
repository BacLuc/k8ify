package converter

import (
	"testing"

	core "k8s.io/api/core/v1"

	assertions "github.com/stretchr/testify/assert"
	"github.com/vshn/k8ify/pkg/ir"
	"github.com/vshn/k8ify/pkg/util"
)

func TestMergePodSecurityContextSpecSvcOnly(t *testing.T) {
	assert := assertions.New(t)
	svc := &ir.PodSecurityContextSpec{FSGroup: util.GetPointer(int64(1000))}
	result := mergePodSecurityContextSpec(svc, nil)
	assert.NotNil(result)
	assert.Equal(util.GetPointer(int64(1000)), result.FSGroup)
}

func TestMergePodSecurityContextSpecDefOnly(t *testing.T) {
	assert := assertions.New(t)
	mismatch := core.FSGroupChangeOnRootMismatch
	def := &ir.PodSecurityContextSpec{FSGroupChangePolicy: &mismatch}
	result := mergePodSecurityContextSpec(nil, def)
	assert.NotNil(result)
	assert.Equal(&mismatch, result.FSGroupChangePolicy)
}

func TestMergePodSecurityContextSpecFieldLevelMerge(t *testing.T) {
	assert := assertions.New(t)
	mismatch := core.FSGroupChangeOnRootMismatch
	svc := &ir.PodSecurityContextSpec{FSGroup: util.GetPointer(int64(1000))}
	def := &ir.PodSecurityContextSpec{FSGroupChangePolicy: &mismatch}
	result := mergePodSecurityContextSpec(svc, def)
	assert.NotNil(result)
	assert.Equal(util.GetPointer(int64(1000)), result.FSGroup)
	assert.Equal(&mismatch, result.FSGroupChangePolicy)
}

func TestMergePodSecurityContextSpecSvcOverridesPerField(t *testing.T) {
	assert := assertions.New(t)
	always := core.FSGroupChangeAlways
	mismatch := core.FSGroupChangeOnRootMismatch
	svc := &ir.PodSecurityContextSpec{FSGroupChangePolicy: &always}
	def := &ir.PodSecurityContextSpec{FSGroupChangePolicy: &mismatch}
	result := mergePodSecurityContextSpec(svc, def)
	assert.Equal(&always, result.FSGroupChangePolicy)
}

func TestMergePodSecurityContextSpecAllNil(t *testing.T) {
	assert := assertions.New(t)
	assert.Nil(mergePodSecurityContextSpec(nil, nil))
	assert.Nil(mergePodSecurityContextSpec(&ir.PodSecurityContextSpec{}, &ir.PodSecurityContextSpec{}))
}

func TestBuildContainerSecurityContextNil(t *testing.T) {
	assert := assertions.New(t)
	assert.Nil(buildContainerSecurityContext(nil))
}

func TestBuildContainerSecurityContextFull(t *testing.T) {
	assert := assertions.New(t)
	rd := core.SeccompProfileTypeRuntimeDefault
	spec := &ir.ContainerSecurityContextSpec{
		RunAsUser:                util.GetPointer(int64(1000)),
		RunAsGroup:               util.GetPointer(int64(2000)),
		RunAsNonRoot:             util.GetPointer(true),
		ReadOnlyRootFilesystem:   util.GetPointer(true),
		AllowPrivilegeEscalation: util.GetPointer(false),
		Privileged:               util.GetPointer(false),
		SeccompProfile:           &core.SeccompProfile{Type: rd},
		SELinuxOptions:           &core.SELinuxOptions{Type: "spc_t"},
		Capabilities:             &core.Capabilities{Add: []core.Capability{"NET_BIND_SERVICE", "SYS_PTRACE"}, Drop: []core.Capability{"ALL"}},
	}
	result := buildContainerSecurityContext(spec)
	assert.NotNil(result)
	assert.Equal(util.GetPointer(int64(1000)), result.RunAsUser)
	assert.Equal(util.GetPointer(int64(2000)), result.RunAsGroup)
	assert.Equal(util.GetPointer(true), result.RunAsNonRoot)
	assert.Equal(util.GetPointer(true), result.ReadOnlyRootFilesystem)
	assert.Equal(util.GetPointer(false), result.AllowPrivilegeEscalation)
	assert.Equal(util.GetPointer(false), result.Privileged)
	assert.Equal(&core.SeccompProfile{Type: rd}, result.SeccompProfile)
	assert.Equal(&core.SELinuxOptions{Type: "spc_t"}, result.SELinuxOptions)
	assert.Equal([]core.Capability{"NET_BIND_SERVICE", "SYS_PTRACE"}, result.Capabilities.Add)
	assert.Equal([]core.Capability{"ALL"}, result.Capabilities.Drop)
}

func TestBuildContainerSecurityContextCapabilitiesOnly(t *testing.T) {
	assert := assertions.New(t)
	spec := &ir.ContainerSecurityContextSpec{
		Capabilities: &core.Capabilities{Drop: []core.Capability{"ALL"}},
	}
	result := buildContainerSecurityContext(spec)
	assert.NotNil(result)
	assert.Nil(result.RunAsUser)
	assert.NotNil(result.Capabilities)
	assert.Equal([]core.Capability{"ALL"}, result.Capabilities.Drop)
	assert.Nil(result.Capabilities.Add)
}
