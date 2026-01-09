# Phase 5: Web UI

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 5 of 7  
> **Status**: In Progress  
> **Estimated Effort**: 3-4 days  
> **Prerequisites**: [Phase 4](./ai-workflow-phase-4.md) completed

---

## Todo List

### Type Definitions
- [x] **5.1.1** Create `ui/src/types/comment.ts` with Comment and CreateCommentInput interfaces
- [x] **5.1.2** Create `ui/src/types/session.ts` with AgentSession and SessionRecord interfaces

### React Hooks
- [x] **5.2.1** Create `ui/src/hooks/useComments.ts` with useComments query hook
- [x] **5.2.2** Add useAddComment mutation hook to useComments.ts
- [x] **5.2.3** Add useCommentsSubscription real-time subscription hook
- [x] **5.2.4** Implement extractMentions helper function for @mentions
- [ ] **5.5.1** Create `ui/src/hooks/useResume.ts` with resume mutation hook
- [ ] **5.5.2** Implement buildContextPrompt function in useResume.ts
- [ ] **5.5.3** Implement buildResumeCommand function for all 3 tools (opencode, claude-code, codex)

### UI Components
- [ ] **5.3.1** Create `ui/src/components/TaskDetail/CommentsPanel.tsx` component
- [ ] **5.3.2** Implement CommentItem sub-component with agent/human styling
- [ ] **5.3.3** Add loading state to CommentsPanel
- [ ] **5.3.4** Add empty state to CommentsPanel
- [ ] **5.3.5** Add comment submission form to CommentsPanel
- [ ] **5.3.6** Add @agent mention warning indicator in textarea
- [ ] **5.4.1** Create `ui/src/components/TaskDetail/SessionInfo.tsx` component
- [ ] **5.4.2** Implement tool icon/name display (opencode, claude-code, codex)
- [ ] **5.4.3** Add blocked/active status indicator
- [ ] **5.4.4** Add Resume button for need_input tasks
- [ ] **5.6.1** Create `ui/src/components/TaskDetail/ResumeModal.tsx` component
- [ ] **5.6.2** Implement generate command functionality
- [ ] **5.6.3** Add copy-to-clipboard functionality
- [ ] **5.6.4** Display working directory and instructions

### Task Card Updates
- [ ] **5.7.1** Add pulsing orange dot indicator for need_input tasks in TaskCard.tsx
- [ ] **5.7.2** Add "Needs Input" label badge for blocked tasks
- [ ] **5.7.3** Add orange ring styling around blocked task cards

### Kanban Board Updates
- [ ] **5.8.1** Add need_input column to columnConfig with distinct styling
- [ ] **5.8.2** Add pulsing indicator to need_input column header
- [ ] **5.8.3** Ensure drag-and-drop works to/from need_input column

### Integration
- [ ] **5.9.1** Update `ui/src/components/TaskDetail/index.tsx` to integrate CommentsPanel
- [ ] **5.9.2** Integrate SessionInfo component into TaskDetail
- [ ] **5.9.3** Integrate ResumeModal with state management
- [ ] **5.9.4** Add StatusBadge and PriorityBadge components if not existing

### Testing
- [ ] **5.10.1** Create `ui/src/components/TaskDetail/CommentsPanel.test.tsx`
- [ ] **5.10.2** Write test: loading state displays correctly
- [ ] **5.10.3** Write test: empty state shows when no comments
- [ ] **5.10.4** Write test: comments display with author, time, content
- [ ] **5.10.5** Write test: add comment form works
- [ ] **5.10.6** Write test: @agent warning shows in textarea
- [ ] **5.10.7** Create `ui/src/components/TaskDetail/SessionInfo.test.tsx`
- [ ] **5.10.8** Create `ui/src/components/TaskDetail/ResumeModal.test.tsx`

### Quality & Polish
- [ ] **5.11.1** Ensure all new components support dark mode
- [ ] **5.11.2** Add proper ARIA labels for accessibility
- [ ] **5.11.3** Add keyboard navigation support
- [ ] **5.11.4** Test on mobile viewport sizes (responsive design)
- [ ] **5.11.5** Handle PocketBase SSE connection errors gracefully
- [ ] **5.11.6** Add React.memo for comment items if rendering many comments

