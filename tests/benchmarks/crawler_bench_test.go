package benchmarks

import (
	"testing"

	"github.com/alonecandies/golwarc/crawlers"
	"github.com/gocolly/colly/v2"
)

// BenchmarkCollySetup benchmarks Colly client setup
func BenchmarkCollySetup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = crawlers.NewDefaultCollyClient()
	}
}

// BenchmarkCollyHTMLCallback benchmarks Colly HTML callbacks
func BenchmarkCollyHTMLCallback(b *testing.B) {
	client := crawlers.NewDefaultCollyClient()

	client.OnHTML("title", func(e *colly.HTMLElement) {
		_ = e.Text
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark overhead of callbacks
	}
}

// BenchmarkCollyLinkExtraction benchmarks link extraction with Colly
func BenchmarkCollyLinkExtraction(b *testing.B) {
	client := crawlers.NewDefaultCollyClient()

	client.OnHTML("a[href]", func(e *colly.HTMLElement) {
		_ = e.Attr("href")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//Benchmark callback overhead
	}
}
