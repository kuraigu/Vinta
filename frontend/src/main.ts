import { mount } from 'svelte'
import './style.css'
import App from './App.svelte'
import DevTestWindow from './devtest/DevTestWindow.svelte'

const urlParams = new URLSearchParams(window.location.search)
const isDevTestWindow = urlParams.get('window') === 'devtest'

if (isDevTestWindow && typeof document !== 'undefined') {
  document.documentElement.classList.add('devtest-window-root')
  document.body.classList.add('devtest-window-body')
}

const app = mount(isDevTestWindow ? DevTestWindow : App, {
  target: document.getElementById('app')!
})

export default app
