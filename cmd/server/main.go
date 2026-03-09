package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "canbankx/docs" // Swagger docs

	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	accountsAdapters "canbankx/internal/accounts/adapters"
	accounts "canbankx/internal/accounts/domain"
	accountsService "canbankx/internal/accounts/service"
	identityAdapters "canbankx/internal/identity/adapters"
	identity "canbankx/internal/identity/domain"
	identityService "canbankx/internal/identity/service"
	ledgerAdapters "canbankx/internal/ledger/adapters"
	ledger "canbankx/internal/ledger/domain"
	ledgerService "canbankx/internal/ledger/service"
	"canbankx/internal/pkg/middleware"
	transfersAdapters "canbankx/internal/transfers/adapters"
	transfers "canbankx/internal/transfers/domain"
	transfersService "canbankx/internal/transfers/service"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	if os.Getenv("DB_USER") == "" {
		dsn = "root:rootpassword@tcp(localhost:3306)/canbankx?charset=utf8mb4&parseTime=True&loc=Local"
	}

	for i := 0; i < 15; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			slog.Info("Successfully connected to MySQL!")

			sqlDB, err := db.DB()
			if err == nil {
				prometheus.MustRegister(collectors.NewDBStatsCollector(sqlDB, "canbankx"))
				slog.Info("Registered DB metrics collector for Prometheus")
			}

			err = db.AutoMigrate(
				&identity.User{},
				&identity.Session{},
				&identity.KYCDocument{},
				&accounts.Account{},
				&transfers.Transfer{},
				&ledger.Transaction{},
				&ledger.LedgerEntry{},
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

// @title CanBankX API
// @version 1.0
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

	// Attempt to load .env file if it exists
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or unable to load, relying on system environment variables")
	}

	slog.Info("Starting CanBankX Monolith...")

	initDB()

	r := gin.Default()

	// Configure CORS
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

	// DI for Ledger
	ledgerRepo := ledgerAdapters.NewLedgerRepository(DB)
	ledgerSvc := ledgerService.NewLedgerService(ledgerRepo)

	// DI for Accounts
	accountRepo := accountsAdapters.NewAccountRepository(DB)
	accountSvc := accountsService.NewAccountService(accountRepo, ledgerSvc)
	accountHandler := accountsAdapters.NewAccountHandler(accountSvc)

	// DI for Transfers
	transferRepo := transfersAdapters.NewTransferRepository(DB)
	transferSvc := transfersService.NewTransferService(transferRepo)
	transferHandler := transfersAdapters.NewTransferHandler(transferSvc)

	// Protected Routes
	protected := api.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())

	// Routes for Accounts
	accountRoutes := protected.Group("/accounts")
	accountHandler.RegisterRoutes(accountRoutes)

	// Routes for Transfers
	transferRoutes := protected.Group("/transfers")
	transferHandler.RegisterRoutes(transferRoutes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Server listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
