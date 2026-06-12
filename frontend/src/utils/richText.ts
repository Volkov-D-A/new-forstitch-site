import type { JSONContent } from '@tiptap/core';

export function parseRichText(content: string): JSONContent {
  const trimmed = content.trim();
  if (trimmed.startsWith('{')) {
    try {
      const parsed = JSON.parse(trimmed) as JSONContent;
      if (parsed.type === 'doc') return parsed;
    } catch {
      // Legacy plain text is converted below.
    }
  }

  const paragraphs = content
    .split(/\n+/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean);

  return {
    type: 'doc',
    content: (paragraphs.length > 0 ? paragraphs : ['']).map((paragraph) => ({
      type: 'paragraph',
      content: paragraph ? [{ type: 'text', text: paragraph }] : undefined,
    })),
  };
}

export function serializeRichText(content: JSONContent): string {
  return JSON.stringify(content);
}
