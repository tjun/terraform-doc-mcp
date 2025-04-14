package terraformdoc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type ProviderConfig struct {
	Repo        string
	DocPathFunc func(resource string) string
}

var providerConfigs = map[string]ProviderConfig{
	"aws": {
		Repo: "hashicorp/terraform-provider-aws",
		DocPathFunc: func(res string) string {
			return fmt.Sprintf("docs/resources/%s.md", strings.TrimPrefix(res, "aws_"))
		},
	},
	"azurerm": {
		Repo: "hashicorp/terraform-provider-azurerm",
		DocPathFunc: func(res string) string {
			return fmt.Sprintf("website/docs/r/%s.html.markdown", strings.TrimPrefix(res, "azurerm_"))
		},
	},
	"google": {
		Repo: "hashicorp/terraform-provider-google",
		DocPathFunc: func(res string) string {
			return fmt.Sprintf("website/docs/r/%s.html.markdown", strings.TrimPrefix(res, "google_"))
		},
	},
	"cloudflare": {
		Repo: "cloudflare/terraform-provider-cloudflare",
		DocPathFunc: func(res string) string {
			return fmt.Sprintf("docs/resources/%s.md", strings.TrimPrefix(res, "cloudflare_"))
		},
	},
	"datadog": {
		Repo: "DataDog/terraform-provider-datadog",
		DocPathFunc: func(res string) string {
			return fmt.Sprintf("docs/resources/%s.md", strings.TrimPrefix(res, "datadog_"))
		},
	},
}

func SupportedProviders() []string {
	keys := make([]string, 0, len(providerConfigs))
	for k := range providerConfigs {
		keys = append(keys, k)
	}
	return keys
}

func FetchTerraformMarkdown(provider, resource, version string) (string, error) {
	if provider == "" || resource == "" || version == "" {
		return "", fmt.Errorf("provider, resource, and version must not be empty")
	}

	config, ok := providerConfigs[provider]
	if !ok {
		keys := SupportedProviders()
		return "", fmt.Errorf("unsupported provider: %s. supported providers are %v", provider, keys)
	}

	if version == "latest" {
		var err error
		version, err = getLatestVersion(config.Repo)
		if err != nil {
			return "", err
		}
	}

	version = "v" + version

	docPath := config.DocPathFunc(resource)
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/refs/tags/%s/%s", config.Repo, version, docPath)
	slog.Info("Fetching Terraform documentation", "url", url)

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			slog.Error("Failed to close response body", "error", closeErr)
		}
	}()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("not found: %s", url)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	slog.Info("Successfully fetched Terraform documentation", "provider", provider, "resource", resource)
	return string(body), nil
}

func getLatestVersion(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", repo)

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			slog.Error("Failed to close response body", "error", closeErr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch tags: %s", res.Status)
	}

	var tags []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(res.Body).Decode(&tags); err != nil {
		return "", err
	}

	var versions []*semver.Version
	for _, tag := range tags {
		clean := strings.TrimPrefix(tag.Name, "v")
		if v, err := semver.NewVersion(clean); err == nil {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no valid semver tags found")
	}

	sort.Sort(semver.Collection(versions))
	latest := versions[len(versions)-1]

	return latest.String(), nil
}
