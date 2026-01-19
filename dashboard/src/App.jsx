import React from 'react'

function App() {
  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui' }}>
      <h1>DevTools Sync Dashboard</h1>
      <p>Local development environment is running.</p>
      <p style={{ color: '#666' }}>
        Server: <a href="http://localhost:8080/health">http://localhost:8080/health</a>
      </p>
    </div>
  )
}

export default App
