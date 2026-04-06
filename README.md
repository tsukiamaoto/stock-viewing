# 全球股市監控中心 (Global Stock Market Monitor)

一個以 React + Vite 建構的現代化全球股市與金融數據監控平台，搭配 **Go (Gin)** 後端，提供多源新聞聚合、社群論壇追蹤、AI 輔助分類與即時爬蟲監控功能。

## 🛠 技術棧

| 類別 | 技術 |
|------|------|
| **前端** | React 19, TypeScript, Vite, React Router DOM v7 |
| **UI** | Material UI, Emotion, 原生 CSS, Lucide React |
| **後端** | Go (Gin) + robfig/cron 背景排程 |
| **資料庫** | Supabase (PostgreSQL 雲端) |
| **AI** | Google Gemini API (多模型輪換) — 新聞分類與摘要 |
| **爬蟲** | CNN, Reuters RSS, NHK RSS, 金十數據, PTT 股版, CMoney 同學會 (chromedp), Yahoo Finance v8, TWSE |
| **翻譯** | Google Translate API (en/ja/zh-CN → zh-TW) |

## 🏛 系統架構

```text
┌─────────────────────────────────────────────────────────────┐
│                    React Frontend (Vite)                     │
│  App.tsx (Router + Global State + localStorage)             │
│  ├── DashboardPage     全球大盤、期指、自選股、恐懼貪婪指數     │
│  ├── NewsPage          CNN / Reuters / NHK / 金十 多欄新聞    │
│  ├── ForumPage         PTT 股版 + CMoney 同學會 (手風琴分類)   │
│  ├── WatchlistPage     自選股管理                             │
│  ├── StockDetailPage   個股詳情與股權分散表                    │
│  └── CrawlerDashboard  爬蟲監控日誌 (即時 SSE + 層級過濾)      │
└───────────────────────────┬─────────────────────────────────┘
                            │ REST API + SSE (localhost:8000)
┌───────────────────────────▼─────────────────────────────────┐
│                   Go (Gin) Backend                          │
│  cmd/server/main.go — HTTP Server + cron 排程器              │
│                                                             │
│  Handler    handler/news.go, quotes.go, stock_detail.go,    │
│             shareholders.go                                 │
│  Service    service/news_service.go, quote_service.go,      │
│             stock_service.go                                │
│  Crawler    crawler/cnn, rss, jin10, ptt_stock,             │
│             cmoney_forum (chromedp), twse_etf, yahoo        │
│  Infra      database/supabase, llm/gemini, translate/,     │
│             logger/ (SSE), config/, middleware/cors          │
└───────────────────────────┬─────────────────────────────────┘
                            │
           ┌────────────────▼────────────────┐
           │       Supabase (PostgreSQL)     │
           │  Table: news (翻譯標題/摘要/來源) │
           └─────────────────────────────────┘
```

## 📁 專案目錄結構

