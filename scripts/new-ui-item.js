/**
 * Vinta UI Scaffolding Generator
 *
 * Creates a new component, Home widget, workspace page, or overlay page from
 * templates/, separates its stylesheet into styles/, and prints registration steps.
 *
 * Usage:
 *   node scripts/new-ui-item.js <component|widget|page|overlay> <PascalName> [--dry-run]
 */

const fs = require('fs')
const path = require('path')

const ROOT = path.resolve(__dirname, '..')
const VALID_KINDS = ['component', 'widget', 'page', 'overlay']

function toKebab(name) {
  return name
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1-$2')
    .toLowerCase()
}

function showHelp() {
  console.log(
    [
      'Usage: node scripts/new-ui-item.js <component|widget|page|overlay> <Name> [--dry-run]',
      '',
      'Examples:',
      '  node scripts/new-ui-item.js component QuickBadge',
      '  node scripts/new-ui-item.js widget WeatherRadar',
      '  node scripts/new-ui-item.js page PomodoroStats',
      '  node scripts/new-ui-item.js overlay CommandPalette'
    ].join('\n')
  )
}

function normalizeName(kind, rawName) {
  if (!/^[A-Z][A-Za-z0-9]*$/.test(rawName)) {
    console.error(`Error: Name "${rawName}" must be PascalCase alphanumeric (e.g. FocusSummary).`)
    process.exit(1)
  }
  if (kind === 'widget' && !rawName.endsWith('Widget')) return `${rawName}Widget`
  if (kind === 'page' && !rawName.endsWith('Page')) return `${rawName}Page`
  if (kind === 'overlay' && !rawName.endsWith('OverlayPage')) {
    return rawName.endsWith('Overlay') ? `${rawName}Page` : `${rawName}OverlayPage`
  }
  return rawName
}

