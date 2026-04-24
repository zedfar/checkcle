
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

// WebhookService handles webhook notifications
type WebhookService struct{}

// NewWebhookService creates a new webhook notification service
func NewWebhookService() *WebhookService {
	return &WebhookService{}
}

// WebhookPayload represents the payload for webhook
type WebhookPayload struct {
	Message      string `json:"message"`
	NotifyName   string `json:"notify_name"`
	Timestamp    string `json:"timestamp"`
	ServiceName  string `json:"service_name,omitempty"`
	Status       string `json:"status,omitempty"`
	Host         string `json:"host,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
	URL          string `json:"url,omitempty"`
	Domain       string `json:"domain,omitempty"`
	Port         int    `json:"port,omitempty"`
	ServiceType  string `json:"service_type,omitempty"`
	ResponseTime int64  `json:"response_time,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	// Server monitoring fields
	CPUUsage     string `json:"cpu_usage,omitempty"`
	RAMUsage     string `json:"ram_usage,omitempty"`
	DiskUsage    string `json:"disk_usage,omitempty"`
	NetworkUsage string `json:"network_usage,omitempty"`
	CPUTemp      string `json:"cpu_temp,omitempty"`
	DiskIO       string `json:"disk_io,omitempty"`
	Threshold    string `json:"threshold,omitempty"`
	// SSL fields
	CertificateName string `json:"certificate_name,omitempty"`
	ExpiryDate      string `json:"expiry_date,omitempty"`
	DaysLeft        string `json:"days_left,omitempty"`
	IssuerCN        string `json:"issuer_cn,omitempty"`
}

// DiscordWebhookPayload represents Discord-specific webhook payload
type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

// SendNotification sends a notification via webhook
func (ws *WebhookService) SendNotification(config *AlertConfiguration, message string) error {
	// fmt.Printf("📡 [WEBHOOK] Attempting to send notification...\n")
	// fmt.Printf("📡 [WEBHOOK] Config - URL: %s, Notify Name: %s\n", config.WebhookURL, config.NotifyName)
	// fmt.Printf("📡 [WEBHOOK] Message: %s\n", message)

	// Use the webhook_url field from alert_configurations
	if config.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Create payload using webhook_payload_template if available
	var jsonData []byte
	var err error

	if config.WebhookPayloadTemplate != "" && strings.TrimSpace(config.WebhookPayloadTemplate) != "" {
		// Use custom payload template
		jsonData, err = ws.generateCustomPayload(config, message)
		if err != nil {
			// fmt.Printf("❌ [WEBHOOK] Custom payload generation error: %v\n", err)
			return err
		}
	} else {
		// Detect Discord webhook and use appropriate payload
		if strings.Contains(strings.ToLower(config.WebhookURL), "discord") {
			jsonData, err = ws.generateDiscordPayload(message)
		} else {
			// Use default payload structure for other webhooks
			jsonData, err = ws.generateDefaultPayload(config, message)
		}
		
		if err != nil {
			// fmt.Printf("❌ [WEBHOOK] JSON marshal error: %v\n", err)
			return err
		}
	}

	// fmt.Printf("📡 [WEBHOOK] Payload: %s\n", string(jsonData))
	// fmt.Printf("📡 [WEBHOOK] Sending POST request...\n")
	
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// fmt.Printf("❌ [WEBHOOK] HTTP error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// fmt.Printf("❌ [WEBHOOK] API error, status: %d\n", resp.StatusCode)
		return fmt.Errorf("webhook error, status: %d", resp.StatusCode)
	}

	// fmt.Printf("✅ [WEBHOOK] Message sent successfully!\n")
	return nil
}

// generateDiscordPayload generates a Discord-compatible webhook payload
func (ws *WebhookService) generateDiscordPayload(message string) ([]byte, error) {
	payload := DiscordWebhookPayload{
		Content: message,
	}
	return json.Marshal(payload)
}

