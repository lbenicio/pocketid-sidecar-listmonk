package listmonk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Subscriber represents a Listmonk subscriber.
type Subscriber struct {
	ID         int               `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Attributes map[string]any    `json:"attribs"`
	Lists      listIDs           `json:"lists"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// listIDs handles Listmonk's varying list format (objects or ints).
type listIDs []int

func (l *listIDs) UnmarshalJSON(data []byte) error {
	// Try as []int first
	var ints []int
	if err := json.Unmarshal(data, &ints); err == nil {
		*l = ints
		return nil
	}
	// Try as array of objects with "id" field
	var objs []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &objs); err != nil {
		return err
	}
	for _, o := range objs {
		*l = append(*l, o.ID)
	}
	return nil
}

// subscriberListResponse is the paginated response from GET /api/subscribers.
type subscriberListResponse struct {
	Data struct {
		Results []Subscriber `json:"results"`
	} `json:"data"`
}

// createSubscriberPayload is sent to POST /api/subscribers.
type createSubscriberPayload struct {
	Email      string         `json:"email"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	Lists      []int          `json:"lists"`
	Attributes map[string]any `json:"attribs"`
	Preconfirm bool           `json:"preconfirm_subscription"`
}

// Client is a minimal HTTP client for the Listmonk API.
type Client struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

// NewClient returns a new Listmonk API client.
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) basicAuth() string {
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	return "Basic " + auth
}

// ListSubscribersByList returns all subscribers in a specific list.
func (c *Client) ListSubscribersByList(listID int) ([]Subscriber, error) {
	url := c.baseURL + "/api/subscribers?list_id=" + strconv.Itoa(listID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing subscribers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var response subscriberListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding subscribers: %w", err)
	}

	return response.Data.Results, nil
}

// CreateSubscriber creates a new subscriber in the given list.
func (c *Client) CreateSubscriber(email, name string, listID int, pocketID string) (*Subscriber, error) {
	payload := createSubscriberPayload{
		Email:  email,
		Name:   name,
		Status: "enabled",
		Lists:  []int{listID},
		Attributes: map[string]any{
			"pocketid_id": pocketID,
		},
		Preconfirm: true,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/subscribers", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creating subscriber: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var sub Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, fmt.Errorf("decoding created subscriber: %w", err)
	}

	return &sub, nil
}

// UpdateSubscriber updates an existing subscriber's email and name.
func (c *Client) UpdateSubscriber(id int, email, name string) (*Subscriber, error) {
	payload := map[string]any{
		"email": email,
		"name":  name,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	url := c.baseURL + "/api/subscribers/" + strconv.Itoa(id)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("updating subscriber: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d while updating subscriber %d: %s", resp.StatusCode, id, string(respBody))
	}

	var sub Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, fmt.Errorf("decoding updated subscriber: %w", err)
	}

	return &sub, nil
}

// DeleteSubscriber deletes a subscriber by ID.
func (c *Client) DeleteSubscriber(id int) error {
	url := c.baseURL + "/api/subscribers/" + strconv.Itoa(id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("deleting subscriber: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d while deleting subscriber %d: %s", resp.StatusCode, id, string(respBody))
	}

	return nil
}
