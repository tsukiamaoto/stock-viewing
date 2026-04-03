import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

export function getPriceColorClass(change: string | number | null | boolean): 'up' | 'down' | 'neutral' {
  if (typeof change === 'boolean') {
      return change ? 'up' : 'down';
  }
  if (change === null) return 'neutral';

  const valNum = typeof change === 'string' ? parseFloat(change) : change;
  if (isNaN(valNum) || valNum === 0) return 'neutral';
  return valNum > 0 ? 'up' : 'down';
}

export const PriceChangeIcon: React.FC<{ change: string | number | null | boolean; size?: number }> = ({ change, size = 14 }) => {
  let isUp = false;
  let isDown = false;
  
  if (typeof change === 'boolean') {
    isUp = change;
    isDown = !change;
  } else if (change !== null) {
      const valNum = typeof change === 'string' ? parseFloat(change) : change;
      if (!isNaN(valNum) && valNum !== 0) {
          isUp = valNum > 0;
          isDown = valNum < 0;
      }
  }

  if (isUp) return <TrendingUp size={size} />;
  if (isDown) return <TrendingDown size={size} />;
  return <Minus size={size} />;
};

interface PriceChangeTextProps {
    change: string;
    percent?: string;
    inlinePercent?: boolean;
    size?: number;
}

export const PriceChangeText: React.FC<PriceChangeTextProps> = ({ change, percent, inlinePercent, size }) => {
    return (
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px', justifyContent: 'flex-end' }}>
           <PriceChangeIcon change={change} size={size} />
           <span>{change}</span>
           {inlinePercent && percent && percent !== '--' ? <span>({percent})</span> : null}
        </span>
    );
};
