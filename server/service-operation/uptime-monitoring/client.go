
package uptimemonitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"service-operation/pocketbase"
)

// UptimeClient handles uptime service operations
type UptimeClient struct {
	pbClient *pocketbase.PocketBaseClient
}

// MetricRecord represents a metric record from any collection
type MetricRecord struct {
	ID           string    `json:"id"`
	ServiceID    string    `json:"service_id"`
	Status       string    `json:"status"`
	ResponseTime int       `json:"response_time"`
	ErrorMessage string    `json:"error_message"` // Added missing ErrorMessage field
	Timestamp    time.Time `json:"timestamp"`
	Created      time.Time `json:"created"`
}

// NewUptimeClient creates a new uptime client
func NewUptimeClient(pbClient *pocketbase.PocketBaseClient) *UptimeClient {
	return &UptimeClient{
		pbClient: pbClient,
	}
}

// GetServices retrieves all services for monitoring
func (uc *UptimeClient) GetServices() ([]UptimeService, error) {
	url := fmt.Sprintf("%s/api/collections/services/records", uc.pbClient.GetBaseURL())
	
	resp, err := uc.pbClient.GetHTTPClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch services, status: %d", resp.StatusCode)
	}

	var response struct {
		Items []UptimeService `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode services: %v", err)
	}

	// log.Printf("✅ Fetched %d services for monitoring", len(response.Items))
	return response.Items, nil
}

// GetLatestMetricRecord gets the latest metric record for a service from specified collection
func (uc *UptimeClient) GetLatestMetricRecord(serviceID, collection string) ([]MetricRecord, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records?filter=service_id='%s'&sort=-timestamp&perPage=1", 
		uc.pbClient.GetBaseURL(), collection, serviceID)
	
	// log.Printf("🔍 [METRICS-QUERY] Getting latest record for service %s from %s", serviceID, collection)
	
	resp, err := uc.pbClient.GetHTTPClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metric record: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch metric record, status: %d", resp.StatusCode)
	}

	var response struct {
		Items []struct {
			ID           string `json:"id"`
			ServiceID    string `json:"service_id"`
			Status       string `json:"status"`
			ResponseTime int    `json:"response_time"`
			ErrorMessage string `json:"error_message"` // Added missing ErrorMessage field
			Timestamp    string `json:"timestamp"`
			Created      string `json:"created"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode metric record: %v", err)
	}

	var records []MetricRecord
	for _, item := range response.Items {
		timestamp, _ := time.Parse(time.RFC3339, item.Timestamp)
		created, _ := time.Parse(time.RFC3339, item.Created)
		
		records = append(records, MetricRecord{
			ID:           item.ID,
			ServiceID:    item.ServiceID,
			Status:       item.Status,
			ResponseTime: item.ResponseTime,
			ErrorMessage: item.ErrorMessage, // Include ErrorMessage in the record
			Timestamp:    timestamp,
			Created:      created,
		})
	}

	// log.Printf("📊 [METRICS-RESULT] Found %d records for service %s in %s", len(records), serviceID, collection)
	return records, nil
}