import { DateRange } from "@yca-software/design-system";

export function getDefaultLast7DaysRange(): DateRange {
  const today = new Date();
  const end = new Date(today);
  end.setHours(0, 0, 0, 0);
  const start = new Date(end);
  start.setDate(end.getDate() - 6);
  return { from: start, to: end };
}

export function toStartOfDay(date: Date | undefined): Date | undefined {
  if (!date) return undefined;
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

export function toEndOfDay(date: Date | undefined): Date | undefined {
  if (!date) return undefined;
  const d = new Date(date);
  d.setHours(23, 59, 59, 999);
  return d;
}
