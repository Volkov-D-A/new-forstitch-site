import { describe, expect, it } from 'vitest';

import { formatPrice } from './currency';

describe('formatPrice', () => {
  it('formats zero with the ruble sign', () => {
    expect(formatPrice(0)).toBe('0 ₽');
  });

  it('adds a localized thousands separator', () => {
    expect(formatPrice(1250)).toBe(`${(1250).toLocaleString('ru-RU')} ₽`);
  });
});
