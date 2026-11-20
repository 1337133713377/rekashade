/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

interface Window {
  go?: { main?: { App?: Record<string, (...args: unknown[]) => Promise<unknown>> } }
  runtime?: {
    EventsOn: (name: string, cb: (...data: unknown[]) => void) => () => void
    EventsOff: (name: string) => void
  }
}
