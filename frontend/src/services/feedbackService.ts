import { Call } from '@wailsio/runtime'

/**
 * System diagnostics information gathered from host environment.
 */
export interface SystemDiagnostics {
  appVersion: string
  gitCommit: string
  buildType: string
  osName: string
  osVersion: string
  windowsBuild: string
  architecture: string
  displayCount: number
  primaryScale: number
  webViewVersion: string
}

/**
 * Retrieves sanitized runtime environment metadata.
 *
 * @returns Diagnostic metadata object
 */
export async function getSystemDiagnostics(): Promise<SystemDiagnostics | null> {
  try {
    const res = await Call.ByName('Vinta/backend.App.GetSystemDiagnostics')
    return res as SystemDiagnostics
  } catch (err) {
    console.error('Failed to get system diagnostics:', err)
    return null
  }
}

/**
 * Launches the default web browser with a pre-populated GitHub issue template
 * containing sanitized runtime diagnostics.
 */
export async function reportFeedback(): Promise<void> {
  try {
    await Call.ByName('Vinta/backend.App.OpenFeedbackIssue')
  } catch (err) {
    console.error('Failed to open feedback issue:', err)
  }
}
