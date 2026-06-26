package util

import (
	"fmt"
	"maps"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"

	"github.com/docker/go-units"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	reTrue = regexp.MustCompile("(?i)^true|yes|1$")
)

// SubConfig extracts the keys that start with a given prefix from a given
// config map
//
// If there is a key that is EQUAL to the prefix, the entry identified by
// `defaultKey` will be updated.
//
// If there is a key that is equal to the prefix, AND an entry corresponding to
// `defaultKey`, the behavior is undefined!
func SubConfig(config map[string]string, prefix string, defaultKey string) map[string]string {
	subConfig := make(map[string]string)
	for key, value := range config {
		if key == prefix {
			subConfig[defaultKey] = value
		}
		if strings.HasPrefix(key, prefix+".") && len(key) > (len(prefix)+1) {
			subKey := key[len(prefix)+1:]
			subConfig[subKey] = value
		}
	}
	return subConfig
}

// ConfigGetInt32 extracts the int32 value from the entry with the given key
//
// Returns `defaultValue` if either the entry does not exist, or is not a
// numeric value.
func ConfigGetInt32(config map[string]string, key string, defaultValue int32) int32 {
	if valStr, ok := config[key]; ok {
		if valInt, err := strconv.Atoi(valStr); err == nil {
			return int32(valInt)
		}
	}
	return defaultValue
}

// IsTruthy determines whether the given string is a representation of a "true"
// state.
//
// Concretly, it currently tests for "true", "yes" or "1", ignoring character
// cases.
func IsTruthy(s string) bool {
	return reTrue.MatchString(s)
}

func GetBoolean(labels map[string]string, key string) bool {
	if val, ok := labels[key]; ok {
		return IsTruthy(val)
	}

	return false
}

func GetOptional(labels map[string]string, key string) *string {
	if val, ok := labels[key]; ok {
		return &val
	}

	return nil
}

// IsSingleton determine whether a resource (according to its labels) should be
// treated as a singleton.
func IsSingleton(labels map[string]string) bool {
	return GetBoolean(labels, "k8ify.singleton")
}

// IsShared determines whether a volume is shared between replicas
func IsShared(labels map[string]string) bool {
	return GetBoolean(labels, "k8ify.shared")
}

// StorageClass determines a storage class from a set of labels
func StorageClass(labels map[string]string) *string {
	return GetOptional(labels, "k8ify.storageClass")
}

func StorageSizeRaw(labels map[string]string) *string {
	return GetOptional(labels, "k8ify.size")
}

func Converter(labels map[string]string) *string {
	return GetOptional(labels, "k8ify.converter")
}

func PartOf(labels map[string]string) *string {
	return GetOptional(labels, "k8ify.partOf")
}

func ImagePullSecret(labels map[string]string) *string {
	return GetOptional(labels, "k8ify.imagePullSecret")
}

// StorageSize determines the requested storage size for a volume, or a
// fallback value.
func StorageSize(labels map[string]string, fallback string) resource.Quantity {
	quantity := fallback
	if q := StorageSizeRaw(labels); q != nil {
		quantity = *q
	}

	size, err := units.RAMInBytes(quantity)
	if err != nil {
		logrus.Errorf("Invalid storage size: %q\n", quantity)
		os.Exit(1)
	}

	return *resource.NewQuantity(size, resource.BinarySI)
}

func ServiceAccountName(labels map[string]string) string {
	serviceAccountName := GetOptional(labels, "k8ify.serviceAccountName")
	if serviceAccountName == nil {
		return ""
	}
	return *serviceAccountName
}

func Annotations(labels map[string]string, kind string) map[string]string {
	annotations := SubConfig(labels, "k8ify.annotations", "")
	maps.Copy(annotations, SubConfig(labels, fmt.Sprintf("k8ify.%s.annotations", kind), ""))
	delete(annotations, "")
	return annotations
}

