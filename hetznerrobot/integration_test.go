package hetznerrobot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestAuthenticationFlow tests the complete authentication flow from provider config to API call
func TestAuthenticationFlow(t *testing.T) {
	// Mock Hetzner Robot API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Missing basic auth")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if username != "testuser" || password != "testpass" {
			t.Errorf("Wrong credentials: %s:%s", username, password)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Route requests based on path
		switch r.URL.Path {
		case "/server":
			// Mock server list response
			servers := map[string]interface{}{
				"server": []map[string]interface{}{
					{
						"server_ip":     "1.2.3.4",
						"server_number": 12345,
						"server_name":   "test-server",
						"product":       "EX41",
						"dc":           "FSN1-DC1",
						"status":       "ready",
						"canceled":     false,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(servers)

		case "/boot/1.2.3.4":
			// Mock boot configuration response
			boot := map[string]interface{}{
				"boot": map[string]interface{}{
					"linux": map[string]interface{}{
						"active": true,
						"os":     "ubuntu_20.04",
						"arch":   "64",
						"lang":   "en",
					},
					"rescue": map[string]interface{}{
						"active": false,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(boot)

		case "/firewall/1.2.3.4":
			// Mock firewall response
			firewall := map[string]interface{}{
				"firewall": map[string]interface{}{
					"server_ip":     "1.2.3.4",
					"status":        "active",
					"whitelist_hos": true,
					"rules": map[string]interface{}{
						"input": []map[string]interface{}{
							{
								"name":       "SSH",
								"src_ip":     "0.0.0.0/0",
								"dst_port":   "22",
								"action":     "accept",
								"protocol":   "tcp",
								"ip_version": "ipv4",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(firewall)

		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test provider configuration with mock server
	provider := Provider()
	resourceData := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
		"url":      server.URL,
	})

	// Configure provider
	client, diags := providerConfigure(context.Background(), resourceData)
	if diags.HasError() {
		t.Fatalf("Failed to configure provider: %v", diags)
	}

	hetznerClient, ok := client.(HetznerRobotClient)
	if !ok {
		t.Fatalf("Expected HetznerRobotClient, got %T", client)
	}

	// Test actual API calls with authentication
	t.Run("GetServer", func(t *testing.T) {
		server, err := hetznerClient.getServer(context.Background(), 12345)
		if err != nil {
			t.Fatalf("Failed to get server: %v", err)
		}
		if server.ServerIP != "1.2.3.4" {
			t.Fatalf("Expected server IP 1.2.3.4, got %s", server.ServerIP)
		}
	})

	t.Run("GetBoot", func(t *testing.T) {
		boot, err := hetznerClient.getBoot(context.Background(), "1.2.3.4")
		if err != nil {
			t.Fatalf("Failed to get boot config: %v", err)
		}
		if boot.ServerIPv4 != "1.2.3.4" {
			t.Fatalf("Expected boot server IP 1.2.3.4, got %s", boot.ServerIPv4)
		}
	})

	t.Run("GetFirewall", func(t *testing.T) {
		firewall, err := hetznerClient.getFirewall(context.Background(), "1.2.3.4")
		if err != nil {
			t.Fatalf("Failed to get firewall: %v", err)
		}
		if firewall.IP != "1.2.3.4" {
			t.Fatalf("Expected firewall IP 1.2.3.4, got %s", firewall.IP)
		}
		if firewall.Status != "active" {
			t.Fatalf("Expected firewall status active, got %s", firewall.Status)
		}
	})
}

// TestAuthenticationFailure tests what happens when authentication fails
func TestAuthenticationFailure(t *testing.T) {
	// Mock server that always returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	// Create client with any credentials (server will reject them)
	client := NewHetznerRobotClient("wronguser", "wrongpass", server.URL)

	// Test that API calls fail appropriately
	_, err := client.getServer(context.Background(), 12345)
	if err == nil {
		t.Fatal("Expected error for unauthorized request")
	}

	expectedError := "hetzner webservice response status 401"
	if !contains(err.Error(), expectedError) {
		t.Fatalf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

// TestProviderWithVariables simulates the actual usage pattern with Terraform variables
func TestProviderWithVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Mock server that validates specific credentials
	expectedUsername := "myuser"
	expectedPassword := "mypass"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != expectedUsername || password != expectedPassword {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return a simple success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "authenticated"}`))
	}))
	defer server.Close()

	// Test configuration similar to user's setup
	testConfig := fmt.Sprintf(`
		provider "hetznerrobot" {
			username = "%s"
			password = "%s"
			url      = "%s"
		}

		# This would be a data source test, but we'll test the provider config
	`, expectedUsername, expectedPassword, server.URL)

	// Test provider configuration directly
	provider := Provider()
	resourceData := schema.TestResourceDataRaw(t, provider.Schema, map[string]interface{}{
		"username": expectedUsername,
		"password": expectedPassword,
		"url":      server.URL,
	})

	client, diags := providerConfigure(context.Background(), resourceData)
	if diags.HasError() {
		t.Fatalf("Provider configuration failed: %v", diags)
	}

	// Verify we can make an authenticated request
	hetznerClient := client.(HetznerRobotClient)
	_, err := hetznerClient.makeAPICall(context.Background(), "GET", server.URL+"/test", nil, []int{http.StatusOK})
	if err != nil {
		t.Fatalf("Authenticated API call failed: %v", err)
	}

	t.Logf("✅ Authentication test passed with config:\n%s", testConfig)
}

// TestRealWorldScenario tests a scenario closer to real-world usage
func TestRealWorldScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a more comprehensive mock API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authenticate
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}

		// In a real test, you might validate against test credentials
		if username == "" || password == "" {
			http.Error(w, "Invalid auth", http.StatusUnauthorized)
			return
		}

		// Mock various API endpoints
		switch {
		case r.URL.Path == "/server" && r.Method == "GET":
			response := map[string]interface{}{
				"server": []map[string]interface{}{
					{
						"server_ip":     "192.168.1.1",
						"server_number": 54321,
						"server_name":   "production-server",
						"product":       "AX41",
						"dc":           "FSN1-DC14",
						"status":       "ready",
						"canceled":     false,
						"paid_until":   "2024-12-31",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/firewall/192.168.1.1" && r.Method == "GET":
			response := map[string]interface{}{
				"firewall": map[string]interface{}{
					"server_ip":     "192.168.1.1",
					"status":        "active",
					"whitelist_hos": true,
					"rules": map[string]interface{}{
						"input": []map[string]interface{}{
							{
								"name":       "SSH Access",
								"src_ip":     "10.0.0.0/8",
								"dst_port":   "22",
								"action":     "accept",
								"protocol":   "tcp",
								"ip_version": "ipv4",
							},
							{
								"name":       "HTTPS",
								"src_ip":     "0.0.0.0/0",
								"dst_port":   "443",
								"action":     "accept",
								"protocol":   "tcp",
								"ip_version": "ipv4",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test the complete flow
	client := NewHetznerRobotClient("realuser", "realpass", server.URL)

	// Test getting server info
	server_info, err := client.getServer(context.Background(), 54321)
	if err != nil {
		t.Fatalf("Failed to get server info: %v", err)
	}

	if server_info.ServerName != "production-server" {
		t.Fatalf("Expected server name 'production-server', got '%s'", server_info.ServerName)
	}

	// Test getting firewall config
	firewall, err := client.getFirewall(context.Background(), "192.168.1.1")
	if err != nil {
		t.Fatalf("Failed to get firewall: %v", err)
	}

	if len(firewall.Rules.Input) != 2 {
		t.Fatalf("Expected 2 firewall rules, got %d", len(firewall.Rules.Input))
	}

	t.Log("✅ Real-world scenario test passed")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
		 s[len(s)-len(substr):] == substr ||
		 findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
