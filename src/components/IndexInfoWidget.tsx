import React, { useEffect, useState, useCallback, memo } from 'react';

interface IndexData {
  symbol: string;
  price: number;
  change: number;
  changePercent: number;
  open: number;
  high: number;
  low: number;
  prevClose: number;
}

interface IndexInfoWidgetProps {
  yfSymbol: string;   // Yahoo Finance ticker, e.g. "^KS11"
}

const IndexInfoWidget: React.FC<IndexInfoWidgetProps> = ({ yfSymbol }) => {
  const [data, setData]       = useState<IndexData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const res  = await fetch(`http://localhost:8000/api/stocks/index?yf_symbol=${encodeURIComponent(yfSymbol)}`);
      const json = await res.json();
      if (json.status === 'success' && json.data) {
        setData(json.data);
        setError(false);
      } else {
        setError(true);
      }
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [yfSymbol]);

  useEffect(() => {
    fetchData();
    const id = setInterval(fetchData, 60_000);
    return () => clearInterval(id);
  }, [fetchData]);

  if (loading) {
    return (
      <div className="index-info-loading">
        <span className="index-info-dot" />
        <span className="index-info-dot" />
        <span className="index-info-dot" />
      </div>
    );
  }

  if (error || !data) {
    return <div className="index-info-error">資料載入失敗</div>;
  }

  const positive = data.change >= 0;
  const color    = positive ? '#22c55e' : '#ef4444';
  const arrow    = positive ? '▲' : '▼';

  return (
    <div className="index-info-widget">
      {/* Main price row */}
      <div className="index-info-price-row">
        <span className="index-info-price" style={{ color }}>
          {data.price.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
        </span>
        <span className="index-info-change" style={{ color }}>
          {arrow} {Math.abs(data.change).toFixed(2)}&nbsp;
          ({Math.abs(data.changePercent).toFixed(2)}%)
        </span>
      </div>

      {/* OHLC row */}
      <div className="index-info-ohlc">
        <span><em>開</em>{data.open.toLocaleString()}</span>
        <span><em>高</em><span style={{ color: '#22c55e' }}>{data.high.toLocaleString()}</span></span>
        <span><em>低</em><span style={{ color: '#ef4444' }}>{data.low.toLocaleString()}</span></span>
        <span><em>昨收</em>{data.prevClose.toLocaleString()}</span>
      </div>
    </div>
  );
};

export default memo(IndexInfoWidget);