### Verification
- [ ] **5.12.1** Test: Real-time updates work (add comment in CLI, see in UI)
- [ ] **5.12.2** Test: Resume button appears only for need_input tasks
- [ ] **5.12.3** Test: Resume modal generates correct command for each tool
- [ ] **5.12.4** Test: Copy button works in resume modal
- [ ] **5.12.5** Test: Tasks draggable to/from need_input column

---

## Overview

This phase implements the Web UI components for the collaborative workflow. Humans need a visual interface to see blocked tasks, read agent questions, and add response comments.

**What we're building:**
- Comments panel on task detail view
- Session info display
- Resume button for blocked tasks
- Visual indicator for tasks needing input
- Real-time comment updates via PocketBase subscriptions

---

## Prerequisites

Before starting this phase:

1. Phase 4 is complete (tool integrations work)
2. Comments collection exists and has data
3. Sessions collection exists
4. Familiar with the existing React/TypeScript UI structure
5. PocketBase real-time subscriptions work

---

## UI Components Overview

```
┌────────────────────────────────────────────────────────────────┐
│                         Task Detail                             │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Title: WRK-123 - Implement authentication                     │
│  Status: need_input    Priority: high                          │
│  Description: ...                                               │
│                                                                 │
├────────────────────────────────────────────────────────────────┤
│  Agent Session                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ ⚡ opencode                                             │    │
│  │ Session: abc123...                                      │    │
│  │ Linked: 2 hours ago                                     │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
│  [Resume Agent Session]                                         │
│                                                                 │
├────────────────────────────────────────────────────────────────┤
│  Comments (3)                                                   │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ 🤖 Agent  10:30                                         │    │
│  │ What authentication approach should I use?              │    │
│  └────────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ 👤 john  11:45                                          │    │
│  │ Use JWT with refresh tokens. The refresh token should   │    │
│  │ have a 7-day expiry...                                  │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ Add a comment...                                        │    │
│  │                                                    [Add] │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Tasks

### Task 5.1: Create Comment Type Definitions

**File**: `ui/src/types/comment.ts`

```typescript
export interface Comment {
  id: string;
  task: string;
  content: string;
  author_type: 'human' | 'agent';
  author_id?: string;
  metadata?: {
    mentions?: string[];
    action?: string;
  };
  created: string;
}

export interface CreateCommentInput {
  task: string;
  content: string;
  author_type: 'human' | 'agent';
  author_id?: string;
}
```

**File**: `ui/src/types/session.ts`

```typescript
export interface AgentSession {
  tool: 'opencode' | 'claude-code' | 'codex';
  ref: string;
  ref_type: 'uuid' | 'path';
  working_dir: string;
  linked_at: string;
}

export interface SessionRecord {
  id: string;
  task: string;
  tool: string;
  external_ref: string;
  ref_type: string;
  working_dir: string;
  status: 'active' | 'paused' | 'completed' | 'abandoned';
  created: string;
  ended_at?: string;
}
```

---

### Task 5.2: Create Comments Hook

**File**: `ui/src/hooks/useComments.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { pb } from '@/lib/pocketbase';
import type { Comment, CreateCommentInput } from '@/types/comment';

// Fetch comments for a task
export function useComments(taskId: string) {
  return useQuery({
    queryKey: ['comments', taskId],
    queryFn: async () => {
      const records = await pb.collection('comments').getFullList<Comment>({
        filter: `task = "${taskId}"`,
        sort: '+created', // Oldest first
      });
      return records;
    },
    enabled: !!taskId,
  });
}

// Add a comment
export function useAddComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateCommentInput) => {
      // Extract mentions from content
      const mentions = extractMentions(input.content);

      const record = await pb.collection('comments').create({
        ...input,
        metadata: { mentions },
      });

      return record;
    },
    onSuccess: (_, variables) => {
      // Invalidate comments query to refetch
      queryClient.invalidateQueries({ queryKey: ['comments', variables.task] });
    },
  });
}

// Real-time subscription for comments
export function useCommentsSubscription(taskId: string) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!taskId) return;

    // Subscribe to comments collection
    const unsubscribe = pb.collection('comments').subscribe('*', (event) => {
      // Only handle events for this task
      if (event.record.task === taskId) {
        // Invalidate to trigger refetch
        queryClient.invalidateQueries({ queryKey: ['comments', taskId] });
      }
    });

    return () => {
      unsubscribe.then((unsub) => unsub());
    };
  }, [taskId, queryClient]);
}

