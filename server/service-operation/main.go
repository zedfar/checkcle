package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"service-operation/config"
	"service-operation/handlers"
	maintenancemonitoring "service-operation/maintenance-monitoring"
	"service-operation/monitoring"
	"service-operation/pocketbase"
	servermonitoring "service-operation/server-monitoring"
	sslmonitoring "service-operation/ssl-monitoring"
	uptimemonitoring "service-operation/uptime-monitoring"
	dataretention "service-operation/data-retention"
)

func main() {
	//log.Println("🚀 === STARTING SERVICE OPERATION SERVER ===")
	
	cfg := config.Load()
	//log.Printf("📋 Configuration loaded:")
	//log.Printf("  - Port: %s", cfg.Port)
	//log.Printf("  - Backend PB Enabled: %t", cfg.PocketBaseEnabled)
	if cfg.PocketBaseEnabled {
		//log.Printf("  - PocketBase URL: %s", cfg.PocketBaseURL)
	}
	
	// Initialize PocketBase client
	var pbClient *pocketbase.PocketBaseClient
	var monitoringService *monitoring.MonitoringService
	var sslMonitoringService *monitoring.SSLMonitoringService
	var sslNotificationService *sslmonitoring.SSLMonitor
	var serverMonitoringService *servermonitoring.ServerMonitoringService
	var uptimeMonitoringService *uptimemonitoring.UptimeMonitor
	var maintenanceMonitor *maintenancemonitoring.MaintenanceMonitor
	var dataRetentionScheduler *dataretention.Scheduler

	if cfg.PocketBaseEnabled {
		var err error
		pbClient, err = pocketbase.NewPocketBaseClient(cfg.PocketBaseURL)
		if err != nil {
			log.Printf("WARNING: Failed to initialize PocketBase client: %v", err)
		} else {
			if cfg.PocketBaseAdminEmail != "" {
				if authErr := pbClient.Authenticate(cfg.PocketBaseAdminEmail, cfg.PocketBaseAdminPassword); authErr != nil {
					log.Printf("WARNING: PocketBase authentication failed: %v", authErr)
				}
			}

			if err := pbClient.TestConnection(); err != nil {
				//log.Printf("⚠️  WARNING: PocketBase connection test failed: %v", err)
			} else {
				//log.Println("✅ PocketBase connection test successful")
				
				// Initialize and start service monitoring with regional support
				//log.Println("🔧 Initializing service monitoring...")
				monitoringService = monitoring.NewMonitoringService(pbClient)
				go monitoringService.Start()
				//log.Println("✅ Service monitoring started with regional agent support")
				
				// Initialize and start SSL monitoring service (original)
				//log.Println("🔧 Initializing SSL monitoring (original)...")
				sslMonitoringService = monitoring.NewSSLMonitoringService(pbClient)
				go sslMonitoringService.Start()
				//log.Println("✅ SSL monitoring started (independent of regional agents)")
				
				// Initialize and start SSL notification service (new)
				//log.Println("🔧 Initializing SSL notification monitoring...")
				sslNotificationService = sslmonitoring.NewSSLMonitor(pbClient)
				go sslNotificationService.Start()
				//log.Println("✅ SSL notification monitoring started with Telegram support")
				
				// Initialize and start server monitoring service
				//log.Println("🔧 Initializing server monitoring...")
				serverMonitoringService = servermonitoring.NewServerMonitoringService(pbClient)
				serverMonitoringService.Start()
				//log.Println("✅ Server monitoring started with notification support")
				
				// Initialize and start uptime monitoring service
				//log.Println("🔧 Initializing uptime monitoring...")
				uptimeMonitoringService = uptimemonitoring.NewUptimeMonitor(pbClient)
				go uptimeMonitoringService.Start()
				//log.Println("✅ Uptime monitoring started with notification support")

				// Initialize and start maintenance monitoring
				maintenanceMonitor = maintenancemonitoring.NewMaintenanceMonitor(pbClient)
				go maintenanceMonitor.Start()

				// Initialize and start data retention scheduler
				//log.Println("🔧 Initializing data retention scheduler...")
				dataRetentionScheduler = dataretention.NewScheduler(pbClient, 24*time.Hour) // Run daily
				go dataRetentionScheduler.Start()
				//log.Println("✅ Data retention scheduler started (daily cleanup)")
			}
		}
	}
	
	//log.Println("🔧 Initializing HTTP handlers...")
	handler := handlers.NewOperationHandler(cfg, pbClient)

	router := mux.NewRouter()

	// Main operation endpoint
	router.HandleFunc("/operation", handler.HandleOperation).Methods("POST")
	
	// Quick operation endpoint with query parameters
	router.HandleFunc("/operation/quick", handler.HandleQuickOperation).Methods("GET")
	
	// Legacy ping endpoint for backward compatibility
	router.HandleFunc("/ping", handler.HandleOperation).Methods("POST")
	router.HandleFunc("/ping/quick", handler.HandleQuickOperation).Methods("GET")
	
	// Health check
	router.HandleFunc("/health", handler.HandleHealth).Methods("GET")

	log.Printf("=== 🌐 CHECKCLE SERVICE OPERATION SERVER READY ===")
	log.Printf("🚀 Starting on port %s", cfg.Port)
	if pbClient != nil {
		log.Printf("✓Backend integration enabled at %s ", pbClient.GetBaseURL())
	}
	if monitoringService != nil {
		log.Printf("✓Service monitoring enabled with regional agent support")
	}
	if sslMonitoringService != nil {
		//log.Printf("🔒 SSL certificate monitoring enabled (independent)")
	}
	if sslNotificationService != nil {
		log.Printf("✓SSL notification monitoring enabled with Telegram support")
	}
	if serverMonitoringService != nil {
		log.Printf("✓Server monitoring enabled with notification support")
	}
	if uptimeMonitoringService != nil {
		log.Printf("✓Uptime monitoring enabled with notification support")
	}
	if maintenanceMonitor != nil {
		log.Printf("✓Maintenance monitoring enabled (60s interval)")
	}
	log.Printf("✓Supported operations: ping, dns, tcp, http, ssl")
	if dataRetentionScheduler != nil {
		log.Printf("✓Data retention scheduler enabled (daily cleanup)")
	}
	

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("🛑 === GRACEFUL SHUTDOWN INITIATED ===")
		log.Println("🛑 Shutting down monitoring services...")
		
		if monitoringService != nil {
			log.Println("🛑 Stopping service monitoring...")
			monitoringService.Stop()
		}
		if sslMonitoringService != nil {
			log.Println("🛑 Stopping SSL monitoring...")
			sslMonitoringService.Stop()
		}
		if sslNotificationService != nil {
			log.Println("🛑 Stopping SSL notification monitoring...")
			sslNotificationService.Stop()
		}
		if serverMonitoringService != nil {
			log.Println("🛑 Stopping server monitoring...")
			serverMonitoringService.Stop()
		}
		if uptimeMonitoringService != nil {
			log.Println("🛑 Stopping uptime monitoring...")
			uptimeMonitoringService.Stop()
		}
		if maintenanceMonitor != nil {
			log.Println("🛑 Stopping maintenance monitoring...")
			maintenanceMonitor.Stop()
		}
		if dataRetentionScheduler != nil {
			log.Println("🛑 Stopping data retention scheduler...")
			dataRetentionScheduler.Stop()
		}
		
		log.Println("✅ All services stopped gracefully")
		log.Println("🛑 === SERVICE OPERATION SERVER STOPPED ===")
		os.Exit(0)
	}()

	//log.Println("🌐 HTTP server starting...")
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("❌ FATAL: Failed to start HTTP server:", err)
	}
}