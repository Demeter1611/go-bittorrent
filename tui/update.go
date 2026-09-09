package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FileSelectedMsg:
		m.fileName = string(msg.FileName)
		m.totalSize = uint64(msg.TotalFileSize)
		return m, nil

	case StatsUpdateMsg:
		m.progress = float32(msg.Progress)
		m.speed = float64(msg.Speed)
		m.bytesDownloaded = uint64(msg.BytesDownloaded)
		m.activePeers = int(msg.ActivePeers)
		m.totalPeers = int(msg.TotalPeers)
		return m, nil

	case DownloadCompleteMsg:
		m.isDone = true
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}
