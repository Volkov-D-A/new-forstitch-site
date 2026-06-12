import React from 'react';
import {
  AdminAPIError,
  createAdminCategory,
  createAdminBlogPost,
  createAdminProduct,
  createAdminGalleryItem,
  createAdminTestimonial,
  deleteAdminBlogPost,
  deleteAdminCategory,
  deleteAdminGalleryItem,
  deleteAdminProduct,
  deleteAdminProductFile,
  deleteAdminProductImage,
  deleteAdminTestimonial,
  getAdminBlog,
  getAdminCategories,
  getAdminGallery,
  getAdminOrders,
  getAdminProducts,
  getAdminSession,
  getAdminSiteSettings,
  getAdminTestimonials,
  loginAdmin,
  logoutAdmin,
  uploadAdminBlogPostImage,
  uploadAdminGalleryItemImage,
  uploadAdminProductAdditionalImage,
  uploadAdminProductFile,
  uploadAdminTestimonialImage,
  uploadAdminProductImage,
  updateAdminCategory,
  updateAdminBlogPost,
  updateAdminGalleryItem,
  updateAdminProduct,
  updateAdminSiteSettings,
  updateAdminTestimonial,
} from '../services/adminApi';
import type { BlogPost, Category, CustomerOrder, GalleryItem, Product, SiteSettings, Testimonial } from '../types/site';
import { formatPrice } from '../utils/currency';

const RichTextEditor = React.lazy(() => import('../components/RichTextEditor').then((module) => ({ default: module.RichTextEditor })));

const emptyCategory: Category = { id: '', label: '' };

const emptyProduct: Product = {
  id: '',
  title: '',
  price: 0,
  cat: '',
  img: '',
  size: '',
  colors: '',
  description: '',
};

const emptyBlogPost: BlogPost = {
  id: '',
  title: '',
  date: new Date().toISOString().slice(0, 10),
  tag: '',
  img: '',
  excerpt: '',
  content: '',
};

const emptyGalleryItem: GalleryItem = {
  img: '',
  title: '',
  description: '',
};

const emptySiteSettings: SiteSettings = { featuredProductId: '' };

const emptyTestimonial: Testimonial = {
  name: '',
  role: '',
  img: '',
  text: '',
};

type AdminTab = 'products' | 'categories' | 'blog' | 'gallery' | 'orders' | 'settings';

