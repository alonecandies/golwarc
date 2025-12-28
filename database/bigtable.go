package database

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
)

// BigTableClient wraps Google Cloud BigTable operations
type BigTableClient struct {
	client      *bigtable.Client
	adminClient *bigtable.AdminClient
	project     string
	instance    string
	ctx         context.Context
}

// BigTableConfig holds BigTable connection configuration
type BigTableConfig struct {
	ProjectID  string
	InstanceID string
}

// NewBigTableClient creates a new BigTable client
func NewBigTableClient(config BigTableConfig) (*BigTableClient, error) {
	ctx := context.Background()

	client, err := bigtable.NewClient(ctx, config.ProjectID, config.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigTable client: %w", err)
	}

	adminClient, err := bigtable.NewAdminClient(ctx, config.ProjectID, config.InstanceID)
	if err != nil {
		_ = client.Close() // Best effort cleanup
		return nil, fmt.Errorf("failed to create BigTable admin client: %w", err)
	}

	return &BigTableClient{
		client:      client,
		adminClient: adminClient,
		project:     config.ProjectID,
		instance:    config.InstanceID,
		ctx:         ctx,
	}, nil
}

// CreateTable creates a new table with a column family
func (c *BigTableClient) CreateTable(tableName string, columnFamily string) error {
	if err := c.adminClient.CreateTable(c.ctx, tableName); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	if err := c.adminClient.CreateColumnFamily(c.ctx, tableName, columnFamily); err != nil {
		return fmt.Errorf("failed to create column family: %w", err)
	}

	return nil
}

// DeleteTable deletes a table
func (c *BigTableClient) DeleteTable(tableName string) error {
	return c.adminClient.DeleteTable(c.ctx, tableName)
}

// TableExists checks if a table exists
func (c *BigTableClient) TableExists(tableName string) (bool, error) {
	tables, err := c.adminClient.Tables(c.ctx)
	if err != nil {
		return false, err
	}

	for _, t := range tables {
		if t == tableName {
			return true, nil
		}
	}
	return false, nil
}

// WriteRow writes a row to a table
func (c *BigTableClient) WriteRow(tableName, rowKey string, columnFamily string, data map[string]string) error {
	tbl := c.client.Open(tableName)
	mut := bigtable.NewMutation()
	timestamp := bigtable.Now()

	for column, value := range data {
		mut.Set(columnFamily, column, timestamp, []byte(value))
	}

	return tbl.Apply(c.ctx, rowKey, mut)
}

// WriteRowWithTimestamp writes a row with a specific timestamp
func (c *BigTableClient) WriteRowWithTimestamp(tableName, rowKey string, columnFamily string, data map[string]string, timestamp time.Time) error {
	tbl := c.client.Open(tableName)
	mut := bigtable.NewMutation()
	ts := bigtable.Time(timestamp)

	for column, value := range data {
		mut.Set(columnFamily, column, ts, []byte(value))
	}

	return tbl.Apply(c.ctx, rowKey, mut)
}

// ReadRow reads a row from a table
func (c *BigTableClient) ReadRow(tableName, rowKey string) (bigtable.Row, error) {
	tbl := c.client.Open(tableName)
	return tbl.ReadRow(c.ctx, rowKey)
}

// ReadRows reads multiple rows from a table
func (c *BigTableClient) ReadRows(tableName string, rowKeys []string) ([]bigtable.Row, error) {
	tbl := c.client.Open(tableName)
	rowSet := bigtable.RowList(rowKeys)

	var rows []bigtable.Row
	err := tbl.ReadRows(c.ctx, rowSet, func(row bigtable.Row) bool {
		rows = append(rows, row)
		return true
	})

	return rows, err
}

// ReadRowsWithFilter reads rows with a filter
func (c *BigTableClient) ReadRowsWithFilter(tableName string, rowRange bigtable.RowRange, filter bigtable.Filter) ([]bigtable.Row, error) {
	tbl := c.client.Open(tableName)

	var rows []bigtable.Row
	err := tbl.ReadRows(c.ctx, rowRange, func(row bigtable.Row) bool {
		rows = append(rows, row)
		return true
	}, bigtable.RowFilter(filter))

	return rows, err
}

// DeleteRow deletes a row from a table
func (c *BigTableClient) DeleteRow(tableName, rowKey string) error {
	tbl := c.client.Open(tableName)
	mut := bigtable.NewMutation()
	mut.DeleteRow()

	return tbl.Apply(c.ctx, rowKey, mut)
}

// DeleteCellsInColumn deletes specific cells in a column
func (c *BigTableClient) DeleteCellsInColumn(tableName, rowKey, columnFamily, column string) error {
	tbl := c.client.Open(tableName)
	mut := bigtable.NewMutation()
	mut.DeleteCellsInColumn(columnFamily, column)

	return tbl.Apply(c.ctx, rowKey, mut)
}

// BulkWrite writes multiple rows in bulk
func (c *BigTableClient) BulkWrite(tableName string, rows map[string]*bigtable.Mutation) ([]error, error) {
	tbl := c.client.Open(tableName)

	rowKeys := make([]string, 0, len(rows))
	muts := make([]*bigtable.Mutation, 0, len(rows))

	for key, mut := range rows {
		rowKeys = append(rowKeys, key)
		muts = append(muts, mut)
	}

	errs, err := tbl.ApplyBulk(c.ctx, rowKeys, muts)
	return errs, err
}

// ScanRows scans all rows in a table with a callback
func (c *BigTableClient) ScanRows(tableName string, callback func(row bigtable.Row) bool) error {
	tbl := c.client.Open(tableName)
	return tbl.ReadRows(c.ctx, bigtable.InfiniteRange(""), callback)
}

// GetTable returns a table instance for advanced operations
func (c *BigTableClient) GetTable(tableName string) *bigtable.Table {
	return c.client.Open(tableName)
}

// Close closes the BigTable client connections
func (c *BigTableClient) Close() error {
	if err := c.adminClient.Close(); err != nil {
		return err
	}
	return c.client.Close()
}
