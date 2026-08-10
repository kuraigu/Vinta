import { defineConfig, type Plugin } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
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

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [optimizeLucideImports(), svelte()],
  optimizeDeps: {
    include: ['svelte', 'svelte/internal', 'svelte/transition', 'svelte/store'],
    exclude: ['lucide-svelte']
  },
  resolve: {
    alias: {
      'lucide-svelte/icons': path.resolve(
        __dirname,
        'node_modules/lucide-svelte/dist/icons'
      )
    }
  }
})
