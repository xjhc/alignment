package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type TypeGenerator struct {
	output strings.Builder
	fileSet *token.FileSet
	structs map[string]*ast.StructType
	fields  map[string][]*ast.Field
}

func (tg *TypeGenerator) parseGoFile(filePath string) error {
	// Parse Go file using go/parser for robust AST parsing
	node, err := parser.ParseFile(tg.fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Parse constants for enum generation
	tg.parseConstants(node)

	// Parse struct types for interface generation
	tg.parseStructTypes(node)

	return nil
}

// parseConstants extracts enum constants from the AST
func (tg *TypeGenerator) parseConstants(file *ast.File) {
	var eventTypes, actionTypes, phaseTypes, roleTypes, kpiTypes, voteTypes []string

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue
				}

				basicLit, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || basicLit.Kind != token.STRING {
					continue
				}

				// Remove quotes from string literal
				value := strings.Trim(basicLit.Value, `"`)
				constName := name.Name

				// Categorize constants by type
				switch {
				case strings.HasPrefix(constName, "Event"):
					eventTypes = append(eventTypes, value)
				case strings.HasPrefix(constName, "Action"):
					actionTypes = append(actionTypes, value)
				case strings.HasPrefix(constName, "Phase"):
					phaseTypes = append(phaseTypes, value)
				case strings.HasPrefix(constName, "Role"):
					roleTypes = append(roleTypes, value)
				case strings.HasPrefix(constName, "KPI"):
					kpiTypes = append(kpiTypes, value)
				case strings.HasPrefix(constName, "Vote"):
					voteTypes = append(voteTypes, value)
				}
			}
		}
	}

	// Generate TypeScript enums
	tg.generateEnum("ServerEventType", eventTypes)
	tg.generateEnum("ClientActionType", actionTypes)
	tg.generateEnum("PhaseType", phaseTypes)
	tg.generateEnum("RoleType", roleTypes)
	tg.generateEnum("KPIType", kpiTypes)
	tg.generateEnum("VoteType", voteTypes)
}

// parseStructTypes extracts struct definitions from the AST
func (tg *TypeGenerator) parseStructTypes(file *ast.File) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Store struct definition for interface generation
			tg.structs[typeSpec.Name.Name] = structType
			tg.fields[typeSpec.Name.Name] = structType.Fields.List
		}
	}
}

func (tg *TypeGenerator) generateEnum(enumName string, values []string) {
	if len(values) == 0 {
		return
	}

	tg.output.WriteString(fmt.Sprintf("// Generated %s enum\n", enumName))
	tg.output.WriteString(fmt.Sprintf("export enum %s {\n", enumName))
	
	for _, value := range values {
		enumKey := tg.toEnumKey(value)
		tg.output.WriteString(fmt.Sprintf("  %s = %q,\n", enumKey, value))
	}
	
	tg.output.WriteString("}\n\n")
}

func (tg *TypeGenerator) toEnumKey(value string) string {
	// Convert UPPER_CASE to PascalCase for enum keys
	parts := strings.Split(value, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	return result.String()
}

func (tg *TypeGenerator) generateInterfaces() {
	tg.output.WriteString("// Generated interfaces from Go structs\n")

	// Generate interfaces for each parsed struct
	for structName, fields := range tg.fields {
		tg.generateInterface(structName, fields)
	}

	// Generate union types for easier usage
	tg.output.WriteString(`
// Union types for easier usage
export type AnyEventType = keyof typeof ServerEventType;
export type AnyActionType = keyof typeof ClientActionType;
export type AnyPhaseType = keyof typeof PhaseType;
export type AnyRoleType = keyof typeof RoleType;
export type AnyKPIType = keyof typeof KPIType;
export type AnyVoteType = keyof typeof VoteType;

`)
}

// generateInterface creates a TypeScript interface from Go struct fields
func (tg *TypeGenerator) generateInterface(structName string, fields []*ast.Field) {
	interfaceName := "Generated" + structName
	tg.output.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))

	for _, field := range fields {
		if len(field.Names) == 0 {
			// Embedded field - skip for now
			continue
		}

		for _, name := range field.Names {
			fieldName := tg.convertFieldName(name.Name, field)
			fieldType := tg.convertGoTypeToTS(field.Type)
			optional := tg.isOptionalField(field)

			if optional {
				tg.output.WriteString(fmt.Sprintf("  %s?: %s;\n", fieldName, fieldType))
			} else {
				tg.output.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
			}
		}
	}

	tg.output.WriteString("}\n\n")
}

