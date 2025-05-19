package main

import (
	"encoding/json"
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"os"
	"path/filepath"
)

type UserPreferences struct {
	LastOpenedProject string   `json:"lastOpenedProject"`
	RecentFiles       []string `json:"recentFiles"`
	EditorSettings    struct {
		FontFamily      string `json:"fontFamily"`
		FontSize        int    `json:"fontSize"`
		TabWidth        int    `json:"tabWidth"`
		ShowLineNumbers bool   `json:"showLineNumbers"`
		WrapText        bool   `json:"wrapText"`
	} `json:"editorSettings"`
	WindowSettings struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		X      int `json:"x"`
		Y      int `json:"y"`
	} `json:"windowSettings"`
	ThemeSettings struct {
		DarkMode  bool   `json:"darkMode"`
		ThemeName string `json:"themeName"`
	} `json:"themeSettings"`
	AutoSaveEnabled  bool `json:"autoSaveEnabled"`
	AutoSaveInterval int  `json:"autoSaveInterval"` // In seconds
}

var preferences UserPreferences

var preferencesPath string

func InitPreferences() error {
	// Determine preferences file location based on OS
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %v", err)
	}

	// Create RISC-GoV IDE config directory if it doesn't exist
	configDir := filepath.Join(userConfigDir, "RISC-GoV-IDE")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	preferencesPath = filepath.Join(configDir, "preferences.json")

	// Load preferences if they exist, otherwise create default
	if _, err := os.Stat(preferencesPath); os.IsNotExist(err) {
		// Create default preferences
		preferences = getDefaultPreferences()
		return SavePreferences()
	}

	// Load existing preferences
	data, err := ioutil.ReadFile(preferencesPath)
	if err != nil {
		return fmt.Errorf("failed to read preferences file: %v", err)
	}

	if err := json.Unmarshal(data, &preferences); err != nil {
		return fmt.Errorf("failed to parse preferences file: %v", err)
	}

	return nil
}

func getDefaultPreferences() UserPreferences {
	prefs := UserPreferences{
		RecentFiles:      []string{},
		AutoSaveEnabled:  true,
		AutoSaveInterval: 60, // Save every 60 seconds
	}

	// Default editor settings
	prefs.EditorSettings.FontFamily = "Courier New"
	prefs.EditorSettings.FontSize = 12
	prefs.EditorSettings.TabWidth = 4
	prefs.EditorSettings.ShowLineNumbers = true
	prefs.EditorSettings.WrapText = false

	// Default window settings
	prefs.WindowSettings.Width = 1200
	prefs.WindowSettings.Height = 800
	prefs.WindowSettings.X = 100
	prefs.WindowSettings.Y = 100

	// Default theme settings
	prefs.ThemeSettings.DarkMode = false
	prefs.ThemeSettings.ThemeName = "default"

	return prefs
}

func SavePreferences() error {
	data, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %v", err)
	}

	if err := ioutil.WriteFile(preferencesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %v", err)
	}

	return nil
}

func AddRecentFile(filePath string) {
	// Check if file is already in the list
	for i, path := range preferences.RecentFiles {
		if path == filePath {
			// If it is, move it to the front
			preferences.RecentFiles = append(preferences.RecentFiles[:i], preferences.RecentFiles[i+1:]...)
			preferences.RecentFiles = append([]string{filePath}, preferences.RecentFiles...)
			SavePreferences()
			return
		}
	}

	// Add file to the front of the list
	preferences.RecentFiles = append([]string{filePath}, preferences.RecentFiles...)

	// Limit list to 10 recent files
	if len(preferences.RecentFiles) > 10 {
		preferences.RecentFiles = preferences.RecentFiles[:10]
	}

	SavePreferences()
}

func SetLastOpenedProject(projectPath string) {
	preferences.LastOpenedProject = projectPath
	SavePreferences()
}

func UpdateWindowSettings(width, height, x, y int) {
	preferences.WindowSettings.Width = width
	preferences.WindowSettings.Height = height
	preferences.WindowSettings.X = x
	preferences.WindowSettings.Y = y
	SavePreferences()
}

