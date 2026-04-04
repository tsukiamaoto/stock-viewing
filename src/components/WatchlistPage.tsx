import React, { useState, useEffect } from 'react';
import { X, ArrowLeft, Search, Loader2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import Autocomplete from '@mui/material/Autocomplete';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
export interface WatchlistStock {
  code: string;
  name: string;
  symbol: string;
}

interface StockEntry {
  code: string;
  name: string;
}

interface WatchlistPageProps {
  stocks: WatchlistStock[];
  onStocksChange: (stocks: WatchlistStock[]) => void;
}

/** Fetch and cache the complete stock list from TWSE OpenAPI */
let stockListCache: StockEntry[] | null = null;
let stockListPromise: Promise<StockEntry[]> | null = null;

async function fetchStockList(): Promise<StockEntry[]> {
  if (stockListCache) return stockListCache;
  if (stockListPromise) return stockListPromise;

  stockListPromise = (async () => {
    try {
      // 透過 vite proxy 抓取證交所 OpenAPI `STOCK_DAY_ALL` (收盤每日統計)
      // 回傳格式為 JSON 陣列： [{ Code: '2330', Name: '台積電', ... }]
      const res = await fetch('/api/twse-open/v1/exchangeReport/STOCK_DAY_ALL');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      
      const list: StockEntry[] = [];
      for (const item of data) {
        if (item.Code && item.Name) {
           list.push({ code: item.Code, name: item.Name });
        }
      }
      stockListCache = list;
      return list;
    } catch (err) {
      console.error('Failed to fetch stock list:', err);
      // 清除 promise 讓下次可以重試
      stockListPromise = null;
      return [];
    }
  })();

  return stockListPromise;
}


const WatchlistPage: React.FC<WatchlistPageProps> = ({ stocks, onStocksChange }) => {
  const [allStocks, setAllStocks] = useState<StockEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [inputValue, setInputValue] = useState('');

  // Load stock list on mount
  useEffect(() => {
    fetchStockList().then((list) => {
      setAllStocks(list);
      setLoading(false);
    });
  }, []);

  const handleSelect = (entry: StockEntry | null) => {
    if (!entry) return;
    if (stocks.some(s => s.code === entry.code)) return;
    onStocksChange([...stocks, { code: entry.code, name: entry.name, symbol: `TWSE:${entry.code}` }]);
  };

  const handleRemove = (code: string) => {
    onStocksChange(stocks.filter(s => s.code !== code));
  };

  return (
    <div className="watchlist-page">
      <div className="watchlist-page-header">
        <Link to="/" className="back-link">
          <ArrowLeft size={20} />
          返回監控首頁
        </Link>
        <h1 className="watchlist-page-title">⭐️ 自選股管理</h1>
        <p className="watchlist-page-subtitle">
          輸入股票代碼或公司名稱即可搜尋新增，變更會即時反映在首頁。
        </p>
      </div>

      {/* Search bar with autocomplete */}
      <div className="watchlist-add-row" style={{ position: 'relative', marginTop: '16px' }}>
        <div className="wl-search-wrapper-mui" style={{ flexGrow: 1 }}>
          <Autocomplete
            options={allStocks}
            getOptionLabel={(option) => `${option.code} ${option.name}`}
            loading={loading}
            fullWidth
            inputValue={inputValue}
            onInputChange={(_, newInputValue) => {
              setInputValue(newInputValue);
            }}
            onChange={(_, newValue) => {
              if (newValue) {
                handleSelect(newValue);
                setInputValue('');
              }
            }}
            filterOptions={(options, state) => {
              // Custom filter to allow space-separated multiple keywords matching
              if (!state.inputValue) return options.slice(0, 100); // return top 100 empty state
              const inputWords = state.inputValue.toLowerCase().split(/\s+/).filter(Boolean);
              
              return options.filter((option) => {
                const targetText = `${option.code} ${option.name}`.toLowerCase();
                // Must match ALL input words
                return inputWords.every(word => targetText.includes(word));
              }).slice(0, 50); // limit 50 results for performance
            }}
            blurOnSelect
            clearOnBlur
            getOptionDisabled={(option) => stocks.some(s => s.code === option.code)}
            renderOption={(props, option) => {
              const alreadyAdded = stocks.some((s) => s.code === option.code);
              const { key, ...otherProps } = props as any;
              return (
                <Box component="li" key={key} {...otherProps} sx={{ display: 'flex', justifyContent: 'space-between', width: '100%', py: 1, borderBottom: '1px solid #f0f0f0' }}>
                  <Box>
                    <span style={{ display: 'inline-block', width: '60px', fontWeight: 'bold', color: '#1a73e8' }}>{option.code}</span>
                    <span>{option.name}</span>
                  </Box>
                  {alreadyAdded && <span style={{ fontSize: '0.8rem', color: '#aaa' }}>已加入</span>}
                </Box>
              );
            }}
            renderInput={(params) => (
              <TextField 
                {...params} 
                placeholder={loading ? '正在載入股票清單...' : '搜尋股票代碼或名稱 (例如 2330 或 台積電)'}
                variant="outlined" 
                size="medium"
                sx={{ 
                  backgroundColor: '#fff', 
                  borderRadius: 2,
                  '& .MuiOutlinedInput-root': {
                    paddingLeft: '32px'
                  }
                }}
                InputProps={{
                  ...params.InputProps,
                  startAdornment: (
                    <>
                      <Search size={20} style={{ color: '#666', position: 'absolute', left: '12px', zIndex: 1 }} />
                      {params.InputProps.startAdornment}
                    </>
                  ),
                  endAdornment: (
                    <>
                      {loading ? <Loader2 size={18} className="spin" style={{ position: 'absolute', right: '40px' }} /> : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
            noOptionsText={inputValue ? "找不到符合的股票" : "請輸入股票代碼或名稱開始搜尋"}
          />
        </div>

        <div className="wl-search-hint" style={{ marginTop: '8px', fontSize: '0.9rem', color: '#666' }}>
          {allStocks.length > 0 && (
            <span>📊 已載入 {allStocks.length.toLocaleString()} 檔上市證券</span>
          )}
        </div>
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
                <td colSpan={5} className="wl-empty">尚未加入任何自選股，請在上方搜尋股票代碼或名稱新增</td>
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
        💡 提示：輸入代碼或名稱後，從下拉選單點選即可新增。也可以直接輸入代碼按 Enter 快速新增。
      </div>
    </div>
  );
};

export default WatchlistPage;