func ServiceType(labels map[string]string, port int32) core.ServiceType {
	subConfig := SubConfig(labels, fmt.Sprintf("k8ify.exposePlain.%d", port), "")
	if len(subConfig) == 0 {
		return ""
	}
	if serviceType, ok := subConfig["type"]; ok {
		// Go does not offer a way to list values of its "ENUMs", see https://github.com/golang/go/issues/19814
		if serviceType == string(core.ServiceTypeClusterIP) {
			return core.ServiceTypeClusterIP
		}
		if serviceType == string(core.ServiceTypeLoadBalancer) {
			return core.ServiceTypeLoadBalancer
		}
		if serviceType == string(core.ServiceTypeExternalName) {
			return core.ServiceTypeExternalName
		}
		if serviceType == string(core.ServiceTypeNodePort) {
			return core.ServiceTypeNodePort
		}
	}
	return core.ServiceTypeLoadBalancer
}

func ServiceExternalTrafficPolicy(labels map[string]string, port int32) core.ServiceExternalTrafficPolicy {
	subConfig := SubConfig(labels, fmt.Sprintf("k8ify.exposePlain.%d", port), "")
	if len(subConfig) == 0 {
		return ""
	}
	if serviceType, ok := subConfig["externalTrafficPolicy"]; ok {
		// Go does not offer a way to list values of its "ENUMs", see https://github.com/golang/go/issues/19814
		if serviceType == string(core.ServiceExternalTrafficPolicyCluster) {
			return core.ServiceExternalTrafficPolicyCluster
		}
		if serviceType == string(core.ServiceExternalTrafficPolicyLocal) {
			return core.ServiceExternalTrafficPolicyLocal
		}
	}
	return core.ServiceExternalTrafficPolicyLocal
}

func ServiceHealthCheckNodePort(labels map[string]string, port int32) int32 {
	subConfig := SubConfig(labels, fmt.Sprintf("k8ify.exposePlain.%d", port), "")
	if len(subConfig) == 0 {
		return 0
	}
	healthCheckNodePort := ConfigGetInt32(subConfig, "healthCheckNodePort", 0)
	if healthCheckNodePort > 65535 || healthCheckNodePort < 0 {
		return 0
	}
	return healthCheckNodePort
}

// ParseInt64 reads an optional int64 value from config[key]. Missing or blank
// keys return (nil, nil); invalid values return an error.
func ParseInt64(config map[string]string, key string) (*int64, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("securityContext: %s must be an integer, got %q", key, value)
	}
	return &parsed, nil
}

