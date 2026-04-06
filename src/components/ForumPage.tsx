import React, { useState, useEffect } from 'react';
import { RefreshCw, MessageCircle, ExternalLink, ThumbsUp } from 'lucide-react';

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6'];
const getAvatarColor = (name: string) => {
  if (!name) return '#6b7280';
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
  return COLORS[Math.abs(hash) % COLORS.length];
};

interface FeedPostProps {
  post: any;
}

const FeedCard: React.FC<FeedPostProps> = ({ post }) => {
  const author = post.category || '匿名使用者';
  const avatarColor = getAvatarColor(author);
  const comments = post.comments || [];
  
  // Format Date gracefully
  let dateText = post.pubDate;
  try {
    const d = new Date(post.pubDate);
    dateText = d.toLocaleString('zh-TW', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {}

  return (
    <div className="feed-card">
      {/* Header: Author & Time */}
      <div className="feed-card-header">
        <div className="feed-avatar" style={{ backgroundColor: avatarColor }}>
          {author.charAt(0).toUpperCase()}
        </div>
        <div className="feed-meta">
          <div className="feed-author">
            <span className="feed-author-name">{author}</span>
            {post.source && <span className="feed-source-badge" style={{ backgroundColor: post.sourceColor }}>{post.source}</span>}
          </div>
          <div className="feed-time">{dateText}</div>
        </div>
        <a href={post.link} target="_blank" rel="noopener noreferrer" className="feed-ext-link" title="原文連結">
          <ExternalLink size={16} />
        </a>
      </div>

      {/* Body: Title and Snippet */}
      <div className="feed-card-body">
        <a href={post.link} target="_blank" rel="noopener noreferrer" className="feed-title-link" style={{ textDecoration: 'none' }}>
          <h3 className="feed-title" style={{ transition: 'color 0.2s', cursor: 'pointer' }} onMouseOver={(e) => e.currentTarget.style.color = '#3b82f6'} onMouseOut={(e) => e.currentTarget.style.color = ''}>
            {post.title}
          </h3>
        </a>
        {post.snippet && <p className="feed-snippet">{post.snippet}</p>}
      </div>

      {/* Interactions summary */}
      <div className="feed-interactions">
        <div className="feed-interaction-item">
          <ThumbsUp size={14} /> 讚
        </div>
        <div className="feed-interaction-item">
          <MessageCircle size={14} /> {comments.length} 留言
        </div>
      </div>

      {/* Comments Section */}
      {comments.length > 0 && (
        <div className="feed-comments">
          {comments.map((comment: any, idx: number) => (
            <div key={idx} className="feed-comment">
              <div className="feed-comment-avatar" style={{ backgroundColor: getAvatarColor(comment.author) }}>
                {comment.author.charAt(0).toUpperCase()}
              </div>
              <div className="feed-comment-bubble">
                <span className="feed-comment-author">{comment.author}</span>
                <span className="feed-comment-text">{comment.content}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export const ForumPage: React.FC = () => {
  const [pttFeed, setPttFeed] = useState<any[]>([]);
  const [cmoneyFeed, setCmoneyFeed] = useState<any[]>([]);
  const [loadingPtt, setLoadingPtt] = useState(true);
  const [loadingCmoney, setLoadingCmoney] = useState(true);

  // Load Watchlist to inject into CMoney API
  const getWatchlistSymbols = () => {
    try {
      const stored = localStorage.getItem('watchlist');
      if (stored) {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) return parsed;
      }
    } catch {}
    return ['2330', '0050']; // Fallback
  };

  const apiBase = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000';

  const fetchPtt = async () => {
    setLoadingPtt(true);
    try {
      const res = await fetch(`${apiBase}/api/news/ptt`);
      const json = await res.json();
      if (json.data) setPttFeed(json.data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingPtt(false);
    }
  };

  const fetchCmoney = async () => {
    setLoadingCmoney(true);
    try {
      const symbols = getWatchlistSymbols().join(',');
      const res = await fetch(`${apiBase}/api/news/cmoney?symbols=${symbols}`);
      const json = await res.json();
      if (json.data) setCmoneyFeed(json.data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingCmoney(false);
    }
  };

  const loadAll = () => {
    fetchPtt();
    fetchCmoney();
  };

  useEffect(() => {
    loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="forum-page">
      <div className="forum-header">
        <div>
          <h1 className="forum-title">社群論壇討論區</h1>
          <p className="forum-subtitle">即時追蹤 PTT 股版與股市爆料同學會的熱門動態</p>
        </div>
        <button className="forum-refresh-btn" onClick={loadAll} disabled={loadingPtt || loadingCmoney}>
          <RefreshCw size={16} className={(loadingPtt || loadingCmoney) ? 'ns-spin' : ''} />
          刷新動態
        </button>
      </div>

      <div className="forum-layout">
        {/* Left Column: CMoney Feed */}
        <div className="feed-column">
          <div className="feed-column-header" style={{ borderTopColor: '#f7931e' }}>
            <h2>股市爆料同學會</h2>
            <span className="feed-badge">自選股追蹤</span>
          </div>
          <div className="feed-container">
            {loadingCmoney ? (
              <div className="feed-loading">連線載入中...</div>
            ) : cmoneyFeed.length === 0 ? (
              <div className="feed-empty">目前沒有追蹤股票的最新動態。</div>
            ) : (
              Object.entries(
                cmoneyFeed.reduce((acc: any, post: any) => {
                  const sym = post.symbol || '未分類';
                  if (!acc[sym]) acc[sym] = [];
                  acc[sym].push(post);
                  return acc;
                }, {})
              ).map(([sym, posts]: [string, any]) => (
                <div key={sym} className="cmoney-stock-group" style={{ marginBottom: '24px' }}>
                  <div style={{ backgroundColor: '#fff', padding: '8px 16px', borderRadius: '8px', marginBottom: '12px', fontWeight: 'bold', color: '#1e293b', boxShadow: '0 1px 3px rgba(0,0,0,0.1)', borderLeft: '4px solid #f7931e', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    📈 代號 {sym} 熱門討論
                  </div>
                  {posts.map((post: any, idx: number) => <FeedCard key={idx} post={post} />)}
                </div>
              ))
            )}
          </div>
        </div>

        {/* Right Column: PTT Feed */}
        <div className="feed-column">
          <div className="feed-column-header" style={{ borderTopColor: '#2c2c2c' }}>
            <h2>PTT 股版 (Stock)</h2>
            <span className="feed-badge">熱門看板</span>
          </div>
          <div className="feed-container">
            {loadingPtt ? (
              <div className="feed-loading">連線載入中...</div>
            ) : pttFeed.length === 0 ? (
              <div className="feed-empty">目前無法取得 PTT 動態。</div>
            ) : (
              pttFeed.map((post, idx) => <FeedCard key={idx} post={post} />)
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
