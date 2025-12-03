package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Confluence Cloud REST API with simple helpers.
type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

// NewClient builds a Confluence API client using the provided base URL and bearer token.
func NewClient(baseURL, authToken string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// SearchResult describes a single Confluence content hit.
type SearchResult struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Space   string `json:"space"`
	Excerpt string `json:"excerpt"`
}

// Page represents a fetched Confluence page with body storage rendered as XHTML.
type Page struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Space struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"space"`
	Body struct {
		Storage struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// Search executes a CQL query scoped to the provided spaces.
func (c *Client) Search(ctx context.Context, query string, spaces []string, limit int) ([]SearchResult, error) {
	cql := fmt.Sprintf("text ~ \"%s\"", escapeQuotes(query))
	if len(spaces) > 0 {
		quoted := make([]string, len(spaces))
		for i, s := range spaces {
			quoted[i] = fmt.Sprintf("\"%s\"", s)
		}
		cql = fmt.Sprintf("%s AND space in (%s)", cql, strings.Join(quoted, ","))
	}

	values := url.Values{}
	values.Set("cql", cql)
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/wiki/rest/api/search", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("confluence search failed: %s", resp.Status)
	}

	var payload struct {
		Results []struct {
			Content struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Links struct {
					WebUI string `json:"webui"`
				} `json:"_links"`
				Space struct {
					Key string `json:"key"`
				} `json:"space"`
			} `json:"content"`
			Excerpt string `json:"excerpt"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(payload.Results))
	for _, item := range payload.Results {
		results = append(results, SearchResult{
			ID:      item.Content.ID,
			Title:   item.Content.Title,
			URL:     c.baseURL + item.Content.Links.WebUI,
			Space:   item.Content.Space.Key,
			Excerpt: item.Excerpt,
		})
	}

	return results, nil
}

// GetPage retrieves a page including its storage format body for downstream processing.
func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/wiki/rest/api/content/%s", c.baseURL, pageID), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("expand", "body.storage,space")
	req.URL.RawQuery = q.Encode()

	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("confluence fetch failed: %s", resp.Status)
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) applyAuth(req *http.Request) {
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	req.Header.Set("Accept", "application/json")
}

func escapeQuotes(input string) string {
	return strings.ReplaceAll(input, "\"", "\\\"")
}
