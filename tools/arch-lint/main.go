package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

// DeprecatedEvent represents a deprecated event constant
type DeprecatedEvent struct {
	Name  string
	Value string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse deprecated events from core/types.go
	deprecatedEvents, err := parseDeprecatedEvents("core/types.go")
	if err != nil {
		return fmt.Errorf("failed to parse deprecated events: %w", err)
	}

	if len(deprecatedEvents) == 0 {
		fmt.Println("✓ No deprecated events found")
		return nil
	}

	fmt.Printf("Found %d deprecated events\n", len(deprecatedEvents))
	for _, event := range deprecatedEvents {
		fmt.Printf("  - %s (%s)\n", event.Name, event.Value)
	}

	// Check frontend WebSocket handlers
	violations, err := checkWebSocketHandlers("client/src/services/websocket.ts", deprecatedEvents)
	if err != nil {
		return fmt.Errorf("failed to check WebSocket handlers: %w", err)
	}

	if len(violations) > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ ARCHITECTURAL VIOLATIONS DETECTED:\n")
		for _, violation := range violations {
			fmt.Fprintf(os.Stderr, "  Line %d: Usage of deprecated event %s\n", violation.Line, violation.EventName)
		}
		fmt.Fprintf(os.Stderr, "\nSee ADR-006 for guidance on using single authoritative events instead.\n")
		return fmt.Errorf("found %d architectural violations", len(violations))
	}

	fmt.Println("✓ No violations found in frontend WebSocket handlers")
	return nil
}

// parseDeprecatedEvents parses the Go source file and extracts deprecated event constants
func parseDeprecatedEvents(filename string) ([]DeprecatedEvent, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	var deprecatedEvents []DeprecatedEvent

	// Walk the AST looking for const declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				// Check each specification in the const block
				for _, spec := range x.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						// Check if there's a comment indicating deprecation
						var hasDeprecatedComment bool
						
						// Check for inline comments (most precise)
						if valueSpec.Comment != nil {
							for _, comment := range valueSpec.Comment.List {
								if strings.Contains(comment.Text, "Deprecated:") || strings.Contains(comment.Text, "DEPRECATED:") {
									hasDeprecatedComment = true
									break
								}
							}
						}

						// Check comments in the preceding lines (for block comments)
						if !hasDeprecatedComment {
							position := fset.Position(valueSpec.Pos())
							// Look for deprecated comment in preceding lines
							if prevComment := findPrecedingDeprecatedComment(filename, position.Line); prevComment {
								hasDeprecatedComment = true
							}
						}

						if hasDeprecatedComment {
							for j, name := range valueSpec.Names {
								var value string
								if j < len(valueSpec.Values) {
									if basicLit, ok := valueSpec.Values[j].(*ast.BasicLit); ok {
										value = strings.Trim(basicLit.Value, `"`)
									}
								}
								deprecatedEvents = append(deprecatedEvents, DeprecatedEvent{
									Name:  name.Name,
									Value: value,
								})
							}
						}
					}
				}
			}
		}
		return true
	})

	return deprecatedEvents, nil
}

// findPrecedingDeprecatedComment checks if there's a deprecated comment in the lines immediately before the given line
func findPrecedingDeprecatedComment(filename string, line int) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	var previousLine string
	
	// Read until we get to our target line
	for scanner.Scan() {
		text := scanner.Text()
		if currentLine == line {
			// Only check the immediately preceding line for a comment-only line
			if strings.Contains(previousLine, "Deprecated:") || strings.Contains(previousLine, "DEPRECATED:") {
				// Only consider it if the previous line is a pure comment (starts with // or /*)
				trimmed := strings.TrimSpace(previousLine)
				if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
					return true
				}
			}
			break
		}
		
		previousLine = text
		currentLine++
	}
	
	return false
}

// Violation represents a usage of a deprecated event
type Violation struct {
	Line      int
	EventName string
}

// checkWebSocketHandlers scans the TypeScript file for usage of deprecated events
func checkWebSocketHandlers(filename string, deprecatedEvents []DeprecatedEvent) ([]Violation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var violations []Violation
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	// Create regex patterns for each deprecated event
	var patterns []*regexp.Regexp
	var eventNames []string
	for _, event := range deprecatedEvents {
		// Look for case statements handling the deprecated event
		// Convert EventName to enum format (remove "Event" prefix for TypeScript enum)
		enumName := strings.TrimPrefix(event.Name, "Event")
		
		// Patterns: case ServerEventType.EnumName: or case "EVENT_VALUE":
		pattern1 := regexp.MustCompile(fmt.Sprintf(`case\s+ServerEventType\.%s\s*:`, enumName))
		pattern2 := regexp.MustCompile(fmt.Sprintf(`case\s+['"]%s['"]s*:`, regexp.QuoteMeta(event.Value)))
		patterns = append(patterns, pattern1, pattern2)
		eventNames = append(eventNames, event.Name, event.Name) // Duplicate for both patterns
	}

	for scanner.Scan() {
		line := scanner.Text()
		
		// Check each pattern against the current line
		for i, pattern := range patterns {
			if pattern.MatchString(line) {
				violations = append(violations, Violation{
					Line:      lineNumber,
					EventName: eventNames[i],
				})
			}
		}
		
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return violations, nil
}