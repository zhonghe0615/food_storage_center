package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Order is a json-friendly representation of an order.
type Order struct {
	ID        string `json:"id"`        // order id
	Name      string `json:"name"`      // food name
	Temp      string `json:"temp"`      // ideal temperature
	Freshness int    `json:"freshness"` // freshness in seconds
}

// Action names
const (
	Place   = "place"
	Move    = "move"
	Pickup  = "pickup"
	Discard = "discard"
)

// Action is a json-friendly representation of an action.
type Action struct {
	Timestamp int64  `json:"timestamp"` // unix timestamp in microseconds
	ID        string `json:"id"`        // order id
	Action    string `json:"action"`    // place, move, pickup or discard
}

type options struct {
	Rate int64 `json:"rate"` // inverse rate in microseconds
	Min  int64 `json:"min"`  // min pickup in microseconds
	Max  int64 `json:"max"`  // max pickup in microseconds
}

type solution struct {
	Options options  `json:"options"`
	Actions []Action `json:"actions"`
}

// Client is a client for fetching and solving challenge test problems.
type Client struct {
	endpoint, auth string
}

func NewClient(endpoint, auth string) *Client {
	return &Client{endpoint: endpoint, auth: auth}
}

// New fetches a new test problem from the server. The URL also works in a browser for convenience.
func (c *Client) New(name string, seed int64) (string, []Order, error) {
	if seed == 0 {
		seed = rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	}

	url := fmt.Sprintf("%v/new?auth=%v&name=%v&seed=%v", c.endpoint, c.auth, name, seed)

	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("%v: %v", url, resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read body: %v", err)
	}
	var orders []Order
	if err := json.Unmarshal(buf, &orders); err != nil {
		return "", nil, fmt.Errorf("failed to deserialize '%v': %v", string(buf), err)
	}
	id := resp.Header.Get("x-test-id")

	log.Printf("Fetched new test problem, id=%v: %v", id, url)
	return id, orders, nil
}

// Solve submits a sequence of actions and parameters as a solution to a test problem. Returns test result.
func (c *Client) Solve(id string, rate, min, max time.Duration, actions []Action) (string, error) {
	url := fmt.Sprintf("%v/solve?auth=%v", c.endpoint, c.auth)

	payload := solution{
		Options: options{
			Rate: rate.Microseconds(),
			Min:  min.Microseconds(),
			Max:  max.Microseconds(),
		},
		Actions: actions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Add("x-test-id", id)
	req.Header.Add("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%v: %v", url, resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
