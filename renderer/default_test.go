package renderer

import (
	"github.com/olekukonko/tablewriter/symbols"
	"testing"
)

func TestDefaultConfigMerging(t *testing.T) {
	tests := []struct {
		name     string
		config   DefaultConfig
		expected DefaultConfig
	}{
		{
			name:   "EmptyConfig",
			config: DefaultConfig{},
			expected: DefaultConfig{
				Borders: Border{Left: On, Right: On, Top: On, Bottom: On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     On,
						ShowFooter:     On,
						BetweenRows:    Off,
						BetweenColumns: On,
					},
					Lines: Lines{
						ShowTop:        On,
						ShowBottom:     On,
						ShowHeaderLine: On,
						ShowFooterLine: On,
					},
					TrimWhitespace: On,
					CompactMode:    Off,
				},
				Symbols: symbols.NewSymbols(symbols.StyleLight),
			},
		},
		{
			name: "PartialBorders",
			config: DefaultConfig{
				Borders: Border{Top: Off},
			},
			expected: DefaultConfig{
				Borders: Border{Left: On, Right: On, Top: Off, Bottom: On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     On,
						ShowFooter:     On,
						BetweenRows:    Off,
						BetweenColumns: On,
					},
					Lines: Lines{
						ShowTop:        On,
						ShowBottom:     On,
						ShowHeaderLine: On,
						ShowFooterLine: On,
					},
					TrimWhitespace: On,
					CompactMode:    Off,
				},
				Symbols: symbols.NewSymbols(symbols.StyleLight),
			},
		},
		{
			name: "PartialSettingsLines",
			config: DefaultConfig{
				Settings: Settings{
					Lines: Lines{ShowFooterLine: Off},
				},
			},
			expected: DefaultConfig{
				Borders: Border{Left: On, Right: On, Top: On, Bottom: On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     On,
						ShowFooter:     On,
						BetweenRows:    Off,
						BetweenColumns: On,
					},
					Lines: Lines{
						ShowTop:        On,
						ShowBottom:     On,
						ShowHeaderLine: On,
						ShowFooterLine: Off,
					},
					TrimWhitespace: On,
					CompactMode:    Off,
				},
				Symbols: symbols.NewSymbols(symbols.StyleLight),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDefault(tt.config)
			got := r.Config()

			// Compare Borders
			if got.Borders != tt.expected.Borders {
				t.Errorf("%s: Borders mismatch - expected %+v, got %+v", tt.name, tt.expected.Borders, got.Borders)
			}
			// Compare Settings.Lines
			if got.Settings.Lines != tt.expected.Settings.Lines {
				t.Errorf("%s: Settings.Lines mismatch - expected %+v, got %+v", tt.name, tt.expected.Settings.Lines, got.Settings.Lines)
			}
			// Compare Settings.Separators
			if got.Settings.Separators != tt.expected.Settings.Separators {
				t.Errorf("%s: Settings.Separators mismatch - expected %+v, got %+v", tt.name, tt.expected.Settings.Separators, got.Settings.Separators)
			}
			// Check Symbols (basic presence check)
			if (tt.expected.Symbols == nil) != (got.Symbols == nil) {
				t.Errorf("%s: Symbols mismatch - expected nil: %v, got nil: %v", tt.name, tt.expected.Symbols == nil, got.Symbols == nil)
			}
		})
	}
}
