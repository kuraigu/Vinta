import { writable } from 'svelte/store'
import type { WeatherData } from '../services/weatherService'

export const activeSettingsTabStore = writable<string>('appearance')
export const activePageIdStore = writable<string>('home')
export const hiddenPillOffsetXStore = writable<number>(0)
export const isIslandHiddenStore = writable<boolean>(false)

export const simulateOfflineStore = writable<boolean>(false)
export const simulateWeatherStore = writable<WeatherData | null>(null)
export const isSimulatingBatteryStore = writable<boolean>(false)
export const isSimulatingWirelessStore = writable<boolean>(false)
