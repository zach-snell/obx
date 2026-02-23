package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListVaultsHandler returns a JSON list of available vaults and the currently active one
func (v *Vault) ListVaultsHandler(ctx context.Context, req *mcp.CallToolRequest, args ListVaultsArgs) (*mcp.CallToolResult, any, error) {
	vaults := v.GetAllowedVaults()

	type VaultInfo struct {
		Alias string `json:"alias"`
		Path  string `json:"path"`
	}

	list := make([]VaultInfo, 0, len(vaults))
	for alias, path := range vaults {
		list = append(list, VaultInfo{Alias: alias, Path: path})
	}

	response := struct {
		ActiveVault string      `json:"active_vault"`
		Vaults      []VaultInfo `json:"vaults"`
	}{
		ActiveVault: v.GetPath(),
		Vaults:      list,
	}

	out, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode vault list: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(out)},
		},
	}, nil, nil
}

// SwitchVaultHandler attempts to switch the active MCP vault target
func (v *Vault) SwitchVaultHandler(ctx context.Context, req *mcp.CallToolRequest, args SwitchVaultArgs) (*mcp.CallToolResult, any, error) {
	vaults := v.GetAllowedVaults()

	// Check against aliases first
	if path, ok := vaults[args.Vault]; ok {
		v.SetPath(path)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Switched active vault to '%s' (%s)", args.Vault, path)},
			},
		}, nil, nil
	}

	// If not an alias, check if it's an exact path
	cleanTarget := filepath.Clean(args.Vault)
	for alias, path := range vaults {
		if filepath.Clean(path) == cleanTarget {
			v.SetPath(path)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Switched active vault to '%s' (%s)", alias, path)},
				},
			}, nil, nil
		}
	}

	// If no match found, reject
	return nil, nil, fmt.Errorf("vault '%s' not found or is not allowed. Use list-vaults to see available vaults", args.Vault)
}
