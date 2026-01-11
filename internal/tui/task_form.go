package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents which field is currently focused
type FormField int

const (
	FieldTitle FormField = iota
	FieldDescription
	FieldType
	FieldPriority
	FieldColumn
	FieldLabels
	FieldDueDate
	FieldSubmit
	FieldCancel
)

const numFields = 9 // Total number of focusable fields

// TaskForm handles task creation and editing
type TaskForm struct {
	mode   FormMode
	taskID string // Only set in edit mode

	// Text inputs
	titleInput   textinput.Model
	descInput    textarea.Model
	labelsInput  textinput.Model
	dueDateInput textinput.Model

	// Select fields (index into options)
	typeSelect     int
	prioritySelect int
	columnSelect   int

	// Options for select fields
	types      []string
	priorities []string
	columns    []string

	// Form state
	focusIndex int
	width      int
	height     int
	keys       taskFormKeyMap
}

type taskFormKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Next   key.Binding
	Prev   key.Binding
	Left   key.Binding
	Right  key.Binding
}

func defaultTaskFormKeyMap() taskFormKeyMap {
	return taskFormKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "prev option"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next option"),
		),
	}
}

// NewTaskForm creates a new task form
func NewTaskForm(mode FormMode, width, height int) *TaskForm {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Task title..."
	ti.CharLimit = 200
	ti.Width = width - 20
	ti.Focus()

	// Description textarea
	ta := textarea.New()
	ta.Placeholder = "Description (markdown supported)..."
	ta.SetWidth(width - 20)
	ta.SetHeight(5)
	ta.CharLimit = 5000

	// Labels input
	li := textinput.New()
	li.Placeholder = "Labels (comma-separated)..."
	li.CharLimit = 200
	li.Width = width - 20

	// Due date input
	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.CharLimit = 10
	di.Width = 15

	return &TaskForm{
		mode:           mode,
		titleInput:     ti,
		descInput:      ta,
		labelsInput:    li,
		dueDateInput:   di,
		types:          []string{"feature", "bug", "chore"},
		priorities:     []string{"low", "medium", "high", "urgent"},
		columns:        []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		typeSelect:     0, // Default: feature
		prioritySelect: 1, // Default: medium
		columnSelect:   0, // Default: backlog
		focusIndex:     0,
		width:          width,
		height:         height,
		keys:           defaultTaskFormKeyMap(),
	}
}

// NewTaskFormWithData creates a form pre-filled with task data (for editing)
func NewTaskFormWithData(task *TaskItem, width, height int) *TaskForm {
	f := NewTaskForm(FormModeEdit, width, height)
	f.taskID = task.ID

	// Pre-fill fields
	f.titleInput.SetValue(task.TaskTitle)
	f.descInput.SetValue(task.TaskDescription)
	f.labelsInput.SetValue(strings.Join(task.Labels, ", "))
	f.dueDateInput.SetValue(task.DueDate)

	// Set select indices
	for i, t := range f.types {
		if t == task.Type {
			f.typeSelect = i
			break
		}
	}
	for i, p := range f.priorities {
		if p == task.Priority {
			f.prioritySelect = i
			break
		}
	}
	for i, c := range f.columns {
		if c == task.Column {
			f.columnSelect = i
			break
		}
	}

	return f
}

