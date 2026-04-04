import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2, ExternalLink, TrendingUp, TrendingDown } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface StockDetailData {
  basic: {
    code: string; shortName: string; longName: string;
    sector: string; industry: string; website: string;
  };
  price: Record<string, any>;
  valuation: Record<string, any>;
  dividends: Record<string, any>;
  ownership: Record<string, any>;
  profitability: Record<string, any>;
  majorHolders: { label: string; value: string }[];
  institutionalHolders: { holder: string; shares: string; dateReported: string; pctHeld: string; value: string }[];
}

interface ShareholderSummary {
  date: string; totalShares: string; totalHolders: string; avgShares: string;
  gt400Shares: string; gt400Pct: string; gt400Count: string;
  range400_600: string; range600_800: string; range800_1000: string;
  gt1000Count: string; gt1000Pct: string; closePrice: string;
  pe: number | string;
}

interface ShareholderDetail {
  dates: string[];
  rows: { range: string; periods: { holders: string; shares: string; pct: string }[] }[];
}

interface ShareholderData {
  code: string;
  eps: number | null;
  summary: ShareholderSummary[];
  detail: ShareholderDetail;
}

const InfoCard: React.FC<{ title: string; items: { label: string; value: any }[] }> = ({ title, items }) => (
  <div className="detail-card">
    <h4 className="detail-card-title">{title}</h4>
    <div className="detail-card-grid">
      {items.map((item, i) => (
        <div key={i} className="detail-item">
          <span className="detail-item-label">{item.label}</span>
          <span className="detail-item-value">{item.value ?? '--'}</span>
        </div>
      ))}
    </div>
  </div>
);

