import React from 'react';
import { Settings, ArrowUp, ArrowDown } from 'lucide-react';

export interface WidgetConfig {
  id: string;
  symbol: string;
  title: string;
  subtitle: string;
  width: '1/2' | '1/1';
  order: number;
  /** Yahoo Finance ticker for real index data (e.g. "^KS11"). When set, uses IndexInfoWidget instead of TradingView SymbolInfo. */
  backendSymbol?: string;
}

interface LayoutSettingsProps {
  configs: WidgetConfig[];
  onConfigChange: (newConfigs: WidgetConfig[]) => void;
}

const LayoutSettings: React.FC<LayoutSettingsProps> = ({ configs, onConfigChange }) => {
  // Sort configs by order for rendering in the settings panel
  const sortedConfigs = [...configs].sort((a, b) => a.order - b.order);

  const handleWidthChange = (id: string, width: '1/2' | '1/1') => {
    onConfigChange(
      configs.map(c => c.id === id ? { ...c, width } : c)
    );
  };

  const handleMoveUp = (id: string) => {
    const idx = sortedConfigs.findIndex(c => c.id === id);
    if (idx > 0) {
      const prev = sortedConfigs[idx - 1];
      const curr = sortedConfigs[idx];
      
      onConfigChange(configs.map(c => {
        if (c.id === curr.id) return { ...c, order: prev.order };
        if (c.id === prev.id) return { ...c, order: curr.order };
        return c;
      }));
    }
  };

  const handleMoveDown = (id: string) => {
    const idx = sortedConfigs.findIndex(c => c.id === id);
    if (idx < sortedConfigs.length - 1) {
      const next = sortedConfigs[idx + 1];
      const curr = sortedConfigs[idx];
      
      onConfigChange(configs.map(c => {
        if (c.id === curr.id) return { ...c, order: next.order };
        if (c.id === next.id) return { ...c, order: curr.order };
        return c;
      }));
    }
  };

  return (
    <aside className="settings-sidebar">
      <div className="settings-header">
        <Settings size={20} color="#3b82f6" />
        自訂排版與位置
      </div>
      
      <div className="settings-list">
        {sortedConfigs.map((config, index) => (
          <div key={config.id} className="settings-item">
            <div className="settings-item-header">
              <span>{config.title}</span>
              <div className="settings-actions">
                <button 
                  className="icon-btn" 
                  onClick={() => handleMoveUp(config.id)}
                  disabled={index === 0}
                  title="上移"
                >
                  <ArrowUp size={16} />
                </button>
                <button 
                  className="icon-btn" 
                  onClick={() => handleMoveDown(config.id)}
                  disabled={index === sortedConfigs.length - 1}
                  title="下移"
                >
                  <ArrowDown size={16} />
                </button>
              </div>
            </div>
            
            <div className="width-control">
              <button 
                className={`width-btn ${config.width === '1/2' ? 'active' : ''}`}
                onClick={() => handleWidthChange(config.id, '1/2')}
              >
                半寬 (50%)
              </button>
              <button 
                className={`width-btn ${config.width === '1/1' ? 'active' : ''}`}
                onClick={() => handleWidthChange(config.id, '1/1')}
              >
                全寬 (100%)
              </button>
            </div>
          </div>
        ))}
      </div>
    </aside>
  );
};

export default LayoutSettings;
