import React from 'react';
import {
  CustomerAPIError,
  loginCustomer,
  startCustomerRegistration,
  startPasswordReset,
  verifyCustomerRegistration,
  verifyPasswordReset,
} from '../services/customerApi';
import type { CustomerSession } from '../types/site';

type AuthMode = 'login' | 'register' | 'register-code' | 'reset' | 'reset-code';

interface AuthModalProps {
  isOpen: boolean;
  onAuthenticated: (session: CustomerSession) => void;
  onClose: () => void;
}

export function AuthModal({ isOpen, onAuthenticated, onClose }: AuthModalProps) {
  const [mode, setMode] = React.useState<AuthMode>('login');
  const [email, setEmail] = React.useState('');
  const [name, setName] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [newPassword, setNewPassword] = React.useState('');
  const [code, setCode] = React.useState('');
  const [error, setError] = React.useState('');
  const [notice, setNotice] = React.useState('');
  const [isSubmitting, setSubmitting] = React.useState(false);

  if (!isOpen) return null;

  const resetMessages = () => {
    setError('');
    setNotice('');
  };

  const finish = (session: CustomerSession) => {
    onAuthenticated(session);
    onClose();
  };

  const submitLogin = async (event: React.FormEvent) => {
    event.preventDefault();
    if (isSubmitting) return;
    setSubmitting(true);
    resetMessages();
    try {
      finish(await loginCustomer(email.trim(), password));
    } catch (err) {
      setError(err instanceof CustomerAPIError && err.code === 'invalid_credentials' ? 'Неверный email или пароль' : 'Не удалось войти');
    } finally {
      setSubmitting(false);
    }
  };

  const submitRegister = async (event: React.FormEvent) => {
    event.preventDefault();
    if (isSubmitting) return;
    setSubmitting(true);
    resetMessages();
    try {
      const response = await startCustomerRegistration(email.trim(), name.trim(), password);
      setEmail(response.email);
      setNotice(response.message);
      setMode('register-code');
    } catch (err) {
      setError(err instanceof CustomerAPIError && err.code === 'customer_exists' ? 'Этот email уже зарегистрирован' : 'Не удалось отправить код');
    } finally {
      setSubmitting(false);
    }
  };

  const submitRegisterCode = async (event: React.FormEvent) => {
    event.preventDefault();
    if (isSubmitting) return;
    setSubmitting(true);
    resetMessages();
    try {
      finish(await verifyCustomerRegistration(email.trim(), code.trim()));
    } catch {
      setError('Неверный или устаревший код');
    } finally {
      setSubmitting(false);
    }
  };

  const submitReset = async (event: React.FormEvent) => {
    event.preventDefault();
    if (isSubmitting) return;
    setSubmitting(true);
    resetMessages();
    try {
      const response = await startPasswordReset(email.trim());
      setEmail(response.email);
      setNotice(response.message);
      setMode('reset-code');
    } catch {
      setError('Не удалось отправить код восстановления');
    } finally {
      setSubmitting(false);
    }
  };

  const submitResetCode = async (event: React.FormEvent) => {
    event.preventDefault();
    if (isSubmitting) return;
    setSubmitting(true);
    resetMessages();
    try {
      finish(await verifyPasswordReset(email.trim(), code.trim(), newPassword));
    } catch {
      setError('Не удалось восстановить пароль');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="auth-modal-veil" onClick={onClose}>
      <div className="auth-modal" onClick={(event) => event.stopPropagation()}>
        <div className="auth-modal-head">
          <h2>{mode === 'login' ? 'Вход' : mode === 'reset' || mode === 'reset-code' ? 'Восстановление пароля' : 'Регистрация'}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="5" y1="5" x2="19" y2="19"></line><line x1="19" y1="5" x2="5" y2="19"></line></svg>
          </button>
        </div>
        <div className="account-auth-tabs">
          <button className={mode === 'login' ? 'active' : ''} onClick={() => { setMode('login'); resetMessages(); }}>Вход</button>
          <button className={mode === 'register' || mode === 'register-code' ? 'active' : ''} onClick={() => { setMode('register'); resetMessages(); }}>Регистрация</button>
        </div>
        {notice ? <div className="admin-alert success">{notice}</div> : null}
        {error ? <div className="admin-alert error">{error}</div> : null}

        {mode === 'login' ? (
          <form className="account-login auth-form" onSubmit={submitLogin}>
            <label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required /></label>
            <label>Пароль<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" required /></label>
            <button className="btn btn-primary" disabled={isSubmitting}>{isSubmitting ? 'Входим...' : 'Войти'}</button>
            <button className="auth-link" type="button" onClick={() => { setMode('reset'); resetMessages(); }}>Забыли пароль?</button>
          </form>
        ) : mode === 'register' ? (
          <form className="account-login auth-form" onSubmit={submitRegister}>
            <label>Имя<input value={name} onChange={(event) => setName(event.target.value)} autoComplete="name" /></label>
            <label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required /></label>
            <label>Пароль<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="new-password" required minLength={6} /></label>
            <button className="btn btn-primary" disabled={isSubmitting}>{isSubmitting ? 'Отправляем...' : 'Получить код'}</button>
          </form>
        ) : mode === 'register-code' ? (
          <form className="account-login auth-form" onSubmit={submitRegisterCode}>
            <label>Код из письма<input value={code} onChange={(event) => setCode(event.target.value)} inputMode="numeric" autoComplete="one-time-code" required /></label>
            <button className="btn btn-primary" disabled={isSubmitting}>{isSubmitting ? 'Проверяем...' : 'Завершить регистрацию'}</button>
          </form>
        ) : mode === 'reset' ? (
          <form className="account-login auth-form" onSubmit={submitReset}>
            <label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required /></label>
            <button className="btn btn-primary" disabled={isSubmitting}>{isSubmitting ? 'Отправляем...' : 'Получить код'}</button>
          </form>
        ) : (
          <form className="account-login auth-form" onSubmit={submitResetCode}>
            <label>Код из письма<input value={code} onChange={(event) => setCode(event.target.value)} inputMode="numeric" autoComplete="one-time-code" required /></label>
            <label>Новый пароль<input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} autoComplete="new-password" required minLength={6} /></label>
            <button className="btn btn-primary" disabled={isSubmitting}>{isSubmitting ? 'Сохраняем...' : 'Сменить пароль'}</button>
          </form>
        )}
      </div>
    </div>
  );
}
