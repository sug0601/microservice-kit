package ksqldb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

type queryRequest struct {
	KSQL string `json:"ksql"`
}

type pullQueryResponse struct {
	Schema     string          `json:"@type,omitempty"`
	QueryID    string          `json:"queryId,omitempty"`
	ColumnNames []string       `json:"columnNames,omitempty"`
	ColumnTypes []string       `json:"columnTypes,omitempty"`
	Row        *rowData        `json:"row,omitempty"`
	FinalMessage string        `json:"finalMessage,omitempty"`
}

type rowData struct {
	Columns []interface{} `json:"columns"`
}

// PullQuery executes a pull query and returns all rows
func (c *Client) PullQuery(query string) ([]map[string]interface{}, error) {
	reqBody := queryRequest{KSQL: query}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/query", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.ksql.v1+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse as JSON array
	var rawResponse []json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []map[string]interface{}
	var columnNames []string

	for _, raw := range rawResponse {
		var item map[string]interface{}
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}

		// Parse header to get schema
		if header, ok := item["header"].(map[string]interface{}); ok {
			if schema, ok := header["schema"].(string); ok {
				columnNames = parseSchema(schema)
			}
		}

		// Parse row data
		if row, ok := item["row"].(map[string]interface{}); ok {
			if columns, ok := row["columns"].([]interface{}); ok && len(columnNames) > 0 {
				rowMap := make(map[string]interface{})
				for i, name := range columnNames {
					if i < len(columns) {
						rowMap[name] = columns[i]
					}
				}
				results = append(results, rowMap)
			}
		}
	}

	return results, nil
}

// parseSchema extracts column names from schema string like "`COL1` TYPE, `COL2` TYPE"
func parseSchema(schema string) []string {
	var names []string
	parts := strings.Split(schema, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "`") {
			end := strings.Index(part[1:], "`")
			if end > 0 {
				names = append(names, part[1:end+1])
			}
		}
	}
	return names
}

// PushQuery executes a push query and streams results
func (c *Client) PushQuery(ctx context.Context, query string, handler func(map[string]interface{})) error {
	reqBody := map[string]interface{}{
		"sql": query,
		"properties": map[string]string{
			"auto.offset.reset": "earliest",
		},
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/query-stream", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.ksql.v1+json")
	req.Header.Set("Accept", "application/vnd.ksqlapi.delimited.v1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var columnNames []string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		switch v := data.(type) {
		case map[string]interface{}:
			// Header with column names
			if names, ok := v["columnNames"].([]interface{}); ok {
				columnNames = make([]string, len(names))
				for i, n := range names {
					columnNames[i] = n.(string)
				}
			}
		case []interface{}:
			// Data row
			if len(columnNames) > 0 {
				row := make(map[string]interface{})
				for i, name := range columnNames {
					if i < len(v) {
						row[name] = v[i]
					}
				}
				handler(row)
			}
		}
	}
}

// ExecuteStatement executes a ksqlDB statement (CREATE, INSERT, etc.)
func (c *Client) ExecuteStatement(statement string) error {
	reqBody := queryRequest{KSQL: statement}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/ksql", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.ksql.v1+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("statement failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetServerInfo returns ksqlDB server information
func (c *Client) GetServerInfo() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/info")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return info, nil
}
