import React, { useState } from 'react';
import { Sparkles, Loader2, ChevronDown, ChevronUp, AlertCircle } from 'lucide-react';
import './AIAssistantWidget.css';

interface AIAssistantWidgetProps {
  symbol: string;
}

interface NewsItem {
  title: string;
  snippet: string;
  source: string;
}

interface Categories {
  trump: NewsItem[];
  hormuz_iran: NewsItem[];
  ai: NewsItem[];
  finance: NewsItem[];
}

const AIAssistantWidget: React.FC<AIAssistantWidgetProps> = ({ symbol }) => {
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [expanded, setExpanded] = useState(false);
  const [categories, setCategories] = useState<Categories | null>(null);

  const handleGenerate = async () => {
    setStatus('loading');
    setExpanded(true);
    
    try {
      const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8000';
      const response = await fetch(`${apiBase}/api/news/categorize/${symbol}`);
      if (!response.ok) throw new Error('API Request Failed');
      const json = await response.json();
      setCategories(json.data.categories);
      setStatus('success');
    } catch (err) {
      console.error(err);
      setStatus('error');
    }
  };

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setExpanded(!expanded);
  };

  return (
    <div className={`ai-widget-container ${expanded ? 'expanded' : ''}`}>
      {!expanded || status === 'idle' ? (
        <button className="ai-generate-btn" onClick={handleGenerate}>
          <Sparkles size={16} className="ai-icon" />
          <span>✨ AI {symbol === 'Macro' ? '總體經濟與大盤' : symbol} 即時洞察</span>
        </button>
      ) : (
        <div className="ai-card glass-effect">
          <div className="ai-card-header" onClick={handleToggle}>
            <div className="ai-header-left">
              <Sparkles size={18} className="ai-icon-active" />
              <span className="ai-title">AI 分析報告: {symbol === 'Macro' ? '全球事件與產業總覽' : symbol}</span>
            </div>
            <button className="ai-toggle-btn" onClick={handleToggle} aria-label={expanded ? '收起' : '展開'}>
              {expanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </button>
          </div>

          {status === 'loading' && (
            <div className="ai-loading-state">
              <Loader2 size={24} className="ai-spinner" />
              <p>正在收集社群情緒與新聞特徵...</p>
            </div>
          )}

          {status === 'success' && categories && (
            <div className="ai-content ai-fade-in">
              <div className="ai-news-categories">
                {categories.trump?.length > 0 && (
                  <div className="ai-category">
                    <div className="ai-category-title"><span className="category-icon">💬</span> 川普相關發言</div>
                    <ul className="ai-news-list">
                      {categories.trump.map((n, i) => <li key={i}>{n.title}</li>)}
                    </ul>
                  </div>
                )}
                {categories.hormuz_iran?.length > 0 && (
                  <div className="ai-category">
                    <div className="ai-category-title"><span className="category-icon">🌊</span> 荷姆茲海峽 & 伊朗</div>
                    <ul className="ai-news-list">
                      {categories.hormuz_iran.map((n, i) => <li key={i}>{n.title}</li>)}
                    </ul>
                  </div>
                )}
                {categories.ai?.length > 0 && (
                  <div className="ai-category">
                    <div className="ai-category-title"><span className="category-icon">🤖</span> AI 相關技術</div>
                    <ul className="ai-news-list">
                      {categories.ai.map((n, i) => <li key={i}>{n.title}</li>)}
                    </ul>
                  </div>
                )}
                {categories.finance?.length > 0 && (
                  <div className="ai-category">
                    <div className="ai-category-title"><span className="category-icon">📈</span> 財經大盤</div>
                    <ul className="ai-news-list">
                      {categories.finance.map((n, i) => <li key={i}>{n.title}</li>)}
                    </ul>
                  </div>
                )}
                {(!categories.trump?.length && !categories.hormuz_iran?.length && !categories.ai?.length && !categories.finance?.length) && (
                  <div style={{ padding: '20px', textAlign: 'center', color: '#64748b' }}>目前無相關特定主題新聞。</div>
                )}
              </div>
              <div className="ai-footer">
                <span className="ai-timestamp">更新時間: {new Date().toLocaleTimeString('zh-TW')}</span>
              </div>
            </div>
          )}

          {status === 'error' && (
            <div className="ai-loading-state" style={{ color: '#ef4444' }}>
              <AlertCircle size={24} />
              <p>無法連接分類伺服器，請確認 backend 已啟動。</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default AIAssistantWidget;
