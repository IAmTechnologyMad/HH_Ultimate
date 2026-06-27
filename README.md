# FirstCry Hot Wheels Telegram Monitor Bot

Instantly notified on Telegram when a new Hot Wheels product appears on FirstCry.

## How to Get Your Telegram Bot and Chat ID

### Step 1 – Create a Telegram Bot

1. Open Telegram, search for `@BotFather`
2. Send `/newbot`
3. Follow prompts, choose a name and username
4. BotFather gives you a **Bot Token** — copy it

### Step 2 – Get Your Chat ID

1. Open your bot in Telegram and press Start
2. Visit this URL in your browser (replace `YOUR_TOKEN`):
   `https://api.telegram.org/botYOUR_TOKEN/getUpdates`
3. Send any message to your bot
4. Refresh the URL above — find `"chat":{"id": XXXXXX}` — that's your Chat ID

### Step 3 – Deploy to Render.com

1. Push this repo to GitHub
2. Go to [render.com](https://render.com) → New → Web Service
3. Connect your GitHub repo
4. Select "Use render.yaml"
5. Under **Environment**, add:
   - `TELEGRAM_BOT_TOKEN` = your bot token
   - `TELEGRAM_CHAT_ID` = your chat ID
6. Click **Deploy**

## Adjusting Poll Speed

Edit `POLL_INTERVAL_SECONDS` in Render environment:

- `30` = check every 30 seconds (recommended)
- `15` = more aggressive, may get blocked occasionally
- `60` = very safe

## CSS Selectors

The scraper targets FirstCry's listing page structure:

- Product cards: `div.list_block`
- Stock status: `data-outstock="true"` (out of stock) / `"false"` (in stock)
- Product ID: `data-pid`, `listingpg-{id}` class, or URL path
- Name: `div.li_txt1 a[title]`
- Price: `div.rupee span.r1`
- MRP: `div.rupee del.regular-price`

If FirstCry changes their HTML, update `scraper/scraper.go` → `parseProducts()`.

## Notification Types

- **NEW PRODUCT** — A product ID never seen before appears on the listing page (in stock only)
- **RESTOCK** — A previously out-of-stock product is now available

## Testing Locally

```bash
export TELEGRAM_BOT_TOKEN="your_token"
export TELEGRAM_CHAT_ID="your_chat_id"
export DB_PATH="./data/products.db"
go run main.go
```

On Windows PowerShell:

```powershell
$env:TELEGRAM_BOT_TOKEN="your_token"
$env:TELEGRAM_CHAT_ID="your_chat_id"
$env:DB_PATH="./data/products.db"
go run main.go
```

## Docker

```bash
docker build -t hotwheels-bot .
docker run -e TELEGRAM_BOT_TOKEN=xxx -e TELEGRAM_CHAT_ID=yyy -e DB_PATH=/var/data/products.db -p 8080:8080 hotwheels-bot
```

## Optional: Prevent Render Free Tier Sleep

Set up a free [UptimeRobot](https://uptimerobot.com) ping to `/health` every 5 minutes.
