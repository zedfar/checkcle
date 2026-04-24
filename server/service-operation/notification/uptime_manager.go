package notification

import (
	"fmt"
	// "log"
	"strings"

	"service-operation/pocketbase"
)

// UptimeNotificationManager handles uptime service notifications with IMMEDIATE sending
type UptimeNotificationManager struct {
	pbClient *pocketbase.PocketBaseClient
	services map[string]NotificationService
}

// NewUptimeNotificationManager creates a new uptime notification manager
func NewUptimeNotificationManager(pbClient *pocketbase.PocketBaseClient, services map[string]NotificationService) *UptimeNotificationManager {
	return &UptimeNotificationManager{
		pbClient: pbClient,
		services: services,
	}
}

// SendUptimeServiceNotification sends notification IMMEDIATELY - NO QUEUING
func (unm *UptimeNotificationManager) SendUptimeServiceNotification(payload *NotificationPayload, notificationID, templateID string) error {
	// log.Printf("📨 [UPTIME-MANAGER] IMMEDIATE send for service: %s, status: %s", payload.ServiceName, payload.Status)
	// log.Printf("📨 [UPTIME-MANAGER] Notification ID: %s, Template ID: %s", notificationID, templateID)
	
	if notificationID == "" {
		return fmt.Errorf("notification ID required for uptime service: %s", payload.ServiceName)
	}

	// Parse notification IDs
	notificationIDs := parseNotificationIDs(notificationID)
	if len(notificationIDs) == 0 {
		return fmt.Errorf("no valid notification IDs for uptime service: %s", payload.ServiceName)
	}

	var errors []string
	successCount := 0

	// Send to each notification channel IMMEDIATELY
	for _, id := range notificationIDs {
		// log.Printf("📤 [UPTIME-SEND] Processing notification ID: %s for service %s", id, payload.ServiceName)
		
		// Check if enabled
		if !isNotificationEnabled(unm.pbClient, id) {
			// log.Printf("⚠️ Notification %s disabled for uptime service %s, skipping", id, payload.ServiceName)
			continue
		}

		// Get alert configuration
		alertConfig, err := getAlertConfiguration(unm.pbClient, id)
		if err != nil {
			// log.Printf("❌ Failed to get alert config for %s (service: %s): %v", id, payload.ServiceName, err)
			errors = append(errors, fmt.Sprintf("config error %s: %v", id, err))
			continue
		}

		// log.Printf("📋 [UPTIME-CONFIG] Alert config for %s: Type=%s, ChatID=%s, Token present=%v", 
		//	id, alertConfig.NotificationType, alertConfig.TelegramChatID, alertConfig.BotToken != "")

		// Get service template
		var serviceTemplate *ServiceNotificationTemplate
		if templateID != "" {
			serviceTemplate, err = getServiceNotificationTemplate(unm.pbClient, templateID)
			if err != nil {
				// log.Printf("⚠️ Template error for %s (service: %s): %v", templateID, payload.ServiceName, err)
			}
		}

		// Generate message
		message := unm.generateUptimeMessage(payload, serviceTemplate)
		// log.Printf("📝 [UPTIME-MESSAGE] Generated for %s (%s): %s", payload.ServiceName, id, message)

		// Get notification service
		service, exists := unm.services[alertConfig.NotificationType]
		if !exists {
			// log.Printf("❌ Unsupported notification type for uptime: %s", alertConfig.NotificationType)
			errors = append(errors, fmt.Sprintf("unsupported type %s", alertConfig.NotificationType))
			continue
		}

		// SEND IMMEDIATELY - NO DELAYS
		// log.Printf("⚡ [UPTIME-TELEGRAM] Sending via %s for uptime service %s", alertConfig.NotificationType, payload.ServiceName)
		if ws, ok := service.(*WebhookService); ok {
			err = ws.SendNotificationWithPayload(alertConfig, message, payload)
		} else {
			err = service.SendNotification(alertConfig, message)
		}
		if err != nil {
			// log.Printf("❌ [UPTIME-FAILED] Failed to send via %s for %s: %v", alertConfig.NotificationType, payload.ServiceName, err)
			errors = append(errors, fmt.Sprintf("send failed %s: %v", alertConfig.NotificationType, err))
		} else {
			// log.Printf("✅ [UPTIME-SUCCESS] Successfully sent via %s for %s", alertConfig.NotificationType, payload.ServiceName)
			successCount++
		}
	}

	// Report results
	if successCount > 0 {
		// log.Printf("✅ [UPTIME-FINAL] Sent %d/%d uptime notifications for %s", successCount, len(notificationIDs), payload.ServiceName)
	}

	if len(errors) > 0 && successCount == 0 {
		return fmt.Errorf("all uptime notifications failed for %s: %v", payload.ServiceName, errors)
	}

	return nil
}

