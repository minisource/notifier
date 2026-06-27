import { describe, it, expect } from 'vitest';
import { parseBooleanEnv } from '../notifier-config';

describe('parseBooleanEnv', () => {
  it('returns default for undefined', () => {
    expect(parseBooleanEnv(undefined)).toBe(false);
    expect(parseBooleanEnv(undefined, true)).toBe(true);
  });

  it('returns default for empty string', () => {
    expect(parseBooleanEnv('')).toBe(false);
    expect(parseBooleanEnv('', true)).toBe(true);
  });

  it('parses "false" as false', () => {
    expect(parseBooleanEnv('false')).toBe(false);
  });

  it('parses "0" as false', () => {
    expect(parseBooleanEnv('0')).toBe(false);
  });

  it('parses "no" as false', () => {
    expect(parseBooleanEnv('no')).toBe(false);
  });

  it('parses "off" as false', () => {
    expect(parseBooleanEnv('off')).toBe(false);
  });

  it('parses "true" as true', () => {
    expect(parseBooleanEnv('true')).toBe(true);
  });

  it('parses "1" as true', () => {
    expect(parseBooleanEnv('1')).toBe(true);
  });

  it('parses "yes" as true', () => {
    expect(parseBooleanEnv('yes')).toBe(true);
  });

  it('parses "on" as true', () => {
    expect(parseBooleanEnv('on')).toBe(true);
  });

  it('is case insensitive', () => {
    expect(parseBooleanEnv('TRUE')).toBe(true);
    expect(parseBooleanEnv('FALSE')).toBe(false);
    expect(parseBooleanEnv('Yes')).toBe(true);
    expect(parseBooleanEnv('No')).toBe(false);
  });

  it('trims whitespace', () => {
    expect(parseBooleanEnv(' true ')).toBe(true);
    expect(parseBooleanEnv(' false ')).toBe(false);
  });

  it('returns default for unknown values', () => {
    expect(parseBooleanEnv('maybe')).toBe(false);
    expect(parseBooleanEnv('maybe', true)).toBe(true);
  });
});
