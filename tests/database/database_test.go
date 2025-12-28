package database_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/alonecandies/golwarc/database"
	"gorm.io/gorm"
)

// =====================
// MySQL Unit Tests
// =====================

func TestMySQLConfig(t *testing.T) {
	config := database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "testdb",
		Charset:  "utf8mb4",
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", config.Host)
	}
	if config.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", config.Port)
	}
	if config.User != "root" {
		t.Errorf("Expected user 'root', got %s", config.User)
	}
	if config.Password != "password" {
		t.Errorf("Expected password 'password', got %s", config.Password)
	}
	if config.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got %s", config.Database)
	}
	if config.Charset != "utf8mb4" {
		t.Errorf("Expected charset 'utf8mb4', got %s", config.Charset)
	}
}

func TestMySQLClientConnection(t *testing.T) {
	// Skip if MySQL is not available
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer client.Close()

	// Test ping
	if err := client.Ping(); err != nil {
		t.Errorf("Failed to ping MySQL: %v", err)
	}
}

func TestMySQLCRUD(t *testing.T) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer client.Close()

	// Create test table
	type TestRecord struct {
		ID        uint   `gorm:"primaryKey"`
		Name      string `gorm:"size:100"`
		CreatedAt time.Time
	}

	// Migrate
	if migrateErr := client.Migrate(&TestRecord{}); migrateErr != nil {
		t.Fatalf("Failed to migrate: %v", migrateErr)
	}

	// Create
	record := &TestRecord{Name: "test_record"}
	if createErr := client.Create(record); createErr != nil {
		t.Fatalf("Failed to create record: %v", createErr)
	}
	if record.ID == 0 {
		t.Error("Expected ID to be set after create")
	}

	// Find
	var found TestRecord
	if findErr := client.First(&found, record.ID); findErr != nil {
		t.Fatalf("Failed to find record: %v", findErr)
	}
	if found.Name != "test_record" {
		t.Errorf("Expected name 'test_record', got %s", found.Name)
	}

	// Update
	if updateErr := client.Update(&found, "name", "updated_record"); updateErr != nil {
		t.Fatalf("Failed to update record: %v", updateErr)
	}

	// Verify update
	var updated TestRecord
	client.First(&updated, record.ID)
	if updated.Name != "updated_record" {
		t.Errorf("Expected name 'updated_record', got %s", updated.Name)
	}

	// Delete
	if deleteErr := client.Delete(&TestRecord{}, record.ID); deleteErr != nil {
		t.Fatalf("Failed to delete record: %v", deleteErr)
	}

	// Verify deletion
	var deleted TestRecord
	err = client.First(&deleted, record.ID)
	if err == nil {
		t.Error("Expected record to be deleted")
	}
}

func TestMySQLTransaction(t *testing.T) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("MySQL not available:", err)
	}
	defer client.Close()

	type TxRecord struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}
	client.Migrate(&TxRecord{})

	// Test successful transaction
	err = client.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&TxRecord{Name: "tx_record"}).Error
	})
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Test rollback on error
	err = client.Transaction(func(tx *gorm.DB) error {
		tx.Create(&TxRecord{Name: "should_rollback"})
		return fmt.Errorf("intentional error")
	})
	if err == nil {
		t.Error("Expected transaction to fail")
	}
}

// =====================
// PostgreSQL Unit Tests
// =====================

func TestPostgreSQLConfig(t *testing.T) {
	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", config.Port)
	}
	if config.User != "postgres" {
		t.Errorf("Expected user 'postgres', got %s", config.User)
	}
	if config.Password != "password" {
		t.Errorf("Expected password 'password', got %s", config.Password)
	}
	if config.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got %s", config.Database)
	}
}

func TestPostgreSQLClientConnection(t *testing.T) {
	client, err := database.NewPostgreSQLClient(database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "golwarc_test",
	})
	if err != nil {
		t.Skip("PostgreSQL not available:", err)
	}
	defer client.Close()

	if err := client.Ping(); err != nil {
		t.Errorf("Failed to ping PostgreSQL: %v", err)
	}
}

// =====================
// ClickHouse Unit Tests
// =====================

func TestClickHouseConfig(t *testing.T) {
	config := database.ClickHouseConfig{
		Host:     "localhost",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "default",
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %s", config.Host)
	}
	if config.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", config.Port)
	}
	if config.User != "default" {
		t.Errorf("Expected user 'default', got %s", config.User)
	}
	if config.Password != "" {
		t.Errorf("Expected password '', got %s", config.Password)
	}
	if config.Database != "default" {
		t.Errorf("Expected database 'default', got %s", config.Database)
	}
}

func TestClickHouseClientConnection(t *testing.T) {
	client, err := database.NewClickHouseClient(database.ClickHouseConfig{
		Host:     "localhost",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "default",
	})
	if err != nil {
		t.Skip("ClickHouse not available:", err)
	}
	defer client.Close()

	if err := client.Ping(); err != nil {
		t.Errorf("Failed to ping ClickHouse: %v", err)
	}
}
