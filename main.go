// Package main implements a program that dumps an up-to-date résumé for me
// into PDF format. It does not depend on any packages aside from the standard
// library.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Résumé contains the information for the resume.
type Résumé struct {
	Name       string
	Trade      string
	Tel        string
	Email      string
	Experience []struct {
		Name         string
		URL          string
		Since        string
		Ended        string `json:",omitempty"`
		Technologies []string
		Summary      string
	}
}

// Obfuscates the résumé using the provided passphrase.
func (r *Résumé) Obfuscate(passphrase string) {
	c := NewCipher(passphrase)
	c.PadStr(&r.Name)
	c.PadStr(&r.Trade)
	c.PadStr(&r.Tel)
	c.PadStr(&r.Email)
	for i := range r.Experience {
		c.PadStr(&r.Experience[i].Name)
		c.PadStr(&r.Experience[i].URL)
		c.PadStr(&r.Experience[i].Since)
		c.PadStr(&r.Experience[i].Ended)
		for j := range r.Experience[i].Technologies {
			c.PadStr(&r.Experience[i].Technologies[j])
		}
		c.PadStr(&r.Experience[i].Summary)
	}
}

// ToPDF produces a PDF for the résumé.
func (r *Résumé) ToPDF() PDFDocument {
	doc := PDFDocument{}

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):     PDFName("Catalog"),
		PDFName("Outlines"): PDFReference(2),
		PDFName("Pages"):    PDFReference(3),
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):  PDFName("Outlines"),
		PDFName("Count"): PDFNumeric(0),
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):  PDFName("Pages"),
		PDFName("Count"): PDFNumeric(1),
		PDFName("Kids"):  PDFArray{PDFReference(4)},
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):   PDFName("Page"),
		PDFName("Parent"): PDFReference(3),
		PDFName("Resources"): PDFDictionary{
			PDFName("Font"): PDFDictionary{
				PDFName("F1"): PDFReference(7),
				PDFName("F2"): PDFReference(8),
				PDFName("F3"): PDFReference(9),
			},
			PDFName("ProcSet"): PDFReference(6),
		},
		// Letter size. Note that this unit is 1/72".
		PDFName("MediaBox"): PDFArray{
			PDFNumeric(0),
			PDFNumeric(0),
			PDFNumeric(612),
			PDFNumeric(792),
		},
		PDFName("Contents"): PDFReference(5),
	})

	regularMetrics := MustGetFontMetrics("Helvetica")
	boldMetrics := MustGetFontMetrics("Helvetica-Bold")
	obliqueMetrics := MustGetFontMetrics("Helvetica-Oblique")

	contents := PDFStream{
		PDFTextObject{
			R: 0.35, G: 0.3, B: 0.35,
			X: 58, Y: 710,

			Font:     PDFName("F2"),
			FontSize: 32,

			Text: PDFString(r.Name),
		},
		PDFRuleObject{
			Width: 2,
			R:     0.35, G: 0.3, B: 0.35,
			X1: 58, X2: 554, Y: 704,
		},
		PDFTextObject{
			R: 0.35, G: 0.3, B: 0.35,
			X: 58, Y: 686,

			Font:     PDFName("F1"),
			FontSize: 12,

			Text: PDFString(r.Trade),
		},
		PDFTextObject{
			R: 0.35, G: 0.3, B: 0.35,
			X: 58, Y: 670,

			Font:     PDFName("F1"),
			FontSize: 12,

			Text: PDFString("E-mail: " + r.Email),
		},
		PDFTextObject{
			R: 0.35, G: 0.3, B: 0.35,
			X: 58, Y: 654,

			Font:     PDFName("F1"),
			FontSize: 12,

			Text: PDFString("Tel: " + r.Tel),
		},
	}

	y := 640.0
	for _, exp := range r.Experience {
		var l OutLines

		// Title
		set := TypeSetter{
			Metrics:    boldMetrics,
			Pt:         20,
			LineHeight: 24,
			X1:         58, Y1: y, X2: 554,
		}
		l, y = set.Set(exp.Name)
		l.AppendToPDFStream("F2", 0.35, 0.3, 0.35, &contents)

		contents = append(contents,
			PDFRuleObject{
				Width: 1,

				R: 0.35, G: 0.3, B: 0.35,
				X1: 58, X2: 554, Y: y - 6,
			},
		)

		y -= 24

		// Timeline
		timeline := fmt.Sprintf("Since %s", exp.Since)
		if exp.Ended != "" {
			timeline = fmt.Sprintf("From %s until %s", exp.Since, exp.Ended)
		}
		contents = append(contents, PDFTextObject{X: 58, Y: y, Font: PDFName("F3"), FontSize: 12, Text: PDFString(timeline)})

		y -= 8

		// Summary
		set = TypeSetter{
			Metrics:    regularMetrics,
			Pt:         12,
			LineHeight: 18,
			X1:         58, Y1: y, X2: 554,
		}
		l, y = set.Set(exp.Summary)
		l.AppendToPDFStream("F1", 0, 0, 0, &contents)

		y -= 8

		// Technologies
		set = TypeSetter{
			Metrics:    obliqueMetrics,
			Pt:         10,
			LineHeight: 18,
			X1:         58, Y1: y, X2: 554,
		}
		l, y = set.Set(fmt.Sprintf("Technologies: %s", strings.Join(exp.Technologies, ", ")))
		l.AppendToPDFStream("F3", 0, 0, 0, &contents)

		y -= 24
	}

	contents = append(contents,
		PDFTextObject{
			R: 0.35, G: 0.3, B: 0.35,
			X: 58, Y: 50,

			Font:     PDFName("F3"),
			FontSize: 10,

			Text: PDFString("This PDF made from scratch with <3 - source: https://github.com/jchv/resume"),
		},
	)

	doc.Objects = append(doc.Objects, contents)

	doc.Objects = append(doc.Objects, PDFArray{
		PDFName("PDF"),
		PDFName("Text"),
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):     PDFName("Font"),
		PDFName("Subtype"):  PDFName("Type1"),
		PDFName("Name"):     PDFName("F1"),
		PDFName("BaseFont"): PDFName("Helvetica"),
		PDFName("Encoding"): PDFName("WinAnsiEncoding"),
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):     PDFName("Font"),
		PDFName("Subtype"):  PDFName("Type1"),
		PDFName("Name"):     PDFName("F2"),
		PDFName("BaseFont"): PDFName("Helvetica-Bold"),
		PDFName("Encoding"): PDFName("WinAnsiEncoding"),
	})

	doc.Objects = append(doc.Objects, PDFDictionary{
		PDFName("Type"):     PDFName("Font"),
		PDFName("Subtype"):  PDFName("Type1"),
		PDFName("Name"):     PDFName("F3"),
		PDFName("BaseFont"): PDFName("Helvetica-Oblique"),
		PDFName("Encoding"): PDFName("WinAnsiEncoding"),
	})

	return doc
}

func main() {
	flag.Parse()

	f, err := os.Open("resume.json")
	if err != nil {
		log.Fatalf("Error opening resume.json: %v", err)
	}
	r := Résumé{}
	if err := json.NewDecoder(f).Decode(&r); err != nil {
		log.Fatalf("Error decoding resume.json: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Printf("Warning: failed to close resume.json: %v", err)
	}

	r.Obfuscate(flag.Arg(0))

	f, err = os.Create("resume.pdf")
	if err != nil {
		log.Fatalf("Error opening resume.pdf: %v", err)
	}
	if err := r.ToPDF().WritePDF(f); err != nil {
		log.Fatalf("Error writing resume.pdf: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Printf("Warning: failed to close resume.pdf: %v", err)
	}
}
