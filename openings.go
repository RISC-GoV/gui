package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/therecipe/qt/widgets"
)

func openFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		widgets.QMessageBox_Critical(mainWindow, "Error",
			fmt.Sprintf("Failed to open file: %v", err),
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	currentFilePath = path
	editor.SetPlainText(string(data))

	// Force update line numbers when file is opened
	editor.updateLineNumberAreaWidth()
	editor.lineNumberArea.Update()

	mainWindow.SetWindowTitle(fmt.Sprintf("RISC-GoV IDE - %s", filepath.Base(path)))

	// Add to recent files list
	AddRecentFile(path)

	// Trigger global syntax highlighting immediately after opening the file
	// This will re-apply highlighting to the entire document.
	if syntaxHighlighter != nil {
		syntaxHighlighter.Rehighlight()
	}
}

func openProjectDialog() {
	projectDir := widgets.QFileDialog_GetExistingDirectory(mainWindow, "Open Project Directory",
		"", widgets.QFileDialog__ShowDirsOnly)

	if projectDir != "" {
		currentProjectPath = projectDir
		fileSystemModel.SetRootPath(currentProjectPath)
		fileTree.SetRootIndex(fileSystemModel.Index2(currentProjectPath, 0))

		// Expand the root directory
		fileTree.Expand(fileSystemModel.Index2(currentProjectPath, 0))

		// Save as last opened project
		SetLastOpenedProject(projectDir)
	}
}

func openFileDialog() {
	filePath := widgets.QFileDialog_GetOpenFileName(mainWindow, "Open File", currentProjectPath,
		"Assembly Files (*.asm);;All Files (*.*)", "", 0)

	if filePath != "" {
		openFile(filePath)
	}
}
