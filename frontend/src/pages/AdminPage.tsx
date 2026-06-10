import React from 'react';
import {
  AdminAPIError,
  createAdminCategory,
  createAdminProduct,
  deleteAdminCategory,
  deleteAdminProduct,
  getAdminCategories,
  getAdminProducts,
  getAdminSession,
  loginAdmin,
  logoutAdmin,
  updateAdminCategory,
  updateAdminProduct,
} from '../services/adminApi';
import type { Category, Product } from '../types/site';

const emptyCategory: Category = { id: '', label: '' };

const emptyProduct: Product = {
  id: '',
  title: '',
  price: 0,
  cat: '',
  sub: '',
  img: '',
  isNew: false,
  size: '',
  colors: '',
  canvas: '',
};

type AdminTab = 'products' | 'categories';

function getErrorMessage(error: unknown) {
  if (error instanceof AdminAPIError) return `${error.message} (${error.code})`;
  if (error instanceof Error) return error.message;
  return 'Неизвестная ошибка';
}

export function AdminPage() {
  const [username, setUsername] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [csrfToken, setCSRFToken] = React.useState('');
  const [adminName, setAdminName] = React.useState('');
  const [tab, setTab] = React.useState<AdminTab>('products');
  const [categories, setCategories] = React.useState<Category[]>([]);
  const [products, setProducts] = React.useState<Product[]>([]);
  const [categoryForm, setCategoryForm] = React.useState<Category>(emptyCategory);
  const [productForm, setProductForm] = React.useState<Product>(emptyProduct);
  const [editingCategoryId, setEditingCategoryId] = React.useState<string | null>(null);
  const [editingProductId, setEditingProductId] = React.useState<string | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [notice, setNotice] = React.useState<string | null>(null);
  const [isLoading, setLoading] = React.useState(false);
  const isAuthenticated = Boolean(csrfToken);

  const loadData = React.useCallback(async () => {
    if (!csrfToken) return;

    setLoading(true);
    setError(null);
    try {
      const [nextCategories, nextProducts] = await Promise.all([
        getAdminCategories(),
        getAdminProducts(),
      ]);
      setCategories(nextCategories);
      setProducts(nextProducts);
    } catch (loadError) {
      setError(getErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [csrfToken]);

  React.useEffect(() => {
    getAdminSession()
      .then((session) => {
        if (!session.authenticated || !session.csrfToken) return;
        setCSRFToken(session.csrfToken);
        setAdminName(session.username || '');
      })
      .catch(() => undefined);
  }, []);

  React.useEffect(() => {
    loadData();
  }, [loadData]);

  const submitLogin = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const session = await loginAdmin(username.trim(), password);
      setCSRFToken(session.csrfToken);
      setAdminName(session.username);
      setPassword('');
      setNotice('Вход выполнен');
    } catch (loginError) {
      setError(getErrorMessage(loginError));
    }
  };

  const logout = async () => {
    try {
      if (csrfToken) await logoutAdmin(csrfToken);
    } catch {
      // Session might already be gone; local reset is still correct.
    }
    setCSRFToken('');
    setAdminName('');
    setCategories([]);
    setProducts([]);
    setNotice('Вы вышли');
  };

  const submitCategory = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const payload = { ...categoryForm, id: categoryForm.id.trim(), label: categoryForm.label.trim() };
      if (editingCategoryId) {
        await updateAdminCategory(csrfToken, payload);
        setNotice('Категория обновлена');
      } else {
        await createAdminCategory(csrfToken, payload);
        setNotice('Категория создана');
      }
      setCategoryForm(emptyCategory);
      setEditingCategoryId(null);
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const submitProduct = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        ...productForm,
        id: productForm.id.trim(),
        title: productForm.title.trim(),
        price: Number(productForm.price) || 0,
        cat: productForm.cat.trim(),
        sub: productForm.sub.trim(),
        img: productForm.img?.trim(),
        size: productForm.size.trim(),
        colors: productForm.colors.trim(),
        canvas: productForm.canvas.trim(),
      };

      if (editingProductId) {
        await updateAdminProduct(csrfToken, payload);
        setNotice('Товар обновлен');
      } else {
        await createAdminProduct(csrfToken, payload);
        setNotice('Товар создан');
      }
      setProductForm(emptyProduct);
      setEditingProductId(null);
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const editCategory = (category: Category) => {
    setEditingCategoryId(category.id);
    setCategoryForm(category);
  };

  const editProduct = (product: Product) => {
    setEditingProductId(product.id);
    setProductForm({ ...product, img: product.img || '' });
  };

  const removeCategory = async (categoryId: string) => {
    if (!window.confirm(`Удалить категорию ${categoryId}?`)) return;
    try {
      await deleteAdminCategory(csrfToken, categoryId);
      setNotice('Категория удалена');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  const removeProduct = async (productId: string) => {
    if (!window.confirm(`Удалить товар ${productId}?`)) return;
    try {
      await deleteAdminProduct(csrfToken, productId);
      setNotice('Товар удален');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="admin-brand">
          <span className="logo-word">forstitch</span>
          <span className="admin-kicker">admin</span>
        </div>
        <button className={'admin-nav-item' + (tab === 'products' ? ' active' : '')} onClick={() => setTab('products')}>
          Товары
        </button>
        <button className={'admin-nav-item' + (tab === 'categories' ? ' active' : '')} onClick={() => setTab('categories')}>
          Категории
        </button>
      </aside>

      <main className="admin-main">
        <header className="admin-topbar">
          <div>
            <h1>Администрирование</h1>
            <p>{isAuthenticated ? (isLoading ? 'Обновляем данные...' : `${products.length} товаров · ${categories.length} категорий`) : 'Вход по защищенной сессии'}</p>
          </div>
          {isAuthenticated ? (
            <div className="admin-userbar">
              <span>{adminName || 'admin'}</span>
              <button className="btn btn-outline btn-sm" onClick={logout}>Выйти</button>
            </div>
          ) : null}
        </header>

        {error ? <div className="admin-alert error">{error}</div> : null}
        {notice ? <div className="admin-alert success">{notice}</div> : null}

        {!isAuthenticated ? (
          <form className="admin-login admin-panel" onSubmit={submitLogin}>
            <div className="admin-panel-head">
              <h2>Вход</h2>
            </div>
            <div className="admin-form">
              <label>Логин<input autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} /></label>
              <label>Пароль<input autoComplete="current-password" type="password" value={password} onChange={(event) => setPassword(event.target.value)} /></label>
              <div className="admin-form-actions">
                <button className="btn btn-primary" type="submit">Войти</button>
              </div>
            </div>
          </form>
        ) : tab === 'products' ? (
          <ProductsAdmin
            categories={categories}
            editingProductId={editingProductId}
            form={productForm}
            onCancel={() => {
              setEditingProductId(null);
              setProductForm(emptyProduct);
            }}
            onChange={setProductForm}
            onEdit={editProduct}
            onRemove={removeProduct}
            onSubmit={submitProduct}
            products={products}
          />
        ) : (
          <CategoriesAdmin
            editingCategoryId={editingCategoryId}
            form={categoryForm}
            onCancel={() => {
              setEditingCategoryId(null);
              setCategoryForm(emptyCategory);
            }}
            onChange={setCategoryForm}
            onEdit={editCategory}
            onRemove={removeCategory}
            onSubmit={submitCategory}
            categories={categories}
          />
        )}
      </main>
    </div>
  );
}

interface CategoriesAdminProps {
  categories: Category[];
  editingCategoryId: string | null;
  form: Category;
  onCancel: () => void;
  onChange: (category: Category) => void;
  onEdit: (category: Category) => void;
  onRemove: (categoryId: string) => void;
  onSubmit: (event: React.FormEvent) => void;
}

function CategoriesAdmin({ categories, editingCategoryId, form, onCancel, onChange, onEdit, onRemove, onSubmit }: CategoriesAdminProps) {
  return (
    <div className="admin-grid two">
      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Категории</h2>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>ID</th><th>Название</th><th></th></tr>
            </thead>
            <tbody>
              {categories.map((category) => (
                <tr key={category.id}>
                  <td><code>{category.id}</code></td>
                  <td>{category.label}</td>
                  <td className="admin-row-actions">
                    <button onClick={() => onEdit(category)}>Изменить</button>
                    <button onClick={() => onRemove(category.id)}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>{editingCategoryId ? 'Редактирование' : 'Новая категория'}</h2>
        </div>
        <form className="admin-form" onSubmit={onSubmit}>
          <label>ID<input disabled={Boolean(editingCategoryId)} value={form.id} onChange={(event) => onChange({ ...form, id: event.target.value })} /></label>
          <label>Название<input value={form.label} onChange={(event) => onChange({ ...form, label: event.target.value })} /></label>
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">{editingCategoryId ? 'Сохранить' : 'Создать'}</button>
            {editingCategoryId ? <button className="btn btn-outline" type="button" onClick={onCancel}>Отмена</button> : null}
          </div>
        </form>
      </section>
    </div>
  );
}

interface ProductsAdminProps {
  categories: Category[];
  editingProductId: string | null;
  form: Product;
  onCancel: () => void;
  onChange: (product: Product) => void;
  onEdit: (product: Product) => void;
  onRemove: (productId: string) => void;
  onSubmit: (event: React.FormEvent) => void;
  products: Product[];
}

function ProductsAdmin({ categories, editingProductId, form, onCancel, onChange, onEdit, onRemove, onSubmit, products }: ProductsAdminProps) {
  return (
    <div className="admin-grid">
      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Товары</h2>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>ID</th><th>Название</th><th>Категория</th><th>Цена</th><th></th></tr>
            </thead>
            <tbody>
              {products.map((product) => (
                <tr key={product.id}>
                  <td><code>{product.id}</code></td>
                  <td>{product.title}</td>
                  <td>{product.cat}</td>
                  <td>{product.price}</td>
                  <td className="admin-row-actions">
                    <button onClick={() => onEdit(product)}>Изменить</button>
                    <button onClick={() => onRemove(product.id)}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>{editingProductId ? 'Редактирование товара' : 'Новый товар'}</h2>
        </div>
        <form className="admin-form product" onSubmit={onSubmit}>
          <label>ID<input disabled={Boolean(editingProductId)} value={form.id} onChange={(event) => onChange({ ...form, id: event.target.value })} /></label>
          <label>Название<input value={form.title} onChange={(event) => onChange({ ...form, title: event.target.value })} /></label>
          <label>Цена<input type="number" value={form.price} onChange={(event) => onChange({ ...form, price: Number(event.target.value) })} /></label>
          <label>Категория<select value={form.cat} onChange={(event) => onChange({ ...form, cat: event.target.value })}>
            <option value="">Выберите</option>
            {categories.filter((category) => category.id !== 'all').map((category) => (
              <option key={category.id} value={category.id}>{category.label}</option>
            ))}
          </select></label>
          <label>Подкатегория<input value={form.sub} onChange={(event) => onChange({ ...form, sub: event.target.value })} /></label>
          <label>Изображение<input value={form.img || ''} onChange={(event) => onChange({ ...form, img: event.target.value })} /></label>
          <label>Размер<input value={form.size} onChange={(event) => onChange({ ...form, size: event.target.value })} /></label>
          <label>Палитра<input value={form.colors} onChange={(event) => onChange({ ...form, colors: event.target.value })} /></label>
          <label>Основа<input value={form.canvas} onChange={(event) => onChange({ ...form, canvas: event.target.value })} /></label>
          <label className="admin-check"><input type="checkbox" checked={Boolean(form.isNew)} onChange={(event) => onChange({ ...form, isNew: event.target.checked })} /> Новинка</label>
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">{editingProductId ? 'Сохранить' : 'Создать'}</button>
            {editingProductId ? <button className="btn btn-outline" type="button" onClick={onCancel}>Отмена</button> : null}
          </div>
        </form>
      </section>
    </div>
  );
}
