package database

import (
	"fmt"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ClickHouseClient wraps GORM ClickHouse database operations
type ClickHouseClient struct {
	db *gorm.DB
}

// ClickHouseConfig holds ClickHouse connection configuration
type ClickHouseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// NewClickHouseClient creates a new ClickHouse client using GORM
func NewClickHouseClient(config ClickHouseConfig) (*ClickHouseClient, error) {
	dsn := fmt.Sprintf("tcp://%s:%d?database=%s&username=%s&password=%s&read_timeout=10&write_timeout=20",
		config.Host,
		config.Port,
		config.Database,
		config.User,
		config.Password,
	)

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &ClickHouseClient{db: db}, nil
}

// GetDB returns the underlying GORM database instance
func (c *ClickHouseClient) GetDB() *gorm.DB {
	return c.db
}

// Close closes the database connection
func (c *ClickHouseClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateTable creates a table for the given model
func (c *ClickHouseClient) CreateTable(model interface{}) error {
	return c.db.AutoMigrate(model)
}

// Ping checks the database connection
func (c *ClickHouseClient) Ping() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Create inserts a new record
func (c *ClickHouseClient) Create(value interface{}) error {
	return c.db.Create(value).Error
}

// CreateInBatches inserts records in batches (efficient for ClickHouse)
func (c *ClickHouseClient) CreateInBatches(value interface{}, batchSize int) error {
	return c.db.CreateInBatches(value, batchSize).Error
}

// Find retrieves records based on conditions
func (c *ClickHouseClient) Find(dest interface{}, conds ...interface{}) error {
	return c.db.Find(dest, conds...).Error
}

// First finds the first record
func (c *ClickHouseClient) First(dest interface{}, conds ...interface{}) error {
	return c.db.First(dest, conds...).Error
}

// Where adds a WHERE clause
func (c *ClickHouseClient) Where(query interface{}, args ...interface{}) *gorm.DB {
	return c.db.Where(query, args...)
}

// Raw executes raw SQL query
func (c *ClickHouseClient) Raw(sql string, values ...interface{}) *gorm.DB {
	return c.db.Raw(sql, values...)
}

// Exec executes raw SQL
func (c *ClickHouseClient) Exec(sql string, values ...interface{}) error {
	return c.db.Exec(sql, values...).Error
}

// Delete deletes records (Note: ClickHouse has limited DELETE support)
func (c *ClickHouseClient) Delete(value interface{}, conds ...interface{}) error {
	return c.db.Delete(value, conds...).Error
}

// Count counts records
func (c *ClickHouseClient) Count(model interface{}, count *int64, conds ...interface{}) error {
	return c.db.Model(model).Where(conds).Count(count).Error
}

// Update updates a single column (Note: ClickHouse has limited UPDATE support)
func (c *ClickHouseClient) Update(model interface{}, column string, value interface{}) error {
	return c.db.Model(model).Update(column, value).Error
}

// Updates updates multiple columns (Note: ClickHouse has limited UPDATE support)
func (c *ClickHouseClient) Updates(model interface{}, values interface{}) error {
	return c.db.Model(model).Updates(values).Error
}

// Migrate automatically migrates the schema for the given models
func (c *ClickHouseClient) Migrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

// Transaction executes a function within a transaction
// Note: ClickHouse has limited transaction support
func (c *ClickHouseClient) Transaction(fn func(*gorm.DB) error) error {
	return c.db.Transaction(fn)
}
