package maintenancemonitoring

// MaintenanceRecord represents a maintenance record from PocketBase
type MaintenanceRecord struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	Description           string `json:"description"`
	Status                string `json:"status"`
	StartTime             string `json:"start_time"`
	EndTime               string `json:"end_time"`
	Priority              string `json:"priority"`
	Field                 string `json:"field"`
	Affected              string `json:"affected"`
	NotifySubscribers     string `json:"notify_subscribers"`
	NotificationChannelID string `json:"notification_channel_id"`
	NotificationID        string `json:"notification_id"`
	NotifiedStart         bool   `json:"notified_start"`
	NotifiedEnd           bool   `json:"notified_end"`
}

// MaintenanceListResponse represents PocketBase list response for maintenance
type MaintenanceListResponse struct {
	Page       int                 `json:"page"`
	PerPage    int                 `json:"perPage"`
	TotalItems int                 `json:"totalItems"`
	Items      []MaintenanceRecord `json:"items"`
}
