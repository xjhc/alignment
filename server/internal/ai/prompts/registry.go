package prompts

import (
	"fmt"
	"math/rand"
	"time"
)

// Global registry of prompt templates
var registry = make(map[string]PromptTemplate)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Register adds a prompt template to the global registry
func Register(template PromptTemplate) {
	registry[template.ID()] = template
}

// Get retrieves a prompt template by ID
func Get(id string) (PromptTemplate, error) {
	template, exists := registry[id]
	if !exists {
		return nil, fmt.Errorf("prompt template '%s' not found", id)
	}
	return template, nil
}

// GetRandom returns a randomly selected prompt template
func GetRandom() PromptTemplate {
	if len(registry) == 0 {
		return nil
	}
	
	// Convert map to slice for random selection
	templates := make([]PromptTemplate, 0, len(registry))
	for _, template := range registry {
		templates = append(templates, template)
	}
	
	return templates[rng.Intn(len(templates))]
}

// List returns all available template IDs
func List() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}

// GetByPersona returns a template that matches the given persona style
func GetByPersona(persona string) (PromptTemplate, error) {
	// Map persona IDs to template IDs
	personaMap := map[string]string{
		"shadow":     "millennial_lean",
		"puppeteer":  "gen_z_cot",
		"default":    "millennial_lean",
	}
	
	templateID, exists := personaMap[persona]
	if !exists {
		templateID = "millennial_lean" // Default fallback
	}
	
	return Get(templateID)
}

// Count returns the number of registered templates
func Count() int {
	return len(registry)
}