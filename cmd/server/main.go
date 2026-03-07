package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	accounts "canbankx/internal/accounts/domain"
	identityAdapters "canbankx/internal/identity/adapters"
	identity "canbankx/internal/identity/domain"
	identityService "canbankx/internal/identity/service"
	ledger "canbankx/internal/ledger/domain"
	transfers "canbankx/internal/transfers/domain"
)

var DB *gorm.DB

func initDB() {
	// Wait to make sure MySQL is started
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

	r.GET("/health", func(c *gin.Context) {
		status := "UP"
		if err := DB.Exec("SELECT 1").Error; err != nil {
			status = "DOWN (MySQL disconnected)"
		}
		c.JSON(http.StatusOK, gin.H{
			"status": status,
		})
	})

	api := r.Group("/api/v1")

	identityRepo := identityAdapters.NewUserRepository(DB)
	sessionRepo := identityAdapters.NewSessionRepository(DB)
	identitySvc := identityService.NewIdentityService(identityRepo, sessionRepo)
	identityHandler := identityAdapters.NewIdentityHandler(identitySvc)

	identityRoutes := api.Group("/identity")
	identityHandler.RegisterRoutes(identityRoutes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
