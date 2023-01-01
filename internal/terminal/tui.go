package tui_test

import (
	"fmt"
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/rog-golang-buddies/rmx/internal/terminal/internal/jam-session"
	"github.com/rog-golang-buddies/rmx/internal/terminal/internal/lobby"
)

// ********
// Code heavily based on "Project Journal"
// https://github.com/bashbunni/pjs
// https://www.youtube.com/watch?v=uJ2egAkSkjg&t=319s
// ********

type Session struct {
	Id suid.UUID `json:"id"`
	// UserCount int    `json:"userCount"`
}

type appView int

const (
	jamView appView = iota
	lobbyView
)

type mainModel struct {
	curView      appView
	lobby        tea.Model
	jam          jam.Model
	RESTendpoint string
	WSendpoint   string
	jamSocket    *websocket.Conn // Websocket connection to a Jam Session
}

func NewModel(serverHostURL string) (mainModel, error) {
	wsHostURL, err := url.Parse(serverHostURL)
	if err != nil {
		return mainModel{}, err
	}
	wsHostURL.Scheme = "ws"

	return mainModel{
		curView:      lobbyView,
		lobby:        lobby.New(wsHostURL.String()+"/ws", serverHostURL+"/api/v1"),
		jam:          jam.New(),
		RESTendpoint: serverHostURL + "/api/v1",
	}, nil
}

func (m mainModel) Init() tea.Cmd {
	return lobby.FetchSessions(m.RESTendpoint)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// Handle incoming messages from I/O
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Ctrl+c exits. Even with short running programs it's good to have
		// a quit key, just incase your logic is off. Users will be very
		// annoyed if they can't exit.
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case lobby.JamConnected:
		m.curView = jamView
		m.jam.Socket = msg.WS
		m.jam.ID = msg.JamID
		cmds = append(cmds, jam.Enter)
	}

	// Call sub-model Updates
	switch m.curView {
	case lobbyView:
		newLobby, newCmd := m.lobby.Update(msg)
		lobbyModel, ok := newLobby.(lobby.Model)
		if !ok {
			panic("could not perform assertion on lobby model")
		}
		m.lobby = lobbyModel
		cmd = newCmd
	case jamView:
		newJam, newCmd := m.jam.Update(msg)
		jamModel, ok := newJam.(jam.Model)
		if !ok {
			panic("could not perform assertion on jam model")
		}
		m.jam = jamModel
		cmd = newCmd
	}
	// Run all commands from sub-model Updates
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)

}

func (m mainModel) View() string {
	switch m.curView {
	case jamView:
		return m.jam.View()
	default:
		return m.lobby.View()
	}
}

func Run() {
	// TODO: Get from args, user input, or env
	const serverHostURL = "http://localhost:9003"
	m, err := NewModel(serverHostURL)
	if err != nil {
		bail(err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		bail(err)
	}
}

func bail(err error) {
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
