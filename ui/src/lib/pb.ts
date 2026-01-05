import PocketBase from 'pocketbase'

// Debug logging - enabled via VITE_DEBUG_REALTIME=true
const DEBUG_REALTIME = import.meta.env.VITE_DEBUG_REALTIME === 'true'
const debugLog = (...args: unknown[]) => {
  if (DEBUG_REALTIME) {
    console.log('[PB]', ...args)
  }
}

// Create a single PocketBase client instance
// In production, the UI and API are served from the same origin
// In development, Vite proxies /api requests to localhost:8090
export const pb = new PocketBase('/')

// Disable auto-cancellation to prevent issues with React strict mode
pb.autoCancellation(false)

// SSE Connection Monitoring (only in browser environment)
if (typeof window !== 'undefined' && DEBUG_REALTIME) {
  // Monitor connection state
  const checkConnection = () => {
    debugLog('Realtime client ID:', pb.realtime.clientId)
    debugLog('Is connected:', !!pb.realtime.clientId)
  }

  // Check periodically during development
  setInterval(checkConnection, 30000)

  // Log initial state after a short delay
  setTimeout(checkConnection, 1000)
}
