<style>
  :global(:root) {
    --vinta-border-width: 1px;
    --vinta-border-color: var(--c-border);

    /* 
     * --vinta-border-shadow
     * 
     * Webview2 has a rendering bug where standard CSS borders (or single 
     * inset box-shadows) can appear faint, clipped, or bleed incorrectly, 
     * especially when using translucent border colors.
     * 
     * To fix this, we use a double inset box-shadow. This custom property
     * defines the double-inset fix as a reusable variable so developers
     * don't have to type it manually.
     * 
     * Usage (Composable):
     *   .my-element {
     *     border: none;
     *     --vinta-border-color: var(--c-primary);
     *     box-shadow: 
     *       var(--vinta-border-shadow),
     *       0 4px 10px rgba(0,0,0,0.5); // Can be composed with drop shadows!
     *   }
     */
    --vinta-border-shadow:
      inset 0 0 0 var(--vinta-border-width) var(--vinta-border-color),
      inset 0 0 0 var(--vinta-border-width) var(--vinta-border-color);
  }

  /* 
   * .vinta-border
   * 
   * Convenience class for elements that ONLY need the border and DO NOT 
   * have their own box-shadows. 
   * 
   * If your element already uses `box-shadow`, do NOT use this class!
   * Compose `var(--vinta-border-shadow)` directly into your CSS instead.
   */
  :global(.vinta-border) {
    border: none !important;
    box-shadow: var(--vinta-border-shadow);
  }
</style>
