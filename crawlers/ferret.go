package crawlers

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/runtime"
)

// FerretClient wraps Ferret FQL (Ferret Query Language) operations
type FerretClient struct {
	compiler *compiler.Compiler
	ctx      context.Context
}

// FerretConfig holds Ferret configuration
type FerretConfig struct {
	UseCDP bool // Use Chrome DevTools Protocol driver
}

// NewFerretClient creates a new Ferret FQL client
func NewFerretClient(config FerretConfig) (*FerretClient, error) {
	ctx := context.Background()

	comp := compiler.New()

	return &FerretClient{
		compiler: comp,
		ctx:      ctx,
	}, nil
}

// NewDefaultFerretClient creates a Ferret client with HTTP driver
func NewDefaultFerretClient() (*FerretClient, error) {
	return NewFerretClient(FerretConfig{
		UseCDP: false,
	})
}

// Execute executes a Ferret FQL query
func (f *FerretClient) Execute(query string) ([]byte, error) {
	program, err := f.compiler.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile query: %w", err)
	}

	result, err := program.Run(f.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

// ExecuteWithParams executes a Ferret FQL query with parameters
func (f *FerretClient) ExecuteWithParams(query string, params map[string]interface{}) ([]byte, error) {
	program, err := f.compiler.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("failed to compile query: %w", err)
	}

	// Convert params to runtime options
	opts := make([]runtime.Option, 0, len(params))
	for k, v := range params {
		k := k
		v := v
		opts = append(opts, runtime.WithParam(k, v))
	}

	result, err := program.Run(f.ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

// ParseHTML parses HTML content using FQL
func (f *FerretClient) ParseHTML(html string, query string) ([]byte, error) {
	fql := fmt.Sprintf(`
		LET doc = PARSE(%s, "html")
		%s
	`, html, query)

	return f.Execute(fql)
}

// LoadDocument loads a document from a URL using FQL
func (f *FerretClient) LoadDocument(url string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		RETURN doc
	`, url)

	return f.Execute(query)
}

// ExtractLinks extracts all links from a URL
func (f *FerretClient) ExtractLinks(url string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET links = (
			FOR link IN ELEMENTS(doc, "a")
			RETURN link.attributes.href
		)
		RETURN links
	`, url)

	return f.Execute(query)
}

// ExtractText extracts text content from elements matching a selector
func (f *FerretClient) ExtractText(url, selector string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET texts = (
			FOR el IN ELEMENTS(doc, "%s")
			RETURN el.innerText
		)
		RETURN texts
	`, url, selector)

	return f.Execute(query)
}

// ExtractAttributes extracts specific attributes from elements
func (f *FerretClient) ExtractAttributes(url, selector, attribute string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET attrs = (
			FOR el IN ELEMENTS(doc, "%s")
			RETURN el.attributes.%s
		)
		RETURN attrs
	`, url, selector, attribute)

	return f.Execute(query)
}

// ScrapeTable scrapes a table from a URL
func (f *FerretClient) ScrapeTable(url, tableSelector string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET table = ELEMENT(doc, "%s")
		LET rows = (
			FOR row IN ELEMENTS(table, "tr")
			LET cells = (
				FOR cell IN ELEMENTS(row, "td, th")
				RETURN cell.innerText
			)
			RETURN cells
		)
		RETURN rows
	`, url, tableSelector)

	return f.Execute(query)
}

// ExampleExtractArticles demonstrates extracting articles from a news site
func (f *FerretClient) ExampleExtractArticles(url string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET articles = (
			FOR article IN ELEMENTS(doc, "article")
			LET title = ELEMENT(article, "h2")
			LET link = ELEMENT(article, "a")
			LET summary = ELEMENT(article, "p")
			
			RETURN {
				title: title.innerText,
				url: link.attributes.href,
				summary: summary.innerText
			}
		)
		RETURN articles
	`, url)

	return f.Execute(query)
}

// ExampleExtractProducts demonstrates extracting products from an e-commerce site
func (f *FerretClient) ExampleExtractProducts(url string) ([]byte, error) {
	query := fmt.Sprintf(`
		LET doc = DOCUMENT("%s")
		LET products = (
			FOR product IN ELEMENTS(doc, ".product")
			LET name = ELEMENT(product, ".product-name")
			LET price = ELEMENT(product, ".product-price")
			LET image = ELEMENT(product, ".product-image img")
			
			RETURN {
				name: name.innerText,
				price: price.innerText,
				imageUrl: image.attributes.src
			}
		)
		RETURN products
	`, url)

	return f.Execute(query)
}

// GetCompiler returns the underlying Ferret compiler for advanced operations
func (f *FerretClient) GetCompiler() *compiler.Compiler {
	return f.compiler
}

// Close closes the Ferret client
func (f *FerretClient) Close() error {
	// Ferret doesn't require explicit cleanup
	return nil
}

// Note: The following functions require CDP driver and may not work in all environments
// They are kept for reference but users should use the standard Execute methods with FQL

// RegisterDriver is a helper to register custom drivers (advanced usage)
func (f *FerretClient) RegisterDriver(name string, driver drivers.Driver) {
	// This is a placeholder - actual implementation depends on Ferret version
	// Users should refer to Ferret documentation for driver registration
}
