package twwidth

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

type Tab rune

// TabWidthDefault is the fallback if detection fails.
const (
	TabWidthDefault     = 4
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
	width int
	mu    sync.RWMutex
}

func (t *Tabinal) String() string { return TabString.String() }

func (t *Tabinal) Size() int {
	t.once.Do(t.init)

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
	w := t.detect()

	t.mu.Lock()
	t.width = w
	t.mu.Unlock()
}

// detect returns the best guess > 0, or 0 to use TabWidthDefault.
func (t *Tabinal) detect() int {
	// 1. Environment override (explicit always wins)
	if w := envInt("TABWIDTH"); w > 0 {
		return clamp(w)
	}
	if w := envInt("TS"); w > 0 {
		return clamp(w)
	}
	if w := envInt("VIM_TABSTOP"); w > 0 {
		return clamp(w)
	}

	// 2. TERM heuristics (safe + deterministic)
	if w := termHeuristic(); w > 0 {
		return w
	}

	return 0
}

func termHeuristic() int {
	termEnv := strings.ToLower(os.Getenv("TERM"))
	if termEnv == "" {
		return 0
	}

	if strings.Contains(termEnv, "vt52") {
		return 2
	}

	if strings.Contains(termEnv, "xterm") ||
		strings.Contains(termEnv, "screen") ||
		strings.Contains(termEnv, "tmux") ||
		strings.Contains(termEnv, "linux") ||
		strings.Contains(termEnv, "ansi") ||
		strings.Contains(termEnv, "rxvt") {
		return 8
	}

	return 0
}

func clamp(w int) int {
	if w <= 0 {
		return 0
	}
	if w > 32 {
		return 32
	}
	return w
}

/* ---------- global singleton ---------- */

var (
	globalTab     *Tabinal
	globalTabOnce sync.Once
)

// TabInstance returns the singleton Tabinal instance.
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
