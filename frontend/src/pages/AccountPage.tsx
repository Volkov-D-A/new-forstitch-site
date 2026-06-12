import React from 'react';
import {
  getCustomerOrders,
  getCustomerSession,
  logoutCustomer,
} from '../services/customerApi';
import { API_BASE_URL } from '../services/siteApi';
import { formatPrice } from '../utils/currency';
import type { CustomerOrder, CustomerSession } from '../types/site';

const statusLabels: Record<string, string> = {
  email_pending: 'Ожидает подтверждения email',
  awaiting_payment: 'Ожидает оплаты',
  paid: 'Оплачен',
  payment_failed: 'Платеж не прошел',
  fulfilled: 'Выполнен',
  cancelled: 'Отменен',
};

function statusText(status: string) {
  return statusLabels[status] || status;
}

function orderTotal(order: CustomerOrder) {
  return order.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
}

function downloadHref(downloadUrl: string) {
  return downloadUrl.startsWith('http') ? downloadUrl : `${API_BASE_URL}${downloadUrl.replace(/^\/api/, '')}`;
}

interface AccountPageProps {
  onAuthRequired: () => void;
  onLoggedOut: () => void;
}

export function AccountPage({ onAuthRequired, onLoggedOut }: AccountPageProps) {
  const [session, setSession] = React.useState<CustomerSession>({ authenticated: false });
  const [orders, setOrders] = React.useState<CustomerOrder[]>([]);
  const [isLoading, setLoading] = React.useState(true);

  const loadOrders = React.useCallback(async () => {
    const nextOrders = await getCustomerOrders();
    setOrders(nextOrders);
  }, []);

  React.useEffect(() => {
    let ignore = false;

    async function load() {
      try {
        const nextSession = await getCustomerSession();
        if (ignore) return;
        setSession(nextSession);
        if (nextSession.authenticated) {
          await loadOrders();
        }
      } catch {
        if (!ignore) setSession({ authenticated: false });
      } finally {
        if (!ignore) setLoading(false);
      }
    }

    load();
    return () => {
      ignore = true;
    };
  }, [loadOrders]);

  const logout = async () => {
    await logoutCustomer();
    onLoggedOut();
    setSession({ authenticated: false });
    setOrders([]);
  };

  if (isLoading) {
    return <div className="app-state">Проверяем вход...</div>;
  }

  if (!session.authenticated) {
    return (
      <div data-screen-label="Личный кабинет">
        <div className="wrap account-page">
          <div className="account-head">
            <h1 className="h-sec page-title">Личный кабинет</h1>
            <p className="lede page-lede">Войдите или зарегистрируйтесь, чтобы оформить заказ и скачать оплаченные схемы.</p>
          </div>
          <button className="btn btn-primary" onClick={onAuthRequired}>Войти или зарегистрироваться</button>
        </div>
      </div>
    );
  }

  return (
    <div data-screen-label="Личный кабинет">
      <div className="wrap account-page">
        <div className="account-top">
          <div>
            <h1 className="h-sec page-title">Личный кабинет</h1>
            <p className="lede page-lede">{session.name || session.email}</p>
          </div>
          <button className="btn btn-outline" onClick={logout}>Выйти</button>
        </div>
        {orders.length === 0 ? (
          <div className="empty-state">
            <p>Заказов пока нет.</p>
          </div>
        ) : (
          <div className="account-orders">
            {orders.map((order) => (
              <article className="account-order" key={order.id}>
                <div className="account-order-head">
                  <div>
                    <h2>{order.id}</h2>
                    <p>{new Date(order.createdAt).toLocaleDateString('ru-RU')}</p>
                  </div>
                  <span className={'order-status status-' + order.status}>{statusText(order.status)}</span>
                </div>
                <div className="account-order-items">
                  {order.items.map((item) => (
                    <div className="account-order-item" key={item.productId}>
                      <div>
                        <strong>{item.productName || item.productId}</strong>
                        <span>{item.quantity} × {formatPrice(item.price)}</span>
                      </div>
                      {item.downloads && item.downloads.length > 0 ? (
                        <div className="account-downloads">
                          {item.downloads.map((file) => (
                            <a className="btn btn-outline btn-sm" href={downloadHref(file.url)} key={file.id}>{file.name}</a>
                          ))}
                        </div>
                      ) : null}
                    </div>
                  ))}
                </div>
                <div className="account-order-foot">
                  <span>{order.message}</span>
                  <strong>{formatPrice(orderTotal(order))}</strong>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
