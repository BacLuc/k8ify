package util_test

import (
	prometheusTypes "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/k8ify/pkg/util"
)

func TestSubConfigEmpty(t *testing.T) {
	actual := util.SubConfig(
		map[string]string{},
		"myapp", "root")
	expected := map[string]string{}
	assert.Equal(t, expected, actual)
}

func TestSubConfigRoot(t *testing.T) {
	actual := util.SubConfig(
		map[string]string{"myapp": "value"},
		"myapp", "root")
	expected := map[string]string{"root": "value"}
	assert.Equal(t, expected, actual)
}

func TestSubConfigRootAnd(t *testing.T) {
	util.SubConfig(
		map[string]string{
			"myapp":      "value",
			"myapp.root": "other",
		},
		"myapp", "root")

	// This is undefined behavior
}

func TestSubConfig(t *testing.T) {
	actual := util.SubConfig(
		map[string]string{
			"myapp.one":   "foo",
			"myapp.two":   "bar",
			"myapp.three": "baz",
		},
		"myapp", "root")
	expected := map[string]string{
		"one":   "foo",
		"two":   "bar",
		"three": "baz",
	}
	assert.Equal(t, expected, actual)
}

func TestConfigGetInt32(t *testing.T) {
	config := map[string]string{
		"num": "1",
		"str": "foo",
	}

	assert.Equal(t, int32(1), util.ConfigGetInt32(config, "num", 99))
	assert.Equal(t, int32(88), util.ConfigGetInt32(config, "str", 88))
	assert.Equal(t, int32(77), util.ConfigGetInt32(config, "zzz", 77))
}

func TestIsTruthy(t *testing.T) {
	assert := assert.New(t)
	assert.True(util.IsTruthy("true"))
	assert.True(util.IsTruthy("True"))
	assert.True(util.IsTruthy("TRUE"))
	assert.True(util.IsTruthy("yes"))
	assert.True(util.IsTruthy("YES"))
	assert.True(util.IsTruthy("1"))

	assert.False(util.IsTruthy("false"))
	assert.False(util.IsTruthy("False"))
	assert.False(util.IsTruthy("FALSE"))
	assert.False(util.IsTruthy("no"))
	assert.False(util.IsTruthy("NO"))
	assert.False(util.IsTruthy("0"))
}

func TestNullableBool(t *testing.T) {
	assert := assert.New(t)

	cases := []TestCase[string, *bool]{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:     "blank",
			input:    " \t",
			expected: nil,
		},
		{
			name:     "true",
			input:    "true",
			expected: &trueToPointTo,
		},
		{
			name:     "false",
			input:    "false",
			expected: &falseToPointTo,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.FilterBlankBool(&tc.input)

			assert.Equal(tc.expected, actual, "NullableBool should be %v", tc.input, tc.expected)
		})
	}
}

func TestServiceMonitorConfig(t *testing.T) {
	assert := assert.New(t)
	type LabelMap map[string]string

	cases := []TestCase[LabelMap, *util.ServiceMonitorConfig]{
		{
			name:     "ServiceMonitorConfig_nothing_set",
			input:    LabelMap{},
			expected: &util.ServiceMonitorConfig{},
		},
		{
			name:  "ServiceMonitorConfig_enabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor": "true"},
			expected: &util.ServiceMonitorConfig{
				Enabled: true,
			},
		},
		{
			name:  "ServiceMonitorConfig_disabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor": "false"},
			expected: &util.ServiceMonitorConfig{
				Enabled: false,
			},
		},
		{
			name: "ServiceMonitorConfig_values_set",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor":               "true",
				"k8ify.prometheus.serviceMonitor.interval":      monitorInterval,
				"k8ify.prometheus.serviceMonitor.path":          monitorPath,
				"k8ify.prometheus.serviceMonitor.scheme":        monitorScheme,
				"k8ify.prometheus.serviceMonitor.endpoint.name": monitorEndpointName,
			},
			expected: &util.ServiceMonitorConfig{
				Enabled:      true,
				Interval:     &monitorInterval,
				Path:         &monitorPath,
				Scheme:       &monitorScheme,
				EndpointName: &monitorEndpointName,
			},
		},
		{
			name: "ServiceMonitorConfig_empty_strings",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor":               "",
				"k8ify.prometheus.serviceMonitor.interval":      "",
				"k8ify.prometheus.serviceMonitor.path":          "",
				"k8ify.prometheus.serviceMonitor.scheme":        "",
				"k8ify.prometheus.serviceMonitor.endpoint.name": "",
			},
			expected: &util.ServiceMonitorConfig{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.ServiceMonitorConfigPointer(tc.input)

			assert.Equal(tc.expected, actual, "ServiceMonitorConfigPointer(%v) should return %v", tc.input, tc.expected)
		})
	}
}

