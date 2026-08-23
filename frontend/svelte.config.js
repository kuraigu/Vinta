import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
import fs from 'node:fs'
import path from 'node:path'

function externalStylePreprocess() {
  return {
    name: 'external-style-preprocessor',
    style({ attributes, filename }) {
      if (attributes.src && typeof attributes.src === 'string' && filename) {
        const stylePath = path.resolve(path.dirname(filename), attributes.src)
        if (fs.existsSync(stylePath)) {
          const code = fs.readFileSync(stylePath, 'utf-8')
          return { code }
        }
      }
    }
  }
}

export default {
  preprocess: [externalStylePreprocess(), vitePreprocess()]
}
