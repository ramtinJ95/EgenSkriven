package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubtaskIndicator_Render(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expanded bool
		contains string
		empty    bool
	}{
		{
			name:   "no subtasks",
			count:  0,
			empty:  true,
		},
		{
			name:     "collapsed with 3 subtasks",
			count:    3,
			expanded: false,
			contains: "+3",
		},
		{
			name:     "expanded with 3 subtasks",
			count:    3,
			expanded: true,
			contains: "-3",
		},
		{
			name:     "collapsed with 1 subtask",
			count:    1,
			expanded: false,
			contains: "+1",
		},
		{
			name:     "collapsed with 10 subtasks",
			count:    10,
			expanded: false,
			contains: "+10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator := SubtaskIndicator{
				Count:    tt.count,
				Expanded: tt.expanded,
			}
			result := indicator.Render()
			if tt.empty {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestExpandedSubtaskView(t *testing.T) {
	view := NewExpandedSubtaskView()

	// Initially not expanded
	assert.False(t, view.IsExpanded("task-1"))

	// Toggle expands
	view.Toggle("task-1")
	assert.True(t, view.IsExpanded("task-1"))

	// Toggle again collapses
	view.Toggle("task-1")
	assert.False(t, view.IsExpanded("task-1"))

	// Set subtasks
	subtasks := []TaskItem{
		{ID: "sub-1", TaskTitle: "Subtask 1"},
		{ID: "sub-2", TaskTitle: "Subtask 2"},
	}
	view.SetSubtasks("task-1", subtasks)
	assert.Len(t, view.GetSubtasks("task-1"), 2)

	// Clear resets everything
	view.Toggle("task-1")
	assert.True(t, view.IsExpanded("task-1"))
	view.Clear()
	assert.False(t, view.IsExpanded("task-1"))
	assert.Empty(t, view.GetSubtasks("task-1"))
}

func TestBuildSubtaskTree(t *testing.T) {
	tasks := []TaskItem{
		{ID: "root-1", TaskTitle: "Root 1", ParentID: ""},
		{ID: "root-2", TaskTitle: "Root 2", ParentID: ""},
		{ID: "child-1-1", TaskTitle: "Child 1-1", ParentID: "root-1"},
		{ID: "child-1-2", TaskTitle: "Child 1-2", ParentID: "root-1"},
		{ID: "child-2-1", TaskTitle: "Child 2-1", ParentID: "root-2"},
		{ID: "grandchild-1-1-1", TaskTitle: "Grandchild 1-1-1", ParentID: "child-1-1"},
	}

	roots := BuildSubtaskTree(tasks)

	// Should have 2 root nodes
	assert.Len(t, roots, 2)

	// Find root-1
	var root1 *SubtaskTreeNode
	for i := range roots {
		if roots[i].Task.ID == "root-1" {
			root1 = &roots[i]
			break
		}
	}
	assert.NotNil(t, root1)

	// root-1 should have 2 children
	assert.Len(t, root1.Children, 2)

	// Find child-1-1 (should have grandchild)
	var child11 *SubtaskTreeNode
	for i := range root1.Children {
		if root1.Children[i].Task.ID == "child-1-1" {
			child11 = &root1.Children[i]
			break
		}
	}
	assert.NotNil(t, child11)
	assert.Len(t, child11.Children, 1)
	assert.Equal(t, "grandchild-1-1-1", child11.Children[0].Task.ID)
}

func TestFilterRootTasks(t *testing.T) {
	tasks := []TaskItem{
		{ID: "root-1", TaskTitle: "Root 1", ParentID: ""},
		{ID: "root-2", TaskTitle: "Root 2", ParentID: ""},
		{ID: "child-1", TaskTitle: "Child 1", ParentID: "root-1"},
		{ID: "child-2", TaskTitle: "Child 2", ParentID: "root-2"},
	}

	roots := FilterRootTasks(tasks)

	assert.Len(t, roots, 2)
	for _, r := range roots {
		assert.Empty(t, r.ParentID)
	}
}

func TestApplySubtaskCounts(t *testing.T) {
	tasks := []TaskItem{
		{ID: "task-1", TaskTitle: "Task 1"},
		{ID: "task-2", TaskTitle: "Task 2"},
		{ID: "task-3", TaskTitle: "Task 3"},
	}

	counts := map[string]int{
		"task-1": 3,
		"task-3": 1,
	}

	result := ApplySubtaskCounts(tasks, counts)

	// task-1 should have 3 subtasks
	assert.Equal(t, 3, result[0].SubtaskCount)
	assert.True(t, result[0].HasSubtasks)

	// task-2 should have 0 subtasks
	assert.Equal(t, 0, result[1].SubtaskCount)
	assert.False(t, result[1].HasSubtasks)

	// task-3 should have 1 subtask
	assert.Equal(t, 1, result[2].SubtaskCount)
	assert.True(t, result[2].HasSubtasks)
}

func TestSubtaskInfo_RenderIndicator(t *testing.T) {
	tests := []struct {
		name     string
		info     SubtaskInfo
		contains string
		empty    bool
	}{
		{
			name:  "no subtasks",
			info:  SubtaskInfo{ParentID: "task-1", Count: 0},
			empty: true,
		},
		{
			name:     "collapsed",
			info:     SubtaskInfo{ParentID: "task-1", Count: 5, Expanded: false},
			contains: "+5",
		},
		{
			name:     "expanded",
			info:     SubtaskInfo{ParentID: "task-1", Count: 5, Expanded: true},
			contains: "-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.RenderIndicator()
			if tt.empty {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestRenderSubtaskItem(t *testing.T) {
	task := TaskItem{
		ID:        "sub-1",
		DisplayID: "WRK-42",
		TaskTitle: "Fix login bug",
		Column:    "in_progress",
	}

	result := RenderSubtaskItem(task, 1)

	assert.Contains(t, result, "WRK-42")
	assert.Contains(t, result, "Fix login bug")
	assert.Contains(t, result, "in_progress")
	assert.Contains(t, result, "|--")
}

func TestRenderSubtaskTree(t *testing.T) {
	nodes := []SubtaskTreeNode{
		{
			Task:  TaskItem{DisplayID: "WRK-1", TaskTitle: "Parent Task", Column: "todo"},
			Depth: 0,
			Children: []SubtaskTreeNode{
				{
					Task:  TaskItem{DisplayID: "WRK-2", TaskTitle: "Child Task", Column: "done"},
					Depth: 1,
				},
			},
		},
	}

	result := RenderSubtaskTree(nodes, 0)

	assert.Contains(t, result, "WRK-1")
	assert.Contains(t, result, "Parent Task")
	assert.Contains(t, result, "WRK-2")
	assert.Contains(t, result, "Child Task")
	assert.Contains(t, result, "[todo]")
	assert.Contains(t, result, "[done]")
}

func TestTaskItem_TitleWithSubtaskIndicator(t *testing.T) {
	// Task without subtasks
	taskNoSubs := TaskItem{
		DisplayID:    "WRK-1",
		TaskTitle:    "No subtasks",
		HasSubtasks:  false,
		SubtaskCount: 0,
	}
	titleNoSubs := taskNoSubs.Title()
	assert.NotContains(t, titleNoSubs, "[+")
	assert.NotContains(t, titleNoSubs, "[-")

	// Task with collapsed subtasks
	taskCollapsed := TaskItem{
		DisplayID:        "WRK-2",
		TaskTitle:        "Has subtasks",
		HasSubtasks:      true,
		SubtaskCount:     3,
		SubtasksExpanded: false,
	}
	titleCollapsed := taskCollapsed.Title()
	assert.Contains(t, titleCollapsed, "+3")

	// Task with expanded subtasks
	taskExpanded := TaskItem{
		DisplayID:        "WRK-3",
		TaskTitle:        "Has subtasks expanded",
		HasSubtasks:      true,
		SubtaskCount:     2,
		SubtasksExpanded: true,
	}
	titleExpanded := taskExpanded.Title()
	assert.Contains(t, titleExpanded, "-2")
}
