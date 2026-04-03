import React, { memo } from 'react';
import { useTradingViewWidget } from '../hooks/useTradingViewWidget';

interface ChartWidgetProps {
  symbol: string;
  interval: string;
}

const ChartWidget: React.FC<ChartWidgetProps> = ({ symbol, interval }) => {
  const containerRef = useTradingViewWidget(
    'https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js',
    {
      autosize: true,
      symbol: symbol,
      interval: interval,
      timezone: 'Asia/Taipei',
      theme: 'light',
      style: '3',
      locale: 'zh_TW',
      enable_publishing: false,
      backgroundColor: 'rgba(255, 255, 255, 1)',
      gridColor: 'rgba(0, 0, 0, 0.05)',
      hide_top_toolbar: true,
      hide_legend: false,
      save_image: false,
      container_id: `tv_${symbol.replace(/[^a-zA-Z0-9]/g, '_')}_${interval}`,
      support_host: 'https://www.tradingview.com',
    }
  );

  return <div className="tradingview-widget-container" ref={containerRef} />;
};

export default memo(ChartWidget);
