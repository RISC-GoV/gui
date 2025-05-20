package main

import (
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Theme constants
const (
	ThemeLight = "Light"
	ThemeDark  = "Dark"
)

// Global theme variable
var currentTheme string

// Let's update the preference-related code to work with our new theme system

func createThemeSettingsTab() *widgets.QWidget {
	tab := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	tab.SetLayout(layout)

	// Create form layout for settings
	formLayout := widgets.NewQFormLayout(nil)

	// Theme selector (Light/Dark)
	themeCombo = widgets.NewQComboBox(nil)
	themeCombo.AddItems([]string{
		"Light",
		"Dark",
	})

	// Set current theme
	if preferences.ThemeSettings.DarkMode {
		themeCombo.SetCurrentText("Dark")
	} else {
		themeCombo.SetCurrentText("Light")
	}

	// Preview of selected theme
	previewGroupBox := widgets.NewQGroupBox2("Theme Preview", nil)
	previewLayout := widgets.NewQVBoxLayout()
	previewGroupBox.SetLayout(previewLayout)

	// Create a preview widget to show how the theme looks
	previewWidget := widgets.NewQWidget(nil, 0)
	previewWidgetLayout := widgets.NewQVBoxLayout()
	previewWidget.SetLayout(previewWidgetLayout)

	// Add a toolbar to the preview
	previewToolBar := widgets.NewQToolBar2(previewWidget)
	previewToolBar.AddAction("File")
	previewToolBar.AddAction("Edit")
	previewToolBar.AddAction("Run")
	previewWidgetLayout.AddWidget(previewToolBar, 0, 0)

	// Add a code editor preview
	previewEditor := widgets.NewQPlainTextEdit(nil)
	previewEditor.SetPlainText("// Sample RISC-V Assembly\n.global _start\n\n_start:\n    li a0, 1       # File descriptor (stdout)\n    la a1, message  # Message address\n    li a2, 13      # Message length\n    li a7, 64      # syscall: write\n    ecall")
	previewEditor.SetReadOnly(true)
	previewWidgetLayout.AddWidget(previewEditor, 0, 0)

	// Add buttons
	buttonLayout := widgets.NewQHBoxLayout()
	runButton := widgets.NewQPushButton2("Run", nil)
	debugButton := widgets.NewQPushButton2("Debug", nil)
	buttonLayout.AddWidget(runButton, 0, 0)
	buttonLayout.AddWidget(debugButton, 0, 0)
	buttonLayout.AddStretch(1)
	previewWidgetLayout.AddLayout(buttonLayout, 0)

	previewLayout.AddWidget(previewWidget, 0, 0)

	// Connect theme changes to update preview in real-time
	themeCombo.ConnectCurrentTextChanged(func(text string) {
		updateThemePreview(previewWidget, text == "Dark")
	})

	// Add widgets to layout
	formLayout.AddRow3("Theme:", themeCombo)

	// Color customization note
	noteLabel := widgets.NewQLabel2("Theme colors are optimized for code visibility and readability.", nil, 0)
	noteLabel.SetWordWrap(true)

	// Add everything to main layout
	layout.AddLayout(formLayout, 0)
	layout.AddWidget(noteLabel, 0, 0)
	layout.AddSpacing(15)
	layout.AddWidget(previewGroupBox, 1, 0) // Give the preview some stretch

	return tab
}

// Update theme preview when user selects a different theme
func updateThemePreview(previewWidget *widgets.QWidget, isDarkMode bool) {
	// Set preview stylesheet based on selected theme
	if isDarkMode {
		// Dark mode preview
		previewWidget.SetStyleSheet(`
			QWidget {
				background-color: #1e1e1e;
				color: #dcdcdc;
			}
			
			QToolBar {
				background-color: #2d2d2d;
				border-bottom: 1px solid #444;
			}
			
			QToolBar QToolButton {
				color: #dcdcdc;
			}
			
			QPlainTextEdit {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
				border-radius: 3px;
			}
			
			QPushButton {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				border-radius: 3px;
				padding: 5px 15px;
			}
		`)
	} else {
		// Light mode preview
		previewWidget.SetStyleSheet(`
			QWidget {
				background-color: #fafafa;
				color: #212121;
			}
			
			QToolBar {
				background-color: #f5f5f5;
				border-bottom: 1px solid #e0e0e0;
			}
			
			QToolBar QToolButton {
				color: #424242;
			}
			
			QPlainTextEdit {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
				border-radius: 3px;
			}
			
			QPushButton {
				background-color: #f5f5f5;
				color: #424242;
				border: 1px solid #e0e0e0;
				border-radius: 3px;
				padding: 5px 15px;
			}
		`)
	}
}

func SetTheme(darkMode bool) {
	preferences.ThemeSettings.DarkMode = darkMode
	preferences.ThemeSettings.ThemeName = ThemeDark
	if !darkMode {
		preferences.ThemeSettings.ThemeName = ThemeLight
	}
	_ = SavePreferences()

	// Apply the theme
	applyTheme(preferences.ThemeSettings.ThemeName)
}

// Replace the existing applyModernTheme function
func applyModernTheme() {
	// Set default theme (will be overridden by preferences)
	currentTheme = ThemeLight

	// Apply default theme (light)
	applyTheme(currentTheme)
}

// Apply the selected theme to the application
func applyTheme(themeName string) {
	currentTheme = themeName

	// Create application-wide stylesheet based on theme
	var styleSheet string

	if themeName == ThemeDark {
		// Dark theme styles
		styleSheet = `
			QWidget {
				background-color: #1e1e1e;
				color: #dcdcdc;
			}
			
			QMenuBar {
				background-color: #2d2d2d;
				color: #dcdcdc;
			}
			
			QMenu {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
			}
			
			QMenu::item:selected {
				background-color: #3e3e3e;
			}
			
			QToolBar {
				background-color: #2d2d2d;
				border-bottom: 1px solid #444;
			}
			
			QPlainTextEdit, QTextEdit {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
			}
			
			QTreeView {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
			}
			
			QTreeView::item:selected {
				background-color: #264f78;
			}
			
			QPushButton {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				padding: 5px 15px;
				border-radius: 3px;
			}
			
			QTabWidget::pane {
				border: 1px solid #444;
			}
			
			QTabBar::tab {
				background-color: #2d2d2d;
				color: #dcdcdc;
				padding: 5px 10px;
				border: 1px solid #444;
				border-bottom: none;
			}
			
			QTabBar::tab:selected {
				background-color: #1e1e1e;
			}
			
			QStatusBar {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border-top: 1px solid #444;
			}
		`
	} else {
		// Light theme styles
		styleSheet = `
			QWidget {
				background-color: #fafafa;
				color: #212121;
			}
			
			QMenuBar {
				background-color: #f5f5f5;
				color: #212121;
			}
			
			QMenu {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
			}
			
			QMenu::item:selected {
				background-color: #e3f2fd;
			}
			
			QToolBar {
				background-color: #f5f5f5;
				border-bottom: 1px solid #e0e0e0;
			}
			
			QPlainTextEdit, QTextEdit {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
			}
			
			QTreeView {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
			}
			
			QTreeView::item:selected {
				background-color: #e3f2fd;
			}
			
			QPushButton {
				background-color: #f5f5f5;
				color: #212121;
				border: 1px solid #e0e0e0;
				padding: 5px 15px;
				border-radius: 3px;
			}
			
			QTabWidget::pane {
				border: 1px solid #e0e0e0;
			}
			
			QTabBar::tab {
				background-color: #f5f5f5;
				color: #212121;
				padding: 5px 10px;
				border: 1px solid #e0e0e0;
				border-bottom: none;
			}
			
			QTabBar::tab:selected {
				background-color: #ffffff;
			}
			
			QStatusBar {
				background-color: #f5f5f5;
				color: #212121;
				border-top: 1px solid #e0e0e0;
			}
		`
	}

	// Apply stylesheet to application
	app.SetStyleSheet(styleSheet)
}

// Setup syntax highlighting
func setupSyntaxHighlighting() {
	// Implement syntax highlighting based on current theme
	var highlightRules []HighlightRule

	if currentTheme == ThemeDark {
		// Dark theme syntax highlighting rules
		highlightRules = []HighlightRule{
			{Pattern: `\b(def|if|else|while|for|return|import|from|as|class|try|except|finally|with|lambda|yield|break|continue|pass|global|nonlocal|in|is|not|and|or)\b`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(86, 156, 214, 255)}, // Keywords
			{Pattern: `".*?"`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(206, 145, 120, 255)},                                                                                                                                            // String literals
			{Pattern: `'.*?'`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(206, 145, 120, 255)},                                                                                                                                            // String literals
			{Pattern: `\b\d+\b`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(181, 206, 168, 255)},                                                                                                                                          // Numbers
			{Pattern: `#.*$`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(106, 153, 85, 255)},                                                                                                                                              // Comments
		}
	} else {
		// Light theme syntax highlighting rules
		highlightRules = []HighlightRule{
			{Pattern: `\b(def|if|else|while|for|return|import|from|as|class|try|except|finally|with|lambda|yield|break|continue|pass|global|nonlocal|in|is|not|and|or)\b`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(0, 0, 255, 255)}, // Keywords
			{Pattern: `".*?"`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(163, 21, 21, 255)},                                                                                                                                           // String literals
			{Pattern: `'.*?'`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(163, 21, 21, 255)},                                                                                                                                           // String literals
			{Pattern: `\b\d+\b`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(9, 136, 90, 255)},                                                                                                                                          // Numbers
			{Pattern: `#.*$`, Format: gui.NewQTextCharFormat(), Color: gui.NewQColor3(0, 128, 0, 255)},                                                                                                                                              // Comments
		}
	}

	// Apply highlighting rules to editor
	for _, rule := range highlightRules {
		rule.Format.SetForeground(gui.NewQBrush3(rule.Color, 1))
		if rule.Bold {
			rule.Format.SetFontWeight(75) // Bold
		}
	}
}

type HighlightRule struct {
	Pattern string
	Format  *gui.QTextCharFormat
	Color   *gui.QColor
	Bold    bool
}
