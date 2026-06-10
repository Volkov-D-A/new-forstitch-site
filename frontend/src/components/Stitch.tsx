import React from 'react';

interface StitchProps {
  center?: boolean;
}

export function Stitch({ center }: StitchProps) {
  return (
    <div className={'x-row stitch-row' + (center ? ' center' : '')}>
      <span className="x-mark">× × ×</span>
    </div>
  );
}
