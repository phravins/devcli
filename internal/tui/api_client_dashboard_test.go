package tui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/phravins/devcli/internal/devtools"
)

func TestNewAPIClientModel(t *testing.T) {
	model := NewAPIClientModel()

	if model.state != StateAPIMethodSelect {
		t.Errorf("expected initial state %d (StateAPIMethodSelect), got %d", StateAPIMethodSelect, model.state)
	}

	if model.methodIndex != 0 {
		t.Errorf("expected initial methodIndex 0, got %d", model.methodIndex)
	}

	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	if len(model.methods) != len(expectedMethods) {
		t.Fatalf("expected %d methods, got %d", len(expectedMethods), len(model.methods))
	}
	for i, m := range expectedMethods {
		if model.methods[i] != m {
			t.Errorf("expected method at %d to be %s, got %s", i, m, model.methods[i])
		}
	}

	if model.urlInput.Value() != "https://httpbin.org/get" {
		t.Errorf("expected default URL 'https://httpbin.org/get', got '%s'", model.urlInput.Value())
	}

	if model.bodyInput.Value() != "" {
		t.Errorf("expected empty initial body, got '%s'", model.bodyInput.Value())
	}
}

func TestAPIClientModel_Init(t *testing.T) {
	model := NewAPIClientModel()
	cmd := model.Init()
	if cmd == nil {
		t.Error("expected Init to return a non-nil Cmd")
	}
}

func TestAPIClientModel_WindowSizeMsg(t *testing.T) {
	model := NewAPIClientModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updated, _ := model.Update(msg)
	m := updated.(APIClientModel)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
	if m.viewport.Width != 96 { // Width - 4
		t.Errorf("expected viewport width 96, got %d", m.viewport.Width)
	}
	if m.viewport.Height != 38 { // Height - 12
		t.Errorf("expected viewport height 38, got %d", m.viewport.Height)
	}
}

