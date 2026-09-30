// Package contract embeds the reviewed public API/MCP contract for use by the
// running service. The JSON file remains the normative, human-reviewable
// artifact; embedding prevents runtime filesystem dependencies.
package contract

import _ "embed"

// MCPTools is the reviewed descriptor for the five public MCP tools.
//
//go:embed mcp-tools.json
var MCPTools []byte

// OpenAPI describes the public HTTP API.
//
//go:embed openapi.json
var OpenAPI []byte

// Schema contains the shared schemas referenced by OpenAPI.
//
//go:embed contratto.schema.json
var Schema []byte
