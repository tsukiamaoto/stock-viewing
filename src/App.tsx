import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { Activity, LayoutDashboard, Star, Newspaper } from 'lucide-react';
import Dashboard from './components/Dashboard';
import LayoutSettings, { type WidgetConfig } from './components/LayoutSettings';
import CustomWatchlist from './components/CustomWatchlist';
import FearGreedIndex from './components/FearGreedIndex';
import TaiwanMarketSection from './components/TaiwanFuturesWidget';
import WatchlistPage from './components/WatchlistPage';
import NewsPage from './components/NewsPage';
import type { WatchlistStock } from './components/WatchlistPage';
import './index.css';

const INITIAL_CONFIG: WidgetConfig[] = [
  { id: 'ni225', symbol: 'OANDA:JP225USD', title: '日經 225', subtitle: 'Japan Nikkei 225', width: '1/2', order: 1 },
  { id: 'kospi', symbol: 'AMEX:EWY', title: '韓國綜合', subtitle: 'South Korea KOSPI', width: '1/2', order: 2 },
  { id: 'spx', symbol: 'OANDA:SPX500USD', title: 'S&P 500', subtitle: 'US S&P 500', width: '1/2', order: 3 },
  { id: 'sox', symbol: 'NASDAQ:SOXX', title: '費城半導體', subtitle: 'PHLX Semiconductor', width: '1/2', order: 4 },
  { id: 'brent', symbol: 'OANDA:BCOUSD', title: '布蘭特原油', subtitle: 'Brent Crude Oil', width: '1/2', order: 5 },
  { id: 'tsm', symbol: 'NYSE:TSM', title: '台積電 ADR', subtitle: 'TSMC ADR (NYSE)', width: '1/2', order: 6 },
  { id: 'dji', symbol: 'OANDA:US30USD', title: '道瓊工業', subtitle: 'Dow Jones Industrial', width: '1/2', order: 7 },
];

const DEFAULT_WATCHLIST: WatchlistStock[] = [
  { code: '2330', name: '台積電', symbol: 'TWSE:2330' },
  { code: '2317', name: '鴻海', symbol: 'TWSE:2317' },
  { code: '2454', name: '聯發科', symbol: 'TWSE:2454' },
  { code: '2881', name: '富邦金', symbol: 'TWSE:2881' },
];

/* ---- Navigation Sidebar ---- */
function NavSidebar() {
  const location = useLocation();
  return (
    <nav className="nav-sidebar">
      <div className="nav-sidebar-title">選單</div>
      <Link to="/" className={`nav-item ${location.pathname === '/' ? 'active' : ''}`}>
        <LayoutDashboard size={18} />
        <span>監控儀表板</span>
      </Link>
      <Link to="/watchlist" className={`nav-item ${location.pathname === '/watchlist' ? 'active' : ''}`}>
        <Star size={18} />
        <span>自選股管理</span>
      </Link>
      <Link to="/news" className={`nav-item ${location.pathname === '/news' ? 'active' : ''}`}>
        <Newspaper size={18} />
        <span>最新消息追蹤</span>
      </Link>
    </nav>
  );
}

/* ---- Dashboard Page ---- */
function DashboardPage({
  interval, setInterval, configs, setConfigs, watchlist,
}: {
  interval: string;
  setInterval: (v: string) => void;
  configs: WidgetConfig[];
  setConfigs: (c: WidgetConfig[]) => void;
  watchlist: WatchlistStock[];
}) {
  const intervals = [
    { label: '當日走勢', value: '5' },
    { label: '日線', value: 'D' },
    { label: '周線', value: 'W' },
    { label: '月線', value: 'M' },
    { label: '年線', value: '12M' },
  ];

  return (
    <>
      <div className="page-header">
        <h2 className="page-title">監控儀表板</h2>
        <div className="controls">
          {intervals.map((itm) => (
            <button
              key={itm.value}
              className={`control-btn ${interval === itm.value ? 'active' : ''}`}
              onClick={() => setInterval(itm.value)}
            >
              {itm.label}
            </button>
          ))}
        </div>
      </div>

      <div className="dashboard-content">
        <div className="dashboard-main">

          {/* Row 1: 台灣加權 + 台指期盤後 */}
          <section>
            <h3 className="section-title">🇹🇼 台灣市場</h3>
            <TaiwanMarketSection />
          </section>

          {/* Row 2: 自選股 + Fear & Greed */}
          <section className="top-section">
            <div className="top-left">
              <h3 className="section-title">⭐️ 我的自選股追蹤</h3>
              <div className="chart-panel" style={{ minHeight: 'auto' }}>
                <CustomWatchlist stocks={watchlist} />
              </div>
            </div>
            <div className="top-right">
              <FearGreedIndex />
            </div>
          </section>

          {/* Row 3: 國際指數 */}
          <section className="indices-section">
            <h3 className="section-title">🌐 全球大盤與期指監控</h3>
            <Dashboard interval={interval} configs={configs} />
          </section>
        </div>

        <LayoutSettings configs={configs} onConfigChange={setConfigs} />
      </div>
    </>
  );
}

/* ---- App Root ---- */
function App() {
  const [interval, setInterval] = useState('D');
  const [configs, setConfigs] = useState<WidgetConfig[]>(INITIAL_CONFIG);
  const [watchlist, setWatchlist] = useState<WatchlistStock[]>(() => {
    try {
      const saved = localStorage.getItem('stock_watchlist');
      if (saved) {
        return JSON.parse(saved);
      }
    } catch (e) {
      console.error('Failed to parse watchlist from local storage', e);
    }
    return DEFAULT_WATCHLIST;
  });

  useEffect(() => {
    localStorage.setItem('stock_watchlist', JSON.stringify(watchlist));
  }, [watchlist]);

  return (
    <BrowserRouter>
      <div className="app-shell">
        <header className="top-header">
          <div className="title-container">
            <Activity className="title-icon" size={26} />
            <h1>全球股市監控中心</h1>
          </div>
        </header>

        <div className="app-body">
          <NavSidebar />
          <main className="main-area">
            <Routes>
              <Route path="/" element={
                <DashboardPage
                  interval={interval}
                  setInterval={setInterval}
                  configs={configs}
                  setConfigs={setConfigs}
                  watchlist={watchlist}
                />
              } />
              <Route path="/watchlist" element={
                <WatchlistPage stocks={watchlist} onStocksChange={setWatchlist} />
              } />
              <Route path="/news" element={<NewsPage />} />
            </Routes>
          </main>
        </div>
      </div>
    </BrowserRouter>
  );
}

export default App;
