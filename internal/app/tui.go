package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	file "github.com/redkenrok/go-file_sorter/internal/file"
)

const BUFFER_HEIGHT = 10
const FORMAT_PLACEHOLDER = "%year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%"

type keyMap struct {
	Help  key.Binding
	Enter key.Binding
	Quit  key.Binding
	Space key.Binding

	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Help,
			k.Quit,
			k.Enter,
			k.Space,
		},
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
		},
	}
}

var keys = keyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),

	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move in"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move out"),
	),
}

type state int

const (
	stateSourcePicker state = iota
	stateDestPicker
	stateFormatInput
	stateConfirm
	stateProcessing
	stateFinished
)

type model struct {
	version   string
	commit    string
	buildDate string

	keys   keyMap
	help   help.Model
	err    error
	height int
	width  int

	sourcePicker filepicker.Model
	destPicker   filepicker.Model

	state        state
	formatInput  textinput.Model
	confirmIndex int

	currentOperation string
	dryRun           bool
	moveMode         bool
	processed        int
	total            int
	lastProcessed    []fileRecord
}

type processingStarted struct {
	totalFiles int
}

type fileProcessed struct {
	path            string
	destinationPath string
	error           error
}

type fileRecord struct {
	action          string
	path            string
	destinationPath string
}

type processingFinished struct{}

func initialModel(
	version string,
	commit string,
	buildDate string,
) model {
	sp := filepicker.New()
	sp.CurrentDirectory, _ = os.Getwd()
	sp.DirAllowed = true
	sp.Height = 20

	fi := textinput.New()
	fi.Placeholder = FORMAT_PLACEHOLDER
	fi.Width = 80

	return model{
		version:   version,
		commit:    commit,
		buildDate: buildDate,

		keys: keys,
		help: help.New(),

		height: 24,
		width:  80,

		state:       stateSourcePicker,
		formatInput: fi,

		sourcePicker: sp,
	}
}

func (
	m model,
) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.sourcePicker.Init(),
		tea.EnterAltScreen,
	)
}

