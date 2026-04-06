import React, { useEffect, useState, useCallback } from 'react';
import { ArrowLeft, ExternalLink, RefreshCw, Loader2, AlertCircle } from 'lucide-react';
import { Link } from 'react-router-dom';
import AIAssistantWidget from './AIAssistantWidget';

interface NewsItem {
  title: string;
  translated_title?: string;
  snippet?: string;
  original_content?: string;
  category?: string;
  link: string;
  pubDate: string;
  source: string;
  sourceColor: string;
}

interface NewsSectionState {
  items: NewsItem[];
  loading: boolean;
  loadingMore: boolean;
  offset: number;
  hasMore: boolean;
  error: string | null;
  lastUpdated: Date | null;
}

// ── 各新聞來源設定 ──────────────────────────────────────────────────
const NEWS_SOURCES = [
  {
    key: 'cnn',
    endpoint: '/api/news/cnn',
    label: 'CNN',
    labelFull: 'CNN 財經',
    color: '#cc0000',
    accentBg: '#fff5f5',
    accentBorder: '#cc0000',
    accentHeader: 'linear-gradient(135deg, #cc0000 0%, #990000 100%)',
    icon: '📺',
    desc: 'CNN Business & World',
    externalUrl: 'https://edition.cnn.com/business',
    titleHoverColor: '#cc0000',
  },
  {
    key: 'reuters',
    endpoint: '/api/news/reuters',
    label: 'REUTERS',
    labelFull: '路透社',
    color: '#e87722',
    accentBg: '#fff8f3',
    accentBorder: '#e87722',
    accentHeader: 'linear-gradient(135deg, #e87722 0%, #c45e0f 100%)',
    icon: '📰',
    desc: 'Reuters Business, World & Markets',
    externalUrl: 'https://www.reuters.com/business/',
    titleHoverColor: '#c45e0f',
  },
  {
    key: 'nhk',
    endpoint: '/api/news/nhk',
    label: 'NHK',
    labelFull: 'NHK World',
    color: '#0068b7',
    accentBg: '#f0f7ff',
    accentBorder: '#0068b7',
    accentHeader: 'linear-gradient(135deg, #0068b7 0%, #004d8a 100%)',
    icon: '🇯🇵',
    desc: 'NHK World News',
    externalUrl: 'https://www3.nhk.or.jp/nhkworld/en/news/',
    titleHoverColor: '#0068b7',
  },
  {
    key: 'jin10',
    endpoint: '/api/news/jin10',
    label: '金十',
    labelFull: '金十數據',
    color: '#c8a000',
    accentBg: '#fffdf0',
    accentBorder: '#c8a000',
    accentHeader: 'linear-gradient(135deg, #c8a000 0%, #9a7b00 100%)',
    icon: '📊',
    desc: 'Jin10 財經快訊',
    externalUrl: 'https://www.jin10.com/',
    titleHoverColor: '#9a7b00',
  },
  {
    key: 'twse-etf',
    endpoint: '/api/news/twse-etf',
    label: 'TWSE',
    labelFull: '台灣證交所',
    color: '#008c95',
    accentBg: '#e0f7fa',
    accentBorder: '#008c95',
    accentHeader: 'linear-gradient(135deg, #008c95 0%, #005662 100%)',
    icon: '📈',
    desc: 'ETF e添富公告',
    externalUrl: 'https://www.twse.com.tw/zh/ETFortune/announcementList',
    titleHoverColor: '#005662',
  },
];

// ── 分類顏色對應 ─────────────────────────────────────────────────────
const CATEGORY_COLORS: Record<string, { bg: string; text: string }> = {
  finance:  { bg: '#dbeafe', text: '#1d4ed8' },
  trade:    { bg: '#dcfce7', text: '#15803d' },
  trump:    { bg: '#fef9c3', text: '#854d0e' },
  ai:       { bg: '#ede9fe', text: '#6d28d9' },
  energy:   { bg: '#ffedd5', text: '#c2410c' },
  geopolitics: { bg: '#fce7f3', text: '#be185d' },
  macro:    { bg: '#e0f2fe', text: '#0369a1' },
  other:    { bg: '#f1f5f9', text: '#475569' },
};

const getCategoryStyle = (cat?: string) => CATEGORY_COLORS[cat || 'other'] || CATEGORY_COLORS['other'];

