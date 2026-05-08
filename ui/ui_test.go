package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"belch/ui"
)

func newTestModel(total int, ch <-chan ui.ResultMsg) ui.Model {
	return ui.NewModel(ui.Config{
		ReqFile: "test.req",
		Mode:    "sniper",
		Target:  "example.com",
		Threads: 1,
		Timeout: 30 * time.Second,
	}, total, ch)
}

func closedCh() <-chan ui.ResultMsg {
	ch := make(chan ui.ResultMsg)
	close(ch)
	return ch
}

// â”€â”€ Initial state â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestNew_InitialState(t *testing.T) {
	m := newTestModel(10, closedCh())
	if m.State != ui.StateReady {
		t.Errorf("NewModel().State: got %v, want StateReady (%d)", m.State, ui.StateReady)
	}
}

func TestNew_NilResults(t *testing.T) {
	m := newTestModel(5, closedCh())
	if len(m.Results()) != 0 {
		t.Error("NewModel() should have no results")
	}
}

// â”€â”€ StateReady â†’ StateRunning â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestEnter_TransitionsToRunning(t *testing.T) {
	ch := make(chan ui.ResultMsg)
	defer close(ch)
	m := newTestModel(1, ch)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := next.(ui.Model)
	if nm.State != ui.StateRunning {
		t.Errorf("after Enter: State = %v, want StateRunning", nm.State)
	}
}

// â”€â”€ ResultMsg accumulation â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestResultMsg_Appended(t *testing.T) {
	ch := make(chan ui.ResultMsg, 1)
	m := newTestModel(1, ch)
	// transition to running
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(ui.Model)

	// deliver a result
	row := ui.ResultMsg{N: 1, StatusCode: 200, Length: 500, Payloads: "admin"}
	next, _ = m.Update(row)
	m = next.(ui.Model)

	if len(m.Results()) != 1 {
		t.Errorf("results count: got %d, want 1", len(m.Results()))
	}
	if m.Results()[0].StatusCode != 200 {
		t.Errorf("result StatusCode: got %d, want 200", m.Results()[0].StatusCode)
	}
}

// â”€â”€ DoneMsg â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestDoneMsg_TransitionsToDone(t *testing.T) {
	ch := make(chan ui.ResultMsg)
	close(ch)
	m := newTestModel(0, ch)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(ui.Model)
	next, _ = m.Update(ui.DoneMsg{})
	m = next.(ui.Model)
	if m.State != ui.StateDone {
		t.Errorf("after DoneMsg: State = %v, want StateDone", m.State)
	}
}

// â”€â”€ Export state â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestExportKey_OpensExportMenu(t *testing.T) {
	m := newTestModel(0, closedCh())
	// put model in StateDone
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(ui.Model)
	next, _ = m.Update(ui.DoneMsg{})
	m = next.(ui.Model)

	// press e
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(ui.Model)
	if m.State != ui.StateExport {
		t.Errorf("after 'e': State = %v, want StateExport", m.State)
	}
}

func TestExportEsc_ReturnsToDone(t *testing.T) {
	m := newTestModel(0, closedCh())
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(ui.Model)
	next, _ = m.Update(ui.DoneMsg{})
	m = next.(ui.Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(ui.Model)
	// esc cancels
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = next.(ui.Model)
	if m.State != ui.StateDone {
		t.Errorf("after esc in StateExport: State = %v, want StateDone", m.State)
	}
}

// â”€â”€ View â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestView_Ready_NonEmpty(t *testing.T) {
	m := newTestModel(10, closedCh())
	if strings.TrimSpace(m.View()) == "" {
		t.Error("View() in StateReady should return non-empty string")
	}
}

func TestView_Ready_ContainsTarget(t *testing.T) {
	m := newTestModel(10, closedCh())
	if !strings.Contains(m.View(), "example.com") {
		t.Error("View() in StateReady should show target")
	}
}

func TestView_Export_NonEmpty(t *testing.T) {
	m := newTestModel(0, closedCh())
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(ui.Model)
	next, _ = m.Update(ui.DoneMsg{})
	m = next.(ui.Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(ui.Model)
	if strings.TrimSpace(m.View()) == "" {
		t.Error("View() in StateExport should return non-empty string")
	}
}

// â”€â”€ State ordering â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func TestStateValues_AreOrdered(t *testing.T) {
	states := []ui.State{
		ui.StateReady,
		ui.StateRunning,
		ui.StateDone,
		ui.StateExport,
		ui.StateExported,
	}
	for i := 1; i < len(states); i++ {
		if states[i] <= states[i-1] {
			t.Errorf("State constants must be in ascending order: %v should be > %v",
				states[i], states[i-1])
		}
	}
}