func SetTheme(darkMode bool, themeName string) {
	preferences.ThemeSettings.DarkMode = darkMode
	preferences.ThemeSettings.ThemeName = themeName
	SavePreferences()
}

func SetEditorSettings(fontFamily string, fontSize, tabWidth int, showLineNumbers, wrapText bool) {
	preferences.EditorSettings.FontFamily = fontFamily
	preferences.EditorSettings.FontSize = fontSize
	preferences.EditorSettings.TabWidth = tabWidth
	preferences.EditorSettings.ShowLineNumbers = showLineNumbers
	preferences.EditorSettings.WrapText = wrapText
	SavePreferences()
}

func SetAutoSave(enabled bool, interval int) {
	preferences.AutoSaveEnabled = enabled
	preferences.AutoSaveInterval = interval
	SavePreferences()
}

func showPreferencesDialog() {
	dialog := widgets.NewQDialog(mainWindow, 0)
	dialog.SetWindowTitle("Preferences")
	dialog.Resize2(500, 400)

	// Create tabbed interface
	tabs := widgets.NewQTabWidget(dialog)

	// Main layout
	mainLayout := widgets.NewQVBoxLayout()
	dialog.SetLayout(mainLayout)
	mainLayout.AddWidget(tabs, 0, 0)

	// Add tabs
	editorTab := createEditorSettingsTab()
	themeTab := createThemeSettingsTab()
	generalTab := createGeneralSettingsTab()

	tabs.AddTab(generalTab, "General")
	tabs.AddTab(editorTab, "Editor")
	tabs.AddTab(themeTab, "Appearance")

	// Button box
	buttonBox := widgets.NewQDialogButtonBox2(core.Qt__Horizontal, dialog)
	buttonBox.SetStandardButtons(widgets.QDialogButtonBox__Ok | widgets.QDialogButtonBox__Cancel)
	mainLayout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		// Save all settings
		savePreferencesFromUI()
		dialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		dialog.Reject()
	})

	dialog.Exec()
}

var (
	fontFamilyCombo         *widgets.QComboBox
	fontSizeSpinner         *widgets.QSpinBox
	tabWidthSpinner         *widgets.QSpinBox
	lineNumbersCheck        *widgets.QCheckBox
	wrapTextCheck           *widgets.QCheckBox
	darkModeCheck           *widgets.QCheckBox
	themeCombo              *widgets.QComboBox
	autoSaveCheck           *widgets.QCheckBox
	autoSaveIntervalSpinner *widgets.QSpinBox
)

func createEditorSettingsTab() *widgets.QWidget {
	tab := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	tab.SetLayout(layout)

	// Font family
	fontFamilyCombo = widgets.NewQComboBox(nil)
	fontFamilyCombo.AddItems([]string{
		"Courier New",
		"Consolas",
		"Monospace",
		"Monaco",
		"Source Code Pro",
	})
	fontFamilyCombo.SetCurrentText(preferences.EditorSettings.FontFamily)
	layout.AddRow3("Font:", fontFamilyCombo)

	// Font size
	fontSizeSpinner = widgets.NewQSpinBox(nil)
	fontSizeSpinner.SetRange(8, 24)
	fontSizeSpinner.SetValue(preferences.EditorSettings.FontSize)
	layout.AddRow3("Font Size:", fontSizeSpinner)

	// Tab width
	tabWidthSpinner = widgets.NewQSpinBox(nil)
	tabWidthSpinner.SetRange(2, 8)
	tabWidthSpinner.SetValue(preferences.EditorSettings.TabWidth)
	layout.AddRow3("Tab Width:", tabWidthSpinner)

	// Line numbers
	lineNumbersCheck = widgets.NewQCheckBox(nil)
	lineNumbersCheck.SetChecked(preferences.EditorSettings.ShowLineNumbers)
	layout.AddRow3("Show Line Numbers:", lineNumbersCheck)

	// Text wrapping
	wrapTextCheck = widgets.NewQCheckBox(nil)
	wrapTextCheck.SetChecked(preferences.EditorSettings.WrapText)
	layout.AddRow3("Wrap Text:", wrapTextCheck)

	return tab
}

