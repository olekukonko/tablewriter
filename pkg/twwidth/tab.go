package twwidth

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

type Tab rune

// TabWidthDefault is the fallback if detection fails (standard unix terminal default).
const (
	TabWidthDefault     = 8
	TabString       Tab = '\t'
)

// IsTab checks if this Tab instance equals default tab
func (t Tab) IsTab() bool {
	return t == TabString
}

func (t Tab) Byte() byte {
	return byte(t)
}

func (t Tab) Rune() rune {
	return rune(t)
}

func (t Tab) String() string {
	return string(t)
}

// IsTab checks if a rune is a tab
func IsTab(r rune) bool {
	return r == TabString.Rune()
}

// Tabinal is a live object whose Size() returns the current tab width.
type Tabinal struct {
	once  sync.Once
	width int // guarded by mu
	mu    sync.RWMutex
}

func (t *Tabinal) String() string { return TabString.String() }

func (t *Tabinal) Size() int {
	t.once.Do(t.init) // first call does the detection
	t.mu.RLock()
	w := t.width
	t.mu.RUnlock()
	if w <= 0 {
		return TabWidthDefault
	}
	return w
}

func (t *Tabinal) SetWidth(w int) {
	if w <= 0 || w > 32 {
		return
	}
	t.mu.Lock()
	t.width = w
	t.mu.Unlock()
}

// init runs exactly once per Tabinal instance.
func (t *Tabinal) init() {
	t.width = t.detect()
}

// detect returns the best guess > 0, or 0 to use TabWidth.
func (t *Tabinal) detect() int {
	// 1. Environment override (Safest & Fastest).
	if w := envInt("TABWIDTH"); w > 0 {
		return w
	}
	if w := envInt("TS"); w > 0 {
		return w
	}
	if w := envInt("VIM_TABSTOP"); w > 0 {
		return w
	}

	// 2. Terminfo (Passive file read).
	if w := t.terminfoIt(); w > 0 {
		return w
	}

	// 3. TERM heuristics (Passive string check).
	if w := t.termHeuristic(); w > 0 {
		return w
	}

	// 4. Ask the terminal directly (Active, Risky).
	if w := t.decRQSS(); w > 0 {
		return w
	}

	return 0 // use default
}

/* ---------- DECRQSS ---------- */

func (t *Tabinal) decRQSS() int {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}

	if err := os.Stdin.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return 0
	}
	defer os.Stdin.SetReadDeadline(time.Time{})

	old, err := term.MakeRaw(fd)
	if err != nil {
		return 0
	}
	defer term.Restore(fd, old)

	buf := make([]byte, 128)
	for {
		if err := os.Stdin.SetReadDeadline(time.Now().Add(1 * time.Millisecond)); err != nil {
			break
		}
		n, _ := os.Stdin.Read(buf)
		if n == 0 {
			break
		}
	}

	if _, err := os.Stdout.WriteString("\x1bP$qit\x1b\\"); err != nil {
		return 0
	}

	if err := os.Stdin.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return 0
	}

	n, _ := os.Stdin.Read(buf)
	resp := string(buf[:n])

	if i := strings.Index(resp, "\x1bP1$r"); i >= 0 {
		rest := resp[i+5:]
		if j := strings.Index(rest, "\x1b\\"); j >= 0 {
			if w, err := strconv.Atoi(rest[:j]); err == nil && w > 0 && w <= 32 {
				return w
			}
		}
	}
	return 0
}

func (t *Tabinal) terminfoIt() int {
	termEnv := os.Getenv("TERM")
	if termEnv == "" {
		return 0
	}

	paths := []string{
		"/usr/share/terminfo/" + string(termEnv[0]) + "/" + termEnv,
		"/lib/terminfo/" + string(termEnv[0]) + "/" + termEnv,
		os.Getenv("HOME") + "/.terminfo/" + string(termEnv[0]) + "/" + termEnv,
	}

	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err == nil {
			if i := strings.Index(string(b), "it#"); i >= 0 {
				s := string(b[i+3:])
				if end := strings.IndexAny(s, ",:\x00"); end >= 0 {
					if n, err := strconv.Atoi(s[:end]); err == nil && n > 0 {
						return n
					}
				}
			}
			return 0
		}
	}
	return 0
}

func (t *Tabinal) termHeuristic() int {
	termEnv := os.Getenv("TERM")
	if strings.Contains(termEnv, "vt52") {
		return 2
	}
	if strings.Contains(termEnv, "xterm") ||
		strings.Contains(termEnv, "screen") ||
		strings.Contains(termEnv, "linux") ||
		strings.Contains(termEnv, "ansi") {
		return 8
	}
	return 0
}

/* ---------- global singleton ---------- */

var (
	globalTab     *Tabinal
	globalTabOnce sync.Once
)

// TabInstance returns the singleton Tabinal instance.
// It performs detection on the first call to Size().
func TabInstance() *Tabinal {
	globalTabOnce.Do(func() {
		globalTab = &Tabinal{}
	})
	return globalTab
}

// TabWidth returns the detected width of a tab character.
func TabWidth() int {
	return TabInstance().Size()
}

func envInt(k string) int {
	v := os.Getenv(k)
	w, _ := strconv.Atoi(v)
	return w
}
