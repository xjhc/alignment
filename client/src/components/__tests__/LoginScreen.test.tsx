import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LoginScreen } from '../LoginScreen'

describe('LoginScreen', () => {
  it('renders login form with correct elements', () => {
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    expect(screen.getByText(/LOEBIAN INC./)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('[ENTER YOUR HANDLE]')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /BROWSE LOBBIES/ })).toBeInTheDocument()
  })

  it('renders all avatar options', () => {
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const avatarButtons = screen.getAllByRole('button').filter(button =>
      ['👤', '🧑‍💻', '🕵️', '🤖', '🧑‍🚀'].includes(button.textContent || '')
    )
    expect(avatarButtons).toHaveLength(5)
  })

  it('has first avatar selected by default', () => {
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const firstAvatar = screen.getByRole('button', { name: '👤' })
    expect(firstAvatar).toHaveClass('selected')
  })

  it('allows avatar selection', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const robotAvatar = screen.getByRole('button', { name: '🤖' })
    await user.click(robotAvatar)

    expect(robotAvatar).toHaveClass('selected')
    expect(screen.getByRole('button', { name: '👤' })).not.toHaveClass('selected')
  })

  it('disables submit button when name is empty', () => {
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const submitButton = screen.getByRole('button', { name: /BROWSE LOBBIES/ })
    expect(submitButton).toBeDisabled()
  })

  it('enables submit button when name is entered', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]')
    const submitButton = screen.getByRole('button', { name: /BROWSE LOBBIES/ })

    await user.type(nameInput, 'TestPlayer')
    expect(submitButton).not.toBeDisabled()
  })

  it('calls onLogin with correct parameters when form is submitted', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]')
    const robotAvatar = screen.getByRole('button', { name: '🤖' })
    const submitButton = screen.getByRole('button', { name: /BROWSE LOBBIES/ })

    await user.click(robotAvatar)
    await user.type(nameInput, 'TestPlayer')
    await user.click(submitButton)

    expect(mockOnLogin).toHaveBeenCalledWith('TestPlayer', '🤖')
  })

  it('trims whitespace from player name', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]')
    const submitButton = screen.getByRole('button', { name: /BROWSE LOBBIES/ })

    await user.type(nameInput, '  TestPlayer  ')
    await user.click(submitButton)

    expect(mockOnLogin).toHaveBeenCalledWith('TestPlayer', '👤')
  })

  it('prevents form submission with only whitespace', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]')

    await user.type(nameInput, '   ')

    const submitButton = screen.getByRole('button', { name: /BROWSE LOBBIES/ })
    expect(submitButton).toBeDisabled()
  })

  it('handles form submission via Enter key', async () => {
    const user = userEvent.setup()
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]')

    await user.type(nameInput, 'TestPlayer')
    await user.keyboard('{Enter}')

    expect(mockOnLogin).toHaveBeenCalledWith('TestPlayer', '👤')
  })

  it('respects maxLength attribute on name input', () => {
    const mockOnLogin = vi.fn()
    render(<LoginScreen onLogin={mockOnLogin} />)

    const nameInput = screen.getByPlaceholderText('[ENTER YOUR HANDLE]') as HTMLInputElement
    expect(nameInput.maxLength).toBe(20)
  })
})