package benchmarks

import (
	"testing"

	"github.com/alonecandies/golwarc/database"
	"github.com/alonecandies/golwarc/models"
	"gorm.io/gorm"
)

// BenchmarkMySQLCreate benchmarks MySQL create operations
func BenchmarkMySQLCreate(b *testing.B) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_bench",
	})
	if err != nil {
		b.Skip("MySQL not available:", err)
	}
	defer client.Close()

	client.Migrate(&models.Page{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := &models.Page{
			URL:    "https://example.com/page",
			Title:  "Benchmark Page",
			Status: 200,
		}
		client.Create(page)
	}

	b.StopTimer()
	// Cleanup
	client.GetDB().Exec("DELETE FROM pages")
}

// BenchmarkMySQLFind benchmarks MySQL find operations
func BenchmarkMySQLFind(b *testing.B) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_bench",
	})
	if err != nil {
		b.Skip("MySQL not available:", err)
	}
	defer client.Close()

	client.Migrate(&models.Page{})

	// Setup: create some test data
	for i := 0; i < 10; i++ {
		page := &models.Page{
			URL:    "https://example.com",
			Title:  "Test Page",
			Status: 200,
		}
		client.Create(page)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var pages []models.Page
		client.Find(&pages)
	}

	b.StopTimer()
	client.GetDB().Exec("DELETE FROM pages")
}

// BenchmarkMySQLUpdate benchmarks MySQL update operations
func BenchmarkMySQLUpdate(b *testing.B) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_bench",
	})
	if err != nil {
		b.Skip("MySQL not available:", err)
	}
	defer client.Close()

	client.Migrate(&models.Page{})

	page := &models.Page{
		URL:    "https://example.com",
		Title:  "Original Title",
		Status: 200,
	}
	client.Create(page)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Update(page, "title", "Updated Title")
	}

	b.StopTimer()
	client.Delete(page, page.ID)
}

// BenchmarkMySQLTransaction benchmarks MySQL transactions
func BenchmarkMySQLTransaction(b *testing.B) {
	client, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		Database: "golwarc_bench",
	})
	if err != nil {
		b.Skip("MySQL not available:", err)
	}
	defer client.Close()

	client.Migrate(&models.Page{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Transaction(func(tx *gorm.DB) error {
			page := &models.Page{
				URL:    "https://example.com",
				Title:  "Transaction Page",
				Status: 200,
			}
			return tx.Create(page).Error
		})
	}

	b.StopTimer()
	client.GetDB().Exec("DELETE FROM pages")
}

// BenchmarkPostgreSQLCreate benchmarks PostgreSQL create operations
func BenchmarkPostgreSQLCreate(b *testing.B) {
	client, err := database.NewPostgreSQLClient(database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "golwarc_bench",
	})
	if err != nil {
		b.Skip("PostgreSQL not available:", err)
	}
	defer client.Close()

	client.Migrate(&models.Page{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := &models.Page{
			URL:    "https://example.com/page",
			Title:  "Benchmark Page",
			Status: 200,
		}
		client.Create(page)
	}

	b.StopTimer()
	client.GetDB().Exec("DELETE FROM pages")
}
