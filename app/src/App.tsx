import { useState } from 'react'
import { AppStateProvider } from './context/AppStateProvider'
import { Header } from './components/Header'
import { LoginModal } from './components/modal/LoginModal'
import { Main } from './components/Main.tsx'

export function App() {
  const [showLoginModal, setShowLoginModal] = useState(false)

  return (
    <AppStateProvider>
      <div className="h-screen flex flex-col overflow-hidden overscroll-none">
        <Header onShowLogin={() => setShowLoginModal(true)} />

        <div className="flex-1 overflow-hidden">
          <Main />
        </div>
        <div className="px-3 py-2 text-xs text-gray-500 text-center">
          © {new Date().getFullYear()} Euclides Project. Avec Privilege du Roy.
        </div>
      </div>

      {showLoginModal && (
        <LoginModal
          onClose={() => setShowLoginModal(false)}
          onSuccess={() => setShowLoginModal(false)}
        />
      )}
    </AppStateProvider>
  )
}
