# eloq-cloud-utils

A Go library of common utilities for Eloq cloud services. This repository will host various reusable packages.

Currently, the available utility is a configuration parser.

## Project Structure

- `parser/`: This directory contains the core logic for the configuration parser.
  - `config_parser.go`: Defines the `ConfigParser` interface and a factory function `NewConfigParser` to create format-specific parsers.
  - `ini_config_parser.go`: Implements the `ConfigParser` interface for INI files.
  - `yaml_config_parser.go`: Implements the `ConfigParser` interface for YAML files.
- `test/`: This directory holds all the tests.
  - `config_integration_test.go`: Contains integration tests that verify the parser's functionality with different configuration formats.
  - `example/`: Contains sample configuration files used for testing.

## Config Parser

A Go library for parsing and manipulating configuration files for different formats (INI, YAML).

### Features

- **Unified Interface**: Provides a single `ConfigParser` interface for different configuration formats.
- **Format Support**: Supports INI and YAML formats.
- **Operations**:
    - `Parse`: Converts configuration files into a flat key-value map.
    - `UpsertConfig`: Adds new keys and updates existing ones.
    - `UpdateConfig`: Updates only existing keys.
    - `InsertConfig`: Adds only new keys.
- **Serialization**: Converts the configuration map back to a string in the original format.

### Installation

```bash
go get github.com/monographdb/eloq-cloud-utils@latest
```

### Usage

Here's a simple example of how to use the config parser:

```go
package main

import (
	"fmt"
	"log"

	"github.com/monographdb/eloq-cloud-utils/parser"
)

func main() {
	// Choose the format (FormatINI or FormatYAML)
	format := parser.FormatINI

	// Create a new config parser
	configParser, err := parser.NewConfigParser(format)
	if err != nil {
		log.Fatalf("Failed to create config parser: %v", err)
	}

	// Raw configuration string
	rawConfig := `
[database]
host = localhost
port = 3306
`

	// Parse the configuration
	configMap, err := configParser.Parse(rawConfig)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Println("Original Config:", configMap)

	// New configuration to merge
	newConfig := map[string]string{
		"database.port": "3307",
		"database.user": "admin",
	}

	// Upsert the configuration
	updatedConfigStr := configParser.UpsertConfig(configMap, newConfig)
	fmt.Println("\nUpdated Config:\n", updatedConfigStr)
}
```

## Running Tests

To run the tests for this project, use the following command:

```bash
go test ./...
```
