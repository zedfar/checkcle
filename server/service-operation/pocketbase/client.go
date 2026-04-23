
package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// authTransport injects Authorization header on every request when token is set.
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		cloned := req.Clone(req.Context())
		cloned.Header.Set("Authorization", "Bearer "+t.token)
		return t.base.RoundTrip(cloned)
	}
	return t.base.RoundTrip(req)
}

type PocketBaseClient struct {
	baseURL   string
	transport *authTransport
	httpClient *http.Client
}

func NewPocketBaseClient(baseURL string) (*PocketBaseClient, error) {
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}

	transport := &authTransport{base: http.DefaultTransport}
	client := &PocketBaseClient{
		baseURL:   baseURL,
		transport: transport,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	return client, nil
}

// Authenticate fetches a superuser token and stores it for all subsequent requests.
func (c *PocketBaseClient) Authenticate(email, password string) error {
	if email == "" || password == "" {
		return nil
	}

	payload, _ := json.Marshal(map[string]string{
		"identity": email,
		"password": password,
	})

	resp, err := http.Post(
		fmt.Sprintf("%s/api/collections/_superusers/auth-with-password", c.baseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("auth response parse failed: %w", err)
	}
	if result.Token == "" {
		return fmt.Errorf("auth response contained no token")
	}

	c.transport.token = result.Token
	return nil
}

func (c *PocketBaseClient) GetBaseURL() string {
	return c.baseURL
}

func (c *PocketBaseClient) TestConnection() error {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/health", c.baseURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PocketBase health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *PocketBaseClient) IsAuthenticated() bool {
	return c.transport.token != ""
}
