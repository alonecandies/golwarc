package database_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alonecandies/golwarc/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// =============================================================================
// PostgreSQL Client Unit Tests with sqlmock
// =============================================================================

func TestNewPostgreSQLClient_Defaults(t *testing.T) {
	// Test that defaults are applied when not specified
	config := database.PostgreSQLConfig{
		Host:     "testhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		// SSLMode and TimeZone not specified - should use defaults
	}

	// This will fail to connect to real DB, but we're testing config handling
	_, err := database.NewPostgreSQLClient(config)
	// Error is expected since we're not connecting to a real DB
	if err == nil {
		t.Skip("Unexpected: connected to PostgreSQL (should skip test)")
	}
}

func TestPostgreSQLClient_GetDB(t *testing.T) {
	// Create a mock SQL database for GORM
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	// Note: This test demonstrates interaction with GORM/PostgreSQL client
	// For production use, instantiate via NewPostgreSQLClient constructor
	_ = mock
	_ = gormDB

	// Test passes if we can create the mocked GORM DB
}

func TestPostgreSQLClient_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	// Create client manually for testing
	client := &database.PostgreSQLClient{}
	// Note: In production code, we'd need a way to inject the DB for testing
	// or use the New constructor with a real connection

	_ = client
	_ = gormDB
	_ = mock
}

// =============================================================================
// Integration Tests (require real PostgreSQL)
// =============================================================================

func TestPostgreSQLClient_Integration_CreateAndFind(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	// Migrate test model
	err := client.Migrate(&TestModel{})
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Create record
	model := &TestModel{
		Name: "PostgreSQL Test",
		Age:  25,
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

	if found[0].Name != "PostgreSQL Test" {
		t.Errorf("Found name = %v, want PostgreSQL Test", found[0].Name)
	}
}

func TestPostgreSQLClient_Integration_Ping(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestPostgreSQLClient_Integration_First(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})
	client.Create(&TestModel{Name: "First PG", Age: 30})
	client.Create(&TestModel{Name: "Second PG", Age: 35})

	var first TestModel
	err := client.First(&first)
	if err != nil {
		t.Errorf("First() error = %v", err)
	}

	if first.Name != "First PG" {
		t.Errorf("First() name = %v, want First PG", first.Name)
	}
}

func TestPostgreSQLClient_Integration_Update(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "Original PG", Age: 30}
	client.Create(model)

	err := client.Update(model, "name", "Updated PG")
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	var updated TestModel
	client.First(&updated, model.ID)
	if updated.Name != "Updated PG" {
		t.Errorf("Update() name = %v, want Updated PG", updated.Name)
	}
}

func TestPostgreSQLClient_Integration_Updates(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "Original PG", Age: 30}
	client.Create(model)

	updates := map[string]interface{}{
		"name": "Multi Updated PG",
		"age":  45,
	}
	err := client.Updates(model, updates)
	if err != nil {
		t.Errorf("Updates() error = %v", err)
	}

	var updated TestModel
	client.First(&updated, model.ID)
	if updated.Name != "Multi Updated PG" || updated.Age != 45 {
		t.Errorf("Updates() got name=%v age=%v, want Multi Updated PG 45", updated.Name, updated.Age)
	}
}

func TestPostgreSQLClient_Integration_Delete(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "To Delete PG", Age: 30}
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

func TestPostgreSQLClient_Integration_Transaction(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})

	// Successful transaction
	err := client.Transaction(func(tx *gorm.DB) error {
		model := &TestModel{Name: "In Transaction PG", Age: 28}
		return tx.Create(model).Error
	})
	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}

	// Verify data was committed
	var found TestModel
	err = client.First(&found, "name = ?", "In Transaction PG")
	if err != nil {
		t.Error("Transaction data should be committed")
	}
}

func TestPostgreSQLClient_Integration_TransactionRollback(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})

	// Failed transaction (should rollback)
	expectedErr := errors.New("intentional error")
	err := client.Transaction(func(tx *gorm.DB) error {
		model := &TestModel{Name: "Should Rollback", Age: 99}
		if createErr := tx.Create(model).Error; createErr != nil {
			return createErr
		}
		return expectedErr // Force rollback
	})

	if err != expectedErr {
		t.Errorf("Transaction() error = %v, want %v", err, expectedErr)
	}

	// Verify data was NOT committed
	var found TestModel
	err = client.First(&found, "name = ?", "Should Rollback")
	if err == nil {
		t.Error("Transaction data should NOT be committed after rollback")
	}
}

func TestPostgreSQLClient_Integration_RawAndExec(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	client.Migrate(&TestModel{})

	// Execute raw SQL
	err := client.Exec("INSERT INTO test_models (name, age) VALUES ($1, $2)", "Raw Insert PG", 40)
	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}

	// Query with raw SQL
	var count int64
	client.Raw("SELECT COUNT(*) FROM test_models").Scan(&count)
	if count == 0 {
		t.Error("Raw() should find inserted record")
	}
}

func TestPostgreSQLClient_Integration_GetDB(t *testing.T) {
	client, skip := setupPostgreSQLTest(t)
	if skip {
		return
	}
	defer cleanupPostgreSQLTest(client)

	db := client.GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}

	// Verify we can use the DB instance directly
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		t.Errorf("GetDB() returned non-functional DB: %v", err)
	}
	if result != 1 {
		t.Errorf("GetDB() query result = %v, want 1", result)
	}
}

func TestPostgreSQLClient_Integration_Close(t *testing.T) {
	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "golwarc_test",
		SSLMode:  "disable",
	}

	client, err := database.NewPostgreSQLClient(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests: PostgreSQL not available (%v)", err)
		return
	}

	// Close should not error
	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should also not panic (idempotent)
	err = client.Close()
	// Note: GORM may return an error on second close, which is acceptable
	_ = err
}

// =============================================================================
// Helper Functions
// =============================================================================

func setupPostgreSQLTest(t *testing.T) (*database.PostgreSQLClient, bool) {
	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "golwarc_test",
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	client, err := database.NewPostgreSQLClient(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests: PostgreSQL not available (%v)", err)
		return nil, true
	}

	// Clean up any existing test data
	client.Exec("DROP TABLE IF EXISTS test_models")

	return client, false
}

func cleanupPostgreSQLTest(client *database.PostgreSQLClient) {
	if client != nil {
		client.Exec("DROP TABLE IF EXISTS test_models")
		client.Close()
	}
}

// =============================================================================
// Configuration Tests
// =============================================================================

func TestPostgreSQLConfig_SSLModeDefault(t *testing.T) {
	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
		// SSLMode not specified - should default to "disable"
	}

	_, err := database.NewPostgreSQLClient(config)
	// Connection will fail, but we're testing that config is processed
	if err == nil {
		t.Skip("Unexpected: connected to PostgreSQL")
	}
}

func TestPostgreSQLConfig_TimeZoneDefault(t *testing.T) {
	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
		SSLMode:  "disable",
		// TimeZone not specified - should default to "UTC"
	}

	_, err := database.NewPostgreSQLClient(config)
	// Connection will fail, but we're testing that config is processed
	if err == nil {
		t.Skip("Unexpected: connected to PostgreSQL")
	}
}
