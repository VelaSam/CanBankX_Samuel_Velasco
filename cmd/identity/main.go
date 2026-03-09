package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "canbankx/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	identityAdapters "canbankx/internal/identity/adapters"
	identity "canbankx/internal/identity/domain"
	identityService "canbankx/internal/identity/service"
	"canbankx/internal/pkg/middleware"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	if os.Getenv("DB_USER") == "" {
		dsn = "root:rootpassword@tcp(localhost:3306)/identity_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	for i := 0; i < 15; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			slog.Info("Successfully connected to MySQL (Identity DB)!")

			sqlDB, err := db.DB()
			if err == nil {
				prometheus.MustRegister(collectors.NewDBStatsCollector(sqlDB, "canbankx_identity"))
				slog.Info("Registered DB metrics collector for Prometheus")
			}

			err = db.AutoMigrate(
				&identity.User{},
				&identity.Session{},
				&identity.KYCDocument{},
			)
			if err != nil {
				slog.Error("Failed to auto-migrate database", "error", err)
				os.Exit(1)
			}
			slog.Info("Database migration completed successfully.")
			return
		}
		slog.Warn("Failed to connect to MySQL", "error", err, "retry_in", "2s")
		time.Sleep(2 * time.Second)
	}
	slog.Error("Could not connect to the database after 15 attempts.")
	os.Exit(1)
}

// @title CanBankX API (Microservices Edition)
// @version 3.0
// @description API de la plateforme bancaire CanBankX
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@canbankx.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting CanBankX Identity Microservice...")

	initDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(middleware.MetricsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		status := "UP"
		if err := DB.Exec("SELECT 1").Error; err != nil {
			status = "DOWN (MySQL disconnected)"
		}
		c.JSON(http.StatusOK, gin.H{
			"status": status,
		})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	identityRepo := identityAdapters.NewUserRepository(DB)
	sessionRepo := identityAdapters.NewSessionRepository(DB)
	identitySvc := identityService.NewIdentityService(identityRepo, sessionRepo)
	identityHandler := identityAdapters.NewIdentityHandler(identitySvc)

	identityRoutes := api.Group("/identity")
	identityHandler.RegisterRoutes(identityRoutes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("Identity Server listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
