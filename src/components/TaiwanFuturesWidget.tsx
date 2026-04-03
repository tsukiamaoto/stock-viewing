import React, { useEffect, useState, memo } from 'react';
import { TrendingUp, TrendingDown, Minus, RefreshCw } from 'lucide-react';

interface TwseQuote {
  name: string;
  price: string;
  change: string;
  changePercent: string;
  open: string;
  high: string;
  low: string;
  volume: string;
  time: string;
  isUp: boolean | null;
}

// TWSE mis API response field mappings
// z=成交價, y=昨收, o=開盤價, h=最高, l=最低, v=成交量(張), n=名稱, t=時間
function parseTwseData(msgArray: Record<string, string>[]): TwseQuote[] {
  return msgArray.map((item) => {
    const price = item.z && item.z !== '-' ? item.z : '--';
    const prevClose = parseFloat(item.y) || 0;
    const currentPrice = parseFloat(item.z) || 0;
    const change = prevClose && currentPrice ? (currentPrice - prevClose).toFixed(2) : '--';
    const changeNum = parseFloat(change);
    const pct = prevClose ? ((changeNum / prevClose) * 100).toFixed(2) : '--';

    return {
      name: item.n || '未知',
      price,
      change: changeNum > 0 ? `+${change}` : change,
      changePercent: !isNaN(changeNum) ? (changeNum > 0 ? `+${pct}%` : `${pct}%`) : '--',
      open: item.o && item.o !== '-' ? item.o : '--',
      high: item.h && item.h !== '-' ? item.h : '--',
      low: item.l && item.l !== '-' ? item.l : '--',
      volume: item.v ? parseInt(item.v).toLocaleString() : '--',
      time: item.t || '--',
      isUp: isNaN(changeNum) ? null : changeNum > 0 ? true : changeNum < 0 ? false : null,
    };
  });
}

