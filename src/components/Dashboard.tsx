import React, { useState, useCallback, useRef } from 'react';
import ChartWidget from './ChartWidget';
import SymbolInfoWidget from './SymbolInfoWidget';
import IndexInfoWidget from './IndexInfoWidget';
import { type WidgetConfig } from './LayoutSettings';
import { Wifi, WifiOff } from 'lucide-react';
import { usePolling } from '../hooks/usePolling';

interface DashboardProps {
  interval: string;
  configs: WidgetConfig[];
}

/*
 * Market schedules: local trading hours + UTC offset.
 * CFD (OANDA) = 24/5, handled separately.
 */
interface MarketSchedule {
  openLocal: number;   // decimal hours, e.g. 9.5 = 09:30
  closeLocal: number;
  utcOffset: number;   // e.g. +8 Taipei, -4 NYSE EDT
}

const MARKET_SCHEDULES: Record<string, MarketSchedule> = {
  // US markets (NYSE/NASDAQ): 09:30-16:00 ET (UTC-4 EDT)
  'AMEX:EWT':     { openLocal: 9.5,  closeLocal: 16,   utcOffset: -4 },
  'SP:SPX500':    { openLocal: 9.5,  closeLocal: 16,   utcOffset: -4 },
  'NASDAQ:SOX':   { openLocal: 9.5,  closeLocal: 16,   utcOffset: -4 },
  'TVC:DJI':      { openLocal: 9.5,  closeLocal: 16,   utcOffset: -4 },
  'NYSE:TSM':     { openLocal: 9.5,  closeLocal: 16,   utcOffset: -4 },
  // Japan: TSE 09:00-15:00 JST (UTC+9)
  'TVC:NI225':    { openLocal: 9.0,  closeLocal: 15,   utcOffset: 9  },
  // Korea: KRX 09:00–15:30 KST (UTC+9)
  'KRX:KOSPI':    { openLocal: 9.0,  closeLocal: 15.5, utcOffset: 9  },
};

function isMarketOpen(symbol: string): boolean {
  const now = new Date();
  const utcDay = now.getUTCDay();

  // CFD (OANDA): Sunday 22:00 UTC → Friday 21:00 UTC
  if (symbol.startsWith('OANDA:')) {
    if (utcDay === 6) return false;
    if (utcDay === 0) return now.getUTCHours() >= 22;
    if (utcDay === 5) return now.getUTCHours() < 21;
    return true;
  }

  // Crypto (BITSTAMP/COINBASE): 24/7
  if (symbol.startsWith('BITSTAMP:') || symbol.startsWith('COINBASE:')) {
    return true;
  }

  if (utcDay === 0 || utcDay === 6) return false;

  const schedule = MARKET_SCHEDULES[symbol];
  if (!schedule) return true;

  const openUtc = (schedule.openLocal - schedule.utcOffset + 24) % 24;
  const closeUtc = (schedule.closeLocal - schedule.utcOffset + 24) % 24;
  const utcHour = now.getUTCHours() + now.getUTCMinutes() / 60;

  if (openUtc < closeUtc) {
    return utcHour >= openUtc && utcHour < closeUtc;
  }
  return utcHour >= openUtc || utcHour < closeUtc;
}

function getStatusText(symbol: string): string {
  if (isMarketOpen(symbol)) return '開盤中';
  const day = new Date().getUTCDay();
  if (day === 0 || day === 6) return '週末休市';
  return '休市';
}

const Dashboard: React.FC<DashboardProps> = ({ interval, configs }) => {
  // refreshKeys: bumped ONCE when market goes closed→open, causing widget remount
  const [refreshKeys, setRefreshKeys] = useState<Record<string, number>>({});
  const [statuses, setStatuses] = useState<Record<string, { text: string; isOpen: boolean }>>({});

  // Track previous open/close state to detect transitions
  const prevOpenRef = useRef<Record<string, boolean>>({});

  const checkMarkets = useCallback(() => {
    const newStatuses: Record<string, { text: string; isOpen: boolean }> = {};
    const prevOpen = prevOpenRef.current;
    const keysToUpdate: string[] = [];

    for (const c of configs) {
      const open = isMarketOpen(c.symbol);
      newStatuses[c.id] = { text: getStatusText(c.symbol), isOpen: open };

      // Detect closed→open transition: refresh the widget once
      if (open && prevOpen[c.id] === false) {
        keysToUpdate.push(c.id);
      }
      prevOpen[c.id] = open;
    }

    setStatuses(newStatuses);

    // Bump keys only for newly-opened markets (causes widget remount = fresh data)
    if (keysToUpdate.length > 0) {
      setRefreshKeys(prev => {
        const next = { ...prev };
        for (const id of keysToUpdate) {
          next[id] = (prev[id] || 0) + 1;
        }
        return next;
      });
    }
  }, [configs]);

  // Check market status every 30 seconds
  usePolling(() => {
    // Initialize previous state
    for (const c of configs) {
      if (prevOpenRef.current[c.id] === undefined) {
        prevOpenRef.current[c.id] = isMarketOpen(c.symbol);
      }
    }
    checkMarkets();
  }, 30000, [checkMarkets, configs]);

  return (
    <div className="dashboard-grid">
      {configs.map((config) => {
        const status = statuses[config.id] || { text: '檢查中...', isOpen: false };
        return (
          <div
            key={config.id}
            className={`chart-panel ${config.width === '1/1' ? 'span-2' : ''}`}
            style={{ order: config.order }}
          >
            <div className="chart-panel-header">
              <div className="chart-panel-title-row">
                <span className="chart-panel-name">{config.title}</span>
                <span className="chart-panel-sub">{config.subtitle}</span>
              </div>
              <div className={`auto-refresh-indicator ${status.isOpen ? 'active' : ''}`}>
                {status.isOpen ? <Wifi size={14} /> : <WifiOff size={14} />}
                <span>{status.text}</span>
              </div>
            </div>
            <div style={{ minHeight: '100px' }}>
              {config.backendSymbol ? (
                <IndexInfoWidget
                  yfSymbol={config.backendSymbol}
                  key={`info-${config.id}-${refreshKeys[config.id] || 0}`}
                />
              ) : (
                <SymbolInfoWidget
                  symbol={config.symbol}
                  key={`info-${config.id}-${refreshKeys[config.id] || 0}`}
                />
              )}
            </div>
            <div className="chart-content">
              <ChartWidget
                symbol={config.chartSymbol || config.symbol}
                interval={interval}
                key={`chart-${config.id}-${refreshKeys[config.id] || 0}`}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default Dashboard;
