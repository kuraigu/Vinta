import { writable } from 'svelte/store'

export const activeSettingsTabStore = writable<string>('appearance')
export const activePageIdStore = writable<string>('home')
export const hiddenPillOffsetXStore = writable<number>(0)

export const simulateOfflineStore = writable<boolean>(false)
export const isSimulatingBatteryStore = writable<boolean>(false)
export const isSimulatingWirelessStore = writable<boolean>(false)
