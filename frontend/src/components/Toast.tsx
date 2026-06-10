import React from 'react';

interface ToastProps {
  text: string;
}

export function Toast({ text }: ToastProps) {
  return <div className="toast"><span className="toast-mark">✓</span>{text}</div>;
}
