/* eslint-disable @typescript-eslint/no-explicit-any */
/// <reference types="svelte" />
/// <reference types="vite/client" />

interface Window {
  go: any
  runtime: any
}

declare module '@appicon' {
  const src: string
  export default src
}

declare module '@appicon-svg' {
  const src: string
  export default src
}