const orderStatusLabels: Record<string, string> = {
  paid: 'Оплачен',
  fulfilled: 'Выполнен',
  cancelled: 'Отменен',
  payment_failed: 'Платеж не прошел',
  awaiting_payment: 'Ожидает оплаты',
  email_pending: 'Ожидает email',
};

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
  const [blogPosts, setBlogPosts] = React.useState<BlogPost[]>([]);
  const [galleryItems, setGalleryItems] = React.useState<GalleryItem[]>([]);
  const [orders, setOrders] = React.useState<CustomerOrder[]>([]);
  const [categoryForm, setCategoryForm] = React.useState<Category>(emptyCategory);
  const [productForm, setProductForm] = React.useState<Product>(emptyProduct);
  const [blogForm, setBlogForm] = React.useState<BlogPost>(emptyBlogPost);
  const [galleryForm, setGalleryForm] = React.useState<GalleryItem>(emptyGalleryItem);
  const [siteSettings, setSiteSettings] = React.useState<SiteSettings>(emptySiteSettings);
  const [testimonials, setTestimonials] = React.useState<Testimonial[]>([]);
  const [testimonialForm, setTestimonialForm] = React.useState<Testimonial>(emptyTestimonial);
  const [selectedProductImage, setSelectedProductImage] = React.useState<File | null>(null);
  const [selectedProductImages, setSelectedProductImages] = React.useState<File[]>([]);
  const [selectedProductFiles, setSelectedProductFiles] = React.useState<File[]>([]);
  const [selectedBlogImage, setSelectedBlogImage] = React.useState<File | null>(null);
  const [selectedGalleryImage, setSelectedGalleryImage] = React.useState<File | null>(null);
  const [selectedTestimonialImage, setSelectedTestimonialImage] = React.useState<File | null>(null);
  const [editingCategoryId, setEditingCategoryId] = React.useState<string | null>(null);
  const [editingProductId, setEditingProductId] = React.useState<string | null>(null);
  const [editingBlogPostId, setEditingBlogPostId] = React.useState<string | null>(null);
  const [editingGalleryItemId, setEditingGalleryItemId] = React.useState<number | null>(null);
  const [editingTestimonialId, setEditingTestimonialId] = React.useState<number | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [notice, setNotice] = React.useState<string | null>(null);
  const [isLoading, setLoading] = React.useState(false);
  const isAuthenticated = Boolean(csrfToken);

  const loadData = React.useCallback(async () => {
    if (!csrfToken) return;

    setLoading(true);
    setError(null);
    try {
      const [nextCategories, nextProducts, nextBlogPosts, nextGalleryItems, nextSettings, nextTestimonials, nextOrders] = await Promise.all([
        getAdminCategories(),
        getAdminProducts(),
        getAdminBlog(),
        getAdminGallery(),
        getAdminSiteSettings(),
        getAdminTestimonials(),
        getAdminOrders(),
      ]);
      setCategories(nextCategories);
      setProducts(nextProducts);
      setBlogPosts(nextBlogPosts);
      setGalleryItems(nextGalleryItems);
      setSiteSettings(nextSettings);
      setTestimonials(nextTestimonials);
      setOrders(nextOrders);
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
    setBlogPosts([]);
    setGalleryItems([]);
    setSiteSettings(emptySiteSettings);
    setTestimonials([]);
    setOrders([]);
    setSelectedTestimonialImage(null);
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
        await createAdminCategory(csrfToken, { ...payload, id: '' });
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
        img: productForm.img?.trim(),
        size: productForm.size.trim(),
        colors: productForm.colors.trim(),
        description: productForm.description?.trim(),
      };

      let savedProduct: Product;
      if (editingProductId) {
        savedProduct = await updateAdminProduct(csrfToken, payload);
        setNotice('Товар обновлен');
      } else {
        savedProduct = await createAdminProduct(csrfToken, { ...payload, id: '' });
        setNotice('Товар создан');
      }

      if (selectedProductImage) {
        await uploadAdminProductImage(csrfToken, savedProduct.id, selectedProductImage);
        setNotice(editingProductId ? 'Товар и изображение обновлены' : 'Товар создан с изображением');
      }
      for (const image of selectedProductImages) {
        await uploadAdminProductAdditionalImage(csrfToken, savedProduct.id, image);
      }
      if (selectedProductImages.length > 0) {
        setNotice('Изображения товара обновлены');
      }
      for (const file of selectedProductFiles) {
        await uploadAdminProductFile(csrfToken, savedProduct.id, file);
      }
      if (selectedProductFiles.length > 0) {
        setNotice('Файлы товара обновлены');
      }

      setProductForm(emptyProduct);
      setSelectedProductImage(null);
      setSelectedProductImages([]);
      setSelectedProductFiles([]);
      setEditingProductId(null);
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const submitSiteSettings = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const updated = await updateAdminSiteSettings(csrfToken, siteSettings);
      setSiteSettings(updated);
      setNotice('Настройки главной обновлены');
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const submitBlogPost = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        ...blogForm,
        id: blogForm.id.trim(),
        title: blogForm.title.trim(),
        date: blogForm.date.trim(),
        tag: blogForm.tag.trim(),
        img: blogForm.img.trim(),
        excerpt: '',
        content: blogForm.content.trim(),
      };

      let savedPost: BlogPost;
      if (editingBlogPostId) {
        savedPost = await updateAdminBlogPost(csrfToken, payload);
        setNotice('Запись блога обновлена');
      } else {
        savedPost = await createAdminBlogPost(csrfToken, { ...payload, id: '' });
        setNotice('Запись блога создана');
      }

      if (selectedBlogImage) {
        await uploadAdminBlogPostImage(csrfToken, savedPost.id, selectedBlogImage);
        setNotice(editingBlogPostId ? 'Запись и обложка обновлены' : 'Запись создана с обложкой');
      }

      setBlogForm(emptyBlogPost);
      setSelectedBlogImage(null);
      setEditingBlogPostId(null);
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const submitGalleryItem = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        ...galleryForm,
        img: galleryForm.img.trim(),
        title: galleryForm.title.trim(),
        description: galleryForm.description.trim(),
      };

      let savedItem: GalleryItem;
      if (editingGalleryItemId) {
        savedItem = await updateAdminGalleryItem(csrfToken, { ...payload, id: editingGalleryItemId });
        setNotice('Элемент галереи обновлен');
      } else {
        savedItem = await createAdminGalleryItem(csrfToken, payload);
        setNotice('Элемент галереи создан');
      }

      if (selectedGalleryImage && savedItem.id) {
        await uploadAdminGalleryItemImage(csrfToken, savedItem.id, selectedGalleryImage);
        setNotice(editingGalleryItemId ? 'Элемент и изображение обновлены' : 'Элемент создан с изображением');
      }

      setGalleryForm(emptyGalleryItem);
      setSelectedGalleryImage(null);
      setEditingGalleryItemId(null);
      await loadData();
    } catch (submitError) {
      setError(getErrorMessage(submitError));
    }
  };

  const submitTestimonial = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    try {
      const payload = {
        ...testimonialForm,
        name: testimonialForm.name.trim(),
        role: testimonialForm.role.trim(),
        img: testimonialForm.img.trim(),
        text: testimonialForm.text.trim(),
      };

      let savedTestimonial: Testimonial;
      if (editingTestimonialId) {
        savedTestimonial = await updateAdminTestimonial(csrfToken, { ...payload, id: editingTestimonialId });
        setNotice('Отзыв обновлен');
      } else {
        savedTestimonial = await createAdminTestimonial(csrfToken, payload);
        setNotice('Отзыв добавлен');
      }

      if (selectedTestimonialImage && savedTestimonial.id) {
        await uploadAdminTestimonialImage(csrfToken, savedTestimonial.id, selectedTestimonialImage);
        setNotice(editingTestimonialId ? 'Отзыв и фото обновлены' : 'Отзыв добавлен с фото');
      }

      setTestimonialForm(emptyTestimonial);
      setSelectedTestimonialImage(null);
      setEditingTestimonialId(null);
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
    setSelectedProductImage(null);
    setSelectedProductImages([]);
    setSelectedProductFiles([]);
  };

  const editBlogPost = (post: BlogPost) => {
    setEditingBlogPostId(post.id);
    setBlogForm({ ...post, img: post.img || '' });
    setSelectedBlogImage(null);
  };

  const editGalleryItem = (item: GalleryItem) => {
    if (!item.id) return;
    setEditingGalleryItemId(item.id);
    setGalleryForm({ ...item, img: item.img || '' });
    setSelectedGalleryImage(null);
  };

  const editTestimonial = (testimonial: Testimonial) => {
    if (!testimonial.id) return;
    setEditingTestimonialId(testimonial.id);
    setTestimonialForm(testimonial);
    setSelectedTestimonialImage(null);
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

  const removeProductImage = async (productId: string, imageId: number) => {
    if (!window.confirm('Удалить дополнительное изображение товара?')) return;
    try {
      await deleteAdminProductImage(csrfToken, productId, imageId);
      setNotice('Изображение товара удалено');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  const removeProductFile = async (productId: string, fileId: number) => {
    if (!window.confirm('Удалить файл товара? Покупатели больше не смогут его скачать.')) return;
    try {
      await deleteAdminProductFile(csrfToken, productId, fileId);
      setNotice('Файл товара удален');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  const removeBlogPost = async (post: BlogPost) => {
    if (!window.confirm(`Удалить запись ${post.title}?`)) return;
    try {
      await deleteAdminBlogPost(csrfToken, post.id);
      setNotice('Запись блога удалена');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  const removeGalleryItem = async (item: GalleryItem) => {
    if (!item.id || !window.confirm(`Удалить элемент ${item.title}?`)) return;
    try {
      await deleteAdminGalleryItem(csrfToken, item.id);
      setNotice('Элемент галереи удален');
      await loadData();
    } catch (deleteError) {
      setError(getErrorMessage(deleteError));
    }
  };

  const removeTestimonial = async (testimonial: Testimonial) => {
    if (!testimonial.id || !window.confirm(`Удалить отзыв ${testimonial.name}?`)) return;
    try {
      await deleteAdminTestimonial(csrfToken, testimonial.id);
      setNotice('Отзыв удален');
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
        <div className="admin-nav-group">
          <div className="admin-nav-label">Информация</div>
          <button className={'admin-nav-item' + (tab === 'orders' ? ' active' : '')} onClick={() => setTab('orders')}>
            Заказы
          </button>
        </div>
        <div className="admin-nav-group">
          <div className="admin-nav-label">Настройки</div>
          <button className={'admin-nav-item' + (tab === 'products' ? ' active' : '')} onClick={() => setTab('products')}>
            Товары
          </button>
          <button className={'admin-nav-item' + (tab === 'categories' ? ' active' : '')} onClick={() => setTab('categories')}>
            Категории
          </button>
          <button className={'admin-nav-item' + (tab === 'blog' ? ' active' : '')} onClick={() => setTab('blog')}>
            Блог
          </button>
          <button className={'admin-nav-item' + (tab === 'gallery' ? ' active' : '')} onClick={() => setTab('gallery')}>
            Галерея
          </button>
          <button className={'admin-nav-item' + (tab === 'settings' ? ' active' : '')} onClick={() => setTab('settings')}>
            Главная
          </button>
        </div>
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
              setSelectedProductImage(null);
              setSelectedProductImages([]);
              setSelectedProductFiles([]);
            }}
            onChange={setProductForm}
            onEdit={editProduct}
            onAdditionalImagesChange={setSelectedProductImages}
            onImageChange={setSelectedProductImage}
            onFilesChange={setSelectedProductFiles}
            onRemoveFile={removeProductFile}
            onRemoveImage={removeProductImage}
            onRemove={removeProduct}
            onSubmit={submitProduct}
            products={products}
            selectedAdditionalImages={selectedProductImages}
            selectedFiles={selectedProductFiles}
            selectedImage={selectedProductImage}
          />
        ) : tab === 'categories' ? (
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
        ) : tab === 'blog' ? (
          <BlogAdmin
            csrfToken={csrfToken}
            editingPostId={editingBlogPostId}
            form={blogForm}
            onCancel={() => {
              setEditingBlogPostId(null);
              setBlogForm(emptyBlogPost);
              setSelectedBlogImage(null);
            }}
            onChange={setBlogForm}
            onEdit={editBlogPost}
            onImageChange={setSelectedBlogImage}
            onEditorError={setError}
            onRemove={removeBlogPost}
            onSubmit={submitBlogPost}
            posts={blogPosts}
            selectedImage={selectedBlogImage}
          />
        ) : tab === 'gallery' ? (
          <GalleryAdmin
            editingItemId={editingGalleryItemId}
            form={galleryForm}
            items={galleryItems}
            onCancel={() => {
              setEditingGalleryItemId(null);
              setGalleryForm(emptyGalleryItem);
              setSelectedGalleryImage(null);
            }}
            onChange={setGalleryForm}
            onEdit={editGalleryItem}
            onImageChange={setSelectedGalleryImage}
            onRemove={removeGalleryItem}
            onSubmit={submitGalleryItem}
            selectedImage={selectedGalleryImage}
          />
        ) : tab === 'orders' ? (
          <OrdersAdmin orders={orders} />
        ) : (
          <SiteSettingsAdmin
            editingTestimonialId={editingTestimonialId}
            form={siteSettings}
            onCancelTestimonial={() => {
              setEditingTestimonialId(null);
              setTestimonialForm(emptyTestimonial);
              setSelectedTestimonialImage(null);
            }}
            onChange={setSiteSettings}
            onChangeTestimonial={setTestimonialForm}
            onChangeTestimonialImage={setSelectedTestimonialImage}
            onEditTestimonial={editTestimonial}
            onRemoveTestimonial={removeTestimonial}
            onSubmit={submitSiteSettings}
            onSubmitTestimonial={submitTestimonial}
            products={products}
            testimonialForm={testimonialForm}
            testimonialImage={selectedTestimonialImage}
            testimonials={testimonials}
          />
        )}
      </main>
    </div>
  );
}

function OrdersAdmin({ orders }: { orders: CustomerOrder[] }) {
  const total = (order: CustomerOrder) => order.items.reduce((sum, item) => sum + item.price * item.quantity, 0);

  return (
    <section className="admin-panel">
      <div className="admin-panel-head">
        <h2>Заказы</h2>
      </div>
      {orders.length === 0 ? (
        <div className="admin-empty">Заказов пока нет.</div>
      ) : (
        <div className="admin-table-wrap">
          <table className="admin-table admin-orders-table">
            <thead>
              <tr>
                <th>Заказ</th>
                <th>Покупатель</th>
                <th>Email</th>
                <th>Схемы</th>
                <th>Статус</th>
                <th>Сумма</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id}>
                  <td>
                    <code>{order.id}</code>
                    <div className="admin-muted">{formatAdminDate(order.createdAt)}</div>
                  </td>
                  <td>{order.customerName || '—'}</td>
                  <td>{order.customerEmail}</td>
                  <td>
                    <div className="admin-order-items">
                      {order.items.map((item) => (
                        <span key={item.productId}>
                          {item.productName || item.productId} × {item.quantity}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td><span className={'order-status status-' + order.status}>{orderStatusLabels[order.status] || order.status}</span></td>
                  <td>{formatPrice(total(order))}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function formatAdminDate(value: string) {
  if (!value) return '';
  return new Date(value).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

interface BlogAdminProps {
  csrfToken: string;
  editingPostId: string | null;
  form: BlogPost;
  onCancel: () => void;
  onChange: (post: BlogPost) => void;
  onEdit: (post: BlogPost) => void;
  onImageChange: (file: File | null) => void;
  onEditorError: (message: string) => void;
  onRemove: (post: BlogPost) => void;
  onSubmit: (event: React.FormEvent) => void;
  posts: BlogPost[];
  selectedImage: File | null;
}

function BlogAdmin({ csrfToken, editingPostId, form, onCancel, onChange, onEdit, onEditorError, onImageChange, onRemove, onSubmit, posts, selectedImage }: BlogAdminProps) {
  return (
    <div className="admin-grid">
      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Блог</h2>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>Дата</th><th>Заголовок</th><th>Тег</th><th></th></tr>
            </thead>
            <tbody>
              {posts.map((post) => (
                <tr key={post.id}>
                  <td>{post.date}</td>
                  <td>{post.title}</td>
                  <td>{post.tag}</td>
                  <td className="admin-row-actions">
                    <button onClick={() => onEdit(post)}>Изменить</button>
                    <button onClick={() => onRemove(post)}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>{editingPostId ? 'Редактирование записи' : 'Новая запись'}</h2>
        </div>
        <form className="admin-form product" onSubmit={onSubmit}>
          {editingPostId ? <div className="admin-current-file">ID: {form.id}</div> : null}
          <label>Заголовок<input value={form.title} onChange={(event) => onChange({ ...form, title: event.target.value })} /></label>
          <label>Дата<input type="date" value={form.date} onChange={(event) => onChange({ ...form, date: event.target.value })} /></label>
          <label>Тег<input value={form.tag} onChange={(event) => onChange({ ...form, tag: event.target.value })} /></label>
          <label className="admin-file-field">Обложка<input accept="image/*" type="file" onChange={(event) => onImageChange(event.target.files?.[0] || null)} /></label>
          {form.img ? <div className="admin-current-file">Текущая обложка: {form.img}</div> : null}
          {selectedImage ? <div className="admin-current-file">Новый файл: {selectedImage.name}</div> : null}
          <div className="admin-form-wide">
            <div className="admin-rich-label">Текст записи</div>
            <React.Suspense fallback={<div className="rich-editor-loading">Загружаем редактор...</div>}>
              <RichTextEditor
                csrfToken={csrfToken}
                value={form.content}
                onChange={(content) => onChange({ ...form, content })}
                onError={onEditorError}
              />
            </React.Suspense>
          </div>
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">{editingPostId ? 'Сохранить' : 'Создать'}</button>
            {editingPostId ? <button className="btn btn-outline" type="button" onClick={onCancel}>Отмена</button> : null}
          </div>
        </form>
      </section>
    </div>
  );
}

interface GalleryAdminProps {
  editingItemId: number | null;
  form: GalleryItem;
  items: GalleryItem[];
  onCancel: () => void;
  onChange: (item: GalleryItem) => void;
  onEdit: (item: GalleryItem) => void;
  onImageChange: (file: File | null) => void;
  onRemove: (item: GalleryItem) => void;
  onSubmit: (event: React.FormEvent) => void;
  selectedImage: File | null;
}

function GalleryAdmin({ editingItemId, form, items, onCancel, onChange, onEdit, onImageChange, onRemove, onSubmit, selectedImage }: GalleryAdminProps) {
  return (
    <div className="admin-grid">
      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Галерея</h2>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>Название</th><th>Описание</th><th></th></tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id || item.title}>
                  <td>{item.title}</td>
                  <td>{item.description}</td>
                  <td className="admin-row-actions">
                    <button onClick={() => onEdit(item)}>Изменить</button>
                    <button onClick={() => onRemove(item)}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>{editingItemId ? 'Редактирование элемента' : 'Новый элемент'}</h2>
        </div>
        <form className="admin-form product" onSubmit={onSubmit}>
          {editingItemId ? <div className="admin-current-file">ID: {editingItemId}</div> : null}
          <label>Название<input value={form.title} onChange={(event) => onChange({ ...form, title: event.target.value })} /></label>
          <label>Описание<input value={form.description} onChange={(event) => onChange({ ...form, description: event.target.value })} /></label>
          <label className="admin-file-field">Изображение<input accept="image/*" type="file" onChange={(event) => onImageChange(event.target.files?.[0] || null)} /></label>
          {form.img ? <div className="admin-current-file">Текущее изображение: {form.img}</div> : null}
          {selectedImage ? <div className="admin-current-file">Новый файл: {selectedImage.name}</div> : null}
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">{editingItemId ? 'Сохранить' : 'Создать'}</button>
            {editingItemId ? <button className="btn btn-outline" type="button" onClick={onCancel}>Отмена</button> : null}
          </div>
        </form>
      </section>
    </div>
  );
}

interface SiteSettingsAdminProps {
  editingTestimonialId: number | null;
  form: SiteSettings;
  onCancelTestimonial: () => void;
  onChange: (settings: SiteSettings) => void;
  onChangeTestimonial: (testimonial: Testimonial) => void;
  onChangeTestimonialImage: (file: File | null) => void;
  onEditTestimonial: (testimonial: Testimonial) => void;
  onRemoveTestimonial: (testimonial: Testimonial) => void;
  onSubmit: (event: React.FormEvent) => void;
  onSubmitTestimonial: (event: React.FormEvent) => void;
  products: Product[];
  testimonialForm: Testimonial;
  testimonialImage: File | null;
  testimonials: Testimonial[];
}

function SiteSettingsAdmin({
  editingTestimonialId,
  form,
  onCancelTestimonial,
  onChange,
  onChangeTestimonial,
  onChangeTestimonialImage,
  onEditTestimonial,
  onRemoveTestimonial,
  onSubmit,
  onSubmitTestimonial,
  products,
  testimonialForm,
  testimonialImage,
  testimonials,
}: SiteSettingsAdminProps) {
  return (
    <div className="admin-settings-layout">
      <div className="admin-settings-column">
        <section className="admin-panel">
          <div className="admin-panel-head">
            <h2>{editingTestimonialId ? 'Редактирование отзыва' : 'Новый отзыв'}</h2>
          </div>
          <form className="admin-form product" onSubmit={onSubmitTestimonial}>
            <label>Имя<input value={testimonialForm.name} onChange={(event) => onChangeTestimonial({ ...testimonialForm, name: event.target.value })} /></label>
            <label>Роль<input value={testimonialForm.role} onChange={(event) => onChangeTestimonial({ ...testimonialForm, role: event.target.value })} /></label>
            <label className="admin-file-field">Фото<input accept="image/*" type="file" onChange={(event) => onChangeTestimonialImage(event.target.files?.[0] || null)} /></label>
            {testimonialForm.img ? <div className="admin-current-file">Текущее фото: {testimonialForm.img}</div> : null}
            {testimonialImage ? <div className="admin-current-file">Новый файл: {testimonialImage.name}</div> : null}
            <label>Текст<textarea value={testimonialForm.text} onChange={(event) => onChangeTestimonial({ ...testimonialForm, text: event.target.value })} /></label>
            <div className="admin-form-actions">
              <button className="btn btn-primary" type="submit">{editingTestimonialId ? 'Сохранить' : 'Добавить'}</button>
              {editingTestimonialId ? <button className="btn btn-outline" type="button" onClick={onCancelTestimonial}>Отмена</button> : null}
            </div>
          </form>
        </section>

        <section className="admin-panel">
          <div className="admin-panel-head">
            <h2>Отзывы</h2>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr><th>Имя</th><th>Роль</th><th>Текст</th><th></th></tr>
              </thead>
              <tbody>
                {testimonials.map((testimonial) => (
                  <tr key={testimonial.id || testimonial.name}>
                    <td>{testimonial.name}</td>
                    <td>{testimonial.role}</td>
                    <td>{testimonial.text}</td>
                    <td className="admin-row-actions">
                      <button onClick={() => onEditTestimonial(testimonial)}>Изменить</button>
                      <button onClick={() => onRemoveTestimonial(testimonial)}>Удалить</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Закрепленная схема</h2>
        </div>
        <form className="admin-form" onSubmit={onSubmit}>
          <label>Схема<select value={form.featuredProductId} onChange={(event) => onChange({ ...form, featuredProductId: event.target.value })}>
            <option value="">Не выбрана</option>
            {products.map((product) => (
              <option key={product.id} value={product.id}>{product.title}</option>
            ))}
          </select></label>
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">Сохранить</button>
          </div>
        </form>
      </section>
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
              <tr><th>Название</th><th></th></tr>
            </thead>
            <tbody>
              {categories.map((category) => (
                <tr key={category.id}>
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
          {editingCategoryId ? <div className="admin-current-file">ID: {form.id}</div> : null}
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
  onAdditionalImagesChange: (files: File[]) => void;
  onFilesChange: (files: File[]) => void;
  onImageChange: (file: File | null) => void;
  onRemoveImage: (productId: string, imageId: number) => void;
  onRemoveFile: (productId: string, fileId: number) => void;
  onRemove: (productId: string) => void;
  onSubmit: (event: React.FormEvent) => void;
  products: Product[];
  selectedAdditionalImages: File[];
  selectedFiles: File[];
  selectedImage: File | null;
}

function ProductsAdmin({
  categories,
  editingProductId,
  form,
  onAdditionalImagesChange,
  onCancel,
  onChange,
  onEdit,
  onFilesChange,
  onImageChange,
  onRemove,
  onRemoveImage,
  onRemoveFile,
  onSubmit,
  products,
  selectedAdditionalImages,
  selectedFiles,
  selectedImage,
}: ProductsAdminProps) {
  const categoryLabels = new Map(categories.map((category) => [category.id, category.label]));
  const additionalImageInputRef = React.useRef<HTMLInputElement>(null);
  const productFileInputRef = React.useRef<HTMLInputElement>(null);

  const addAdditionalImage = (file: File | null) => {
    if (!file) return;
    onAdditionalImagesChange([...selectedAdditionalImages, file]);
    if (additionalImageInputRef.current) {
      additionalImageInputRef.current.value = '';
    }
  };

  const removeSelectedAdditionalImage = (indexToRemove: number) => {
    onAdditionalImagesChange(selectedAdditionalImages.filter((_, index) => index !== indexToRemove));
  };

  const addProductFiles = (files: File[]) => {
    if (files.length === 0) return;
    onFilesChange([...selectedFiles, ...files]);
    if (productFileInputRef.current) {
      productFileInputRef.current.value = '';
    }
  };

  const removeSelectedFile = (indexToRemove: number) => {
    onFilesChange(selectedFiles.filter((_, index) => index !== indexToRemove));
  };

  return (
    <div className="admin-grid">
      <section className="admin-panel">
        <div className="admin-panel-head">
          <h2>Товары</h2>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>Название</th><th>Категория</th><th>Цена</th><th></th></tr>
            </thead>
            <tbody>
              {products.map((product) => (
                <tr key={product.id}>
                  <td>{product.title}</td>
                  <td>{categoryLabels.get(product.cat) || product.cat}</td>
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
          {editingProductId ? <div className="admin-current-file">ID: {form.id}</div> : null}
          <label>Название<input value={form.title} onChange={(event) => onChange({ ...form, title: event.target.value })} /></label>
          <label>Цена<input type="number" value={form.price} onChange={(event) => onChange({ ...form, price: Number(event.target.value) })} /></label>
          <label>Категория<select value={form.cat} onChange={(event) => onChange({ ...form, cat: event.target.value })}>
            <option value="">Выберите</option>
            {categories.filter((category) => category.id !== 'all').map((category) => (
              <option key={category.id} value={category.id}>{category.label}</option>
            ))}
          </select></label>
          <label className="admin-file-field">Изображение<input accept="image/*" type="file" onChange={(event) => onImageChange(event.target.files?.[0] || null)} /></label>
          {form.img ? <div className="admin-current-file">Текущее изображение: {form.img}</div> : null}
          {selectedImage ? <div className="admin-current-file">Новый файл: {selectedImage.name}</div> : null}
          <label className="admin-file-field admin-form-wide">Добавить дополнительное изображение<input ref={additionalImageInputRef} accept="image/*" type="file" onChange={(event) => addAdditionalImage(event.target.files?.[0] || null)} /></label>
          {selectedAdditionalImages.length > 0 ? (
            <div className="admin-selected-images admin-form-wide">
              <div className="admin-selected-images-title">К загрузке</div>
              {selectedAdditionalImages.map((file, index) => (
                <div className="admin-selected-image" key={`${file.name}-${file.size}-${file.lastModified}-${index}`}>
                  <span>{file.name}</span>
                  <button type="button" onClick={() => removeSelectedAdditionalImage(index)}>Убрать</button>
                </div>
              ))}
            </div>
          ) : null}
          {form.images && form.images.length > 0 ? (
            <div className="admin-product-images admin-form-wide">
              {form.images.map((image) => (
                <div className="admin-product-image" key={image.id}>
                  <img src={image.url} alt="" />
                  <button type="button" onClick={() => onRemoveImage(form.id, image.id)}>Удалить</button>
                </div>
              ))}
            </div>
          ) : null}
          <label className="admin-file-field admin-form-wide">Файлы для покупателя<input ref={productFileInputRef} multiple type="file" onChange={(event) => addProductFiles(Array.from(event.target.files || []))} /></label>
          {selectedFiles.length > 0 ? (
            <div className="admin-selected-images admin-form-wide">
              <div className="admin-selected-images-title">Файлы к загрузке</div>
              {selectedFiles.map((file, index) => (
                <div className="admin-selected-image" key={`${file.name}-${file.size}-${file.lastModified}-${index}`}>
                  <span>{file.name}</span>
                  <button type="button" onClick={() => removeSelectedFile(index)}>Убрать</button>
                </div>
              ))}
            </div>
          ) : null}
          {form.files && form.files.length > 0 ? (
            <div className="admin-product-files admin-form-wide">
              <div className="admin-selected-images-title">Загруженные файлы</div>
              {form.files.map((file) => (
                <div className="admin-product-file" key={file.id}>
                  <span>{file.name}</span>
                  <button type="button" onClick={() => onRemoveFile(form.id, file.id)}>Удалить</button>
                </div>
              ))}
            </div>
          ) : null}
          <label>Размер<input value={form.size} onChange={(event) => onChange({ ...form, size: event.target.value })} /></label>
          <label>Палитра<input value={form.colors} onChange={(event) => onChange({ ...form, colors: event.target.value })} /></label>
          <label className="admin-form-wide">Описание<textarea value={form.description || ''} onChange={(event) => onChange({ ...form, description: event.target.value })} /></label>
          <div className="admin-form-actions">
            <button className="btn btn-primary" type="submit">{editingProductId ? 'Сохранить' : 'Создать'}</button>
            {editingProductId ? <button className="btn btn-outline" type="button" onClick={onCancel}>Отмена</button> : null}
          </div>
        </form>
      </section>
    </div>
  );
}
