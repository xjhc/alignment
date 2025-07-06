
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// SQLGenerator generates SQL CREATE TABLE statements from Go structs
type SQLGenerator struct {
	output  strings.Builder
	fileSet *token.FileSet
}

func main() {
	generator := &SQLGenerator{
		fileSet: token.NewFileSet(),
	}

	// Generate schema from core/types.go
	if err := generator.generateFromGoFile("core/types.go"); err != nil {
		fmt.Printf("Error generating from core/types.go: %v\n", err)
		os.Exit(1)
	}

	// Generate schema from store/models.go
	if err := generator.generateFromGoFile("server/internal/store/models.go"); err != nil {
		fmt.Printf("Error generating from store/models.go: %v\n", err)
		os.Exit(1)
	}

	// Write to output file
	if err := os.WriteFile("schema.sql", []byte(generator.output.String()), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated schema.sql successfully")
}

func (g *SQLGenerator) generateFromGoFile(filePath string) error {
	node, err := parser.ParseFile(g.fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	g.output.WriteString(fmt.Sprintf("-- Schema generated from %s --\n\n", filePath))

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		g.generateCreateTable(typeSpec.Name.Name, structType.Fields)
		return true
	})

	return nil
}

func (g *SQLGenerator) generateCreateTable(structName string, fields *ast.FieldList) {
	tableName := toSnakeCase(structName)
	g.output.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))

	var columns []string
	for _, field := range fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name
			columnName := toSnakeCase(fieldName)

			sqlType := g.goTypeToSQL(field.Type)

			// Extract constraints from struct tags
			constraints := g.extractConstraints(field.Tag)

			columns = append(columns, fmt.Sprintf("    %s %s %s", columnName, sqlType, constraints))
		}
	}

	g.output.WriteString(strings.Join(columns, ",\n"))
	g.output.WriteString("\n);\n\n")
}

func (g *SQLGenerator) goTypeToSQL(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "VARCHAR(255)"
		case "int", "int64":
			return "BIGINT"
		case "int32":
			return "INTEGER"
		case "float64":
			return "DOUBLE PRECISION"
		case "float32":
			return "REAL"
		case "bool":
			return "BOOLEAN"
		case "Time":
			return "TIMESTAMP WITH TIME ZONE"
		default:
			return "TEXT"
		}
	case *ast.SelectorExpr:
		if sel, ok := t.X.(*ast.Ident); ok && sel.Name == "time" && t.Sel.Name == "Time" {
			return "TIMESTAMP WITH TIME ZONE"
		}
	case *ast.StarExpr: // Pointer to a type
		return g.goTypeToSQL(t.X)
	case *ast.ArrayType, *ast.MapType:
		return "JSONB"
	}
	return "TEXT" // Default fallback
}

func (g *SQLGenerator) extractConstraints(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}

	tagValue := strings.Trim(tag.Value, "`")
	var constraints []string

	// Example: `db:"id,primary_key" json:"id"`

	dbTag := extractTag(tagValue, "db")
	if dbTag != "" {
		parts := strings.Split(dbTag, ",")

		if contains(parts, "primary_key") {
			constraints = append(constraints, "PRIMARY KEY")
		}

		if contains(parts, "not_null") {
			constraints = append(constraints, "NOT NULL")
		}

		if contains(parts, "unique") {
			constraints = append(constraints, "UNIQUE")
		}
	}

	if strings.Contains(tagValue, "default:") {
		// More robust default value parsing needed here
		// For now, we'll keep it simple
		if strings.Contains(tagValue, "default:NOW()") {
			constraints = append(constraints, "DEFAULT NOW()")
		}
	}

	return strings.Join(constraints, " ")
}

func extractTag(tagString, key string) string {
	key = key + `:"`
	startIndex := strings.Index(tagString, key)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(key)
	endIndex := strings.Index(tagString[startIndex:], `"`)
	if endIndex == -1 {
		return ""
	}

	return tagString[startIndex : startIndex+endIndex]
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
    var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}