package internal

import (
	"testing"

	assertions "github.com/stretchr/testify/assert"
	"github.com/vshn/k8ify/pkg/ir"
)

func targetCfgWithSecurityContext(sec map[string]interface{}) ir.TargetCfg {
	return ir.TargetCfg{"securityContext": sec}
}

func TestValidatePodSecurityContextLabelsPartRejected(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{"k8ify.podSecurityContext.fsGroup": "1000"}
	errs := validatePodSecurityContextLabels("my-part", labels, true)
	assert.Len(errs, 1)
	assert.Contains(errs[0].Error(), "parent")
}

func TestValidatePodSecurityContextLabelsPartContainerAllowed(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{"k8ify.securityContext.runAsUser": "1000"}
	errs := validatePodSecurityContextLabels("my-part", labels, true)
	assert.Empty(errs)
}

func TestValidatePodSecurityContextLabelsParentInvalidEnum(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{"k8ify.podSecurityContext.fsGroupChangePolicy": "Never"}
	errs := validatePodSecurityContextLabels("parent", labels, false)
	assert.NotEmpty(errs)
}

func TestValidateContainerSecurityContextLabelsSeccompError(t *testing.T) {
	assert := assertions.New(t)
	labels := map[string]string{"k8ify.securityContext.seccompProfile.type": "Localhost"}
	errs := validateContainerSecurityContextLabels(labels)
	assert.NotEmpty(errs)
}

func TestValidateTargetCfgSecurityContextContainerLevelKey(t *testing.T) {
	assert := assertions.New(t)
	cfg := targetCfgWithSecurityContext(map[string]interface{}{
		"capabilities": map[string]interface{}{"add": "SYS_PTRACE"},
	})
	errs := validateTargetCfgSecurityContext(cfg)
	assert.NotEmpty(errs)
	found := false
	for _, err := range errs {
		if err.Error() == (&TargetCfgContainerLevelKeyError{Key: "capabilities"}).Error() {
			found = true
		}
	}
	assert.True(found)
}

func TestValidateTargetCfgSecurityContextValidPodLevel(t *testing.T) {
	assert := assertions.New(t)
	cfg := targetCfgWithSecurityContext(map[string]interface{}{
		"fsGroupChangePolicy": "OnRootMismatch",
	})
	errs := validateTargetCfgSecurityContext(cfg)
	assert.Empty(errs)
}

func TestValidateTargetCfgSecurityContextSeccompLocalhostWithoutProfile(t *testing.T) {
	assert := assertions.New(t)
	cfg := targetCfgWithSecurityContext(map[string]interface{}{
		"seccompProfile": map[string]interface{}{"type": "Localhost"},
	})
	errs := validateTargetCfgSecurityContext(cfg)
	assert.NotEmpty(errs)
}
