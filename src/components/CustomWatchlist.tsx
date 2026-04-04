import React, { useState, memo } from 'react';
import { Link } from 'react-router-dom';
import { RefreshCw } from 'lucide-react';
import { usePolling } from '../hooks/usePolling';
import { getPriceColorClass, PriceChangeIcon } from './shared/PriceChangeDisplay';

interface WatchlistStock {
  code: string;
  name: string;
  symbol: string;
}

interface StockPrice {
  code: string;
  name: string;
  price: string;
  change: string;
  changePercent: string;
  d5_change: string;
  d5_pct: string;
  d7_change: string;
  d7_pct: string;
  volume: string;
}

interface CustomWatchlistProps {
  stocks: WatchlistStock[];
}

const CustomWatchlist: React.FC<CustomWatchlistProps> = ({ stocks }) => {
  const [prices, setPrices] = useState<StockPrice[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<string>('');

  const fetchPrices = async () => {
    if (stocks.length === 0) {
      setPrices([]);
      setLoading(false);
      return;
    }
    
    setLoading(true);
    try {
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8000';
      const symbols = stocks.map(s => s.code).join(',');
      const res = await fetch(`${apiUrl}/api/stocks/watchlist?symbols=${symbols}`);
      
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json = await res.json();
      
      if (json.status === 'success') {
        // Merge the backend data with stock name from props
        const matched: StockPrice[] = stocks.map(stock => {
          const found = json.data.find((d: any) => d.code === stock.code);
          if (found) {
             return {
               code: stock.code,
               name: stock.name,
               price: found.price || '--',
               change: found.change || '--',
               changePercent: found.changePercent || '--',
               d5_change: found.d5_change || '--',
               d5_pct: found.d5_pct || '--',
               d7_change: found.d7_change || '--',
               d7_pct: found.d7_pct || '--',
               volume: found.volume || '--',
             };
          }
          return {
            code: stock.code, name: stock.name, price: '--', change: '--',
            changePercent: '--', d5_change: '--', d5_pct: '--', d7_change: '--', d7_pct: '--', volume: '--'
          };
        });
        setPrices(matched);
      }
      setLastUpdate(new Date().toLocaleTimeString('zh-TW'));
    } catch (err) {
      console.error('Failed to fetch historical quotes:', err);
      // Fallback
      setPrices(stocks.map(s => ({
        code: s.code, name: s.name, price: '--', change: '--', changePercent: '--',
        d5_change: '--', d5_pct: '--', d7_change: '--', d7_pct: '--', volume: '--'
      })));
    } finally {
      setLoading(false);
    }
  };

  usePolling(fetchPrices, 60000, [stocks]);

  return (
    <div className="custom-watchlist">
      <div className="watchlist-toolbar">
        <span className="watchlist-update-time">
          {lastUpdate ? `最後更新：${lastUpdate}` : '載入中...'}
        </span>
        <button className="watchlist-refresh" onClick={fetchPrices} disabled={loading}>
          <RefreshCw size={14} className={loading ? 'spin' : ''} />
        </button>
      </div>
      <div className="watchlist-table-wrapper" style={{ overflowX: 'auto' }}>
        <table className="watchlist-price-table">
          <thead>
            <tr>
              <th>代碼</th>
              <th>名稱 / 股東資訊</th>
              <th className="text-right">收盤價</th>
              <th className="text-right">今日漲跌</th>
              <th className="text-right">今日幅%</th>
              <th className="text-right">5日漲跌</th>
              <th className="text-right">5日幅%</th>
              <th className="text-right">7日漲跌</th>
              <th className="text-right">7日幅%</th>
              <th className="text-right">成交量</th>
            </tr>
          </thead>
          <tbody>
            {prices.length === 0 && !loading ? (
              <tr>
                <td colSpan={10} className="wl-empty">尚未加入自選股，請前往「自選股管理」新增</td>
              </tr>
            ) : (
              prices.map((stock) => {
                const d1dir = getPriceColorClass(stock.change);
                const d5dir = getPriceColorClass(stock.d5_change);
                const d7dir = getPriceColorClass(stock.d7_change);
                
                return (
                  <tr key={stock.code} className={`stock-row stock-${d1dir}`}>
                    <td><span className="wl-code-badge">{stock.code}</span></td>
                    <td className="stock-name">
                      <Link to={`/stock/${stock.code}`} title="查看個股詳細資料" className="stock-detail-link">
                        {stock.name} →
                      </Link>
                    </td>
                    <td className="text-right stock-price">{stock.price}</td>
                    <td className="text-right">
                      <span className={`stock-change ${d1dir}`}>
                        <PriceChangeIcon change={stock.change} />
                        {stock.change}
                      </span>
                    </td>
                    <td className={`text-right ${d1dir}`}>
                      {stock.changePercent !== '--' ? `${stock.changePercent}%` : '--'}
                    </td>
                    <td className="text-right">
                      <span className={`stock-change ${d5dir}`}>
                        <PriceChangeIcon change={stock.d5_change} />
                        {stock.d5_change}
                      </span>
                    </td>
                    <td className={`text-right ${d5dir}`}>
                      {stock.d5_pct !== '--' ? `${stock.d5_pct}%` : '--'}
                    </td>
                    <td className="text-right">
                      <span className={`stock-change ${d7dir}`}>
                        <PriceChangeIcon change={stock.d7_change} />
                        {stock.d7_change}
                      </span>
                    </td>
                    <td className={`text-right ${d7dir}`}>
                      {stock.d7_pct !== '--' ? `${stock.d7_pct}%` : '--'}
                    </td>
                    <td className="text-right">{stock.volume}</td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default memo(CustomWatchlist);