func (
	m model,
) Update(
	msg tea.Msg,
) (
	tea.Model,
	tea.Cmd,
) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case processingStarted:
		m.total = msg.totalFiles
		m.processed = 0
		return m, m.processNextFile(0)

	case fileProcessed:
		if msg.error != nil {
			m.err = msg.error
			return m, tea.Quit
		}

		m.processed++

		action := ""
		if m.dryRun {
			action += "Would "
			if m.moveMode {
				action += "move"
			} else {
				action += "copy"
			}
		} else if m.moveMode {
			action = "Move"
		} else {
			action = "Copy"
		}

		if msg.path != "" {
			record := fileRecord{
				action:          action,
				path:            msg.path,
				destinationPath: msg.destinationPath,
			}
			m.lastProcessed = append(m.lastProcessed, record)
			if len(m.lastProcessed) > 5 {
				m.lastProcessed = m.lastProcessed[len(m.lastProcessed)-5:]
			}
		}

		if !m.dryRun {
			m.currentOperation = fmt.Sprintf("%s %s\n → %s", action, msg.path, msg.destinationPath)
		} else {
			m.currentOperation = fmt.Sprintf("%s %s\n → %s", action, msg.path, msg.destinationPath)
		}

		if m.processed < m.total {
			return m, m.processNextFile(m.processed)
		}
		return m, func() tea.Msg {
			return processingFinished{}
		}

	case processingFinished:
		if m.processed == m.total {
			m.state = stateFinished
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.sourcePicker.Height = m.height - BUFFER_HEIGHT
		m.destPicker.Height = m.height - BUFFER_HEIGHT
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}

		switch m.state {
		case stateSourcePicker:
			if key.Matches(msg, m.keys.Enter) {
				selectedDir := m.sourcePicker.CurrentDirectory
				if selectedDir != "" {
					m.state = stateDestPicker
					m.destPicker = filepicker.New()
					m.destPicker.DirAllowed = true
					m.destPicker.Height = m.height - BUFFER_HEIGHT
					m.destPicker.CurrentDirectory, _ = os.Getwd()
					return m, m.destPicker.Init()
				}
			}

		case stateDestPicker:
			if key.Matches(msg, m.keys.Enter) {
				if m.destPicker.CurrentDirectory != "" {
					m.state = stateFormatInput
					m.formatInput.Focus()
					return m, textinput.Blink
				}
			}

		case stateFormatInput:
			if key.Matches(msg, m.keys.Enter) {
				m.state = stateConfirm
				return m, nil
			}

		case stateConfirm:
			if key.Matches(msg, m.keys.Up) && m.confirmIndex > 0 {
				m.confirmIndex--
			}
			if key.Matches(msg, m.keys.Down) && m.confirmIndex < 1 {
				m.confirmIndex++
			}

			if key.Matches(msg, m.keys.Space) {
				if m.confirmIndex == 0 {
					m.dryRun = !m.dryRun
				} else if m.confirmIndex == 1 {
					m.moveMode = !m.moveMode
				}
				return m, nil
			}

			if key.Matches(msg, m.keys.Enter) {
				m.state = stateProcessing
				return m, m.processFiles()
			}

		case stateFinished:
			if key.Matches(msg, m.keys.Enter) || key.Matches(msg, m.keys.Quit) {
				return m, tea.Quit
			}
		}
	}

	// Update the active component based on the current state
	var cmd tea.Cmd
	switch m.state {
	case stateSourcePicker:
		m.sourcePicker, cmd = m.sourcePicker.Update(msg)
		cmds = append(cmds, cmd)
	case stateDestPicker:
		m.destPicker, cmd = m.destPicker.Update(msg)
		cmds = append(cmds, cmd)
	case stateFormatInput:
		m.formatInput, cmd = m.formatInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (
	m model,
) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(0)

	s.WriteString(
		titleStyle.Render(
			fmt.Sprintf("File sorter v%s", m.version),
		) + "\n",
	)

	switch m.state {
	case stateSourcePicker:
		s.WriteString(lipgloss.NewStyle().MarginBottom(1).Render("Select source directory:"))
		s.WriteString("\n" + m.sourcePicker.View())
		break

	case stateDestPicker:
		s.WriteString(lipgloss.NewStyle().MarginBottom(1).Render("Select destination directory:"))
		s.WriteString("\n" + m.destPicker.View())
		break

	case stateFormatInput:
		s.WriteString("Enter file format pattern:\n")
		s.WriteString(m.formatInput.View())
		s.WriteString("\n\nFormat placeholders:\n")
		s.WriteString("%year%, %month%, %day%, %hour%, %minute%, %second%\n")
		s.WriteString("%index%     - Incremental file count\n")
		s.WriteString("%ext%       - File extension\n")
		s.WriteString("%type%      - File type (flac,svg+xml,webm)\n")
		s.WriteString("%mime-type% - File's mime-type (audio/flac,image/svg+xml,video/webm)\n")
		s.WriteString(fmt.Sprintf("\nDefault format:\n%s", FORMAT_PLACEHOLDER))
		break

	case stateConfirm:
		format := m.formatInput.Value()
		if format == "" {
			format = m.formatInput.Placeholder
		}
		s.WriteString(fmt.Sprintf("Source:      %s\n", m.sourcePicker.CurrentDirectory))
		s.WriteString(fmt.Sprintf("Destination: %s\n", m.destPicker.CurrentDirectory))
		s.WriteString(fmt.Sprintf("Format:      %s\n", format))

		dryRunCheckbox := " Log without changing files"
		if m.dryRun {
			dryRunCheckbox = "[X]" + dryRunCheckbox
		} else {
			dryRunCheckbox = "[ ]" + dryRunCheckbox
		}
		moveModeCheckbox := " Move instead of copy"
		if m.moveMode {
			moveModeCheckbox = "[X]" + moveModeCheckbox
		} else {
			moveModeCheckbox = "[ ]" + moveModeCheckbox
		}

		if m.confirmIndex == 0 {
			dryRunCheckbox = "> " + dryRunCheckbox
		} else {
			dryRunCheckbox = "  " + dryRunCheckbox
		}
		if m.confirmIndex == 1 {
			moveModeCheckbox = "> " + moveModeCheckbox
		} else {
			moveModeCheckbox = "  " + moveModeCheckbox
		}

		s.WriteString("\n" + dryRunCheckbox)
		s.WriteString("\n" + moveModeCheckbox)
		break

	case stateProcessing:
		s.WriteString("Processing...\n")
		if len(m.lastProcessed) > 0 {
			s.WriteString("\nLast processed files:\n")
			for _, record := range m.lastProcessed {
				s.WriteString(fmt.Sprintf("%s %s\n → %s\n", record.action, record.path, record.destinationPath))
			}
		}
		if m.currentOperation != "" {
			s.WriteString("\nCurrent operation:\n")
			s.WriteString(m.currentOperation + "\n")
		}
		s.WriteString(fmt.Sprintf("Processed %d out of %d files.", m.processed, m.total))
		if m.err != nil {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
			s.WriteString("\n" + style.Render(m.err.Error()))
		}
		break

	case stateFinished:
		s.WriteString("Processing done.\n")
		if len(m.lastProcessed) > 0 {
			s.WriteString("\nLast processed files:\n")
			for _, record := range m.lastProcessed {
				s.WriteString(fmt.Sprintf("%s %s\n → %s\n", record.action, record.path, record.destinationPath))
			}
		}
		s.WriteString(fmt.Sprintf("\nProcessed %d files in total.", m.total))
		break
	}

	content := s.String()
	helpView := m.help.View(m.keys)

	contentLines := strings.Split(content, "\n")
	contentHeight := len(contentLines)
	helpLines := strings.Split(helpView, "\n")
	helpHeight := len(helpLines)
	paddingLines := m.height - contentHeight - (helpHeight + 1)
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	finalView := content + padding + "\n" + helpView

	return finalView
}

