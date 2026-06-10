import React from 'react';

interface EmptyStateProps {
  title: string;
  text?: string;
  action?: React.ReactNode;
}

export function EmptyState({ title, text, action }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <span className="x-mark">× × ×</span>
      <h1 className="h-sec">{title}</h1>
      {text ? <p>{text}</p> : null}
      {action}
    </div>
  );
}
