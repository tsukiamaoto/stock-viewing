import React, { useEffect, useRef, memo } from 'react';

interface SymbolInfoWidgetProps {
  symbol: string;
}

const SymbolInfoWidget: React.FC<SymbolInfoWidgetProps> = ({ symbol }) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    
    containerRef.current.innerHTML = '';
    
    const script = document.createElement('script');
    script.src = 'https://s3.tradingview.com/external-embedding/embed-widget-symbol-info.js';
    script.type = 'text/javascript';
    script.async = true;
    script.innerHTML = `
      {
        "symbol": "${symbol}",
        "width": "100%",
        "locale": "zh_TW",
        "colorTheme": "light",
        "isTransparent": true
      }
    `;
    
    containerRef.current.appendChild(script);
  }, [symbol]);

  return (
    <div className="tradingview-widget-container" ref={containerRef} style={{ height: '100%', width: '100%' }} />
  );
};

export default memo(SymbolInfoWidget);
