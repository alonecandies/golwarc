package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLClient wraps GORM PostgreSQL database operations
type PostgreSQLClient struct {
	db *gorm.DB
}

// PostgreSQLConfig holds PostgreSQL connection configuration
type PostgreSQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	TimeZone string
}

// NewPostgreSQLClient creates a new PostgreSQL client using GORM
func NewPostgreSQLClient(config PostgreSQLConfig) (*PostgreSQLClient, error) {
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.User,
		config.Password,
		config.Database,
		config.Port,
		config.SSLMode,
		config.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &PostgreSQLClient{db: db}, nil
}

// GetDB returns the underlying GORM database instance
func (c *PostgreSQLClient) GetDB() *gorm.DB {
	return c.db
}

// Close closes the database connection
func (c *PostgreSQLClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate automatically migrates the schema for the given models
func (c *PostgreSQLClient) Migrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

// Ping checks the database connection
func (c *PostgreSQLClient) Ping() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Create inserts a new record
func (c *PostgreSQLClient) Create(value interface{}) error {
	return c.db.Create(value).Error
}

// Find retrieves records based on conditions
func (c *PostgreSQLClient) Find(dest interface{}, conds ...interface{}) error {
	return c.db.Find(dest, conds...).Error
}

// First finds the first record ordered by primary key
func (c *PostgreSQLClient) First(dest interface{}, conds ...interface{}) error {
	return c.db.First(dest, conds...).Error
}

// Update updates attributes with callbacks
func (c *PostgreSQLClient) Update(model interface{}, column string, value interface{}) error {
	return c.db.Model(model).Update(column, value).Error
}

// Updates updates multiple attributes
func (c *PostgreSQLClient) Updates(model interface{}, values interface{}) error {
	return c.db.Model(model).Updates(values).Error
}

// Delete deletes a record
func (c *PostgreSQLClient) Delete(value interface{}, conds ...interface{}) error {
	return c.db.Delete(value, conds...).Error
}

// Transaction executes a function within a transaction
func (c *PostgreSQLClient) Transaction(fn func(*gorm.DB) error) error {
	return c.db.Transaction(fn)
}

// Raw executes raw SQL query
func (c *PostgreSQLClient) Raw(sql string, values ...interface{}) *gorm.DB {
	return c.db.Raw(sql, values...)
}

// Exec executes raw SQL
func (c *PostgreSQLClient) Exec(sql string, values ...interface{}) error {
	return c.db.Exec(sql, values...).Error
}
