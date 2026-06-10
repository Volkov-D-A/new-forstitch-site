import React from 'react';
import { Link } from 'react-router-dom';

interface LogoProps {
  onClick?: () => void;
}

export function Logo({ onClick }: LogoProps) {
  return (
    <Link className="logo" to="/" onClick={onClick}>
      <div>
        <span className="logo-word">f<span className="logo-x">×</span>rstitch</span>
        <span className="logo-sub">авторские схемы для вышивки</span>
      </div>
    </Link>
  );
}