// generateDefaultPayload generates the default webhook payload
func (ws *WebhookService) generateDefaultPayload(config *AlertConfiguration, message string) ([]byte, error) {
	payload := WebhookPayload{
		Message:    message,
		NotifyName: config.NotifyName,
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	return json.Marshal(payload)
}

// generateCustomPayload generates payload using the custom template
func (ws *WebhookService) generateCustomPayload(config *AlertConfiguration, message string) ([]byte, error) {
	return ws.generateCustomPayloadWithData(config, message, nil)
}

// generateCustomPayloadWithData generates payload using the custom template, with optional structured data
func (ws *WebhookService) generateCustomPayloadWithData(config *AlertConfiguration, message string, payload *NotificationPayload) ([]byte, error) {
	tmpl, err := template.New("webhook").Parse(config.WebhookPayloadTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload template: %v", err)
	}

	templateData := map[string]interface{}{
		"message":     message,
		"notify_name": config.NotifyName,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if payload != nil {
		templateData["service_name"] = payload.ServiceName
		templateData["status"] = strings.ToUpper(payload.Status)
		templateData["host"] = payload.Host
		templateData["hostname"] = payload.Hostname
		templateData["url"] = payload.URL
		templateData["domain"] = payload.Domain
		templateData["port"] = payload.Port
		templateData["service_type"] = payload.ServiceType
		templateData["response_time"] = payload.ResponseTime
		templateData["error_message"] = payload.ErrorMessage
		templateData["cpu_usage"] = payload.CPUUsage
		templateData["ram_usage"] = payload.RAMUsage
		templateData["disk_usage"] = payload.DiskUsage
		templateData["days_left"] = payload.DaysLeft
		templateData["expiry_date"] = payload.ExpiryDate
		templateData["issuer_cn"] = payload.IssuerCN
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute webhook payload template: %v", err)
	}

	var jsonCheck interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonCheck); err != nil {
		return nil, fmt.Errorf("webhook payload template did not generate valid JSON: %v", err)
	}

	return buf.Bytes(), nil
}

// generateStructuredPayload builds a rich JSON payload from a NotificationPayload
func (ws *WebhookService) generateStructuredPayload(config *AlertConfiguration, message string, payload *NotificationPayload) ([]byte, error) {
	p := WebhookPayload{
		Message:         message,
		NotifyName:      config.NotifyName,
		Timestamp:       time.Now().Format(time.RFC3339),
		ServiceName:     payload.ServiceName,
		Status:          strings.ToUpper(payload.Status),
		Host:            payload.Host,
		Hostname:        payload.Hostname,
		URL:             payload.URL,
		Domain:          payload.Domain,
		Port:            payload.Port,
		ServiceType:     payload.ServiceType,
		ResponseTime:    payload.ResponseTime,
		ErrorMessage:    payload.ErrorMessage,
		CPUUsage:        payload.CPUUsage,
		RAMUsage:        payload.RAMUsage,
		DiskUsage:       payload.DiskUsage,
		NetworkUsage:    payload.NetworkUsage,
		CPUTemp:         payload.CPUTemp,
		DiskIO:          payload.DiskIO,
		Threshold:       payload.Threshold,
		CertificateName: payload.CertificateName,
		ExpiryDate:      payload.ExpiryDate,
		DaysLeft:        payload.DaysLeft,
		IssuerCN:        payload.IssuerCN,
	}
	return json.Marshal(p)
}

// SendNotificationWithPayload sends a webhook notification with full structured payload data
func (ws *WebhookService) SendNotificationWithPayload(config *AlertConfiguration, message string, payload *NotificationPayload) error {
	if config.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	var jsonData []byte
	var err error

	if config.WebhookPayloadTemplate != "" && strings.TrimSpace(config.WebhookPayloadTemplate) != "" {
		jsonData, err = ws.generateCustomPayloadWithData(config, message, payload)
	} else {
		jsonData, err = ws.generateStructuredPayload(config, message, payload)
	}
	if err != nil {
		return err
	}

	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook error, status: %d", resp.StatusCode)
	}

	return nil
}

// SendServerNotification sends a server-specific notification via webhook
func (ws *WebhookService) SendServerNotification(config *AlertConfiguration, payload *NotificationPayload, template *ServerNotificationTemplate, resourceType string) error {
	message := ws.generateServerMessage(payload, template, resourceType)
	return ws.SendNotification(config, message)
}

// SendServiceNotification sends a service-specific notification via webhook
func (ws *WebhookService) SendServiceNotification(config *AlertConfiguration, payload *NotificationPayload, template *ServiceNotificationTemplate) error {
	message := ws.generateServiceMessage(payload, template)
	return ws.SendNotification(config, message)
}

// generateServerMessage creates a message for server notifications using server template
func (ws *WebhookService) generateServerMessage(payload *NotificationPayload, template *ServerNotificationTemplate, resourceType string) string {
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
		templateMessage = ws.generateDefaultServerMessage(payload, resourceType)
	}
	
	return ws.replacePlaceholders(templateMessage, payload)
}

