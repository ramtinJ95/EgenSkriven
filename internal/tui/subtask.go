package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// SubtaskIndicator renders the expandable subtask indicator.
type SubtaskIndicator struct {
	Count    int
	Expanded bool
}

// Render creates the subtask indicator string.
func (s SubtaskIndicator) Render() string {
	if s.Count == 0 {
		return ""
	}

	icon := "+"
	if s.Expanded {
		icon = "-"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render(fmt.Sprintf("[%s%d]", icon, s.Count))
}

// GetSubtaskCounts returns a map of parentID -> subtask count for all tasks.
// This is more efficient than counting one at a time.
func GetSubtaskCounts(app *pocketbase.PocketBase, boardID string) (map[string]int, error) {
	counts := make(map[string]int)

	// Query all tasks with parents in this board
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board} AND parent != '' AND parent IS NOT NULL",
			dbx.Params{"board": boardID}),
	)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		parentID := r.GetString("parent")
		if parentID != "" {
			counts[parentID]++
		}
	}

	return counts, nil
}

// LoadSubtasks loads child tasks for a parent task.
func LoadSubtasks(app *pocketbase.PocketBase, parentID, boardPrefix string) ([]TaskItem, error) {
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("parent = {:parent}", dbx.Params{"parent": parentID}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %w", err)
	}

	var subtasks []TaskItem
	for _, r := range records {
		displayID := fmt.Sprintf("%s-%d", boardPrefix, r.GetInt("seq"))
		subtasks = append(subtasks, NewTaskItemFromRecord(r, displayID))
	}

	return subtasks, nil
}

// SubtaskTreeNode represents a task in the subtask hierarchy.
type SubtaskTreeNode struct {
	Task     TaskItem
	Children []SubtaskTreeNode
	Depth    int
}

