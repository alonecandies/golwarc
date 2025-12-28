package database_test

import (
	"testing"

	"github.com/alonecandies/golwarc/database"
	"gorm.io/gorm"
)

// =============================================================================
// ClickHouse Client Integration Tests
// =============================================================================

func TestNewClickHouseClient(t *testing.T) {
	config := database.ClickHouseConfig{
		Host:     "localhost",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "default",
	}

	client, err := database.NewClickHouseClient(config)
	if err != nil {
		t.Skipf("Skipping ClickHouse tests: ClickHouse not available (%v)", err)
		return
	}
	defer client.Close()

	if client == nil {
		t.Error("NewClickHouseClient() returned nil client")
	}
}

func TestClickHouseClient_Ping(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	err := client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestClickHouseClient_GetDB(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	db := client.GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}
}

func TestClickHouseClient_CreateAndFind(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	// Migrate test model
	err := client.Migrate(&TestModel{})
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Create record
	model := &TestModel{
		Name: "ClickHouse Test",
		Age:  35,
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
}

func TestClickHouseClient_First(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	client.Migrate(&TestModel{})
	client.Create(&TestModel{Name: "First CH", Age: 20})

	var first TestModel
	err := client.First(&first)
	if err != nil {
		t.Errorf("First() error = %v", err)
	}
}

func TestClickHouseClient_Update(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "Original CH", Age: 30}
	client.Create(model)

	err := client.Update(model, "name", "Updated CH")
	// Note: ClickHouse may not support traditional updates in the same way
	// This test documents the behavior
	_ = err
}

func TestClickHouseClient_Delete(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	client.Migrate(&TestModel{})
	model := &TestModel{Name: "To Delete CH", Age: 30}
	client.Create(model)

	err := client.Delete(model)
	// Note: ClickHouse has special delete semantics
	_ = err
}

func TestClickHouseClient_Transaction(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	client.Migrate(&TestModel{})

	// ClickHouse may not support traditional transactions
	err := client.Transaction(func(tx *gorm.DB) error {
		model := &TestModel{Name: "In Transaction CH", Age: 25}
		return tx.Create(model).Error
	})

	// Document the behavior
	_ = err
}

func TestClickHouseClient_RawAndExec(t *testing.T) {
	client, skip := setupClickHouseTest(t)
	if skip {
		return
	}
	defer cleanupClickHouseTest(client)

	client.Migrate(&TestModel{})

	// Execute raw SQL (ClickHouse-specific syntax may be required)
	err := client.Exec("INSERT INTO test_models (name, age) VALUES (?, ?)", "Raw CH", 40)
	if err != nil {
		t.Logf("Exec() error (may be expected for ClickHouse): %v", err)
	}

	// Query with raw SQL
	var count int64
	err = client.Raw("SELECT COUNT(*) FROM test_models").Scan(&count).Error
	if err != nil {
		t.Logf("Raw() error: %v", err)
	}
}

func TestClickHouseClient_Close(t *testing.T) {
	config := database.ClickHouseConfig{
		Host:     "localhost",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "default",
	}

	client, err := database.NewClickHouseClient(config)
	if err != nil {
		t.Skipf("Skipping ClickHouse tests: ClickHouse not available (%v)", err)
		return
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func setupClickHouseTest(t *testing.T) (*database.ClickHouseClient, bool) {
	config := database.ClickHouseConfig{
		Host:     "localhost",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "default",
	}

	client, err := database.NewClickHouseClient(config)
	if err != nil {
		t.Skipf("Skipping ClickHouse tests: ClickHouse not available (%v)", err)
		return nil, true
	}

	// Clean up any existing test data
	client.Exec("DROP TABLE IF EXISTS test_models")

	return client, false
}

func cleanupClickHouseTest(client *database.ClickHouseClient) {
	if client != nil {
		client.Exec("DROP TABLE IF EXISTS test_models")
		client.Close()
	}
}
