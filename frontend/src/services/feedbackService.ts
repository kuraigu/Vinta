/**
 * @fileoverview Gathers system info for bug reports and opens GitHub issues.
 *
 * Collects Windows build, architecture, and display scale details so users can
 * file pre-filled bug reports on GitHub without having to copy system info manually.
 */

import {
  GetSystemDiagnostics,
  OpenFeedbackIssue,
  OpenFeatureRequest
} from '../../bindings/Vinta/backend/app'

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
    const res = await GetSystemDiagnostics()
    return res as unknown as SystemDiagnostics
  } catch (err) {
    console.error('Failed to get system diagnostics:', err)
    return null
  }
}

/**
 * Launches the default web browser with the consolidated bug report template
 * pre-populated with user input and sanitized runtime diagnostics.
 */
export async function reportFeedback(
  title: string = '',
  describe: string = '',
  reproduce: string = ''
): Promise<void> {
  try {
    await OpenFeedbackIssue(title, describe, reproduce)
  } catch (err) {
    console.error('Failed to open feedback issue:', err)
  }
}

/**
 * Launches the default web browser with the feature request template pre-populated.
 */
export async function requestFeature(
  title: string = '',
  describe: string = '',
  useCase: string = ''
): Promise<void> {
  try {
    await OpenFeatureRequest(title, describe, useCase)
  } catch (err) {
    console.error('Failed to open feature request:', err)
  }
}
