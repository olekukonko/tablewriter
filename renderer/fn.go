package renderer

import "github.com/olekukonko/tablewriter/tw"

func defaultConfig() DefaultConfig {
	return DefaultConfig{
		Borders: Border{
			Left:   tw.On,
			Right:  tw.On,
			Top:    tw.On,
			Bottom: tw.On,
		},
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
	}
}

func NewDefault(configs ...DefaultConfig) *Default {
	cfg := defaultConfig()
	cfg.Debug = true
	if len(configs) > 0 {
		userCfg := configs[0]
		if userCfg.Borders.Left != 0 {
			cfg.Borders.Left = userCfg.Borders.Left
		}
		if userCfg.Borders.Right != 0 {
			cfg.Borders.Right = userCfg.Borders.Right
		}
		if userCfg.Borders.Top != 0 {
			cfg.Borders.Top = userCfg.Borders.Top
		}
		if userCfg.Borders.Bottom != 0 {
			cfg.Borders.Bottom = userCfg.Borders.Bottom
		}
		if userCfg.Symbols != nil {
			cfg.Symbols = userCfg.Symbols
		}
		cfg.Settings = mergeSettings(cfg.Settings, userCfg.Settings)
	}
	return &Default{config: cfg}
}

func mergeSettings(defaults, overrides Settings) Settings {
	if overrides.Separators.ShowHeader != 0 {
		defaults.Separators.ShowHeader = overrides.Separators.ShowHeader
	}
	if overrides.Separators.ShowFooter != 0 {
		defaults.Separators.ShowFooter = overrides.Separators.ShowFooter
	}
	if overrides.Separators.BetweenRows != 0 {
		defaults.Separators.BetweenRows = overrides.Separators.BetweenRows
	}
	if overrides.Separators.BetweenColumns != 0 {
		defaults.Separators.BetweenColumns = overrides.Separators.BetweenColumns
	}
	if overrides.Lines.ShowTop != 0 {
		defaults.Lines.ShowTop = overrides.Lines.ShowTop
	}
	if overrides.Lines.ShowBottom != 0 {
		defaults.Lines.ShowBottom = overrides.Lines.ShowBottom
	}
	if overrides.Lines.ShowHeaderLine != 0 {
		defaults.Lines.ShowHeaderLine = overrides.Lines.ShowHeaderLine
	}
	if overrides.Lines.ShowFooterLine != 0 {
		defaults.Lines.ShowFooterLine = overrides.Lines.ShowFooterLine
	}
	if overrides.TrimWhitespace != 0 {
		defaults.TrimWhitespace = overrides.TrimWhitespace
	}
	if overrides.CompactMode != 0 {
		defaults.CompactMode = overrides.CompactMode
	}
	return defaults
}
