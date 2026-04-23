package maintenancemonitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"service-operation/notification"
	"service-operation/pocketbase"
)

// MaintenanceMonitor checks maintenance windows and sends notifications
type MaintenanceMonitor struct {
	pbClient      *pocketbase.PocketBaseClient
	checkInterval time.Duration
	stopChan      chan bool
	telegram      *notification.TelegramService
}

// NewMaintenanceMonitor creates a new MaintenanceMonitor
func NewMaintenanceMonitor(pbClient *pocketbase.PocketBaseClient) *MaintenanceMonitor {
	return &MaintenanceMonitor{
		pbClient:      pbClient,
		checkInterval: 60 * time.Second,
		stopChan:      make(chan bool, 1),
		telegram:      notification.NewTelegramService(),
	}
}

// Start begins the maintenance monitoring loop
func (mm *MaintenanceMonitor) Start() {
	mm.check()

	ticker := time.NewTicker(mm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.check()
		case <-mm.stopChan:
			return
		}
	}
}

// Stop gracefully stops the monitor
func (mm *MaintenanceMonitor) Stop() {
	select {
	case mm.stopChan <- true:
	default:
	}
}

// check fetches active maintenance records and processes each one
func (mm *MaintenanceMonitor) check() {
	filter := url.QueryEscape(`status = "scheduled" || status = "in_progress"`)
	apiURL := fmt.Sprintf("%s/api/collections/maintenance/records?filter=%s&perPage=100",
		mm.pbClient.GetBaseURL(), filter)

	resp, err := mm.pbClient.GetHTTPClient().Get(apiURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var listResp MaintenanceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return
	}

	now := time.Now()
	for _, record := range listResp.Items {
		mm.process(record, now)
	}
}

// process evaluates a single maintenance record and acts if needed
func (mm *MaintenanceMonitor) process(m MaintenanceRecord, now time.Time) {
	startTime, err := parseTime(m.StartTime)
	if err != nil {
		return
	}
	endTime, err := parseTime(m.EndTime)
	if err != nil {
		return
	}

	// scheduled → in_progress: start_time has passed, not yet notified
	if m.Status == "scheduled" && !now.Before(startTime) && now.Before(endTime) && !m.NotifiedStart {
		if err := mm.updateStatus(m.ID, "in_progress"); err != nil {
			log.Printf("[MAINTENANCE] Failed to update status to in_progress for %s: %v", m.ID, err)
			return
		}
		mm.sendNotification(m, "start")
		mm.setFlag(m.ID, "notified_start", true)
	}

	// in_progress → completed: end_time has passed, not yet notified
	if m.Status == "in_progress" && !now.Before(endTime) && !m.NotifiedEnd {
		if err := mm.updateStatus(m.ID, "completed"); err != nil {
			log.Printf("[MAINTENANCE] Failed to update status to completed for %s: %v", m.ID, err)
			return
		}
		mm.sendNotification(m, "end")
		mm.setFlag(m.ID, "notified_end", true)
	}
}

// updateStatus patches the status field of a maintenance record
func (mm *MaintenanceMonitor) updateStatus(id, status string) error {
	apiURL := fmt.Sprintf("%s/api/collections/maintenance/records/%s",
		mm.pbClient.GetBaseURL(), id)

	body, _ := json.Marshal(map[string]string{"status": status})
	req, err := http.NewRequest(http.MethodPatch, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := mm.pbClient.GetHTTPClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// setFlag patches a boolean idempotency flag on the maintenance record
func (mm *MaintenanceMonitor) setFlag(id, field string, value bool) {
	apiURL := fmt.Sprintf("%s/api/collections/maintenance/records/%s",
		mm.pbClient.GetBaseURL(), id)

	body, _ := json.Marshal(map[string]bool{field: value})
	req, err := http.NewRequest(http.MethodPatch, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := mm.pbClient.GetHTTPClient().Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// sendNotification resolves the alert config and sends a Telegram message
func (mm *MaintenanceMonitor) sendNotification(m MaintenanceRecord, notifType string) {
	if m.NotifySubscribers != "yes" {
		return
	}

	channelID := m.NotificationChannelID
	if channelID == "" {
		channelID = m.NotificationID
	}
	if channelID == "" {
		return
	}

	configURL := fmt.Sprintf("%s/api/collections/alert_configurations/records/%s",
		mm.pbClient.GetBaseURL(), channelID)

	resp, err := mm.pbClient.GetHTTPClient().Get(configURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var config notification.AlertConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return
	}

	// Check enabled (support both status field and legacy bool)
	enabled := config.Status == "enabled"
	if config.Status == "" {
		enabled = strings.EqualFold(config.Enabled, "true")
	}
	if !enabled {
		return
	}

	if config.NotificationType != "telegram" {
		return
	}

	message := buildMessage(m, notifType)
	if err := mm.telegram.SendNotification(&config, message); err != nil {
		log.Printf("[MAINTENANCE] Telegram notification failed for %s: %v", m.ID, err)
	}
}

// buildMessage generates the Telegram HTML message for a maintenance event
func buildMessage(m MaintenanceRecord, notifType string) string {
	var emoji, statusText, timeText string

	switch notifType {
	case "start":
		emoji = "⚠️"
		statusText = "has started"
		timeText = fmt.Sprintf("Scheduled until: %s", formatTime(m.EndTime))
	case "end":
		emoji = "✅"
		statusText = "has completed"
		timeText = "All systems are back to normal operation"
	}

	affected := strings.Join(
		filterEmpty(strings.Split(m.Affected, ",")),
		", ",
	)

	return fmt.Sprintf(
		"%s <b>Maintenance %s</b>\n\n<b>Title:</b> %s\n<b>Description:</b> %s\n<b>Affected Services:</b> %s\n<b>%s</b>\n\n<b>Priority:</b> %s\n<b>Impact:</b> %s",
		emoji, statusText,
		m.Title,
		m.Description,
		affected,
		timeText,
		strings.ToUpper(m.Priority),
		strings.ToUpper(m.Field),
	)
}

// parseTime parses PocketBase datetime strings (supports space and T separator)
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05.000Z",
		"2006-01-02 15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %q", s)
}

// formatTime formats a PocketBase datetime string for display
func formatTime(s string) string {
	t, err := parseTime(s)
	if err != nil {
		return s
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

// filterEmpty removes empty/whitespace-only strings from a slice
func filterEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
