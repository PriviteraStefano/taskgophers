package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const divisor = 3

type status int

const (
	todo status = iota
	inProgress
	done
)

// Model management
var models []tea.Model

const (
	model status = iota
	form
)

// Styling
var (
	columnStyle = lipgloss.NewStyle().
			Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// Task represents a task in a to-do lists.
type Task struct {
	ID          int
	title       string
	description string
	status      status
}

func (t Task) FilterValue() string {
	return t.title
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

func (t *Task) Next() {
	if t.status < done {
		t.status++
	} else {
		t.status = todo
	}
}

func NewTask(status status, title, description string) Task {
	return Task{
		title:       title,
		description: description,
		status:      status,
	}
}

// Model TaskList represents a lists of tasks.
type Model struct {
	focused  status
	lists    []list.Model
	err      error
	loaded   bool
	quitting bool
}

func New() *Model {
	return &Model{}
}

// initList initializes the lists of tasks.
func (m *Model) initList(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height-divisor)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	// Init To Do
	m.lists[todo].Title = "To Do"
	m.lists[todo].SetItems([]list.Item{
		Task{ID: 1, title: "Write documentation", description: "Write documentation for the project", status: todo},
		Task{ID: 2, title: "Implement feature X", description: "Implement feature X according to the specification", status: todo},
		Task{ID: 3, title: "Fix bug Y", description: "Fix bug Y that causes the application to crash", status: todo},
	})

	// Init In Progress
	m.lists[inProgress].Title = "In Progress"
	m.lists[inProgress].SetItems([]list.Item{
		Task{ID: 1, title: "In progress", description: "In progress", status: inProgress},
	})

	// Init Done
	m.lists[done].Title = "Done"
	m.lists[done].SetItems([]list.Item{
		Task{ID: 1, title: "Done", description: "Done", status: done},
	})
	m.loaded = true
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Next() {
	if m.focused == done {
		m.focused = todo
	} else {
		m.focused++
	}
}

func (m *Model) Previous() {
	if m.focused > todo {
		m.focused--
	} else {
		m.focused = done
	}
}

func (m *Model) MoveToNext() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	if selectedItem != nil {
		selectedTask := selectedItem.(Task)
		m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
		selectedTask.Next()
		m.lists[selectedTask.status].
			InsertItem(
				len(m.lists[selectedTask.status].Items()),
				list.Item(selectedTask),
			)
	}

	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			m.Previous()
		case "right", "l":
			m.Next()
		case "enter":
			return m, m.MoveToNext
		case "n":
			models[model] = m // save state of the current model
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), list.Item(task))
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.loaded {
		todoView := m.lists[todo].View()
		inProgressView := m.lists[inProgress].View()
		doneView := m.lists[done].View()
		switch m.focused {
		case inProgress:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				focusedStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		case done:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				focusedStyle.Render(doneView),
			)
		default:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				focusedStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		}
	} else {
		return "loading..."
	}
}

// form model
type Form struct {
	focused     status
	title       textinput.Model
	description textarea.Model
}

func NewForm(focused status) *Form {
	form := &Form{
		title:       textinput.New(),
		description: textarea.New(),
		focused:     focused,
	}
	form.title.Focus()
	return form
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, f.title.View(), f.description.View())
}

func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		{
			switch msg.String() {
			case "ctrl+c", "q":
				return f, tea.Quit
			case "enter":
				if f.title.Focused() {
					f.title.Blur()
					f.description.Focus()
					return f, textarea.Blink
				} else {
					models[form] = f
					return models[model], f.CreateTask
				}
			}
		}
	}
	if f.title.Focused() {
		f.title, cmd = f.title.Update(msg)
		return f, cmd
	} else {
		f.description, cmd = f.description.Update(msg)
		return f, cmd
	}
}

func (f Form) CreateTask() tea.Msg {
	task := NewTask(f.focused, f.title.Value(), f.description.Value())
	return task
}

func main() {
	models = []tea.Model{New(), NewForm(todo)}
	m := models[model]
	p := tea.NewProgram(m)
	if err, _ := p.Run(); err != nil {
		panic(err)
	}
}
