/**
 * @fileoverview Artwork URL passthrough for album art.
 */

interface UpscaleResult {
  url: string
  isUpscaled: boolean
  width: number
  height: number
}

/**
 * Passes artwork URLs through directly so WebView2's GPU compositor handles scaling
 * without blocking the main thread on synchronous canvas resampling.
 */
export async function enhanceArtwork(
  src: string,
  _minDisplayWidth = 400
): Promise<UpscaleResult> {
  if (!src) {
    return { url: '', isUpscaled: false, width: 0, height: 0 }
  }
  return { url: src, isUpscaled: false, width: 0, height: 0 }
}
