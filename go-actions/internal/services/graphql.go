package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func ExecuteGraphQLRequest(req GraphQLRequest) (json.RawMessage, error) {
	hasuraEndpoint := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	adminSecret := os.Getenv("HASURA_ADMIN_SECRET")

	if hasuraEndpoint == "" || adminSecret == "" {
		return nil, errors.New("Hasura GraphQL endpoint or admin secret is not configured")
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", hasuraEndpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-hasura-admin-secret", adminSecret)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Hasura response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var gqlResponse GraphQLResponse
	if err := json.Unmarshal(bodyBytes, &gqlResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(gqlResponse.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL API returned error: %s", gqlResponse.Errors[0].Message)
	}

	return gqlResponse.Data, nil
}
