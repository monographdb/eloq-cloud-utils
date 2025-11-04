package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/ini.v1"
)

// IniConfigParser provides methods to handle INI-formatted configuration files.
// It implements the ConfigParser interface.
type IniConfigParser struct{}

// UpsertConfig merges a new configuration map into an old one.
// It adds new key-value pairs and updates existing ones.
// The result is returned as an INI-formatted string.
func (c *IniConfigParser) UpsertConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		oldConfMap[key] = val
	}
	return writeConfigToStr(oldConfMap)
}

// UpdateConfig updates an existing configuration map with new values.
// Only keys that already exist in the old map are updated.
// The result is returned as an INI-formatted string.
func (c *IniConfigParser) UpdateConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		if _, ok := oldConfMap[key]; ok {
			oldConfMap[key] = val
		}
	}
	return writeConfigToStr(oldConfMap)
}

// InsertConfig inserts new key-value pairs into a configuration map.
// Only keys that do not already exist in the old map are added.
// The result is returned as an INI-formatted string.
func (c *IniConfigParser) InsertConfig(oldConfMap, newConfigMap map[string]string) string {
	for key, val := range newConfigMap {
		if _, ok := ignoreFields[key]; ok {
			oldConfMap[key] = ignoreFields[key]
			continue
		}
		if _, ok := oldConfMap[key]; !ok {
			oldConfMap[key] = val
		}
	}
	return writeConfigToStr(oldConfMap)
}

// Parse converts a raw INI-formatted string into a flat key-value map.
// It uses dot notation for section and key (e.g., "section.key").
// It can handle flag-style keys (keys without values).
func (c *IniConfigParser) Parse(rawConf string) (map[string]string, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{SkipUnrecognizableLines: true}, []byte(rawConf))
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, section := range cfg.SectionStrings() {
		sectionObj, _ := cfg.GetSection(section)
		keyNames := sectionObj.KeyStrings()
		for _, name := range keyNames {
			key := sectionObj.Key(name)
			v := key.Value()
			if v == "" { // treat as flag-style key without value
				result[fmt.Sprintf("%s.%s", section, name)] = IgnoreValue
			} else {
				result[fmt.Sprintf("%s.%s", section, name)] = v
			}
		}
	}
	// Supplement: detect flag-only keys that some parsers may skip
	augmentFlagOnlyKeys(rawConf, result)
	return result, nil
}

// augmentFlagOnlyKeys scans raw INI to capture bare keys (no '=') as boolean/flag keys.
// This is a helper function for Parse.
func augmentFlagOnlyKeys(raw string, out map[string]string) {
	lines := strings.Split(raw, "\n")
	curSection := ""
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, ";") || strings.HasPrefix(s, "#") {
			continue
		}
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			curSection = strings.TrimSpace(s[1 : len(s)-1])
			continue
		}
		if strings.Contains(s, "=") {
			continue
		}
		// bare key
		if curSection != "" {
			key := curSection + "." + s
			if _, exists := out[key]; !exists {
				out[key] = IgnoreValue
			}
		}
	}
}

// sortMapKeys sorts the keys of a map alphabetically.
// This ensures a consistent order when writing the configuration back to a string.
func sortMapKeys(originMap map[string]string) []string {
	keys := make([]string, 0, len(originMap))
	for k := range originMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// writeConfigToStr converts a configuration map back into an INI-formatted string.
// It handles sections and both value-bearing and flag-style keys.
func writeConfigToStr(confMap map[string]string) string {
	keys := sortMapKeys(confMap)
	builder := strings.Builder{}
	section := make([]string, 0)

	for _, key := range keys {
		val := confMap[key]
		tmpSection := strings.Split(key, ".")
		if !lo.Contains(section, tmpSection[0]) {
			section = append(section, tmpSection[0])
			builder.WriteString("[" + tmpSection[0] + "]\n")
		}
		if val == IgnoreValue {
			builder.WriteString(tmpSection[1] + "=\n")
		} else {
			builder.WriteString(tmpSection[1] + "=" + val + "\n")
		}
	}
	return builder.String()
}
