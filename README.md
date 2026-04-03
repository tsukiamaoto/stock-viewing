# 全球股市監控中心 (Global Stock Market Monitor)

這是一個以 React + Vite 建構的現代化全球股市與金融數據監控前端應用程式，搭配 Python FastAPI 作為後端，提供 AI 輔助新聞分類與 RSS 聚合功能。

## 🛠 技術棧 (Tech Stack)

### 前端 (Frontend)
*   **核心框架:** React 19, TypeScript
*   **建構與開發工具:** Vite, ESLint
*   **路由管理:** React Router DOM (v7)
*   **UI 系統與樣式:** Material UI (MUI), Emotion, 原生 CSS
*   **圖示庫:** Lucide React, MUI Icons Material

### 後端 (Backend)
*   **框架:** Python FastAPI + Uvicorn
*   **AI 分類:** Google Gemini API (`gemini-1.5-flash`)
*   **新聞來源:** RSS (`feedparser`) + yfinance

## 🏛 系統架構 (System Architecture)

```
┌───────────────────────────────────────────────────────┐
│                    React Frontend                     │
│  App.tsx (Router + Global State + localStorage)       │
│  ├── DashboardPage (儀表板)                           │
│  │   ├── TaiwanFuturesWidget  (TWSE/TAIFEX API)       │
│  │   ├── CustomWatchlist      (TWSE OpenAPI)          │
│  │   ├── FearGreedIndex       (alternative.me API)    │
│  │   └── Dashboard            (TradingView Widgets)   │
│  ├── WatchlistPage (自選股管理, TWSE ISIN)             │
│  └── NewsPage      (RSS 新聞, AI 摘要)                │
│                                                       │
│  Shared Hooks / Components:                           │
│  ├── hooks/usePolling.ts           (計時器輪詢)        │
│  ├── hooks/useTradingViewWidget.ts (TV 圖表嵌入)      │
│  └── components/shared/PriceChangeDisplay.tsx         │
└────────────────────────┬──────────────────────────────┘
                         │ REST API (localhost:8000)
┌────────────────────────▼──────────────────────────────┐
│                   FastAPI Backend                     │
│  main.py                                              │
│  ├── GET /api/news/latest?symbol=Macro               │
│  │   └── news_crawler.py (RSS + yfinance)             │
│  └── GET /api/news/categorize/{symbol}               │
│      ├── news_crawler.py                              │
│      └── llm_classifier.py (Gemini API)              │
└───────────────────────────────────────────────────────┘
```

### 核心 Hooks 說明
| Hook / Component | 位置 | 用途 |
|---|---|---|
| `usePolling` | `src/hooks/usePolling.ts` | 封裝 `setInterval` 輪詢邏輯，統一管理計時器生命週期 |
| `useTradingViewWidget` | `src/hooks/useTradingViewWidget.ts` | 統一管理 TradingView 腳本注入與清除 |
| `PriceChangeDisplay` | `src/components/shared/` | 共用的漲跌顏色、方向圖示判斷 |

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

> [!IMPORTANT]
> `GEMINI_API_KEY` 為啟動 AI 新聞分類功能的必要條件。未填寫時系統會自動退回至本地關鍵字分類器 (mock fallback)，其他功能不受影響。

## 🚀 快速開始 (Getting Started)

### 1. 安裝並啟動前端

```bash
npm install
cp .env.example .env   # 依需求修改環境變數
npm run dev
```

### 2. 啟動後端 (選用，新聞功能需要)

```bash
cd backend
python -m venv venv
venv\Scripts\activate  # Windows
pip install -r requirements.txt
cp .env.example .env   # 填入 GEMINI_API_KEY
python main.py
```

### 3. 建置正式環境版本

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
│   ├── main.py                         # FastAPI 應用主程式
│   ├── news_crawler.py                 # RSS + yfinance 爬蟲
│   ├── llm_classifier.py              # Gemini AI 分類器
│   ├── requirements.txt               # Python 套件清單
│   ├── .env                           # 後端環境變數 (gitignored)
│   └── .env.example                   # 後端環境變數範本
├── .env                               # 前端環境變數 (gitignored)
├── .env.example                       # 前端環境變數範本
├── vite.config.ts                     # Vite 設定 (API Proxy)
└── package.json
```
