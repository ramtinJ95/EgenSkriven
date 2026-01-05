import styles from './Skeleton.module.css';

interface SkeletonProps {
  /** Width of skeleton (CSS value) */
  width?: string;
  /** Height of skeleton (CSS value) */
  height?: string;
  /** Border radius (CSS value) */
  radius?: string;
  /** Additional class name */
  className?: string;
}

/**
 * Generic skeleton loader component.
 * 
 * @example
 * <Skeleton width="100%" height="20px" />
 */
export function Skeleton({ 
  width = '100%', 
  height = '16px', 
  radius = 'var(--radius-sm)',
  className = ''
}: SkeletonProps) {
  return (
    <div 
      className={`${styles.skeleton} ${className}`}
      style={{ width, height, borderRadius: radius }}
      role="status"
      aria-label="Loading..."
    />
  );
}

/**
 * Skeleton for a task card.
 * Matches the visual structure of TaskCard component.
 */
export function TaskCardSkeleton() {
  return (
    <div className={styles.taskCardSkeleton}>
      <div className={styles.taskCardSkeletonHeader}>
        <Skeleton width="12px" height="12px" radius="50%" />
        <Skeleton width="60px" height="12px" />
      </div>
      <Skeleton width="100%" height="16px" className={styles.taskCardSkeletonTitle} />
      <Skeleton width="70%" height="14px" />
      <div className={styles.taskCardSkeletonFooter}>
        <Skeleton width="50px" height="20px" radius="var(--radius-sm)" />
        <Skeleton width="24px" height="16px" />
      </div>
    </div>
  );
}

/**
 * Skeleton for a column of tasks.
 * Shows header and multiple task card skeletons.
 */
export function ColumnSkeleton({ cardCount = 3 }: { cardCount?: number }) {
  return (
    <div className={styles.columnSkeleton}>
      <div className={styles.columnSkeletonHeader}>
        <Skeleton width="100px" height="16px" />
        <Skeleton width="24px" height="16px" radius="var(--radius-sm)" />
      </div>
      <div className={styles.columnSkeletonCards}>
        {Array.from({ length: cardCount }).map((_, i) => (
          <TaskCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}

/**
 * Full board skeleton for initial load.
 */
export function BoardSkeleton() {
  return (
    <div className={styles.boardSkeleton} role="status" aria-label="Loading board...">
      <ColumnSkeleton cardCount={2} />
      <ColumnSkeleton cardCount={4} />
      <ColumnSkeleton cardCount={2} />
      <ColumnSkeleton cardCount={1} />
      <ColumnSkeleton cardCount={3} />
    </div>
  );
}
