package database

import (
	"fmt"
	"time"

	"github.com/alonecandies/golwarc/libs"
	"github.com/go-sql-driver/mysql"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLClient wraps GORM MySQL database operations
type MySQLClient struct {
	db *gorm.DB
}

// MySQLConfig holds MySQL connection configuration
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Charset  string
	TLS      *libs.TLSConfig
}

// NewMySQLClient creates a new MySQL client using GORM
func NewMySQLClient(config MySQLConfig) (*MySQLClient, error) {
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}

	tlsParam := ""
	// Configure TLS if enabled
	if config.TLS != nil && config.TLS.Enabled {
		tlsConfig, err := libs.CreateTLSConfig(config.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}

		// Register TLS config with unique name
		tlsConfigName := "custom"
		if err := mysql.RegisterTLSConfig(tlsConfigName, tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to register TLS config: %w", err)
		}
		tlsParam = fmt.Sprintf("&tls=%s", tlsConfigName)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		tlsParam,
	)

	db, err := gorm.Open(mysqldriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &MySQLClient{db: db}, nil
}

// GetDB returns the underlying GORM database instance
func (c *MySQLClient) GetDB() *gorm.DB {
	return c.db
}

// Close closes the database connection
func (c *MySQLClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate automatically migrates the schema for the given models
func (c *MySQLClient) Migrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

// Ping checks the database connection
func (c *MySQLClient) Ping() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Create inserts a new record
func (c *MySQLClient) Create(value interface{}) error {
	return c.db.Create(value).Error
}

// Find retrieves records based on conditions
func (c *MySQLClient) Find(dest interface{}, conds ...interface{}) error {
	return c.db.Find(dest, conds...).Error
}

// First finds the first record ordered by primary key
func (c *MySQLClient) First(dest interface{}, conds ...interface{}) error {
	return c.db.First(dest, conds...).Error
}

// Update updates attributes with callbacks
func (c *MySQLClient) Update(model interface{}, column string, value interface{}) error {
	return c.db.Model(model).Update(column, value).Error
}

// Updates updates multiple attributes
func (c *MySQLClient) Updates(model interface{}, values interface{}) error {
	return c.db.Model(model).Updates(values).Error
}

// Delete deletes a record
func (c *MySQLClient) Delete(value interface{}, conds ...interface{}) error {
	return c.db.Delete(value, conds...).Error
}

// Transaction executes a function within a transaction
func (c *MySQLClient) Transaction(fn func(*gorm.DB) error) error {
	return c.db.Transaction(fn)
}

// Raw executes raw SQL query
func (c *MySQLClient) Raw(sql string, values ...interface{}) *gorm.DB {
	return c.db.Raw(sql, values...)
}

// Exec executes raw SQL
func (c *MySQLClient) Exec(sql string, values ...interface{}) error {
	return c.db.Exec(sql, values...).Error
}
