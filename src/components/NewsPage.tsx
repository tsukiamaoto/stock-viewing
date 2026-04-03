import React, { useEffect, useState } from 'react';
import { ArrowLeft, ExternalLink, RefreshCw, Rss } from 'lucide-react';
import { Link } from 'react-router-dom';
import AIAssistantWidget from './AIAssistantWidget';

interface FeedItem {
  title: string;
  link: string;
  pubDate: string;
  source: string;
  sourceColor: string;
}

// Backend will handle the RSS sources

const QUICK_LINKS = [
  { name: 'CNN 財經', url: 'https://edition.cnn.com/business', color: '#cc0000', icon: '📺' },
  { name: '路透社 財經', url: 'https://www.reuters.com/business/', color: '#ff8000', icon: '📰' },
  { name: 'NHK 新聞', url: 'https://www3.nhk.or.jp/nhkworld/en/news/', color: '#0068b7', icon: '🇯🇵' },
  { name: '金十數據', url: 'https://www.jin10.com/', color: '#1a1a2e', icon: '📊' },
  { name: 'TWSE ETF 公告', url: 'https://www.twse.com.tw/zh/ETFortune/announcementList', color: '#003d79', icon: '🏛️' },
];

const NewsPage: React.FC = () => {
  const [feedItems, setFeedItems] = useState<FeedItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFeeds = async () => {
    setLoading(true);
    setError(null);
    const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8000';
    
    try {
      const res = await fetch(`${apiBase}/api/news/latest?symbol=Macro`);
      if (!res.ok) throw new Error('Network error');
      const json = await res.json();
      
      if (json.status === 'success' && json.data) {
        // Sort by date descending safely
        const sorted = json.data.sort((a: any, b: any) => {
          try {
             return new Date(b.pubDate || 0).getTime() - new Date(a.pubDate || 0).getTime();
          } catch { return 0; }
        });
        setFeedItems(sorted);
      } else {
        setError('無法取得資料，請使用下方的快速連結前往各網站查看。');
      }
    } catch (e) {
      console.error(e);
      setError(`無法連接後端伺服器 (${apiBase})，請確認 FastAPI 已啟動。`);
    }
    
    setLoading(false);
  };


  useEffect(() => {
    fetchFeeds();
  }, []);

  const formatDate = (dateStr: string) => {
    try {
      const d = new Date(dateStr);
      return d.toLocaleString('zh-TW', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <div className="news-page">
      <div className="news-page-header">
        <Link to="/" className="back-link">
          <ArrowLeft size={20} />
          返回監控首頁
        </Link>
        <h1 className="news-page-title">📰 最新財經消息</h1>
        <p className="news-page-subtitle">
          整合 CNN、路透、NHK 等國際財經消息，以及金十數據和 TWSE ETF 公告的快速連結。
        </p>
      </div>

      <div style={{ marginBottom: '24px' }}>
        <AIAssistantWidget symbol="Macro" />
      </div>

      {/* Quick Links */}
      <h2 className="news-section-title">🔗 快速連結</h2>
      <div className="quick-links-grid">
        {QUICK_LINKS.map((link) => (
          <a
            key={link.name}
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            className="quick-link-card"
            style={{ borderLeftColor: link.color }}
          >
            <span className="quick-link-icon">{link.icon}</span>
            <span className="quick-link-name">{link.name}</span>
            <ExternalLink size={14} className="quick-link-ext" />
          </a>
        ))}
      </div>

      {/* RSS Feed */}
      <div className="news-feed-header">
        <h2 className="news-section-title">
          <Rss size={20} />
          RSS 新聞串流
        </h2>
        <button className="news-refresh-btn" onClick={fetchFeeds} disabled={loading}>
          <RefreshCw size={16} className={loading ? 'spin' : ''} />
          重新整理
        </button>
      </div>

      {loading ? (
        <div className="news-loading">正在載入新聞...</div>
      ) : error ? (
        <div className="news-error">{error}</div>
      ) : (
        <div className="news-feed-list">
          {feedItems.map((item, i) => (
            <a
              key={`${item.source}-${i}`}
              href={item.link}
              target="_blank"
              rel="noopener noreferrer"
              className="news-feed-item"
            >
              <span className="news-source-badge" style={{ background: item.sourceColor }}>
                {item.source}
              </span>
              <span className="news-item-title">{item.title}</span>
              <span className="news-item-date">{formatDate(item.pubDate)}</span>
            </a>
          ))}
        </div>
      )}
    </div>
  );
};

export default NewsPage;
