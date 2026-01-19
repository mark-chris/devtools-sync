import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import App from './App'

describe('App', () => {
  afterEach(() => {
    cleanup()
  })

  it('renders dashboard title', () => {
    render(<App />)
    const heading = screen.getByRole('heading', { level: 1 })
    expect(heading).toHaveTextContent('DevTools Sync Dashboard')
  })

  it('renders status message', () => {
    render(<App />)
    const message = screen.getByText(/local development environment is running/i)
    expect(message).toBeInTheDocument()
  })

  it('renders server health link', () => {
    render(<App />)
    const link = screen.getByRole('link', { name: /http:\/\/localhost:8080\/health/i })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', 'http://localhost:8080/health')
  })
})
