package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	rcore "github.com/RISC-GoV/core"
	assembler "github.com/RISC-GoV/risc-assembler"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func stepDebugCode() {
	if !debugInfo.isDebugging || debugInfo.cpu == nil {
		return
	}

	// Execute the current instruction
	state := debugInfo.cpu.ExecuteSingle()
	updateRegistersDisplay()

	// Calculate the line to highlight in the editor
	// Here we add 1 to show the next line that will execute, not the one that just ran
	lineNum := 1 // Default to line 1
	if debugInfo.cpu.PC != 0 {
		// Calculate line number based on PC value
		// This assumes each instruction is 4 bytes and maps to source lines
		lineNum = int(debugInfo.cpu.PC / 4)
	}

	// Highlight the next execution line
	editor.HighlightLine(lineNum)

	switch state {
	case rcore.PROGRAM_EXIT:
		terminalOutput.SetText(terminalOutput.ToPlainText() + "Program exited normally\n")
		stopDebugging()
	case rcore.PROGRAM_EXIT_FAILURE:
		terminalOutput.SetText(terminalOutput.ToPlainText() + "Program exited with failure\n")
		stopDebugging()
	case rcore.E_BREAK:
		terminalOutput.SetText(terminalOutput.ToPlainText() + fmt.Sprintf("Breakpoint hit at 0x%0x\n", debugInfo.cpu.PC))
	default:
		terminalOutput.SetText(terminalOutput.ToPlainText() + fmt.Sprintf("PC (4byte/instructions) = %d\n", debugInfo.cpu.PC))
	}
}

func continueDebugCode() {
	if !debugInfo.isDebugging || debugInfo.cpu == nil {
		return
	}

	go func() {
		for debugInfo.isDebugging {
			state := debugInfo.cpu.ExecuteSingle()

			switch state {
			case rcore.PROGRAM_EXIT:
				terminalOutput.SetText(terminalOutput.ToPlainText() + "Program exited normally\n")
				stopDebugging()
				return
			case rcore.PROGRAM_EXIT_FAILURE:
				terminalOutput.SetText(terminalOutput.ToPlainText() + "Program exited with failure\n")
				stopDebugging()
				return
			case rcore.E_BREAK:
				terminalOutput.SetText(terminalOutput.ToPlainText() + fmt.Sprintf("Breakpoint hit at 0x%0x\n", debugInfo.cpu.PC))

				// Calculate the line to highlight in the editor - show the next line to execute
				lineNum := 1 // Default to line 1
				if debugInfo.cpu.PC != 0 {
					// Calculate line number based on PC value - point to next instruction
					lineNum = int(debugInfo.cpu.PC / 4)
				}

				// Update registers and highlight the current line
				updateRegistersDisplay()
				editor.HighlightLine(lineNum)
				return
			}
		}
		updateRegistersDisplay()
	}()
}

