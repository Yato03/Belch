package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State represents the phase the TUI is currently in.
type State int

const (
	StateReady    State = iota // config summary, press Enter to start
	StateRunning               // results streaming in
	StateDone                  // all results received
	StateExport                // export format picker
	StateExported              // export confirmation
)

// Config holds the immutable run parameters (set from CLI flags).
type Config struct {
	ReqFile    string
	Mode       string
	Target     string
	Threads    int
	SkipVerify bool
	Timeout    time.Duration
}

// ResultRow is one completed HTTP request.
type ResultRow struct {
	N          int
	StatusCode int
	Length     int
	Duration   time.Duration
	Payloads   string
	Error      string
}

// ResultMsg is the Bubble Tea message for a completed request.
type ResultMsg ResultRow

// DoneMsg signals that all requests have finished.
type DoneMsg struct{}

// ExportMsg is the result of an export operation.
type ExportMsg struct {
	Path string
	Err  error
}

// Model is the top-level Bubble Tea model.
type Model struct {
	State        State
	config       Config
	total        int
	results      []ResultRow
	filtered     []ResultRow
	resultCh     <-chan ResultMsg
	tbl          table.Model
	filterInput  textinput.Model
	filterActive bool
	exportMsg    string
	exportErr    bool
	width        int
	height       int
}

// Results returns the accumulated result rows (read-only copy).
func (m Model) Results() []ResultRow { return m.results }

// NewModel constructs the initial model. resultCh is read-only; main closes it when done.
func NewModel(config Config, total int, resultCh <-chan ResultMsg) Model {
	ti := textinput.New()
	ti.Placeholder = "filtrar por status, length, payload…"
	ti.CharLimit = 64
	ti.Width = 40

	cols := tableColumns(80)
	s := table.DefaultStyles()
	s.Header = styleTableHeader
	s.Cell = styleTableCell
	s.Selected = styleTableSelected

	tbl := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithStyles(s),
	)

	return Model{
		State:       StateReady,
		config:      config,
		total:       total,
		results:     make([]ResultRow, 0, total),
		filtered:    make([]ResultRow, 0, total),
		resultCh:    resultCh,
		tbl:         tbl,
		filterInput: ti,
	}
}

func tableColumns(width int) []table.Column {
	payloadW := width - 6 - 8 - 8 - 12 - 4
	if payloadW < 20 {
		payloadW = 20
	}
	return []table.Column{
		{Title: "#", Width: 6},
		{Title: "status", Width: 8},
		{Title: "length", Width: 8},
		{Title: "time", Width: 12},
		{Title: "payload", Width: payloadW},
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tbl.SetColumns(tableColumns(msg.Width))
		m.syncTable()
		return m, nil

	case ResultMsg:
		m.results = append(m.results, ResultRow(msg))
		m.refilter()
		m.syncTable()
		return m, waitForResult(m.resultCh)

	case DoneMsg:
		m.State = StateDone
		return m, nil

	case ExportMsg:
		m.State = StateExported
		if msg.Err != nil {
			m.exportMsg = "Error: " + msg.Err.Error()
			m.exportErr = true
		} else {
			m.exportMsg = "Guardado en " + msg.Path
			m.exportErr = false
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m.delegateToComponent(msg)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.State {
	case StateReady:
		if key == "enter" {
			m.State = StateRunning
			return m, waitForResult(m.resultCh)
		}
		if key == "q" {
			return m, tea.Quit
		}

	case StateRunning:
		return m.handleTableAndFilter(key, msg)

	case StateDone:
		switch key {
		case "q":
			return m, tea.Quit
		case "e":
			m.State = StateExport
			return m, nil
		}
		return m.handleTableAndFilter(key, msg)

	case StateExport:
		switch key {
		case "1":
			return m, doExport(m.results, "csv")
		case "2":
			return m, doExport(m.results, "json")
		case "3":
			return m, doExport(m.results, "txt")
		case "esc", "q":
			m.State = StateDone
		}
		return m, nil

	case StateExported:
		m.State = StateDone
		m.exportMsg = ""
		return m, nil
	}

	return m, nil
}

func (m Model) handleTableAndFilter(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "/", "f":
		m.filterActive = !m.filterActive
		if m.filterActive {
			m.filterInput.Focus()
			return m, textinput.Blink
		}
		m.filterInput.Blur()
		return m, nil
	case "esc":
		if m.filterActive {
			m.filterActive = false
			m.filterInput.Blur()
			return m, nil
		}
	case "r":
		m.filterActive = false
		m.filterInput.Blur()
		m.filterInput.SetValue("")
		m.refilter()
		m.syncTable()
		return m, nil
	}

	if m.filterActive {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.refilter()
		m.syncTable()
		return m, cmd
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m Model) delegateToComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.filterActive {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.refilter()
		m.syncTable()
		return m, cmd
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.State {
	case StateReady:
		return m.viewReady()
	case StateRunning, StateDone:
		return m.viewTable()
	case StateExport:
		return m.viewExport()
	case StateExported:
		return m.viewExported()
	}
	return ""
}

func (m Model) viewReady() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styleTitle.Render("  fuzzer") + "\n\n")
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		styleGray.Render("target:"),
		styleBold.Render(m.config.Target)))
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		styleGray.Render("mode:  "),
		styleBold.Render(m.config.Mode)))
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		styleGray.Render("jobs:  "),
		styleBold.Render(fmt.Sprintf("%d", m.total))))
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		styleGray.Render("threads:"),
		styleBold.Render(fmt.Sprintf("%d", m.config.Threads))))
	sb.WriteString("\n")
	sb.WriteString(styleGray.Render("  Pulsa Enter para empezar · q para salir") + "\n")
	return sb.String()
}

