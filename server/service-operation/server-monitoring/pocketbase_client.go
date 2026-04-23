package servermonitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"service-operation/pocketbase"
)

// ServerMetrics represents server metrics record in PocketBase
type ServerMetrics struct {
	ID             string `json:"id"`
	ServerID       string `json:"server_id"`
	Timestamp      string `json:"timestamp"` // Changed to string for custom parsing
	RAMTotal       string `json:"ram_total"`
	RAMUsed        string `json:"ram_used"`
	RAMFree        string `json:"ram_free"`
	CPUCores       string `json:"cpu_cores"`
	CPUUsage       string `json:"cpu_usage"`
	CPUFree        string `json:"cpu_free"`
	DiskTotal      string `json:"disk_total"`
	DiskUsed        string `json:"disk_used"`
	DiskFree        string `json:"disk_free"`
	Status         string `json:"status"`
	NetworkRxBytes int64  `json:"network_rx_bytes"`
	NetworkTxBytes int64  `json:"network_tx_bytes"`
	NetworkRxSpeed int64  `json:"network_rx_speed"`
	NetworkTxSpeed int64  `json:"network_tx_speed"`
	Created        string `json:"created"` // Changed to string for custom parsing
	Updated        string `json:"updated"` // Changed to string for custom parsing
}

// ParsedServerMetrics represents ServerMetrics with parsed time fields
type ParsedServerMetrics struct {
	ServerMetrics
	CreatedTime   time.Time
	UpdatedTime   time.Time
	TimestampTime time.Time
}

// ParseServerMetrics converts ServerMetrics to ParsedServerMetrics with proper time parsing
func ParseServerMetrics(sm ServerMetrics) (ParsedServerMetrics, error) {
	psm := ParsedServerMetrics{ServerMetrics: sm}
	
	var err error
	
	// Parse Created time
	if sm.Created != "" {
		psm.CreatedTime, err = parsePocketBaseTime(sm.Created)
		if err != nil {
			log.Printf("Warning: Failed to parse Created time '%s': %v", sm.Created, err)
		}
	}
	
	// Parse Updated time
	if sm.Updated != "" {
		psm.UpdatedTime, err = parsePocketBaseTime(sm.Updated)
		if err != nil {
			log.Printf("Warning: Failed to parse Updated time '%s': %v", sm.Updated, err)
		}
	}
	
	// Parse Timestamp
	if sm.Timestamp != "" {
		psm.TimestampTime, err = parsePocketBaseTime(sm.Timestamp)
		if err != nil {
			log.Printf("Warning: Failed to parse Timestamp '%s': %v", sm.Timestamp, err)
		}
	}
	
	return psm, nil
}

// parsePocketBaseTime parses PocketBase time format "2025-08-11 13:09:13.243Z"
func parsePocketBaseTime(timeStr string) (time.Time, error) {
	// First try the PocketBase format with space separator
	if t, err := time.Parse("2006-01-02 15:04:05.000Z", timeStr); err == nil {
		return t, nil
	}
	
	// Fallback to RFC3339 format with T separator
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}
	
	// Fallback to RFC3339Nano format
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

// ServerPocketBaseClient is a wrapper around the PocketBase client for server monitoring
type ServerPocketBaseClient struct {
	client *pocketbase.PocketBaseClient
}

// NewServerPocketBaseClient creates a new server PocketBase client
func NewServerPocketBaseClient(client *pocketbase.PocketBaseClient) *ServerPocketBaseClient {
	return &ServerPocketBaseClient{
		client: client,
	}
}

// GetServerThreshold fetches threshold configuration for a server
func (spc *ServerPocketBaseClient) GetServerThreshold(thresholdID string) (*ServerThreshold, error) {
	if thresholdID == "" {
		log.Printf("No threshold ID provided")
		_ = thresholdID
		return nil, nil
	}

	url := fmt.Sprintf("%s/api/collections/server_threshold_templates/records/%s", spc.client.GetBaseURL(), thresholdID)
	//log.Printf("🔍 Fetching server threshold from: %s", url)

	resp, err := spc.client.GetHTTPClient().Get(url)
	if err != nil {
		//log.Printf("❌ HTTP error fetching server threshold: %v", err)
		_ = err
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//log.Printf("❌ Failed to fetch server threshold, status: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch server threshold, status: %d", resp.StatusCode)
	}

	var threshold ServerThreshold
	if err := json.NewDecoder(resp.Body).Decode(&threshold); err != nil {
		//log.Printf("❌ Error decoding server threshold JSON: %v", err)
		_ = err
		return nil, err
	}

	//log.Printf("✅ Successfully fetched server threshold: %+v", threshold)
	_ = url
	_ = threshold
	return &threshold, nil
}