func createThemeSettingsTab() *widgets.QWidget {
	tab := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	tab.SetLayout(layout)

	// Dark mode
	darkModeCheck = widgets.NewQCheckBox(nil)
	darkModeCheck.SetChecked(preferences.ThemeSettings.DarkMode)
	layout.AddRow3("Dark Mode:", darkModeCheck)

	// Theme selector
	themeCombo = widgets.NewQComboBox(nil)
	themeCombo.AddItems([]string{
		"Default",
		"Monokai",
		"Solarized",
	})

	// Set current theme
	index := themeCombo.FindText(preferences.ThemeSettings.ThemeName, core.Qt__MatchFixedString)
	if index >= 0 {
		themeCombo.SetCurrentIndex(index)
	}

	layout.AddRow3("Theme:", themeCombo)

	return tab
}

func createGeneralSettingsTab() *widgets.QWidget {
	tab := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQFormLayout(nil)
	tab.SetLayout(layout)

	// Auto-save
	autoSaveCheck = widgets.NewQCheckBox(nil)
	autoSaveCheck.SetChecked(preferences.AutoSaveEnabled)
	layout.AddRow3("Enable Auto-save:", autoSaveCheck)

	// Auto-save interval
	autoSaveIntervalSpinner = widgets.NewQSpinBox(nil)
	autoSaveIntervalSpinner.SetRange(10, 300)
	autoSaveIntervalSpinner.SetValue(preferences.AutoSaveInterval)
	autoSaveIntervalSpinner.SetSuffix(" seconds")
	layout.AddRow3("Auto-save Interval:", autoSaveIntervalSpinner)

	// Recent files section
	recentFilesGroup := widgets.NewQGroupBox2("Recent Files", nil)
	recentFilesLayout := widgets.NewQVBoxLayout()
	recentFilesGroup.SetLayout(recentFilesLayout)

	recentFilesList := widgets.NewQListWidget(nil)
	for _, file := range preferences.RecentFiles {
		recentFilesList.AddItem(file)
	}

	clearRecentButton := widgets.NewQPushButton2("Clear Recent Files", nil)
	clearRecentButton.ConnectClicked(func(bool) {
		recentFilesList.Clear()
		preferences.RecentFiles = []string{}
		SavePreferences()
	})

	recentFilesLayout.AddWidget(recentFilesList, 0, 0)
	recentFilesLayout.AddWidget(clearRecentButton, 0, 0)

	layout.AddRow3("", recentFilesGroup)

	return tab
}

func savePreferencesFromUI() {
	// Save editor settings
	SetEditorSettings(
		fontFamilyCombo.CurrentText(),
		fontSizeSpinner.Value(),
		tabWidthSpinner.Value(),
		lineNumbersCheck.IsChecked(),
		wrapTextCheck.IsChecked(),
	)

	// Save theme settings
	SetTheme(
		darkModeCheck.IsChecked(),
		themeCombo.CurrentText(),
	)

	// Save auto-save settings
	SetAutoSave(
		autoSaveCheck.IsChecked(),
		autoSaveIntervalSpinner.Value(),
	)

	// Apply settings to current editor session
	applyPreferencesToEditor()
}