// Extract @mentions from text
function extractMentions(text: string): string[] {
  const matches = text.match(/@\w+/g) || [];
  return [...new Set(matches)]; // Dedupe
}
```

---

### Task 5.3: Create Comments Panel Component

**File**: `ui/src/components/TaskDetail/CommentsPanel.tsx`

```typescript
import React, { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { useComments, useAddComment, useCommentsSubscription } from '@/hooks/useComments';
import type { Comment } from '@/types/comment';

interface CommentsPanelProps {
  taskId: string;
}

export function CommentsPanel({ taskId }: CommentsPanelProps) {
  const { data: comments = [], isLoading, error } = useComments(taskId);
  const addComment = useAddComment();
  const [newComment, setNewComment] = useState('');

  // Subscribe to real-time updates
  useCommentsSubscription(taskId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const trimmed = newComment.trim();
    if (!trimmed) return;

    try {
      await addComment.mutateAsync({
        task: taskId,
        content: trimmed,
        author_type: 'human',
        // author_id could come from auth context
      });
      setNewComment('');
    } catch (err) {
      console.error('Failed to add comment:', err);
    }
  };

  if (isLoading) {
    return (
      <div className="border-t border-gray-200 dark:border-gray-700 p-4">
        <div className="animate-pulse space-y-3">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4"></div>
          <div className="h-20 bg-gray-200 dark:bg-gray-700 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="border-t border-gray-200 dark:border-gray-700 p-4">
        <p className="text-red-500 text-sm">Failed to load comments</p>
      </div>
    );
  }

  return (
    <div className="border-t border-gray-200 dark:border-gray-700">
      {/* Header */}
      <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-800">
        <h3 className="font-semibold text-sm text-gray-700 dark:text-gray-300">
          Comments ({comments.length})
        </h3>
      </div>

      {/* Comments list */}
      <div className="max-h-80 overflow-y-auto">
        {comments.length === 0 ? (
          <div className="px-4 py-8 text-center text-gray-500 text-sm">
            No comments yet
          </div>
        ) : (
          comments.map((comment) => (
            <CommentItem key={comment.id} comment={comment} />
          ))
        )}
      </div>

      {/* Add comment form */}
      <form onSubmit={handleSubmit} className="p-4 border-t border-gray-100 dark:border-gray-800">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Add a comment... (use @agent to trigger auto-resume)"
          className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                     rounded-lg resize-none bg-white dark:bg-gray-800
                     focus:ring-2 focus:ring-blue-500 focus:border-transparent
                     placeholder-gray-400 dark:placeholder-gray-500"
          rows={3}
        />
        <div className="flex justify-between items-center mt-2">
          <span className="text-xs text-gray-400">
            {newComment.includes('@agent') && (
              <span className="text-orange-500">
                Will trigger auto-resume (if enabled)
              </span>
            )}
          </span>
          <button
            type="submit"
            disabled={!newComment.trim() || addComment.isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 
                       rounded-lg hover:bg-blue-700 disabled:opacity-50 
                       disabled:cursor-not-allowed transition-colors"
          >
            {addComment.isPending ? 'Adding...' : 'Add Comment'}
          </button>
        </div>
      </form>
    </div>
  );
}

// Individual comment component
function CommentItem({ comment }: { comment: Comment }) {
  const isAgent = comment.author_type === 'agent';
  const authorDisplay = comment.author_id || comment.author_type;
  const timeAgo = formatDistanceToNow(new Date(comment.created), { addSuffix: true });

  return (
    <div
      className={`px-4 py-3 border-b border-gray-100 dark:border-gray-800 
                  ${isAgent ? 'bg-blue-50 dark:bg-blue-900/10' : ''}`}
    >
      {/* Header */}
      <div className="flex items-center gap-2 mb-2">
        {/* Author badge */}
        <span
          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium
                      ${isAgent
                        ? 'bg-blue-100 text-blue-700 dark:bg-blue-800 dark:text-blue-200'
                        : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-200'
                      }`}
        >
          {isAgent ? '🤖' : '👤'} {authorDisplay}
        </span>
        
        {/* Timestamp */}
        <span className="text-xs text-gray-400">{timeAgo}</span>
        
        {/* Mentions */}
        {comment.metadata?.mentions?.map((mention) => (
          <span
            key={mention}
            className="text-xs text-orange-500 font-medium"
          >
            {mention}
          </span>
        ))}
      </div>

      {/* Content */}
      <div className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
        {comment.content}
      </div>
    </div>
  );
}
```

---

### Task 5.4: Create Session Info Component

**File**: `ui/src/components/TaskDetail/SessionInfo.tsx`

```typescript
import React from 'react';
import { formatDistanceToNow } from 'date-fns';
import type { AgentSession } from '@/types/session';

interface SessionInfoProps {
  session: AgentSession | null;
  taskColumn: string;
  displayId: string;
  onResume?: () => void;
}

const toolConfig: Record<string, { icon: string; name: string; color: string }> = {
  opencode: { icon: '⚡', name: 'OpenCode', color: 'text-yellow-500' },
  'claude-code': { icon: '🤖', name: 'Claude Code', color: 'text-purple-500' },
  codex: { icon: '🔮', name: 'Codex', color: 'text-green-500' },
};

export function SessionInfo({ session, taskColumn, displayId, onResume }: SessionInfoProps) {
  if (!session) {
    return null;
  }

  const tool = toolConfig[session.tool] || { icon: '🔧', name: session.tool, color: 'text-gray-500' };
  const linkedAgo = formatDistanceToNow(new Date(session.linked_at), { addSuffix: true });
  const isBlocked = taskColumn === 'need_input';

  return (
    <div className="border-t border-gray-200 dark:border-gray-700">
      <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800/50">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
          Agent Session
        </h4>

        <div className="flex items-start gap-3">
          {/* Tool icon */}
          <span className={`text-2xl ${tool.color}`}>{tool.icon}</span>

          {/* Session details */}
          <div className="flex-1 min-w-0">
            <div className="font-medium text-gray-900 dark:text-gray-100">
              {tool.name}
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400 font-mono truncate">
              {truncateMiddle(session.ref, 30)}
            </div>
            <div className="text-xs text-gray-400 mt-1">
              Linked {linkedAgo}
            </div>
          </div>

          {/* Status indicator */}
          <div className={`px-2 py-1 rounded text-xs font-medium
                          ${isBlocked 
                            ? 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' 
                            : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                          }`}>
            {isBlocked ? 'Blocked' : 'Active'}
          </div>
        </div>

        {/* Resume button */}
        {isBlocked && onResume && (
          <button
            onClick={onResume}
            className="w-full mt-4 px-4 py-2 bg-green-600 text-white rounded-lg 
                       hover:bg-green-700 transition-colors flex items-center 
                       justify-center gap-2 font-medium"
          >
            <PlayIcon className="w-4 h-4" />
            Resume Agent Session
          </button>
        )}
      </div>
    </div>
  );
}

// Simple play icon
function PlayIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 20 20" fill="currentColor">
      <path
        fillRule="evenodd"
        d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function truncateMiddle(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str;
  const half = Math.floor((maxLen - 3) / 2);
  return `${str.slice(0, half)}...${str.slice(-half)}`;
}
```

---

### Task 5.5: Create Resume Hook

**File**: `ui/src/hooks/useResume.ts`

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { pb } from '@/lib/pocketbase';

interface ResumeInput {
  taskId: string;
  displayId: string;
  exec?: boolean;
}

interface ResumeResult {
  command: string;
  prompt: string;
  tool: string;
  sessionRef: string;
  workingDir: string;
}

export function useResume() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, displayId, exec }: ResumeInput): Promise<ResumeResult> => {
      // Call the resume endpoint/function
      // This could be a custom API endpoint or exec a CLI command
      
      // For now, we'll construct the resume info client-side
      // In a real implementation, you might call a backend endpoint
      
      const task = await pb.collection('tasks').getOne(taskId);
      const session = task.agent_session;
      
      if (!session) {
        throw new Error('No session linked to this task');
      }
      
      // Fetch comments for context
      const comments = await pb.collection('comments').getFullList({
        filter: `task = "${taskId}"`,
        sort: '+created',
      });
      
      // Build prompt (simplified - full version would match CLI)
      const prompt = buildContextPrompt(task, comments);
      
      // Build command
      const command = buildResumeCommand(session.tool, session.ref, prompt);
      
      if (exec) {
        // Update task to in_progress
        await pb.collection('tasks').update(taskId, {
          column: 'in_progress',
        });
        
        // In a real implementation, you might:
        // 1. Call a backend endpoint that executes the command
        // 2. Open a terminal with the command
        // 3. Copy to clipboard and show instructions
      }
      
      return {
        command,
        prompt,
        tool: session.tool,
        sessionRef: session.ref,
        workingDir: session.working_dir,
      };
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tasks', variables.taskId] });
    },
  });
}