func TestServiceMonitorBasicAuthConfig(t *testing.T) {
	assert := assert.New(t)
	type LabelMap map[string]string

	cases := []TestCase[LabelMap, *util.ServiceMonitorBasicAuthConfig]{
		{
			name:     "BasicAuthConfig_nothing_set",
			input:    LabelMap{},
			expected: &util.ServiceMonitorBasicAuthConfig{},
		},
		{
			name:  "BasicAuthConfig_enabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor.endpoint.basicAuth": "true"},
			expected: &util.ServiceMonitorBasicAuthConfig{
				Enabled:  true,
				Username: "",
				Password: "",
			},
		},
		{
			name:  "BasicAuthConfig_disabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor.endpoint.basicAuth": "false"},
			expected: &util.ServiceMonitorBasicAuthConfig{
				Enabled: false,
			},
		},
		{
			name: "BasicAuthConfig_values_set",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth":          "true",
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth.username": monitorUsername,
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth.password": monitorPassword,
			},
			expected: &util.ServiceMonitorBasicAuthConfig{
				Enabled:  true,
				Username: monitorUsername,
				Password: monitorPassword,
			},
		},
		{
			name: "BasicAuthConfig_empty_strings",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth":          "",
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth.username": "",
				"k8ify.prometheus.serviceMonitor.endpoint.basicAuth.password": "",
			},
			expected: &util.ServiceMonitorBasicAuthConfig{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.ServiceMonitorBasicAuthConfigPointer(tc.input)

			assert.Equal(tc.expected, actual, "BasicAuthConfigPointer(%v) should return %v", tc.input, tc.expected)
		})
	}
}

func TestServiceMonitorTlsConfig(t *testing.T) {
	assert := assert.New(t)
	type LabelMap map[string]string

	cases := []TestCase[LabelMap, *util.ServiceMonitorTlsConfig]{
		{
			name:     "TlsConfig_nothing_set",
			input:    LabelMap{},
			expected: &util.ServiceMonitorTlsConfig{},
		},
		{
			name:  "TlsConfig_enabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig": "true"},
			expected: &util.ServiceMonitorTlsConfig{
				Enabled: true,
			},
		},
		{
			name:  "TlsConfig_disabled",
			input: LabelMap{"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig": "false"},
			expected: &util.ServiceMonitorTlsConfig{
				Enabled: false,
			},
		},
		{
			name: "TlsConfig_values_set",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig":                    "true",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.ca":                 monitorCa,
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.cert":               monitorCert,
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.keySecretValue":     monitorKeySecretValue,
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.insecureSkipVerify": "true",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.maxVersion":         monitorTlsMaxVersion,
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.minVersion":         monitorTlsMinVersion,
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.serverName":         monitorServerName,
			},
			expected: &util.ServiceMonitorTlsConfig{
				Enabled:            true,
				Ca:                 &monitorCa,
				Cert:               &monitorCert,
				KeySecretValue:     &monitorKeySecretValue,
				InsecureSkipVerify: &trueToPointTo,
				MaxVersion:         &monitorTlsMaxVersion,
				MinVersion:         &monitorTlsMinVersion,
				ServerName:         &monitorServerName,
			},
		},
		{
			name: "TlsConfig_empty_strings",
			input: LabelMap{
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig":                    "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.ca":                 "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.cert":               "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.keySecretValue":     "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.insecureSkipVerify": "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.maxVersion":         "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.minVersion":         "",
				"k8ify.prometheus.serviceMonitor.endpoint.tlsConfig.serverName":         "",
			},
			expected: &util.ServiceMonitorTlsConfig{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := util.ServiceMonitorTlsConfigPointer(tc.input)

			assert.Equal(tc.expected, actual, "BasicAuthConfigPointer(%v) should return %v", tc.input, tc.expected)
		})
	}
}

