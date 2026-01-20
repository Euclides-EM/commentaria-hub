import { useState } from 'react'
import { AppStateProvider } from './context/AppStateContext'
import { Header } from './components/Header'
import { LoginModal } from './components/LoginModal'
import { MainApp } from './components/MainApp'

function App() {
  const [showLoginModal, setShowLoginModal] = useState(false)

  return (
    <AppStateProvider>
      <div className="h-screen flex flex-col overflow-hidden overscroll-none">
        <Header onShowLogin={() => setShowLoginModal(true)} />

        <div className="flex-1 overflow-hidden">
          <MainApp />
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

export default App
