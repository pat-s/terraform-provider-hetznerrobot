package hetznerrobot

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderConfigure(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		password  string
		url       string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid credentials",
			username:  "testuser",
			password:  "testpass",
			url:       "https://robot-ws.your-server.de",
			expectErr: false,
		},
		{
			name:      "empty username",
			username:  "",
			password:  "testpass",
			url:       "https://robot-ws.your-server.de",
			expectErr: true,
			errMsg:    "username is required for Hetzner Robot authentication",
		},
		{
			name:      "empty password",
			username:  "testuser",
			password:  "",
			url:       "https://robot-ws.your-server.de",
			expectErr: true,
			errMsg:    "password is required for Hetzner Robot authentication",
		},
		{
			name:      "valid with custom url",
			username:  "testuser",
			password:  "testpass",
			url:       "https://custom-robot-api.example.com",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a resource data with test values
			resourceData := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
				"username": tt.username,
				"password": tt.password,
				"url":      tt.url,
			})

			// Test provider configuration
			client, diags := providerConfigure(context.Background(), resourceData)

			if tt.expectErr {
				if !diags.HasError() {
					t.Fatalf("Expected error but got none")
				}
				if diags[0].Summary != tt.errMsg {
					t.Fatalf("Expected error message '%s', got '%s'", tt.errMsg, diags[0].Summary)
				}
				if client != nil {
					t.Fatalf("Expected nil client when error occurs")
				}
			} else {
				if diags.HasError() {
					t.Fatalf("Unexpected error: %v", diags)
				}
				if client == nil {
					t.Fatalf("Expected client but got nil")
				}

				// Verify client is properly configured
				hetznerClient, ok := client.(HetznerRobotClient)
				if !ok {
					t.Fatalf("Expected HetznerRobotClient, got %T", client)
				}

				if hetznerClient.username != tt.username {
					t.Fatalf("Expected username '%s', got '%s'", tt.username, hetznerClient.username)
				}
				if hetznerClient.password != tt.password {
					t.Fatalf("Expected password '%s', got '%s'", tt.password, hetznerClient.password)
				}
				if hetznerClient.url != tt.url {
					t.Fatalf("Expected url '%s', got '%s'", tt.url, hetznerClient.url)
				}
			}
		})
	}
}

func TestProviderConfigureWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("HETZNERROBOT_USERNAME", "env_user")
	os.Setenv("HETZNERROBOT_PASSWORD", "env_pass")
	os.Setenv("HETZNERROBOT_URL", "https://env-robot-api.example.com")
	defer func() {
		os.Unsetenv("HETZNERROBOT_USERNAME")
		os.Unsetenv("HETZNERROBOT_PASSWORD")
		os.Unsetenv("HETZNERROBOT_URL")
	}()

	// Create resource data without explicit values (should use env vars)
	resourceData := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{})

	// Test provider configuration
	client, diags := providerConfigure(context.Background(), resourceData)

	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}
	if client == nil {
		t.Fatalf("Expected client but got nil")
	}

	// Verify client uses environment variables
	hetznerClient, ok := client.(HetznerRobotClient)
	if !ok {
		t.Fatalf("Expected HetznerRobotClient, got %T", client)
	}

	if hetznerClient.username != "env_user" {
		t.Fatalf("Expected username from env 'env_user', got '%s'", hetznerClient.username)
	}
	if hetznerClient.password != "env_pass" {
		t.Fatalf("Expected password from env 'env_pass', got '%s'", hetznerClient.password)
	}
	if hetznerClient.url != "https://env-robot-api.example.com" {
		t.Fatalf("Expected url from env 'https://env-robot-api.example.com', got '%s'", hetznerClient.url)
	}
}

func TestProviderConfigureOverridesEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("HETZNERROBOT_USERNAME", "env_user")
	os.Setenv("HETZNERROBOT_PASSWORD", "env_pass")
	defer func() {
		os.Unsetenv("HETZNERROBOT_USERNAME")
		os.Unsetenv("HETZNERROBOT_PASSWORD")
	}()

	// Create resource data with explicit values (should override env vars)
	resourceData := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"username": "config_user",
		"password": "config_pass",
	})

	// Test provider configuration
	client, diags := providerConfigure(context.Background(), resourceData)

	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	// Verify client uses config values, not env vars
	hetznerClient, ok := client.(HetznerRobotClient)
	if !ok {
		t.Fatalf("Expected HetznerRobotClient, got %T", client)
	}

	if hetznerClient.username != "config_user" {
		t.Fatalf("Expected config username 'config_user', got '%s'", hetznerClient.username)
	}
	if hetznerClient.password != "config_pass" {
		t.Fatalf("Expected config password 'config_pass', got '%s'", hetznerClient.password)
	}
}

func TestProviderDefaultURL(t *testing.T) {
	// Create resource data without URL (should use default)
	resourceData := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	})

	// Test provider configuration
	client, diags := providerConfigure(context.Background(), resourceData)

	if diags.HasError() {
		t.Fatalf("Unexpected error: %v", diags)
	}

	// Verify client uses default URL
	hetznerClient, ok := client.(HetznerRobotClient)
	if !ok {
		t.Fatalf("Expected HetznerRobotClient, got %T", client)
	}

	expectedURL := "https://robot-ws.your-server.de"
	if hetznerClient.url != expectedURL {
		t.Fatalf("Expected default URL '%s', got '%s'", expectedURL, hetznerClient.url)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HETZNERROBOT_USERNAME"); v == "" {
		t.Fatal("HETZNERROBOT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("HETZNERROBOT_PASSWORD"); v == "" {
		t.Fatal("HETZNERROBOT_PASSWORD must be set for acceptance tests")
	}
}

func testAccProviders() map[string]*schema.Provider {
	return map[string]*schema.Provider{
		"hetznerrobot": Provider(),
	}
}

func testAccProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"hetznerrobot": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}