```text
stock-viewing/
├── src/                                    # React 前端
│   ├── components/
│   │   ├── shared/
│   │   │   └── PriceChangeDisplay.tsx      # 共用漲跌顯示組件
│   │   ├── AIAssistantWidget.tsx           # AI 新聞洞察小組件
│   │   ├── ChartWidget.tsx                 # TradingView 走勢圖
│   │   ├── CrawlerDashboard.tsx            # 爬蟲監控日誌 + 層級過濾
│   │   ├── CrawlerDashboard.css            # 爬蟲監控樣式
│   │   ├── CustomWatchlist.tsx             # 首頁自選股模塊
│   │   ├── Dashboard.tsx                   # 國際大盤模塊
│   │   ├── FearGreedIndex.tsx              # 恐懼與貪婪指數
│   │   ├── ForumPage.tsx                   # PTT + CMoney 論壇 (手風琴分類)
│   │   ├── IndexInfoWidget.tsx             # 指數資訊組件
│   │   ├── LayoutSettings.tsx              # 儀表板排版設定
│   │   ├── NewsPage.tsx                    # 多源新聞資訊頁面
│   │   ├── StockDetailPage.tsx             # 個股詳情與股權結構
│   │   ├── SymbolInfoWidget.tsx            # 基礎報價資訊組件
│   │   ├── TaiwanFuturesWidget.tsx         # 台灣市場模塊
│   │   └── WatchlistPage.tsx               # 自選股管理頁面
│   ├── hooks/
│   │   ├── usePolling.ts                   # 計時器輪詢 Hook
│   │   └── useTradingViewWidget.ts         # TradingView 嵌入 Hook
│   ├── App.tsx                             # 應用進入點 (路由/全局狀態)
│   ├── main.tsx                            # React 掛載點
│   └── index.css                           # 全局 CSS 變數與核心樣式
│
├── backend-go/                             # Go 後端
│   ├── cmd/
│   │   ├── server/main.go                  # 入口：HTTP Server + 排程器
│   │   ├── clear_db/main.go                # 工具：清空 Supabase news 表
│   │   └── trigger/main.go                 # 工具：手動觸發一次完整爬取
│   ├── internal/
│   │   ├── config/config.go                # 環境變數 + LLM Prompts
│   │   ├── database/supabase.go            # Supabase PostgREST Client
│   │   ├── handler/                        # HTTP 路由層
│   │   │   ├── news.go                     # /api/news/* (含 ptt, cmoney)
│   │   │   ├── quotes.go                   # /api/stocks/index, /watchlist
│   │   │   ├── stock_detail.go             # /api/stocks/detail/:code
│   │   │   └── shareholders.go             # /api/stocks/shareholders/:code
│   │   ├── service/                        # 業務邏輯層
│   │   │   ├── news_service.go             # 爬蟲→翻譯→DB 管線
│   │   │   ├── quote_service.go            # Yahoo Finance 報價聚合
│   │   │   └── stock_service.go            # 個股+股權分散表
│   │   ├── crawler/                        # 爬蟲模組 (7 個資料源)
│   │   │   ├── common.go                   # 通用 HTTP Client + User-Agent
│   │   │   ├── cnn.go                      # CNN 頁面爬取
│   │   │   ├── rss.go                      # Reuters / NHK RSS 聚合
│   │   │   ├── jin10.go                    # 金十數據 (API→RSSHub→直接爬取)
│   │   │   ├── ptt_stock.go                # PTT 股版爬取
│   │   │   ├── cmoney_forum.go             # CMoney 同學會 (chromedp 無頭瀏覽器)
│   │   │   ├── twse_etf.go                 # TWSE ETF 公告
│   │   │   └── yahoo.go                    # Yahoo Finance v8 API
│   │   ├── llm/gemini.go                   # Gemini 多模型輪換分類器
│   │   ├── translate/                      # Google Translate API
│   │   ├── logger/                         # 結構化日誌 + SSE 即時串流
│   │   ├── model/                          # 通用資料模型
│   │   │   ├── response.go                 # APIResponse 通用回應
│   │   │   ├── news.go                     # 新聞結構體
│   │   │   ├── quote.go                    # 報價結構體
│   │   │   └── stock.go                    # 個股+股權分散表
│   │   └── middleware/cors.go              # CORS 中介層
│   ├── vendor/                             # 本地化依賴 (go mod vendor)
│   ├── go.mod / go.sum                     # Go 模組定義與校驗
│   ├── .air.toml                           # Air 熱更新設定
│   └── .env.example                        # 後端環境變數範本
│
├── .env.example                            # 前端環境變數範本
├── vite.config.ts                          # Vite 設定 (API Proxy)
├── package.json                            # 前端依賴
└── .gitignore
```

## 核心模組說明

| 模組 | 位置 | 用途 |
|---|---|---|
| **News Handler** | `handler/news.go` | 多源新聞 API + PTT / CMoney 論壇 API |
| **Quotes Handler** | `handler/quotes.go` | 自選股與指數報價 (Yahoo Finance) |
| **Stock Detail** | `handler/stock_detail.go` | 個股基本面、財報、大戶持股比例 |
| **Shareholders** | `handler/shareholders.go` | 集保結算所每週股權分散表爬取 |
| **News Service** | `service/news_service.go` | 協調 crawler → translate → DB 管線 |
| **CMoney Crawler** | `crawler/cmoney_forum.go` | 使用 chromedp 無頭瀏覽器爬取 SPA 論壇 |
| **PTT Crawler** | `crawler/ptt_stock.go` | PTT 股版 HTML 爬取 |
| **Jin10 Crawler** | `crawler/jin10.go` | 三層 fallback (API → RSSHub → 直接爬取) |
| **Logger** | `logger/` | 結構化日誌 + SSE 即時推送至前端 |
| **Gemini LLM** | `llm/gemini.go` | 多模型輪換、額度追蹤、JSON 擷取 |

