import { useEffect, useRef } from 'react';

/**
 * Custom hook to safely inject and initialize a TradingView widget.
 * @param scriptSrc The URL of the TradingView embedded script
 * @param config The JS object configuration for the widget
 * @returns containerRef to be attached to the target div
 */
export function useTradingViewWidget<T>(scriptSrc: string, config: T) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    
    // Clear previous widget content completely to prevent duplicates
    containerRef.current.innerHTML = '';
    
    const script = document.createElement('script');
    script.src = scriptSrc;
    script.type = 'text/javascript';
    script.async = true;
    script.innerHTML = JSON.stringify(config);
    
    containerRef.current.appendChild(script);
  }, [scriptSrc, JSON.stringify(config)]);

  return containerRef;
}