// BuildSubtaskTree builds a tree from flat task list.
func BuildSubtaskTree(tasks []TaskItem) []SubtaskTreeNode {
	// Create lookup map
	byID := make(map[string]TaskItem)
	for _, t := range tasks {
		byID[t.ID] = t
	}

	// Find root tasks (no parent or parent not in list)
	var roots []SubtaskTreeNode
	childrenOf := make(map[string][]TaskItem)

	for _, t := range tasks {
		if t.ParentID == "" {
			roots = append(roots, SubtaskTreeNode{Task: t, Depth: 0})
		} else {
			childrenOf[t.ParentID] = append(childrenOf[t.ParentID], t)
		}
	}

	// Build tree recursively
	var buildChildren func(parentID string, depth int) []SubtaskTreeNode
	buildChildren = func(parentID string, depth int) []SubtaskTreeNode {
		children := childrenOf[parentID]
		var nodes []SubtaskTreeNode
		for _, child := range children {
			node := SubtaskTreeNode{
				Task:     child,
				Depth:    depth,
				Children: buildChildren(child.ID, depth+1),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	// Add children to roots
	for i := range roots {
		roots[i].Children = buildChildren(roots[i].Task.ID, 1)
	}

	return roots
}

// RenderSubtaskTree renders the subtask tree with indentation.
func RenderSubtaskTree(nodes []SubtaskTreeNode, indent int) string {
	var lines []string

	indentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	for _, node := range nodes {
		prefix := strings.Repeat("  ", indent+node.Depth)
		connector := indentStyle.Render("|-- ")

		taskLine := fmt.Sprintf("%s%s%s %s",
			prefix, connector,
			node.Task.DisplayID,
			Truncate(node.Task.TaskTitle, 40))

		// Add status indicator
		statusColors := map[string]string{
			"done":        "82",  // green
			"in_progress": "214", // orange
			"review":      "205", // pink
			"todo":        "39",  // cyan
			"backlog":     "240", // gray
		}
		color := statusColors[node.Task.Column]
		if color == "" {
			color = "240"
		}
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render("[" + node.Task.Column + "]")

		lines = append(lines, taskLine+" "+status)

		// Render children recursively
		if len(node.Children) > 0 {
			childLines := RenderSubtaskTree(node.Children, indent)
			lines = append(lines, childLines)
		}
	}

	return strings.Join(lines, "\n")
}

// ExpandedSubtaskView tracks which parent tasks have expanded subtasks.
type ExpandedSubtaskView struct {
	expanded map[string]bool       // parentID -> expanded
	subtasks map[string][]TaskItem // cached subtasks
}

// NewExpandedSubtaskView creates a new subtask view tracker.
func NewExpandedSubtaskView() *ExpandedSubtaskView {
	return &ExpandedSubtaskView{
		expanded: make(map[string]bool),
		subtasks: make(map[string][]TaskItem),
	}
}

// IsExpanded returns true if the task's subtasks are visible.
func (v *ExpandedSubtaskView) IsExpanded(taskID string) bool {
	return v.expanded[taskID]
}

// Toggle expands/collapses subtasks for a task.
func (v *ExpandedSubtaskView) Toggle(taskID string) {
	v.expanded[taskID] = !v.expanded[taskID]
}

// SetSubtasks caches loaded subtasks.
func (v *ExpandedSubtaskView) SetSubtasks(parentID string, subtasks []TaskItem) {
	v.subtasks[parentID] = subtasks
}

// GetSubtasks returns cached subtasks.
func (v *ExpandedSubtaskView) GetSubtasks(parentID string) []TaskItem {
	return v.subtasks[parentID]
}

// Clear clears all expanded state.
func (v *ExpandedSubtaskView) Clear() {
	v.expanded = make(map[string]bool)
	v.subtasks = make(map[string][]TaskItem)
}

// RenderSubtaskItem renders a single subtask with indentation for inline display.
func RenderSubtaskItem(task TaskItem, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	connector := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("|-- ")

	// Status color
	statusColors := map[string]string{
		"done":        "82",
		"in_progress": "214",
		"review":      "205",
		"todo":        "39",
		"backlog":     "240",
	}
	color := statusColors[task.Column]
	if color == "" {
		color = "240"
	}
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render("[" + task.Column + "]")

	return fmt.Sprintf("%s%s%s %s %s",
		indentStr, connector,
		task.DisplayID,
		Truncate(task.TaskTitle, 35),
		status)
}

// ApplySubtaskCounts updates tasks with their subtask counts.
func ApplySubtaskCounts(tasks []TaskItem, counts map[string]int) []TaskItem {
	for i := range tasks {
		if count, ok := counts[tasks[i].ID]; ok {
			tasks[i].SubtaskCount = count
			tasks[i].HasSubtasks = count > 0
		}
	}
	return tasks
}

// LoadSubtasksForTask is a tea.Cmd that loads subtasks for a parent task.
func LoadSubtasksForTask(app *pocketbase.PocketBase, parentID, boardPrefix string) func() ([]TaskItem, error) {
	return func() ([]TaskItem, error) {
		return LoadSubtasks(app, parentID, boardPrefix)
	}
}

// SubtaskInfo holds subtask display info for a parent task.
type SubtaskInfo struct {
	ParentID     string
	Count        int
	Expanded     bool
	Subtasks     []TaskItem
	LoadedOnce   bool
}

// NewSubtaskInfo creates new subtask info.
func NewSubtaskInfo(parentID string, count int) SubtaskInfo {
	return SubtaskInfo{
		ParentID: parentID,
		Count:    count,
		Expanded: false,
	}
}

// RenderIndicator renders the subtask indicator badge.
func (s SubtaskInfo) RenderIndicator() string {
	if s.Count == 0 {
		return ""
	}

	icon := "+"
	if s.Expanded {
		icon = "-"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render(fmt.Sprintf("[%s%d]", icon, s.Count))
}

// SubtasksLoadedMsg is sent when subtasks are loaded.
type SubtasksLoadedMsg struct {
	ParentID string
	Subtasks []TaskItem
	Err      error
}

// LoadSubtasksCmd creates a command to load subtasks.
func LoadSubtasksCmd(app *pocketbase.PocketBase, parentID, boardPrefix string) func() SubtasksLoadedMsg {
	return func() SubtasksLoadedMsg {
		subtasks, err := LoadSubtasks(app, parentID, boardPrefix)
		return SubtasksLoadedMsg{
			ParentID: parentID,
			Subtasks: subtasks,
			Err:      err,
		}
	}
}

// FilterRootTasks returns only tasks that have no parent.
func FilterRootTasks(tasks []TaskItem) []TaskItem {
	var roots []TaskItem
	for _, t := range tasks {
		if t.ParentID == "" {
			roots = append(roots, t)
		}
	}
	return roots
}

// LoadTasksWithSubtaskInfo loads tasks with parent info from records.
func LoadTasksWithSubtaskInfo(records []*core.Record, boardPrefix string) []TaskItem {
	var tasks []TaskItem
	for _, r := range records {
		displayID := fmt.Sprintf("%s-%d", boardPrefix, r.GetInt("seq"))
		task := NewTaskItemFromRecord(r, displayID)
		task.ParentID = r.GetString("parent")
		tasks = append(tasks, task)
	}
	return tasks
}
