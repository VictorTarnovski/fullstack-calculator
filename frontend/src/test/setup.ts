import '@testing-library/jest-dom/vitest'

/* jsdom does not implement ResizeObserver. useFitText only needs an instance
   that can be constructed and disconnected; tests that care about the resize
   callback itself install their own stub via vi.stubGlobal. */
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

if (!('ResizeObserver' in globalThis)) {
  globalThis.ResizeObserver = ResizeObserverStub as unknown as typeof ResizeObserver
}
