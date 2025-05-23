package main

import (
	"regexp"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

var (
	reRiscvRegisters          *regexp.Regexp
	reRiscvInstructions       *regexp.Regexp
	reRiscvDirectives         *regexp.Regexp
	reRiscvPseudoInstructions *regexp.Regexp
	reComment                 *regexp.Regexp
	reString                  *regexp.Regexp
	reChar                    *regexp.Regexp
	reNumber                  *regexp.Regexp
	reLabel                   *regexp.Regexp

	registerFormat    *gui.QTextCharFormat
	instructionFormat *gui.QTextCharFormat
	directiveFormat   *gui.QTextCharFormat
	pseudoFormat      *gui.QTextCharFormat
	commentFormat     *gui.QTextCharFormat
	stringFormat      *gui.QTextCharFormat
	numberFormat      *gui.QTextCharFormat
	labelFormat       *gui.QTextCharFormat
)

// SyntaxHighlighter for the editor
var syntaxHighlighter *gui.QSyntaxHighlighter

func init() {
	// Compile all regexps once when the package is initialized
	reRiscvRegisters = regexp.MustCompile(`\b(zero|ra|sp|gp|tp|t[0-6]|s[0-11]|a[0-7]|x\d+)\b`)
	reRiscvInstructions = regexp.MustCompile(`\b(add|addi|sub|lui|auipc|jal|jalr|beq|bne|blt|bge|bltu|bgeu|lb|lh|lw|lbu|lhu|sb|sh|sw|and|andi|or|ori|xor|xori|sll|slli|srl|srli|sra|srai|slt|slti|sltu|sltiu|ecall|ebreak|fence|fence\.i|csrrw|csrrs|csrrc|csrrwi|csrrsi|csrrci|mul|mulh|mulhsu|mulhu|div|divu|rem|remu)\b`)
	reRiscvDirectives = regexp.MustCompile(`\.(text|data|section|global|globl|byte|half|word|dword|zero|align|file|ident|size|type|string|ascii|asciiz)`)
	reRiscvPseudoInstructions = regexp.MustCompile(`\b(nop|li|la|mv|not|neg|seqz|snez|sltz|sgtz|beqz|bnez|blez|bgez|bltz|bgtz|j|jr|ret|call|tail)\b`)
	reComment = regexp.MustCompile(`#.*$`)
	reString = regexp.MustCompile(`".*?"`)
	reChar = regexp.MustCompile(`'.*?'`)
	reNumber = regexp.MustCompile(`\b(0x[0-9a-fA-F]+|0b[01]+|\d+)\b`)
	reLabel = regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*):\b`)
}

func applyFormatToPattern(text string, compiledRegex *regexp.Regexp, format *gui.QTextCharFormat, highlighter *gui.QSyntaxHighlighter) {
	matches := compiledRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		highlighter.SetFormat(match[0], match[1]-match[0], format)
	}
}

func setupSyntaxHighlighting() {
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

	// Initialize formats once, update only if colors change
	registerFormat = gui.NewQTextCharFormat()
	registerFormat.SetForeground(gui.NewQBrush3(registerColor, core.Qt__SolidPattern))
	registerFormat.SetFontWeight(75) // Bold

	instructionFormat = gui.NewQTextCharFormat()
	instructionFormat.SetForeground(gui.NewQBrush3(instructionColor, core.Qt__SolidPattern))

	directiveFormat = gui.NewQTextCharFormat()
	directiveFormat.SetForeground(gui.NewQBrush3(directiveColor, core.Qt__SolidPattern))

	pseudoFormat = gui.NewQTextCharFormat()
	pseudoFormat.SetForeground(gui.NewQBrush3(pseudoColor, core.Qt__SolidPattern))

	commentFormat = gui.NewQTextCharFormat()
	commentFormat.SetForeground(gui.NewQBrush3(commentColor, core.Qt__SolidPattern))

	stringFormat = gui.NewQTextCharFormat()
	stringFormat.SetForeground(gui.NewQBrush3(stringColor, core.Qt__SolidPattern))

	numberFormat = gui.NewQTextCharFormat()
	numberFormat.SetForeground(gui.NewQBrush3(numberColor, core.Qt__SolidPattern))

	labelFormat = gui.NewQTextCharFormat()
	labelFormat.SetForeground(gui.NewQBrush3(labelColor, core.Qt__SolidPattern))

	// Connect the highlightBlock function
	syntaxHighlighter.ConnectHighlightBlock(func(text string) {
		// Apply formats using the pre-compiled regexps and pre-initialized formats
		applyFormatToPattern(text, reRiscvRegisters, registerFormat, syntaxHighlighter)
		applyFormatToPattern(text, reRiscvInstructions, instructionFormat, syntaxHighlighter)
		applyFormatToPattern(text, reRiscvDirectives, directiveFormat, syntaxHighlighter)
		applyFormatToPattern(text, reRiscvPseudoInstructions, pseudoFormat, syntaxHighlighter)
		applyFormatToPattern(text, reComment, commentFormat, syntaxHighlighter)
		applyFormatToPattern(text, reString, stringFormat, syntaxHighlighter)
		applyFormatToPattern(text, reChar, stringFormat, syntaxHighlighter)
		applyFormatToPattern(text, reNumber, numberFormat, syntaxHighlighter)
		applyFormatToPattern(text, reLabel, labelFormat, syntaxHighlighter)
	})

	syntaxHighlighter.Rehighlight()
}
func (e *CodeEditor) lineNumberAreaPaint(event *gui.QPaintEvent) {
	painter := gui.NewQPainter2(e.lineNumberArea)
	defer painter.End()

	painter.SetFont(e.Font())

	// Pre-create common pens and brushes
	breakpointPen := gui.NewQPen()
	breakpointPen.SetColor(gui.NewQColor3(255, 0, 0, 255))

	breakpointBrush := gui.NewQBrush()
	breakpointBrush.SetColor(gui.NewQColor3(255, 0, 0, 255))
	breakpointBrush.SetStyle(core.Qt__SolidPattern)

	lineNumberPen := gui.NewQPen()
	lineNumberPen.SetColor(gui.NewQColor3(120, 120, 120, 255))

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
				painter.SetPen(breakpointPen)
				painter.SetBrush(breakpointBrush)

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
			painter.SetPen(lineNumberPen)                     // Use the pre-created pen
			painter.DrawText3(width-45, top+height-4, number) // Right-align with more space
		}

		block = block.Next()
		top = bottom
		bottom = top + int(e.BlockBoundingRect(block).Height())
		blockNumber++
	}
}
