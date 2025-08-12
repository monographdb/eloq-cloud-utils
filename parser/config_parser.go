package parser

import (
	"errors"
	"strings"
)

// ConfigParser defines the interface for configuration file parsers
// It provides methods to parse, merge, and manipulate configuration data
// for different formats (INI, YAML, etc.)
type ConfigParser interface {
	// Parse converts a raw configuration string into a flat key-value map
	// The keys use dot notation for nested structures (e.g., "section.key" for INI, "parent.child" for YAML)
	// Parsers must not depend on any product/module concept; they are format-only.
	Parse(rawConf string) (map[string]string, error)

	// UpsertConfig merges newConfigMap into oldConfMap, adding new keys and updating existing ones
	// Returns the serialized configuration string in the appropriate format
	UpsertConfig(oldConfMap, newConfigMap map[string]string) string

	// UpdateConfig updates only existing keys in oldConfMap with values from newConfigMap
	// Keys that don't exist in oldConfMap are ignored
	// Returns the serialized configuration string in the appropriate format
	UpdateConfig(oldConfMap, newConfigMap map[string]string) string

	// InsertConfig adds only new keys from newConfigMap to oldConfMap
	// Keys that already exist in oldConfMap are ignored
	// Returns the serialized configuration string in the appropriate format
	InsertConfig(oldConfMap, newConfigMap map[string]string) string
}

// Format identifies the configuration file format.
type Format string

const (
	FormatINI  Format = "ini"
	FormatYAML Format = "yaml"
	// IgnoreValue is a special value to indicate a key that has no value.
	IgnoreValue = "__ELOQ_NULL__"
)

var (
	// ignoreFields specifies keys that should be set to a fixed value, ignoring the value from the new configuration.
	ignoreFields = map[string]string{
		"mariadb.skip-log-bin": IgnoreValue,
		"mariadb.skip-innodb":  IgnoreValue,
		"mariadb.core_file":    IgnoreValue,
		"mariadb.eloq":         IgnoreValue,
	}
)

// AddIgnoreFields adds fields to the ignore list.
func AddIgnoreFields(fields map[string]string) {
	for k, v := range fields {
		ignoreFields[k] = v
	}
}

// RemoveIgnoreFields removes fields from the ignore list.
func RemoveIgnoreFields(keys []string) {
	for _, k := range keys {
		delete(ignoreFields, k)
	}
}

// GetIgnoreFields returns a copy of the current ignoreFields.
func GetIgnoreFields() map[string]string {
	m := make(map[string]string)
	for k, v := range ignoreFields {
		m[k] = v
	}
	return m
}

// NewConfigParser returns the appropriate parser for the given configuration format
func NewConfigParser(format Format) (ConfigParser, error) {
	switch format {
	case FormatYAML:
		return &YamlConfigParser{}, nil
	case FormatINI:
		return &IniConfigParser{}, nil
	default:
		return nil, errors.New("unsupported config format: " + string(format))
	}
}

// TemplateReplaceAll replaces all placeholders in the template with values from the entries map.
func TemplateReplaceAll(template string, entries map[string]string) string {
	for placeholder, value := range entries {
		if !strings.Contains(template, placeholder) {
			continue
		}
		template = strings.ReplaceAll(template, placeholder, value)
	}
	return template
}
