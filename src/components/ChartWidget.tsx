import React, { useEffect, useRef, memo } from 'react';

interface ChartWidgetProps {
  symbol: string;
  interval: string;
}

const ChartWidget: React.FC<ChartWidgetProps> = ({ symbol, interval }) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    
    containerRef.current.innerHTML = '';
    
    const script = document.createElement('script');
    script.src = 'https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js';
    script.type = 'text/javascript';
    script.async = true;
    script.innerHTML = `
      {
        "autosize": true,
        "symbol": "${symbol}",
        "interval": "${interval}",
        "timezone": "Asia/Taipei",
        "theme": "light",
        "style": "3",
        "locale": "zh_TW",
        "enable_publishing": false,
        "backgroundColor": "rgba(255, 255, 255, 1)",
        "gridColor": "rgba(0, 0, 0, 0.05)",
        "hide_top_toolbar": true,
        "hide_legend": false,
        "save_image": false,
        "container_id": "tv_${symbol.replace(/[^a-zA-Z0-9]/g, '_')}_${interval}",
        "support_host": "https://www.tradingview.com"
      }
    `;
    
    containerRef.current.appendChild(script);
  }, [symbol, interval]);

  return <div className="tradingview-widget-container" ref={containerRef} />;
};

export default memo(ChartWidget);
