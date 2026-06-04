package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"time"

	"api-service/config"
	"api-service/db"
	"api-service/handler"
	logger "api-service/utils"
	"api-service/middleware"
	"api-service/repository"
	"api-service/service"
)

//go:embed web/* public/*
var embeddedFS embed.FS

func main() {
	if err := logger.InitLogger("runtimes"); err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	logger.Info("Initializing API Service...")

	// 1. Load application configurations
	cfg := config.LoadConfig()

	// 2. Setup MySQL database connection
	sqlDB, err := db.InitDB(cfg.DBDSN)
	if err != nil {
		logger.Errorf("failed to connect to database: %v", err)
		logger.Fatal("Failed to initialize database connection. Exiting.")
	}
	defer sqlDB.Close()

	// 3. Initialize repositories (Data Access Layer)
	appRepo := repository.NewAppRepository(sqlDB)
	tokenRepo := repository.NewTokenRepository(sqlDB)
	adminRepo := repository.NewAdminRepository(sqlDB)
	blacklistRepo := repository.NewBlacklistRepository(sqlDB)
	logRepo := repository.NewLogRepository(sqlDB)

	// Seed default administrator if DB is empty
	db.SeedAdminUser(adminRepo)

	// 4. Initialize services (Business Logic Layer)
	appService := service.NewAppService(appRepo)
	tokenService := service.NewTokenService(tokenRepo, appRepo, blacklistRepo, logRepo)
	adminService := service.NewAdminService(adminRepo)

	// 5. Initialize handlers (Controller Layer)
	webHandler := handler.NewWebHandler(embeddedFS, adminService, appService, tokenService)

	// 6. Register routes
	mux := http.NewServeMux()

	// Redirect root "/" to "/admin"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	// Public Admin routes
	mux.Handle("/admin/login", webHandler)
	mux.Handle("/admin/captcha", webHandler)

	// Protected Admin Routes (requires Administrator session cookie)
	adminSessionAuth := middleware.AdminSessionMiddleware(adminService)

	mux.Handle("/admin", adminSessionAuth(webHandler))
	mux.Handle("/admin/apps", adminSessionAuth(webHandler))
	mux.Handle("/admin/users", adminSessionAuth(webHandler))
	mux.Handle("/admin/tokens", adminSessionAuth(webHandler))
	mux.Handle("/admin/blacklist", adminSessionAuth(webHandler))
	mux.Handle("/admin/blacklist/add", adminSessionAuth(webHandler))
	mux.Handle("/admin/blacklist/delete", adminSessionAuth(webHandler))
	mux.Handle("/admin/logs", adminSessionAuth(webHandler))
	mux.Handle("/admin/logs/blacklist", adminSessionAuth(webHandler))
	mux.Handle("/admin/logout", adminSessionAuth(webHandler))
	mux.Handle("/admin/apps/register", adminSessionAuth(webHandler))
	mux.Handle("/admin/apps/toggle", adminSessionAuth(webHandler))
	mux.Handle("/admin/apps/delete", adminSessionAuth(webHandler))
	mux.Handle("/admin/tokens/generate", adminSessionAuth(webHandler))
	mux.Handle("/admin/tokens/revoke", adminSessionAuth(webHandler))
	mux.Handle("/admin/users/create", adminSessionAuth(webHandler))

	// Public Protected API endpoint for client applications (requires Client Access Token)
	protectedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handler.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		appID := middleware.GetAppID(r.Context())
		version := middleware.GetVersion(r.Context())

		handler.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"message":               "Access granted to protected resource!",
			"authenticated_app_id":  appID,
			"authenticated_version": version,
			"timestamp":             time.Now().Format(time.RFC3339),
		})
	})

	clientAuth := middleware.AuthMiddleware(tokenService)
	mux.Handle("/api/protected/resource", clientAuth(protectedMux))

	// Serve public static files (checks local directory first for development, falls back to embedded FS)
	var publicHandler http.Handler
	if _, err := os.Stat("public"); err == nil {
		publicHandler = http.FileServer(http.Dir("public"))
	} else {
		publicSubFS, err := fs.Sub(embeddedFS, "public")
		if err != nil {
			logger.Fatalf("Failed to initialize embedded public assets: %v", err)
		}
		publicHandler = http.FileServer(http.FS(publicSubFS))
	}
	mux.Handle("/public/", http.StripPrefix("/public/", publicHandler))

	// 7. Configure and start HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      middleware.LoggerMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Infof("Server is listening on port %s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server stopped unexpectedly: %v", err)
	}
}

