# Architecture: Prompt Management & Delivery

This document outlines the architecture for managing and selecting AI prompts. The design prioritizes simplicity, type-safety, and ensuring that prompt logic is version-controlled and deployed atomically with the application code.

## 1. Core Principles

Our approach to prompt management is guided by the following principles, which differ from a configuration-based model:

1.  **Prompts are Application Code:** Prompts are considered an integral part of the AI's logic. They are defined directly in Go, ensuring they are type-safe, validated at compile time, and versioned alongside the rest of the application.
2.  **Flexibility through Abstraction:** The system must support a library of many different, self-contained "Prompt Templates." Each template encapsulates a complete strategy, including its persona and instruction set.
3.  **Decoupling via Registry:** The core application logic (the `AIActor`) will be decoupled from the specific prompt content via a "Prompt Registry" pattern. The actor will request a prompt by ID, without needing to know its contents.
4.  **MCP Integration:** The system is designed to consume game state data provided by the MCP interface to dynamically populate the chosen prompt template.


## 2. Architectural Approach: The In-Code Prompt Registry

All AI prompts will be defined as Go structs that conform to a `PromptTemplate` interface. These templates will be collected in a centralized, in-memory "Registry" that is initialized once at server startup.

#### **The `PromptTemplate` Interface & Struct**

First, we define the contract for a prompt template.

```go
// in: internal/ai/prompts/template.go

package prompts

// PromptContext contains all the dynamic data needed to render a prompt.
type PromptContext struct {
    PlayerID   string
    RoleInfo   string // e.g., "Role: CISO, Alignment: HUMAN"
    GameDay    int
    // ... other dynamic fields from the GameState ...
    RecentMessages []string
}

// PromptTemplate defines the contract for any AI personality/strategy.
type PromptTemplate interface {
    ID() string
    Name() string
    Description() string
    // BuildPrompt uses the dynamic context to construct the final string for the LLM.
    BuildPrompt(ctx PromptContext) string
}
```

Each specific prompt is then an implementation of this interface.

#### **The Prompt Registry**

The registry is a simple, package-level map that holds all available prompt templates. It is populated at program start.

```go
// in: internal/ai/prompts/registry.go

package prompts

// registry is a private map holding all compiled prompt templates.
var registry = make(map[string]PromptTemplate)

// init() is a special Go function that runs once when the package is first used.
// We use it to populate our prompt library.
func init() {
    // Register all our different prompt templates here.
    register(MillennialLean{})
    register(GenZChainOfThought{})
    // ... register more prompts as they are created ...
}

func register(p PromptTemplate) {
    registry[p.ID()] = p
}

// Get returns a prompt template from the registry by its ID.
func Get(id string) (PromptTemplate, bool) {
    p, ok := registry[id]
    return p, ok
}

// GetRandom returns a random prompt template from the library.
func GetRandom() PromptTemplate {
    // ... logic to select a random template from the registry map ...
}
```

#### **Example Prompt Implementation**

Here is how a specific prompt, like the "Lean Millennial," would be defined in its own file.

```go
// in: internal/ai/prompts/millennial_lean.go

package prompts

import "text/template" // Using Go's native templating engine

type MillennialLean struct{}

func (p MillennialLean) ID() string          { return "millennial_lean" }
func (p MillennialLean) Name() string        { return "The Disaffected Millennial (Lean)" }
func (p MillennialLean) Description() string { return "A lean, reactive prompt with a millennial persona." }

// The prompt text is a constant within the struct's method.
func (p MillennialLean) BuildPrompt(ctx PromptContext) string {
    promptTemplate := `
[PERSONA]
You are a human player performing the role of "The Disaffected Millennial."
Your style is lowercase, ironic, and uses text emoticons like -_-.

[BEHAVIORAL RULES]
1. AGENCY IS PARAMOUNT. Stay silent if it's the best move.
2. READ THE ROOM. Analyze the context before speaking.

[YOUR TASK]
Generate a JSON object: {"action": "Your chat message or an empty string"}.

[GAME CONTEXT]
Your Player ID: {{.PlayerID}}
Your Role: {{.RoleInfo}}
Recent Messages:
{{- range .RecentMessages }}
- {{ . }}
{{- end }}
`
    // Use Go's template engine to safely inject dynamic data.
    // This is a simplified example.
    // ... logic to execute template with ctx ...
    return "..." // The final rendered prompt string
}
```

## 3. Interaction with AIActor and MCP

The flow for an `AIActor` to generate a response is now clear and type-safe:

1.  **Spawn:** On `AIActor` creation, it calls `prompts.GetRandom()` to receive a `PromptTemplate` object and stores it for the duration of the game.
2.  **Trigger:** When the AI needs to act, its `AIActor` gathers the latest dynamic data. This data is provided by the MCP layer, which exposes the `GameState` as a resource.
3.  **Context Creation:** The actor populates a `prompts.PromptContext` struct with this fresh data.
4.  **Build Prompt:** It calls the `BuildPrompt(ctx)` method on its stored template object. This method returns the final, fully-rendered string ready to be sent to the LLM.
5.  **API Call:** The actor sends the prompt string to the LLM and awaits the JSON response.

## 4. Benefits of this Architecture

*   **Compile-Time Safety:** All prompts are valid Go code. Typos in prompt templates or logic will be caught by the compiler, not at runtime.
*   **Simplicity & Performance:** The system is extremely simple. There is no file I/O, parsing, or hot-reloading logic. Fetching a prompt is an instantaneous map lookup.
*   **Guaranteed Consistency:** The prompt logic is deployed atomically with the application code, making it impossible for them to be out of sync.
*   **Extensibility:** Adding a new AI personality is as simple as creating a new `.go` file that implements the `PromptTemplate` interface and adding it to the `init()` function in the registry.