func applyPreferencesToEditor() {
	// Apply font settings
	font := gui.NewQFont()
	font.SetFamily(preferences.EditorSettings.FontFamily)
	font.SetPointSize(preferences.EditorSettings.FontSize)
	font.SetFixedPitch(true)
	editor.SetFont(font)

	// Set tab width
	metrics := gui.NewQFontMetrics(font)
	editor.SetTabStopWidth(preferences.EditorSettings.TabWidth * metrics.HorizontalAdvance(" ", 0))

	// Apply text wrapping
	if preferences.EditorSettings.WrapText {
		editor.SetLineWrapMode(widgets.QPlainTextEdit__WidgetWidth)
	} else {
		editor.SetLineWrapMode(widgets.QPlainTextEdit__NoWrap)
	}

	// Apply theme if dark mode is enabled
	if preferences.ThemeSettings.DarkMode {
		// Create dark palette
		darkPalette := gui.NewQPalette()
		darkColor := gui.NewQColor3(45, 45, 45, 255)
		darkPalette.SetColor2(gui.QPalette__Window, darkColor)
		darkPalette.SetColor2(gui.QPalette__WindowText, gui.NewQColor3(255, 255, 255, 255))
		darkPalette.SetColor2(gui.QPalette__Base, gui.NewQColor3(25, 25, 25, 255))
		darkPalette.SetColor2(gui.QPalette__AlternateBase, darkColor)
		darkPalette.SetColor2(gui.QPalette__ToolTipBase, gui.NewQColor3(255, 255, 220, 255))
		darkPalette.SetColor2(gui.QPalette__ToolTipText, gui.NewQColor3(0, 0, 0, 255))
		darkPalette.SetColor2(gui.QPalette__Text, gui.NewQColor3(255, 255, 255, 255))
		darkPalette.SetColor2(gui.QPalette__Button, darkColor)
		darkPalette.SetColor2(gui.QPalette__ButtonText, gui.NewQColor3(255, 255, 255, 255))
		darkPalette.SetColor2(gui.QPalette__Link, gui.NewQColor3(42, 130, 218, 255))
		darkPalette.SetColor2(gui.QPalette__Highlight, gui.NewQColor3(42, 130, 218, 255))
		darkPalette.SetColor2(gui.QPalette__HighlightedText, gui.NewQColor3(0, 0, 0, 255))

		app.SetPalette(darkPalette, "")
	} else {
		// Reset to default palette
		app.SetPalette(app.Style().StandardPalette(), "")
	}

	// Force update of line number area
	editor.updateLineNumberAreaWidth()
	editor.lineNumberArea.Update()
}

func initializeFromPreferences() {
	// Initialize preferences system
	err := InitPreferences()
	if err != nil {
		fmt.Printf("Failed to initialize preferences: %v\n", err)
		return
	}

	// Apply settings to editor
	applyPreferencesToEditor()

	// Restore window geometry
	mainWindow.Resize2(preferences.WindowSettings.Width, preferences.WindowSettings.Height)
	mainWindow.Move2(preferences.WindowSettings.X, preferences.WindowSettings.Y)

	// Open last project if available
	if preferences.LastOpenedProject != "" {
		currentProjectPath = preferences.LastOpenedProject
		fileSystemModel.SetRootPath(currentProjectPath)
		fileTree.SetRootIndex(fileSystemModel.Index2(currentProjectPath, 0))
		fileTree.Expand(fileSystemModel.Index2(currentProjectPath, 0))

		// Open most recent file if available
		if len(preferences.RecentFiles) > 0 {
			openFile(preferences.RecentFiles[0])
		}
	}

	// Setup auto-save timer if enabled
	if preferences.AutoSaveEnabled && preferences.AutoSaveInterval > 0 {
		setupAutoSaveTimer()
	}
}

var autoSaveTimer *core.QTimer

func setupAutoSaveTimer() {
	if autoSaveTimer != nil {
		autoSaveTimer.Stop()
	}

	autoSaveTimer = core.NewQTimer(nil)
	autoSaveTimer.ConnectTimeout(func() {
		if currentFilePath != "" {
			saveCurrentFile()
		}
	})

	// Convert seconds to milliseconds
	autoSaveTimer.Start(preferences.AutoSaveInterval * 1000)
}

func saveWindowState() {
	// Save current window position and size
	UpdateWindowSettings(
		mainWindow.Width(),
		mainWindow.Height(),
		mainWindow.X(),
		mainWindow.Y(),
	)
}
