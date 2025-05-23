package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	rcore "github.com/RISC-GoV/core"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Static shared variables
var (
	// Application-wide static components
	app        *widgets.QApplication
	mainWindow *widgets.QMainWindow

	// UI components
	editor          *CodeEditor
	terminalOutput  *widgets.QTextEdit
	terminalInput   *widgets.QTextEdit
	fileTree        *widgets.QTreeView
	fileSystemModel *widgets.QFileSystemModel
	debugToolbar    *widgets.QToolBar
	debugContainer  *widgets.QSplitter

	// Debug components
	registersView   *widgets.QTableWidget
	memoryView      *widgets.QTableWidget
	debugInfo       *DebugState
	currentHighline int

	// File handling
	debugFileSplit     []string
	realFileSplit      []string
	currentFilePath    string
	currentProjectPath string
	wg                 sync.WaitGroup
)

type DebugState struct {
	sync.RWMutex
	isDebugging bool
	cpu         *rcore.CPU
	breakpoints map[int]bool
}

type CodeEditor struct {
	*widgets.QPlainTextEdit
	lineNumberArea *LineNumberArea
}

type LineNumberArea struct {
	*widgets.QWidget
	codeEditor *CodeEditor
}

func NewCodeEditor() *CodeEditor {
	editor := &CodeEditor{
		QPlainTextEdit: widgets.NewQPlainTextEdit(nil),
	}

	syntaxHighlighter = gui.NewQSyntaxHighlighter2(editor.Document())
	font := gui.NewQFont()
	font.SetFamily(preferences.EditorSettings.FontFamily)
	font.SetFixedPitch(true)
	font.SetPointSize(preferences.EditorSettings.FontSize)

	metrics := gui.NewQFontMetrics(font)
	editor.SetTabStopWidth(4 * metrics.HorizontalAdvance(" ", 0))
	editor.SetFont(font)

	editor.lineNumberArea = NewLineNumberArea(editor)
	editor.ConnectUpdateRequest(editor.updateLineNumberArea)
	editor.lineNumberArea.ConnectMousePressEvent(editor.lineNumberAreaMousePress)
	editor.ConnectBlockCountChanged(func(int) { editor.updateLineNumberAreaWidth() })
	editor.SetLineWrapMode(widgets.QPlainTextEdit__NoWrap)
	editor.updateLineNumberAreaWidth()

	return editor
}

func (e *CodeEditor) updateLineNumberArea(rect *core.QRect, dy int) {
	if dy != 0 {
		e.lineNumberArea.Scroll(0, dy)
		return
	}
	e.lineNumberArea.Update2(0, rect.Y(), e.lineNumberArea.Width(), rect.Height())
	if rect.Contains(e.Viewport().Rect().TopLeft(), true) {
		e.updateLineNumberAreaWidth()
	}
}

func (e *CodeEditor) updateLineNumberAreaWidth() {
	lineNumberWidth := 70
	e.SetViewportMargins(lineNumberWidth, 0, 0, 0)
	e.lineNumberArea.SetGeometry2(0, 0, lineNumberWidth, e.Rect().Height())
}

func createToolbars() {
	mainToolbar := widgets.NewQToolBar("Main Toolbar", mainWindow)
	mainToolbar.SetWindowTitle("Main")
	mainWindow.AddToolBar(core.Qt__TopToolBarArea, mainToolbar)

	actions := map[string]func(){
		"New File":  createNewFile,
		"Open File": openFileDialog,
		"Save":      saveCurrentFile,
		"Assemble":  AssembleCode,
		"Run":       runCode,
		"Debug":     debugCode,
	}

	for name, handler := range actions {
		action := mainToolbar.AddAction2(gui.NewQIcon(), name)
		action.ConnectTriggered(func(bool) { handler() }) // wrap it
	}

	debugToolbar = widgets.NewQToolBar("Debug", mainWindow)
	debugToolbar.SetWindowTitle("Debug")
	mainWindow.AddToolBar(core.Qt__TopToolBarArea, debugToolbar)
	debugToolbar.SetVisible(false)

	debugActions := map[string]func(){
		"HotReload": hotReloadCode,
		"Step":      stepDebugCode,
		"Continue":  continueDebugCode,
		"Stop":      stopDebugging,
	}

	for name, handler := range debugActions {
		action := debugToolbar.AddAction2(gui.NewQIcon(), name)
		action.ConnectTriggered(func(bool) { handler() }) // wrap it
	}

}

func main() {
	app = widgets.NewQApplication(len(os.Args), os.Args)

	wg.Add(2)
	go func() {
		defer wg.Done()
		applyModernTheme()
		debugInfo = &DebugState{
			breakpoints: make(map[int]bool),
		}
	}()

	mainWindow = widgets.NewQMainWindow(nil, 0)
	mainWindow.SetWindowTitle("RISC-GoV IDE")
	mainWindow.Resize2(1200, 800)

	centralWidget := widgets.NewQWidget(nil, 0)
	mainLayout := widgets.NewQVBoxLayout()
	centralWidget.SetLayout(mainLayout)

	createMenus()
	createToolbars()
	mainContent := createMainContent()
	mainLayout.AddWidget(mainContent, 0, 0)
	go func() {
		defer wg.Done()
		initializeFromPreferences()
	}()

	mainWindow.SetCentralWidget(centralWidget)

	wg.Wait()

	mainWindow.ShowMaximized()
	editor.lineNumberArea.ConnectPaintEvent(editor.lineNumberAreaPaint)

	go initTerminalIO()

	mainWindow.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		go saveWindowState()
		event.Accept()
	})
	setupSyntaxHighlighting()
	mainWindow.Show()
	app.Exec()
}

func NewLineNumberArea(editor *CodeEditor) *LineNumberArea {
	lineNumberArea := &LineNumberArea{
		QWidget:    widgets.NewQWidget(editor, 0),
		codeEditor: editor,
	}

	return lineNumberArea
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

func createMenus() {
	menuBar := mainWindow.MenuBar()

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

	preferencesAction := fileMenu.AddAction("Pre&ferences...")
	preferencesAction.ConnectTriggered(func(bool) { showPreferencesDialog() })

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction("E&xit")
	exitAction.SetShortcut(gui.NewQKeySequence2("Alt+F4", gui.QKeySequence__NativeText))
	exitAction.ConnectTriggered(func(bool) {
		saveWindowState()
		app.Quit()
	})

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

	helpMenu := menuBar.AddMenu2("&Help")

	aboutAction := helpMenu.AddAction("&About")
	aboutAction.ConnectTriggered(func(bool) {
		widgets.QMessageBox_About(mainWindow, "About RISC-GoV IDE",
			"RISC-GoV IDE\nA development environment for RISC-V assembly.")
	})
}
func initTerminalIO() {
	stdinR, stdinW, _ := os.Pipe()
	stdoutR, stdoutW, _ := os.Pipe()

	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Stderr = stdoutW

	// Hold original stdout to write back to terminal for debugging
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

	// Update PC label
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