// ParseBoolPtr reads an optional boolean pointer from config[key]. Missing or
// blank keys return nil; otherwise a *bool built via IsTruthy. No error.
func ParseBoolPtr(config map[string]string, key string) *bool {
	value, ok := config[key]
	if !ok {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	result := IsTruthy(trimmed)
	return &result
}

// ParseFSGroupChangePolicy reads an optional fsGroupChangePolicy enum value.
// Missing/blank returns (nil, nil); case-insensitive match against Always and
// OnRootMismatch; anything else returns an error.
func ParseFSGroupChangePolicy(config map[string]string, key string) (*core.PodFSGroupChangePolicy, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	switch strings.ToLower(trimmed) {
	case strings.ToLower(string(core.FSGroupChangeAlways)):
		policy := core.FSGroupChangeAlways
		return &policy, nil
	case strings.ToLower(string(core.FSGroupChangeOnRootMismatch)):
		policy := core.FSGroupChangeOnRootMismatch
		return &policy, nil
	default:
		return nil, fmt.Errorf("fsGroupChangePolicy must be Always or OnRootMismatch, got %q", value)
	}
}

// ParseSeccompProfile reads a seccomp profile from the "type" and
// "localhostProfile" keys of config. Missing type returns (nil, nil).
// `type=Localhost` requires `localhostProfile`; a `localhostProfile` set
// without `type=Localhost` is an error.
func ParseSeccompProfile(config map[string]string) (*core.SeccompProfile, error) {
	typeValue, ok := config["type"]
	if !ok {
		return nil, nil
	}
	trimmedType := strings.TrimSpace(typeValue)
	if trimmedType == "" {
		return nil, nil
	}
	switch strings.ToLower(trimmedType) {
	case strings.ToLower(string(core.SeccompProfileTypeRuntimeDefault)):
		if profileValue, ok := config["localhostProfile"]; ok && strings.TrimSpace(profileValue) != "" {
			return nil, fmt.Errorf("securityContext: seccompProfile.localhostProfile must only be set with type=Localhost, got type %q", typeValue)
		}
		st := core.SeccompProfileTypeRuntimeDefault
		return &core.SeccompProfile{Type: st}, nil
	case strings.ToLower(string(core.SeccompProfileTypeLocalhost)):
		profileValue, hasProfile := config["localhostProfile"]
		trimmedProfile := ""
		if hasProfile {
			trimmedProfile = strings.TrimSpace(profileValue)
		}
		if trimmedProfile == "" {
			return nil, fmt.Errorf("securityContext: seccompProfile.localhostProfile is required when seccompProfile.type is Localhost")
		}
		st := core.SeccompProfileTypeLocalhost
		return &core.SeccompProfile{Type: st, LocalhostProfile: &trimmedProfile}, nil
	default:
		return nil, fmt.Errorf("securityContext: seccompProfile.type must be RuntimeDefault or Localhost, got %q", typeValue)
	}
}

// ParseCapabilities reads a comma-separated capability list from config[key].
// Entries are trimmed, uppercased, deduplicated and sorted. Empty entries
// produce an error. Missing or blank values return nil.
func ParseCapabilities(config map[string]string, key string) ([]core.Capability, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	seen := make(map[string]bool)
	var result []string
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			return nil, fmt.Errorf("capabilities: empty entry in %q", value)
		}
		upper := strings.ToUpper(entry)
		if !seen[upper] {
			seen[upper] = true
			result = append(result, upper)
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	sort.Strings(result)
	capabilities := make([]core.Capability, len(result))
	for i, c := range result {
		capabilities[i] = core.Capability(c)
	}
	return capabilities, nil
}

// ParseInt64List reads a comma-separated int64 list from config[key]. Entries
// are trimmed, parsed, deduplicated and sorted. Missing/blank returns nil.
// Empty entries or non-integer values produce an error.
func ParseInt64List(config map[string]string, key string) ([]int64, error) {
	value, ok := config[key]
	if !ok {
		return nil, nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	seen := make(map[int64]bool)
	var values []int64
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			return nil, fmt.Errorf("securityContext: %s contains an empty entry in %q", key, value)
		}
		parsed, err := strconv.ParseInt(entry, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("securityContext: %s must be an integer, got %q", key, entry)
		}
		if !seen[parsed] {
			seen[parsed] = true
			values = append(values, parsed)
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values, nil
}

// ParseSELinuxOptions reads the user/role/type/level keys from config. If all
// four are blank/absent it returns nil; otherwise it returns a populated
// *SELinuxOptions. level is passed verbatim (it may contain ':' or ',').
func ParseSELinuxOptions(config map[string]string) *core.SELinuxOptions {
	user := strings.TrimSpace(config["user"])
	role := strings.TrimSpace(config["role"])
	selinuxType := strings.TrimSpace(config["type"])
	level := config["level"]
	if user == "" && role == "" && selinuxType == "" && level == "" {
		return nil
	}
	return &core.SELinuxOptions{
		User:  user,
		Role:  role,
		Type:  selinuxType,
		Level: level,
	}
}

// NormalizeInterfaceConfig recursively flattens a nested
// map[string]interface{} (as produced for x-targetCfg) into a flat
// map[string]string with dotted keys. Scalar string/int/int64/float64/bool
// values are converted via fmt.Sprint; []interface{} lists are comma-joined;
// nil values are skipped.
func NormalizeInterfaceConfig(m map[string]interface{}, prefix string) map[string]string {
	result := make(map[string]string)
	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case nil:
			continue
		case map[string]interface{}:
			for k, val := range NormalizeInterfaceConfig(v, fullKey) {
				result[k] = val
			}
		case []interface{}:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if item == nil {
					continue
				}
				parts = append(parts, fmt.Sprint(item))
			}
			result[fullKey] = strings.Join(parts, ",")
		default:
			result[fullKey] = fmt.Sprint(v)
		}
	}
	return result
}
