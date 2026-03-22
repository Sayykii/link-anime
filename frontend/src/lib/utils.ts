import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return bytes + ' B'
}

// Build a proxied Shoko image URL. Images are served through our backend.
export function shokoImageUrl(source: string, type: string, id: string): string {
  return `/api/shoko/image/${source}/${type}/${id}`
}

// Get the preferred poster URL for a Shoko series, or null if none.
export function seriesPosterUrl(images: { Posters?: { ID: string; Source: string; Preferred: boolean }[] }): string | null {
  if (!images.Posters?.length) return null
  const preferred = images.Posters.find(p => p.Preferred) || images.Posters[0]
  return shokoImageUrl(preferred.Source, 'Poster', preferred.ID)
}

// Get poster URL from a dashboard episode's SeriesPoster field.
export function dashboardPosterUrl(poster?: { ID: string; Source: string } | null): string | null {
  if (!poster) return null
  return shokoImageUrl(poster.Source, 'Poster', poster.ID)
}
