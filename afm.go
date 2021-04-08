package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// fontMetrics is a map of known fonts to font metrics.
var fontMetrics = map[string]*FontMetrics{}

// FontMetrics is a structure containing AFM font metrics.
type FontMetrics struct {
	RuneWidth  map[rune]int
	NamedWidth map[string]int
}

// Width calculates the width of a string in the given font at a given pt.
func (m *FontMetrics) Width(pt float64, text string) float64 {
	var w float64
	for _, r := range text {
		w += float64(m.RuneWidth[r]) * pt / 1000.0
	}
	return w
}

// This initializer loads AFM files from afm/*.afm.
func init() {
	filenames, err := filepath.Glob("afm/*.afm")
	if err != nil {
		log.Fatalf("Searching for afm files failed: %v", err)
	}

	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Error opening afm file %q: %v", filename, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		metrics := &FontMetrics{RuneWidth: map[rune]int{}, NamedWidth: map[string]int{}}
		for scanner.Scan() {
			l := scanner.Text()

			// Hackish way of ignoring lines without metrics.
			// This does not work for all AFM files, but it works for Core14.
			if !strings.HasPrefix(l, "C ") {
				continue
			}

			var c rune
			var w int
			var n string

			// Not all AFM lines will be formatted this way. This hack is
			// only suitable for parsing the Core14.
			if _, err := fmt.Sscanf(l, "C %d ; WX %d ; N %s ", &c, &w, &n); err != nil {
				log.Fatalf("Error parsing afm line %q: %v", l, err)
			}
			metrics.NamedWidth[n] = w
			if c == -1 {
				continue
			}
			metrics.RuneWidth[c] = w
		}
		basename := filepath.Base(filename)
		fontMetrics[basename[:len(basename)-4]] = metrics
	}
}

func MustGetFontMetrics(font string) *FontMetrics {
	m, ok := fontMetrics[font]
	if !ok {
		log.Fatalf("Could not find font metrics for font %q.", font)
	}
	return m
}

// TypeSetter is an object that performs typesetting.
type TypeSetter struct {
	Metrics    *FontMetrics
	Pt         float64
	LineHeight float64
	X1, Y1, X2 float64
}

// OutLine contains one typesetted output line.
type OutLine struct {
	X, Y float64
	Pt   float64
	Text string
}

// OutLines is a wrapper around a sequence of output lines.
type OutLines []OutLine

// AppendToPDFStream converts the output lines into PDF text objects.
func (l OutLines) AppendToPDFStream(font string, r float64, g float64, b float64, s *PDFStream) {
	for _, line := range l {
		*s = append(*s,
			PDFTextObject{
				R: r, G: g, B: b,
				X: line.X, Y: line.Y,

				Font:     PDFName(font),
				FontSize: line.Pt,

				Text: PDFString(line.Text),
			},
		)
	}
}

// Set performs typesetting and returns the output lines and the new Y position.
func (s *TypeSetter) Set(Text string) (OutLines, float64) {
	avail := s.X2 - s.X1
	result := OutLines{}

	y := s.Y1
	start := 0
	lastBreak := -1

	var w float64
Outer:
	for {
		y -= s.LineHeight
		w = 0
		for i, r := range Text[start:] {
			if r == ' ' || r == '-' {
				lastBreak = i
			}
			w += float64(s.Metrics.RuneWidth[r]) * s.Pt / 1000.0
			if w >= avail && lastBreak != -1 {
				result = append(result, OutLine{
					X: s.X1, Y: y, Pt: s.Pt, Text: Text[start : start+lastBreak],
				})
				start = start + lastBreak + 1
				continue Outer
			}
		}
		break
	}

	result = append(result, OutLine{
		X: s.X1, Y: y, Pt: s.Pt, Text: Text[start:],
	})
	return result, y
}
