import { describe, expect, it } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

import en from './en.json';
import fa from './fa.json';

/**
 * Flatten a nested messages object into dot-notation keys, e.g.
 * { common: { active: 'Active' } } -> ['common.active'].
 */
function flattenKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  const keys: string[] = [];
  for (const [key, value] of Object.entries(obj)) {
    const full = prefix ? `${prefix}.${key}` : key;
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      keys.push(...flattenKeys(value as Record<string, unknown>, full));
    } else {
      keys.push(full);
    }
  }
  return keys;
}

function hasKey(obj: Record<string, unknown>, dotted: string): boolean {
  return dotted
    .split('.')
    .every((part) => {
      if (obj && typeof obj === 'object' && part in obj) {
        obj = obj[part] as Record<string, unknown>;
        return true;
      }
      return false;
    });
}

/**
 * Collect literal translation keys passed to `t('...')` / `t("...")` across the
 * source tree. Dynamic template-literal keys (e.g. t(`statuses.${s}`)) are
 * resolved at runtime and intentionally skipped here.
 */
function collectSourceKeys(): string[] {
  const root = path.join(__dirname, '..');
  const keys = new Set<string>();
  const walk = (dir: string) => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        // Skip generated/build dirs and the message catalogs themselves —
        // only static app source matters here.
        if (['node_modules', '.next', 'messages'].includes(entry.name)) continue;
        walk(full);
      } else if (
        /\.(ts|tsx)$/.test(entry.name) &&
        !/\.(test|spec)\.(ts|tsx)$/.test(entry.name) &&
        entry.name !== 'i18n.ts' // config file — holds t() examples in comments
      ) {
        const source = fs.readFileSync(full, 'utf8');
        for (const match of source.matchAll(/\bt\(\s*(['"])([^'"`]+)\1/g)) {
          keys.add(match[2]);
        }
      }
    }
  };
  walk(root);
  return [...keys];
}

describe('locale message files', () => {
  const enKeys = flattenKeys(en).sort();
  const faKeys = flattenKeys(fa).sort();

  it('en.json and fa.json expose the same key set', () => {
    expect(faKeys).toEqual(enKeys);
  });

  it('every literal translation key used in src exists in both locales', () => {
    const missing = collectSourceKeys().filter(
      (key) => !hasKey(en, key) || !hasKey(fa, key),
    );
    expect(missing).toEqual([]);
  });
});
