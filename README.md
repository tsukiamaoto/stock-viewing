# 全球股市監控中心 (Global Stock Market Monitor)

這是一個以 React + Vite 建構的現代化全球股市與金融數據監控前端應用程式，搭配 Python FastAPI 作為後端，提供 AI 輔助新聞分類與 RSS 聚合功能。

## 🛠 技術棧 (Tech Stack)

### 核心框架與工具
*   **前端:** React 19, TypeScript, Vite, React Router DOM (v7)
*   **UI 系統:** Material UI (MUI), Emotion, 原生 CSS, Lucide React
*   **後端:** Python FastAPI + Uvicorn
*   **背景任務:** APScheduler (背景定時爬蟲)
*   **資料庫與服務:** Supabase (PostgreSQL 雲端資料庫)
*   **AI 整合:** Google Gemini API (`gemini-1.5-flash`), 用於新聞分類與摘要
*   **爬蟲資料源:** RSS (`feedparser`), yfinance, Google API (網頁翻譯器), TSWE, 集保結算所

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
│                   FastAPI Backend                     │
│  main.py (主入口與 apscheduler 定時背景爬取任務)            │
│  ├── news_api.py           (/api/news/*)              │
│  │   ├── GET /latest, /cnn, /reuters, /nhk            │
│  │   └── GET /categorize/{symbol} (LLM Analysis)      │
│  │                                                    │
│  ├── quotes_api.py         (/api/stocks/watchlist)    │
│  ├── stock_detail_api.py   (/api/stocks/detail)       │
│  └── shareholders_api.py   (/api/stocks/shareholders) │
│                                                       │
│  Helper Modules:                                      │
│  ├── news_crawler.py       (Web & RSS Crawling)       │
│  ├── llm_classifier.py     (Gemini Prompting)         │
│  ├── database.py           (Supabase Connection)      │
│  └── config.py             (Environment Configuration)│
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
| **News API** | `backend/news_api.py` | 負責從 Supabase 讀取多個新聞源資料，也呼叫 LLM 進行新聞分類。 |
| **Quotes API** | `backend/quotes_api.py` | 負責自選股與報價相關。 |
| **Stock Detail API** | `backend/stock_detail_api.py` | 抓取 yfinance 提供的個股基本面、財報、與大戶持股比例。 |
| **Shareholders API** | `backend/shareholders_api.py` | 爬取集保結算所的每週股權分散表 (持有張數分布)。 |

## ⚙️ 環境設定 (Environment Setup)

本專案使用 `.env` 存放可配置的參數，請勿將 `.env` 提交至版本控制。

### 前端 (根目錄)

複製範本並填入對應值：
```bash
cp .env.example .env
```

| 變數 | 說明 | 預設值 |
|---|---|---|
| `VITE_API_URL` | FastAPI 後端服務的 URL | `http://localhost:8000` |
| `VITE_TWSE_OPEN_API_URL` | TWSE 收盤資料 API URL | `https://openapi.twse.com.tw/v1/exchangeReport/STOCK_DAY_ALL` |

### 後端 (`backend/` 目錄)

```bash
cp backend/.env.example backend/.env
```

| 變數 | 說明 | 範例值 |
|---|---|---|
| `GEMINI_API_KEY` | **必填**，Google Gemini API 金鑰 | `AIza...` |
| `FRONTEND_URLS` | 允許 CORS 的前端 URL，逗號分隔 | `http://localhost:5173,http://localhost:5174` |
| `API_HOST` | 後端監聽的 Host | `0.0.0.0` |
| `API_PORT` | 後端監聽的 Port | `8000` |
| `VITE_SUPABASE_URL` | Supabase 專案 URL | `https://xxxx.supabase.co` |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase Server 端用私鑰 (具寫入權限) | `eyJhbG...` |

> [!IMPORTANT]
> `GEMINI_API_KEY` 為啟動 AI 新聞分類功能的必要條件。未填寫時系統會自動退回至 locally mockup 機制。

## 🚀 開發與啟動方式 (Development & Getting Started)

此版本依賴 Supabase 與定時任務，強烈建議**同時啟動後端與前端**以取得完整功能。

### 1. 啟動後端 API (FastAPI)

後端負責每分鐘在背景定時抓取並翻譯新聞，同時寫入至 Supabase。

```bash
cd backend

# 建立並啟動虛擬環境 (Windows)
python -m venv venv
venv\Scripts\activate
# (Mac/Linux 使用: source venv/bin/activate)

# 安裝所需套件
pip install -r requirements.txt

# 複製環境變數範本並填入必要金鑰 (重點: SUPABASE_SERVICE_ROLE_KEY, GEMINI_API_KEY)
cp .env.example .env

# 啟動 FastAPI 伺服器 (附帶熱重載 hot-reload 開發模式)
uvicorn main:app --reload
```
*(後端啟動後會預設運行於 http://localhost:8000，並自動開啟背景任務與 API 服務)*

### 2. 啟動前端 UI (React + Vite)

請開啟**另一個**新的終端機視窗，並回到專案根目錄：

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

## 📁 專案目錄結構

```text
stock-viewing/
├── src/
│   ├── components/
│   │   ├── shared/
│   │   │   └── PriceChangeDisplay.tsx  # 共用漲跌顯示組件
│   │   ├── AIAssistantWidget.tsx       # AI 新聞洞察小組件
│   │   ├── ChartWidget.tsx             # TradingView 走勢圖
│   │   ├── CustomWatchlist.tsx         # 首頁自選股模塊
│   │   ├── Dashboard.tsx               # 國際大盤模塊
│   │   ├── FearGreedIndex.tsx          # 恐懼與貪婪指數
│   │   ├── LayoutSettings.tsx          # 儀表板排版設定
│   │   ├── NewsPage.tsx                # 新聞資訊頁面
│   │   ├── StockDetailPage.tsx         # 個股詳情與股權結構
│   │   ├── SymbolInfoWidget.tsx        # 基礎報價資訊組件
│   │   ├── TaiwanFuturesWidget.tsx     # 台灣市場模塊
│   │   └── WatchlistPage.tsx           # 自選股管理頁面
│   ├── hooks/
│   │   ├── usePolling.ts               # 計時器輪詢 Hook
│   │   └── useTradingViewWidget.ts     # TradingView 嵌入 Hook
│   ├── App.tsx                         # 應用進入點 (路由/全局狀態)
│   ├── main.tsx                        # React 掛載點
│   └── index.css                       # 全局 CSS 變數與核心樣式
├── backend/
│   ├── main.py                         # FastAPI 主應用與排程器
│   ├── news_api.py                     # 新聞路由
│   ├── quotes_api.py                   # 報價路由
│   ├── stock_detail_api.py             # 個股與財務路由
│   ├── shareholders_api.py             # 股權分散表爬蟲路由
│   ├── news_crawler.py                 # RSS + yfinance 爬蟲工具
│   ├── llm_classifier.py               # Gemini AI 分析器
│   ├── database.py                     # Supabase 連線封裝
│   ├── config.py                       # 環境變數管理
│   ├── requirements.txt                # 依賴清單
│   ├── .env                            # 後端環境變數 (gitignored)
│   └── .env.example                    # 後端環境變數範本
├── .env                                # 前端環境變數 (gitignored)
├── .env.example                        # 前端環境變數範本
├── vite.config.ts                      # Vite 設定 (API Proxy)
└── package.json
```
