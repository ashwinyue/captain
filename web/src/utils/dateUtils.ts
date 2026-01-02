// Date utilities for parsing API timestamps and computing relative times
import i18n from '@/i18n';


/**
 * Parse API timestamp string as UTC (if timezone missing), return Date in local timezone
 * - Trims microseconds to milliseconds
 * - Adds 'Z' if timezone is missing
 */
export function parseAPITimestampToLocalDate(iso?: string | null): Date | null {
  if (!iso) return null;
  try {
    let s = iso.trim();
    // Normalize: replace space with T
    if (s.includes(' ')) s = s.replace(' ', 'T');
    // Keep only milliseconds (JS Date doesn't support microseconds)
    s = s.replace(/(\.\d{3})\d+$/, '$1');
    const hasTZ = /[zZ]|[+-]\d{2}:\d{2}$/.test(s);
    if (!hasTZ) s += 'Z'; // treat as UTC if TZ is missing
    const d = new Date(s);
    if (!Number.isFinite(d.getTime())) {
      // Fallback: manual UTC parse or more aggressive cleaning
      const cleanS = s.split('.')[0] + 'Z';
      const d2 = new Date(cleanS);
      if (Number.isFinite(d2.getTime())) return d2;

      const m = s.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d{1,3}))?/);
      if (m) {
        const [_, Y, M, D, h, m2, s2, ms] = m;
        return new Date(Date.UTC(+Y, +M - 1, +D, +h, +m2, +s2, +(ms || 0)));
      }
    }
    return d;
  } catch {
    return null;
  }
}

/**
 * Get positive minutes difference from now to the provided ISO timestamp.
 * Returns null if timestamp invalid.
 */
export function diffMinutesFromNow(iso?: string | null): number | null {
  const d = parseAPITimestampToLocalDate(iso);
  if (!d) return null;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  if (!Number.isFinite(diffMs)) return null;
  const minutes = Math.floor(Math.max(0, diffMs) / 60000);
  return minutes;
}

/**
 * Build last-seen text with special handling for <1 minute: "刚刚在线".
 * Otherwise, if within 60 minutes, show "X分钟前在线".
 */
export function buildLastSeenText(iso?: string | null, isOnline?: boolean | null): string | null {
  if (isOnline) return null;
  const mins = diffMinutesFromNow(iso);
  if (mins == null) return null;
  if (mins === 0) return i18n.t('time.lastSeen.justNow', { defaultValue: '刚刚在线' });
  if (mins > 0 && mins <= 60)   return i18n.t('time.lastSeen.minutesAgo', { mins, defaultValue: `${mins}分钟前在线` });
  return null;
}

/**
 * Format date as YYYY/MM/DD HH:mm
 */
export function formatLocalDateTime(iso?: string | null): string {
  const d = parseAPITimestampToLocalDate(iso);
  if (!d) return '-';
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  return `${y}/${m}/${day} ${hh}:${mm}`;
}

/**
 * Format online duration in minutes to a human-readable string
 * @param minutes Duration in minutes
 * @param isOnline Whether the visitor is currently online
 */
export function formatOnlineDuration(minutes: number | null | undefined, isOnline: boolean): string {
  if (minutes == null) {
    return '-';
  }

  if (isOnline && minutes === 0) {
    return i18n.t('visitor.onlineStatus.online', '在线');
  }

  if (minutes === 0) {
    return i18n.t('visitor.onlineStatus.justNow', '刚刚');
  }

  if (minutes < 60) {
    return i18n.t('visitor.onlineStatus.minutesAgo', { count: minutes, defaultValue: `${minutes} 分钟前` });
  }

  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return i18n.t('visitor.onlineStatus.hoursAgo', { count: hours, defaultValue: `${hours} 小时前` });
  }

  const days = Math.floor(hours / 24);
  if (days < 7) {
    return i18n.t('visitor.onlineStatus.daysAgo', { count: days, defaultValue: `${days} 天前` });
  }

  const weeks = Math.floor(days / 7);
  return i18n.t('visitor.onlineStatus.weeksAgo', { count: weeks, defaultValue: `${weeks} 周前` });
}

