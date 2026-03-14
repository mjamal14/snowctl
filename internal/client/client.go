package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SNClient is the ServiceNow Table API client.
type SNClient struct {
	httpClient *http.Client
	baseURL    string
	auth       Authenticator
	userAgent  string
	debug      bool
}

// ListOptions maps to ServiceNow Table API query parameters.
type ListOptions struct {
	Query        string
	Fields       []string
	Limit        int
	Offset       int
	DisplayValue string
	OrderBy      string
	OrderByDesc  string
}

// TableResult is the parsed response from the Table API.
type TableResult struct {
	Records    []map[string]interface{}
	TotalCount int
	HasMore    bool
}

// New creates a new ServiceNow API client.
func New(baseURL string, auth Authenticator) *SNClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &SNClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		auth:       auth,
		userAgent:  "snowctl/0.1.0",
	}
}

// SetDebug enables debug logging of HTTP requests.
func (c *SNClient) SetDebug(debug bool) {
	c.debug = debug
}

// List retrieves records from a ServiceNow table.
func (c *SNClient) List(table string, opts ListOptions) (*TableResult, error) {
	u := c.tableURL(table)
	q := c.buildQueryParams(opts)
	u.RawQuery = q.Encode()

	resp, body, err := c.doRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp, body)
	}

	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	totalCount := len(result.Result)
	if tc := resp.Header.Get("X-Total-Count"); tc != "" {
		if n, err := strconv.Atoi(tc); err == nil {
			totalCount = n
		}
	}

	hasMore := false
	if opts.Limit > 0 {
		hasMore = totalCount > opts.Offset+opts.Limit
	}

	return &TableResult{
		Records:    result.Result,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// Get retrieves a single record by sys_id.
func (c *SNClient) Get(table string, sysID string, opts ListOptions) (map[string]interface{}, error) {
	u := c.tableURL(table)
	u.Path += "/" + sysID
	q := url.Values{}
	if len(opts.Fields) > 0 {
		q.Set("sysparm_fields", strings.Join(opts.Fields, ","))
	}
	if opts.DisplayValue != "" {
		q.Set("sysparm_display_value", opts.DisplayValue)
	}
	u.RawQuery = q.Encode()

	resp, body, err := c.doRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp, body)
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Result, nil
}

// Create creates a new record in the given table.
func (c *SNClient) Create(table string, data map[string]interface{}) (map[string]interface{}, error) {
	u := c.tableURL(table)

	resp, body, err := c.doRequest("POST", u.String(), data)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, parseAPIError(resp, body)
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Result, nil
}

// Update replaces a record entirely (PUT).
func (c *SNClient) Update(table string, sysID string, data map[string]interface{}) (map[string]interface{}, error) {
	u := c.tableURL(table)
	u.Path += "/" + sysID

	resp, body, err := c.doRequest("PUT", u.String(), data)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp, body)
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Result, nil
}

// Patch partially updates a record (PATCH).
func (c *SNClient) Patch(table string, sysID string, data map[string]interface{}) (map[string]interface{}, error) {
	u := c.tableURL(table)
	u.Path += "/" + sysID

	resp, body, err := c.doRequest("PATCH", u.String(), data)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp, body)
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Result, nil
}

// Delete removes a record by sys_id.
func (c *SNClient) Delete(table string, sysID string) error {
	u := c.tableURL(table)
	u.Path += "/" + sysID

	resp, body, err := c.doRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseAPIError(resp, body)
	}

	return nil
}

func (c *SNClient) tableURL(table string) *url.URL {
	u, _ := url.Parse(c.baseURL + "/api/now/table/" + table)
	return u
}

func (c *SNClient) buildQueryParams(opts ListOptions) url.Values {
	q := url.Values{}
	if opts.Query != "" {
		q.Set("sysparm_query", opts.Query)
	}
	if len(opts.Fields) > 0 {
		q.Set("sysparm_fields", strings.Join(opts.Fields, ","))
	}
	if opts.Limit > 0 {
		q.Set("sysparm_limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("sysparm_offset", strconv.Itoa(opts.Offset))
	}
	if opts.DisplayValue != "" {
		q.Set("sysparm_display_value", opts.DisplayValue)
	}
	if opts.OrderBy != "" {
		q.Set("sysparm_orderby", opts.OrderBy)
	}
	if opts.OrderByDesc != "" {
		q.Set("ORDERBYDESCSYSPARM", opts.OrderByDesc)
	}
	q.Set("sysparm_exclude_reference_link", "true")
	return q
}

func (c *SNClient) doRequest(method, url string, body interface{}) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if err := c.auth.Apply(req); err != nil {
		return nil, nil, fmt.Errorf("applying auth: %w", err)
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", method, url)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("reading response: %w", err)
	}

	return resp, respBody, nil
}
