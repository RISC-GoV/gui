package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	rcore "github.com/RISC-GoV/core"
	assembler "github.com/RISC-GoV/risc-assembler"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Global variables
var (
	debugFileSplit []string
	realFileSplit  []string
	// File related variables
	currentFilePath    string
	currentProjectPath string
	app                *widgets.QApplication

	// Main UI components
	mainWindow      *widgets.QMainWindow
	editor          *CodeEditor
	terminalOutput  *widgets.QTextEdit // For stdout (read-only)
	terminalInput   *widgets.QTextEdit // For stdin (input)
	fileTree        *widgets.QTreeView
	fileSystemModel *widgets.QFileSystemModel

	// Debug related UI elements and state
	registersView   *widgets.QTableWidget
	memoryView      *widgets.QTableWidget
	debugToolbar    *widgets.QToolBar
	debugInfo       *DebugState
	debugContainer  *widgets.QSplitter
	currentHighline int // Current highlighted line in debugging
)

type DebugState struct {
	isDebugging bool
	cpu         *rcore.CPU
	breakpoints map[int]bool
}

type CodeEditor struct {
	*widgets.QPlainTextEdit
	lineNumberArea *LineNumberArea
}

// LineNumberArea shows line numbers and breakpoints
type LineNumberArea struct {
	*widgets.QWidget
	codeEditor *CodeEditor
}

func NewCodeEditor() *CodeEditor {
	editor := &CodeEditor{
		QPlainTextEdit: widgets.NewQPlainTextEdit(nil),
	}

	syntaxHighlighter = gui.NewQSyntaxHighlighter2(editor.Document())
	// Set up editor appearance
	font := gui.NewQFont()
	font.SetFamily(preferences.EditorSettings.FontFamily)
	font.SetFixedPitch(true)
	font.SetPointSize(preferences.EditorSettings.FontSize)
	editor.SetFont(font)

	// Set tab width
	metrics := gui.NewQFontMetrics(font)
	editor.SetTabStopWidth(4 * metrics.HorizontalAdvance(" ", 0))

	// Create line number area
	editor.lineNumberArea = NewLineNumberArea(editor)

	// Connect signals
	editor.ConnectUpdateRequest(editor.updateLineNumberArea)
	editor.lineNumberArea.ConnectMousePressEvent(editor.lineNumberAreaMousePress)

	// Update line number area width
	editor.updateLineNumberAreaWidth()
	editor.ConnectBlockCountChanged(func(int) { editor.updateLineNumberAreaWidth() })

	// Configure editor
	editor.SetLineWrapMode(widgets.QPlainTextEdit__NoWrap)

	return editor
}

func NewLineNumberArea(editor *CodeEditor) *LineNumberArea {
	lineNumberArea := &LineNumberArea{
		QWidget:    widgets.NewQWidget(editor, 0),
		codeEditor: editor,
	}

	return lineNumberArea
}

func (e *CodeEditor) updateLineNumberArea(rect *core.QRect, dy int) {
	if dy != 0 {
		e.lineNumberArea.Scroll(0, dy)
	} else {
		e.lineNumberArea.Update2(0, rect.Y(), e.lineNumberArea.Width(), rect.Height())
	}

	if rect.Contains(e.Viewport().Rect().TopLeft(), true) {
		e.updateLineNumberAreaWidth()
	}
}

func (e *CodeEditor) updateLineNumberAreaWidth() {
	// Calculate width needed for 5-digit line numbers plus breakpoint indicator
	// 5 digits @ ~8px each + breakpoint circle (16px) + padding (10px) = ~66px
	lineNumberWidth := 70
	e.SetViewportMargins(lineNumberWidth, 0, 0, 0)

	// Get the editor window size and set line number area to same height
	editorRect := e.Rect()
	e.lineNumberArea.SetGeometry2(0, 0, lineNumberWidth, editorRect.Height())
}

func (e *CodeEditor) BlockAtPosition(y int) int {
	block := e.FirstVisibleBlock()
	if !block.IsValid() {
		return 1
	}

	blockNumber := block.BlockNumber() + 1
	offset := e.ContentOffset()
	top := int(e.BlockBoundingGeometry(block).Translated(offset.X(), offset.Y()).Top())
	bottom := top + int(e.BlockBoundingRect(block).Height())

	for block.IsValid() && top <= y {
		if y <= bottom {
			return blockNumber
		}

		block = block.Next()
		top = bottom
		if block.IsValid() {
			bottom = top + int(e.BlockBoundingRect(block).Height())
		}
		blockNumber++
	}

	return blockNumber
}

