const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const fetchHealth    = () => fetch(`${BASE}/health`).then(r => r.json())
export const fetchCluster   = () => fetch(`${BASE}/api/v1/cluster`).then(r => r.json())
export const fetchServices  = () => fetch(`${BASE}/api/v1/services`).then(r => r.json())
export const fetchResources = () => fetch(`${BASE}/api/v1/resources`).then(r => r.json())
