import type { ReactNode } from 'react'
import { Header } from './Header'
import styles from './Layout.module.css'

interface LayoutProps {
  children: ReactNode
}

/**
 * Main application layout.
 * 
 * Structure:
 * - Header: App title and actions
 * - Main: Content area (board/list view)
 * 
 * Note: Sidebar will be added in Phase 5 (Multi-board support)
 */
export function Layout({ children }: LayoutProps) {
  return (
    <div className={styles.layout}>
      <Header />
      <main className={styles.main}>{children}</main>
    </div>
  )
}
