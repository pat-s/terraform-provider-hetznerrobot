package hetznerrobot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
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
	return slices.Contains(expectedStatusCodes, statusCode)
}

func (c *HetznerRobotClient) makeAPICall(ctx context.Context, method string, uri string, data url.Values, expectedStatusCodes []int) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}
	request, err := http.NewRequestWithContext(ctx, method, uri, body)
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
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if !codeIsInExpected(response.StatusCode, expectedStatusCodes) {
		return nil, fmt.Errorf("hetzner webservice response status %d: %s", response.StatusCode, responseBytes)
	}

	return responseBytes, nil
}
