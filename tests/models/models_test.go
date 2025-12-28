package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alonecandies/golwarc/models"
)

// TestPageTableName tests the TableName method
func TestPageTableName(t *testing.T) {
	page := models.Page{}
	expected := "pages"
	if got := page.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

// TestPageJSONMarshalUnmarshal tests JSON serialization
func TestPageJSONMarshalUnmarshal(t *testing.T) {
	original := models.Page{
		ID:      1,
		URL:     "https://example.com/test",
		Title:   "Test Page",
		Content: "Test content",
		Status:  200,
		Domain:  "example.com",
		HTML:    "<html><body>Test</body></html>",
		Headers: "Content-Type: text/html",
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Page: %v", err)
	}

	// Unmarshal
	var unmarshaled models.Page
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Page: %v", err)
	}

	// Verify key fields
	if unmarshaled.URL != original.URL {
		t.Errorf("URL mismatch: got %v, want %v", unmarshaled.URL, original.URL)
	}
	if unmarshaled.Title != original.Title {
		t.Errorf("Title mismatch: got %v, want %v", unmarshaled.Title, original.Title)
	}
	if unmarshaled.Status != original.Status {
		t.Errorf("Status mismatch: got %v, want %v", unmarshaled.Status, original.Status)
	}
}

// TestPageEmptyValues tests Page with empty/default values
func TestPageEmptyValues(t *testing.T) {
	page := models.Page{}

	if page.URL != "" {
		t.Errorf("Expected empty URL, got %v", page.URL)
	}
	if page.Status != 0 {
		t.Errorf("Expected zero Status, got %v", page.Status)
	}
}

// TestPageAllFields tests Page with all fields populated
func TestPageAllFields(t *testing.T) {
	page := models.Page{
		ID:      123,
		URL:     "https://example.com/full",
		Title:   "Full Page Title",
		Content: "Full page content with lots of text",
		HTML:    "<html><head><title>Full</title></head><body><h1>Full Page</h1></body></html>",
		Headers: "Content-Type: text/html; charset=utf-8\nContent-Length: 1234",
	}
	_ = page.Status    // Just testing struct initialization
	_ = page.Domain    // Just testing struct initialization
	_ = page.CreatedAt // Just testing struct initialization
	_ = page.UpdatedAt // Just testing struct initialization

	// Verify all fields are set
	if page.ID != 123 {
		t.Errorf("ID mismatch: got %v, want 123", page.ID)
	}
	if page.URL != "https://example.com/full" {
		t.Errorf("URL mismatch")
	}
	if page.Title != "Full Page Title" {
		t.Errorf("Title mismatch")
	}
	if page.Content == "" {
		t.Error("Content should not be empty")
	}
	if page.HTML == "" {
		t.Error("HTML should not be empty")
	}
	if page.Headers == "" {
		t.Error("Headers should not be empty")
	}
}

// TestProductTableName tests the TableName method
func TestProductTableName(t *testing.T) {
	product := models.Product{}
	expected := "products"
	if got := product.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

// TestProductJSONMarshalUnmarshal tests JSON serialization
func TestProductJSONMarshalUnmarshal(t *testing.T) {
	original := models.Product{
		ID:          1,
		Name:        "Test Product",
		Price:       99.99,
		Currency:    "USD",
		Description: "A test product",
		ImageURL:    "https://example.com/image.jpg",
		SourceURL:   "https://example.com/product",
		Category:    "Electronics",
		Brand:       "TestBrand",
		SKU:         "TEST-001",
		InStock:     true,
		Rating:      4.5,
		ReviewCount: 42,
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Product: %v", err)
	}

	// Unmarshal
	var unmarshaled models.Product
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Product: %v", err)
	}

	// Verify key fields
	if unmarshaled.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, original.Name)
	}
	if unmarshaled.Price != original.Price {
		t.Errorf("Price mismatch: got %v, want %v", unmarshaled.Price, original.Price)
	}
	if unmarshaled.InStock != original.InStock {
		t.Errorf("InStock mismatch: got %v, want %v", unmarshaled.InStock, original.InStock)
	}
}

// TestProductPriceHandling tests various price scenarios
func TestProductPriceHandling(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		currency string
	}{
		{"Zero price", 0.0, "USD"},
		{"Normal price", 99.99, "USD"},
		{"High price", 9999.99, "EUR"},
		{"Low price", 0.99, "GBP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := models.Product{
				Price:    tt.price,
				Currency: tt.currency,
			}
			_ = product.Name // Just testing struct initialization

			if product.Price != tt.price {
				t.Errorf("Price = %v, want %v", product.Price, tt.price)
			}
			if product.Currency != tt.currency {
				t.Errorf("Currency = %v, want %v", product.Currency, tt.currency)
			}
		})
	}
}

