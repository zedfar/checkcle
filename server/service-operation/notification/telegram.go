
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// TelegramService handles Telegram notifications
type TelegramService struct{}

// NewTelegramService creates a new Telegram notification service
func NewTelegramService() *TelegramService {
	return &TelegramService{}
}

// TelegramPayload represents the payload for Telegram API
type TelegramPayload struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	ParseMode       string `json:"parse_mode,omitempty"`
	MessageThreadID int    `json:"message_thread_id,omitempty"`
}

// SendNotification sends a notification via Telegram
func (ts *TelegramService) SendNotification(config *AlertConfiguration, message string) error {
	// fmt.Printf("📱 [TELEGRAM] Attempting to send notification...\n")
	// fmt.Printf("📱 [TELEGRAM] Config - Chat ID: %s, Bot Token present: %v\n", config.TelegramChatID, config.BotToken != "")
	// fmt.Printf("📱 [TELEGRAM] Message: %s\n", message)

	if config.BotToken == "" || config.TelegramChatID == "" {
		return fmt.Errorf("telegram bot token and chat ID are required")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
	// fmt.Printf("📱 [TELEGRAM] API URL: %s\n", strings.Replace(url, config.BotToken, "[REDACTED]", 1))

	payload := TelegramPayload{
		ChatID: config.TelegramChatID,
		Text:   message,
	}

	if config.TelegramThreadID != "" {
		if threadID, err := strconv.Atoi(config.TelegramThreadID); err == nil {
			payload.MessageThreadID = threadID
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// fmt.Printf("📱 [TELEGRAM] Sending POST request...\n")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// fmt.Printf("❌ [TELEGRAM] HTTP error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	var result NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// fmt.Printf("❌ [TELEGRAM] JSON decode error: %v\n", err)
		return err
	}

	// fmt.Printf("📱 [TELEGRAM] Response - OK: %v, Description: %s\n", result.OK, result.Description)

	if !result.OK {
		// fmt.Printf("❌ [TELEGRAM] API error: %s\n", result.Description)
		return fmt.Errorf("telegram API error: %s", result.Description)
	}

	// fmt.Printf("✅ [TELEGRAM] Message sent successfully!\n")
	return nil
}

// SendServerNotification sends a server-specific notification via Telegram
func (ts *TelegramService) SendServerNotification(config *AlertConfiguration, payload *NotificationPayload, template *ServerNotificationTemplate, resourceType string) error {
	message := ts.generateServerMessage(payload, template, resourceType)
	return ts.SendNotification(config, message)
}

// SendServiceNotification sends a service-specific notification via Telegram
func (ts *TelegramService) SendServiceNotification(config *AlertConfiguration, payload *NotificationPayload, template *ServiceNotificationTemplate) error {
	message := ts.generateServiceMessage(payload, template)
	return ts.SendNotification(config, message)
}

// generateServerMessage creates a message for server notifications using server template
func (ts *TelegramService) generateServerMessage(payload *NotificationPayload, template *ServerNotificationTemplate, resourceType string) string {
	var templateMessage string
	
	// Select appropriate template message based on status and resource type
	switch strings.ToLower(payload.Status) {
	case "down":
		templateMessage = template.DownMessage
	case "up":
		templateMessage = template.UpMessage
	case "warning":
		templateMessage = template.WarningMessage
	case "paused":
		templateMessage = template.PausedMessage
	default:
		// Handle resource-specific messages
		switch resourceType {
		case "cpu":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreCPUMessage
			} else {
				templateMessage = template.CPUMessage
			}
		case "ram", "memory":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreRAMMessage
			} else {
				templateMessage = template.RAMMessage
			}
		case "disk":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreDiskMessage
			} else {
				templateMessage = template.DiskMessage
			}
		case "network":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreNetworkMessage
			} else {
				templateMessage = template.NetworkMessage
			}
		case "cpu_temp", "cpu_temperature":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreCPUTempMessage
			} else {
				templateMessage = template.CPUTempMessage
			}
		case "disk_io":
			if strings.Contains(strings.ToLower(payload.Status), "restore") {
				templateMessage = template.RestoreDiskIOMessage
			} else {
				templateMessage = template.DiskIOMessage
			}
		default:
			templateMessage = template.WarningMessage
		}
	}
	
	// If no template message found, use a default
	if templateMessage == "" {
		templateMessage = ts.generateDefaultServerMessage(payload, resourceType)
	}
	
	return ts.replacePlaceholders(templateMessage, payload)
}

