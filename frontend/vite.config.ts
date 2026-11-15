import { defineConfig, type Plugin } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import wails from '@wailsio/runtime/plugins/vite'
import path from 'path'

// Converts PascalCase icon names to kebab-case (e.g., BellRing -> bell-ring, Minimize2 -> minimize-2, Volume2 -> volume-2)
function pascalToKebab(str: string): string {
  return str
    .replace(/([a-zA-Z])([0-9])/g, '$1-$2')
    .replace(/([0-9])([a-zA-Z])/g, '$1-$2')
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1-$2')
    .toLowerCase()
}

// Vite plugin to automatically rewrite named lucide-svelte imports into direct per-icon imports.
// Example: import { Eye, Minimize2 } from 'lucide-svelte' ->
//          import Eye from 'lucide-svelte/icons/eye.js'
//          import Minimize2 from 'lucide-svelte/icons/minimize-2.js'
// This eliminates bundling/evaluating all 1500+ icons in lucide-svelte, dramatically boosting startup & build speed.
function optimizeLucideImports(): Plugin {
  return {
    name: 'optimize-lucide-imports',
    enforce: 'pre',
    transform(code, id) {
      if (!id.endsWith('.svelte') && !id.endsWith('.ts') && !id.endsWith('.js'))
        return null

      const lucideImportRegex =
        /import\s+\{([^}]+)\}\s+from\s+['"]lucide-svelte['"]/g
      if (!lucideImportRegex.test(code)) return null

      const transformed = code.replace(lucideImportRegex, (_, importsStr) => {
        const specifiers = importsStr
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)

        const directImports = specifiers.map((specifier) => {
          const parts = specifier.split(/\s+as\s+/).map((p) => p.trim())
          const iconName = parts[0]
          const alias = parts[1] || iconName
          const kebabName = pascalToKebab(iconName)
          return `import ${alias} from 'lucide-svelte/icons/${kebabName}.js'`
        })

        return directImports.join('\n')
      })

      return {
        code: transformed,
        map: null
      }
    }
  }
}

function optimizeBackdropFilter(): Plugin {
  const transformCss = (css: string): string => {
    if (!css.includes('backdrop-filter')) return css
    return css.replace(
      /([^{}]+)\{([^{}]*backdrop-filter\s*:\s*(?!none\b)[^;{}]+;[^{}]*)\}/g,
      (fullMatch, selector: string, body: string) => {
        const trimmedSel = selector.trim()
        // Never inject declarations into @keyframes step blocks (from, to, 0%, 100%).
        if (
          trimmedSel === 'from' ||
          trimmedSel === 'to' ||
          /(?:^|[\s,])\d+(?:\.\d+)?%$/.test(trimmedSel) ||
          trimmedSel.includes('@keyframes')
        ) {
          return fullMatch
        }

        let nextBody = body
        const match = body.match(/(^|[;\s])backdrop-filter\s*:\s*([^;{}]+);/)
        if (!match) return fullMatch

        const val = match[2].trim()
        if (!/-webkit-backdrop-filter\s*:/.test(nextBody)) {
          nextBody = nextBody.replace(
            /(^|[;\s])(backdrop-filter\s*:\s*[^;{}]+;)/,
            `$1-webkit-backdrop-filter: ${val};\n  $2`
          )
        }
        // Bound the backdrop-filter stacking context without transform: translateZ(0),
        // which breaks position: fixed popovers/modals and disables Windows ClearType text smoothing.
        if (!/\bisolation\s*:/.test(nextBody)) {
          nextBody += '\n  isolation: isolate;'
        }
        return `${selector}{${nextBody}}`
      }
    )
  }

  return {
    name: 'optimize-backdrop-filter',
    enforce: 'pre',
    transform(code, id) {
      const cleanId = id.split('?')[0]
      if (!cleanId.endsWith('.css') && !cleanId.endsWith('.svelte')) return null
      if (!code.includes('backdrop-filter')) return null

      let transformed = code
      if (cleanId.endsWith('.svelte')) {
        // Scope CSS transforms strictly to <style> blocks so script/template expressions are untouched.
        transformed = code.replace(
          /(<style\b[^>]*>)([\s\S]*?)(<\/style>)/gi,
          (_m, open: string, css: string, close: string) =>
            `${open}${transformCss(css)}${close}`
        )
      } else {
        transformed = transformCss(code)
      }

      if (transformed === code) return null
      return { code: transformed, map: null }
    }
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true
  },
  plugins: [
    optimizeLucideImports(),
    optimizeBackdropFilter(),
    svelte(),
    wails('./bindings')
  ],
  optimizeDeps: {
    include: ['svelte', 'svelte/internal', 'svelte/transition', 'svelte/store'],
    exclude: ['lucide-svelte']
  },
  build: {
    emptyOutDir: true,
    target: 'esnext',
    cssTarget: 'chrome120'
  },
  resolve: {
    alias: {
      'lucide-svelte/icons': path.resolve(
        import.meta.dirname,
        'node_modules/lucide-svelte/dist/icons'
      ),
      '@appicon': path.resolve(import.meta.dirname, '../build/appicon.png')
    }
  }
})
