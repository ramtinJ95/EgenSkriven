import { useState, type ReactNode } from 'react'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { FilterBar } from './FilterBar'
import type { Task } from '../types/task'
import styles from './Layout.module.css'

interface LayoutProps {
  children: ReactNode
  totalTasks: number
  filteredTasks: number
  onOpenFilterBuilder: () => void
  onOpenDisplayOptions: () => void
  onOpenSettings?: () => void
  onOpenHelp?: () => void
  /** All tasks (unfiltered) for epic counting in sidebar */
  tasks?: Task[]
  /** Currently selected epic filter */
  selectedEpicId?: string | null
  /** Callback when epic filter changes */
  onSelectEpic?: (epicId: string | null) => void
  /** Callback when epic detail view should be opened */
  onEpicDetailClick?: (epicId: string) => void
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
  onOpenHelp,
  tasks = [],
  selectedEpicId,
  onSelectEpic,
  onEpicDetailClick,
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
        <Sidebar
          collapsed={sidebarCollapsed}
          onToggle={handleToggleSidebar}
          tasks={tasks}
          selectedEpicId={selectedEpicId}
          onSelectEpic={onSelectEpic}
          onEpicDetailClick={onEpicDetailClick}
        />
        <div className={styles.content}>
          <Header onDisplayOptionsClick={onOpenDisplayOptions} onSettingsClick={onOpenSettings} onHelpClick={onOpenHelp} />
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
