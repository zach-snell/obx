package vault

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	modeCompact  = "compact"
	modeDetailed = "detailed"
)

type compactResponse struct {
	Status    string         `json:"status"`
	Mode      string         `json:"mode"`
	Summary   string         `json:"summary"`
	Truncated bool           `json:"truncated,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Next      map[string]any `json:"next,omitempty"`
}

func normalizeMode(mode string) string {
	if mode == modeDetailed {
		return modeDetailed
	}
	return modeCompact
}

func isDetailedMode(mode string) bool {
	return normalizeMode(mode) == modeDetailed
}

func compactResult(summary string, truncated bool, data map[string]any, next map[string]any) (*mcp.CallToolResult, any, error) {
	response := compactResponse{
		Status:    "ok",
		Mode:      modeCompact,
		Summary:   summary,
		Truncated: truncated,
		Data:      data,
		Next:      next,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build compact response: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, nil, nil
}
