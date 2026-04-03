import React, { useEffect, useState, memo } from 'react';

interface FearGreedData {
  score: number;
  label: string;
  color: string;
}

function getLabel(score: number): { label: string; color: string } {
  if (score <= 25) return { label: '極度恐懼', color: '#dc2626' };
  if (score <= 45) return { label: '恐懼', color: '#f97316' };
  if (score <= 55) return { label: '中性', color: '#eab308' };
  if (score <= 75) return { label: '貪婪', color: '#22c55e' };
  return { label: '極度貪婪', color: '#16a34a' };
}

const FearGreedIndex: React.FC = () => {
  const [data, setData] = useState<FearGreedData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Try fetching from the alternative API
    const fetchData = async () => {
      try {
        const res = await fetch('https://api.alternative.me/fng/?limit=1');
        const json = await res.json();
        if (json?.data?.[0]) {
          const score = parseInt(json.data[0].value, 10);
          const { label, color } = getLabel(score);
          setData({ score, label, color });
        }
      } catch {
        // Fallback: show a static placeholder
        setData(null);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  // SVG gauge drawing
  const renderGauge = (score: number, label: string, color: string) => {
    const radius = 80;
    const circumference = Math.PI * radius; // semi-circle
    const progress = (score / 100) * circumference;

    return (
      <svg viewBox="0 0 200 120" width="100%" style={{ maxWidth: '280px' }}>
        {/* Background arc */}
        <path
          d="M 20 110 A 80 80 0 0 1 180 110"
          fill="none"
          stroke="#e5e7eb"
          strokeWidth="16"
          strokeLinecap="round"
        />
        {/* Progress arc */}
        <path
          d="M 20 110 A 80 80 0 0 1 180 110"
          fill="none"
          stroke={color}
          strokeWidth="16"
          strokeLinecap="round"
          strokeDasharray={`${progress} ${circumference}`}
          style={{ transition: 'stroke-dasharray 1s ease' }}
        />
        {/* Score text */}
        <text x="100" y="85" textAnchor="middle" fontSize="32" fontWeight="700" fill={color}>
          {score}
        </text>
        {/* Label */}
        <text x="100" y="110" textAnchor="middle" fontSize="13" fontWeight="600" fill="#64748b">
          {label}
        </text>
      </svg>
    );
  };

  return (
    <div className="fear-greed-container">
      <div className="fear-greed-header">
        <span className="fear-greed-title">📊 恐懼與貪婪指數</span>
        <a
          href="https://edition.cnn.com/markets/fear-and-greed"
          target="_blank"
          rel="noopener noreferrer"
          className="fear-greed-link"
        >
          查看 CNN 完整報告 →
        </a>
      </div>
      <div className="fear-greed-body">
        {loading ? (
          <div className="fear-greed-loading">載入中...</div>
        ) : data ? (
          <>
            <div className="fear-greed-gauge">
              {renderGauge(data.score, data.label, data.color)}
            </div>
            <div className="fear-greed-scale">
              <span style={{ color: '#dc2626' }}>極度恐懼</span>
              <span style={{ color: '#f97316' }}>恐懼</span>
              <span style={{ color: '#eab308' }}>中性</span>
              <span style={{ color: '#22c55e' }}>貪婪</span>
              <span style={{ color: '#16a34a' }}>極度貪婪</span>
            </div>
            <p className="fear-greed-source">資料來源：Alternative.me Crypto Fear & Greed Index</p>
          </>
        ) : (
          <div className="fear-greed-fallback">
            <p>無法取得即時資料</p>
            <a
              href="https://edition.cnn.com/markets/fear-and-greed"
              target="_blank"
              rel="noopener noreferrer"
              className="fear-greed-link"
            >
              前往 CNN 查看 →
            </a>
          </div>
        )}
      </div>
    </div>
  );
};

export default memo(FearGreedIndex);
