package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	ledgerAdapters "canbankx/internal/ledger/adapters"
	ledger "canbankx/internal/ledger/domain"
	ledgerService "canbankx/internal/ledger/service"
	"canbankx/internal/pkg/cache"
	"canbankx/internal/pkg/middleware"
	transfersAdapters "canbankx/internal/transfers/adapters"
	transfers "canbankx/internal/transfers/domain"
	transfersService "canbankx/internal/transfers/service"

	"canbankx/internal/pkg/kafka"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	if os.Getenv("DB_USER") == "" {
		dsn = "root:rootpassword@tcp(localhost:3306)/transactions_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	for i := 0; i < 15; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			slog.Info("Successfully connected to MySQL (Transactions DB)!")

			sqlDB, err := db.DB()
			if err == nil {
				prometheus.MustRegister(collectors.NewDBStatsCollector(sqlDB, "canbankx_transactions"))
			}

			err = db.AutoMigrate(
				&transfers.Transfer{},
				&ledger.Transaction{},
				&ledger.LedgerEntry{},
			)
			if err != nil {
				slog.Error("Failed to auto-migrate database", "error", err)
				os.Exit(1)
			}
			return
		}
		slog.Warn("Failed to connect to MySQL", "error", err, "retry_in", "2s")
		time.Sleep(2 * time.Second)
	}
	os.Exit(1)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting CanBankX Transactions Microservice...")

	cache.InitRedis()
	initDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(middleware.MetricsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		status := "UP"
		if err := DB.Exec("SELECT 1").Error; err != nil {
			status = "DOWN"
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	ledgerRepo := ledgerAdapters.NewLedgerRepository(DB)
	ledgerSvc := ledgerService.NewLedgerService(ledgerRepo)

	transferRepo := transfersAdapters.NewTransferRepository(DB)
	transferSvc := transfersService.NewTransferService(transferRepo)
	transferHandler := transfersAdapters.NewTransferHandler(transferSvc)

	internal := r.Group("/internal/v1/ledger")
	internal.GET("/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
			return
		}

		entries, balance, apiErr := ledgerSvc.GetAccountHistory(uint(id))
		if apiErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": apiErr.Message})
			return
		}

		c.JSON(http.StatusOK, gin.H{"entries": entries, "balance": balance})
	})

	api := r.Group("/api/v1")
	protected := api.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())

	transferRoutes := protected.Group("/transfers")
	transferHandler.RegisterRoutes(transferRoutes)

	bbcRoutes := protected.Group("/bbc")
	transfersAdapters.RegisterBBCRoutes(bbcRoutes)

	kafkaHost := os.Getenv("KAFKA_URL")
	if kafkaHost == "" {
		kafkaHost = "kafka:9092"
	}
	kafka.StartConsumer(DB, kafkaHost, "bbc.events.2", "sam-bank-group")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	slog.Info("Transactions Server listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