function buildFiles(kind, componentName) {
  const kebab = toKebab(componentName)
  const baseKebab = kebab.replace(/-(widget|overlay-page|page)$/, '')

  if (kind === 'component') {
    const templatePath = path.join(ROOT, 'templates/component/ExampleComponent.svelte')
    const raw = fs.readFileSync(templatePath, 'utf8')
    const styleMatch = raw.match(/<style>([\s\S]*?)<\/style>/)
    const css = styleMatch ? styleMatch[1].trim() + '\n' : ''
    const svelte = raw
      .replace(/example-container/g, `${baseKebab}-container`)
      .replace(/<style>[\s\S]*?<\/style>/, `<style src="./styles/${componentName}.css"></style>`)

    return {
      sveltePath: path.join(ROOT, 'frontend/src/components', `${componentName}.svelte`),
      cssPath: path.join(ROOT, 'frontend/src/components/styles', `${componentName}.css`),
      svelte,
      css,
      steps: [
        `Import ${componentName} from '../components/${componentName}.svelte' where needed.`,
        'Replace placeholder text with $tStore(...) keys in frontend/src/locales/en.json and run task sync-locales.'
      ]
    }
  }

  if (kind === 'widget') {
    const templatePath = path.join(ROOT, 'templates/widget/ExampleWidget.svelte')
    const raw = fs.readFileSync(templatePath, 'utf8')
    const styleMatch = raw.match(/<style>([\s\S]*?)<\/style>/)
    const css = styleMatch ? styleMatch[1].trim() + '\n' : ''
    const svelte = raw
      .replace(/ExampleWidget/g, componentName)
      .replace(/id="example"/g, `id="${baseKebab}"`)
      .replace(/example-widget-/g, `${baseKebab}-widget-`)
      .replace(/Example Widget/g, componentName.replace(/Widget$/, ''))
      .replace(/<style>[\s\S]*?<\/style>/, `<style src="../styles/${componentName}.css"></style>`)

    return {
      sveltePath: path.join(ROOT, 'frontend/src/components/widgets', `${componentName}.svelte`),
      cssPath: path.join(ROOT, 'frontend/src/components/styles', `${componentName}.css`),
      svelte,
      css,
      steps: [
        `Register '${baseKebab}' in HomeWidgetsVisibility / widget order in frontend/src/store.ts.`,
        `Mount <${componentName} /> inside frontend/src/pages/HomePage.svelte.`,
        'Add i18n keys in frontend/src/locales/en.json and run task sync-locales.'
      ]
    }
  }

  if (kind === 'page') {
    const templatePath = path.join(ROOT, 'templates/page/ExamplePage.svelte')
    const raw = fs.readFileSync(templatePath, 'utf8')
    const styleMatch = raw.match(/<style>([\s\S]*?)<\/style>/)
    const css = styleMatch ? styleMatch[1].trim() + '\n' : ''
    const svelte = raw
      .replace(/\.\.\/\.\.\/frontend\/src\//g, '../')
      .replace(/ExamplePage/g, componentName)
      .replace(/data-vinta="example-page"/g, `data-vinta="${baseKebab}-page"`)
      .replace(/example-minimal/g, `${baseKebab}-minimal`)
      .replace(/example-page/g, `${baseKebab}-page`)
      .replace(/example-scroll-area/g, `${baseKebab}-scroll-area`)
      .replace(/example-body/g, `${baseKebab}-body`)
      .replace(/Example Page/g, componentName.replace(/Page$/, ''))
      .replace(/<style>[\s\S]*?<\/style>/, `<style src="./styles/${componentName}.css"></style>`)

    return {
      sveltePath: path.join(ROOT, 'frontend/src/pages', `${componentName}.svelte`),
      cssPath: path.join(ROOT, 'frontend/src/pages/styles', `${componentName}.css`),
      svelte,
      css,
      steps: [
        `Add '${baseKebab}' to page visibility & routing in frontend/src/store.ts and frontend/src/App.svelte.`,
        `Add navigation helper in frontend/src/utils/navigation.ts (AppControls).`,
        `If listed in AppsOverlayPage.svelte, add .app-tile.app-${baseKebab} .tile-icon-box gradient in AppsOverlayPage.css.`,
        'Add i18n keys in frontend/src/locales/en.json and run task sync-locales.'
      ]
    }
  }

  // overlay
  const templatePath = path.join(ROOT, 'templates/overlay/ExampleOverlayPage.svelte')
  const raw = fs.readFileSync(templatePath, 'utf8')
  const styleMatch = raw.match(/<style>([\s\S]*?)<\/style>/)
  const css = styleMatch ? styleMatch[1].trim() + '\n' : ''
  const svelte = raw
    .replace(/ExampleOverlayPage/g, componentName)
    .replace(/dataVinta="example-overlay"/g, `dataVinta="${baseKebab}-overlay"`)
    .replace(/example-overlay-scroll/g, `${baseKebab}-overlay-scroll`)
    .replace(/Example Overlay/g, componentName.replace(/OverlayPage$/, ''))
    .replace(/<style>[\s\S]*?<\/style>/, `<style src="../styles/${componentName}.css"></style>`)

  return {
    sveltePath: path.join(ROOT, 'frontend/src/pages/overlays', `${componentName}.svelte`),
    cssPath: path.join(ROOT, 'frontend/src/pages/styles', `${componentName}.css`),
    svelte,
    css,
    steps: [
      `Add '${baseKebab}' to OverlayPageId in frontend/src/store.ts.`,
      `Mount <${componentName} onClose={onCloseOverlay} /> in frontend/src/components/MainIsland.svelte.`,
      'Add i18n keys in frontend/src/locales/en.json and run task sync-locales.'
    ]
  }
}

function run() {
  const args = process.argv.slice(2)
  if (args.includes('--help') || args.includes('-h') || args.length < 2) {
    showHelp()
    process.exit(args.includes('--help') || args.includes('-h') ? 0 : 1)
  }

  const kind = args[0].toLowerCase()
  const rawName = args[1]
  const dryRun = args.includes('--dry-run')

  if (!VALID_KINDS.includes(kind)) {
    console.error(`Error: Unknown type "${kind}". Expected one of: ${VALID_KINDS.join(', ')}`)
    process.exit(1)
  }

  const componentName = normalizeName(kind, rawName)
  const plan = buildFiles(kind, componentName)

  if (fs.existsSync(plan.sveltePath)) {
    console.error(`Error: Target already exists: ${path.relative(ROOT, plan.sveltePath)}`)
    process.exit(1)
  }

  if (dryRun) {
    console.log('[dry-run] Would create:')
    console.log(`  - ${path.relative(ROOT, plan.sveltePath)}`)
    console.log(`  - ${path.relative(ROOT, plan.cssPath)}`)
  } else {
    fs.mkdirSync(path.dirname(plan.sveltePath), { recursive: true })
    fs.mkdirSync(path.dirname(plan.cssPath), { recursive: true })
    fs.writeFileSync(plan.sveltePath, plan.svelte, 'utf8')
    fs.writeFileSync(plan.cssPath, plan.css, 'utf8')
    console.log('Created:')
    console.log(`  - ${path.relative(ROOT, plan.sveltePath)}`)
    console.log(`  - ${path.relative(ROOT, plan.cssPath)}`)
  }

  console.log('\nNext steps:')
  plan.steps.forEach((step, idx) => {
    console.log(`  ${idx + 1}. ${step}`)
  })
}

run()
