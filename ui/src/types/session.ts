import type { RecordModel } from 'pocketbase'

// Supported AI tools for session tracking
export type SessionTool = 'opencode' | 'claude-code' | 'codex'

// Reference type for session identifiers
export type SessionRefType = 'uuid' | 'path'

// Session status values
export type SessionStatus = 'active' | 'paused' | 'completed' | 'abandoned'

// Agent session embedded in task record (agent_session JSON field)
// This represents the currently linked session for a task
// Aligns with the agent_session field schema in tasks collection
export interface AgentSession {
  tool: SessionTool // Which AI tool owns this session
  ref: string // Session/thread ID from the tool
  ref_type: SessionRefType // Type of reference (UUID or path)
  working_dir: string // Project directory where session operates
  linked_at: string // ISO timestamp when session was linked
}

// Session record from sessions collection
// Tracks historical sessions for a task (including abandoned ones)
// Aligns with migrations/14_sessions_collection.go schema
export interface SessionRecord extends RecordModel {
  task: string // Task ID (relation to tasks collection)
  tool: SessionTool // Which AI tool owns this session
  external_ref: string // Session/thread ID from the tool
  ref_type: SessionRefType // Type of reference
  working_dir: string // Project directory
  status: SessionStatus // Current status of the session
  created: string // ISO timestamp (auto-populated by PocketBase)
  ended_at?: string // ISO timestamp when session ended (optional)
}