func TestAPIClientModel_MethodSelectState(t *testing.T) {
	model := NewAPIClientModel()

	// Navigate down with "j" and "down"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updated.(APIClientModel)
	if m.methodIndex != 1 {
		t.Errorf("expected methodIndex 1 after 'j', got %d", m.methodIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(APIClientModel)
	if m.methodIndex != 2 {
		t.Errorf("expected methodIndex 2 after Down, got %d", m.methodIndex)
	}

	// Navigate up with "k" and "up"
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(APIClientModel)
	if m.methodIndex != 1 {
		t.Errorf("expected methodIndex 1 after 'k', got %d", m.methodIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(APIClientModel)
	if m.methodIndex != 0 {
		t.Errorf("expected methodIndex 0 after Up, got %d", m.methodIndex)
	}

	// Test boundary up (cannot go below 0)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(APIClientModel)
	if m.methodIndex != 0 {
		t.Errorf("expected methodIndex 0 at lower bound, got %d", m.methodIndex)
	}

	// Move to bottom boundary
	for i := 0; i < 10; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(APIClientModel)
	}
	if m.methodIndex != len(m.methods)-1 {
		t.Errorf("expected methodIndex %d at upper bound, got %d", len(m.methods)-1, m.methodIndex)
	}

	// Test Esc key returns BackMsg
	m.state = StateAPIMethodSelect
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Esc in StateAPIMethodSelect")
	}
	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Errorf("expected BackMsg on Esc, got %T", msg)
	}

	// Test Enter key moves to StateAPIURLInput
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(APIClientModel)
	if m.state != StateAPIURLInput {
		t.Errorf("expected state StateAPIURLInput after Enter, got %d", m.state)
	}
	if !m.urlInput.Focused() {
		t.Error("expected urlInput to be focused")
	}
}

func TestAPIClientModel_URLInputState(t *testing.T) {
	// 1. GET method (methodIndex 0) -> Enter should switch directly to StateAPISending
	model := NewAPIClientModel()
	model.state = StateAPIURLInput
	model.urlInput.Focus()
	model.methodIndex = 0 // GET

	// Esc returns to StateAPIMethodSelect
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(APIClientModel)
	if m.state != StateAPIMethodSelect {
		t.Errorf("expected state StateAPIMethodSelect after Esc, got %d", m.state)
	}
	if m.urlInput.Focused() {
		t.Error("expected urlInput to be blurred on Esc")
	}

	// Enter on GET -> StateAPISending
	model.state = StateAPIURLInput
	model.urlInput.Focus()
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(APIClientModel)
	if m.state != StateAPISending {
		t.Errorf("expected state StateAPISending for GET, got %d", m.state)
	}
	if cmd == nil {
		t.Error("expected execution cmd for GET request")
	}

	// 2. DELETE method (methodIndex 3) -> Enter should switch to StateAPISending
	model = NewAPIClientModel()
	model.state = StateAPIURLInput
	model.urlInput.Focus()
	model.methodIndex = 3 // DELETE
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(APIClientModel)
	if m.state != StateAPISending {
		t.Errorf("expected state StateAPISending for DELETE, got %d", m.state)
	}
	if cmd == nil {
		t.Error("expected execution cmd for DELETE request")
	}

	// 3. POST method (methodIndex 1) -> Enter should switch to StateAPIBodyInput
	model = NewAPIClientModel()
	model.state = StateAPIURLInput
	model.urlInput.Focus()
	model.methodIndex = 1 // POST
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(APIClientModel)
	if m.state != StateAPIBodyInput {
		t.Errorf("expected state StateAPIBodyInput for POST, got %d", m.state)
	}
	if !m.bodyInput.Focused() {
		t.Error("expected bodyInput to be focused for POST")
	}

	// 4. Typing input into URL field
	model = NewAPIClientModel()
	model.state = StateAPIURLInput
	model.urlInput.Focus()
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(APIClientModel)
	if !strings.HasSuffix(m.urlInput.Value(), "a") {
		t.Errorf("expected URL input to end with 'a', got '%s'", m.urlInput.Value())
	}
}

func TestAPIClientModel_BodyInputState(t *testing.T) {
	model := NewAPIClientModel()
	model.state = StateAPIBodyInput
	model.bodyInput.Focus()
	model.methodIndex = 1 // POST

	// Esc returns to StateAPIURLInput
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(APIClientModel)
	if m.state != StateAPIURLInput {
		t.Errorf("expected state StateAPIURLInput after Esc, got %d", m.state)
	}
	if !m.urlInput.Focused() {
		t.Error("expected urlInput to be focused after returning from body input")
	}
	if m.bodyInput.Focused() {
		t.Error("expected bodyInput to be blurred on Esc")
	}

	// Enter sends request -> StateAPISending
	model.state = StateAPIBodyInput
	model.bodyInput.Focus()
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(APIClientModel)
	if m.state != StateAPISending {
		t.Errorf("expected state StateAPISending, got %d", m.state)
	}
	if cmd == nil {
		t.Error("expected execution cmd after Enter on body input")
	}

	// Typing input into Body field
	model.state = StateAPIBodyInput
	model.bodyInput.Focus()
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	m = updated.(APIClientModel)
	if m.bodyInput.Value() != "{" {
		t.Errorf("expected body input '{', got '%s'", m.bodyInput.Value())
	}
}

func TestAPIClientModel_APIExecMsg(t *testing.T) {
	model := NewAPIClientModel()

	// 1. Success response
	respSuccess := devtools.APIResponse{
		Status:     "200 OK",
		StatusCode: 200,
		LatencyMs:  45,
		Formatted:  `{"status": "ok"}`,
		Err:        nil,
	}

	updated, _ := model.Update(apiExecMsg{response: respSuccess})
	m := updated.(APIClientModel)

	if m.state != StateAPIResponseView {
		t.Errorf("expected state StateAPIResponseView, got %d", m.state)
	}
	if m.lastResponse.Status != "200 OK" {
		t.Errorf("expected status '200 OK', got '%s'", m.lastResponse.Status)
	}
	if !strings.Contains(m.viewport.View(), "200 OK") && !strings.Contains(m.viewport.View(), "status") {
		t.Error("expected viewport to contain response data")
	}

	// 2. Error response in APIResponse
	respErr := devtools.APIResponse{
		Err: errors.New("connection refused"),
	}

	updated, _ = model.Update(apiExecMsg{response: respErr})
	m = updated.(APIClientModel)

	if m.state != StateAPIResponseView {
		t.Errorf("expected state StateAPIResponseView, got %d", m.state)
	}
	if !strings.Contains(m.viewport.View(), "connection refused") {
		t.Errorf("expected viewport to render error message, got '%s'", m.viewport.View())
	}
}

func TestAPIClientModel_ResponseViewState(t *testing.T) {
	model := NewAPIClientModel()
	model.state = StateAPIResponseView

	// Esc returns to StateAPIMethodSelect
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(APIClientModel)
	if m.state != StateAPIMethodSelect {
		t.Errorf("expected state StateAPIMethodSelect after Esc, got %d", m.state)
	}

	// 'q' returns to StateAPIMethodSelect
	model.state = StateAPIResponseView
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(APIClientModel)
	if m.state != StateAPIMethodSelect {
		t.Errorf("expected state StateAPIMethodSelect after 'q', got %d", m.state)
	}

	// 'r' resends request (transitions to StateAPISending)
	model.state = StateAPIResponseView
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(APIClientModel)
	if m.state != StateAPISending {
		t.Errorf("expected state StateAPISending after 'r', got %d", m.state)
	}
	if cmd == nil {
		t.Error("expected resend execution cmd")
	}
}

func TestAPIClientModel_View(t *testing.T) {
	model := NewAPIClientModel()
	model.width = 80
	model.height = 24

	// 1. StateAPIMethodSelect
	model.state = StateAPIMethodSelect
	viewSelect := model.View()
	if !strings.Contains(viewSelect, "DevCLI API & HTTP Playground") || !strings.Contains(viewSelect, "Select HTTP Method:") {
		t.Errorf("unexpected view output for StateAPIMethodSelect:\n%s", viewSelect)
	}

	// 2. StateAPIURLInput
	model.state = StateAPIURLInput
	viewURL := model.View()
	if !strings.Contains(viewURL, "Enter Target URL:") {
		t.Errorf("unexpected view output for StateAPIURLInput:\n%s", viewURL)
	}

	// 3. StateAPIBodyInput
	model.state = StateAPIBodyInput
	viewBody := model.View()
	if !strings.Contains(viewBody, "Enter Request Body JSON") {
		t.Errorf("unexpected view output for StateAPIBodyInput:\n%s", viewBody)
	}

	// 4. StateAPISending
	model.state = StateAPISending
	viewSending := model.View()
	if !strings.Contains(viewSending, "Sending GET request") {
		t.Errorf("unexpected view output for StateAPISending:\n%s", viewSending)
	}

	// 5. StateAPIResponseView - Status 200 (Green badge)
	model.state = StateAPIResponseView
	model.lastResponse = devtools.APIResponse{
		StatusCode: 200,
		Status:     "200 OK",
		LatencyMs:  12,
		Formatted:  "OK",
	}
	viewResp200 := model.View()
	if !strings.Contains(viewResp200, "200 OK") || !strings.Contains(viewResp200, "12 ms") {
		t.Errorf("unexpected view output for StateAPIResponseView (200):\n%s", viewResp200)
	}

	// 6. StateAPIResponseView - Status 404 (Red badge)
	model.lastResponse = devtools.APIResponse{
		StatusCode: 404,
		Status:     "404 Not Found",
		LatencyMs:  15,
		Formatted:  "Not Found",
	}
	viewResp404 := model.View()
	if !strings.Contains(viewResp404, "404 Not Found") {
		t.Errorf("unexpected view output for StateAPIResponseView (404):\n%s", viewResp404)
	}

	// 7. Unknown state fallback
	model.state = 999
	viewUnknown := model.View()
	if viewUnknown != "Unknown API Client state" {
		t.Errorf("expected 'Unknown API Client state', got '%s'", viewUnknown)
	}
}

func TestExecuteAPIReqCmd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"hello": "world"}`))
	}))
	defer ts.Close()

	cmd := executeAPIReqCmd("GET", ts.URL, "")
	msg := cmd()

	execMsg, ok := msg.(apiExecMsg)
	if !ok {
		t.Fatalf("expected msg to be apiExecMsg, got %T", msg)
	}

	if execMsg.response.Err != nil {
		t.Fatalf("unexpected error: %v", execMsg.response.Err)
	}

	if execMsg.response.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", execMsg.response.StatusCode)
	}
}