type TestCase[InParam any, OutParam any] struct {
	name     string
	input    InParam
	expected OutParam
}

var (
	falseToPointTo        = false
	monitorCa             = "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIUL8fmlL3Z1OSjE+9GHNrCuDGWKZgwCgYIKoZIzj0EAwIw\nGDEWMBQGA1UEAwwNTXkgTWluaW1hbCBDQTAeFw0yNTA3MTUxMzUyMTJaFw0yNjA3\nMTAxMzUyMTJaMBgxFjAUBgNVBAMMDU15IE1pbmltYWwgQ0EwWTATBgcqhkjOPQIB\nBggqhkjOPQMBBwNCAAQ6GrfF/1dVy3v97b+c6ZWRBAmdlBNV3qxfhdWS6KIwMvCr\nDiRUhXOpcLA49HjLX9RfDpxyI8Nz/Nv12bMg5f3go1MwUTAdBgNVHQ4EFgQU7Zcx\nnhcTn8t5cdCumGg7IKL39YwwHwYDVR0jBBgwFoAU7ZcxnhcTn8t5cdCumGg7IKL3\n9YwwDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiBeHxk5JKc1JpKF\nTZU6u6Yo4ozduWSQxIH6jSzh7BOCTAIhAMDNTO4ilY+DAna/udskuXMcjsfI0kQY\nU95t8zPBdnxh\n-----END CERTIFICATE-----\n"
	monitorCert           = "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIUAfxOSXWbSYYnhBSQGezqo+D6ia4wCgYIKoZIzj0EAwIw\nGDEWMBQGA1UEAwwNTXkgTWluaW1hbCBDQTAeFw0yNTA3MTUxMzU3MzBaFw0yNjA3\nMTAxMzU3MzBaMBgxFjAUBgNVBAMMDU15IE1pbmltYWwgQ0EwWTATBgcqhkjOPQIB\nBggqhkjOPQMBBwNCAATxEpwQy5oTno/HH+w9lUMsQxDZWFADZzt2xuI1Q33/TsBV\nKCwmZv3ywDwP1n2rHSoR7pZrQwUvNx/gyAobTPeDo1MwUTAdBgNVHQ4EFgQUjOmk\n2Q1r4qrwPIEjnWUlcyOAmTUwHwYDVR0jBBgwFoAUjOmk2Q1r4qrwPIEjnWUlcyOA\nmTUwDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiA7A3wxscDg/3rE\nqz7dR6899fxypP+nTwVw1M9SYgwmpAIhAO72sDxX86Y6Qikv1TCEQpO5t43clkoo\nekgUTlxCHY0H\n-----END CERTIFICATE-----\n"
	monitorEndpointName   = "default"
	monitorInterval       = "30s"
	monitorKeySecretValue = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIHA0lQcWCa/He3w/MDaQS1c/YJte4mx3Bg9gzAr4P35BoAoGCCqGSM49\nAwEHoUQDQgAEc4ivr46eO4DOxArTOGP+5sxjFHDQpF02tRnuQBa9R433GDOSvdqb\nTEmIlxovk6eif+/2yLxFIsaA8aXaMbH+wQ==\n-----END EC PRIVATE KEY-----"
	monitorPassword       = "mypassword"
	monitorPath           = "/actuator/health"
	monitorScheme         = "http"
	monitorServerName     = "service.svc"
	monitorTlsMaxVersion  = string(prometheusTypes.TLSVersion13)
	monitorTlsMinVersion  = string(prometheusTypes.TLSVersion10)
	monitorUsername       = "myuser"
	trueToPointTo         = true
)
