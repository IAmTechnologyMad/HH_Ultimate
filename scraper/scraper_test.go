package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

const sampleListingHTML = `
<div class="list_block lft" data-outstock="true">
	<div class="li_inner_block listingpg-20257824">
		<div class="list_img wifi">
			<a href="//www.firstcry.com/hot-wheels/example/20257824/product-detail">
				<img src="//cdn.fcglcdn.com/brainbees/images/products/219x265/20257824a.jpg" alt="Hot Wheels Die Cast Pack">
			</a>
		</div>
		<div class="li_txt1 wifi lft">
			<a href="//www.firstcry.com/hot-wheels/example/20257824/product-detail"
			   title="Hot Wheels Die Cast Free Wheel Vehicle Toys in 1:64 Scale Pack of 2 - Multicolor">
				Hot Wheels Die Cast Free Wheel Vehicle Toys in 1:64...
			</a>
		</div>
		<div class="rupee fw lft" aria-label="Sale price RS 262.07 and Regular price RS 359">
			<span class="r1 B14_42"><a><span></span>262.07</a></span>
			<span class="r2 R12_42"><a><del class="regular-price"><span></span>359</del></a></span>
		</div>
	</div>
</div>
<div class="list_block lft" data-outstock="false">
	<div class="li_inner_block listingpg-14933549">
		<div class="list_img wifi">
			<a href="//www.firstcry.com/hot-wheels/example/14933549/product-detail">
				<img src="//cdn.fcglcdn.com/brainbees/images/products/219x265/14933549a.jpg">
			</a>
		</div>
		<div class="li_txt1 wifi lft">
			<a href="//www.firstcry.com/hot-wheels/example/14933549/product-detail"
			   title="Hot Wheels Color Shifters Track and 1 Car - Multicolour">
				Hot Wheels Color Shifters Track...
			</a>
		</div>
		<div class="rupee fw lft" aria-label="Sale price RS 499 and Regular price RS 699">
			<span class="r1 B14_42"><a><span></span>499</a></span>
			<span class="r2 R12_42"><a><del class="regular-price"><span></span>699</del></a></span>
		</div>
	</div>
</div>
`

func TestParseProducts(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(sampleListingHTML))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	s := New("https://example.com", 0, 1, 0)
	products := s.parseProducts(doc)

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	first := products[0]
	if first.ID != "20257824" {
		t.Errorf("expected ID 20257824, got %q", first.ID)
	}
	if first.InStock {
		t.Error("expected first product to be out of stock")
	}
	if first.Price != "262.07" {
		t.Errorf("expected price 262.07, got %q", first.Price)
	}
	if first.OrigPrice != "359" {
		t.Errorf("expected MRP 359, got %q", first.OrigPrice)
	}
	if !strings.Contains(first.Name, "Hot Wheels Die Cast") {
		t.Errorf("unexpected name: %q", first.Name)
	}

	second := products[1]
	if second.ID != "14933549" {
		t.Errorf("expected ID 14933549, got %q", second.ID)
	}
	if !second.InStock {
		t.Error("expected second product to be in stock")
	}
}
