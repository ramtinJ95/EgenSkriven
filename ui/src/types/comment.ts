import type { RecordModel } from 'pocketbase'

// Author type for comments
export type AuthorType = 'human' | 'agent'

// Comment metadata for additional context
export interface CommentMetadata {
  mentions?: string[] // @agent, @user mentions extracted from content
  action?: string // Optional action that triggered this comment
}

// Comment record from PocketBase
// Aligns with migrations/13_comments_collection.go schema
export interface Comment extends RecordModel {
  task: string // Task ID (relation to tasks collection)
  content: string // Comment text content
  author_type: AuthorType // Who created the comment
  author_id?: string // Username, agent name, or identifier
  metadata?: CommentMetadata // Additional context
  created: string // ISO timestamp (auto-populated by PocketBase)
}

// Input for creating a new comment
export interface CreateCommentInput {
  task: string
  content: string
  author_type: AuthorType
  author_id?: string
}
