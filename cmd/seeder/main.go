package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"

	"canbankx/internal/accounts/domain"
	identityAdmin "canbankx/internal/identity/domain"
	ledger "canbankx/internal/ledger/domain"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBIdentity *gorm.DB
var DBAccounts *gorm.DB
var DBTransactions *gorm.DB

func initDB(dbName string, defaultPort string) *gorm.DB {
	host := os.Getenv("DB_HOST")
	if host == "db" || host == "" {
		host = "127.0.0.1"
	}
	port := defaultPort

	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + host + ":" + port + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	if os.Getenv("DB_USER") == "" {
		dsn = "canbankx_user:SecurePassword123!@tcp(" + host + ":" + port + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	}

	for i := 0; i < 15; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			slog.Info(fmt.Sprintf("Successfully connected to MySQL for %s!", dbName))
			return db
		}
		slog.Warn("Failed to connect to MySQL", "db", dbName, "error", err, "retry_in", "2s")
		time.Sleep(2 * time.Second)
	}
	slog.Error("Could not connect to the database after 15 attempts.", "db", dbName)
	os.Exit(1)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Data Seeder (Microservices Edition)...")

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found or unable to load, relying on system environment variables")
	}

	DBIdentity = initDB("identity_db", "3306")
	DBAccounts = initDB("accounts_db", "3307")
	DBTransactions = initDB("transactions_db", "3308")

	DBIdentity.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	DBIdentity.Exec("TRUNCATE TABLE users;")
	DBIdentity.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	DBAccounts.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	DBAccounts.Exec("TRUNCATE TABLE accounts;")
	DBAccounts.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	DBTransactions.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	DBTransactions.Exec("TRUNCATE TABLE transfers;")
	DBTransactions.Exec("TRUNCATE TABLE ledger_entries;")
	DBTransactions.Exec("TRUNCATE TABLE transactions;")
	DBTransactions.Exec("SET FOREIGN_KEY_CHECKS = 1;")

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

	DBIdentity.CreateInBatches(users, 100)

	slog.Info("Generating accounts and initial ledger deposits...")

	for _, u := range users {
		for j := 0; j < accountsPerUser; j++ {
			accType := "CHECKING"
			if j == 1 {
				accType = "SAVINGS"
			}
			acc := domain.Account{
				UserID:  u.ID,
				Type:    accType,
				Status:  "ACTIVE",
				Balance: initialBalance,
			}
			DBAccounts.Create(&acc)
			accountIDs = append(accountIDs, acc.ID)

			tx := ledger.Transaction{
				Reference: fmt.Sprintf("INIT-DEPOSIT-U%d-A%d", u.ID, acc.ID),
				Type:      "DEPOSIT",
			}
			DBTransactions.Create(&tx)

			entry := ledger.LedgerEntry{
				TransactionID: tx.ID,
				AccountID:     acc.ID,
				Amount:        initialBalance,
			}
			DBTransactions.Create(&entry)
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
