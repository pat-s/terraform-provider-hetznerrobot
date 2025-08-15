package hetznerrobot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type HetznerRobotClient struct {
	username string
	password string
	url      string
}

func NewHetznerRobotClient(username string, password string, url string) HetznerRobotClient {
	return HetznerRobotClient{
		username: username,
		password: password,
		url:      url,
	}
}

func codeIsInExpected(statusCode int, expectedStatusCodes []int) bool {
	for _, expectedStatusCode := range expectedStatusCodes {
		if statusCode == expectedStatusCode {
			return true
		}
	}
	return false
}

func (c *HetznerRobotClient) makeAPICall(ctx context.Context, method string, uri string, data url.Values, expectedStatusCodes []int) ([]byte, error) {
	tflog.Debug(ctx, "requesting Hetzner webservice", map[string]any{
		"uri":    uri,
		"method": method,
		"data":   data,
	})

	request, err := http.NewRequestWithContext(ctx, method, uri, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	if data != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	request.SetBasicAuth(c.username, c.password)

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "got hetzner webservice response", map[string]any{
		"status": response.StatusCode,
		"body":   string(responseBytes),
	})

	if !codeIsInExpected(response.StatusCode, expectedStatusCodes) {
		return nil, fmt.Errorf("hetzner webservice response status %d: %s", response.StatusCode, responseBytes)
	}

	return responseBytes, nil
}
