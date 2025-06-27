package mcp

// This file defines the core structures for handling MCP JSON-RPC messages.
// Based on https://modelcontextprotocol.io/specification/2025-06-18/

// Request represents a standard JSON-RPC request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response represents a standard JSON-RPC response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Notification represents a JSON-RPC notification.
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Error represents a JSON-RPC error object.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// --- Capability & Initialization Structs ---

type InitializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	// We don't need to parse client capabilities for our simple server.
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Resources map[string]interface{} `json:"resources"`
	Tools     map[string]interface{} `json:"tools"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

// --- Resource Structs ---

type Resource struct {
	URITemplate string `json:"uriTemplate,omitempty"`
	URI         string `json:"uri,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ListResourcesResult struct {
	ResourceTemplates []Resource `json:"resourceTemplates"`
}

type ReadResourceParams struct {
	URI string `json:"uri"`
}

type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

type ResourceContent struct {
	URI  string      `json:"uri"`
	Text interface{} `json:"text"`
}

// --- Tool Structs ---

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResult struct {
	Content []ToolResultContent `json:"content"`
}

type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Error Codes ---
const (
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)