func createToolbars() {
	// Create main toolbar
	mainToolbar := widgets.NewQToolBar("Main Toolbar", mainWindow)
	mainToolbar.SetWindowTitle("Main")
	mainWindow.AddToolBar(core.Qt__TopToolBarArea, mainToolbar)

	// File operations
	newAction := mainToolbar.AddAction2(gui.NewQIcon(), "New File")
	newAction.ConnectTriggered(func(bool) { createNewFile() })

	openAction := mainToolbar.AddAction2(gui.NewQIcon(), "Open File")
	openAction.ConnectTriggered(func(bool) { openFileDialog() })

	saveAction := mainToolbar.AddAction2(gui.NewQIcon(), "Save")
	saveAction.ConnectTriggered(func(bool) { saveCurrentFile() })

	mainToolbar.AddSeparator()

	// Run operations
	assembleAction := mainToolbar.AddAction2(gui.NewQIcon(), "Assemble")
	assembleAction.ConnectTriggered(func(bool) { AssembleCode() })

	// Run operations
	runAction := mainToolbar.AddAction2(gui.NewQIcon(), "Run")
	runAction.ConnectTriggered(func(bool) { runCode() })

	debugAction := mainToolbar.AddAction2(gui.NewQIcon(), "Debug")
	debugAction.ConnectTriggered(func(bool) { debugCode() })

	// Create debug toolbar (initially hidden)
	debugToolbar = widgets.NewQToolBar("Debug", mainWindow)
	debugToolbar.SetWindowTitle("Debug")
	mainWindow.AddToolBar(core.Qt__TopToolBarArea, debugToolbar)
	debugToolbar.SetVisible(false)

	hotReloadAction := debugToolbar.AddAction2(gui.NewQIcon(), "HotReload")
	hotReloadAction.ConnectTriggered(func(bool) { hotReloadCode() })

	stepAction := debugToolbar.AddAction2(gui.NewQIcon(), "Step")
	stepAction.ConnectTriggered(func(bool) { stepDebugCode() })

	continueAction := debugToolbar.AddAction2(gui.NewQIcon(), "Continue")
	continueAction.ConnectTriggered(func(bool) { continueDebugCode() })

	stopAction := debugToolbar.AddAction2(gui.NewQIcon(), "Stop")
	stopAction.ConnectTriggered(func(bool) { stopDebugging() })
}

func createMainContent() *widgets.QWidget {
	// Create main content widget
	contentWidget := widgets.NewQWidget(nil, 0)
	contentLayout := widgets.NewQVBoxLayout()
	contentWidget.SetLayout(contentLayout)

	// Create main splitter
	mainSplitter := widgets.NewQSplitter2(core.Qt__Horizontal, nil)

	// Left side: File browser
	fileSystemModel = widgets.NewQFileSystemModel(nil)
	// Initialize with current directory
	currentDir := "."
	fileSystemModel.SetRootPath(currentDir)

	fileTree = widgets.NewQTreeView(nil)
	fileTree.SetModel(fileSystemModel)
	fileTree.SetHeaderHidden(true)
	fileTree.HideColumn(1) // Hide Size column
	fileTree.HideColumn(2) // Hide Type column
	fileTree.HideColumn(3) // Hide Date Modified column

	// Connect file tree selection change
	fileTree.ConnectClicked(func(index *core.QModelIndex) {
		path := fileSystemModel.FilePath(index)
		if !core.NewQFileInfo3(path).IsDir() {
			openFile(path)
		}
	})

	filePanel := widgets.NewQWidget(nil, 0)
	filePanelLayout := widgets.NewQVBoxLayout()
	filePanelLayout.AddWidget(widgets.NewQLabel2("Files", nil, 0), 0, 0)
	filePanelLayout.AddWidget(fileTree, 0, 0)
	filePanel.SetLayout(filePanelLayout)

	mainSplitter.AddWidget(filePanel)

	// Right side: Editor and terminal
	rightSplitter := widgets.NewQSplitter2(core.Qt__Vertical, nil)

	// Code editor
	editor = NewCodeEditor()
	editorPanel := widgets.NewQWidget(nil, 0)
	editorLayout := widgets.NewQVBoxLayout()
	editorLayout.AddWidget(editor, 0, 0)
	editorPanel.SetLayout(editorLayout)

	rightSplitter.AddWidget(editorPanel)

	// Terminal panel with output and input terminals
	terminalPanel := widgets.NewQWidget(nil, 0)
	terminalLayout := widgets.NewQVBoxLayout()

	// Output terminal (stdout) - read-only
	terminalOutput = widgets.NewQTextEdit(nil)
	terminalOutput.SetReadOnly(true)
	terminalOutput.SetFontFamily(preferences.EditorSettings.FontFamily)
	terminalOutput.SetFontPointSize(float64(preferences.EditorSettings.TFontSize))

	// Input terminal (stdin)
	terminalInput = widgets.NewQTextEdit(nil)
	terminalInput.SetReadOnly(false)
	terminalInput.SetFontFamily(preferences.EditorSettings.FontFamily)
	terminalInput.SetFontPointSize(float64(preferences.EditorSettings.TFontSize))
	terminalInput.SetMaximumHeight(60) // Limit height of input terminal

	// Add terminals to layout
	terminalLayout.AddWidget(widgets.NewQLabel2("Output (stdout)", nil, 0), 0, 0)
	terminalLayout.AddWidget(terminalOutput, 0, 0)
	terminalLayout.AddWidget(widgets.NewQLabel2("Input (stdin)", nil, 0), 0, 0)
	terminalLayout.AddWidget(terminalInput, 0, 0)

	terminalPanel.SetLayout(terminalLayout)

	rightSplitter.AddWidget(terminalPanel)

	// Set initial splitter sizes for right panel
	rightSplitter.SetSizes([]int{600, 200})

	mainSplitter.AddWidget(rightSplitter)

	// Set initial splitter sizes for main splitter
	mainSplitter.SetSizes([]int{250, 950})

	contentLayout.AddWidget(mainSplitter, 0, 0)

	return contentWidget
}

