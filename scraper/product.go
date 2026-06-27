package scraper

import (
	"fmt"
	"html"
	"strings"
)

// Product represents a single Hot Wheels listing on FirstCry.
type Product struct {
	ID        string
	Name      string
	Price     string
	OrigPrice string
	URL       string
	ImageURL  string
	InStock   bool
	IsNew     bool
	IsRestock bool
}

// Fingerprint returns a unique key for this product.
func (p *Product) Fingerprint() string {
	return fmt.Sprintf("%s|%v", p.ID, p.InStock)
}

// TelegramMessage builds the Telegram alert message with HTML formatting.
func (p *Product) TelegramMessage() string {
	label := "🆕 NEW PRODUCT"
	if p.IsRestock {
		label = "🔄 RESTOCK ALERT"
	}

	inStockStr := "✅ IN STOCK – ORDER NOW!"
	if !p.InStock {
		inStockStr = "⚠️ Out of Stock"
	}

	price := p.Price
	if price != "" && !strings.HasPrefix(price, "₹") {
		price = "₹" + price
	}
	origPrice := p.OrigPrice
	if origPrice != "" && !strings.HasPrefix(origPrice, "₹") {
		origPrice = "₹" + origPrice
	}

	return fmt.Sprintf(
		"<b>%s</b> 🚨\n\n"+
			"🏎️ <b>%s</b>\n\n"+
			"💰 Price: <b>%s</b> (MRP: %s)\n"+
			"📦 Status: %s\n\n"+
			"🔗 <a href=\"%s\">Buy Now on FirstCry</a>\n\n"+
			"⚡ Act fast before it sells out!",
		label,
		html.EscapeString(p.Name),
		html.EscapeString(price),
		html.EscapeString(origPrice),
		inStockStr,
		html.EscapeString(p.URL),
	)
}
