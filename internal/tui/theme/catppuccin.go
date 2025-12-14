// Package theme provides Catppuccin color schemes.
package theme

import "github.com/charmbracelet/lipgloss"

// Mocha returns the Catppuccin Mocha theme (dark).
func Mocha() *Theme {
	return &Theme{
		Name:   "Catppuccin Mocha",
		IsDark: true,

		// Primary colors
		Mauve:    lipgloss.Color("#cba6f7"),
		Blue:     lipgloss.Color("#89b4fa"),
		Green:    lipgloss.Color("#a6e3a1"),
		Yellow:   lipgloss.Color("#f9e2af"),
		Red:      lipgloss.Color("#f38ba8"),
		Peach:    lipgloss.Color("#fab387"),
		Teal:     lipgloss.Color("#94e2d5"),
		Pink:     lipgloss.Color("#f5c2e7"),
		Flamingo: lipgloss.Color("#f2cdcd"),

		// Text colors
		Text:    lipgloss.Color("#cdd6f4"),
		Subtext: lipgloss.Color("#a6adc8"),

		// Surface colors
		Surface:  lipgloss.Color("#313244"),
		Surface0: lipgloss.Color("#313244"),
		Surface1: lipgloss.Color("#45475a"),
		Base:     lipgloss.Color("#1e1e2e"),
		Mantle:   lipgloss.Color("#181825"),
		Crust:    lipgloss.Color("#11111b"),

		// Overlay colors
		Overlay0: lipgloss.Color("#6c7086"),
		Overlay1: lipgloss.Color("#7f849c"),
		Overlay2: lipgloss.Color("#9399b2"),
	}
}

// Macchiato returns the Catppuccin Macchiato theme (dark).
func Macchiato() *Theme {
	return &Theme{
		Name:   "Catppuccin Macchiato",
		IsDark: true,

		// Primary colors
		Mauve:    lipgloss.Color("#c6a0f6"),
		Blue:     lipgloss.Color("#8aadf4"),
		Green:    lipgloss.Color("#a6da95"),
		Yellow:   lipgloss.Color("#eed49f"),
		Red:      lipgloss.Color("#ed8796"),
		Peach:    lipgloss.Color("#f5a97f"),
		Teal:     lipgloss.Color("#8bd5ca"),
		Pink:     lipgloss.Color("#f5bde6"),
		Flamingo: lipgloss.Color("#f0c6c6"),

		// Text colors
		Text:    lipgloss.Color("#cad3f5"),
		Subtext: lipgloss.Color("#a5adcb"),

		// Surface colors
		Surface:  lipgloss.Color("#363a4f"),
		Surface0: lipgloss.Color("#363a4f"),
		Surface1: lipgloss.Color("#494d64"),
		Base:     lipgloss.Color("#24273a"),
		Mantle:   lipgloss.Color("#1e2030"),
		Crust:    lipgloss.Color("#181926"),

		// Overlay colors
		Overlay0: lipgloss.Color("#6e738d"),
		Overlay1: lipgloss.Color("#8087a2"),
		Overlay2: lipgloss.Color("#939ab7"),
	}
}

// Frappe returns the Catppuccin Frappe theme (dark).
func Frappe() *Theme {
	return &Theme{
		Name:   "Catppuccin Frappe",
		IsDark: true,

		// Primary colors
		Mauve:    lipgloss.Color("#ca9ee6"),
		Blue:     lipgloss.Color("#8caaee"),
		Green:    lipgloss.Color("#a6d189"),
		Yellow:   lipgloss.Color("#e5c890"),
		Red:      lipgloss.Color("#e78284"),
		Peach:    lipgloss.Color("#ef9f76"),
		Teal:     lipgloss.Color("#81c8be"),
		Pink:     lipgloss.Color("#f4b8e4"),
		Flamingo: lipgloss.Color("#eebebe"),

		// Text colors
		Text:    lipgloss.Color("#c6d0f5"),
		Subtext: lipgloss.Color("#a5adce"),

		// Surface colors
		Surface:  lipgloss.Color("#414559"),
		Surface0: lipgloss.Color("#414559"),
		Surface1: lipgloss.Color("#51576d"),
		Base:     lipgloss.Color("#303446"),
		Mantle:   lipgloss.Color("#292c3c"),
		Crust:    lipgloss.Color("#232634"),

		// Overlay colors
		Overlay0: lipgloss.Color("#737994"),
		Overlay1: lipgloss.Color("#838ba7"),
		Overlay2: lipgloss.Color("#949cbb"),
	}
}

// Latte returns the Catppuccin Latte theme (light).
func Latte() *Theme {
	return &Theme{
		Name:   "Catppuccin Latte",
		IsDark: false,

		// Primary colors
		Mauve:    lipgloss.Color("#8839ef"),
		Blue:     lipgloss.Color("#1e66f5"),
		Green:    lipgloss.Color("#40a02b"),
		Yellow:   lipgloss.Color("#df8e1d"),
		Red:      lipgloss.Color("#d20f39"),
		Peach:    lipgloss.Color("#fe640b"),
		Teal:     lipgloss.Color("#179299"),
		Pink:     lipgloss.Color("#ea76cb"),
		Flamingo: lipgloss.Color("#dd7878"),

		// Text colors
		Text:    lipgloss.Color("#4c4f69"),
		Subtext: lipgloss.Color("#6c6f85"),

		// Surface colors
		Surface:  lipgloss.Color("#ccd0da"),
		Surface0: lipgloss.Color("#ccd0da"),
		Surface1: lipgloss.Color("#bcc0cc"),
		Base:     lipgloss.Color("#eff1f5"),
		Mantle:   lipgloss.Color("#e6e9ef"),
		Crust:    lipgloss.Color("#dce0e8"),

		// Overlay colors
		Overlay0: lipgloss.Color("#9ca0b0"),
		Overlay1: lipgloss.Color("#8c8fa1"),
		Overlay2: lipgloss.Color("#7c7f93"),
	}
}
