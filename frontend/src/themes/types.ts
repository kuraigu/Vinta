/**
 * @fileoverview Type definitions for theme packages.
 *
 * Defines the JSON schema for `.vintatheme` manifest files (name, author, entry CSS).
 */

/**
 * VintaThemeManifest represents the schema definition of a .vintatheme manifest.json.
 */
export interface VintaThemeManifest {
  schemaVersion: number
  id: string
  name: string
  author: string
  version: string
  description?: string
  entry: string
  preview?: string
  accentColor?: string
  mode?: 'dark' | 'light' | 'oled' | 'custom'
  meta?: Record<string, unknown>
}

/**
 * ThemeInfo encapsulates manifest data and filesystem origin for frontend consumers.
 */
export interface ThemeInfo {
  manifest: VintaThemeManifest
  isBuiltIn: boolean
  path: string
}