// convertFieldName converts Go field names to camelCase for TypeScript
func (tg *TypeGenerator) convertFieldName(goFieldName string, field *ast.Field) string {
	// Look for json tag first
	if field.Tag != nil {
		tagValue := strings.Trim(field.Tag.Value, "`")
		if strings.Contains(tagValue, "json:") {
			// Extract JSON field name from tag
			parts := strings.Fields(tagValue)
			for _, part := range parts {
				if strings.HasPrefix(part, "json:") {
					jsonTag := strings.TrimPrefix(part, "json:")
					jsonTag = strings.Trim(jsonTag, `"`)
					jsonName := strings.Split(jsonTag, ",")[0]
					if jsonName != "" && jsonName != "-" {
						return jsonName
					}
				}
			}
		}
	}

	// No json tag, convert to camelCase
	runes := []rune(goFieldName)
	if len(runes) == 0 {
		return goFieldName
	}

	// Convert first character to lowercase
	runes[0] = rune(strings.ToLower(string(runes[0]))[0])
	return string(runes)
}

// isOptionalField determines if a field should be optional in TypeScript
func (tg *TypeGenerator) isOptionalField(field *ast.Field) bool {
	// Check if field has json:",omitempty" tag
	if field.Tag != nil {
		tagValue := strings.Trim(field.Tag.Value, "`")
		return strings.Contains(tagValue, "omitempty")
	}

	// Check if field is a pointer type
	if _, ok := field.Type.(*ast.StarExpr); ok {
		return true
	}

	return false
}

// convertGoTypeToTS converts Go types to TypeScript types
func (tg *TypeGenerator) convertGoTypeToTS(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		case "Time":
			return "string" // time.Time serializes to string in JSON
		case "Duration":
			return "number" // time.Duration serializes to number in JSON
		default:
			// Check if it's a custom type that should reference generated interface
			if _, exists := tg.structs[t.Name]; exists {
				return "Generated" + t.Name
			}
			// Assume it's a simple type alias
			return "string"
		}
	case *ast.StarExpr:
		// Pointer types - recursively get the underlying type
		return tg.convertGoTypeToTS(t.X)
	case *ast.ArrayType:
		// Arrays and slices
		elementType := tg.convertGoTypeToTS(t.Elt)
		return elementType + "[]"
	case *ast.MapType:
		// Maps
		keyType := tg.convertGoTypeToTS(t.Key)
		valueType := tg.convertGoTypeToTS(t.Value)
		if keyType == "string" {
			return fmt.Sprintf("Record<string, %s>", valueType)
		}
		return "Record<string, any>"
	case *ast.InterfaceType:
		// Interface{} types
		return "any"
	case *ast.SelectorExpr:
		// Package.Type expressions (like time.Time)
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				return "string"
			}
			if ident.Name == "time" && t.Sel.Name == "Duration" {
				return "number"
			}
		}
		return "any"
	default:
		return "any"
	}
}

func (tg *TypeGenerator) generateAll() {
	tg.output.WriteString("// AUTO-GENERATED FILE - DO NOT EDIT\n")
	tg.output.WriteString("// Generated from Go core package types\n")
	tg.output.WriteString("// Run 'npm run generate:types' to update\n\n")
	
	// Initialize data structures
	tg.fileSet = token.NewFileSet()
	tg.structs = make(map[string]*ast.StructType)
	tg.fields = make(map[string][]*ast.Field)
	
	// Parse the core types file
	if err := tg.parseGoFile("../../core/types.go"); err != nil {
		fmt.Printf("Error parsing Go file: %v\n", err)
		os.Exit(1)
	}
	
	// Generate interfaces
	tg.generateInterfaces()
}

func main() {
	generator := &TypeGenerator{}
	generator.generateAll()
	
	// Write to the output file
	outputPath := "../../client/src/types/generated.ts"
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	
	if err := os.WriteFile(outputPath, []byte(generator.output.String()), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Generated TypeScript definitions at %s\n", outputPath)
}