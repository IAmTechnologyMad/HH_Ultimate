package scraper

import (
	"math/rand"
	"net/http"
	"sync/atomic"
)

// HeaderPool rotates User-Agent and other headers to avoid bot detection.
type HeaderPool struct {
	agents []string
	idx    uint64
}

func NewHeaderPool() *HeaderPool {
	return &HeaderPool{
		agents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
		},
	}
}

// Apply sets rotating browser-like headers on the request.
func (h *HeaderPool) Apply(req *http.Request) {
	idx := atomic.AddUint64(&h.idx, 1) % uint64(len(h.agents))
	agent := h.agents[idx]

	req.Header.Set("User-Agent", agent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-IN,en;q=0.9,hi;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", "https://www.google.com/")

	viewports := []string{"1920,1080", "1440,900", "1366,768", "2560,1440"}
	req.Header.Set("Sec-CH-Viewport-Width", viewports[rand.Intn(len(viewports))])
}
