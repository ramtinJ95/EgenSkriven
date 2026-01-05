import { useState, type ReactNode } from 'react'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import styles from './Layout.module.css'

interface LayoutProps {
  children: ReactNode
}

const SIDEBAR_COLLAPSED_KEY = 'egenskriven-sidebar-collapsed'

/**
 * Main application layout.
 *
 * Structure:
 * - Sidebar: Board navigation (collapsible)
 * - Header: App title and actions
 * - Main: Content area (board/list view)
 *
 * Note: CurrentBoardProvider must wrap App, not Layout,
 * because AppContent uses useCurrentBoard before Layout renders.
 */
export function Layout({ children }: LayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    const saved = localStorage.getItem(SIDEBAR_COLLAPSED_KEY)
    return saved === 'true'
  })

  const handleToggleSidebar = () => {
    const newValue = !sidebarCollapsed
    setSidebarCollapsed(newValue)
    localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(newValue))
  }

  return (
    <div className={styles.layout}>
      <div className={styles.body}>
        <Sidebar collapsed={sidebarCollapsed} onToggle={handleToggleSidebar} />
        <div className={styles.content}>
          <Header />
          <main className={styles.main}>{children}</main>
        </div>
      </div>
    </div>
  )
}