function buildContextPrompt(task: any, comments: any[]): string {
  let prompt = `## Task Context\n\n`;
  prompt += `**Task**: ${task.display_id || task.id.slice(0, 8)} - ${task.title}\n`;
  prompt += `**Priority**: ${task.priority}\n\n`;
  
  prompt += `## Comments\n\n`;
  for (const c of comments) {
    const author = c.author_id || c.author_type;
    prompt += `[${author}]: ${c.content}\n\n`;
  }
  
  prompt += `## Instructions\n\n`;
  prompt += `Continue working based on the above context.\n`;
  
  return prompt;
}

function buildResumeCommand(tool: string, ref: string, prompt: string): string {
  // Escape prompt for shell (simplified)
  const escaped = prompt.replace(/'/g, "'\\''");
  
  switch (tool) {
    case 'opencode':
      return `opencode run '${escaped}' --session ${ref}`;
    case 'claude-code':
      return `claude --resume ${ref} '${escaped}'`;
    case 'codex':
      return `codex exec resume ${ref} '${escaped}'`;
    default:
      return `# Unknown tool: ${tool}`;
  }
}
```

---

### Task 5.6: Create Resume Modal

**File**: `ui/src/components/TaskDetail/ResumeModal.tsx`

```typescript
import React, { useState } from 'react';
import { useResume } from '@/hooks/useResume';

interface ResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
  taskId: string;
  displayId: string;
}