func (
	m *model,
) processFiles() tea.Cmd {
	return func() tea.Msg {
		var totalFiles int
		filepath.Walk(m.sourcePicker.CurrentDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.HasPrefix(path, m.destPicker.CurrentDirectory) {
				return nil
			}

			rel, err := filepath.Rel(m.destPicker.CurrentDirectory, path)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return nil
			}

			totalFiles++
			return nil
		})

		return processingStarted{totalFiles: totalFiles}
	}
}

func (
	m *model,
) processNextFile(
	index int,
) tea.Cmd {
	return func() tea.Msg {
		var currentFile string
		var currentIndex int

		// Find the nth unprocessed file
		err := filepath.Walk(
			m.sourcePicker.CurrentDirectory,
			func(
				path string,
				info os.FileInfo,
				err error,
			) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if strings.HasPrefix(path, m.destPicker.CurrentDirectory) {
					return nil
				}

				if currentIndex == index {
					currentFile = path
					return fmt.Errorf("found") // Use error to break the walk
				}

				currentIndex++
				return nil
			},
		)

		if err != nil && err.Error() != "found" {
			return fileProcessed{error: err}
		}

		if currentFile == "" {
			return processingFinished{}
		}

		// Process the file
		creationDate, err := file.GetFileCreationDate(currentFile)
		if err != nil {
			return fileProcessed{
				path:  currentFile,
				error: fmt.Errorf("error getting creation date: %w", err),
			}
		}

		format := m.formatInput.Value()
		if format == "" {
			format = m.formatInput.Placeholder
		}

		newFileName := file.FormatName(format, creationDate, index+1, currentFile)
		destinationPath := filepath.Join(m.destPicker.CurrentDirectory, newFileName)

		if !m.dryRun {
			if err := os.MkdirAll(filepath.Dir(destinationPath), os.ModePerm); err != nil {
				return fileProcessed{
					path:  currentFile,
					error: fmt.Errorf("failed to create directory: %w", err),
				}
			}

			if m.moveMode {
				err = os.Rename(currentFile, destinationPath)
			} else {
				err = file.CopyFile(currentFile, destinationPath)
			}

			if err != nil {
				return fileProcessed{
					path: currentFile,
					error: fmt.Errorf("failed to %s file: %w",
						map[bool]string{true: "move", false: "copy"}[m.moveMode],
						err),
				}
			}
		}

		return fileProcessed{
			path:            currentFile,
			destinationPath: destinationPath,
		}
	}
}

func RunTUI(
	version string,
	commit string,
	buildDate string,
) {
	p := tea.NewProgram(
		initialModel(
			version,
			commit,
			buildDate,
		),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