// generateServiceMessage creates a message for service notifications using service template
func (ws *WebhookService) generateServiceMessage(payload *NotificationPayload, template *ServiceNotificationTemplate) string {
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
		templateMessage = ws.generateDefaultUptimeMessage(payload)
	}
	
	return ws.replacePlaceholders(templateMessage, payload)
}

// replacePlaceholders replaces all placeholders in the message with actual values
func (ws *WebhookService) replacePlaceholders(message string, payload *NotificationPayload) string {
	// Replace basic placeholders
	message = strings.ReplaceAll(message, "${service_name}", payload.ServiceName)
	message = strings.ReplaceAll(message, "${server_name}", payload.ServiceName) // server_name maps to service_name
	message = strings.ReplaceAll(message, "${status}", strings.ToUpper(payload.Status))
	message = strings.ReplaceAll(message, "${host}", ws.safeString(payload.Host))
	message = strings.ReplaceAll(message, "${hostname}", ws.safeString(payload.Hostname))
	
	// Replace URL with fallback to host
	url := ws.safeString(payload.URL)
	if url == "N/A" && payload.Host != "" {
		url = payload.Host
	}
	message = strings.ReplaceAll(message, "${url}", url)
	
	// Replace domain
	message = strings.ReplaceAll(message, "${domain}", ws.safeString(payload.Domain))
	
	// Replace service type
	if payload.ServiceType != "" {
		message = strings.ReplaceAll(message, "${service_type}", strings.ToUpper(payload.ServiceType))
	} else {
		message = strings.ReplaceAll(message, "${service_type}", "N/A")
	}
	
	// Replace region and agent info
	message = strings.ReplaceAll(message, "${region_name}", ws.safeString(payload.RegionName))
	message = strings.ReplaceAll(message, "${agent_id}", ws.safeString(payload.AgentID))
	
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
	message = strings.ReplaceAll(message, "${cpu_usage}", ws.safeString(payload.CPUUsage))
	message = strings.ReplaceAll(message, "${ram_usage}", ws.safeString(payload.RAMUsage))
	message = strings.ReplaceAll(message, "${disk_usage}", ws.safeString(payload.DiskUsage))
	message = strings.ReplaceAll(message, "${network_usage}", ws.safeString(payload.NetworkUsage))
	message = strings.ReplaceAll(message, "${cpu_temp}", ws.safeString(payload.CPUTemp))
	message = strings.ReplaceAll(message, "${disk_io}", ws.safeString(payload.DiskIO))
	message = strings.ReplaceAll(message, "${threshold}", ws.safeString(payload.Threshold))
	
	// Replace SSL certificate fields
	message = strings.ReplaceAll(message, "${certificate_name}", ws.safeString(payload.CertificateName))
	message = strings.ReplaceAll(message, "${expiry_date}", ws.safeString(payload.ExpiryDate))
	message = strings.ReplaceAll(message, "${days_left}", ws.safeString(payload.DaysLeft))
	message = strings.ReplaceAll(message, "${issuer_cn}", ws.safeString(payload.IssuerCN))
	message = strings.ReplaceAll(message, "${issuer}", ws.safeString(payload.IssuerCN)) // alias
	
	// Replace error message - important for uptime services
	message = strings.ReplaceAll(message, "${error_message}", ws.safeString(payload.ErrorMessage))
	message = strings.ReplaceAll(message, "${error}", ws.safeString(payload.ErrorMessage))
	message = strings.ReplaceAll(message, "${message}", ws.safeString(payload.Message))
	
	// Replace time placeholders
	message = strings.ReplaceAll(message, "${time}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	message = strings.ReplaceAll(message, "${timestamp}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	
	return message
}

// safeString returns the string value or "N/A" if empty
func (ws *WebhookService) safeString(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// generateDefaultUptimeMessage creates a default uptime message with proper formatting
func (ws *WebhookService) generateDefaultUptimeMessage(payload *NotificationPayload) string {
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
	
	message := fmt.Sprintf("%s Service %s is %s.", statusEmoji, payload.ServiceName, strings.ToUpper(payload.Status))
	
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
func (ws *WebhookService) generateDefaultServerMessage(payload *NotificationPayload, resourceType string) string {
	statusEmoji := "🔵"
	switch strings.ToLower(payload.Status) {
	case "up":
		statusEmoji = "🟢"
	case "down":
		statusEmoji = "🔴"
	case "warning":
		statusEmoji = "🟡"
	}
	
	return fmt.Sprintf("%s 🖥️ Server %s (%s) status: %s", statusEmoji, payload.ServiceName, payload.Hostname, strings.ToUpper(payload.Status))
}