
package notification

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"service-operation/pocketbase"
)

// parseNotificationIDs parses comma-separated notification IDs
func parseNotificationIDs(notificationID string) []string {
	if notificationID == "" {
		return []string{}
	}
	
	// Split by comma and trim whitespace
	ids := strings.Split(notificationID, ",")
	var cleanIDs []string
	for _, id := range ids {
		cleanID := strings.TrimSpace(id)
		if cleanID != "" {
			cleanIDs = append(cleanIDs, cleanID)
		}
	}
	
	//log.Printf("📋 Parsed notification IDs: %v", cleanIDs)
	return cleanIDs
}

// isNotificationEnabled checks if the notification is enabled
func isNotificationEnabled(pbClient *pocketbase.PocketBaseClient, notificationID string) bool {
	config, err := getAlertConfiguration(pbClient, notificationID)
	if err != nil {
		return false
	}

	// Support both: new "status" field ("enabled"/"disabled") and legacy boolean "enabled" field
	if config.Status != "" {
		return config.Status == "enabled"
	}
	enabled, err := strconv.ParseBool(config.Enabled)
	if err != nil {
		return false
	}
	return enabled
}

// getAlertConfiguration fetches alert configuration from PocketBase
func getAlertConfiguration(pbClient *pocketbase.PocketBaseClient, notificationID string) (*AlertConfiguration, error) {
	url := fmt.Sprintf("%s/api/collections/alert_configurations/records/%s", pbClient.GetBaseURL(), notificationID)

	resp, err := pbClient.GetHTTPClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Failed to fetch alert configuration %s, status: %d", notificationID, resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch alert configuration, status: %d", resp.StatusCode)
	}

	var config AlertConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// getNotificationTemplate fetches server notification template from PocketBase
func getNotificationTemplate(pbClient *pocketbase.PocketBaseClient, templateID string) (*ServerNotificationTemplate, error) {
	url := fmt.Sprintf("%s/api/collections/server_notification_templates/records/%s", pbClient.GetBaseURL(), templateID)

	resp, err := pbClient.GetHTTPClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Failed to fetch notification template %s, status: %d", templateID, resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch notification template, status: %d", resp.StatusCode)
	}

	var template ServerNotificationTemplate
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		return nil, err
	}

	return &template, nil
}

// getServiceNotificationTemplate fetches service notification template from PocketBase
func getServiceNotificationTemplate(pbClient *pocketbase.PocketBaseClient, templateID string) (*ServiceNotificationTemplate, error) {
	url := fmt.Sprintf("%s/api/collections/service_notification_templates/records/%s", pbClient.GetBaseURL(), templateID)

	resp, err := pbClient.GetHTTPClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch service notification template, status: %d", resp.StatusCode)
	}

	var template ServiceNotificationTemplate
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		return nil, err
	}

	return &template, nil
}

// Helper function to get map keys
func getKeys(m map[string]NotificationService) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}