// generateServiceMessage creates a message for service notifications using service template
func (ts *TelegramService) generateServiceMessage(payload *NotificationPayload, template *ServiceNotificationTemplate) string {
	var templateMessage string
	
	// Select appropriate template message based on status
	switch strings.ToLower(payload.Status) {
	case "up":
		templateMessage = template.UpMessage
	case "down":
		templateMessage = template.DownMessage
	case "maintenance":
		templateMessage = template.MaintenanceMessage
	case "incident":
		templateMessage = template.IncidentMessage
	case "resolved":
		templateMessage = template.ResolvedMessage
	case "warning":
		templateMessage = template.WarningMessage
	default:
		templateMessage = template.WarningMessage
	}
	
	// If no template message found, use a default
	if templateMessage == "" {
		templateMessage = ts.generateDefaultUptimeMessage(payload)
	}
	
	return ts.replacePlaceholders(templateMessage, payload)
}

// replacePlaceholders replaces all placeholders in the message with actual values
func (ts *TelegramService) replacePlaceholders(message string, payload *NotificationPayload) string {
	// Replace basic placeholders
	message = strings.ReplaceAll(message, "${service_name}", payload.ServiceName)
	message = strings.ReplaceAll(message, "${status}", strings.ToUpper(payload.Status))
	message = strings.ReplaceAll(message, "${host}", ts.safeString(payload.Host))
	message = strings.ReplaceAll(message, "${hostname}", ts.safeString(payload.Hostname))
	
	// Replace URL with fallback to host
	url := ts.safeString(payload.URL)
	if url == "N/A" && payload.Host != "" {
		url = payload.Host
	}
	message = strings.ReplaceAll(message, "${url}", url)
	
	// Replace domain
	message = strings.ReplaceAll(message, "${domain}", ts.safeString(payload.Domain))
	
	// Replace service type
	if payload.ServiceType != "" {
		message = strings.ReplaceAll(message, "${service_type}", strings.ToUpper(payload.ServiceType))
	} else {
		message = strings.ReplaceAll(message, "${service_type}", "N/A")
	}
	
	// Replace region and agent info
	message = strings.ReplaceAll(message, "${region_name}", ts.safeString(payload.RegionName))
	message = strings.ReplaceAll(message, "${agent_id}", ts.safeString(payload.AgentID))
	
	// Handle numeric fields safely
	if payload.Port > 0 {
		message = strings.ReplaceAll(message, "${port}", fmt.Sprintf("%d", payload.Port))
	} else {
		message = strings.ReplaceAll(message, "${port}", "N/A")
	}
	
	if payload.ResponseTime > 0 {
		message = strings.ReplaceAll(message, "${response_time}", fmt.Sprintf("%dms", payload.ResponseTime))
	} else {
		message = strings.ReplaceAll(message, "${response_time}", "N/A")
	}
	
	if payload.Uptime > 0 {
		message = strings.ReplaceAll(message, "${uptime}", fmt.Sprintf("%d%%", payload.Uptime))
	} else {
		message = strings.ReplaceAll(message, "${uptime}", "N/A")
	}
	
	// Replace server monitoring fields
	message = strings.ReplaceAll(message, "${cpu_usage}", ts.safeString(payload.CPUUsage))
	message = strings.ReplaceAll(message, "${ram_usage}", ts.safeString(payload.RAMUsage))
	message = strings.ReplaceAll(message, "${disk_usage}", ts.safeString(payload.DiskUsage))
	message = strings.ReplaceAll(message, "${network_usage}", ts.safeString(payload.NetworkUsage))
	message = strings.ReplaceAll(message, "${cpu_temp}", ts.safeString(payload.CPUTemp))
	message = strings.ReplaceAll(message, "${disk_io}", ts.safeString(payload.DiskIO))
	message = strings.ReplaceAll(message, "${threshold}", ts.safeString(payload.Threshold))
	
	// Replace error message - important for uptime services
	message = strings.ReplaceAll(message, "${error_message}", ts.safeString(payload.ErrorMessage))
	message = strings.ReplaceAll(message, "${error}", ts.safeString(payload.ErrorMessage))
	
	// Replace time placeholders
	message = strings.ReplaceAll(message, "${time}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	message = strings.ReplaceAll(message, "${timestamp}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	
	return message
}

// safeString returns the string value or "N/A" if empty
func (ts *TelegramService) safeString(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// generateDefaultUptimeMessage creates a default uptime message with proper formatting
func (ts *TelegramService) generateDefaultUptimeMessage(payload *NotificationPayload) string {
	// Status emoji mapping
	statusEmoji := "🔵"
	switch strings.ToLower(payload.Status) {
	case "up":
		statusEmoji = "🟢"
	case "down":
		statusEmoji = "🔴"
	case "warning":
		statusEmoji = "🟡"
	case "maintenance", "paused":
		statusEmoji = "🟠"
	}
	
	message := fmt.Sprintf("%sService %s is %s.", statusEmoji, payload.ServiceName, strings.ToUpper(payload.Status))
	
	// Build formatted details
	details := []string{}
	
	// Add URL or host
	if payload.URL != "" {
		details = append(details, fmt.Sprintf(" - Host URL: %s", payload.URL))
	} else if payload.Host != "" {
		details = append(details, fmt.Sprintf(" - Host: %s", payload.Host))
	}
	
	// Add service type
	if payload.ServiceType != "" {
		details = append(details, fmt.Sprintf(" - Type: %s", strings.ToUpper(payload.ServiceType)))
	}
	
	// Add port if available
	if payload.Port > 0 {
		details = append(details, fmt.Sprintf(" - Port: %d", payload.Port))
	}
	
	// Add domain if available
	if payload.Domain != "" {
		details = append(details, fmt.Sprintf(" - Domain: %s", payload.Domain))
	}
	
	// Add response time
	if payload.ResponseTime > 0 {
		details = append(details, fmt.Sprintf(" - Response time: %dms", payload.ResponseTime))
	} else {
		details = append(details, " - Response time: N/A")
	}
	
	// Add region info
	if payload.RegionName != "" {
		details = append(details, fmt.Sprintf(" - Region: %s", payload.RegionName))
	}
	
	// Add agent info
	if payload.AgentID != "" {
		details = append(details, fmt.Sprintf(" - Agent: %s", payload.AgentID))
	}
	
	// Add uptime if available
	if payload.Uptime > 0 {
		details = append(details, fmt.Sprintf(" - Uptime: %d%%", payload.Uptime))
	}
	
	// Add timestamp
	details = append(details, fmt.Sprintf(" - Time: %s", payload.Timestamp.Format("2006-01-02 15:04:05")))
	
	// Combine message with details
	if len(details) > 0 {
		message += "\n" + strings.Join(details, "\n")
	}
	
	return message
}

// generateDefaultServerMessage creates a default server message
func (ts *TelegramService) generateDefaultServerMessage(payload *NotificationPayload, resourceType string) string {
	statusEmoji := "🔵"
	switch strings.ToLower(payload.Status) {
	case "up":
		statusEmoji = "🟢"
	case "down":
		statusEmoji = "🔴"
	case "warning":
		statusEmoji = "🟡"
	}
	
	return fmt.Sprintf("%s🖥️ Server %s (%s) status: %s", statusEmoji, payload.ServiceName, payload.Hostname, strings.ToUpper(payload.Status))
}