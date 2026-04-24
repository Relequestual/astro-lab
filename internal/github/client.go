package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Relequestual/astro-lab/internal/auth"
)

const graphqlEndpoint = "https://api.github.com/graphql"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

func (e GraphQLError) Error() string {
	return auth.RedactSecrets(e.Message)
}

func (c *Client) Do(ctx context.Context, req GraphQLRequest) (*GraphQLResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "astro-lab/0.1")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, &AuthError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if resp.StatusCode != 200 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: auth.RedactSecrets(string(respBody))}
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, &QueryError{Errors: gqlResp.Errors}
	}

	return &gqlResp, nil
}

// Error types

type AuthError struct {
	StatusCode int
	Body       string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed (HTTP %d): %s", e.StatusCode, auth.RedactSecrets(e.Body))
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.StatusCode, e.Body)
}

type QueryError struct {
	Errors []GraphQLError
}

func (e *QueryError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("GraphQL error: %s", e.Errors[0].Error())
	}
	return "unknown GraphQL error"
}
