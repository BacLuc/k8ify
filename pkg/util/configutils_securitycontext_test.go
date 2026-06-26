package util_test

import (
	"testing"

	core "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/k8ify/pkg/util"
)

func TestParseInt64(t *testing.T) {
	assert := assert.New(t)
	type in struct {
		config map[string]string
		key    string
	}
	cases := []struct {
		name        string
		input       in
		expected    *int64
		expectError bool
	}{
		{name: "absent", input: in{map[string]string{}, "x"}, expected: nil},
		{name: "blank", input: in{map[string]string{"x": "  "}, "x"}, expected: nil},
		{name: "valid", input: in{map[string]string{"x": "123"}, "x"}, expected: util.GetPointer(int64(123))},
		{name: "invalid", input: in{map[string]string{"x": "abc"}, "x"}, expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := util.ParseInt64(tc.input.config, tc.input.key)
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseBoolPtr(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		name     string
		input    map[string]string
		expected *bool
	}{
		{name: "absent", input: map[string]string{}, expected: nil},
		{name: "blank", input: map[string]string{"x": "  "}, expected: nil},
		{name: "true", input: map[string]string{"x": "true"}, expected: util.GetPointer(true)},
		{name: "yes", input: map[string]string{"x": "YES"}, expected: util.GetPointer(true)},
		{name: "one", input: map[string]string{"x": "1"}, expected: util.GetPointer(true)},
		{name: "false", input: map[string]string{"x": "false"}, expected: util.GetPointer(false)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.ParseBoolPtr(tc.input, "x")
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseFSGroupChangePolicy(t *testing.T) {
	assert := assert.New(t)
	always := core.FSGroupChangeAlways
	mismatch := core.FSGroupChangeOnRootMismatch
	cases := []struct {
		name        string
		input       map[string]string
		expected    *core.PodFSGroupChangePolicy
		expectError bool
	}{
		{name: "absent", input: map[string]string{}, expected: nil},
		{name: "blank", input: map[string]string{"x": "  "}, expected: nil},
		{name: "Always", input: map[string]string{"policy": "Always"}, expected: &always},
		{name: "always", input: map[string]string{"policy": "always"}, expected: &always},
		{name: "OnRootMismatch", input: map[string]string{"policy": "OnRootMismatch"}, expected: &mismatch},
		{name: "invalid", input: map[string]string{"policy": "Never"}, expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := util.ParseFSGroupChangePolicy(tc.input, "policy")
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseSeccompProfile(t *testing.T) {
	assert := assert.New(t)
	runtimeDefault := core.SeccompProfileTypeRuntimeDefault
	localhost := core.SeccompProfileTypeLocalhost
	profile := "profiles/my.json"
	cases := []struct {
		name        string
		input       map[string]string
		expected    *core.SeccompProfile
		expectError bool
	}{
		{name: "absent", input: map[string]string{}, expected: nil},
		{name: "blank", input: map[string]string{"type": ""}, expected: nil},
		{name: "RuntimeDefault", input: map[string]string{"type": "RuntimeDefault"}, expected: &core.SeccompProfile{Type: runtimeDefault}},
		{name: "Localhost_with_profile", input: map[string]string{"type": "Localhost", "localhostProfile": "profiles/my.json"}, expected: &core.SeccompProfile{Type: localhost, LocalhostProfile: &profile}},
		{name: "Localhost_without_profile", input: map[string]string{"type": "Localhost"}, expectError: true},
		{name: "Localhost_blank_profile", input: map[string]string{"type": "Localhost", "localhostProfile": "  "}, expectError: true},
		{name: "profile_without_localhost", input: map[string]string{"type": "RuntimeDefault", "localhostProfile": "profiles/my.json"}, expectError: true},
		{name: "bad_type", input: map[string]string{"type": "Unconfined"}, expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := util.ParseSeccompProfile(tc.input)
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseCapabilities(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		name        string
		input       map[string]string
		expected    []core.Capability
		expectError bool
	}{
		{name: "absent", input: map[string]string{}, expected: nil},
		{name: "blank", input: map[string]string{"add": "  "}, expected: nil},
		{name: "two_sorted_upper", input: map[string]string{"add": "SYS_PTRACE,net_bind_service"}, expected: []core.Capability{"NET_BIND_SERVICE", "SYS_PTRACE"}},
		{name: "duplicate_dedupe", input: map[string]string{"add": "SYS_PTRACE,SYS_PTRACE"}, expected: []core.Capability{"SYS_PTRACE"}},
		{name: "empty_entry", input: map[string]string{"add": "SYS_PTRACE,"}, expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := util.ParseCapabilities(tc.input, "add")
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseInt64List(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		name        string
		input       map[string]string
		expected    []int64
		expectError bool
	}{
		{name: "absent", input: map[string]string{}, expected: nil},
		{name: "blank", input: map[string]string{"groups": "  "}, expected: nil},
		{name: "sorted_deduped", input: map[string]string{"groups": "100,2,2,50"}, expected: []int64{2, 50, 100}},
		{name: "empty_entry", input: map[string]string{"groups": "100,,"}, expectError: true},
		{name: "non_int", input: map[string]string{"groups": "100,abc"}, expectError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := util.ParseInt64List(tc.input, "groups")
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestParseSELinuxOptions(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		name     string
		input    map[string]string
		expected *core.SELinuxOptions
	}{
		{name: "all_blank", input: map[string]string{}, expected: nil},
		{name: "only_level_verbatim", input: map[string]string{"level": "s0:c1,c2"}, expected: &core.SELinuxOptions{Level: "s0:c1,c2"}},
		{name: "full_set", input: map[string]string{"user": "u", "role": "r", "type": "t", "level": "s0:c1,c2"}, expected: &core.SELinuxOptions{User: "u", Role: "r", Type: "t", Level: "s0:c1,c2"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.ParseSELinuxOptions(tc.input)
			assert.Equal(tc.expected, actual)
		})
	}
}

func TestNormalizeInterfaceConfig(t *testing.T) {
	assert := assert.New(t)
	input := map[string]interface{}{
		"flat": "a",
		"nested": map[string]interface{}{
			"deep": "b",
			"deeper": map[string]interface{}{
				"leaf": "c",
			},
		},
		"num":  int64(42),
		"flag": true,
		"list": []interface{}{"x", "y", "z"},
		"nul":  nil,
	}
	expected := map[string]string{
		"flat":               "a",
		"nested.deep":        "b",
		"nested.deeper.leaf": "c",
		"num":                "42",
		"flag":               "true",
		"list":               "x,y,z",
	}
	assert.Equal(expected, util.NormalizeInterfaceConfig(input, ""))
}
