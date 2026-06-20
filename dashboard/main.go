package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const refreshInterval = 3 * time.Second

type tickMsg time.Time

type model struct {
	serverPID   string
	serverRSS   string
	serverCPU   string
	s3Status    string
	dataUsage   string
	logLines    []string
	lastRefresh string
	err         string
}

func (m model) Init() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.refresh()
		m.lastRefresh = time.Now().UTC().Format("15:04:05")
		return m, tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Fields    struct {
		Message string `json:"message"`
	} `json:"fields"`
}

func parseLog(line string) string {
	var e logEntry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		return line
	}
	ts := e.Timestamp
	if len(ts) > 19 {
		ts = ts[11:19] // just HH:MM:SS
	}
	lvl := e.Level
	if len(lvl) > 5 {
		lvl = lvl[:5]
	}
	msg := e.Fields.Message
	return fmt.Sprintf("%s %-5s %s", ts, lvl, msg)
}
func (m *model) refresh() {
	// RustFS server process
	out, err := exec.Command("pgrep", "-f", "/usr/local/bin/rustfs server").Output()
	if err == nil && len(out) > 0 {
		pid := strings.TrimSpace(string(out))
		pid = strings.Split(pid, "\n")[0]
		m.serverPID = pid

		// RSS memory
		data, _ := os.ReadFile(fmt.Sprintf("/proc/%s/status", pid))
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb := fields[1]
					mb := 0
					fmt.Sscanf(kb, "%d", &mb)
					m.serverRSS = fmt.Sprintf("%d MB", mb/1024)
				}
			}
		}

		// CPU from /proc/pid/stat
		statData, _ := os.ReadFile(fmt.Sprintf("/proc/%s/stat", pid))
		fields := strings.Fields(string(statData))
		if len(fields) >= 15 {
			// fields[13] = utime, fields[14] = stime (in clock ticks)
			utime, _ := strconv.Atoi(fields[13])
			stime, _ := strconv.Atoi(fields[14])
			total := utime + stime
			// Show CPU in jiffies (simple), or compute percentage vs uptime
			uptime, _ := os.ReadFile("/proc/uptime")
			uptimeSec, _ := strconv.ParseFloat(strings.Fields(string(uptime))[0], 64)
			hz := 100.0 // Linux clock ticks per second
			if uptimeSec > 0 {
				cpuPct := (float64(total) / hz / uptimeSec) * 100
				m.serverCPU = fmt.Sprintf("%.1f%%", cpuPct)
			} else {
				m.serverCPU = fmt.Sprintf("%d jiffies", total)
			}
		} else {
			m.serverCPU = "—"
		}
	} else {
		m.serverPID = "—"
		m.serverRSS = "—"
		m.serverCPU = "—"
	}

	// S3 API health
	httpOut, err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"http://localhost:9000").Output()
	if err == nil && strings.TrimSpace(string(httpOut)) != "000" {
		m.s3Status = "✓  http://0.0.0.0:9000  (" + strings.TrimSpace(string(httpOut)) + ")"
	} else {
		m.s3Status = "✗  not responding"
	}

	// Data usage (all disk mounts)
	duOut, _ := exec.Command("sh", "-c", "du -sh /data/disk* 2>/dev/null | awk '{s=$1; for(i=2;i<=NF;i++) printf \"%s%s\",(i>2?\" \":\"\"),$i; print \"=\"s}' | paste -sd ' ' -").Output()
	m.dataUsage = strings.TrimSpace(string(duOut))
	if m.dataUsage == "" {
		// fallback: /data
		duOut, _ = exec.Command("du", "-sh", "/data").Output()
		m.dataUsage = strings.Fields(strings.TrimSpace(string(duOut)))[0]
	}
	if m.dataUsage == "" {
		m.dataUsage = "(no disks)"
	}

	// Log tail (parsed)
	logOut, _ := exec.Command("tail", "-n", "8", "/var/log/rustfs.log").Output()
	rawLines := strings.Split(strings.TrimSpace(string(logOut)), "\n")
	m.logLines = nil
	for _, line := range rawLines {
		if line != "" {
			m.logLines = append(m.logLines, parseLog(line))
		}
	}
	if len(m.logLines) == 0 {
		m.logLines = []string{"(waiting for RustFS...)"}
	}
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("┌─── RustFS Appliance ────────────────────────────────────────────┐\n")

	// Server line
	b.WriteString(fmt.Sprintf("│  Server   %-4s  │  RSS %-7s  │  CPU %-5s                │\n",
		m.serverPID, m.serverRSS, m.serverCPU))

	// S3 API
	b.WriteString(fmt.Sprintf("│  S3 API   │  %-47s │\n", m.s3Status))

	// Console
	b.WriteString("│  Console  │  http://0.0.0.0:9001                                  │\n")

	// Data
	b.WriteString(fmt.Sprintf("│  Data     │  /data  %-48s │\n", m.dataUsage))

	b.WriteString("├─── Logs ──────────────────────────────────────────────────────────┤\n")

	// Log lines (show last 8, truncated)
	maxLogLines := 8
	start := 0
	if len(m.logLines) > maxLogLines {
		start = len(m.logLines) - maxLogLines
	}
	for i := start; i < len(m.logLines); i++ {
		line := m.logLines[i]
		if len(line) > 76 {
			line = line[:73] + "..."
		}
		b.WriteString(fmt.Sprintf("│  %s\n", line))
	}

	// Fill empty log space
	shown := len(m.logLines) - start
	if shown < 0 {
		shown = 0
	}
	for i := shown; i < maxLogLines; i++ {
		b.WriteString("│\n")
	}

	// Footer
	b.WriteString("└────────────────────────────────────────────────────────────────────┘\n")
	b.WriteString(fmt.Sprintf("  %s  |  Refresh: %ds  |  Read-only\n", m.lastRefresh, int(refreshInterval.Seconds())))

	return b.String()
}

func main() {
	m := model{}
	m.refresh()
	m.lastRefresh = time.Now().UTC().Format("15:04:05")

	p := tea.NewProgram(m, tea.WithInput(nil)) // no stdin — read-only
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dashboard: %v\n", err)
		os.Exit(1)
	}
}
