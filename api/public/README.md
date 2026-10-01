# Public API/MCP contracts

These JSON artifacts are the authoritative, reviewable public contracts:

- `openapi.json`: HTTP operations.
- `contratto.schema.json`: shared public data schemas.
- `mcp-tools.json`: public MCP tool descriptors.

`embed.go` embeds them for the running Go service. Keep relative schema references
and published URLs compatible when updating a contract. Runtime contract files
belong here; explanatory guides and executable synthetic usage examples live in
[docs/backend](../../docs/backend/README.md).
