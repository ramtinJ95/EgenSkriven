import PocketBase from 'pocketbase'

// Create a single PocketBase client instance
// In production, the UI and API are served from the same origin
// In development, Vite proxies /api requests to localhost:8090
export const pb = new PocketBase('/')

// Disable auto-cancellation to prevent issues with React strict mode
pb.autoCancellation(false)