export function ResumeModal({ isOpen, onClose, taskId, displayId }: ResumeModalProps) {
  const resume = useResume();
  const [result, setResult] = useState<any>(null);
  const [copied, setCopied] = useState(false);

  if (!isOpen) return null;

  const handleGenerate = async () => {
    try {
      const res = await resume.mutateAsync({ taskId, displayId, exec: false });
      setResult(res);
    } catch (err) {
      console.error('Failed to generate resume command:', err);
    }
  };

  const handleCopy = () => {
    if (result?.command) {
      navigator.clipboard.writeText(result.command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />
      
      {/* Modal */}
      <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl 
                      max-w-2xl w-full mx-4 max-h-[80vh] overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Resume Session for {displayId}
          </h2>
        </div>
        
        {/* Content */}
        <div className="px-6 py-4 overflow-y-auto max-h-[60vh]">
          {!result ? (
            <div className="text-center py-8">
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                Generate the command to resume the agent session with full context.
              </p>
              <button
                onClick={handleGenerate}
                disabled={resume.isPending}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg 
                           hover:bg-blue-700 disabled:opacity-50"
              >
                {resume.isPending ? 'Generating...' : 'Generate Resume Command'}
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Tool info */}
              <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                <span>Tool:</span>
                <span className="font-medium text-gray-900 dark:text-gray-100">
                  {result.tool}
                </span>
              </div>
              
              {/* Command */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Resume Command
                </label>
                <div className="relative">
                  <pre className="bg-gray-100 dark:bg-gray-900 p-4 rounded-lg 
                                  text-sm overflow-x-auto font-mono text-gray-800 
                                  dark:text-gray-200">
                    {result.command}
                  </pre>
                  <button
                    onClick={handleCopy}
                    className="absolute top-2 right-2 px-3 py-1 bg-gray-200 
                               dark:bg-gray-700 rounded text-xs hover:bg-gray-300
                               dark:hover:bg-gray-600"
                  >
                    {copied ? 'Copied!' : 'Copy'}
                  </button>
                </div>
              </div>
              
              {/* Working directory */}
              <div className="text-sm">
                <span className="text-gray-500">Working directory: </span>
                <code className="text-gray-700 dark:text-gray-300">
                  {result.workingDir}
                </code>
              </div>
              
              {/* Instructions */}
              <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
                <h4 className="font-medium text-blue-800 dark:text-blue-200 mb-2">
                  Instructions
                </h4>
                <ol className="text-sm text-blue-700 dark:text-blue-300 space-y-1 list-decimal list-inside">
                  <li>Copy the command above</li>
                  <li>Open a terminal in: {result.workingDir}</li>
                  <li>Paste and run the command</li>
                  <li>The agent will resume with full context</li>
                </ol>
              </div>
            </div>
          )}
        </div>
        
        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 
                        flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 dark:text-gray-300 
                       hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
```

---

### Task 5.7: Update Task Card with Need Input Indicator

**File**: `ui/src/components/TaskCard.tsx` (modify existing)

Add a visual indicator for tasks in `need_input` state:

```typescript
// Add to TaskCard component

interface TaskCardProps {
  task: Task;
  // ... other props
}

export function TaskCard({ task, ...props }: TaskCardProps) {
  const isBlocked = task.column === 'need_input';
  
  return (
    <div className={`relative ... ${isBlocked ? 'ring-2 ring-orange-400' : ''}`}>
      {/* Blocked indicator */}
      {isBlocked && (
        <div className="absolute -top-1 -right-1">
          <span className="relative flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full 
                             rounded-full bg-orange-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 
                             bg-orange-500"></span>
          </span>
        </div>
      )}
      
      {/* Blocked label */}
      {isBlocked && (
        <div className="absolute top-2 left-2">
          <span className="px-2 py-0.5 bg-orange-100 text-orange-700 
                           dark:bg-orange-900/30 dark:text-orange-400 
                           rounded text-xs font-medium">
            Needs Input
          </span>
        </div>
      )}
      
      {/* Rest of card content */}
      {/* ... */}
    </div>
  );
}
```

---

### Task 5.8: Update Kanban Board for Need Input Column

**File**: `ui/src/components/KanbanBoard.tsx` (modify existing)

Add styling for the `need_input` column:

```typescript
const columnConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  backlog: { label: 'Backlog', color: 'text-gray-600', bgColor: 'bg-gray-100' },
  todo: { label: 'To Do', color: 'text-blue-600', bgColor: 'bg-blue-50' },
  in_progress: { label: 'In Progress', color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
  need_input: { label: 'Needs Input', color: 'text-orange-600', bgColor: 'bg-orange-50' },
  review: { label: 'Review', color: 'text-purple-600', bgColor: 'bg-purple-50' },
  done: { label: 'Done', color: 'text-green-600', bgColor: 'bg-green-50' },
};

// In the column header:
<div className={`px-3 py-2 rounded-t-lg ${columnConfig[column].bgColor}`}>
  <h3 className={`font-semibold ${columnConfig[column].color} flex items-center gap-2`}>
    {columnConfig[column].label}
    {column === 'need_input' && (
      <span className="animate-pulse text-orange-500">●</span>
    )}
    <span className="text-gray-400 font-normal">({tasks.length})</span>
  </h3>
</div>
```

---

### Task 5.9: Integrate Components into Task Detail

**File**: `ui/src/components/TaskDetail/index.tsx` (modify existing)

```typescript
import React, { useState } from 'react';
import { CommentsPanel } from './CommentsPanel';
import { SessionInfo } from './SessionInfo';
import { ResumeModal } from './ResumeModal';
import type { Task } from '@/types/task';

interface TaskDetailProps {
  task: Task;
  onClose: () => void;
}

export function TaskDetail({ task, onClose }: TaskDetailProps) {
  const [showResumeModal, setShowResumeModal] = useState(false);
  
  const displayId = task.display_id || `WRK-${task.seq}`;
  
  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {displayId}: {task.title}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <XIcon className="w-5 h-5" />
          </button>
        </div>
        
        {/* Status badges */}
        <div className="flex items-center gap-2 mt-2">
          <StatusBadge column={task.column} />
          <PriorityBadge priority={task.priority} />
        </div>
      </div>
      
      {/* Content - scrollable */}
      <div className="flex-1 overflow-y-auto">
        {/* Description */}
        {task.description && (
          <div className="px-6 py-4">
            <h4 className="text-sm font-medium text-gray-500 mb-2">Description</h4>
            <p className="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
              {task.description}
            </p>
          </div>
        )}
        
        {/* Session info */}
        <SessionInfo
          session={task.agent_session}
          taskColumn={task.column}
          displayId={displayId}
          onResume={() => setShowResumeModal(true)}
        />
        
        {/* Comments */}
        <CommentsPanel taskId={task.id} />
      </div>
      
      {/* Resume modal */}
      <ResumeModal
        isOpen={showResumeModal}
        onClose={() => setShowResumeModal(false)}
        taskId={task.id}
        displayId={displayId}
      />
    </div>
  );
}
```

---

### Task 5.10: Write Component Tests

**File**: `ui/src/components/TaskDetail/CommentsPanel.test.tsx`

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CommentsPanel } from './CommentsPanel';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

describe('CommentsPanel', () => {
  beforeEach(() => {
    queryClient.clear();
  });

  it('shows loading state', () => {
    render(<CommentsPanel taskId="test-123" />, { wrapper });
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('shows empty state when no comments', async () => {
    // Mock empty response
    jest.spyOn(global, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: async () => ({ items: [] }),
    } as Response);

    render(<CommentsPanel taskId="test-123" />, { wrapper });
    
    await waitFor(() => {
      expect(screen.getByText(/no comments yet/i)).toBeInTheDocument();
    });
  });

  it('displays comments', async () => {
    const mockComments = [
      {
        id: '1',
        content: 'Test comment',
        author_type: 'human',
        author_id: 'john',
        created: new Date().toISOString(),
      },
    ];

    jest.spyOn(global, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: async () => ({ items: mockComments }),
    } as Response);

    render(<CommentsPanel taskId="test-123" />, { wrapper });

    await waitFor(() => {
      expect(screen.getByText('Test comment')).toBeInTheDocument();
      expect(screen.getByText('john')).toBeInTheDocument();
    });
  });

  it('adds new comment', async () => {
    render(<CommentsPanel taskId="test-123" />, { wrapper });

    const textarea = screen.getByPlaceholderText(/add a comment/i);
    const submitButton = screen.getByText(/add comment/i);

    fireEvent.change(textarea, { target: { value: 'New comment' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(textarea).toHaveValue(''); // Should clear after submit
    });
  });

  it('shows @agent warning', () => {
    render(<CommentsPanel taskId="test-123" />, { wrapper });

    const textarea = screen.getByPlaceholderText(/add a comment/i);
    fireEvent.change(textarea, { target: { value: '@agent please continue' } });

    expect(screen.getByText(/auto-resume/i)).toBeInTheDocument();
  });
});
```

---

## Testing Checklist

Before considering this phase complete:

### Comments Panel

- [ ] Loading state displays correctly
- [ ] Empty state shows when no comments
- [ ] Comments display with author, time, content
- [ ] Agent comments have distinct styling
- [ ] Human comments have distinct styling
- [ ] @mentions are highlighted
- [ ] Add comment form works
- [ ] @agent warning shows in textarea
- [ ] Real-time updates work (add comment in CLI, see in UI)

### Session Info

- [ ] Shows nothing when no session
- [ ] Displays tool icon and name
- [ ] Shows session ref (truncated)
- [ ] Shows linked time
- [ ] Shows blocked/active status
- [ ] Resume button appears for need_input tasks
- [ ] Resume button hidden for other states

### Resume Modal

- [ ] Opens when clicking resume button
- [ ] Generates command correctly
- [ ] Copy button works
- [ ] Shows working directory
- [ ] Shows instructions
- [ ] Closes properly

### Task Card

- [ ] Needs input indicator (pulsing dot)
- [ ] Needs input label
- [ ] Orange ring around blocked tasks

### Kanban Board

- [ ] need_input column styled distinctly
- [ ] Column has pulsing indicator
- [ ] Tasks draggable to/from need_input

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `ui/src/types/comment.ts` | New | Comment types |
| `ui/src/types/session.ts` | New | Session types |
| `ui/src/hooks/useComments.ts` | New | Comments hook |
| `ui/src/hooks/useResume.ts` | New | Resume hook |
| `ui/src/components/TaskDetail/CommentsPanel.tsx` | New | Comments panel |
| `ui/src/components/TaskDetail/SessionInfo.tsx` | New | Session info |
| `ui/src/components/TaskDetail/ResumeModal.tsx` | New | Resume modal |
| `ui/src/components/TaskDetail/index.tsx` | Modified | Integrate new components |
| `ui/src/components/TaskCard.tsx` | Modified | Need input indicator |
| `ui/src/components/KanbanBoard.tsx` | Modified | Need input column |
| `ui/src/components/TaskDetail/*.test.tsx` | New | Component tests |

---

## Notes for Implementer

1. **Styling consistency**: Match the existing UI design system (colors, spacing, typography).

2. **Dark mode**: Ensure all new components support dark mode.

3. **Accessibility**: Add proper ARIA labels and keyboard navigation.

4. **Real-time subscriptions**: PocketBase subscriptions use SSE. Handle connection errors gracefully.

5. **Mobile responsiveness**: Test on mobile viewport sizes.

6. **Performance**: Use React.memo for comment items if rendering many comments.
