package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
			log.Println("Successfully connected to MySQL!")

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
				log.Fatalf("Failed to auto-migrate database: %v", err)
			}
			log.Println("Database migration completed successfully.")
			return
		}
		log.Printf("Failed to connect to MySQL (%v). Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to the database after 15 attempts.")
}

func main() {
	log.Println("Starting CanBankX Monolith...")

	initDB()

	r := gin.Default()

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

	log.Printf("Server listening on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
