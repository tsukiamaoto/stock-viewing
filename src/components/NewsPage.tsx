import React, { useEffect, useState } from 'react';
import { ArrowLeft, ExternalLink, RefreshCw, Rss } from 'lucide-react';
import { Link } from 'react-router-dom';

interface FeedItem {
  title: string;
  link: string;
  pubDate: string;
  source: string;
  sourceColor: string;
}

const RSS_SOURCES = [
  {
    name: 'CNN Money',
    url: 'https://rss.cnn.com/rss/money_latest.rss',
    color: '#cc0000',
  },
  {
    name: 'Reuters',
    url: 'https://www.reutersagency.com/feed/?best-topics=business-finance',
    color: '#ff8000',
  },
  {
    name: 'NHK World',
    url: 'https://www3.nhk.or.jp/rss/news/cat0.xml',
    color: '#0068b7',
  },
];

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
    const allItems: FeedItem[] = [];

    for (const source of RSS_SOURCES) {
      try {
        const res = await fetch(
          `https://api.rss2json.com/v1/api.json?rss_url=${encodeURIComponent(source.url)}`
        );
        const json = await res.json();
        if (json.status === 'ok' && json.items) {
          for (const item of json.items.slice(0, 8)) {
            allItems.push({
              title: item.title,
              link: item.link,
              pubDate: item.pubDate,
              source: source.name,
              sourceColor: source.color,
            });
          }
        }
      } catch {
        // Skip failed source
      }
    }

    // Sort by date descending
    allItems.sort((a, b) => new Date(b.pubDate).getTime() - new Date(a.pubDate).getTime());
    setFeedItems(allItems);
    if (allItems.length === 0) {
      setError('無法取得 RSS 資料，請使用下方的快速連結前往各網站查看。');
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
