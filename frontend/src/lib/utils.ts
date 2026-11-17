import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return "0 o"
  const units = ["o", "Ko", "Mo", "Go"]
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value >= 10 || unit === 0 ? Math.round(value) : value.toFixed(1)} ${units[unit]}`
}

export function clock(ts: number): string {
  if (!ts) return "--:--:--"
  return new Date(ts).toLocaleTimeString("fr-FR", { hour12: false })
}
