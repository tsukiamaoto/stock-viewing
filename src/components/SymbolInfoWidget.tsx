import React, { memo } from 'react';
import { useTradingViewWidget } from '../hooks/useTradingViewWidget';

interface SymbolInfoWidgetProps {
  symbol: string;
}

const SymbolInfoWidget: React.FC<SymbolInfoWidgetProps> = ({ symbol }) => {
  const containerRef = useTradingViewWidget(
    'https://s3.tradingview.com/external-embedding/embed-widget-symbol-info.js',
    {
      symbol: symbol,
      width: '100%',
      locale: 'zh_TW',
      colorTheme: 'light',
      isTransparent: true,
    }
  );

  return (
    <div className="tradingview-widget-container" ref={containerRef} style={{ height: '100%', width: '100%' }} />
  );
};

export default memo(SymbolInfoWidget);