func debugCode() {
	if currentFilePath == "" {
		widgets.QMessageBox_Information(mainWindow, "No File", "No file is currently open to debug", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Stop any existing debug session first to ensure clean state
	if debugInfo.isDebugging {
		stopDebugging()
	}

	saveCurrentFile()

	// Create hidden directory for assembled output
	outputDir := filepath.Join(filepath.Dir(currentFilePath), ".riscgov_ide/assembling")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		widgets.QMessageBox_Critical(mainWindow, "Error", fmt.Sprintf("Failed to create output directory: %v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Process breakpoints - add ebreak instructions
	lines := strings.Split(editor.ToPlainText(), "\n")
	tempFile := filepath.Join(outputDir, "temp_"+filepath.Base(currentFilePath))

	var modifiedContent strings.Builder

	for lineIndex, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments for breakpoint purposes
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "//") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 0 {
				_, isInstruction := assembler.InstructionToOpType[parts[0]]
				_, isInstruction2 := assembler.PseudoToInstruction[parts[0]]

				// If this is an instruction and we have a breakpoint on this line
				if (isInstruction || isInstruction2) && debugInfo.breakpoints[lineIndex] {
					modifiedContent.WriteString("ebreak\n")
				}
			}
		}

		modifiedContent.WriteString(line + "\n")
	}
	terminalOutput.Clear()

	debugFileContent := modifiedContent.String()

	debugFileSplit = strings.Split(debugFileContent, "\n")
	realFileSplit = strings.Split(editor.ToPlainText(), "\n")
	if err := os.WriteFile(tempFile, []byte(modifiedContent.String()), 0644); err != nil {
		terminalOutput.SetPlainText("Failed to create temporary file with breakpoints.")
		return
	}

	// Assemble code
	terminalOutput.SetPlainText("Assembling code with breakpoints...\n")

	asm := assembler.Assembler{}
	err := asm.Assemble(tempFile, outputDir)
	if err != nil {
		errMsg := fmt.Sprintf("Assembly failed: %v\n", err)
		terminalOutput.SetPlainText(errMsg)
		return
	}

	setTerminal("Assembly successful.\nStarting debugger...\n")

	// Start debug session with fresh state
	debugInfo.isDebugging = true
	debugInfo.cpu = rcore.NewCPU(rcore.NewMemory())
	rcore.Kernel.Init()
	// Show debug UI
	showDebugWindows()

	outputFile := filepath.Join(outputDir, "output.exe")
	// Load program in CPU
	err = debugInfo.cpu.LoadFile(outputFile)
	if err != nil {
		setTerminal(fmt.Sprintf("Debug failed: %v\n", err))
		stopDebugging()
		return
	}

	// Update registers display
	updateRegistersDisplay()

	lineNum := 1 // Default to line 1
	if debugInfo.cpu.PC != 0 {
		// Calculate line number based on PC value - point to next instruction
		lineNum = int(debugInfo.cpu.PC/4) + 1
	}
	editor.HighlightLine(lineNum)
	setTerminal("Debug session started. Use Step or Continue.\n")
}

func hotReloadCode() {
	saveCurrentFile()

	// Create hidden directory for assembled output
	outputDir := filepath.Join(filepath.Dir(currentFilePath), ".riscgov_ide/assembling")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		widgets.QMessageBox_Critical(mainWindow, "Error", fmt.Sprintf("Failed to create output directory: %v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Process breakpoints - add ebreak instructions
	lines := strings.Split(editor.ToPlainText(), "\n")
	tempFile := filepath.Join(outputDir, "temp_"+filepath.Base(currentFilePath))

	var modifiedContent strings.Builder

	for lineIndex, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments for breakpoint purposes
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "//") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 0 {
				_, isInstruction := assembler.InstructionToOpType[parts[0]]
				_, isInstruction2 := assembler.PseudoToInstruction[parts[0]]

				// If this is an instruction and we have a breakpoint on this line
				if (isInstruction || isInstruction2) && debugInfo.breakpoints[lineIndex] {
					modifiedContent.WriteString("ebreak\n")
				}
			}
		}

		modifiedContent.WriteString(line + "\n")
	}
	terminalOutput.Clear()

	debugFileContent := modifiedContent.String()

	debugFileSplit = strings.Split(debugFileContent, "\n")
	realFileSplit = strings.Split(editor.ToPlainText(), "\n")
	if err := os.WriteFile(tempFile, []byte(modifiedContent.String()), 0644); err != nil {
		terminalOutput.SetPlainText("Failed to create temporary file with breakpoints.")
		return
	}

	asm := assembler.Assembler{}
	err := asm.Assemble(tempFile, outputDir)
	if err != nil {
		widgets.QMessageBox_Critical(mainWindow, "Error", fmt.Sprintf("Hot reload failed, error Assembling:\n %v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Show debug UI
	showDebugWindows()

	outputFile := filepath.Join(outputDir, "output.exe")
	debugInfo.cpu.Memory = rcore.NewMemory()
	oldPC := debugInfo.cpu.PC
	// Load program in CPU
	err = debugInfo.cpu.LoadFile(outputFile)
	if err != nil {
		widgets.QMessageBox_Critical(mainWindow, "Error", fmt.Sprintf("Hot reload failed, error LoadingFile:\n %v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}
	debugInfo.cpu.PC = oldPC
}

func stopDebugging() {
	if !debugInfo.isDebugging {
		return
	}
	debugInfo.isDebugging = false
	debugInfo.cpu = nil

	// Restore normal UI
	hideDebugWindows()
	debugFileSplit = nil
	realFileSplit = nil

	if editor != nil && editor.lineNumberArea != nil {
		editor.lineNumberArea.Update()
	}

	terminalOutput.SetText(terminalOutput.ToPlainText() + "Debug session stopped.\n")
}

func showDebugWindows() {
	// Make debug toolbar visible
	debugToolbar.SetVisible(true)

	// Create debug panels if they don't exist
	if debugContainer == nil {
		// Create registers view
		registersView = widgets.NewQTableWidget(nil)
		registersView.SetColumnCount(2)
		registersView.SetRowCount(32) // 32 RISC-V registers
		registersView.SetHorizontalHeaderLabels([]string{"Register(ABI)", "Hex(Dec)"})
		registersView.VerticalHeader().SetVisible(false)
		registersView.SetEditTriggers(widgets.QAbstractItemView__NoEditTriggers)
		header := registersView.HorizontalHeader()
		header.SetDefaultAlignment(core.Qt__AlignLeft)

		// Initialize register rows
		regNames := []string{
			"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
			"s0/fp", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
			"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
			"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
		}

		for i := 0; i < 32; i++ {
			registerItem := widgets.NewQTableWidgetItem2(fmt.Sprintf("x%d(%s)", i, regNames[i]), 0)
			ValueItem := widgets.NewQTableWidgetItem2("0x0(0)", 0)

			registersView.SetItem(i, 0, registerItem)
			registersView.SetItem(i, 1, ValueItem)
		}

		// Create memory view with address input controls
		memoryView = widgets.NewQTableWidget(nil)
		memoryView.SetColumnCount(3)
		memoryView.SetRowCount(16) // Display 16 memory locations by default
		memoryView.SetHorizontalHeaderLabels([]string{"Address", "Hex", "ASCII"})
		memoryView.VerticalHeader().SetVisible(false)
		memoryView.SetEditTriggers(widgets.QAbstractItemView__NoEditTriggers)

		// Memory view controls
		addressLabel := widgets.NewQLabel2("Address:", nil, 0)
		addressInput := widgets.NewQLineEdit(nil)
		addressInput.SetPlaceholderText("0x0")

		viewButton := widgets.NewQPushButton2("View Memory", nil)
		viewButton.ConnectClicked(func(bool) {
			addr := addressInput.Text()
			viewMemory(addr)
		})

		memoryControls := widgets.NewQWidget(nil, 0)
		memoryControlsLayout := widgets.NewQHBoxLayout()
		memoryControlsLayout.AddWidget(addressLabel, 0, 0)
		memoryControlsLayout.AddWidget(addressInput, 0, 0)
		memoryControlsLayout.AddWidget(viewButton, 0, 0)
		memoryControls.SetLayout(memoryControlsLayout)

		// Create memory panel with controls
		memoryPanel := widgets.NewQWidget(nil, 0)
		memoryLayout := widgets.NewQVBoxLayout()
		memoryLayout.AddWidget(widgets.NewQLabel2("Memory", nil, 0), 0, 0)
		memoryLayout.AddWidget(memoryControls, 0, 0)
		memoryLayout.AddWidget(memoryView, 0, 0)
		memoryPanel.SetLayout(memoryLayout)

		// Create registers panel
		registersPanel := widgets.NewQWidget(nil, 0)
		registersLayout := widgets.NewQVBoxLayout()
		registersLayout.AddWidget(widgets.NewQLabel2("Registers", nil, 0), 0, 0)
		registersLayout.AddWidget(registersView, 0, 0)
		registersPanel.SetLayout(registersLayout)

		// Create debug panel container
		debugPanel := widgets.NewQSplitter2(core.Qt__Vertical, nil)
		debugPanel.AddWidget(registersPanel)
		debugPanel.AddWidget(memoryPanel)
		debugPanel.SetSizes([]int{400, 400})

		// Replace editor with a splitter containing editor and debug panel
		editorParent := editor.ParentWidget()
		editorLayout := editorParent.Layout()
		// Remove editor from its parent
		editorLayout.RemoveWidget(editor)

		// Create new container for editor and debug view
		debugContainer = widgets.NewQSplitter2(core.Qt__Horizontal, nil)
		debugContainer.AddWidget(editor)
		debugContainer.AddWidget(debugPanel)
		debugContainer.SetSizes([]int{700, 500})

		// Add splitter to layout
		editorLayout.AddWidget(debugContainer)
	} else {
		// If debug container already exists, just make it visible
		debugContainer.Widget(1).SetVisible(true)
	}
}

func hideDebugWindows() {
	// Hide debug toolbar
	debugToolbar.SetVisible(false)

	// Hide debug panels if they exist
	if debugContainer != nil {
		debugContainer.Widget(1).SetVisible(false)
	}

	// Clear highlight
	currentHighline = -1
	if editor != nil && editor.lineNumberArea != nil {
		editor.lineNumberArea.Update()
	}
}

func (e *CodeEditor) lineNumberAreaMousePress(event *gui.QMouseEvent) {
	// Get line number at cursor position
	blockNumber := e.BlockAtPosition(event.Y())

	// Calculate the actual source code line number (1-based)
	lineNumber := blockNumber - 1

	// Toggle breakpoint
	if debugInfo.breakpoints[lineNumber] {
		delete(debugInfo.breakpoints, lineNumber)
	} else {
		debugInfo.breakpoints[lineNumber] = true
	}

	// Update the line number area
	e.lineNumberArea.Update()
}

func getRelevantLine(lineNum int, lines []string) (int, int) {
	instructionCount := 0
	nBreaks := 0
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
			continue // Skip empty lines and comments
		}

		// Check if this line contains an instruction
		parts := strings.Fields(trimmedLine)
		if len(parts) > 0 {
			_, isInstruction := assembler.InstructionToOpType[parts[0]]
			retFunc, isInstruction2 := assembler.PseudoToInstruction[parts[0]]
			if isInstruction || isInstruction2 {

				instructionCount++

				if instructionCount == lineNum {
					return i, nBreaks
				}
				if isInstruction2 {
					result := retFunc(parts)
					instructionCount += len(result) - 1
				}
				if parts[0] == "ebreak" {
					nBreaks++
				}
			}
		}
	}
	return -1, nBreaks
}

func (e *CodeEditor) HighlightLine(lineNum int) {
	// Calculate the actual line number in the source code
	currentHighline = -1
	realFileVal, realBreaks := getRelevantLine(lineNum, realFileSplit)
	_, debugBreaks := getRelevantLine(lineNum, debugFileSplit)
	currentHighline = realFileVal + (realBreaks - debugBreaks)
	if currentHighline < 1 {
		return // Line doesn't exist or isn't an instruction
	}

	// Scroll to make sure the line is visible
	block := e.Document().FindBlockByLineNumber(currentHighline)
	cursor := e.TextCursor()
	cursor.SetPosition(block.Position(), gui.QTextCursor__MoveAnchor)
	e.SetTextCursor(cursor)
	e.CenterCursor()

	// Redraw line number area to show highlight
	e.lineNumberArea.Update()
}
