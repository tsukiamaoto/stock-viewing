import React, { useEffect, useState, useRef } from 'react';
import { Activity, XCircle, CheckCircle2, Terminal, RefreshCw } from 'lucide-react';
import './CrawlerDashboard.css';

interface CrawlerStats {
  source: string;
  success_count: number;
  failure_count: number;
  last_run: string;
}

const CrawlerDashboard: React.FC = () => {
  const [stats, setStats] = useState<CrawlerStats[]>([]);
  const [logs, setLogs] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [logLevel, setLogLevel] = useState('ALL');
  const logContainerRef = useRef<HTMLDivElement>(null);

  const fetchStats = async () => {
    try {
      const res = await fetch('http://localhost:8000/api/system/stats');
      const data = await res.json();
      if (data.status === 'success') {
        setStats(data.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch stats', err);
    }
  };

  const fetchLogs = async () => {
    try {
      const res = await fetch('http://localhost:8000/api/system/logs');
      const data = await res.json();
      if (data.status === 'success') {
        setLogs(data.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch logs', err);
    }
  };

  const refreshAll = async () => {
    setLoading(true);
    await Promise.all([fetchStats(), fetchLogs()]);
    setLoading(false);
  };

  useEffect(() => {
    refreshAll();
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;
    const idx = setInterval(() => {
      fetchStats();
      fetchLogs();
    }, 15000); // 15 seconds refresh
    return () => clearInterval(idx);
  }, [autoRefresh]);

  useEffect(() => {
    // Auto-scroll logs to bottom
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs]);

  const totalSuccess = stats.reduce((acc, curr) => acc + curr.success_count, 0);
  const totalFailure = stats.reduce((acc, curr) => acc + curr.failure_count, 0);

  return (
    <div className="crawler-dashboard-container">
      <div className="crawler-header">
        <h1 className="crawler-title">
          <Activity size={28} /> 系統爬蟲監控
        </h1>
        <div className="crawler-controls">
          <span className="last-update">自動更新: {autoRefresh ? '開啟 (15s)' : '關閉'}</span>
          <button 
            className={`control-btn ${autoRefresh ? 'active' : ''}`}
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw size={16} /> 排程
          </button>
          <button className="control-btn" onClick={refreshAll}>
            立即重新整理
          </button>
        </div>
      </div>

      <div className="stats-grid">
        <div className="stat-card success-card">
          <div className="stat-icon"><CheckCircle2 size={32} /></div>
          <div className="stat-content">
            <div className="stat-label">總成功抓取</div>
            <div className="stat-value">{totalSuccess.toLocaleString()} <span className="text-sm">筆</span></div>
          </div>
        </div>
        <div className="stat-card failure-card">
          <div className="stat-icon"><XCircle size={32} /></div>
          <div className="stat-content">
            <div className="stat-label">總失敗次數</div>
            <div className="stat-value">{totalFailure.toLocaleString()} <span className="text-sm">次</span></div>
          </div>
        </div>
      </div>

      <div className="content-grid">
        <div className="sources-panel">
          <h2 className="panel-title">各來源爬蟲狀態</h2>
          {stats.length === 0 && !loading && (
            <div className="no-data">目前尚無爬蟲執行紀錄</div>
          )}
          <div className="sources-list">
            {stats.map((s, idx) => (
              <div key={idx} className="source-item">
                <div className="source-name">{s.source}</div>
                <div className="source-metrics">
                  <span className="metric-pill success">{s.success_count} 成功</span>
                  {s.failure_count > 0 && <span className="metric-pill failure">{s.failure_count} 失敗</span>}
                </div>
                <div className="source-time">
                  最後執行: {new Date(s.last_run).toLocaleTimeString()}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="logs-panel">
          <div className="panel-title logs-title-bar">
            <span><Terminal size={18} /> 系統實時日誌</span>
            <select 
              className="log-level-select"
              value={logLevel}
              onChange={(e) => setLogLevel(e.target.value)}
            >
              <option value="ALL">全部層級 (ALL)</option>
              <option value="INFO">一般資訊 (INFO)</option>
              <option value="WARN">警告 (WARN)</option>
              <option value="ERROR">錯誤 (ERROR)</option>
            </select>
          </div>
          <div className="logs-terminal" ref={logContainerRef}>
            {loading && logs.length === 0 ? (
              <div className="log-loading">載入中...</div>
            ) : logs.length === 0 ? (
              <div className="log-empty">目前尚無日誌記錄。</div>
            ) : (
              logs.filter(line => {
                if (logLevel === 'ALL') return true;
                return line.includes(`level=${logLevel}`);
              }).map((line, idx) => {
                // Colorize common patterns slightly
                const isError = line.includes("level=ERROR") || line.includes("level=WARN");
                const isInfo = line.includes("level=INFO");
                return (
                  <div key={idx} className={`log-line ${isError ? 'log-error' : ''} ${isInfo ? 'log-info' : ''}`}>
                    {line}
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default CrawlerDashboard;
