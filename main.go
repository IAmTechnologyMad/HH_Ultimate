package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kents/firstcry-hotwheels-bot/config"
	"github.com/kents/firstcry-hotwheels-bot/notifier"
	"github.com/kents/firstcry-hotwheels-bot/scraper"
	"github.com/kents/firstcry-hotwheels-bot/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting FirstCry Hot Wheels Monitor Bot...")

	cfg := config.Load()

	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open store: %v", err)
	}
	defer db.Close()

	tg, err := notifier.New(cfg.TelegramBotToken, cfg.TelegramChatID)
	if err != nil {
		log.Fatalf("Failed to init Telegram: %v", err)
	}

	sc := scraper.New(
		cfg.FirstCryURL,
		cfg.RequestTimeout,
		cfg.MaxRetries,
		cfg.RetryDelay,
	)

	go startHealthServer(cfg.Port)

	count, _ := db.CountProducts()
	if err := tg.SendStartup(count); err != nil {
		log.Printf("Warning: Could not send startup message: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startPinger(ctx, cfg.PingURL)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	log.Printf("Polling every %s | URL: %s", cfg.PollInterval, cfg.FirstCryURL)
	runMonitorLoop(ctx, cfg, sc, db, tg)

	log.Println("Bot stopped. Goodbye!")
}

func runMonitorLoop(
	ctx context.Context,
	cfg *config.Config,
	sc *scraper.Scraper,
	db *store.Store,
	tg *notifier.TelegramNotifier,
) {
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	checkProducts(ctx, sc, db, tg)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkProducts(ctx, sc, db, tg)
		}
	}
}

func checkProducts(
	ctx context.Context,
	sc *scraper.Scraper,
	db *store.Store,
	tg *notifier.TelegramNotifier,
) {
	log.Printf("[Monitor] Checking for new products...")

	products, err := sc.FetchProducts(ctx)
	if err != nil {
		log.Printf("[Monitor] ERROR fetching products: %v", err)
		return
	}

	log.Printf("[Monitor] Found %d products on page", len(products))

	// Collect products to save and alerts to send
	var toSave []store.ProductInput
	var alerts []string

	for _, p := range products {
		status, err := db.GetProductStatus(p.ID)
		if err != nil {
			log.Printf("[Monitor] DB error checking product %s: %v", p.ID, err)
			continue
		}

		shouldAlert := false

		if status.IsNew {
			p.IsNew = true
			shouldAlert = true
			log.Printf("[Monitor] NEW PRODUCT: %s (%s)", p.Name, p.ID)
		} else if status.WasOutOfStock && p.InStock {
			p.IsRestock = true
			shouldAlert = true
			log.Printf("[Monitor] RESTOCK: %s (%s)", p.Name, p.ID)
		}

		toSave = append(toSave, store.ProductInput{
			ID:      p.ID,
			Name:    p.Name,
			InStock: p.InStock,
		})

		if shouldAlert && p.InStock {
			alerts = append(alerts, p.TelegramMessage())
		}
	}

	// Batch save all products in a single write transaction
	if len(toSave) > 0 {
		if err := db.BatchSaveProducts(toSave); err != nil {
			log.Printf("[Monitor] Failed to batch save products: %v", err)
		}
	}

	// Send alerts after saving
	for _, msg := range alerts {
		if err := tg.SendAlert(msg); err != nil {
			log.Printf("[Monitor] Failed to send alert: %v", err)
		}
	}
}

func startHealthServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"firstcry-hotwheels-bot"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("FirstCry Hot Wheels Monitor is running"))
	})

	log.Printf("[HTTP] Health-check server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Printf("[HTTP] Server error: %v", err)
	}
}

func startPinger(ctx context.Context, url string) {
	if url == "" {
		log.Println("[Pinger] No RENDER_EXTERNAL_URL provided, self-pinger disabled")
		return
	}

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	pingUrl := url + "/health"
	log.Printf("[Pinger] Started self-pinger for %s every 10m", pingUrl)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := http.Get(pingUrl)
			if err != nil {
				log.Printf("[Pinger] Failed to ping self: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("[Pinger] Self-ping successful (status: %s)", resp.Status)
			}
		}
	}
}
