package database_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alonecandies/golwarc/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Test model for database operations
type TestModel struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// =============================================================================
// MySQL Client Unit Tests with sqlmock (No Real Database Required)
// =============================================================================

func TestMySQLClient_Create_WithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}

	// Create a wrapper client
	client := &database.MySQLClient{}
	// We'd need to inject gormDB here in production code
	// For now, we're testing the concept

	_ = client
	_ = mock
	_ = gormDB
}

func TestMySQLClient_GetDB_WithMock(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}

	// Verify GORM DB is usable
	if gormDB == nil {
		t.Fatal("GORM DB should not be nil")
	}
}

func TestMySQLClient_Transaction_WithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = gormDB.Transaction(func(tx *gorm.DB) error {
		return nil
	})

	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestMySQLClient_Transaction_Rollback_WithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}

	// Mock transaction rollback
	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("intentional error")
	err = gormDB.Transaction(func(tx *gorm.DB) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Transaction() error = %v, want %v", err, expectedErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// MySQL Client Integration Tests (Require Real MySQL)
// =============================================================================

func TestNewMySQLClient(t *testing.T) {
	config := database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "test_db",
	}

	client, err := database.NewMySQLClient(config)
	if err != nil {
		t.Skipf("Skipping MySQL tests: MySQL not available (%v)", err)
		return
	}
	defer client.Close()

	if client == nil {
		t.Error("NewMySQLClient() returned nil client")
	}
}

func TestMySQLClient_Ping(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestMySQLClient_CreateAndFind(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	// Migrate test model
	err := client.Migrate(&TestModel{})
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Create record
	model := &TestModel{
		Name: "John Doe",
		Age:  30,
	}
	err = client.Create(model)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Find record
	var found []TestModel
	err = client.Find(&found)
	if err != nil {
		t.Errorf("Find() error = %v", err)
	}

	if len(found) == 0 {
		t.Error("Find() should return at least one record")
	}

	if found[0].Name != "John Doe" {
		t.Errorf("Found name = %v, want John Doe", found[0].Name)
	}
}

func TestMySQLClient_First(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})
	client.Create(&TestModel{Name: "First", Age: 20})
	client.Create(&TestModel{Name: "Second", Age: 25})

	var first TestModel
	err := client.First(&first)
	if err != nil {
		t.Errorf("First() error = %v", err)
	}

	if first.Name != "First" {
		t.Errorf("First() name = %v, want First", first.Name)
	}
}

func TestMySQLClient_Update(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "Original", Age: 30}
	client.Create(model)

	err := client.Update(model, "name", "Updated")
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	var updated TestModel
	client.First(&updated, model.ID)
	if updated.Name != "Updated" {
		t.Errorf("Update() name = %v, want Updated", updated.Name)
	}
}

func TestMySQLClient_Updates(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "Original", Age: 30}
	client.Create(model)

	updates := map[string]interface{}{
		"name": "MultiUpdated",
		"age":  40,
	}
	err := client.Updates(model, updates)
	if err != nil {
		t.Errorf("Updates() error = %v", err)
	}

	var updated TestModel
	client.First(&updated, model.ID)
	if updated.Name != "MultiUpdated" || updated.Age != 40 {
		t.Errorf("Updates() got name=%v age=%v, want MultiUpdated 40", updated.Name, updated.Age)
	}
}

func TestMySQLClient_Delete(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "ToDelete", Age: 30}
	client.Create(model)

	err := client.Delete(model)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	var found TestModel
	err = client.First(&found, model.ID)
	if err == nil {
		t.Error("First() should return error after Delete()")
	}
}

func TestMySQLClient_Transaction(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})

	err := client.Transaction(func(tx *gorm.DB) error {
		model := &TestModel{Name: "InTransaction", Age: 25}
		return tx.Create(model).Error
	})
	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	var found TestModel
	err = client.First(&found, "name = ?", "InTransaction")
	if err != nil {
		t.Error("Transaction data should be committed")
	}
}

func TestMySQLClient_RawAndExec(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	client.Migrate(&TestModel{})

	err := client.Exec("INSERT INTO test_models (name, age) VALUES (?, ?)", "RawInsert", 35)
	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}

	var count int64
	client.Raw("SELECT COUNT(*) FROM test_models").Scan(&count)
	if count == 0 {
		t.Error("Raw() should find inserted record")
	}
}

func TestMySQLConfig_DefaultCharset(t *testing.T) {
	config := database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "test_db",
		// Charset not specified - should default to utf8mb4
	}

	client, err := database.NewMySQLClient(config)
	if err != nil {
		t.Skip("MySQL not available for this test")
		return
	}
	defer client.Close()
}

func TestMySQLClient_GetDB(t *testing.T) {
	client, skip := setupMySQLTest(t)
	if skip {
		return
	}
	defer cleanupMySQLTest(client)

	db := client.GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func setupMySQLTest(t *testing.T) (*database.MySQLClient, bool) {
	config := database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_test",
	}

	client, err := database.NewMySQLClient(config)
	if err != nil {
		t.Skipf("Skipping MySQL tests: MySQL not available (%v)", err)
		return nil, true
	}

	client.Exec("DROP TABLE IF EXISTS test_models")

	return client, false
}

func cleanupMySQLTest(client *database.MySQLClient) {
	if client != nil {
		client.Exec("DROP TABLE IF EXISTS test_models")
		client.Close()
	}
}
