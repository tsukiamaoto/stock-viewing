import React, { useState } from 'react';
import { Plus, X, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';

export interface WatchlistStock {
  code: string;
  name: string;
  symbol: string;
}

interface WatchlistPageProps {
  stocks: WatchlistStock[];
  onStocksChange: (stocks: WatchlistStock[]) => void;
}

const WatchlistPage: React.FC<WatchlistPageProps> = ({ stocks, onStocksChange }) => {
  const [inputCode, setInputCode] = useState('');
  const [inputName, setInputName] = useState('');

  const handleAdd = () => {
    const code = inputCode.trim();
    const name = inputName.trim() || code;
    if (!code) return;
    if (stocks.some(s => s.code === code)) return;

    onStocksChange([...stocks, { code, name, symbol: `TWSE:${code}` }]);
    setInputCode('');
    setInputName('');
  };

  const handleRemove = (code: string) => {
    onStocksChange(stocks.filter(s => s.code !== code));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleAdd();
  };

  return (
    <div className="watchlist-page">
      <div className="watchlist-page-header">
        <Link to="/" className="back-link">
          <ArrowLeft size={20} />
          返回監控首頁
        </Link>
        <h1 className="watchlist-page-title">⭐️ 自選股管理</h1>
        <p className="watchlist-page-subtitle">在此新增或移除您追蹤的台股股票，變更會即時反映在首頁的自選股區塊。</p>
      </div>

      {/* Add form */}
      <div className="watchlist-add-row">
        <input
          type="text"
          className="wl-input"
          placeholder="台股代碼 (例如 2330)"
          value={inputCode}
          onChange={(e) => setInputCode(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <input
          type="text"
          className="wl-input"
          placeholder="自訂名稱 (選填，例如 台積電)"
          value={inputName}
          onChange={(e) => setInputName(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <button className="wl-add-btn" onClick={handleAdd} disabled={!inputCode.trim()}>
          <Plus size={16} />
          新增股票
        </button>
      </div>

      {/* Table */}
      <div className="wl-table-container">
        <table className="wl-table">
          <thead>
            <tr>
              <th style={{ width: '60px' }}>#</th>
              <th style={{ width: '140px' }}>股票代碼</th>
              <th>名稱</th>
              <th style={{ width: '200px' }}>TradingView 代號</th>
              <th style={{ width: '100px' }}>操作</th>
            </tr>
          </thead>
          <tbody>
            {stocks.length === 0 ? (
              <tr>
                <td colSpan={5} className="wl-empty">尚未加入任何自選股，請在上方輸入股票代碼新增</td>
              </tr>
            ) : (
              stocks.map((stock, index) => (
                <tr key={stock.code}>
                  <td className="wl-cell-center">{index + 1}</td>
                  <td><span className="wl-code-badge">{stock.code}</span></td>
                  <td>{stock.name}</td>
                  <td className="wl-cell-mono">{stock.symbol}</td>
                  <td className="wl-cell-center">
                    <button className="wl-remove-btn" onClick={() => handleRemove(stock.code)} title="移除此股票">
                      <X size={16} />
                      移除
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="wl-tip">
        💡 提示：輸入台股代碼後按 Enter 可以快速新增。新增的股票會自動出現在首頁的「我的自選股追蹤」中。
      </div>
    </div>
  );
};

export default WatchlistPage;