// TestProductRating tests rating field
func TestProductRating(t *testing.T) {
	product := models.Product{
		Rating: 4.75,
	}
	_ = product.Name // Just testing struct initialization

	if product.Rating != 4.75 {
		t.Errorf("Rating = %v, want 4.75", product.Rating)
	}
}

// TestArticleTableName tests the TableName method
func TestArticleTableName(t *testing.T) {
	article := models.Article{}
	expected := "articles"
	if got := article.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

// TestArticleJSONMarshalUnmarshal tests JSON serialization
func TestArticleJSONMarshalUnmarshal(t *testing.T) {
	publishedAt := time.Now()
	original := models.Article{
		ID:          1,
		Title:       "Test Article",
		Author:      "John Doe",
		Content:     "Article content here",
		Summary:     "Article summary",
		PublishedAt: &publishedAt,
		SourceURL:   "https://example.com/article",
		SourceName:  "Example News",
		Category:    "Technology",
		Tags:        "tech,news,ai",
		ImageURL:    "https://example.com/article.jpg",
		Language:    "en",
		WordCount:   500,
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Article: %v", err)
	}

	// Unmarshal
	var unmarshaled models.Article
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Article: %v", err)
	}

	// Verify key fields
	if unmarshaled.Title != original.Title {
		t.Errorf("Title mismatch: got %v, want %v", unmarshaled.Title, original.Title)
	}
	if unmarshaled.Author != original.Author {
		t.Errorf("Author mismatch: got %v, want %v", unmarshaled.Author, original.Author)
	}
	if unmarshaled.WordCount != original.WordCount {
		t.Errorf("WordCount mismatch: got %v, want %v", unmarshaled.WordCount, original.WordCount)
	}
}

// TestArticlePublishedAtNil tests Article with nil PublishedAt
func TestArticlePublishedAtNil(t *testing.T) {
	article := models.Article{
		PublishedAt: nil,
	}
	_ = article.Title  // Just testing struct initialization
	_ = article.Author // Just testing struct initialization

	if article.PublishedAt != nil {
		t.Error("PublishedAt should be nil")
	}
}

// TestArticlePublishedAtSet tests Article with set PublishedAt
func TestArticlePublishedAtSet(t *testing.T) {
	now := time.Now()
	article := models.Article{
		PublishedAt: &now,
	}
	_ = article.Title  // Just testing struct initialization
	_ = article.Author // Just testing struct initialization

	if article.PublishedAt == nil {
		t.Error("PublishedAt should not be nil")
	}
	if !article.PublishedAt.Equal(now) {
		t.Errorf("PublishedAt = %v, want %v", article.PublishedAt, now)
	}
}

// TestArticleTags tests tag handling
func TestArticleTags(t *testing.T) {
	tests := []struct {
		name string
		tags string
	}{
		{"Single tag", "technology"},
		{"Multiple tags", "tech,news,ai,ml"},
		{"Empty tags", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := models.Article{
				Tags: tt.tags,
			}
			_ = article.Title // Just testing struct initialization

			if article.Tags != tt.tags {
				t.Errorf("Tags = %v, want %v", article.Tags, tt.tags)
			}
		})
	}
}

// TestArticleLanguage tests language field
func TestArticleLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
	}{
		{"English", "en"},
		{"Spanish", "es"},
		{"French", "fr"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := models.Article{
				Language: tt.language,
			}
			_ = article.Title // Just testing struct initialization

			if article.Language != tt.language {
				t.Errorf("Language = %v, want %v", article.Language, tt.language)
			}
		})
	}
}

// TestArticleWordCount tests word count field
func TestArticleWordCount(t *testing.T) {
	article := models.Article{
		WordCount: 2,
	}
	_ = article.Title   // Just testing struct initialization
	_ = article.Content // Just testing struct initialization

	if article.WordCount != 2 {
		t.Errorf("WordCount = %v, want 2", article.WordCount)
	}
}

// TestAllModelsTableNames ensures all models have correct table names
func TestAllModelsTableNames(t *testing.T) {
	tests := []struct {
		name      string
		model     interface{ TableName() string }
		wantTable string
	}{
		{"Page", models.Page{}, "pages"},
		{"Product", models.Product{}, "products"},
		{"Article", models.Article{}, "articles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.TableName(); got != tt.wantTable {
				t.Errorf("%s.TableName() = %v, want %v", tt.name, got, tt.wantTable)
			}
		})
	}
}
