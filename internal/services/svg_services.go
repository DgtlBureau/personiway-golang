package services

import (
	"fmt"
	"html"
	"strings"

	"github.com/ledongthuc/pdf"
)

type SimpleSVG struct {
	width, height int
	elements      []string
}

func NewSimpleSVG(width, height int) *SimpleSVG {
	return &SimpleSVG{
		width:    width,
		height:   height,
		elements: make([]string, 0),
	}
}

func (s *SimpleSVG) AddText(x, y float64, text string, fontSize int, color string) {
	escapedText := html.EscapeString(text)
	textElement := fmt.Sprintf(
		`<text x="%.2f" y="%.2f" font-family="Arial, sans-serif" font-size="%d" fill="%s">%s</text>`,
		x, y, fontSize, color, escapedText,
	)
	s.elements = append(s.elements, textElement)
}

func (s *SimpleSVG) ToString() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">
<rect width="%d" height="%d" fill="white"/>
`,
		s.width, s.height, s.width, s.height, s.width, s.height,
	))

	for _, element := range s.elements {
		builder.WriteString(element)
		builder.WriteString("\n")
	}

	builder.WriteString("</svg>")
	return builder.String()
}

func addTextRowsToSVG(svg *SimpleSVG, textRows pdf.Rows) {
	y := 30.0
	lineHeight := 16.0

	for _, row := range textRows {
		x := 20.0
		lineText := ""

		for _, word := range row.Content {
			lineText += word.S + " "
		}

		lineText = strings.TrimSpace(lineText)
		if lineText != "" {
			svg.AddText(x, y, lineText, 12, "black")
			y += lineHeight

			if y > 800 {
				break
			}
		}
	}
}
