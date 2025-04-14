package terraformdoc

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchTerraformMarkdown(t *testing.T) {
	// オリジナルのトランスポートを保存
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	// カスタムトランスポートを設定
	http.DefaultTransport = &mockTransport{}

	tests := []struct {
		name        string
		provider    string
		resource    string
		version     string
		expectedDoc string
		expectError bool
	}{
		{
			name:        "正常系 - 最新バージョンを使用",
			provider:    "aws",
			resource:    "aws_instance",
			version:     "latest",
			expectedDoc: "# aws_instance\n\nProvides an EC2 instance resource.",
			expectError: false,
		},
		{
			name:        "異常系 - サポートされていないプロバイダー",
			provider:    "unsupported",
			resource:    "resource",
			version:     "latest",
			expectedDoc: "",
			expectError: true,
		},
		{
			name:        "異常系 - 空のパラメータ",
			provider:    "",
			resource:    "",
			version:     "",
			expectedDoc: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := FetchTerraformMarkdown(tt.provider, tt.resource, tt.version)

			if tt.expectError {
				if err == nil {
					t.Errorf("期待されるエラーが発生しませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期せぬエラーが発生しました: %v", err)
				}
			}

			if doc != tt.expectedDoc && !tt.expectError {
				t.Errorf("期待されるドキュメントと一致しません。\n期待値: %s\n実際: %s", tt.expectedDoc, doc)
			}
		})
	}
}

// モック用トランスポート
type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// リクエストURLに基づいてモックレスポンスを返す
	url := req.URL.String()

	// GitHub APIのタグ一覧リクエスト
	if strings.Contains(url, "api.github.com/repos/hashicorp/terraform-provider-aws/tags") {
		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(`[
				{"name": "v5.0.0"},
				{"name": "v4.0.0"},
				{"name": "v3.0.0"}
			]`)),
			Header: make(http.Header),
		}, nil
	}

	// AWSプロバイダーのドキュメントリクエスト
	if strings.Contains(url, "raw.githubusercontent.com/hashicorp/terraform-provider-aws/refs/tags/v5.0.0/docs/resources/instance.md") {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("# aws_instance\n\nProvides an EC2 instance resource.")),
			Header:     make(http.Header),
		}, nil
	}

	// その他のリクエストは404を返す
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
		Header:     make(http.Header),
	}, nil
}
