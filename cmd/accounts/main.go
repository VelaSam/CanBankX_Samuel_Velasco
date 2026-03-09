package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "canbankx/docs" // Swagger docs

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
	ledgerDomain "canbankx/internal/ledger/domain"
	ledgerService "canbankx/internal/ledger/service"
	"canbankx/internal/pkg/cache"
	apierrors "canbankx/internal/pkg/errors"
	"canbankx/internal/pkg/middleware"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	if os.Getenv("DB_USER") == "" {
		dsn = "root:rootpassword@tcp(localhost:3306)/accounts_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	for i := 0; i < 15; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			slog.Info("Successfully connected to MySQL (Accounts DB)!")

			sqlDB, err := db.DB()
			if err == nil {
				prometheus.MustRegister(collectors.NewDBStatsCollector(sqlDB, "canbankx_accounts"))
			}

			err = db.AutoMigrate(
				&accounts.Account{},
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

type httpLedgerService struct {
	baseURL string
}

type ledgerResponse struct {
	Entries []ledgerDomain.LedgerEntry `json:"entries"`
	Balance float64                    `json:"balance"`
}

func (s *httpLedgerService) GetAccountHistory(accountID uint) ([]ledgerDomain.LedgerEntry, float64, *apierrors.APIError) {
	url := fmt.Sprintf("%s/internal/v1/ledger/%d", s.baseURL, accountID)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		slog.Error("Failed to fetch ledger history from Transactions API", "error", err)
		return nil, 0, apierrors.ErrInternalServer
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data ledgerResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, 0, apierrors.ErrInternalServer
	}

	return data.Entries, data.Balance, nil
}

func setupInternalRoutes(router *gin.Engine) {
	internal := router.Group("/internal/v1/accounts")

	internal.GET("/:id/balance", func(c *gin.Context) {
		id := c.Param("id")
		var acc accounts.Account
		if err := DB.First(&acc, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, acc)
	})

	type TransferReq struct {
		SourceAccount uint    `json:"sourceAccount"`
		TargetAccount uint    `json:"targetAccount"`
		Amount        float64 `json:"amount"`
	}

	internal.POST("/transfer", func(c *gin.Context) {
		var req TransferReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		err := DB.Transaction(func(tx *gorm.DB) error {
			var source accounts.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&source, req.SourceAccount).Error; err != nil {
				return fmt.Errorf("source account not found")
			}

			var target accounts.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&target, req.TargetAccount).Error; err != nil {
				return fmt.Errorf("target account not found")
			}

			if source.Balance < req.Amount {
				return fmt.Errorf("insufficient funds")
			}

			source.Balance -= req.Amount
			target.Balance += req.Amount

			if err := tx.Save(&source).Error; err != nil {
				return err
			}
			return tx.Save(&target).Error
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting CanBankX Accounts Microservice...")

	cache.InitRedis()
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
			status = "DOWN"
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	repo := accountsAdapters.NewAccountRepository(DB)

	setupInternalRoutes(r)

	transactionsURL := os.Getenv("TRANSACTIONS_SERVICE_URL")
	if transactionsURL == "" {
		transactionsURL = "http://transactions-api:8083"
	}

	var ledgerSvc ledgerService.LedgerService = &httpLedgerService{baseURL: transactionsURL}

	accountSvc := accountsService.NewAccountService(repo, ledgerSvc)
	accountHandler := accountsAdapters.NewAccountHandler(accountSvc)

	api := r.Group("/api/v1")
	protected := api.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())

	accountRoutes := protected.Group("/accounts")
	accountHandler.RegisterRoutes(accountRoutes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	slog.Info("Accounts Server listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
