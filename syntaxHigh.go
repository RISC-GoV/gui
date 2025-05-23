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
	// More comprehensive register pattern including all valid RISC-V registers
	reRiscvRegisters = regexp.MustCompile(`\b(?:x(?:[0-9]|[12][0-9]|3[01])|zero|ra|sp|gp|tp|t[0-6]|s[0-9]|s1[0-1]|a[0-7])\b`)

	// Extended instruction set including all RV32I, RV64I, M, A, and F extensions
	reRiscvInstructions = regexp.MustCompile(`\b(?:add|addi|sub|lui|auipc|jal|jalr|beq|bne|blt|bge|bltu|bgeu|lb|lh|lw|ld|lbu|lhu|lwu|sb|sh|sw|sd|and|andi|or|ori|xor|xori|sll|slli|srl|srli|sra|srai|slt|slti|sltu|sltiu|mul|mulh|mulhsu|mulhu|div|divu|rem|remu|fence|fence\.i|ecall|ebreak|csrr[wsrc]|csrr[ws]i|csrrc|amoadd\.[wd]|amoswap\.[wd]|amoand\.[wd]|amoor\.[wd]|amoxor\.[wd]|amomax[u]?\.[wd]|amomin[u]?\.[wd]|flw|fsw|fadd\.s|fsub\.s|fmul\.s|fdiv\.s|fsqrt\.s|fmin\.s|fmax\.s|fcvt\.[ws]\.s|fcvt\.s\.[ws]|fmv\.[wx]\.s|fmv\.s\.[wx]|feq\.s|flt\.s|fle\.s|fclass\.s)\b`)

	// Extended directives including common assembler directives
	reRiscvDirectives = regexp.MustCompile(`\.(?:text|data|section|global|globl|local|weak|byte|2byte|half|4byte|word|8byte|dword|quad|zero|align|balign|p2align|file|ident|size|type|string|ascii|asciiz|set|equ|option|attribute|include|macro|endm|rept|endr|extern|rodata|bss)`)

	// Extended pseudo-instructions
	reRiscvPseudoInstructions = regexp.MustCompile(`\b(?:nop|li|la|mv|not|neg|seqz|snez|sltz|sgtz|beqz|bnez|blez|bgez|bltz|bgtz|bgt|ble|bgtu|bleu|j|jr|ret|call|tail|sext\.w|zext\.w)\b`)

	// Improved comment pattern to handle both single-line and multi-line comments
	reComment = regexp.MustCompile(`#.*$|/\*(?s:.*?)\*/`)

	// Improved string and character patterns
	reString = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	reChar = regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)

	// Extended number pattern to handle all common number formats
	reNumber = regexp.MustCompile(`\b(?:0x[0-9a-fA-F]+|0b[01]+|0o[0-7]+|\d+)\b`)

	// Improved label pattern
	reLabel = regexp.MustCompile(`^[ \t]*[a-zA-Z_.][a-zA-Z0-9_$.]*:`)
}

func applyFormatToPattern(text string, compiledRegex *regexp.Regexp, format *gui.QTextCharFormat, highlighter *gui.QSyntaxHighlighter) {
	matches := compiledRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		highlighter.SetFormat(match[0], match[1]-match[0], format)
	}
}

func setupSyntaxHighlighting() {
	var (
		registerColor    *gui.QColor
		instructionColor *gui.QColor
		directiveColor   *gui.QColor
		pseudoColor      *gui.QColor
		commentColor     *gui.QColor
		stringColor      *gui.QColor
		numberColor      *gui.QColor
		labelColor       *gui.QColor
	)

	if currentTheme == ThemeDark {
		registerColor = gui.NewQColor3(255, 128, 128, 255)    // Brighter red
		instructionColor = gui.NewQColor3(130, 177, 255, 255) // Brighter blue
		directiveColor = gui.NewQColor3(216, 160, 223, 255)   // Brighter purple
		pseudoColor = gui.NewQColor3(100, 223, 223, 255)      // Brighter teal
		commentColor = gui.NewQColor3(128, 178, 128, 255)     // Brighter green
		stringColor = gui.NewQColor3(230, 192, 160, 255)      // Brighter brown
		numberColor = gui.NewQColor3(200, 230, 180, 255)      // Brighter light green
		labelColor = gui.NewQColor3(240, 240, 190, 255)       // Brighter yellow
	} else {
		// More vibrant light theme colors
		registerColor = gui.NewQColor3(204, 0, 0, 255)      // Vivid red
		instructionColor = gui.NewQColor3(0, 102, 204, 255) // Strong blue
		directiveColor = gui.NewQColor3(153, 0, 204, 255)   // Rich purple
		pseudoColor = gui.NewQColor3(0, 153, 153, 255)      // Deep teal
		commentColor = gui.NewQColor3(0, 153, 0, 255)       // Clear green
		stringColor = gui.NewQColor3(204, 102, 0, 255)      // Deep orange
		numberColor = gui.NewQColor3(0, 153, 102, 255)      // Forest green
		labelColor = gui.NewQColor3(153, 102, 0, 255)       // Rich brown
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
