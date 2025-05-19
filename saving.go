package main

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func createNewFile() {
	if currentProjectPath == "" {
		widgets.QMessageBox_Information(mainWindow, "No Project",
			"Please open a project folder first", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// Create dialog for new file name
	dialog := widgets.NewQDialog(mainWindow, 0)
	dialog.SetWindowTitle("New File")
	dialogLayout := widgets.NewQVBoxLayout()

	// Create input field
	label := widgets.NewQLabel2("Enter filename:", nil, 0)
	entry := widgets.NewQLineEdit(nil)
	entry.SetPlaceholderText("filename.asm")

	// Create button box
	buttonBox := widgets.NewQDialogButtonBox2(core.Qt__Horizontal, nil)
	buttonBox.SetStandardButtons(widgets.QDialogButtonBox__Ok | widgets.QDialogButtonBox__Cancel)
	buttonBox.ConnectAccepted(func() { dialog.Accept() })
	buttonBox.ConnectRejected(func() { dialog.Reject() })

	// Add widgets to layout
	dialogLayout.AddWidget(label, 0, 0)
	dialogLayout.AddWidget(entry, 0, 0)
	dialogLayout.AddWidget(buttonBox, 0, 0)
	dialog.SetLayout(dialogLayout)

	// Show dialog and process result
	if dialog.Exec() == int(widgets.QDialog__Accepted) {
		filename := entry.Text()
		if filename == "" {
			return
		}

		// Add .asm extension if not present
		if !strings.HasSuffix(filename, ".asm") {
			filename += ".asm"
		}

		path := filepath.Join(currentProjectPath, filename)
		err := ioutil.WriteFile(path, []byte(""), 0644)
		if err != nil {
			widgets.QMessageBox_Critical(mainWindow, "Error",
				fmt.Sprintf("Failed to create file: %v", err),
				widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			return
		}

		// Refresh file tree and open the new file
		fileSystemModel.SetRootPath(fileSystemModel.RootPath())
		openFile(path)
	}
}

func saveCurrentFile() {
	if currentFilePath == "" {
		saveFileAs()
		return
	}

	content := editor.ToPlainText()
	err := os.WriteFile(currentFilePath, []byte(content), 0644)
	if err != nil {
		widgets.QMessageBox_Critical(mainWindow, "Error",
			fmt.Sprintf("Failed to save file: %v", err),
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	}
}

func saveFileAs() {
	filePath := widgets.QFileDialog_GetSaveFileName(mainWindow, "Save File As", currentProjectPath,
		"Assembly Files (*.asm);;All Files (*.*)", "", 0)

	if filePath != "" {
		// Add .asm extension if not present
		if !strings.HasSuffix(filePath, ".asm") && !strings.Contains(filePath, ".") {
			filePath += ".asm"
		}

		currentFilePath = filePath
		saveCurrentFile()
		mainWindow.SetWindowTitle(fmt.Sprintf("RISC-GoV IDE - %s", filepath.Base(filePath)))
	}
}
