package scraper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var productIDFromURL = regexp.MustCompile(`/(\d+)/product-detail`)

type Scraper struct {
	client     *http.Client
	url        string
	headers    *HeaderPool
	retries    int
	retryDelay time.Duration
}

func New(url string, timeout time.Duration, maxRetries int, retryDelay time.Duration) *Scraper {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	}
	return &Scraper{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		url:        url,
		headers:    NewHeaderPool(),
		retries:    maxRetries,
		retryDelay: retryDelay,
	}
}

// FetchProducts scrapes the FirstCry Hot Wheels page and returns all visible products.
func (s *Scraper) FetchProducts(ctx context.Context) ([]*Product, error) {
	var lastErr error

	for attempt := 0; attempt < s.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.retryDelay):
			}
			log.Printf("[Scraper] Retry %d/%d after error: %v", attempt, s.retries, lastErr)
		}

		products, err := s.fetchOnce(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		return products, nil
	}

	return nil, fmt.Errorf("all %d attempts failed, last error: %w", s.retries, lastErr)
}

func (s *Scraper) fetchOnce(ctx context.Context) ([]*Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	s.headers.Apply(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	return s.parseProducts(doc), nil
}

// parseProducts extracts product data from FirstCry listing HTML.
func (s *Scraper) parseProducts(doc *goquery.Document) []*Product {
	var products []*Product

	doc.Find("div.list_block").Each(func(_ int, sel *goquery.Selection) {
		product := &Product{}

		outStock, _ := sel.Attr("data-outstock")
		product.InStock = outStock != "true"

		inner := sel.Find("div.li_inner_block").First()
		if inner.Length() == 0 {
			inner = sel
		}

		product.ID = extractProductID(inner)
		if product.ID == "" {
			return
		}

		link := inner.Find("a[href*='product-detail']").First()
		href, _ := link.Attr("href")
		if href != "" {
			product.URL = normalizeURL(href)
		}
		if product.URL == "" {
			return
		}
		if product.ID == "" {
			product.ID = extractIDFromURL(product.URL)
		}
		if product.ID == "" {
			return
		}

		nameLink := inner.Find("div.li_txt1 a").First()
		if title, ok := nameLink.Attr("title"); ok && title != "" {
			product.Name = strings.TrimSpace(title)
		} else {
			product.Name = strings.TrimSpace(nameLink.Text())
		}

		product.Price = cleanPrice(inner.Find("div.rupee span.r1").First().Text())
		if product.Price == "" {
			product.Price = extractPriceFromAriaLabel(inner.Find("div.rupee").First().AttrOr("aria-label", ""), "Sale price RS ")
		}

		product.OrigPrice = cleanPrice(inner.Find("div.rupee del.regular-price").First().Text())
		if product.OrigPrice == "" {
			product.OrigPrice = extractPriceFromAriaLabel(inner.Find("div.rupee").First().AttrOr("aria-label", ""), "Regular price RS ")
		}

		imgSel := inner.Find("div.list_img img").First()
		product.ImageURL, _ = imgSel.Attr("src")
		if product.ImageURL == "" {
			product.ImageURL, _ = imgSel.Attr("data-src")
		}
		product.ImageURL = normalizeURL(product.ImageURL)

		if product.Name != "" || product.Price != "" {
			products = append(products, product)
		}
	})

	log.Printf("[Scraper] Parsed %d products from page", len(products))
	return products
}

func extractProductID(sel *goquery.Selection) string {
	if pid, ok := sel.Find("[data-pid]").First().Attr("data-pid"); ok && pid != "" {
		return pid
	}
	if pid, ok := sel.Attr("data-pid"); ok && pid != "" {
		return pid
	}

	class, _ := sel.Attr("class")
	for _, part := range strings.Fields(class) {
		if strings.HasPrefix(part, "listingpg-") {
			return strings.TrimPrefix(part, "listingpg-")
		}
	}

	href, _ := sel.Find("a[href*='product-detail']").First().Attr("href")
	return extractIDFromURL(href)
}

func extractIDFromURL(href string) string {
	matches := productIDFromURL.FindStringSubmatch(href)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func normalizeURL(href string) string {
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if strings.HasPrefix(href, "/") {
		return "https://www.firstcry.com" + href
	}
	if strings.HasPrefix(href, "http") {
		return href
	}
	return "https://www.firstcry.com/" + strings.TrimPrefix(href, "/")
}

func cleanPrice(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, raw)
	return raw
}

func extractPriceFromAriaLabel(label, prefix string) string {
	idx := strings.Index(label, prefix)
	if idx < 0 {
		return ""
	}
	rest := label[idx+len(prefix):]
	if end := strings.IndexAny(rest, " and"); end >= 0 {
		rest = rest[:end]
	}
	return strings.TrimSpace(rest)
}
