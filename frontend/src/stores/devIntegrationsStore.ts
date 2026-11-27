/**
 * @fileoverview Reactive store for Google/Firebase Cloud Sync, GitHub Developer Hub,
 * and DPAPI-backed AI provider telemetry.
 */
import { writable } from 'svelte/store'
import {
  GetIntegrationCapabilities,
  GetGoogleAccountStatus,
  SignInWithGoogle,
  SignOutGoogle,
  PushCloudSync,
  PullCloudSync,
  GetGitHubStatus,
  ConnectGitHubWithPAT,
  StartGitHubDeviceFlow,
  PollGitHubDeviceFlow,
  DisconnectGitHub,
  GetGitHubDashboard,
  GetAIKeysStatus,
  SaveAIProviderKey,
  DeleteAIProviderKey,
  GetAIUsageOverview
} from '../../bindings/Vinta/backend/app.js'
import type {
  IntegrationCapabilities,
  GoogleAccountStatus,
  CloudSyncResult,
  GitHubStatus,
  GitHubDeviceCodeInfo,
  GitHubDashboard,
  AIKeysStatus,
  AIProviderUsage
} from '../../bindings/Vinta/backend/models.js'
import {
  refreshGoogleCalendarStatus,
  syncGoogleCalendar
} from './calendarStore'
import { beginAuthSignIn, endAuthSignIn } from './ui'

export const integrationCapabilitiesStore = writable<IntegrationCapabilities>({
  hasGoogleOAuth: false,
  hasFirebase: false,
  hasGitHubOAuth: false
})

export const googleAccountStore = writable<GoogleAccountStatus>({
  connected: false,
  email: '',
  name: '',
  picture: '',
  hasGoogleClientId: false,
  firebaseEnabled: false,
  cloudSyncReady: false,
  lastCloudSyncAt: 0
})

export const githubStatusStore = writable<GitHubStatus>({
  connected: false,
  username: '',
  avatarUrl: '',
  hasOAuthClientId: false
})

export const githubDashboardStore = writable<GitHubDashboard | null>(null)

export const aiKeysStatusStore = writable<AIKeysStatus>({
  hasClaude: false,
  hasOpenAI: false
})

export const aiUsageStore = writable<AIProviderUsage[]>([])

function formatError(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message
  if (typeof err === 'string' && err.trim()) return err
  return fallback
}

/**
 * Loads initial account, GitHub, and AI key states from the Go DPAPI vault.
 */
export async function initDevIntegrations(): Promise<void> {
  try {
    const [caps, google, github, aiKeys] = await Promise.all([
      typeof GetIntegrationCapabilities === 'function'
        ? GetIntegrationCapabilities()
        : Promise.resolve(null),
      typeof GetGoogleAccountStatus === 'function'
        ? GetGoogleAccountStatus()
        : Promise.resolve(null),
      typeof GetGitHubStatus === 'function'
        ? GetGitHubStatus()
        : Promise.resolve(null),
      typeof GetAIKeysStatus === 'function'
        ? GetAIKeysStatus()
        : Promise.resolve(null)
    ])
    if (caps) integrationCapabilitiesStore.set(caps)
    if (google) googleAccountStore.set(google)
    if (github) {
      githubStatusStore.set(github)
      if (github.connected) {
        void refreshGitHubDashboard(false)
      }
    }
    if (aiKeys) aiKeysStatusStore.set(aiKeys)
    void refreshAIUsage(false)
  } catch {
    // Non-fatal in browser-only preview mode
  }
}

export async function signInGoogleAccount(): Promise<{
  ok: boolean
  error?: string
}> {
  if (typeof SignInWithGoogle !== 'function') {
    return { ok: false, error: 'Google sign-in unavailable in preview mode' }
  }
  beginAuthSignIn()
  try {
    const status = await SignInWithGoogle()
    if (status) googleAccountStore.set(status)
    await refreshGoogleCalendarStatus()
    if (status?.connected) {
      void syncGoogleCalendar()
    }
    return { ok: true }
  } catch (err) {
    return { ok: false, error: formatError(err, 'Google sign-in failed') }
  } finally {
    endAuthSignIn()
  }
}

export async function signOutGoogleAccount(): Promise<void> {
  if (typeof SignOutGoogle === 'function') {
    await SignOutGoogle()
  }
  const status =
    typeof GetGoogleAccountStatus === 'function'
      ? await GetGoogleAccountStatus()
      : null
  if (status) googleAccountStore.set(status)
  await refreshGoogleCalendarStatus()
}

export async function pushCloudBackup(): Promise<{
  ok: boolean
  result?: CloudSyncResult
  error?: string
}> {
  if (typeof PushCloudSync !== 'function') {
    return { ok: false, error: 'Cloud sync unavailable' }
  }
  try {
    const result = await PushCloudSync()
    const status = await GetGoogleAccountStatus()
    if (status) googleAccountStore.set(status)
    return { ok: true, result }
  } catch (err) {
    return { ok: false, error: formatError(err, 'Cloud backup failed') }
  }
}

