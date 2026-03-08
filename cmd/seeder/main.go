package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"canbankx/internal/accounts/domain"
	identityAdmin "canbankx/internal/identity/domain"
	ledger "canbankx/internal/ledger/domain"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
			slog.Info("Successfully connected to MySQL for Seeding!")
			return
		}
		slog.Warn("Failed to connect to MySQL", "error", err, "retry_in", "2s")
		time.Sleep(2 * time.Second)
	}
	slog.Error("Could not connect to the database after 15 attempts.")
	os.Exit(1)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Data Seeder...")
	initDB()

	//truncate existing data to start fresh pour les load test (pas obligatoire, but good for consistent load tests)
	DB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	DB.Exec("TRUNCATE TABLE users;")
	DB.Exec("TRUNCATE TABLE accounts;")
	DB.Exec("TRUNCATE TABLE transfers;")
	DB.Exec("TRUNCATE TABLE ledger_entries;")
	DB.Exec("TRUNCATE TABLE transactions;")
	DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	numUsers := 1000
	accountsPerUser := 2
	initialBalance := 10000.0

	var accountIDs []uint

	users := make([]*identityAdmin.User, 0, numUsers)

	slog.Info(fmt.Sprintf("Generating %d users...", numUsers))

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)

	for i := 1; i <= numUsers; i++ {
		users = append(users, &identityAdmin.User{
			FirstName:    fmt.Sprintf("User%d", i),
			LastName:     "Test",
			Email:        fmt.Sprintf("user%d@test.com", i),
			PasswordHash: string(passwordHash),
			SIN:          "000000000",
			Status:       "ACTIVE",
		})
	}

	DB.CreateInBatches(users, 100)

	slog.Info("Generating accounts and initial ledger deposits...")

	for _, u := range users {
		for j := 0; j < accountsPerUser; j++ {
			accType := "CHECKING"
			if j == 1 {
				accType = "SAVINGS"
			}
			acc := domain.Account{
				UserID: u.ID,
				Type:   accType,
				Status: "ACTIVE",
			}
			DB.Create(&acc)
			accountIDs = append(accountIDs, acc.ID)

			tx := ledger.Transaction{
				Reference: fmt.Sprintf("INIT-DEPOSIT-U%d-A%d", u.ID, acc.ID),
				Type:      "DEPOSIT",
			}
			DB.Create(&tx)

			entry := ledger.LedgerEntry{
				TransactionID: tx.ID,
				AccountID:     acc.ID,
				Amount:        initialBalance,
			}
			DB.Create(&entry)
		}
	}

	slog.Info("Seeding completed successfully!")

	fileData := map[string]interface{}{
		"accountIds": accountIDs,
	}
	bytes, _ := json.MarshalIndent(fileData, "", "  ")
	os.WriteFile("tests/k6/users.json", bytes, 0644)

	slog.Info("Exported account IDs to tests/k6/users.json")
}
