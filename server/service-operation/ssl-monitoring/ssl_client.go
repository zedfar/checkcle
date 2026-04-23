
package sslmonitoring

import (
	"encoding/json"
	"fmt"
	"net/http"

	"service-operation/pocketbase"
)

// SSLClient handles SSL certificate data operations
type SSLClient struct {
	pbClient *pocketbase.PocketBaseClient
}

// NewSSLClient creates a new SSL client
func NewSSLClient(pbClient *pocketbase.PocketBaseClient) *SSLClient {
	return &SSLClient{
		pbClient: pbClient,
	}
}

// GetSSLCertificates fetches all SSL certificates from PocketBase
func (sc *SSLClient) GetSSLCertificates() ([]SSLCertificate, error) {
	url := fmt.Sprintf("%s/api/collections/ssl_certificates/records", sc.pbClient.GetBaseURL())
	// log.Printf("🔍 [SSL-CLIENT] Fetching SSL certificates from: %s", url)
	
	resp, err := sc.pbClient.GetHTTPClient().Get(url)
	if err != nil {
		// log.Printf("❌ [SSL-CLIENT] HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// log.Printf("❌ [SSL-CLIENT] Failed to fetch SSL certificates, status: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch SSL certificates, status: %d", resp.StatusCode)
	}

	var response struct {
		Items []SSLCertificate `json:"items"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// log.Printf("❌ [SSL-CLIENT] Error decoding response: %v", err)
		return nil, err
	}

	// log.Printf("✅ [SSL-CLIENT] Successfully fetched %d SSL certificates", len(response.Items))
	return response.Items, nil
}

// UpdateSSLCertificate updates SSL certificate status and notification time
func (sc *SSLClient) UpdateSSLCertificate(certID string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/collections/ssl_certificates/records/%s", sc.pbClient.GetBaseURL(), certID)
	
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSL certificate update request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.pbClient.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to update SSL certificate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update SSL certificate, status: %d", resp.StatusCode)
	}

	// log.Printf("✅ [SSL-CLIENT] Successfully updated SSL certificate: %s", certID)
	return nil
}