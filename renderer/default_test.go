package renderer

import (
	"github.com/olekukonko/tablewriter/tw"
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
				Borders: Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.On,
					},
					TrimWhitespace: tw.On,
					CompactMode:    tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
			},
		},
		{
			name: "PartialBorders",
			config: DefaultConfig{
				Borders: Border{Top: tw.Off},
			},
			expected: DefaultConfig{
				Borders: Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.On,
					},
					TrimWhitespace: tw.On,
					CompactMode:    tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
			},
		},
		{
			name: "PartialSettingsLines",
			config: DefaultConfig{
				Settings: Settings{
					Lines: Lines{ShowFooterLine: tw.Off},
				},
			},
			expected: DefaultConfig{
				Borders: Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
				Settings: Settings{
					Separators: Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.Off,
					},
					TrimWhitespace: tw.On,
					CompactMode:    tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
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
