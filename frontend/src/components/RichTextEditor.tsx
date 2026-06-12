import React from 'react';
import Image from '@tiptap/extension-image';
import Link from '@tiptap/extension-link';
import { TableKit } from '@tiptap/extension-table';
import { EditorContent, useEditor, useEditorState } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import {
  Bold,
  Columns3,
  Heading2,
  ImagePlus,
  Italic,
  Link2,
  List,
  ListOrdered,
  Minus,
  Pilcrow,
  Plus,
  Quote,
  Redo2,
  Rows3,
  Strikethrough,
  Table2,
  Trash2,
  Undo2,
  Unlink,
} from 'lucide-react';
import { uploadAdminBlogContentImage } from '../services/adminApi';
import { parseRichText, serializeRichText } from '../utils/richText';

const extensions = [
  StarterKit.configure({ link: false }),
  Link.configure({ openOnClick: false }),
  Image.configure({ allowBase64: false }),
  TableKit.configure({ table: { resizable: true } }),
];

interface RichTextEditorProps {
  csrfToken: string;
  value: string;
  onChange: (value: string) => void;
  onError: (message: string) => void;
}

interface ToolbarButtonProps {
  active?: boolean;
  disabled?: boolean;
  label: string;
  onClick: () => void;
  children: React.ReactNode;
}

function ToolbarButton({ active, children, disabled, label, onClick }: ToolbarButtonProps) {
  return (
    <button
      className={'rich-editor-button' + (active ? ' active' : '')}
      disabled={disabled}
      onClick={onClick}
      title={label}
      type="button"
      aria-label={label}
    >
      {children}
    </button>
  );
}

export function RichTextEditor({ csrfToken, value, onChange, onError }: RichTextEditorProps) {
  const imageInputRef = React.useRef<HTMLInputElement>(null);
  const [isUploading, setUploading] = React.useState(false);
  const editor = useEditor({
    extensions,
    content: parseRichText(value),
    onUpdate: ({ editor: currentEditor }) => onChange(serializeRichText(currentEditor.getJSON())),
  });
  const state = useEditorState({
    editor,
    selector: ({ editor: currentEditor }) => ({
      bold: currentEditor?.isActive('bold') || false,
      italic: currentEditor?.isActive('italic') || false,
      strike: currentEditor?.isActive('strike') || false,
      heading: currentEditor?.isActive('heading', { level: 2 }) || false,
      bulletList: currentEditor?.isActive('bulletList') || false,
      orderedList: currentEditor?.isActive('orderedList') || false,
      blockquote: currentEditor?.isActive('blockquote') || false,
      link: currentEditor?.isActive('link') || false,
      table: currentEditor?.isActive('table') || false,
    }),
  });

  React.useEffect(() => {
    if (!editor) return;
    const next = parseRichText(value);
    if (JSON.stringify(editor.getJSON()) !== JSON.stringify(next)) {
      editor.commands.setContent(next, { emitUpdate: false });
    }
  }, [editor, value]);

  if (!editor) return null;

  const setLink = () => {
    const previous = editor.getAttributes('link').href as string | undefined;
    const href = window.prompt('Адрес ссылки', previous || 'https://');
    if (href === null) return;
    if (!href.trim()) {
      editor.chain().focus().extendMarkRange('link').unsetLink().run();
      return;
    }
    editor.chain().focus().extendMarkRange('link').setLink({ href: href.trim() }).run();
  };

  const uploadImage = async (file: File | null) => {
    if (!file) return;
    setUploading(true);
    try {
      const uploaded = await uploadAdminBlogContentImage(csrfToken, file);
      editor.chain().focus().setImage({ src: uploaded.url, alt: file.name }).run();
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Не удалось загрузить изображение');
    } finally {
      setUploading(false);
      if (imageInputRef.current) imageInputRef.current.value = '';
    }
  };

  return (
    <div className="rich-editor">
      <div className="rich-editor-toolbar">
        <ToolbarButton label="Обычный текст" active={editor.isActive('paragraph')} onClick={() => editor.chain().focus().setParagraph().run()}><Pilcrow /></ToolbarButton>
        <ToolbarButton label="Заголовок" active={state?.heading} onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}><Heading2 /></ToolbarButton>
        <ToolbarButton label="Жирный" active={state?.bold} onClick={() => editor.chain().focus().toggleBold().run()}><Bold /></ToolbarButton>
        <ToolbarButton label="Курсив" active={state?.italic} onClick={() => editor.chain().focus().toggleItalic().run()}><Italic /></ToolbarButton>
        <ToolbarButton label="Зачеркнутый" active={state?.strike} onClick={() => editor.chain().focus().toggleStrike().run()}><Strikethrough /></ToolbarButton>
        <ToolbarButton label="Маркированный список" active={state?.bulletList} onClick={() => editor.chain().focus().toggleBulletList().run()}><List /></ToolbarButton>
        <ToolbarButton label="Нумерованный список" active={state?.orderedList} onClick={() => editor.chain().focus().toggleOrderedList().run()}><ListOrdered /></ToolbarButton>
        <ToolbarButton label="Цитата" active={state?.blockquote} onClick={() => editor.chain().focus().toggleBlockquote().run()}><Quote /></ToolbarButton>
        <ToolbarButton label="Ссылка" active={state?.link} onClick={setLink}><Link2 /></ToolbarButton>
        <ToolbarButton label="Удалить ссылку" disabled={!state?.link} onClick={() => editor.chain().focus().unsetLink().run()}><Unlink /></ToolbarButton>
        <ToolbarButton label="Изображение" disabled={isUploading} onClick={() => imageInputRef.current?.click()}><ImagePlus /></ToolbarButton>
        <ToolbarButton label="Добавить таблицу" onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}><Table2 /></ToolbarButton>
        {state?.table ? (
          <>
            <ToolbarButton label="Добавить строку" onClick={() => editor.chain().focus().addRowAfter().run()}><Rows3 /><Plus /></ToolbarButton>
            <ToolbarButton label="Удалить строку" onClick={() => editor.chain().focus().deleteRow().run()}><Rows3 /><Minus /></ToolbarButton>
            <ToolbarButton label="Добавить столбец" onClick={() => editor.chain().focus().addColumnAfter().run()}><Columns3 /><Plus /></ToolbarButton>
            <ToolbarButton label="Удалить столбец" onClick={() => editor.chain().focus().deleteColumn().run()}><Columns3 /><Minus /></ToolbarButton>
            <ToolbarButton label="Удалить таблицу" onClick={() => editor.chain().focus().deleteTable().run()}><Trash2 /></ToolbarButton>
          </>
        ) : null}
        <ToolbarButton label="Отменить" disabled={!editor.can().undo()} onClick={() => editor.chain().focus().undo().run()}><Undo2 /></ToolbarButton>
        <ToolbarButton label="Повторить" disabled={!editor.can().redo()} onClick={() => editor.chain().focus().redo().run()}><Redo2 /></ToolbarButton>
      </div>
      <EditorContent editor={editor} />
      <input ref={imageInputRef} accept="image/*" hidden type="file" onChange={(event) => uploadImage(event.target.files?.[0] || null)} />
    </div>
  );
}

export function RichTextContent({ value }: { value: string }) {
  const editor = useEditor({
    extensions,
    content: parseRichText(value),
    editable: false,
  });

  React.useEffect(() => {
    if (editor) editor.commands.setContent(parseRichText(value), { emitUpdate: false });
  }, [editor, value]);

  return <EditorContent className="rich-content" editor={editor} />;
}