// generateUptimeMessage creates notification message for uptime services
func (unm *UptimeNotificationManager) generateUptimeMessage(payload *NotificationPayload, template *ServiceNotificationTemplate) string {
	var baseMessage string

	// Use template if available
	if template != nil {
		// log.Printf("🔧 Using uptime template for status: %s", strings.ToLower(payload.Status))
		switch strings.ToLower(payload.Status) {
		case "up":
			baseMessage = template.UpMessage
		case "down":
			baseMessage = template.DownMessage
		case "maintenance":
			baseMessage = template.MaintenanceMessage
		case "incident":
			baseMessage = template.IncidentMessage
		case "resolved":
			baseMessage = template.ResolvedMessage
		case "warning":
			baseMessage = template.WarningMessage
		default:
			baseMessage = template.WarningMessage
		}
	}

	// Use default if no template
	if baseMessage == "" {
		// log.Printf("🔧 Using default uptime message (no template)")
		baseMessage = unm.getDefaultUptimeMessage(payload)
	}

	// Replace placeholders
	message := unm.replacePlaceholders(baseMessage, payload)
	
	// log.Printf("📝 Final uptime message: %s", message)
	return message
}

// replacePlaceholders replaces all placeholders in the message
func (unm *UptimeNotificationManager) replacePlaceholders(message string, payload *NotificationPayload) string {
	message = strings.ReplaceAll(message, "${service_name}", payload.ServiceName)
	message = strings.ReplaceAll(message, "${status}", strings.ToUpper(payload.Status))
	message = strings.ReplaceAll(message, "${host}", payload.Host)
	message = strings.ReplaceAll(message, "${ip}", payload.Host)
	
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
	
	if payload.URL != "" {
		message = strings.ReplaceAll(message, "${url}", payload.URL)
	} else {
		message = strings.ReplaceAll(message, "${url}", "N/A")
	}
	
	if payload.Domain != "" {
		message = strings.ReplaceAll(message, "${domain}", payload.Domain)
	} else {
		message = strings.ReplaceAll(message, "${domain}", "N/A")
	}
	
	if payload.ServiceType != "" {
		message = strings.ReplaceAll(message, "${service_type}", strings.ToUpper(payload.ServiceType))
	} else {
		message = strings.ReplaceAll(message, "${service_type}", "N/A")
	}
	
	if payload.RegionName != "" {
		message = strings.ReplaceAll(message, "${region_name}", payload.RegionName)
	} else {
		message = strings.ReplaceAll(message, "${region_name}", "N/A")
	}
	
	if payload.AgentID != "" {
		message = strings.ReplaceAll(message, "${agent_id}", payload.AgentID)
	} else {
		message = strings.ReplaceAll(message, "${agent_id}", "N/A")
	}
	
	if payload.Uptime > 0 {
		message = strings.ReplaceAll(message, "${uptime}", fmt.Sprintf("%d%%", payload.Uptime))
	} else {
		message = strings.ReplaceAll(message, "${uptime}", "N/A")
	}
	
	message = strings.ReplaceAll(message, "${timestamp}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	message = strings.ReplaceAll(message, "${time}", payload.Timestamp.Format("2006-01-02 15:04:05"))
	message = strings.ReplaceAll(message, "${date}", payload.Timestamp.Format("2006-01-02"))
	
	// Handle error message placeholder - critical for uptime services
	if payload.ErrorMessage != "" {
		message = strings.ReplaceAll(message, "${error}", payload.ErrorMessage)
		message = strings.ReplaceAll(message, "${error_message}", payload.ErrorMessage)
	} else {
		message = strings.ReplaceAll(message, "${error}", "")
		message = strings.ReplaceAll(message, "${error_message}", "")
	}
	
	if payload.Message != "" {
		message = strings.ReplaceAll(message, "${message}", payload.Message)
	}

	return message
}

// getDefaultUptimeMessage provides a default notification message for uptime services
func (unm *UptimeNotificationManager) getDefaultUptimeMessage(payload *NotificationPayload) string {
	statusEmoji := "✅"
	if payload.Status == "down" {
		statusEmoji = "❌"
	} else if payload.Status == "warning" {
		statusEmoji = "⚠️"
	}

	message := fmt.Sprintf("%s [UPTIME] %s is %s", statusEmoji, payload.ServiceName, strings.ToUpper(payload.Status))
	
	if payload.Host != "" {
		message += fmt.Sprintf(" | Host: %s", payload.Host)
	}
	
	if payload.ResponseTime > 0 {
		message += fmt.Sprintf(" | %dms", payload.ResponseTime)
	}
	
	message += fmt.Sprintf(" | %s", payload.Timestamp.Format("15:04:05"))

	return message
}