const StockDetailPage: React.FC = () => {
  const { code } = useParams<{ code: string }>();
  const [data, setData] = useState<StockDetailData | null>(null);
  const [shData, setShData] = useState<ShareholderData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showWeeks, setShowWeeks] = useState(12);

  useEffect(() => {
    if (!code) return;
    setLoading(true);
    setError('');
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8000';
    fetch(`${apiUrl}/api/stocks/detail/${code}`)
      .then(r => r.json())
      .then(json => {
        if (json.status === 'success') {
          setData(json.data);
        } else {
          setError(json.message || '載入失敗');
        }
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));

    // Fetch shareholder distribution data (async, non-blocking)
    fetch(`${apiUrl}/api/stocks/shareholders/${code}`)
      .then(r => r.json())
      .then(json => {
        if (json.status === 'success') setShData(json.data);
      })
      .catch(() => {});
  }, [code]);

  if (loading) {
    return (
      <div className="stock-detail-page">
        <div className="stock-detail-loading">
          <Loader2 size={32} className="spin" />
          <p>正在載入 {code} 的詳細資料...</p>
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="stock-detail-page">
        <Link to="/" className="back-link"><ArrowLeft size={20} /> 返回監控首頁</Link>
        <div className="stock-detail-error">⚠️ {error || '載入失敗'}</div>
      </div>
    );
  }

  const { basic, price, valuation, dividends, ownership, profitability, majorHolders, institutionalHolders } = data;

  const priceChange = price.currentPrice !== '--' && price.previousClose !== '--'
    ? (price.currentPrice - price.previousClose) : 0;
  const priceChangePct = price.previousClose && price.previousClose !== '--'
    ? ((priceChange / price.previousClose) * 100) : 0;
  const isUp = priceChange > 0;
  const isDown = priceChange < 0;

  return (
    <div className="stock-detail-page">
      <Link to="/" className="back-link"><ArrowLeft size={20} /> 返回監控首頁</Link>

      {/* Header */}
      <div className="stock-detail-header">
        <div className="stock-detail-header-left">
          <div className="stock-detail-code-row">
            <span className="wl-code-badge" style={{ fontSize: '1.1rem', padding: '4px 14px' }}>{basic.code}</span>
            <h1 className="stock-detail-name">{basic.longName || basic.shortName}</h1>
          </div>
          <div className="stock-detail-tags">
            <span className="stock-detail-tag">{basic.sector}</span>
            <span className="stock-detail-tag">{basic.industry}</span>
          </div>
        </div>
        <div className="stock-detail-header-right">
          <div className={`stock-detail-price ${isUp ? 'up' : isDown ? 'down' : ''}`}>
            {price.currentPrice}
          </div>
          <div className={`stock-detail-price-change ${isUp ? 'up' : isDown ? 'down' : ''}`}>
            {isUp ? <TrendingUp size={18} /> : isDown ? <TrendingDown size={18} /> : null}
            <span>{priceChange >= 0 ? '+' : ''}{priceChange.toFixed(2)} ({priceChangePct >= 0 ? '+' : ''}{priceChangePct.toFixed(2)}%)</span>
          </div>
        </div>
      </div>

      {/* External Links */}
      <div className="stock-detail-links">
        <a href={`https://norway.twsthr.info/StockHolders.aspx?stock=${basic.code}`} target="_blank" rel="noreferrer" className="stock-detail-ext-link">
          <ExternalLink size={14} /> 神秘金字塔
        </a>
        <a href={`https://tw.stock.yahoo.com/quote/${basic.code}.TW`} target="_blank" rel="noreferrer" className="stock-detail-ext-link">
          <ExternalLink size={14} /> Yahoo 股市
        </a>
        <a href={`https://statementdog.com/analysis/tpe/${basic.code}`} target="_blank" rel="noreferrer" className="stock-detail-ext-link">
          <ExternalLink size={14} /> 財報狗
        </a>
        {basic.website && (
          <a href={basic.website} target="_blank" rel="noreferrer" className="stock-detail-ext-link">
            <ExternalLink size={14} /> 公司官網
          </a>
        )}
      </div>

      {/* Cards Grid */}
      <div className="stock-detail-grid">
        <InfoCard title="📊 交易行情" items={[
          { label: '今日開盤', value: price.open },
          { label: '今日最高', value: price.dayHigh },
          { label: '今日最低', value: price.dayLow },
          { label: '前日收盤', value: price.previousClose },
          { label: '成交量', value: price.volume },
          { label: '平均成交量', value: price.averageVolume },
          { label: '52週最高', value: price.fiftyTwoWeekHigh },
          { label: '52週最低', value: price.fiftyTwoWeekLow },
          { label: '50日均線', value: price.fiftyDayAverage },
          { label: '200日均線', value: price.twoHundredDayAverage },
          { label: 'Beta', value: price.beta },
        ]} />

        <InfoCard title="💰 估值指標" items={[
          { label: '市值', value: valuation.marketCap },
          { label: '企業價值', value: valuation.enterpriseValue },
          { label: '本益比 (TTM)', value: valuation.trailingPE },
          { label: '預估本益比', value: valuation.forwardPE },
          { label: '股價淨值比', value: valuation.priceToBook },
          { label: 'EPS (TTM)', value: valuation.trailingEps },
          { label: '預估 EPS', value: valuation.forwardEps },
        ]} />

        <InfoCard title="🏦 股利資訊" items={[
          { label: '股利 (每股)', value: dividends.dividendRate },
          { label: '殖利率', value: dividends.dividendYield },
          { label: '配發率', value: dividends.payoutRatio },
        ]} />

        <InfoCard title="👥 股權結構" items={[
          { label: '流通股數', value: ownership.sharesOutstanding },
          { label: '浮動股數', value: ownership.floatShares },
          { label: '內部人持股', value: ownership.heldPercentInsiders },
          { label: '法人持股', value: ownership.heldPercentInstitutions },
        ]} />

        <InfoCard title="📈 獲利能力" items={[
          { label: '毛利率', value: profitability.grossMargins },
          { label: '營業利益率', value: profitability.operatingMargins },
          { label: '淨利率', value: profitability.profitMargins },
          { label: 'ROE', value: profitability.returnOnEquity },
          { label: 'ROA', value: profitability.returnOnAssets },
          { label: '營收成長', value: profitability.revenueGrowth },
          { label: '盈餘成長', value: profitability.earningsGrowth },
          { label: '總營收', value: profitability.totalRevenue },
          { label: '淨利', value: profitability.netIncome },
        ]} />
      </div>

      {/* Major Holders */}
      {majorHolders.length > 0 && (
        <div className="detail-card" style={{ marginTop: '1.5rem' }}>
          <h4 className="detail-card-title">🏛️ 主要持股概況</h4>
          <table className="detail-holders-table">
            <tbody>
              {majorHolders.map((h, i) => (
                <tr key={i}>
                  <td className="detail-holder-value">{h.label}</td>
                  <td className="detail-holder-label">{h.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Institutional Holders */}
      {institutionalHolders.length > 0 && (
        <div className="detail-card" style={{ marginTop: '1.5rem' }}>
          <h4 className="detail-card-title">🏢 法人機構持股</h4>
          <div style={{ overflowX: 'auto' }}>
            <table className="detail-holders-table">
              <thead>
                <tr>
                  <th>機構名稱</th>
                  <th className="text-right">持有股數</th>
                  <th className="text-right">持股比例</th>
                  <th className="text-right">市值</th>
                  <th className="text-right">更新日</th>
                </tr>
              </thead>
              <tbody>
                {institutionalHolders.map((h, i) => (
                  <tr key={i}>
                    <td style={{ fontWeight: 600 }}>{h.holder}</td>
                    <td className="text-right">{h.shares}</td>
                    <td className="text-right">{h.pctHeld}</td>
                    <td className="text-right">{h.value}</td>
                    <td className="text-right" style={{ color: '#64748b' }}>{h.dateReported}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TDCC 集保戶股權分散表 */}
      {shData && shData.summary.length > 0 && (
        <div className="detail-card" style={{ marginTop: '1.5rem' }}>
          <div className="detail-card-title" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>📊 集保戶股權分散表</span>
            <div style={{ display: 'flex', gap: '6px' }}>
              {[12, 24, 52].map(n => (
                <button key={n} className={`control-btn ${showWeeks === n ? 'active' : ''}`}
                  style={{ padding: '2px 10px', fontSize: '0.75rem' }}
                  onClick={() => setShowWeeks(n)}>
                  {n}週
                </button>
              ))}
            </div>
          </div>

          {/* Line Charts */}
          {(() => {
            const chartData = shData.summary.slice(0, showWeeks).map(r => ({
              date: r.date.slice(5),
              closePrice: parseFloat(r.closePrice) || 0,
              pe: typeof r.pe === 'number' ? r.pe : 0,
              gt400Pct: parseFloat(r.gt400Pct) || 0,
              gt1000Pct: parseFloat(r.gt1000Pct) || 0,
              totalHolders: parseInt(r.totalHolders.replace(/,/g, '')) || 0,
              avgShares: parseFloat(r.avgShares) || 0,
            })).reverse();

            return (
              <div className="shareholders-charts">
                {/* Chart 1: 收盤價 + PE */}
                <div className="shareholders-chart-card">
                  <h5 className="shareholders-chart-title">收盤價 / 本益比 (PE)</h5>
                  <ResponsiveContainer width="100%" height={240}>
                    <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                      <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" />
                      <YAxis yAxisId="price" tick={{ fontSize: 11 }} />
                      <YAxis yAxisId="pe" orientation="right" tick={{ fontSize: 11, fill: '#8b5cf6' }} />
                      <Tooltip contentStyle={{ fontSize: '0.82rem' }} />
                      <Legend wrapperStyle={{ fontSize: '0.82rem' }} />
                      <Line yAxisId="price" type="monotone" dataKey="closePrice" name="收盤價" stroke="#3b82f6" strokeWidth={2} dot={false} />
                      <Line yAxisId="pe" type="monotone" dataKey="pe" name="PE" stroke="#8b5cf6" strokeWidth={2} dot={false} strokeDasharray="5 3" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                {/* Chart 2: 大股東持有比例 */}
                <div className="shareholders-chart-card">
                  <h5 className="shareholders-chart-title">大股東持有比例 (%)</h5>
                  <ResponsiveContainer width="100%" height={240}>
                    <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                      <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" />
                      <YAxis tick={{ fontSize: 11 }} domain={['dataMin - 0.5', 'dataMax + 0.5']} />
                      <Tooltip contentStyle={{ fontSize: '0.82rem' }} formatter={(v: any) => `${Number(v).toFixed(2)}%`} />
                      <Legend wrapperStyle={{ fontSize: '0.82rem' }} />
                      <Line type="monotone" dataKey="gt400Pct" name=">400張 %" stroke="#f59e0b" strokeWidth={2} dot={false} />
                      <Line type="monotone" dataKey="gt1000Pct" name=">1000張 %" stroke="#ef4444" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                {/* Chart 3: 總股東人數 + 平均持股 */}
                <div className="shareholders-chart-card">
                  <h5 className="shareholders-chart-title">總股東人數 / 平均張數</h5>
                  <ResponsiveContainer width="100%" height={240}>
                    <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                      <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" />
                      <YAxis yAxisId="holders" tick={{ fontSize: 11 }} />
                      <YAxis yAxisId="avg" orientation="right" tick={{ fontSize: 11, fill: '#10b981' }} />
                      <Tooltip contentStyle={{ fontSize: '0.82rem' }} />
                      <Legend wrapperStyle={{ fontSize: '0.82rem' }} />
                      <Line yAxisId="holders" type="monotone" dataKey="totalHolders" name="總股東人數" stroke="#6366f1" strokeWidth={2} dot={false} />
                      <Line yAxisId="avg" type="monotone" dataKey="avgShares" name="平均張數/人" stroke="#10b981" strokeWidth={2} dot={false} strokeDasharray="5 3" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>
            );
          })()}

          <div style={{ overflowX: 'auto' }}>
            <table className="detail-holders-table shareholders-table">
              <thead>
                <tr>
                  <th>資料日期</th>
                  <th className="text-right">集保總張數</th>
                  <th className="text-right">總股東人數</th>
                  <th className="text-right">平均張數/人</th>
                  <th className="text-right">&gt;400張持有張數</th>
                  <th className="text-right">&gt;400張持有%</th>
                  <th className="text-right">&gt;400張人數</th>
                  <th className="text-right">400~600</th>
                  <th className="text-right">600~800</th>
                  <th className="text-right">800~1000</th>
                  <th className="text-right">&gt;1000張人數</th>
                  <th className="text-right">&gt;1000張%</th>
                  <th className="text-right">收盤價</th>
                  <th className="text-right">本益比</th>
                </tr>
              </thead>
              <tbody>
                {shData.summary.slice(0, showWeeks).map((row, i) => (
                  <tr key={i}>
                    <td style={{ fontWeight: 600, color: '#3b82f6' }}>{row.date}</td>
                    <td className="text-right">{row.totalShares}</td>
                    <td className="text-right">{row.totalHolders}</td>
                    <td className="text-right">{row.avgShares}</td>
                    <td className="text-right">{row.gt400Shares}</td>
                    <td className="text-right" style={{ fontWeight: 700 }}>{row.gt400Pct}%</td>
                    <td className="text-right">{row.gt400Count}</td>
                    <td className="text-right">{row.range400_600}</td>
                    <td className="text-right">{row.range600_800}</td>
                    <td className="text-right">{row.range800_1000}</td>
                    <td className="text-right">{row.gt1000Count}</td>
                    <td className="text-right" style={{ fontWeight: 700 }}>{row.gt1000Pct}%</td>
                    <td className="text-right" style={{ fontWeight: 700 }}>{row.closePrice}</td>
                    <td className="text-right" style={{ fontWeight: 700, color: '#8b5cf6' }}>{row.pe}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TDCC 詳細分級 */}
      {shData && shData.detail.dates.length > 0 && (
        <div className="detail-card" style={{ marginTop: '1.5rem' }}>
          <h4 className="detail-card-title">📋 近三週持股分級明細</h4>
          <div style={{ overflowX: 'auto' }}>
            <table className="detail-holders-table shareholders-table">
              <thead>
                <tr>
                  <th rowSpan={2}>持股分級</th>
                  {shData.detail.dates.map((d, i) => (
                    <th key={i} colSpan={3} className="text-center" style={{ borderLeft: '2px solid var(--panel-border)' }}>{d}</th>
                  ))}
                </tr>
                <tr>
                  {shData.detail.dates.map((_, i) => (
                    <React.Fragment key={i}>
                      <th className="text-right" style={{ borderLeft: i > 0 ? '2px solid var(--panel-border)' : 'none', fontSize: '0.72rem' }}>人數</th>
                      <th className="text-right" style={{ fontSize: '0.72rem' }}>張數</th>
                      <th className="text-right" style={{ fontSize: '0.72rem' }}>%</th>
                    </React.Fragment>
                  ))}
                </tr>
              </thead>
              <tbody>
                {shData.detail.rows.map((row, i) => (
                  <tr key={i} style={row.range.startsWith('*') ? { fontWeight: 700, background: '#f8fafc' } : {}}>
                    <td style={{ fontWeight: 600, whiteSpace: 'nowrap' }}>{row.range}</td>
                    {row.periods.map((p, pi) => (
                      <React.Fragment key={pi}>
                        <td className="text-right" style={{ borderLeft: pi > 0 ? '2px solid var(--panel-border)' : 'none' }}>{p.holders}</td>
                        <td className="text-right">{p.shares}</td>
                        <td className="text-right">{p.pct}%</td>
                      </React.Fragment>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};

export default StockDetailPage;
