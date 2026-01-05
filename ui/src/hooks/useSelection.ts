import { useContext } from 'react'
import {
  SelectionContext,
  type SelectionContextValue,
} from '../contexts/SelectionContext'

/**
 * Hook to access selection state and actions.
 * Must be used within a SelectionProvider.
 *
 * @example
 * function TaskCard({ task }) {
 *   const { isSelected, selectTask } = useSelection();
 *   return (
 *     <div
 *       className={isSelected(task.id) ? 'selected' : ''}
 *       onClick={() => selectTask(task.id)}
 *     >
 *       {task.title}
 *     </div>
 *   );
 * }
 */
export function useSelection(): SelectionContextValue {
  const context = useContext(SelectionContext)
  if (!context) {
    throw new Error('useSelection must be used within a SelectionProvider')
  }
  return context
}
