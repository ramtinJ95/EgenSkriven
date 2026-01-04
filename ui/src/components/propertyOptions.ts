import type { PropertyOption } from './PropertyPicker'

// Pre-configured options for common properties

export const STATUS_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'backlog',
    label: 'Backlog',
    icon: '●',
    color: 'var(--status-backlog, #6B7280)',
  },
  {
    value: 'todo',
    label: 'Todo',
    icon: '●',
    color: 'var(--status-todo, #E5E5E5)',
  },
  {
    value: 'in_progress',
    label: 'In Progress',
    icon: '●',
    color: 'var(--status-in-progress, #F59E0B)',
  },
  {
    value: 'review',
    label: 'Review',
    icon: '●',
    color: 'var(--status-review, #A855F7)',
  },
  {
    value: 'done',
    label: 'Done',
    icon: '●',
    color: 'var(--status-done, #22C55E)',
  },
]

export const PRIORITY_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'urgent',
    label: 'Urgent',
    icon: '🔴',
    color: 'var(--priority-urgent, #EF4444)',
  },
  {
    value: 'high',
    label: 'High',
    icon: '🟠',
    color: 'var(--priority-high, #F97316)',
  },
  {
    value: 'medium',
    label: 'Medium',
    icon: '🟡',
    color: 'var(--priority-medium, #EAB308)',
  },
  {
    value: 'low',
    label: 'Low',
    icon: '⚪',
    color: 'var(--priority-low, #6B7280)',
  },
]

export const TYPE_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'bug',
    label: 'Bug',
    icon: '🐛',
    color: 'var(--type-bug, #EF4444)',
  },
  {
    value: 'feature',
    label: 'Feature',
    icon: '✨',
    color: 'var(--type-feature, #A855F7)',
  },
  {
    value: 'chore',
    label: 'Chore',
    icon: '🔧',
    color: 'var(--type-chore, #6B7280)',
  },
]
