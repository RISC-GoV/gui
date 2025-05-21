package main

import (
	"regexp"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

// SyntaxHighlighter for the editor
var syntaxHighlighter *gui.QSyntaxHighlighter

func setupSyntaxHighlighting() {

	// Define RISC-V specific syntax highlighting patterns
	riscvRegisters := `\b(zero|ra|sp|gp|tp|t[0-6]|s[0-11]|a[0-7]|x\d+)\b`
	riscvInstructions := `\b(add|addi|sub|lui|auipc|jal|jalr|beq|bne|blt|bge|bltu|bgeu|lb|lh|lw|lbu|lhu|sb|sh|sw|and|andi|or|ori|xor|xori|sll|slli|srl|srli|sra|srai|slt|slti|sltu|sltiu|ecall|ebreak|fence|fence\.i|csrrw|csrrs|csrrc|csrrwi|csrrsi|csrrci|mul|mulh|mulhsu|mulhu|div|divu|rem|remu)\b`
	riscvDirectives := `\.(text|data|section|global|globl|byte|half|word|dword|zero|align|file|ident|size|type|string|ascii|asciiz)`
	riscvPseudoInstructions := `\b(nop|li|la|mv|not|neg|seqz|snez|sltz|sgtz|beqz|bnez|blez|bgez|bltz|bgtz|j|jr|ret|call|tail)\b`

	// Determine highlighting colors based on current theme
	var registerColor, instructionColor, directiveColor, pseudoColor, commentColor, stringColor, numberColor, labelColor *gui.QColor

	if currentTheme == ThemeDark {
		// Dark theme colors
		registerColor = gui.NewQColor3(209, 105, 105, 255)   // Red
		instructionColor = gui.NewQColor3(86, 156, 214, 255) // Blue
		directiveColor = gui.NewQColor3(197, 134, 192, 255)  // Purple
		pseudoColor = gui.NewQColor3(78, 201, 176, 255)      // Teal
		commentColor = gui.NewQColor3(106, 153, 85, 255)     // Green
		stringColor = gui.NewQColor3(206, 145, 120, 255)     // Brown
		numberColor = gui.NewQColor3(181, 206, 168, 255)     // Light green
		labelColor = gui.NewQColor3(220, 220, 170, 255)      // Light yellow
	} else {
		// Light theme colors
		registerColor = gui.NewQColor3(170, 43, 43, 255)   // Dark red
		instructionColor = gui.NewQColor3(0, 0, 255, 255)  // Blue
		directiveColor = gui.NewQColor3(163, 21, 163, 255) // Purple
		pseudoColor = gui.NewQColor3(0, 128, 128, 255)     // Teal
		commentColor = gui.NewQColor3(0, 128, 0, 255)      // Green
		stringColor = gui.NewQColor3(163, 21, 21, 255)     // Dark red
		numberColor = gui.NewQColor3(9, 136, 90, 255)      // Green
		labelColor = gui.NewQColor3(121, 94, 38, 255)      // Brown
	}

	// Connect the highlightBlock function
	syntaxHighlighter.ConnectHighlightBlock(func(text string) {
		// Registers (bold)
		registerFormat := gui.NewQTextCharFormat()
		registerFormat.SetForeground(gui.NewQBrush3(registerColor, core.Qt__SolidPattern))
		registerFormat.SetFontWeight(75) // Bold
		applyFormatToPattern(text, riscvRegisters, registerFormat, syntaxHighlighter)

		// Instructions
		instructionFormat := gui.NewQTextCharFormat()
		instructionFormat.SetForeground(gui.NewQBrush3(instructionColor, core.Qt__SolidPattern))
		applyFormatToPattern(text, riscvInstructions, instructionFormat, syntaxHighlighter)

		// Directives
		directiveFormat := gui.NewQTextCharFormat()
		directiveFormat.SetForeground(gui.NewQBrush3(directiveColor, core.Qt__SolidPattern))
		applyFormatToPattern(text, riscvDirectives, directiveFormat, syntaxHighlighter)

		// Pseudo instructions
		pseudoFormat := gui.NewQTextCharFormat()
		pseudoFormat.SetForeground(gui.NewQBrush3(pseudoColor, core.Qt__SolidPattern))
		applyFormatToPattern(text, riscvPseudoInstructions, pseudoFormat, syntaxHighlighter)

		// Comments
		commentFormat := gui.NewQTextCharFormat()
		commentFormat.SetForeground(gui.NewQBrush3(commentColor, core.Qt__SolidPattern))
		commentRegex := regexp.MustCompile(`#.*$`)
		matches := commentRegex.FindAllStringIndex(text, -1)
		for _, match := range matches {
			syntaxHighlighter.SetFormat(match[0], match[1]-match[0], commentFormat)
		}

		// String literals
		stringFormat := gui.NewQTextCharFormat()
		stringFormat.SetForeground(gui.NewQBrush3(stringColor, core.Qt__SolidPattern))
		stringRegex := regexp.MustCompile(`".*?"`)
		matches = stringRegex.FindAllStringIndex(text, -1)
		for _, match := range matches {
			syntaxHighlighter.SetFormat(match[0], match[1]-match[0], stringFormat)
		}

		// Character literals
		charRegex := regexp.MustCompile(`'.*?'`)
		matches = charRegex.FindAllStringIndex(text, -1)
		for _, match := range matches {
			syntaxHighlighter.SetFormat(match[0], match[1]-match[0], stringFormat)
		}

		// Numbers (hex, binary, decimal)
		numberFormat := gui.NewQTextCharFormat()
		numberFormat.SetForeground(gui.NewQBrush3(numberColor, core.Qt__SolidPattern))
		numberRegex := regexp.MustCompile(`\b(0x[0-9a-fA-F]+|0b[01]+|\d+)\b`)
		matches = numberRegex.FindAllStringIndex(text, -1)
		for _, match := range matches {
			syntaxHighlighter.SetFormat(match[0], match[1]-match[0], numberFormat)
		}

		// Labels
		labelFormat := gui.NewQTextCharFormat()
		labelFormat.SetForeground(gui.NewQBrush3(labelColor, core.Qt__SolidPattern))
		labelRegex := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*):\b`)
		matches = labelRegex.FindAllStringIndex(text, -1)
		for _, match := range matches {
			syntaxHighlighter.SetFormat(match[0], match[1]-match[0], labelFormat)
		}
	})
}

// Helper function to apply formatting to all matches of a pattern
func applyFormatToPattern(text, pattern string, format *gui.QTextCharFormat, highlighter *gui.QSyntaxHighlighter) {
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		highlighter.SetFormat(match[0], match[1]-match[0], format)
	}
}

func (e *CodeEditor) lineNumberAreaPaint(event *gui.QPaintEvent) {
	painter := gui.NewQPainter2(e.lineNumberArea)
	defer painter.End()

	painter.SetFont(e.Font())

	// Fill background - fill the entire visible area
	r := event.Rect()
	painter.FillRect5(r.X(), r.Y(), r.Width(), r.Height(), preferences.ThemeSettings.LineNumberAreaColor)

	// Draw line numbers and breakpoint indicators
	block := e.FirstVisibleBlock()
	if !block.IsValid() {
		return
	}

	blockGeom := e.BlockBoundingGeometry(block)
	offset := e.ContentOffset()
	translated := blockGeom.Translated(offset.X(), offset.Y())
	top := int(translated.Y())
	bottom := top + int(e.BlockBoundingRect(block).Height())

	width := e.lineNumberArea.Width()
	height := e.FontMetrics().Height()

	blockNumber := block.BlockNumber()

	for block.IsValid() && top <= event.Rect().Bottom() {
		if block.IsVisible() && bottom >= event.Rect().Top() {
			number := strconv.Itoa(blockNumber + 1)

			if debugInfo.breakpoints[blockNumber] {
				pen := gui.NewQPen()
				pen.SetColor(gui.NewQColor3(255, 0, 0, 255))
				painter.SetPen(pen)

				brush := gui.NewQBrush()
				brush.SetColor(gui.NewQColor3(255, 0, 0, 255))
				brush.SetStyle(core.Qt__SolidPattern)
				painter.SetBrush(brush)

				size := (height - 4) / 2
				x := 3 + size/2
				y := top + 2 + size/2
				painter.DrawEllipse3(x, y, size, size)
			}

			// Highlight current debug line
			if debugInfo.isDebugging && blockNumber == currentHighline {
				painter.FillRect5(0, top, width, height, gui.NewQColor3(255, 255, 0, 100))
			}

			// Draw line number (right-aligned, with more space from the right edge)
			pen := gui.NewQPen()
			pen.SetColor(gui.NewQColor3(120, 120, 120, 255))
			painter.SetPen(pen)
			painter.DrawText3(width-45, top+height-4, number) // Right-align with more space
		}

		block = block.Next()
		top = bottom
		bottom = top + int(e.BlockBoundingRect(block).Height())
		blockNumber++
	}
}