// Init initializes the form
func (f *TaskForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the form
func (f *TaskForm) Update(msg tea.Msg) (*TaskForm, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.Cancel):
			return f, func() tea.Msg {
				return closeTaskFormMsg{cancelled: true}
			}

		case key.Matches(msg, f.keys.Submit):
			return f, f.submit()

		case key.Matches(msg, f.keys.Next):
			f.nextField()
			return f, nil

		case key.Matches(msg, f.keys.Prev):
			f.prevField()
			return f, nil

		case key.Matches(msg, f.keys.Left):
			if f.isSelectField() {
				f.selectPrev()
				return f, nil
			}

		case key.Matches(msg, f.keys.Right):
			if f.isSelectField() {
				f.selectNext()
				return f, nil
			}

		case msg.String() == "enter":
			if FormField(f.focusIndex) == FieldSubmit {
				return f, f.submit()
			} else if FormField(f.focusIndex) == FieldCancel {
				return f, func() tea.Msg {
					return closeTaskFormMsg{cancelled: true}
				}
			}
		}

	case tea.WindowSizeMsg:
		f.SetSize(msg.Width/2, msg.Height-10)
	}

	// Update the currently focused input
	cmd := f.updateFocusedInput(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *TaskForm) View() string {
	title := "Add Task"
	if f.mode == FormModeEdit {
		title = "Edit Task"
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render(title)

	// Build form fields
	var fields []string

	fields = append(fields, f.renderField("Title", f.titleInput.View(), FieldTitle))
	fields = append(fields, f.renderField("Description", f.descInput.View(), FieldDescription))
	fields = append(fields, f.renderSelect("Type", f.types, f.typeSelect, FieldType))
	fields = append(fields, f.renderSelect("Priority", f.priorities, f.prioritySelect, FieldPriority))
	fields = append(fields, f.renderSelect("Column", f.columns, f.columnSelect, FieldColumn))
	fields = append(fields, f.renderField("Labels", f.labelsInput.View(), FieldLabels))
	fields = append(fields, f.renderField("Due Date", f.dueDateInput.View(), FieldDueDate))

	// Buttons
	buttons := f.renderButtons()

	// Help text
	help := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("tab:next field  ctrl+s:save  esc:cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(fields, "\n"),
		"",
		buttons,
		"",
		help,
	)

	return formStyle.
		Width(f.width).
		MaxHeight(f.height).
		Render(content)
}

// SetSize updates form dimensions
func (f *TaskForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.titleInput.Width = width - 20
	f.descInput.SetWidth(width - 20)
	f.labelsInput.Width = width - 20
}

func (f *TaskForm) renderField(label, input string, field FormField) string {
	focused := FormField(f.focusIndex) == field

	labelStyle := lipgloss.NewStyle().Width(12)
	if focused {
		labelStyle = labelStyle.Foreground(primaryColor).Bold(true)
	}

	return labelStyle.Render(label+":") + " " + input
}

func (f *TaskForm) renderSelect(label string, options []string, selected int, field FormField) string {
	focused := FormField(f.focusIndex) == field

	labelStyle := lipgloss.NewStyle().Width(12)
	if focused {
		labelStyle = labelStyle.Foreground(primaryColor).Bold(true)
	}

	var rendered []string
	for i, opt := range options {
		style := lipgloss.NewStyle()
		if i == selected {
			style = style.Bold(true).Foreground(primaryColor)
			if focused {
				opt = "[" + opt + "]"
			} else {
				opt = " " + opt + " "
			}
		} else {
			style = style.Foreground(mutedColor)
			opt = " " + opt + " "
		}
		rendered = append(rendered, style.Render(opt))
	}

	return labelStyle.Render(label+":") + " " + strings.Join(rendered, "")
}

func (f *TaskForm) renderButtons() string {
	submitFocused := FormField(f.focusIndex) == FieldSubmit
	cancelFocused := FormField(f.focusIndex) == FieldCancel

	submitStyle := buttonStyle
	if submitFocused {
		submitStyle = buttonFocusedStyle
	}
	cancelStyle := buttonStyle
	if cancelFocused {
		cancelStyle = buttonFocusedStyle
	}

	submit := submitStyle.Render("[ Save ]")
	cancel := cancelStyle.Render("[ Cancel ]")

	return lipgloss.JoinHorizontal(lipgloss.Center, submit, "  ", cancel)
}

func (f *TaskForm) nextField() {
	f.blurCurrent()
	f.focusIndex = (f.focusIndex + 1) % numFields
	f.focusCurrent()
}

func (f *TaskForm) prevField() {
	f.blurCurrent()
	f.focusIndex = (f.focusIndex - 1 + numFields) % numFields
	f.focusCurrent()
}

func (f *TaskForm) blurCurrent() {
	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput.Blur()
	case FieldDescription:
		f.descInput.Blur()
	case FieldLabels:
		f.labelsInput.Blur()
	case FieldDueDate:
		f.dueDateInput.Blur()
	}
}

func (f *TaskForm) focusCurrent() {
	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput.Focus()
	case FieldDescription:
		f.descInput.Focus()
	case FieldLabels:
		f.labelsInput.Focus()
	case FieldDueDate:
		f.dueDateInput.Focus()
	}
}

func (f *TaskForm) isSelectField() bool {
	field := FormField(f.focusIndex)
	return field == FieldType || field == FieldPriority || field == FieldColumn
}

func (f *TaskForm) selectNext() {
	switch FormField(f.focusIndex) {
	case FieldType:
		f.typeSelect = (f.typeSelect + 1) % len(f.types)
	case FieldPriority:
		f.prioritySelect = (f.prioritySelect + 1) % len(f.priorities)
	case FieldColumn:
		f.columnSelect = (f.columnSelect + 1) % len(f.columns)
	}
}

func (f *TaskForm) selectPrev() {
	switch FormField(f.focusIndex) {
	case FieldType:
		f.typeSelect = (f.typeSelect - 1 + len(f.types)) % len(f.types)
	case FieldPriority:
		f.prioritySelect = (f.prioritySelect - 1 + len(f.priorities)) % len(f.priorities)
	case FieldColumn:
		f.columnSelect = (f.columnSelect - 1 + len(f.columns)) % len(f.columns)
	}
}

func (f *TaskForm) updateFocusedInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput, cmd = f.titleInput.Update(msg)
	case FieldDescription:
		f.descInput, cmd = f.descInput.Update(msg)
	case FieldLabels:
		f.labelsInput, cmd = f.labelsInput.Update(msg)
	case FieldDueDate:
		f.dueDateInput, cmd = f.dueDateInput.Update(msg)
	}

	return cmd
}

func (f *TaskForm) submit() tea.Cmd {
	// Validate
	title := strings.TrimSpace(f.titleInput.Value())
	if title == "" {
		return showStatus("Title is required", true, 3*time.Second)
	}

	// Parse labels
	var labels []string
	labelsStr := strings.TrimSpace(f.labelsInput.Value())
	if labelsStr != "" {
		for _, l := range strings.Split(labelsStr, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				labels = append(labels, l)
			}
		}
	}

	data := TaskFormData{
		Title:       title,
		Description: f.descInput.Value(),
		Type:        f.types[f.typeSelect],
		Priority:    f.priorities[f.prioritySelect],
		Column:      f.columns[f.columnSelect],
		Labels:      labels,
		DueDate:     strings.TrimSpace(f.dueDateInput.Value()),
	}

	return func() tea.Msg {
		return submitTaskFormMsg{
			mode:   f.mode,
			taskID: f.taskID,
			data:   data,
		}
	}
}

// Mode returns the form mode
func (f *TaskForm) Mode() FormMode {
	return f.mode
}

// TaskID returns the task ID (for edit mode)
func (f *TaskForm) TaskID() string {
	return f.taskID
}
