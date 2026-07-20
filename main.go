package main

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"api-service/config"
	"api-service/db"
	"api-service/handler"
	"api-service/handler/admin"
	"api-service/handler/api"
	"api-service/middleware"
	"api-service/repository"
	"api-service/service"
	logger "api-service/utils"
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
	calendarRepo := repository.NewCalendarRepository(sqlDB)
	holidayRepo := repository.NewHolidayRepository(sqlDB)
	menuRepo := repository.NewMenuRepository(sqlDB)
	roleRepo := repository.NewRoleRepository(sqlDB)

	// Seed default administrator if DB is empty
	db.SeedAdminUser(adminRepo)
	db.SeedRBAC(sqlDB)

	// 4. Initialize services (Business Logic Layer)
	appService := service.NewAppService(appRepo)
	tokenService := service.NewTokenService(tokenRepo, appRepo, blacklistRepo, logRepo)
	adminService := service.NewAdminService(adminRepo, cfg)
	calendarService := service.NewCalendarService(calendarRepo)
	holidayService := service.NewHolidayService(holidayRepo)
	menuService := service.NewMenuService(menuRepo)
	roleService := service.NewRoleService(roleRepo)

	// 5. Initialize handlers (Controller Layer)
	adminBase := admin.NewBaseHandler(embeddedFS)
	apiBase := api.NewBaseHandler()
	adminSessionAuth := middleware.AdminSessionMiddleware(adminService)
	clientAuth := middleware.AuthMiddleware(tokenService)

	controllers := []handler.Controller{
		// 后台数据管理
		admin.NewAuthHandler(adminBase, adminService, adminSessionAuth),
		admin.NewDashboardHandler(adminBase, appService, tokenService, adminSessionAuth),
		admin.NewAppHandler(adminBase, appService, tokenService, adminSessionAuth),
		admin.NewTokenHandler(adminBase, tokenService, appService, adminSessionAuth),
		admin.NewUserHandler(adminBase, adminService, adminSessionAuth),
		admin.NewBlacklistHandler(adminBase, tokenService, adminSessionAuth),
		admin.NewLogHandler(adminBase, tokenService, adminSessionAuth),
		admin.NewCalendarHandler(adminBase, calendarService, adminSessionAuth),
		admin.NewHolidayHandler(adminBase, holidayService, adminSessionAuth),
		admin.NewMenuHandler(adminBase, menuService, adminSessionAuth),
		admin.NewRoleHandler(adminBase, roleService, adminSessionAuth),
		// 开放API接口
		api.NewResourceHandler(apiBase, calendarService),
		api.NewCalendarHandler(apiBase, calendarService, holidayService, clientAuth),
	}

	// 6. Register routes
	mux := http.NewServeMux()

	// Catch-all handler: route all non-/api and non-/admin requests to public/index.html (SPA Fallback)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 1. Check if the file physically exists under the public directory
		// (e.g. /favicon.ico -> public/favicon.ico, /assets/index.js -> public/assets/index.js)
		cleanPath := filepath.Clean(path)
		localFilePath := filepath.Join("public", cleanPath)
		if fileInfo, err := os.Stat(localFilePath); err == nil && !fileInfo.IsDir() {
			http.ServeFile(w, r, localFilePath)
			return
		}

		// Also check the embedded filesystem
		embedPath := filepath.ToSlash(filepath.Join("public", cleanPath))
		if fileInfo, err := fs.Stat(embeddedFS, embedPath); err == nil && !fileInfo.IsDir() {
			content, err := embeddedFS.ReadFile(embedPath)
			if err == nil {
				contentType := mime.TypeByExtension(filepath.Ext(embedPath))
				if contentType != "" {
					w.Header().Set("Content-Type", contentType)
				}
				w.Write(content)
				return
			}
		}

		// 2. If the request did not match any file, but starts with /api/ or /admin/, let it return 404 (NotFound).
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/admin/") || path == "/api" || path == "/admin" {
			http.NotFound(w, r)
			return
		}

		// 3. Otherwise, fallback to serving the Vue app's index.html
		if _, err := os.Stat("public/index.html"); err == nil {
			http.ServeFile(w, r, "public/index.html")
			return
		}

		htmlContent, err := embeddedFS.ReadFile("public/index.html")
		if err != nil {
			logger.Errorf("Failed to read public/index.html: %v", err)
			http.Error(w, "Frontend build not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlContent)
	})

	for _, c := range controllers {
		handlerImpl, ok := c.(http.Handler)
		if !ok {
			continue // ensure the controller embeds *handler.Router
		}
		for _, route := range c.InitRoutes() {
			var finalHandler http.Handler = handlerImpl
			for _, mw := range route.Middlewares {
				finalHandler = mw(finalHandler)
			}

			// 组装 Go 1.22 的路由 Pattern
			pattern := route.Path
			if route.Method != "" {
				pattern = route.Method + " " + route.Path
			}
			mux.Handle(pattern, finalHandler)
		}
	}

	// 6. Configure and start HTTP Server
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