const formatDate = (dateStr: string) => {
  if (!dateStr) return '—';
  try {
    const d = new Date(dateStr);
    return d.toLocaleString('zh-TW', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
  } catch { return dateStr; }
};


// ── 單一新聞區塊元件 ─────────────────────────────────────────────────
interface NewsSectionProps {
  source: (typeof NEWS_SOURCES)[number];
}

const NewsSection: React.FC<NewsSectionProps> = ({ source }) => {
  const [state, setState] = useState<NewsSectionState>({
    items: [],
    loading: true,
    loadingMore: false,
    offset: 0,
    hasMore: true,
    error: null,
    lastUpdated: null,
  });

  const fetchNews = useCallback(async (isLoadMore = false) => {
    if (isLoadMore) {
        setState(prev => ({ ...prev, loadingMore: true, error: null }));
    } else {
        setState(prev => ({ ...prev, loading: true, error: null, offset: 0 }));
    }

    const currentOffset = isLoadMore ? state.offset + 15 : 0;
    const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8000';
    try {
      const res = await fetch(`${apiBase}${source.endpoint}?limit=15&offset=${currentOffset}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      if (json.status === 'success' && json.data) {
        setState(prev => ({ 
            items: isLoadMore ? [...prev.items, ...json.data] : json.data, 
            loading: false, 
            loadingMore: false,
            offset: currentOffset,
            hasMore: json.data.length === 15,
            error: null, 
            lastUpdated: new Date() 
        }));
      } else {
        setState(prev => ({ ...prev, loading: false, loadingMore: false, hasMore: false, error: json.message || '無資料' }));
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setState(prev => ({ ...prev, loading: false, loadingMore: false, error: msg }));
    }
  }, [source.endpoint, state.offset]);

  useEffect(() => { fetchNews(); }, [source.endpoint]);

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, clientHeight, scrollHeight } = e.currentTarget;
    if (scrollHeight - scrollTop <= clientHeight + 50 && !state.loading && !state.loadingMore && state.hasMore) {
      fetchNews(true);
    }
  };

  return (
    <div className="ns-block">
      {/* 區塊 Header */}
      <div className="ns-block-header" style={{ background: source.accentHeader }}>
        <div className="ns-block-header-left">
          <span className="ns-block-icon">{source.icon}</span>
          <div>
            <span className="ns-block-label">{source.label}</span>
            <span className="ns-block-desc">{source.desc}</span>
          </div>
        </div>
        <div className="ns-block-header-right">
          {state.lastUpdated && (
            <span className="ns-block-updated">
              {state.lastUpdated.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' })}
            </span>
          )}
          <button
            className="ns-refresh-btn"
            onClick={() => fetchNews(false)}
            disabled={state.loading}
            title="重新整理"
          >
            {state.loading
              ? <Loader2 size={14} className="ns-spin" />
              : <RefreshCw size={14} />
            }
          </button>
          <a
            href={source.externalUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="ns-ext-link"
            title={`前往 ${source.labelFull}`}
          >
            <ExternalLink size={13} />
          </a>
        </div>
      </div>

      {/* 內容區 */}
      <div className="ns-block-body" onScroll={handleScroll}>
        {state.loading ? (
          <div className="ns-skeleton-list">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="ns-skeleton-row">
                <div className="ns-skeleton-badge" />
                <div className="ns-skeleton-title" style={{ width: `${60 + i * 6}%` }} />
              </div>
            ))}
          </div>
        ) : state.error ? (
          <div className="ns-error-box">
            <AlertCircle size={16} />
            <span>{state.error}</span>
          </div>
        ) : state.items.length === 0 ? (
          <div className="ns-empty-box">暫無資料，請稍後重試</div>
        ) : (
          <ul className="ns-item-list">
            {state.items.map((item, idx) => {
              const catStyle = getCategoryStyle(item.category);
              const isJin10 = source.key === 'jin10';
              return (
                <li key={`${source.key}-${idx}`} className="ns-item" style={{ '--hover-bg': source.accentBg } as React.CSSProperties}>
                  <a
                    href={item.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="ns-item-link"
                  >
                    {/* 左側：序號 + 分類徽章 */}
                    <div className="ns-item-left">
                      <span className="ns-item-no">{idx + 1}</span>
                      {item.category && item.category !== 'other' && (
                        <span
                          className="ns-cat-badge"
                          style={{ background: catStyle.bg, color: catStyle.text }}
                        >
                          {item.category.toUpperCase()}
                        </span>
                      )}
                    </div>

                    {/* 中間：標題 + 翻譯 + 摘要 */}
                    <div className="ns-item-content">
                      {isJin10 ? (
                        <>
                          {/* Jin10: 翻譯後的繁體中文標題為主 */}
                          <span
                            className="ns-item-title"
                            style={{ '--link-hover': source.titleHoverColor } as React.CSSProperties}
                          >
                            {item.translated_title || item.title}
                          </span>
                          {item.snippet && item.snippet !== item.translated_title && (
                            <span className="ns-item-snippet">{item.snippet}</span>
                          )}
                        </>
                      ) : (
                        <>
                          {/* CNN/Reuters/NHK: 原文標題 */}
                          <span
                            className="ns-item-title"
                            style={{ '--link-hover': source.titleHoverColor } as React.CSSProperties}
                          >
                            {item.title}
                          </span>
                          {/* 翻譯後的中文標題 */}
                          {item.translated_title && item.translated_title !== item.title && (
                            <span className="ns-item-translated">{item.translated_title}</span>
                          )}
                          {/* 翻譯後的中文摘要 */}
                          {item.snippet && (
                            <span className="ns-item-snippet">{item.snippet}</span>
                          )}
                        </>
                      )}
                    </div>

                    {/* 右側：來源 + 時間 */}
                    <div className="ns-item-right">
                      {!isJin10 && item.source && (
                        <span
                          className="ns-source-pill"
                          style={{ background: source.color + '1a', color: source.color, border: `1px solid ${source.color}33` }}
                        >
                          {item.source.replace(`${source.label} `, '').replace(source.labelFull, '').trim() || item.source}
                        </span>
                      )}
                      <span className="ns-item-date">{formatDate(item.pubDate)}</span>
                    </div>
                  </a>
                </li>
              );
            })}
            
            {/* Loading More Indicator */}
            {state.loadingMore && (
              <div style={{ textAlign: 'center', padding: '10px 0', color: '#666' }}>
                <Loader2 size={16} className="ns-spin" style={{ display: 'inline' }} /> 載入中...
              </div>
            )}
            
            {!state.hasMore && state.items.length > 0 && (
              <div style={{ textAlign: 'center', padding: '15px 0', fontSize: '0.8rem', color: '#999' }}>
                已呈現所有資料
              </div>
            )}
          </ul>
        )}
      </div>
    </div>
  );
};


// ── 主頁面 ────────────────────────────────────────────────────────────
const NewsPage: React.FC = () => {
  const [refreshKey, setRefreshKey] = useState(0);

  const refreshAll = () => setRefreshKey(k => k + 1);

  return (
    <div className="ns-page">
      {/* Header */}
      <div className="ns-page-header">
        <Link to="/" className="back-link">
          <ArrowLeft size={20} />
          返回監控首頁
        </Link>
        <div className="ns-page-title-row">
          <h1 className="ns-page-title">📡 最新消息追蹤</h1>
          <button className="ns-refresh-all-btn" onClick={refreshAll}>
            <RefreshCw size={15} />
            全部重新整理
          </button>
        </div>
        <p className="ns-page-subtitle">
          即時監控 CNN、路透社、NHK、金十數據 — 四大國際財經資訊同步呈現
        </p>
      </div>

      {/* AI Assistant */}
      <div style={{ marginBottom: '1.5rem' }}>
        <AIAssistantWidget symbol="Macro" />
      </div>

      {/* 快速連結列 */}
      <div className="ns-quick-bar">
        {NEWS_SOURCES.map(src => (
          <a
            key={src.key}
            href={src.externalUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="ns-quick-chip"
            style={{ '--chip-color': src.color } as React.CSSProperties}
          >
            <span>{src.icon}</span>
            <span>{src.labelFull}</span>
            <ExternalLink size={11} />
          </a>
        ))}
        <a
          href="https://www.twse.com.tw/zh/ETFortune/announcementList"
          target="_blank"
          rel="noopener noreferrer"
          className="ns-quick-chip"
          style={{ '--chip-color': '#003d79' } as React.CSSProperties}
        >
          <span>🏛️</span>
          <span>TWSE ETF 公告</span>
          <ExternalLink size={11} />
        </a>
      </div>

      {/* 四大新聞區塊 */}
      <div className="ns-grid" key={refreshKey}>
        {NEWS_SOURCES.map(source => (
          <NewsSection key={source.key} source={source} />
        ))}
      </div>
    </div>
  );
};

export default NewsPage;
