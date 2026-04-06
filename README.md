# 全球股市監控中心 (Global Stock Market Monitor)

這是一個以 React + Vite 建構的現代化全球股市與金融數據監控前端應用程式，搭配 **Go (Gin)** 後端，提供 AI 輔助新聞分類與 RSS 聚合功能。

## 🛠 技術棧 (Tech Stack)

### 核心框架與工具
*   **前端:** React 19, TypeScript, Vite, React Router DOM (v7)
*   **UI 系統:** Material UI (MUI), Emotion, 原生 CSS, Lucide React
*   **後端:** Go (Gin HTTP Framework) + robfig/cron 背景排程
*   **資料庫與服務:** Supabase (PostgreSQL 雲端資料庫)
*   **AI 整合:** Google Gemini API (多模型輪換), 用於新聞分類與摘要
*   **爬蟲資料源:** RSS (gofeed), Yahoo Finance v8 API, TWSE, 集保結算所

## 🏛 系統架構 (System Architecture)

```text
┌───────────────────────────────────────────────────────┐
│                    React Frontend                     │
│  App.tsx (Router + Global State + localStorage)       │
│  ├── DashboardPage (儀表板)                           │
│  │   ├── TaiwanFuturesWidget                          │
│  │   ├── CustomWatchlist                              │
│  │   ├── FearGreedIndex                               │
│  │   └── Dashboard                                    │
│  ├── WatchlistPage  (自選股管理)                        │
│  ├── NewsPage       (各平台新聞)                        │
│  └── StockDetailPage(個股詳細分析與股權分散表)             │
│                                                       │
│  Shared Config / Hooks:                               │
│  ├── hooks/usePolling.ts                              │
│  ├── hooks/useTradingViewWidget.ts                    │
│  └── components/shared/PriceChangeDisplay.tsx         │
└────────────────────────┬──────────────────────────────┘
                         │ REST API (localhost:8000)
┌────────────────────────▼──────────────────────────────┐
│                  Go (Gin) Backend                     │
│  cmd/server/main.go (入口 + cron 排程器)               │
│                                                       │
│  Handler Layer (路由):                                 │
│  ├── handler/news.go         (/api/news/*)            │
│  │   ├── GET /latest, /cnn, /reuters, /nhk, /jin10    │
│  │   └── GET /categorize/:symbol (LLM Analysis)       │
│  ├── handler/quotes.go       (/api/stocks/index,      │
│  │                            /api/stocks/watchlist)   │
│  ├── handler/stock_detail.go (/api/stocks/detail/:code)│
│  └── handler/shareholders.go (/api/stocks/shareholders)│
│                                                       │
│  Service Layer (業務邏輯):                              │
│  ├── service/news_service.go     (爬蟲→LLM→DB 管線)    │
│  ├── service/quote_service.go    (報價聚合)             │
│  └── service/stock_service.go    (個股+股權分散表)       │
│                                                       │
│  Infrastructure:                                      │
│  ├── crawler/ (cnn, rss, jin10, yahoo)                │
│  ├── llm/gemini.go           (Gemini 多模型輪換)       │
│  ├── database/supabase.go    (PostgREST Client)       │
│  └── config/config.go        (環境變數管理)             │
└────────────────────────┬──────────────────────────────┘
                         │ Read / Write
┌────────────────────────▼──────────────────────────────┐
│                 Supabase (Database)                   │
│  Table: news (自動翻譯的標題, 新聞摘要片段, 來源等)        │
└───────────────────────────────────────────────────────┘
```

### 核心模組說明
| 模組 | 位置 | 用途 |
|---|---|---|
| **News Handler** | `backend-go/internal/handler/news.go` | 從 Supabase 讀取多源新聞，呼叫 LLM 分類 |
| **Quotes Handler** | `backend-go/internal/handler/quotes.go` | 自選股與指數報價 (Yahoo Finance API) |
| **Stock Detail Handler** | `backend-go/internal/handler/stock_detail.go` | 個股基本面、財報、大戶持股比例 |
| **Shareholders Handler** | `backend-go/internal/handler/shareholders.go` | 爬取集保結算所的每週股權分散表 |
| **News Service** | `backend-go/internal/service/news_service.go` | 協調 crawler → LLM → database 管線 |
| **Yahoo Crawler** | `backend-go/internal/crawler/yahoo.go` | 直接呼叫 Yahoo Finance v8 REST API |
| **Gemini LLM** | `backend-go/internal/llm/gemini.go` | 多模型輪換、額度追蹤、JSON 擷取 |

## ⚙️ 環境設定 (Environment Setup)

本專案使用 `.env` 存放可配置的參數，請勿將 `.env` 提交至版本控制。

### 前端 (根目錄)

複製範本並填入對應值：
```bash
cp .env.example .env
```

| 變數 | 說明 | 預設值 |
|---|---|---|
| `VITE_API_URL` | 後端服務的 URL | `http://localhost:8000` |
| `VITE_TWSE_OPEN_API_URL` | TWSE 收盤資料 API URL | `https://openapi.twse.com.tw/v1/exchangeReport/STOCK_DAY_ALL` |

### 後端 (`backend-go/` 目錄)

```bash
cp backend-go/.env.example backend-go/.env
```