// GetAllServers retrieves all servers from PocketBase
func (spc *ServerPocketBaseClient) GetAllServers() ([]Server, error) {
	url := fmt.Sprintf("%s/api/collections/servers/records?perPage=500", spc.client.GetBaseURL())
	//log.Printf("🌐 Fetching all servers from: %s", url)

	resp, err := spc.client.GetHTTPClient().Get(url)
	if err != nil {
		//log.Printf("❌ HTTP error fetching servers: %v", err)
		_ = err
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//log.Printf("❌ Failed to fetch servers, status: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch servers, status: %d", resp.StatusCode)
	}

	var response struct {
		Items []Server `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		//log.Printf("❌ Error decoding servers JSON: %v", err)
		_ = err
		return nil, err
	}

	//log.Printf("✅ Successfully fetched %d servers", len(response.Items))
	_ = url
	return response.Items, nil
}

// GetLatestServerMetrics retrieves the latest server metrics from PocketBase
func (spc *ServerPocketBaseClient) GetLatestServerMetrics(serverID string, timeout time.Duration) ([]ParsedServerMetrics, error) {
	// Calculate the time before which metrics are considered too old
	cutoff := time.Now().Add(-timeout).UTC().Format("2006-01-02 15:04:05.000Z")

	// Construct the filter string with proper formatting (no spaces around operators)
	filter := fmt.Sprintf("server_id='%s'&&created>'%s'", serverID, cutoff)
	
	// URL encode the filter parameter
	encodedFilter := url.QueryEscape(filter)

	// Construct the URL with properly encoded parameters
	requestURL := fmt.Sprintf("%s/api/collections/server_metrics/records?filter=%s&sort=-created&perPage=1", 
		spc.client.GetBaseURL(), encodedFilter)
	
	//log.Printf("🌐 Fetching latest server metrics from: %s", requestURL)
	//log.Printf("🔍 Filter used: %s", filter)
	//log.Printf("🔍 Cutoff time: %s", cutoff)

	resp, err := spc.client.GetHTTPClient().Get(requestURL)
	if err != nil {
		//log.Printf("❌ HTTP error fetching server metrics: %v", err)
		_ = err
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//log.Printf("❌ Failed to fetch server metrics, status: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch server metrics, status: %d", resp.StatusCode)
	}

	var response struct {
		Items []ServerMetrics `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		//log.Printf("❌ Error decoding server metrics JSON: %v", err)
		_ = err
		return nil, err
	}

	// Convert to ParsedServerMetrics
	var parsedMetrics []ParsedServerMetrics
	for _, metric := range response.Items {
		parsed, err := ParseServerMetrics(metric)
		if err != nil {
			//log.Printf("❌ Error parsing server metric: %v", err)
			_ = err
			continue
		}
		parsedMetrics = append(parsedMetrics, parsed)
	}

	//log.Printf("✅ Successfully fetched and parsed %d server metrics", len(parsedMetrics))
	_ = requestURL
	_ = filter
	_ = cutoff
	return parsedMetrics, nil
}

// UpdateServerStatus updates the server status in PocketBase
func (spc *ServerPocketBaseClient) UpdateServerStatus(serverID string, status string) error {
	url := fmt.Sprintf("%s/api/collections/servers/records/%s", spc.client.GetBaseURL(), serverID)
	//log.Printf("🌐 Updating server status at: %s", url)

	payload := map[string]interface{}{
		"status":       status,
		"last_checked": time.Now().UTC().Format("2006-01-02 15:04:05.000Z"),
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		//log.Printf("❌ Error marshaling payload: %v", err)
		_ = err
		return err
	}

	//log.Printf("📝 Update payload: %s", string(payloadBytes))

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		//log.Printf("❌ Error creating request: %v", err)
		_ = err
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		//log.Printf("❌ HTTP error updating server status: %v", err)
		_ = err
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//log.Printf("❌ Failed to update server status, status: %d", resp.StatusCode)
		return fmt.Errorf("failed to update server status, status: %d", resp.StatusCode)
	}

	//log.Printf("✅ Successfully updated server status to %s", status)
	_ = url
	_ = payloadBytes
	_ = status
	return nil
}