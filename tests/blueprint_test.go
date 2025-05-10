package tests

import (
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestDefaultConfigMerging(t *testing.T) {
	tests := []struct {
		name     string
		config   tw.Rendition
		expected tw.Rendition
	}{
		{
			name:   "EmptyConfig",
			config: tw.Rendition{},
			expected: tw.Rendition{
				Borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: tw.Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.On,
					},
					// TrimWhitespace: tw.On,
					CompactMode: tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
			},
		},
		{
			name: "PartialBorders",
			config: tw.Rendition{
				Borders: tw.Border{Top: tw.Off},
			},
			expected: tw.Rendition{
				Borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.On},
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: tw.Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.On,
					},
					// TrimWhitespace: tw.On,
					CompactMode: tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
			},
		},
		{
			name: "PartialSettingsLines",
			config: tw.Rendition{
				Settings: tw.Settings{
					Lines: tw.Lines{ShowFooterLine: tw.Off},
				},
			},
			expected: tw.Rendition{
				Borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.On,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: tw.Lines{
						ShowTop:        tw.On,
						ShowBottom:     tw.On,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.Off,
					},
					// TrimWhitespace: tw.On,
					CompactMode: tw.Off,
				},
				Symbols: tw.NewSymbols(tw.StyleLight),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := renderer.NewBlueprint(tt.config)
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