func runCode() {
	if currentFilePath == "" {
		widgets.QMessageBox_Information(mainWindow, "No File", "No file is currently open to run", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	saveCurrentFile()

	// Create hidden directory for assembled output
	outputDir := filepath.Join(filepath.Dir(currentFilePath), ".riscgov_ide/assembling")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		widgets.QMessageBox_Critical(mainWindow, "Error", fmt.Sprintf("Failed to create output directory: %v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Assemble code
	terminalOutput.SetPlainText("Assembling code...\n")

	asm := assembler.Assembler{}

	err := asm.Assemble(currentFilePath, outputDir)
	if err != nil {
		errMsg := fmt.Sprintf("Assembly failed: %v\n", err)
		terminalOutput.SetPlainText(errMsg)
		return
	}

	setTerminal("Assembly successful.\nRunning code...\n")
	go func() {
		// Execute code
		outputFile := filepath.Join(outputDir, "output.exe")
		cpu := rcore.NewCPU(rcore.NewMemory())
		rcore.Kernel.Init()
		err = cpu.ExecuteFile(outputFile)
		if err != nil {
			setTerminal(fmt.Sprintf("Execution failed: %v\n", err))
			return
		}
		//wait 100ms
		time.Sleep(100 * time.Millisecond)
		setTerminal("\nProgram execution finished.\n")
	}()
}

func setTerminal(newMSG string) {
	const maxLines = 50

	old := terminalOutput.ToPlainText()
	combined := old + newMSG

	// Split into lines
	lines := strings.Split(combined, "\n")

	// Keep only the last maxLines
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	// Join and set the updated text
	terminalOutput.SetPlainText(strings.Join(lines, "\n"))

	// Scroll to bottom using gui.QTextCursor
	cursor := terminalOutput.TextCursor()
	cursor.MovePosition(gui.QTextCursor__End, gui.QTextCursor__MoveAnchor, 1)
	terminalOutput.SetTextCursor(cursor)
}

func updateRegistersDisplay() {
	if debugInfo.cpu == nil {
		return
	}

	// Update PC label (create if needed)
	pcLabel := widgets.NewQLabel2(fmt.Sprintf("PC: 0x%0x", debugInfo.cpu.PC), nil, 0)
	pcLabel.SetAlignment(core.Qt__AlignLeft)
	pcLabel.SetFont(gui.NewQFont2("Courier New", 12, 1, false))

	// Clear and update registers view
	for i := 0; i < 32; i++ {
		regValue := debugInfo.cpu.Registers[i]
		// Set register value in hex
		hexItem := registersView.Item(i, 1)
		hexItem.SetText(fmt.Sprintf("0x%0x(%d)", regValue, int32(regValue)))
	}
}

func viewMemory(addrStr string) {
	if debugInfo.cpu == nil {
		return
	}

	// Parse address
	var startAddr uint32
	_, err := fmt.Sscanf(addrStr, "%x", &startAddr)
	if err != nil {
		widgets.QMessageBox_Warning(mainWindow, "Invalid Address",
			"Please enter a valid hexadecimal address", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Clear memory view table
	memoryView.SetRowCount(16)

	// Display memory contents
	for i := 0; i < 16; i++ {
		addr := startAddr + uint32(i)
		value, err := debugInfo.cpu.Memory.ReadByte(addr)

		// Create address item
		addrItem := widgets.NewQTableWidgetItem2(fmt.Sprintf("0x%0x", addr), 0)
		memoryView.SetItem(i, 0, addrItem)

		// Create hex value item
		var hexItem *widgets.QTableWidgetItem
		var asciiItem *widgets.QTableWidgetItem

		if err != nil {
			hexItem = widgets.NewQTableWidgetItem2("Error", 0)
			asciiItem = widgets.NewQTableWidgetItem2("-", 0)
		} else {
			hexItem = widgets.NewQTableWidgetItem2(fmt.Sprintf("0x%02x", value), 0)

			// ASCII representation
			char := '.'
			if value >= 32 && value <= 126 {
				char = rune(value)
			}
			asciiItem = widgets.NewQTableWidgetItem2(string(char), 0)
		}

		memoryView.SetItem(i, 1, hexItem)
		memoryView.SetItem(i, 2, asciiItem)
	}
}

func createMenus() {
	menuBar := mainWindow.MenuBar()

	// File menu
	fileMenu := menuBar.AddMenu2("&File")

	newAction := fileMenu.AddAction("&New File")
	newAction.SetShortcut(gui.NewQKeySequence2("Ctrl+N", gui.QKeySequence__NativeText))
	newAction.ConnectTriggered(func(bool) { createNewFile() })

	openAction := fileMenu.AddAction("&Open File...")
	openAction.SetShortcut(gui.NewQKeySequence2("Ctrl+O", gui.QKeySequence__NativeText))
	openAction.ConnectTriggered(func(bool) { openFileDialog() })

	openProjectAction := fileMenu.AddAction("Open &Project...")
	openProjectAction.ConnectTriggered(func(bool) { openProjectDialog() })

	fileMenu.AddSeparator()

	saveAction := fileMenu.AddAction("&Save")
	saveAction.SetShortcut(gui.NewQKeySequence2("Ctrl+S", gui.QKeySequence__NativeText))
	saveAction.ConnectTriggered(func(bool) { saveCurrentFile() })

	saveAsAction := fileMenu.AddAction("Save &As...")
	saveAsAction.ConnectTriggered(func(bool) { saveFileAs() })

	fileMenu.AddSeparator()

	// Add Preferences menu item
	preferencesAction := fileMenu.AddAction("Pre&ferences...")
	preferencesAction.ConnectTriggered(func(bool) { showPreferencesDialog() })

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction("E&xit")
	exitAction.SetShortcut(gui.NewQKeySequence2("Alt+F4", gui.QKeySequence__NativeText))
	exitAction.ConnectTriggered(func(bool) {
		saveWindowState() // Save window state before exiting
		app.Quit()
	})

	// Edit menu
	editMenu := menuBar.AddMenu2("&Edit")

	undoAction := editMenu.AddAction("&Undo")
	undoAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Z", gui.QKeySequence__NativeText))
	undoAction.ConnectTriggered(func(bool) {
		if editor != nil {
			editor.Undo()
		}
	})

	redoAction := editMenu.AddAction("&Redo")
	redoAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Y", gui.QKeySequence__NativeText))
	redoAction.ConnectTriggered(func(bool) {
		if editor != nil {
			editor.Redo()
		}
	})

	editMenu.AddSeparator()

	cutAction := editMenu.AddAction("Cu&t")
	cutAction.SetShortcut(gui.NewQKeySequence2("Ctrl+X", gui.QKeySequence__NativeText))
	cutAction.ConnectTriggered(func(bool) {
		if editor != nil {
			editor.Cut()
		}
	})

	copyAction := editMenu.AddAction("&Copy")
	copyAction.SetShortcut(gui.NewQKeySequence2("Ctrl+C", gui.QKeySequence__NativeText))
	copyAction.ConnectTriggered(func(bool) {
		if editor != nil {
			editor.Copy()
		}
	})

	pasteAction := editMenu.AddAction("&Paste")
	pasteAction.SetShortcut(gui.NewQKeySequence2("Ctrl+V", gui.QKeySequence__NativeText))
	pasteAction.ConnectTriggered(func(bool) {
		if editor != nil {
			editor.Paste()
		}
	})

	// Run menu
	runMenu := menuBar.AddMenu2("&Run")

	assembleAction := runMenu.AddAction("&Assemble")
	assembleAction.SetShortcut(gui.NewQKeySequence2("F5", gui.QKeySequence__NativeText))
	assembleAction.ConnectTriggered(func(bool) { AssembleCode() })

	runAction := runMenu.AddAction("&Run")
	runAction.SetShortcut(gui.NewQKeySequence2("F6", gui.QKeySequence__NativeText))
	runAction.ConnectTriggered(func(bool) { runCode() })

	debugAction := runMenu.AddAction("&Debug")
	debugAction.SetShortcut(gui.NewQKeySequence2("F7", gui.QKeySequence__NativeText))
	debugAction.ConnectTriggered(func(bool) { debugCode() })

	// Help menu
	helpMenu := menuBar.AddMenu2("&Help")

	aboutAction := helpMenu.AddAction("&About")
	aboutAction.ConnectTriggered(func(bool) {
		widgets.QMessageBox_About(mainWindow, "About RISC-GoV IDE",
			"RISC-GoV IDE\nA development environment for RISC-V assembly.")
	})
}

func main() {
	// Initialize Qt application
	app = widgets.NewQApplication(len(os.Args), os.Args)

	applyModernTheme()
	// Initialize global variables
	debugInfo = &DebugState{
		isDebugging: false,
		breakpoints: make(map[int]bool),
	}

	// Create main window
	mainWindow = widgets.NewQMainWindow(nil, 0)
	mainWindow.SetWindowTitle("RISC-GoV IDE")
	mainWindow.Resize2(1200, 800)

	// Create central widget and main layout
	centralWidget := widgets.NewQWidget(nil, 0)
	mainLayout := widgets.NewQVBoxLayout()
	centralWidget.SetLayout(mainLayout)

	// Create menus and toolbars
	createMenus()
	createToolbars()

	// Create main content
	mainContent := createMainContent()
	mainLayout.AddWidget(mainContent, 0, 0)

	// Set central widget
	mainWindow.SetCentralWidget(centralWidget)

	// Start with window maximized to ensure proper rendering of all components
	mainWindow.ShowMaximized()
	initializeFromPreferences()
	editor.lineNumberArea.ConnectPaintEvent(editor.lineNumberAreaPaint)
	initTerminalIO()
	// Connect close event to save window state
	mainWindow.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		saveWindowState()
		event.Accept()
	})

	editor.ConnectTextChanged(func() {
		// Reapply syntax highlighting when text changes
		setupSyntaxHighlighting()
	})
	// Show window and run application
	mainWindow.Show()
	app.Exec()

}

func initTerminalIO() {
	stdinR, stdinW, _ := os.Pipe()
	stdoutR, stdoutW, _ := os.Pipe()

	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Stderr = stdoutW

	// Hold original stdout to write back to terminal for debugging (optional)
	originalStdout := os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")

	var currentInput string
	updateCh := make(chan string, 100)

	// UI: Capture key events
	terminalInput.ConnectKeyPressEvent(func(event *gui.QKeyEvent) {
		key := event.Key()

		// Handle Enter key - send input to stdin
		if key == int(core.Qt__Key_Return) || key == int(core.Qt__Key_Enter) {
			if currentInput != "" {
				stdinW.Write([]byte(currentInput))
				updateCh <- currentInput + "\n"
				currentInput = ""
				terminalInput.Clear()
			}
			event.Accept()
			return
		}

		// Handle Backspace
		if key == int(core.Qt__Key_Backspace) && len(currentInput) > 0 {
			currentInput = currentInput[:len(currentInput)-1]
			text := terminalInput.ToPlainText()
			if len(text) > 0 {
				terminalInput.SetPlainText(text[:len(text)-1])
			}
			event.Accept()
			return
		}

		// Handle regular character input
		if event.Text() != "" {
			char := event.Text()
			currentInput += char
			terminalInput.InsertPlainText(char)
			event.Accept()
		}
	})

	// Timer-based UI update from stdout pipe
	timer := core.NewQTimer(nil)
	timer.ConnectTimeout(func() {
		for {
			select {
			case out := <-updateCh:
				terminalOutput.Append(out)
				cursor := terminalOutput.TextCursor()
				cursor.MovePosition(gui.QTextCursor__End, gui.QTextCursor__MoveAnchor, 1)
				terminalOutput.SetTextCursor(cursor)
				terminalOutput.EnsureCursorVisible()

			default:
				return
			}
		}
	})
	timer.Start(5)

	// Goroutine to read from redirected stdout/stderr
	go func() {
		reader := bufio.NewReader(stdoutR)
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err != nil {
				break
			}
			output := string(buffer[:n])
			// Mirror to original stdout
			fmt.Fprintln(originalStdout, output)
			// Send to GUI
			updateCh <- output
		}
	}()
}
