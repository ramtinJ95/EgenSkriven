import { useState, type ReactNode } from 'react'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { FilterBar } from './FilterBar'
import styles from './Layout.module.css'

interface LayoutProps {
  children: ReactNode
  totalTasks: number
  filteredTasks: number
  onOpenFilterBuilder: () => void
  onOpenDisplayOptions: () => void
  onOpenSettings?: () => void
}

const SIDEBAR_COLLAPSED_KEY = 'egenskriven-sidebar-collapsed'

/**
 * Main application layout.
 *
 * Structure:
 * - Sidebar: Board navigation (collapsible)
 * - Header: App title, search, and display options
 * - FilterBar: Active filters display
 * - Main: Content area (board/list view)
 *
 * Note: CurrentBoardProvider must wrap App, not Layout,
 * because AppContent uses useCurrentBoard before Layout renders.
 */
export function Layout({
  children,
  totalTasks,
  filteredTasks,
  onOpenFilterBuilder,
  onOpenDisplayOptions,
  onOpenSettings,
}: LayoutProps) {
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
          <Header onDisplayOptionsClick={onOpenDisplayOptions} onSettingsClick={onOpenSettings} />
          <FilterBar
            totalTasks={totalTasks}
            filteredTasks={filteredTasks}
            onOpenFilterBuilder={onOpenFilterBuilder}
          />
          <main className={styles.main}>{children}</main>
        </div>
      </div>
    </div>
  )
}
