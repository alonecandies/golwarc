package database

import "gorm.io/gorm"

// DatabaseClient defines the interface for database operations
// This enables mocking in tests and provides a consistent API across different database implementations
type DatabaseClient interface {
	// GetDB returns the underlying GORM database instance for advanced operations
	GetDB() *gorm.DB

	// Create inserts a new record into the database
	Create(value interface{}) error

	// Find retrieves records based on conditions
	Find(dest interface{}, conds ...interface{}) error

	// First finds the first record ordered by primary key
	First(dest interface{}, conds ...interface{}) error

	// Update updates a single column on a model
	Update(model interface{}, column string, value interface{}) error

	// Updates updates multiple columns on a model
	Updates(model interface{}, values interface{}) error

	// Delete removes a record from the database
	Delete(value interface{}, conds ...interface{}) error

	// Migrate automatically migrates the schema for the given models
	Migrate(models ...interface{}) error

	// Ping checks the database connection
	Ping() error

	// Close closes the database connection
	Close() error

	// Transaction executes a function within a database transaction
	Transaction(fn func(*gorm.DB) error) error
}

// Ensure all database clients implement the interface
var (
	_ DatabaseClient = (*MySQLClient)(nil)
	_ DatabaseClient = (*PostgreSQLClient)(nil)
	_ DatabaseClient = (*ClickHouseClient)(nil)
)
