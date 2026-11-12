/**
 * Auto-cleaning Wails v3 event wrapper for Svelte components.
 *
 * Wails `Events.On()` listeners stay in memory forever unless you manually
 * unsubscribe. When island views mount and unmount frequently, forgotten
 * listeners stack up and cause IPC lag.
 *
 * This wrapper hooks into Svelte's `onDestroy` lifecycle so listeners clean
 * themselves up automatically when a component unmounts.
 *
 * Usage:
 * Swap `import { Events } from "@wailsio/runtime"` with `import { Events } from "$lib/events"`
 * at the top level of your component `<script>` block.
 */

import { Events as WailsEvents } from '@wailsio/runtime'
import { onDestroy } from 'svelte'

export const Events = {
  ...WailsEvents,

  /**
   * Listens for a Wails event and unsubscribes when the component unmounts.
   * Call this at the top level of your `<script>` block, not inside async functions.
   */
  On(eventName: string, callback: (event: any) => void) {
    const unsub = WailsEvents.On(eventName, callback)
    onDestroy(() => unsub())
    return unsub
  },

  /**
   * Listens for a Wails event up to `maxCallbacks` times, or until the component unmounts.
   */
  OnMultiple(
    eventName: string,
    callback: (event: any) => void,
    maxCallbacks: number
  ) {
    const unsub = WailsEvents.OnMultiple(eventName, callback, maxCallbacks)
    onDestroy(() => unsub())
    return unsub
  },

  /**
   * Raw `Events.On` for global listeners in plain `.ts` files outside Svelte components.
   * You must handle unsubscribing manually when using this.
   */
  OnGlobal: WailsEvents.On
}