| 變數 | 說明 | 範例值 |
|---|---|---|
| `GEMINI_API_KEY` | **必填**，Google Gemini API 金鑰 | `AIza...` |
| `GEMINI_MODELS` | 可用的 Gemini 模型清單 (逗號分隔) | `gemma-3-4b-it,gemini-2.5-flash` |
| `FRONTEND_URLS` | 允許 CORS 的前端 URL，逗號分隔 | `http://localhost:5173,http://localhost:5174` |
| `API_HOST` | 後端監聽的 Host | `0.0.0.0` |
| `API_PORT` | 後端監聽的 Port | `8000` |
| `VITE_SUPABASE_URL` | Supabase 專案 URL | `https://xxxx.supabase.co` |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase Server 端用私鑰 (具寫入權限) | `eyJhbG...` |
| `CRAWLER_INTERVAL_MINUTES` | 定時爬蟲間隔 (分鐘) | `5` |

> [!IMPORTANT]
> `GEMINI_API_KEY` 為啟動 AI 新聞分類功能的必要條件。未填寫時系統會自動退回至 keyword-based fallback 機制。

## 🚀 開發與啟動方式 (Development & Getting Started)

此版本依賴 Supabase 與定時任務，強烈建議**同時啟動後端與前端**以取得完整功能。

### 1. 啟動 Go 後端

```bash
cd backend-go

# 複製環境變數範本並填入必要金鑰
cp .env.example .env

# 方法 A：開發模式 (熱更新 Hot Reload)
# 直接運行我們包含的 air.exe，只要您儲存 Go 檔案，伺服器就會自動重新編譯並重啟
./air.exe

# 方法 B：生產模式 (直接編譯並啟動)
go build -mod=vendor -o server.exe ./cmd/server/
./server.exe
```

> [!TIP]
> Go 後端使用 `go mod vendor` 將所有套件下載至 `backend-go/vendor/` 目錄，無需全域安裝任何依賴。編譯時請加上 `-mod=vendor` 旗標。

*(後端啟動後會預設運行於 http://localhost:8000，並自動開啟背景排程與 API 服務)*

### 2. 啟動前端 UI (React + Vite)

請開啟**另一個**新的終端機視窗：

```bash
# 安裝前端依賴
npm install

# 複製環境變數範本
cp .env.example .env

# 啟動 Vite 開發伺服器
npm run dev
```

### 3. 建置打包 (Production Build)

```bash
npm run build
```

### 4. 強制關閉背景後端 (Kill Backend Process)

如果在開發過程中遇到啟動埠號衝突或後端卡滯的情況：

**Windows:**
```bash
taskkill -F -IM server.exe -T
```

**Mac / Linux:**
```bash
pkill -f server
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
│   │   ├── CustomWatchlist.tsx             # 首頁自選股模塊
│   │   ├── Dashboard.tsx                   # 國際大盤模塊
│   │   ├── FearGreedIndex.tsx              # 恐懼與貪婪指數
│   │   ├── LayoutSettings.tsx              # 儀表板排版設定
│   │   ├── NewsPage.tsx                    # 新聞資訊頁面
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
├── backend-go/                             # Go 後端 (主要)
│   ├── cmd/server/
│   │   └── main.go                         # 入口：HTTP Server + 排程器
│   ├── internal/
│   │   ├── config/config.go                # 環境變數 + LLM Prompts
│   │   ├── database/supabase.go            # Supabase PostgREST Client
│   │   ├── handler/                        # HTTP 路由層
│   │   │   ├── news.go                     # /api/news/*
│   │   │   ├── quotes.go                   # /api/stocks/index, /watchlist
│   │   │   ├── stock_detail.go             # /api/stocks/detail/:code
│   │   │   └── shareholders.go             # /api/stocks/shareholders/:code
│   │   ├── service/                        # 業務邏輯層
│   │   │   ├── news_service.go             # 爬蟲→LLM→DB 管線
│   │   │   ├── quote_service.go            # Yahoo Finance 報價
│   │   │   └── stock_service.go            # 個股+股權分散表
│   │   ├── crawler/                        # 爬蟲模組
│   │   │   ├── common.go                   # 通用 HTTP Client
│   │   │   ├── cnn.go                      # CNN 頁面爬取
│   │   │   ├── rss.go                      # Reuters / NHK RSS
│   │   │   ├── jin10.go                    # 金十數據快訊
│   │   │   └── yahoo.go                    # Yahoo Finance v8 API
│   │   ├── llm/gemini.go                   # Gemini 多模型輪換分類器
│   │   ├── model/                          # 通用資料模型
│   │   │   ├── response.go                 # APIResponse 通用回應
│   │   │   ├── news.go                     # 新聞結構體
│   │   │   ├── quote.go                    # 報價結構體
│   │   │   └── stock.go                    # 個股+股權分散表
│   │   └── middleware/cors.go              # CORS 中介層
│   ├── vendor/                             # 本地化依賴 (go mod vendor)
│   ├── go.mod                              # Go 模組定義
│   ├── go.sum                              # 依賴校驗
│   ├── .env                                # 後端環境變數 (gitignored)
│   └── .env.example                        # 後端環境變數範本
├── .env                                    # 前端環境變數 (gitignored)
├── .env.example                            # 前端環境變數範本
├── vite.config.ts                          # Vite 設定 (API Proxy)
└── package.json
```
