package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	terraformdoc "github.com/tjun/terraform-doc-mcp"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (rerr error) {
	s := server.NewMCPServer(
		"Terraform Doc MCP",
		"1.0.0",
		server.WithLogging(),
	)

	tool := mcp.NewTool("terraform-doc",
		mcp.WithDescription("get the terraform document of the given provider, version, and resource"),
		mcp.WithString("provider",
			mcp.Required(),
			mcp.Description("terraform provider name. supported providers are "+strings.Join(terraformdoc.SupportedProviders(), ", ")),
		),
		mcp.WithString("version",
			mcp.Description("provider version. default is latest"),
		),
		mcp.WithString("resource",
			mcp.Required(),
			mcp.Description("terraform resource name. format will be like aws_instance, google_compute_instance, datadog_monitor"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		provider := request.Params.Arguments["provider"].(string)
		version := "latest"
		if v, ok := request.Params.Arguments["version"].(string); ok {
			version = v
		}
		resource := request.Params.Arguments["resource"].(string)

		docs, err := terraformdoc.FetchTerraformMarkdown(provider, resource, version)
		if err != nil {
			return nil, fmt.Errorf("failed to get terraform doc: %w", err)
		}

		return mcp.NewToolResultText(string(docs)), nil
	})

	slog.Info("run mcp server")
	if err := server.ServeStdio(s); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