export async function pullCloudBackup(): Promise<{
  ok: boolean
  result?: CloudSyncResult
  error?: string
}> {
  if (typeof PullCloudSync !== 'function') {
    return { ok: false, error: 'Cloud sync unavailable' }
  }
  try {
    const result = await PullCloudSync()
    const status = await GetGoogleAccountStatus()
    if (status) googleAccountStore.set(status)
    return { ok: true, result }
  } catch (err) {
    return { ok: false, error: formatError(err, 'Cloud restore failed') }
  }
}

export async function connectGitHubPAT(
  token: string
): Promise<{ ok: boolean; error?: string }> {
  if (typeof ConnectGitHubWithPAT !== 'function') {
    return { ok: false, error: 'GitHub integration unavailable' }
  }
  try {
    const status = await ConnectGitHubWithPAT(token)
    if (status) githubStatusStore.set(status)
    await refreshGitHubDashboard(true)
    return { ok: true }
  } catch (err) {
    return { ok: false, error: formatError(err, 'GitHub connection failed') }
  }
}

export interface GitHubAuthScopeOptions {
  includeNotifications?: boolean
  includePrivateRepos?: boolean
}

let githubAuthPinHeld = false

export async function startGitHubDeviceAuth(
  options: GitHubAuthScopeOptions = {}
): Promise<{
  ok: boolean
  info?: GitHubDeviceCodeInfo
  error?: string
}> {
  if (typeof StartGitHubDeviceFlow !== 'function') {
    return { ok: false, error: 'GitHub Device Flow unavailable' }
  }
  if (!githubAuthPinHeld) {
    githubAuthPinHeld = true
    beginAuthSignIn()
  }
  try {
    const info = await StartGitHubDeviceFlow(
      Boolean(options.includeNotifications),
      Boolean(options.includePrivateRepos)
    )
    return { ok: true, info }
  } catch (err) {
    if (githubAuthPinHeld) {
      githubAuthPinHeld = false
      endAuthSignIn()
    }
    return {
      ok: false,
      error: formatError(err, 'Could not start GitHub sign-in')
    }
  }
}

export async function completeGitHubDeviceAuth(
  deviceCode: string,
  interval: number
): Promise<{ ok: boolean; error?: string }> {
  if (typeof PollGitHubDeviceFlow !== 'function') {
    if (githubAuthPinHeld) {
      githubAuthPinHeld = false
      endAuthSignIn()
    }
    return { ok: false, error: 'GitHub Device Flow unavailable' }
  }
  try {
    const status = await PollGitHubDeviceFlow(deviceCode, interval)
    if (status) githubStatusStore.set(status)
    if (status?.connected) {
      await refreshGitHubDashboard(true)
    }
    return { ok: true }
  } catch (err) {
    return {
      ok: false,
      error: formatError(err, 'GitHub sign-in did not complete')
    }
  } finally {
    if (githubAuthPinHeld) {
      githubAuthPinHeld = false
      endAuthSignIn()
    }
  }
}

export async function disconnectGitHubAccount(): Promise<void> {
  if (typeof DisconnectGitHub === 'function') {
    await DisconnectGitHub()
  }
  githubDashboardStore.set(null)
  const status =
    typeof GetGitHubStatus === 'function' ? await GetGitHubStatus() : null
  if (status) githubStatusStore.set(status)
}

export async function refreshGitHubDashboard(
  forceRefresh = false
): Promise<void> {
  if (typeof GetGitHubDashboard !== 'function') return
  try {
    const dash = await GetGitHubDashboard(forceRefresh)
    if (dash) githubDashboardStore.set(dash)
  } catch {
    // Keep existing cached view on transient offline errors
  }
}

export async function saveAIKey(
  provider: 'claude' | 'openai',
  apiKey: string
): Promise<{ ok: boolean; error?: string }> {
  if (typeof SaveAIProviderKey !== 'function') {
    return { ok: false, error: 'Credential vault unavailable' }
  }
  try {
    await SaveAIProviderKey(provider, apiKey)
    const keys = await GetAIKeysStatus()
    if (keys) aiKeysStatusStore.set(keys)
    await refreshAIUsage(true)
    return { ok: true }
  } catch (err) {
    return { ok: false, error: formatError(err, 'Failed to save API key') }
  }
}

export async function deleteAIKey(
  provider: 'claude' | 'openai'
): Promise<void> {
  if (typeof DeleteAIProviderKey === 'function') {
    await DeleteAIProviderKey(provider)
  }
  const keys =
    typeof GetAIKeysStatus === 'function' ? await GetAIKeysStatus() : null
  if (keys) aiKeysStatusStore.set(keys)
  await refreshAIUsage(true)
}

export async function refreshAIUsage(forceRefresh = false): Promise<void> {
  if (typeof GetAIUsageOverview !== 'function') return
  try {
    const list = await GetAIUsageOverview(forceRefresh)
    if (Array.isArray(list)) aiUsageStore.set(list)
  } catch {
    // ignore
  }
}
