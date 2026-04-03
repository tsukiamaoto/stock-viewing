import React, { useEffect, useState, memo } from 'react';
import { TrendingUp, TrendingDown, Minus, RefreshCw } from 'lucide-react';

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
  open: string;
  high: string;
  low: string;
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
    setLoading(true);
    try {
      // Use TWSE open API for real-time stock data
      const res = await fetch('https://openapi.twse.com.tw/v1/exchangeReport/STOCK_DAY_ALL');
      const data = await res.json();

      const matched: StockPrice[] = [];
      for (const stock of stocks) {
        const found = data.find((d: Record<string, string>) => d.Code === stock.code);
        if (found) {
          const close = parseFloat(found.ClosingPrice) || 0;
          const prevClose = parseFloat(found.ClosingPrice) || 0; // STOCK_DAY_ALL may not have prev
          matched.push({
            code: stock.code,
            name: stock.name || found.Name,
            price: found.ClosingPrice || '--',
            change: found.Change || '--',
            changePercent: close && prevClose ? ((parseFloat(found.Change || '0') / (close - parseFloat(found.Change || '0'))) * 100).toFixed(2) : '--',
            open: found.OpeningPrice || '--',
            high: found.HighestPrice || '--',
            low: found.LowestPrice || '--',
            volume: found.TradeVolume ? parseInt(found.TradeVolume).toLocaleString() : '--',
          });
        } else {
          matched.push({
            code: stock.code,
            name: stock.name,
            price: '--',
            change: '--',
            changePercent: '--',
            open: '--',
            high: '--',
            low: '--',
            volume: '--',
          });
        }
      }
      setPrices(matched);
      setLastUpdate(new Date().toLocaleTimeString('zh-TW'));
    } catch {
      // If API fails, show placeholders
      setPrices(stocks.map(s => ({
        code: s.code,
        name: s.name,
        price: '--',
        change: '--',
        changePercent: '--',
        open: '--',
        high: '--',
        low: '--',
        volume: '--',
      })));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPrices();
    // Auto refresh every 60 seconds
    const timer = setInterval(fetchPrices, 60000);
    return () => clearInterval(timer);
  }, [stocks]);

  const getChangeColor = (change: string) => {
    const val = parseFloat(change);
    if (isNaN(val) || val === 0) return 'neutral';
    return val > 0 ? 'up' : 'down';
  };

  const getChangeIcon = (change: string) => {
    const val = parseFloat(change);
    if (isNaN(val) || val === 0) return <Minus size={14} />;
    return val > 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />;
  };

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
      <div className="watchlist-table-wrapper">
        <table className="watchlist-price-table">
          <thead>
            <tr>
              <th>代碼</th>
              <th>名稱</th>
              <th className="text-right">收盤價</th>
              <th className="text-right">漲跌</th>
              <th className="text-right">漲跌%</th>
              <th className="text-right">開盤</th>
              <th className="text-right">最高</th>
              <th className="text-right">最低</th>
              <th className="text-right">成交量</th>
            </tr>
          </thead>
          <tbody>
            {prices.length === 0 && !loading ? (
              <tr>
                <td colSpan={9} className="wl-empty">尚未加入自選股，請前往「自選股管理」新增</td>
              </tr>
            ) : (
              prices.map((stock) => {
                const dir = getChangeColor(stock.change);
                return (
                  <tr key={stock.code} className={`stock-row stock-${dir}`}>
                    <td><span className="wl-code-badge">{stock.code}</span></td>
                    <td className="stock-name">{stock.name}</td>
                    <td className="text-right stock-price">{stock.price}</td>
                    <td className={`text-right stock-change ${dir}`}>
                      {getChangeIcon(stock.change)}
                      {stock.change}
                    </td>
                    <td className={`text-right stock-change ${dir}`}>
                      {stock.changePercent !== '--' ? `${stock.changePercent}%` : '--'}
                    </td>
                    <td className="text-right">{stock.open}</td>
                    <td className="text-right">{stock.high}</td>
                    <td className="text-right">{stock.low}</td>
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
