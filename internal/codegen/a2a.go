package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// GenerateA2ATypes generates Go types from A2A JSON/YAML schema
func GenerateA2ATypes(destination string, schemaPath string) error {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read A2A schema: %w", err)
	}

	var schema map[string]interface{}

	switch {
	case strings.HasSuffix(schemaPath, ".json"):
		if err := json.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("failed to parse JSON schema: %w", err)
		}
	case strings.HasSuffix(schemaPath, ".yaml"), strings.HasSuffix(schemaPath, ".yml"):
		if err := yaml.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("failed to parse YAML schema: %w", err)
		}
	default:
		return fmt.Errorf("unsupported schema format: must be .json, .yaml, or .yml")
	}

	definitions, ok := schema["definitions"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema does not contain definitions")
	}

	outputFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil {
			fmt.Printf("Warning: Failed to close output file: %v\n", closeErr)
		}
	}()

	header := `// Code generated from A2A schema. DO NOT EDIT.
package a2a

`
	if _, err := outputFile.WriteString(header); err != nil {
		return fmt.Errorf("failed to write file header: %w", err)
	}

	processedTypes := map[string]bool{}

	acronyms := map[string]bool{
		"id":      true,
		"uri":     true,
		"url":     true,
		"api":     true,
		"html":    true,
		"http":    true,
		"https":   true,
		"json":    true,
		"jsonrpc": true,
		"rpc":     true,
		"mime":    true,
		"a2a":     true,
		"sse":     true,
		"uuid":    true,
	}

	typeNames := make([]string, 0, len(definitions))
	for typeName := range definitions {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	for _, typeName := range typeNames {
		definition := definitions[typeName]

		defMap, ok := definition.(map[string]interface{})
		if !ok {
			continue
		}

		isEnum := false
		var enumValues []interface{}
		if enum, ok := defMap["enum"].([]interface{}); ok && len(enum) > 0 {
			isEnum = true
			enumValues = enum
		}

		if !isEnum {
			continue
		}

		description := ""
		if desc, ok := defMap["description"].(string); ok {
			description = desc
		}

		if description != "" {
			formattedDescription := formatDescription(description)
			if _, err := outputFile.WriteString(formattedDescription + "\n"); err != nil {
				return err
			}
		}

		typeStr := "string"
		if t, ok := defMap["type"].(string); ok {
			typeStr = t
		}

		typeDecl := fmt.Sprintf("type %s %s\n\n", typeName, typeStr)
		if _, err := outputFile.WriteString(typeDecl); err != nil {
			return err
		}

		constDecl := fmt.Sprintf("// %s enum values\nconst (\n", typeName)
		if _, err := outputFile.WriteString(constDecl); err != nil {
			return err
		}

		enumStrings := make([]string, 0, len(enumValues))
		for _, val := range enumValues {
			if strVal, ok := val.(string); ok {
				enumStrings = append(enumStrings, strVal)
			}
		}
		sort.Strings(enumStrings)

		for _, val := range enumStrings {
			enumVal := fmt.Sprintf("\t%s%s %s = \"%s\"\n", typeName, convertToGoFieldName(val, acronyms), typeName, val)
			if _, err := outputFile.WriteString(enumVal); err != nil {
				return err
			}
		}

		if _, err := outputFile.WriteString(")\n\n"); err != nil {
			return err
		}

		processedTypes[typeName] = true
	}

	for _, typeName := range typeNames {
		definition := definitions[typeName]

		defMap, ok := definition.(map[string]interface{})
		if !ok {
			continue
		}

		if processedTypes[typeName] {
			continue
		}

		if anyOf, ok := defMap["anyOf"].([]interface{}); ok && len(anyOf) > 0 {
			description := ""
			if desc, ok := defMap["description"].(string); ok {
				description = desc
			}

			if description != "" {
				formattedDescription := formatDescription(description)
				if _, err := outputFile.WriteString(formattedDescription + "\n"); err != nil {
					return err
				}
			}

			typeDecl := fmt.Sprintf("type %s interface{}\n\n", typeName)
			if _, err := outputFile.WriteString(typeDecl); err != nil {
				return err
			}
			continue
		}

		description := ""
		if desc, ok := defMap["description"].(string); ok {
			description = desc
		}

		if description != "" {
			formattedDescription := formatDescription(description)
			if _, err := outputFile.WriteString(formattedDescription + "\n"); err != nil {
				return err
			}
		}

		structDef := fmt.Sprintf("type %s struct {\n", typeName)
		if _, err := outputFile.WriteString(structDef); err != nil {
			return err
		}

		properties, ok := defMap["properties"].(map[string]interface{})
		if ok {
			propNames := make([]string, 0, len(properties))
			for propName := range properties {
				propNames = append(propNames, propName)
			}
			sort.Strings(propNames)

			for _, propName := range propNames {
				propDef := properties[propName]
				propMap, ok := propDef.(map[string]interface{})
				if !ok {
					continue
				}

				fieldName := convertToGoFieldName(propName, acronyms)

				propType := determineGoType(propMap, definitions)
				propDefStr := fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, propType, propName)
				if _, err := outputFile.WriteString(propDefStr); err != nil {
					return err
				}
			}
		}

		if _, err := outputFile.WriteString("}\n\n"); err != nil {
			return err
		}
	}

	cmd := exec.Command("go", "fmt", destination)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to format %s: %v\n", destination, err)
	}

	return nil
}

// formatDescription formats a description string as proper Go comments
// with each line prefixed by "// "
func formatDescription(description string) string {
	if description == "" {
		return ""
	}

	lines := strings.Split(description, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines[i] = "// " + line
		} else {
			lines[i] = "//"
		}
	}

	return strings.Join(lines, "\n")
}

// convertToGoFieldName converts a JSON property name to a properly capitalized Go field name
func convertToGoFieldName(name string, acronyms map[string]bool) string {
	if name == "_meta" {
		return "Meta"
	}

	name = strings.ReplaceAll(name, "-", "_")

	var parts []string
	var current strings.Builder

	for i, r := range name {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	var finalParts []string
	for _, part := range parts {
		subParts := strings.Split(part, "_")
		for _, subPart := range subParts {
			if subPart != "" {
				finalParts = append(finalParts, subPart)
			}
		}
	}

	for i, part := range finalParts {
		lowerPart := strings.ToLower(part)
		if acronyms[lowerPart] {
			finalParts[i] = strings.ToUpper(lowerPart)
		} else {
			finalParts[i] = cases.Title(language.English).String(lowerPart)
		}
	}

	return strings.Join(finalParts, "")
}

// determineGoType determines the Go type for a JSON schema property
func determineGoType(propMap map[string]interface{}, definitions map[string]interface{}) string {
	if ref, ok := propMap["$ref"].(string); ok {
		parts := strings.Split(ref, "/")
		refType := parts[len(parts)-1]
		return refType
	}

	if propType, ok := propMap["type"].(string); ok && propType == "array" {
		if items, ok := propMap["items"].(map[string]interface{}); ok {
			itemType := determineGoType(items, definitions)
			return "[]" + itemType
		}
		return "[]interface{}"
	}

	if propType, ok := propMap["type"].(string); ok {
		format := ""
		if fmt, ok := propMap["format"].(string); ok {
			format = fmt
		}

		switch propType {
		case "string":
			if format == "date-time" {
				return "time.Time"
			}
			return "string"
		case "integer":
			if format == "int64" {
				return "int64"
			}
			return "int"
		case "number":
			return "float64"
		case "boolean":
			return "bool"
		case "object":
			return "map[string]interface{}"
		}
	}

	return "interface{}"
}