const TaiwanMarketSection: React.FC = () => {
  const [quotes, setQuotes] = useState<TwseQuote[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState('');
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      const allQuotes: TwseQuote[] = [];

      // 1) TWSE: 加權指數 + 櫃買指數
      const twseSymbols = ['tse_t00.tw', 'otc_o00.tw'].join('|');
      const twseRes = await fetch(`/api/twse/getStockInfo.jsp?ex_ch=${twseSymbols}&json=1&delay=0&_=${Date.now()}`);
      if (twseRes.ok) {
        const twseData = await twseRes.json();
        if (twseData?.msgArray) {
          allQuotes.push(...parseTwseData(twseData.msgArray));
        }
      }

      // 2) TAIFEX: 台指期 TXF (透過 TAIFEX 行情 API)
      try {
        const taifexRes = await fetch('/api/taifex/futures/api/getQuoteList', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            CID: 'TXF',
            SymbolType: 'F',
          }),
        });
        if (taifexRes.ok) {
          const taifexData = await taifexRes.json();
          if (taifexData?.RtData?.QuoteList) {
            const txfList = taifexData.RtData.QuoteList as Record<string, string>[];
            // Find regular session (-F) and night/after-hours session (-M)
            const regularSession = txfList.find((q) => q.SymbolID?.endsWith('-F'));
            const nightSession = txfList.find((q) => q.SymbolID?.endsWith('-M'));

            const parseTxf = (txf: Record<string, string>, sessionLabel: string) => {
              const cprice = txf.CLastPrice && txf.CLastPrice !== '0.00' ? txf.CLastPrice : '--';
              const diff = parseFloat(txf.CDiff) || 0;
              const diffRate = txf.CDiffRate || '0.00';
              const timeStr = txf.CTime || '';
              const formattedTime = timeStr.length >= 6
                ? `${timeStr.slice(0, 2)}:${timeStr.slice(2, 4)}:${timeStr.slice(4, 6)}`
                : timeStr;

              return {
                name: `${txf.DispCName || '台指期'} ${sessionLabel}`,
                price: cprice,
                change: diff > 0 ? `+${diff.toFixed(0)}` : `${diff.toFixed(0)}`,
                changePercent: diff > 0 ? `+${diffRate}%` : `${diffRate}%`,
                open: txf.COpenPrice || '--',
                high: txf.CHighPrice || '--',
                low: txf.CLowPrice || '--',
                volume: txf.CTotalVolume ? parseInt(txf.CTotalVolume).toLocaleString() : '--',
                time: formattedTime,
                isUp: diff > 0 ? true : diff < 0 ? false : null,
              } as TwseQuote;
            };

            if (regularSession) {
              allQuotes.push(parseTxf(regularSession, '(日盤)'));
            }
            if (nightSession && nightSession.CLastPrice && nightSession.CLastPrice !== '0.00') {
              allQuotes.push(parseTxf(nightSession, '(夜盤)'));
            }
          }
        }
      } catch {
        console.warn('TAIFEX API unavailable, skipping TXF data');
      }

      if (allQuotes.length > 0) {
        setQuotes(allQuotes);
        setLastUpdate(new Date().toLocaleTimeString('zh-TW'));
        setError(null);
      } else {
        setError('目前非交易時段，暫無即時資料');
      }
    } catch (err) {
      setError('無法連接 TWSE API，請確認網路或稍後再試');
      console.error('TWSE fetch error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    // 每 10 秒自動更新
    const timer = setInterval(fetchData, 10000);
    return () => clearInterval(timer);
  }, []);

  const getIcon = (isUp: boolean | null) => {
    if (isUp === true) return <TrendingUp size={16} />;
    if (isUp === false) return <TrendingDown size={16} />;
    return <Minus size={16} />;
  };

  return (
    <div className="tw-market-section">
      {/* 加權指數大卡片 */}
      {quotes.length > 0 ? (
        quotes.map((q, i) => (
          <div key={i} className="tw-market-card">
            <div className="tw-market-card-header">
              <span className="tw-market-card-name">{q.name}</span>
              <div className="tw-market-card-refresh">
                <RefreshCw size={13} className="spin" />
                <span>即時 · {q.time}</span>
              </div>
            </div>
            <div className="tw-market-card-body">
              <div className="tw-market-price-row">
                <span className={`tw-market-price ${q.isUp === true ? 'up' : q.isUp === false ? 'down' : ''}`}>
                  {q.price}
                </span>
                <span className={`tw-market-change ${q.isUp === true ? 'up' : q.isUp === false ? 'down' : ''}`}>
                  {getIcon(q.isUp)}
                  {q.change} ({q.changePercent})
                </span>
              </div>
              <div className="tw-market-stats">
                <div className="tw-stat">
                  <span className="tw-stat-label">開盤</span>
                  <span className="tw-stat-value">{q.open}</span>
                </div>
                <div className="tw-stat">
                  <span className="tw-stat-label">最高</span>
                  <span className="tw-stat-value">{q.high}</span>
                </div>
                <div className="tw-stat">
                  <span className="tw-stat-label">最低</span>
                  <span className="tw-stat-value">{q.low}</span>
                </div>
                <div className="tw-stat">
                  <span className="tw-stat-label">成交量</span>
                  <span className="tw-stat-value">{q.volume}</span>
                </div>
              </div>
            </div>
          </div>
        ))
      ) : loading ? (
        <div className="tw-market-card">
          <div className="tw-market-card-body" style={{ textAlign: 'center', padding: '2rem' }}>
            <RefreshCw size={20} className="spin" style={{ marginBottom: 8 }} />
            <p style={{ color: '#64748b', margin: 0 }}>正在載入台灣市場資料...</p>
          </div>
        </div>
      ) : error ? (
        <div className="tw-market-card">
          <div className="tw-market-card-body" style={{ textAlign: 'center', padding: '2rem' }}>
            <p style={{ color: '#64748b', margin: 0 }}>{error}</p>
            <button onClick={fetchData} className="tw-retry-btn">🔄 重試</button>
          </div>
        </div>
      ) : null}

      {lastUpdate && (
        <div className="tw-market-update-bar">
          最後更新：{lastUpdate} · 每 10 秒自動刷新
        </div>
      )}
    </div>
  );
};

export default memo(TaiwanMarketSection);
