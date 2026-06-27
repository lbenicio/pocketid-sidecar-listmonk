package pocketid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// User represents a PocketID user as returned by the admin API.
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
}

// userListResponse is the paginated response from GET /api/users.
type userListResponse struct {
	Data       []User `json:"data"`
	Pagination struct {
		TotalPages   int `json:"totalPages"`
		TotalItems   int `json:"totalItems"`
		CurrentPage  int `json:"currentPage"`
		ItemsPerPage int `json:"itemsPerPage"`
	} `json:"pagination"`
}

// Client is a minimal HTTP client for the PocketID admin API.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient returns a new PocketID API client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListUsers fetches all users from PocketID (handles pagination).
func (c *Client) ListUsers() ([]User, error) {
	var allUsers []User
	page := 1

	for {
		url := fmt.Sprintf("%s/api/users?page=%d&per_page=100", c.baseURL, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("listing users: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d while listing users", resp.StatusCode)
		}

		var response userListResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding users response: %w", err)
		}
		resp.Body.Close()

		allUsers = append(allUsers, response.Data...)

		if page >= response.Pagination.TotalPages {
			break
		}
		page++
	}

	return allUsers, nil
}
