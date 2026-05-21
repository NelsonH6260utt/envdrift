package envparser

import "os"

// FromEnvironment reads the current process environment and returns an EnvMap.
// If keys is non-empty, only those keys are included.
func FromEnvironment(keys []string) EnvMap {
	envMap := make(EnvMap)

	if len(keys) == 0 {
		for _, entry := range os.Environ() {
			key, value, err := parseLine(entry)
			if err == nil {
				envMap[key] = value
			}
		}
		return envMap
	}

	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			envMap[k] = v
		}
	}
	return envMap
}

// Keys returns the sorted list of keys present in the EnvMap.
func (e EnvMap) Keys() []string {
	keys := make([]string, 0, len(e))
	for k := range e {
		keys = append(keys, k)
	}
	return keys
}

// Merge combines two EnvMaps; values from override take precedence.
func Merge(base, override EnvMap) EnvMap {
	result := make(EnvMap, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// Subset returns a new EnvMap containing only the specified keys.
func (e EnvMap) Subset(keys []string) EnvMap {
	sub := make(EnvMap, len(keys))
	for _, k := range keys {
		if v, ok := e[k]; ok {
			sub[k] = v
		}
	}
	return sub
}
