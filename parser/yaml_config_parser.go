package parser

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"
)

// YamlConfigParser provides methods to handle YAML-formatted configuration files.
// It implements the ConfigParser interface.
type YamlConfigParser struct{}

// UpsertConfig merges a new configuration map into an old one.
// It adds new key-value pairs and updates existing ones.
// The result is returned as a YAML-formatted string.
func (c *YamlConfigParser) UpsertConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		oldConfMap[key] = val
	}
	return writeYamlConfigToStr(oldConfMap)
}

// UpdateConfig updates an existing configuration map with new values.
// Only keys that already exist in the old map are updated.
// The result is returned as a YAML-formatted string.
func (c *YamlConfigParser) UpdateConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		if _, ok := oldConfMap[key]; ok {
			oldConfMap[key] = val
		}
	}
	return writeYamlConfigToStr(oldConfMap)
}

// InsertConfig inserts new key-value pairs into a configuration map.
// Only keys that do not already exist in the old map are added.
// The result is returned as a YAML-formatted string.
func (c *YamlConfigParser) InsertConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		if _, ok := oldConfMap[key]; !ok {
			oldConfMap[key] = val
		}
	}
	return writeYamlConfigToStr(oldConfMap)
}

// Parse converts a raw YAML-formatted string into a flat key-value map.
// It uses dot notation for nested structures (e.g., "parent.child.key").
func (c *YamlConfigParser) Parse(rawConf string) (map[string]string, error) {
	var yamlData any
	err := yaml.Unmarshal([]byte(rawConf), &yamlData)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	flattenYaml(yamlData, "", result)

	return result, nil
}

// flattenYaml recursively flattens a YAML structure into a flat map with dot-separated keys.
// This is a helper function for Parse.
func flattenYaml(data any, prefix string, result map[string]string) {
	// Treat explicit nil as empty string to avoid "<nil>" output
	if data == nil {
		result[prefix] = ""
		return
	}

	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + "." + key
			}
			flattenYaml(value, newKey, result)
		}
	case map[any]any:
		for key, value := range v {
			keyStr := fmt.Sprintf("%v", key)
			newKey := keyStr
			if prefix != "" {
				newKey = prefix + "." + keyStr
			}
			flattenYaml(value, newKey, result)
		}
	case []any:
		for i, value := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			flattenYaml(value, newKey, result)
		}
	default:
		result[prefix] = fmt.Sprintf("%v", v)
	}
}

// writeYamlConfigToStr converts a flat key-value map back to a YAML-formatted string.
// It reconstructs the nested structure from the dot-separated keys.
func writeYamlConfigToStr(confMap map[string]string) string {
	// Build nested structure from flat keys
	yamlData := make(map[string]any)

	// Sort keys to ensure consistent output
	keys := make([]string, 0, len(confMap))
	for k := range confMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := confMap[key]
		if value == IgnoreValue {
			continue // Skip ignored values
		}
		setNestedValue(yamlData, key, value)
	}

	// Convert back to YAML
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return ""
	}
	return string(yamlBytes)
}

// setNestedValue sets a value in a nested map structure using a dot-separated key.
// This is a helper function for writeYamlConfigToStr.
func setNestedValue(data map[string]any, key string, value string) {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		// Handle array indices
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			arrayKey := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			// Ensure the array exists
			if _, ok := current[arrayKey]; !ok {
				current[arrayKey] = make([]any, 0)
			}

			// Convert to slice if needed
			if reflect.TypeOf(current[arrayKey]).Kind() != reflect.Slice {
				current[arrayKey] = make([]any, 0)
			}

			// Handle array element (simplified for basic cases)
			if i == len(parts)-1 {
				// This is the final part, set the value
				arr := current[arrayKey].([]any)
				// Extend array if needed and set value
				for len(arr) <= parseIndex(indexStr) {
					arr = append(arr, nil)
				}
				arr[parseIndex(indexStr)] = parseValue(value)
				current[arrayKey] = arr
			}
			continue
		}

		if i == len(parts)-1 {
			// Last part, set the value
			current[part] = parseValue(value)
		} else {
			// Intermediate part, ensure it's a map
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]any)
			}
			if nested, ok := current[part].(map[string]any); ok {
				current = nested
			} else {
				// If it's not a map, create a new one
				current[part] = make(map[string]any)
				current = current[part].(map[string]any)
			}
		}
	}
}

// parseValue attempts to parse a string value to an appropriate type (bool, int, float, or string).
// This is a helper function for setNestedValue.
func parseValue(value string) any {
	// Try to parse as boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Try to parse as number
	if strings.Contains(value, ".") {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	} else {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}

	// Return as string
	return value
}

// parseIndex parses an array index from a string.
// This is a helper function for setNestedValue.
func parseIndex(indexStr string) int {
	if i, err := strconv.Atoi(indexStr); err == nil {
		return i
	}
	return 0
}
