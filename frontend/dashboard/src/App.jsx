import { useState, useEffect, useRef } from 'react'
import { fetchHealth, fetchCluster, fetchServices, fetchResources } from './api.js'
import Header from './components/Header.jsx'
import ClusterStatus from './components/ClusterStatus.jsx'
import ResourceGauges from './components/ResourceGauges.jsx'
import ServicesGrid from './components/ServicesGrid.jsx'
import Pipeline from './components/Pipeline.jsx'
import EventTicker from './components/EventTicker.jsx'

const POLL_INTERVAL = 5000

export default function App() {
  const [health, setHealth]       = useState(null)
  const [cluster, setCluster]     = useState(null)
  const [services, setServices]   = useState(null)
  const [resources, setResources] = useState(null)

  // Ref-backed history for sparklines (kept outside React state to avoid re-render
  // storms; gauge components read it via props on each poll).
  const cpuHistory  = useRef([])
  const memHistory  = useRef([])
  const diskHistory = useRef([])

  useEffect(() => {
    const load = async () => {
      const [h, c, svc, res] = await Promise.allSettled([
        fetchHealth(),
        fetchCluster(),
        fetchServices(),
        fetchResources(),
      ])

      if (h.status   === 'fulfilled') setHealth(h.value)
      if (c.status   === 'fulfilled') setCluster(c.value)
      if (svc.status === 'fulfilled') setServices(svc.value)
      if (res.status === 'fulfilled') {
        const r = res.value
        setResources(r)
        // Append to rolling sparkline history (keep last 12 samples).
        cpuHistory.current  = [...cpuHistory.current,  r.cpu_percent].slice(-12)
        memHistory.current  = [...memHistory.current,  r.memory_percent].slice(-12)
        diskHistory.current = [...diskHistory.current, r.disk_io_percent].slice(-12)
      }
    }

    load()
    const interval = setInterval(load, POLL_INTERVAL)
    return () => clearInterval(interval)
  }, [])

  return (
    <div style={{ minHeight: '100vh', display: 'grid', gridTemplateRows: 'auto 1fr auto' }}>
      <Header health={health} />

      <div style={{
        padding: '0 24px 24px',
        display: 'grid',
        gap: '16px',
        gridTemplateColumns: '1fr 1fr 1fr',
        gridTemplateRows: 'auto auto auto',
        alignItems: 'start',
      }}>
        <ClusterStatus cluster={cluster} />
        <ResourceGauges
          resources={resources}
          cpuHistory={cpuHistory.current}
          memHistory={memHistory.current}
          diskHistory={diskHistory.current}
          cluster={cluster}
        />
        <ServicesGrid services={services} />
        <Pipeline />
      </div>

      <EventTicker services={services} cluster={cluster} />
    </div>
  )
}
