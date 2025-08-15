package hetznerrobot

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewHetznerRobotClient(t *testing.T) {
	username := "testuser"
	password := "testpass"
	url := "https://robot-ws.your-server.de"

	client := NewHetznerRobotClient(username, password, url)

	if client.username != username {
		t.Fatalf("Expected username '%s', got '%s'", username, client.username)
	}
	if client.password != password {
		t.Fatalf("Expected password '%s', got '%s'", password, client.password)
	}
	if client.url != url {
		t.Fatalf("Expected url '%s', got '%s'", url, client.url)
	}
}

func TestMakeAPICallAuthentication(t *testing.T) {
	username := "testuser"
	password := "testpass"

	// Create a test server that verifies authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the Authorization header is present and correct
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Missing Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify it's Basic auth
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Basic auth, got: %s", authHeader)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Decode and verify credentials
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Errorf("Failed to decode auth header: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		expectedCreds := fmt.Sprintf("%s:%s", username, password)
		if string(decoded) != expectedCreds {
			t.Errorf("Expected credentials '%s', got '%s'", expectedCreds, string(decoded))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewHetznerRobotClient(username, password, server.URL)

	// Test GET request
	data, err := client.makeAPICall(context.Background(), "GET", server.URL+"/test", nil, []int{http.StatusOK})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"status": "success"}`
	if string(data) != expected {
		t.Fatalf("Expected response '%s', got '%s'", expected, string(data))
	}
}

func TestMakeAPICallWithFormData(t *testing.T) {
	username := "testuser"
	password := "testpass"

	// Create a test server that verifies form data and auth
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth (simplified check)
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check Content-Type for form data
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected form content type, got: %s", r.Header.Get("Content-Type"))
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Verify form values
		if r.Form.Get("test_param") != "test_value" {
			t.Errorf("Expected test_param=test_value, got: %s", r.Form.Get("test_param"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"form_received": true}`))
	}))
	defer server.Close()

	// Create client
	client := NewHetznerRobotClient(username, password, server.URL)

	// Prepare form data
	formData := url.Values{}
	formData.Set("test_param", "test_value")

	// Test POST request with form data
	data, err := client.makeAPICall(context.Background(), "POST", server.URL+"/test", formData, []int{http.StatusOK})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"form_received": true}`
	if string(data) != expected {
		t.Fatalf("Expected response '%s', got '%s'", expected, string(data))
	}
}

func TestMakeAPICallUnauthorized(t *testing.T) {
	// Create a test server that always returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	// Create client
	client := NewHetznerRobotClient("wronguser", "wrongpass", server.URL)

	// Test request that should fail
	_, err := client.makeAPICall(context.Background(), "GET", server.URL+"/test", nil, []int{http.StatusOK})

	if err == nil {
		t.Fatal("Expected error for unauthorized request")
	}

	expectedError := "hetzner webservice response status 401"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

func TestMakeAPICallWrongCredentials(t *testing.T) {
	correctUsername := "correctuser"
	correctPassword := "correctpass"
	wrongUsername := "wronguser"
	wrongPassword := "wrongpass"

	// Create a test server that checks for correct credentials
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No auth header", http.StatusUnauthorized)
			return
		}

		// Decode credentials
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			http.Error(w, "Bad auth header", http.StatusUnauthorized)
			return
		}

		expectedCreds := fmt.Sprintf("%s:%s", correctUsername, correctPassword)
		if string(decoded) != expectedCreds {
			http.Error(w, "Wrong credentials", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tests := []struct {
		name     string
		username string
		password string
		expectOK bool
	}{
		{
			name:     "correct credentials",
			username: correctUsername,
			password: correctPassword,
			expectOK: true,
		},
		{
			name:     "wrong username",
			username: wrongUsername,
			password: correctPassword,
			expectOK: false,
		},
		{
			name:     "wrong password",
			username: correctUsername,
			password: wrongPassword,
			expectOK: false,
		},
		{
			name:     "wrong both",
			username: wrongUsername,
			password: wrongPassword,
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewHetznerRobotClient(tt.username, tt.password, server.URL)

			_, err := client.makeAPICall(context.Background(), "GET", server.URL+"/test", nil, []int{http.StatusOK})

			if tt.expectOK && err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}
			if !tt.expectOK && err == nil {
				t.Fatal("Expected error but got success")
			}
		})
	}
}

func TestCodeIsInExpected(t *testing.T) {
	tests := []struct {
		statusCode      int
		expectedCodes   []int
		shouldBePresent bool
	}{
		{200, []int{200, 201, 202}, true},
		{201, []int{200, 201, 202}, true},
		{202, []int{200, 201, 202}, true},
		{404, []int{200, 201, 202}, false},
		{500, []int{200, 201, 202}, false},
		{200, []int{}, false},
		{200, []int{200}, true},
	}

	for _, tt := range tests {
		result := codeIsInExpected(tt.statusCode, tt.expectedCodes)
		if result != tt.shouldBePresent {
			t.Errorf("codeIsInExpected(%d, %v) = %v, want %v",
				tt.statusCode, tt.expectedCodes, result, tt.shouldBePresent)
		}
	}
}

func TestMakeAPICallExpectedStatusCodes(t *testing.T) {
	// Create a test server that can return different status codes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for status code in query parameter
		statusStr := r.URL.Query().Get("status")
		if statusStr == "" {
			w.WriteHeader(http.StatusOK)
		} else {
			switch statusStr {
			case "201":
				w.WriteHeader(http.StatusCreated)
			case "202":
				w.WriteHeader(http.StatusAccepted)
			case "404":
				w.WriteHeader(http.StatusNotFound)
			case "500":
				w.WriteHeader(http.StatusInternalServerError)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}
		w.Write([]byte("response"))
	}))
	defer server.Close()

	client := NewHetznerRobotClient("user", "pass", server.URL)

	tests := []struct {
		name          string
		statusCode    string
		expectedCodes []int
		shouldSucceed bool
	}{
		{
			name:          "200 expected and received",
			statusCode:    "",
			expectedCodes: []int{http.StatusOK},
			shouldSucceed: true,
		},
		{
			name:          "201 expected and received",
			statusCode:    "201",
			expectedCodes: []int{http.StatusCreated},
			shouldSucceed: true,
		},
		{
			name:          "multiple expected codes with match",
			statusCode:    "202",
			expectedCodes: []int{http.StatusOK, http.StatusCreated, http.StatusAccepted},
			shouldSucceed: true,
		},
		{
			name:          "404 not expected",
			statusCode:    "404",
			expectedCodes: []int{http.StatusOK},
			shouldSucceed: false,
		},
		{
			name:          "500 not expected",
			statusCode:    "500",
			expectedCodes: []int{http.StatusOK, http.StatusCreated},
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/test"
			if tt.statusCode != "" {
				url += "?status=" + tt.statusCode
			}

			_, err := client.makeAPICall(context.Background(), "GET", url, nil, tt.expectedCodes)

			if tt.shouldSucceed && err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Fatal("Expected error but got success")
			}
		})
	}
}
