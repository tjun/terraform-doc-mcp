# terraform-doc-mcp

## Overview

terraform-doc-mcp is an MCP server for retrieving documentation for Terraform providers and resources. 
It can be used to reference Terraform resource documentation from LLMs such as Claude Desktop.

## Features

- Supports multiple Terraform providers (aws, azurerm, google, cloudflare, datadog)
- Ability to fetch documentation for specific versions or the latest version

### Parameters

- `provider`: Terraform provider name (required). Supported providers: aws, azurerm, google, cloudflare, datadog
- `resource`: Terraform resource name (required). Examples: aws_instance, google_compute_instance, datadog_monitor
- `version`: Provider version (optional, default is "latest")

## Setup

### Using with Claude Desktop

To use this with Claude Desktop, add the following to your claude_desktop_config.json:

```json
{
  "mcpServers": {
    "terraform-doc": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "ghcr.io/tjun/terraform-doc-mcp:1.0"
      ],
      "env": {}
    }
  }
}
```

## License

This project is released under the license specified in the LICENSE file.
