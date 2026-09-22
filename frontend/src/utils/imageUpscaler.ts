/**
 * Multi-stage pyramid upsampler for low-resolution media artwork.
 *
 * Uses fetch + blob URL to bypass cross-origin restrictions that block canvas
 * operations on remote thumbnail URLs (e.g. YouTube's i.ytimg.com CDN).
 *
 * Only processes images that are genuinely smaller than the display card.
 * High-res sources (e.g. Spotify 640×640) are passed through untouched and
 * left to CSS bilinear downscaling, which looks clean at sub-display-size ratios.
 */

interface UpscaleResult {
  url: string
  isUpscaled: boolean
  width: number
  height: number
}

/** In-memory LRU cache to prevent redundant re-processing. */
const artCache = new Map<string, string>()

/**
 * Loads an image from a URL via fetch (as a Blob) to sidestep canvas CORS tainting.
 *
 * Wails 3's WebView2 does not enforce same-origin policy for fetch calls the way
 * a browser does, so plain fetch generally succeeds for external CDN URLs.
 *
 * @param src Remote image URL
 * @returns A local blob: URL, or null if both fetch attempts fail
 */
async function fetchAsBlobUrl(src: string): Promise<string | null> {
  // Attempt 1: plain fetch — works in Wails WebView2 without CORS enforcement
  try {
    const resp = await fetch(src)
    if (resp.ok) {
      const blob = await resp.blob()
      return URL.createObjectURL(blob)
    }
  } catch {
    // fall through to attempt 2
  }

  // Attempt 2: explicit CORS mode
  try {
    const resp = await fetch(src, { mode: 'cors', credentials: 'omit' })
    if (!resp.ok) return null
    const blob = await resp.blob()
    return URL.createObjectURL(blob)
  } catch {
    return null
  }
}

/**
 * Draws srcCanvas into a new canvas at targetW × targetH using the browser's
 * built-in high-quality bilinear resampler. No pixel-level sharpening is applied
 * to avoid amplifying JPEG compression noise into visible color speckles.
 */
function bilinearUpscaleStage(
  srcCanvas: HTMLCanvasElement,
  targetW: number,
  targetH: number
): HTMLCanvasElement {
  const dst = document.createElement('canvas')
  dst.width = targetW
  dst.height = targetH
  const ctx = dst.getContext('2d')!
  ctx.imageSmoothingEnabled = true
  ctx.imageSmoothingQuality = 'high'
  ctx.drawImage(srcCanvas, 0, 0, targetW, targetH)
  return dst
}

/**
 * Upscales low-resolution media artwork using a 2-stage bilinear pyramid.
 *
 * Only images whose width is below `minDisplayWidth × 1.4` are processed.
 * Anything larger is passed through as-is; CSS handles the rest via object-fit.
 *
 * @param src Source image URI or base64 data URL
 * @param minDisplayWidth Width of the display card in CSS pixels (default 400)
 * @returns Promise resolving to the enhanced URL and metadata
 */
export async function enhanceArtwork(
  src: string,
  minDisplayWidth = 400
): Promise<UpscaleResult> {
  if (!src) {
    return { url: '', isUpscaled: false, width: 0, height: 0 }
  }

  const cached = artCache.get(src)
  if (cached) {
    return { url: cached, isUpscaled: true, width: minDisplayWidth, height: minDisplayWidth }
  }

  // Fetch as blob to avoid CORS canvas tainting on remote URLs
  const isDataUrl = src.startsWith('data:')
  const blobUrl = isDataUrl ? null : await fetchAsBlobUrl(src)

  // If fetch failed and this is not an inline data URL, bail out gracefully
  const processUrl = blobUrl ?? (isDataUrl ? src : null)
  if (!processUrl) {
    return { url: src, isUpscaled: false, width: 0, height: 0 }
  }

  return new Promise((resolve) => {
    const img = new Image()

    img.onload = () => {
      const srcW = img.naturalWidth
      const srcH = img.naturalHeight

      // Revoke temporary blob URL now that the image is decoded
      if (blobUrl) URL.revokeObjectURL(blobUrl)

      // Images already adequate for the display card are passed through untouched.
      // For a 400px card, anything ≥ 560px (1.4×) is more than sufficient for
      // CSS bilinear downscaling to look sharp — no canvas work needed.
      const adequateWidth = Math.round(minDisplayWidth * 1.4)
      if (srcW >= adequateWidth) {
        resolve({ url: src, isUpscaled: false, width: srcW, height: srcH })
        return
      }

      try {
        // Seed canvas at native resolution
        const seedCanvas = document.createElement('canvas')
        seedCanvas.width = srcW
        seedCanvas.height = srcH
        const seedCtx = seedCanvas.getContext('2d')!
        seedCtx.imageSmoothingEnabled = true
        seedCtx.imageSmoothingQuality = 'high'
        seedCtx.drawImage(img, 0, 0, srcW, srcH)

        let currentCanvas = seedCanvas

        // Stage 1: gentle 1.5× intermediate step for very small images (< 300px)
        // to avoid a single-jump stretch that exaggerates blockiness.
        if (srcW < 300) {
          const step1W = Math.round(srcW * 1.5)
          const step1H = Math.round(srcH * 1.5)
          currentCanvas = bilinearUpscaleStage(currentCanvas, step1W, step1H)
        }

        // Stage 2: upsample to 2× source width, capped at 800px
        const targetW = Math.min(Math.round(srcW * 2), 800)
        const targetH = Math.round((targetW / srcW) * srcH)
        const finalCanvas = bilinearUpscaleStage(currentCanvas, targetW, targetH)

        // Prefer WebP for compact high-quality encoding; fall back to JPEG
        let enhancedUrl = finalCanvas.toDataURL('image/webp', 0.92)
        if (!enhancedUrl || enhancedUrl.startsWith('data:image/png')) {
          enhancedUrl = finalCanvas.toDataURL('image/jpeg', 0.92)
        }

        // Maintain small LRU memory footprint
        if (artCache.size > 20) {
          const firstKey = artCache.keys().next().value
          if (firstKey) artCache.delete(firstKey)
        }
        artCache.set(src, enhancedUrl)

        resolve({ url: enhancedUrl, isUpscaled: true, width: targetW, height: targetH })
      } catch {
        resolve({ url: src, isUpscaled: false, width: srcW, height: srcH })
      }
    }

    img.onerror = () => {
      if (blobUrl) URL.revokeObjectURL(blobUrl)
      resolve({ url: src, isUpscaled: false, width: 0, height: 0 })
    }

    img.src = processUrl
  })
}
