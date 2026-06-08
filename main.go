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

	// Seed default administrator if DB is empty
	db.SeedAdminUser(adminRepo)

	// 4. Initialize services (Business Logic Layer)
	appService := service.NewAppService(appRepo)
	tokenService := service.NewTokenService(tokenRepo, appRepo, blacklistRepo, logRepo)
	adminService := service.NewAdminService(adminRepo)
	calendarService := service.NewCalendarService(calendarRepo)
	holidayService := service.NewHolidayService(holidayRepo)

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
		// 开放API接口
		api.NewResourceHandler(apiBase, calendarService),
		api.NewCalendarHandler(apiBase, calendarService, holidayService, clientAuth),
	}

	// 6. Register routes
	mux := http.NewServeMux()

	// TODO
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
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