func (m Model) viewTable() string {
	header := m.renderHeader()
	tblView := m.tbl.View()
	footer := m.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, tblView, footer)
}

func (m Model) renderHeader() string {
	done := len(m.results)
	progress := fmt.Sprintf("%d/%d", done, m.total)

	left := styleTitle.Render("fuzzer") + "  " +
		styleGray.Render(m.config.Mode) + " · " +
		styleGray.Render(m.config.Target)

	var status string
	if m.State == StateRunning {
		status = styleStatus3xx.Render("⟳ running")
	} else {
		status = styleStatus2xx.Render("✓ done")
	}

	right := status + "  " + styleGray.Render(progress)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	line := styleHeaderBar.Render(left + strings.Repeat(" ", gap) + right)
	return line
}

func (m Model) renderFooter() string {
	var filterPart string
	if m.filterActive {
		filterPart = styleFilterLabel.Render("Filter: ") + m.filterInput.View()
	} else {
		filterPart = styleGray.Render("Filter: ") + styleGray.Render(m.filterInput.Value())
	}

	shown := len(m.filtered)
	total := len(m.results)
	countPart := styleGray.Render(fmt.Sprintf("%d shown / %d total", shown, total))

	gap := m.width - lipgloss.Width(filterPart) - lipgloss.Width(countPart) - 2
	if gap < 1 {
		gap = 1
	}
	row1 := filterPart + strings.Repeat(" ", gap) + countPart

	var keys string
	if m.State == StateDone {
		keys = styleGray.Render("[/] filtrar  [r] reset  [e] exportar  [q] salir")
	} else {
		keys = styleGray.Render("[/] filtrar  [r] reset  [ctrl+c] salir")
	}

	return "\n" + row1 + "\n" + keys
}

func (m Model) viewExport() string {
	content := styleExportBox.Render(
		styleTitle.Render("Exportar resultados") + "\n\n" +
			styleGray.Render("[1]") + " CSV\n" +
			styleGray.Render("[2]") + " JSON\n" +
			styleGray.Render("[3]") + " Texto plano\n\n" +
			styleGray.Render("[esc] cancelar"),
	)
	return "\n\n" + lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
}

func (m Model) viewExported() string {
	var msg string
	if m.exportErr {
		msg = styleErrorMsg.Render(m.exportMsg)
	} else {
		msg = styleSuccess.Render("✓ " + m.exportMsg)
	}
	hint := styleGray.Render("Pulsa cualquier tecla para continuar")
	content := styleExportBox.Render(msg + "\n\n" + hint)
	return "\n\n" + lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
}

func (m *Model) refilter() {
	q := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	filtered := make([]ResultRow, 0, len(m.results))
	for _, r := range m.results {
		if q == "" {
			filtered = append(filtered, r)
			continue
		}
		haystack := strings.ToLower(fmt.Sprintf("%d %d %s %s",
			r.StatusCode, r.Length, r.Payloads, r.Error))
		if strings.Contains(haystack, q) {
			filtered = append(filtered, r)
		}
	}
	m.filtered = filtered
}

func (m *Model) syncTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, r := range m.filtered {
		var statusStr string
		if r.Error != "" {
			statusStr = statusStyle(0, r.Error).Render("ERR")
		} else {
			statusStr = statusStyle(r.StatusCode, "").Render(fmt.Sprintf("%d", r.StatusCode))
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", r.N),
			statusStr,
			fmt.Sprintf("%d", r.Length),
			r.Duration.Round(time.Millisecond).String(),
			r.Payloads,
		}
	}
	m.tbl.SetRows(rows)
	if m.height > 10 {
		m.tbl.SetHeight(m.height - 7)
	}
}

func waitForResult(ch <-chan ResultMsg) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return DoneMsg{}
		}
		return r
	}
}

func doExport(rows []ResultRow, format string) tea.Cmd {
	return func() tea.Msg {
		ts := time.Now().Format("20060102-150405")
		var (
			path string
			err  error
		)
		switch format {
		case "csv":
			path = "fuzzer-results-" + ts + ".csv"
			err = ExportCSV(rows, path)
		case "json":
			path = "fuzzer-results-" + ts + ".json"
			err = ExportJSON(rows, path)
		case "txt":
			path = "fuzzer-results-" + ts + ".txt"
			err = ExportText(rows, path)
		}
		return ExportMsg{Path: path, Err: err}
	}
}