## ⚙️ 環境設定

本專案使用 `.env` 存放可配置的參數，請勿將 `.env` 提交至版本控制。

### 前端 (根目錄)

```bash
cp .env.example .env
```

| 變數 | 說明 | 預設值 |
|---|---|---|
| `VITE_API_URL` | 後端服務的 URL | `http://localhost:8000` |
| `VITE_TWSE_OPEN_API_URL` | TWSE 收盤資料 API URL | `https://openapi.twse.com.tw/...` |

### 後端 (`backend-go/`)

```bash
cp backend-go/.env.example backend-go/.env
```

| 變數 | 說明 | 範例值 |
|---|---|---|
| `GEMINI_API_KEY` | **必填** Google Gemini API 金鑰 | `AIza...` |
| `GEMINI_MODELS` | 可用的 Gemini 模型清單 (逗號分隔) | `gemma-3-4b-it,gemini-2.5-flash` |
| `FRONTEND_URLS` | CORS 前端 URL (逗號分隔) | `http://localhost:5173` |
| `API_HOST` / `API_PORT` | 後端監聽位址 | `0.0.0.0` / `8000` |
| `VITE_SUPABASE_URL` | Supabase 專案 URL | `https://xxxx.supabase.co` |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase Server 私鑰 | `eyJhbG...` |
| `CRAWLER_INTERVAL_MINUTES` | 定時爬蟲間隔 (分鐘) | `5` |

> [!IMPORTANT]
> `GEMINI_API_KEY` 為 AI 新聞分類功能的必要條件。未填寫時系統自動退回至 keyword-based fallback。

## 🚀 開發與啟動

### 1. 啟動 Go 後端

```bash
cd backend-go

# 複製環境變數並填入金鑰
cp .env.example .env

# 開發模式 (熱更新)
./air.exe

# 或直接運行
go run -mod=vendor ./cmd/server
```

> [!TIP]
> Go 後端使用 `go mod vendor` 管理依賴，無需全域安裝套件。編譯時請加上 `-mod=vendor`。

### 2. 啟動前端

```bash
# 安裝依賴
npm install

# 複製環境變數
cp .env.example .env

# 啟動 Vite 開發伺服器
npm run dev
```

### 3. 實用工具指令

```bash
# 手動觸發一次完整爬取 (不需啟動 server)
go run -mod=vendor ./cmd/trigger

# 清空 Supabase news 資料表
go run -mod=vendor ./cmd/clear_db
```

### 4. 強制關閉背景後端

**Windows:**
```bash
taskkill -F -IM server.exe -T
# 或
taskkill -F -IM go.exe -T
```

**Mac / Linux:**
```bash
pkill -f server
```

## 📡 API 端點一覽

### 新聞與論壇
| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/news/latest` | 最新新聞 (所有來源) |
| GET | `/api/news/cnn` | CNN 新聞 |
| GET | `/api/news/reuters` | Reuters 新聞 |
| GET | `/api/news/nhk` | NHK 新聞 |
| GET | `/api/news/jin10` | 金十數據快訊 |
| GET | `/api/news/twse-etf` | TWSE ETF 公告 |
| GET | `/api/news/ptt` | PTT 股版論壇 |
| GET | `/api/news/cmoney?symbols=2330,0050` | CMoney 同學會 (依自選股分類) |
| GET | `/api/news/categorize/:symbol` | LLM 新聞分類分析 |

### 報價與個股
| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/stocks/index` | 全球大盤指數 |
| GET | `/api/stocks/watchlist?symbols=...` | 自選股報價 |
| GET | `/api/stocks/detail/:code` | 個股詳情 (Yahoo Finance) |
| GET | `/api/stocks/shareholders/:code` | 股權分散表 (集保結算所) |

### 系統
| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/crawler/logs` | SSE 即時爬蟲日誌串流 |
