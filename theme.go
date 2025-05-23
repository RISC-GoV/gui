package main

import (
	"github.com/therecipe/qt/core"
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

func applyTheme(themeName string) {
	currentTheme = themeName

	// Create application-wide stylesheet based on theme
	var styleSheet string

	if themeName == ThemeDark {
		preferences.ThemeSettings.LineNumberAreaColor = gui.NewQColor3(45, 45, 45, 255)
		// Dark theme styles
		styleSheet = `
			* {
				transition: background-color 0ms, color 0ms, border 0ms;
			}

			QWidget {
				background-color: #1e1e1e;
				color: #dcdcdc;
			}
			
			QMenuBar {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border-bottom: 1px solid #444;
			}
			
			QMenuBar::item:selected {
				background-color: #3e3e3e;
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
				spacing: 3px;
			}
			
			QToolButton {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: none;
				border-radius: 3px;
				padding: 3px;
			}
			
			QToolButton:hover {
				background-color: #3e3e3e;
			}
			
			QToolButton:pressed {
				background-color: #505050;
			}
			
			QLineEdit {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
				border-radius: 2px;
				padding: 2px;
			}
			
			QPlainTextEdit, QTextEdit {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
				selection-background-color: #264f78;
				selection-color: #ffffff;
			}
			
			QTreeView, QListView, QTableView {
				background-color: #1c1c1c;
				color: #dcdcdc;
				border: 1px solid #444;
				alternate-background-color: #262626;
			}
			
			QTreeView::item:selected, QListView::item:selected, QTableView::item:selected {
				background-color: #264f78;
				color: #ffffff;
			}
			
			QTreeView::item:hover, QListView::item:hover, QTableView::item:hover {
				background-color: #323232;
			}
			
			QTreeView::branch {
				background-color: #1c1c1c;
			}
			
			QHeaderView::section {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				padding: 4px;
			}
			
			QPushButton {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				padding: 5px 15px;
				border-radius: 3px;
			}
			
			QPushButton:hover {
				background-color: #3e3e3e;
			}
			
			QPushButton:pressed {
				background-color: #505050;
			}
			
			QPushButton:disabled {
				background-color: #1e1e1e;
				color: #666666;
				border: 1px solid #333;
			}
			
			QTabWidget::pane {
				border: 1px solid #444;
				background-color: #1e1e1e;
			}
			
			QTabBar::tab {
				background-color: #2d2d2d;
				color: #b0b0b0;
				padding: 5px 10px;
				border: 1px solid #444;
				border-bottom: none;
				border-top-left-radius: 3px;
				border-top-right-radius: 3px;
			}
			
			QTabBar::tab:selected {
				background-color: #1e1e1e;
				color: #dcdcdc;
			}
			
			QTabBar::tab:hover:!selected {
				background-color: #3e3e3e;
			}
			
			QStatusBar {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border-top: 1px solid #444;
			}
			
			QScrollBar:vertical {
				background-color: #292929;
				width: 14px;
				margin: 14px 0px 14px 0px;
			}
			
			QScrollBar::handle:vertical {
				background-color: #555555;
				min-height: 20px;
				border-radius: 3px;
			}
			
			QScrollBar::handle:vertical:hover {
				background-color: #666666;
			}
			
			QScrollBar::add-line:vertical, QScrollBar::sub-line:vertical {
				border: none;
				background: none;
				height: 14px;
			}
			
			QScrollBar:horizontal {
				background-color: #292929;
				height: 14px;
				margin: 0px 14px 0px 14px;
			}
			
			QScrollBar::handle:horizontal {
				background-color: #555555;
				min-width: 20px;
				border-radius: 3px;
			}
			
			QScrollBar::handle:horizontal:hover {
				background-color: #666666;
			}
			
			QScrollBar::add-line:horizontal, QScrollBar::sub-line:horizontal {
				border: none;
				background: none;
				width: 14px;
			}
			
			QLabel {
				color: #dcdcdc;
			}
			
			QComboBox {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				border-radius: 3px;
				padding: 2px 8px;
			}
			
			QComboBox::drop-down {
				subcontrol-origin: padding;
				subcontrol-position: top right;
				width: 20px;
				border-left: 1px solid #444;
			}
			
			QComboBox QAbstractItemView {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				selection-background-color: #3e3e3e;
			}
			
			QSpinBox, QDoubleSpinBox {
				background-color: #2d2d2d;
				color: #dcdcdc;
				border: 1px solid #444;
				border-radius: 3px;
				padding: 2px 5px;
			}
			
			QGroupBox {
				border: 1px solid #444;
				border-radius: 5px;
				margin-top: 8px;
				padding-top: 8px;
			}
			
			QGroupBox::title {
				subcontrol-origin: margin;
				subcontrol-position: top left;
				left: 10px;
				color: #dcdcdc;
			}
		`
	} else {
		preferences.ThemeSettings.LineNumberAreaColor = gui.NewQColor3(240, 240, 240, 255)
		// Light theme styles
		styleSheet = `
			* {
				transition: background-color 0ms, color 0ms, border 0ms;
			}
			
			QWidget {
				background-color: #f5f5f5;
				color: #212121;
			}
			
			QMenuBar {
				background-color: #f0f0f0;
				color: #212121;
				border-bottom: 1px solid #e0e0e0;
			}
			
			QMenuBar::item:selected {
				background-color: #e3f2fd;
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
				background-color: #f0f0f0;
				border-bottom: 1px solid #e0e0e0;
				spacing: 3px;
			}
			
			QToolButton {
				background-color: #f0f0f0;
				color: #212121;
				border: none;
				border-radius: 3px;
				padding: 3px;
			}
			
			QToolButton:hover {
				background-color: #e3f2fd;
			}
			
			QToolButton:pressed {
				background-color: #bbdefb;
			}
			
			QLineEdit {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
				border-radius: 2px;
				padding: 2px;
			}
			
			QPlainTextEdit, QTextEdit {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
				selection-background-color: #bbdefb;
				selection-color: #212121;
			}
			
			QTreeView, QListView, QTableView {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
				alternate-background-color: #f9f9f9;
			}
			
			QTreeView::item:selected, QListView::item:selected, QTableView::item:selected {
				background-color: #bbdefb;
				color: #212121;
			}
			
			QTreeView::item:hover, QListView::item:hover, QTableView::item:hover {
				background-color: #e3f2fd;
			}
			
			QTreeView::branch {
				background-color: #ffffff;
			}
			
			QHeaderView::section {
				background-color: #f0f0f0;
				color: #212121;
				border: 1px solid #e0e0e0;
				padding: 4px;
			}
			
			QPushButton {
				background-color: #f0f0f0;
				color: #212121;
				border: 1px solid #e0e0e0;
				padding: 5px 15px;
				border-radius: 3px;
			}
			
			QPushButton:hover {
				background-color: #e3f2fd;
			}
			
			QPushButton:pressed {
				background-color: #bbdefb;
			}
			
			QPushButton:disabled {
				background-color: #f5f5f5;
				color: #9e9e9e;
				border: 1px solid #e0e0e0;
			}
			
			QTabWidget::pane {
				border: 1px solid #e0e0e0;
				background-color: #ffffff;
			}
			
			QTabBar::tab {
				background-color: #f0f0f0;
				color: #757575;
				padding: 5px 10px;
				border: 1px solid #e0e0e0;
				border-bottom: none;
				border-top-left-radius: 3px;
				border-top-right-radius: 3px;
			}
			
			QTabBar::tab:selected {
				background-color: #ffffff;
				color: #212121;
			}
			
			QTabBar::tab:hover:!selected {
				background-color: #e3f2fd;
			}
			
			QStatusBar {
				background-color: #f0f0f0;
				color: #212121;
				border-top: 1px solid #e0e0e0;
			}
			
			QScrollBar:vertical {
				background-color: #f0f0f0;
				width: 14px;
				margin: 14px 0px 14px 0px;
			}
			
			QScrollBar::handle:vertical {
				background-color: #bdbdbd;
				min-height: 20px;
				border-radius: 3px;
			}
			
			QScrollBar::handle:vertical:hover {
				background-color: #9e9e9e;
			}
			
			QScrollBar::add-line:vertical, QScrollBar::sub-line:vertical {
				border: none;
				background: none;
				height: 14px;
			}
			
			QScrollBar:horizontal {
				background-color: #f0f0f0;
				height: 14px;
				margin: 0px 14px 0px 14px;
			}
			
			QScrollBar::handle:horizontal {
				background-color: #bdbdbd;
				min-width: 20px;
				border-radius: 3px;
			}
			
			QScrollBar::handle:horizontal:hover {
				background-color: #9e9e9e;
			}
			
			QScrollBar::add-line:horizontal, QScrollBar::sub-line:horizontal {
				border: none;
				background: none;
				width: 14px;
			}
			
			QLabel {
				color: #212121;
			}
			
			QComboBox {
				background-color: #f0f0f0;
				color: #212121;
				border: 1px solid #e0e0e0;
				border-radius: 3px;
				padding: 2px 8px;
			}
			
			QComboBox::drop-down {
				subcontrol-origin: padding;
				subcontrol-position: top right;
				width: 20px;
				border-left: 1px solid #e0e0e0;
			}
			
			QComboBox QAbstractItemView {
				background-color: #ffffff;
				color: #212121;
				border: 1px solid #e0e0e0;
				selection-background-color: #e3f2fd;
			}
			
			QSpinBox, QDoubleSpinBox {
				background-color: #f0f0f0;
				color: #212121;
				border: 1px solid #e0e0e0;
				border-radius: 3px;
				padding: 2px 5px;
			}
			
			QGroupBox {
				border: 1px solid #e0e0e0;
				border-radius: 5px;
				margin-top: 8px;
				padding-top: 8px;
			}
			
			QGroupBox::title {
				subcontrol-origin: margin;
				subcontrol-position: top left;
				left: 10px;
				color: #212121;
			}
		`
	}

	// Apply stylesheet to application
	app.SetStyleSheet(styleSheet)
	// Force immediate update to prevent white flash
	app.ProcessEvents(core.QEventLoop__AllEvents)
}

type HighlightRule struct {
	Pattern string
	Format  *gui.QTextCharFormat
	Color   *gui.QColor
	Bold    bool
}
