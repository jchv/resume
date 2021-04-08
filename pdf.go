package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
)

// This file contains a minimalist PDF writer. On its own, it is not a
// particularly strict implementation of the PDF specification, though it
// should be more than possible to write fully-comformant documents using a
// small subset of PDF features.

// PDFObject is an interface for types that represent PDF objects. Normally,
// it is customary to write Go interfaces as an "-er" suffix version of their
// verb function, but a decision was made to deviate here just to improve code
// clarity.
type PDFObject interface {
	WritePDF(w io.Writer) error
}

// PDFRaw contains raw, uninterpreted PDF data.
type PDFRaw string

// WritePDF implements PDFObject.
func (r PDFRaw) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, string(r)); err != nil {
		return err
	}
	return nil
}

// PDFBoolean encapsulates a PDF boolean type.
type PDFBoolean bool

// WritePDF implements PDFObject.
func (b PDFBoolean) WritePDF(w io.Writer) error {
	var err error

	switch b {
	case true:
		_, err = io.WriteString(w, "true")
	case false:
		_, err = io.WriteString(w, "false")
	}

	return err
}

// PDFReference encapsulates an indirect reference.
type PDFReference int

// WritePDF implements PDFObject.
func (n PDFReference) WritePDF(w io.Writer) error {
	formatted := fmt.Sprintf("%d %d R", int(n), 0)

	if _, err := io.WriteString(w, formatted); err != nil {
		return err
	}

	return nil
}

// PDFNumeric encapsulates a PDF numeric type.
type PDFNumeric float64

// WritePDF implements PDFObject.
func (n PDFNumeric) WritePDF(w io.Writer) error {
	formatted := strconv.FormatFloat(float64(n), 'f', -1, 64)

	if _, err := io.WriteString(w, formatted); err != nil {
		return err
	}

	return nil
}

// PDFName encapsulates a PDF name type.
type PDFName string

// WritePDF implements PDFObject.
func (n PDFName) WritePDF(w io.Writer) error {
	escaped := make([]byte, 1, len(n)+1)
	escaped[0] = '/'

	for _, c := range n {
		// Recommended by PDF Reference, 3rd Edition.
		// Note that we do not support UTF-8 in names.
		if c >= 33 && c <= 126 {
			escaped = append(escaped, byte(c))
		} else {
			escaped = append(escaped, '#')
			escaped = append(escaped, strconv.FormatInt(int64(c), 16)...)
		}
	}

	if _, err := w.Write(escaped); err != nil {
		return err
	}

	return nil
}

// PDFString encapsulates a PDF string type.
type PDFString string

// WritePDF implements PDFObject.
func (d PDFString) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}

	if _, err := hex.NewEncoder(w).Write([]byte(d)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	return nil
}

// PDFArray encapsulates a PDF dictionary type.
type PDFArray []PDFObject

// WritePDF implements PDFObject.
func (a PDFArray) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, "[ "); err != nil {
		return err
	}

	for _, v := range a {
		if err := v.WritePDF(w); err != nil {
			return err
		}
		if _, err := w.Write([]byte{' '}); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, " ]"); err != nil {
		return err
	}

	return nil
}

// PDFDictionary encapsulates a PDF dictionary type.
type PDFDictionary map[PDFName]PDFObject

// WritePDF implements PDFObject.
func (d PDFDictionary) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, "<<\n"); err != nil {
		return err
	}

	for k, v := range d {
		if _, err := io.WriteString(w, "/"+string(k)+" "); err != nil {
			return err
		}
		if err := v.WritePDF(w); err != nil {
			return err
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, ">>"); err != nil {
		return err
	}

	return nil
}

// PDFTextObject provides a basic text object with fixed operators.
// This could be improved by composing it out of operators instead.
type PDFTextObject struct {
	R, G, B  float64
	Font     PDFName
	FontSize float64
	X, Y     float64
	Text     PDFString
}

// WritePDF implements PDFObject.
func (t PDFTextObject) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, "BT\n"); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("%f %f %f rg\n", t.R, t.G, t.B)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("/%s %f Tf\n", string(t.Font), t.FontSize)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("%.2f %.2f Td\n", t.X, t.Y)); err != nil {
		return err
	}

	if err := t.Text.WritePDF(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, " Tj\nET"); err != nil {
		return err
	}

	return nil
}

// PDFRuleObject provides a basic horizontal rule.
type PDFRuleObject struct {
	Width     float64
	R, G, B   float64
	X1, X2, Y float64
}

// WritePDF implements PDFObject.
func (r PDFRuleObject) WritePDF(w io.Writer) error {
	if _, err := io.WriteString(w, fmt.Sprintf("%f w\n", r.Width)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("%f %f %f RG\n", r.R, r.G, r.B)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("%.2f %.2f m\n", r.X1, r.Y)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("%.2f %.2f l s\n", r.X2, r.Y)); err != nil {
		return err
	}

	return nil
}

// PDFStream contains a stream of PDF objects.
type PDFStream []PDFObject

// WritePDF implements PDFObject.
func (s PDFStream) WritePDF(w io.Writer) error {
	b := bytes.Buffer{}
	for i, obj := range s {
		if i != 0 {
			if _, err := b.WriteRune('\n'); err != nil {
				return err
			}
		}
		if err := obj.WritePDF(&b); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, fmt.Sprintf("<< /Length %d >>\nstream\n", b.Len())); err != nil {
		return err
	}
	if _, err := io.Copy(w, &b); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\nendstream\n"); err != nil {
		return err
	}
	return nil
}

// PDFDocument contains the basic structure of a PDF document.
type PDFDocument struct {
	Objects []PDFObject
}

// WriteCounter is a wrapper around an io.Writer that counts bytes written. We
// can use this to be able to write byte offsets into the file.
type WriteCounter struct {
	Writer io.Writer
	Count  int
}

// Write implements io.Writer
func (w *WriteCounter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	w.Count += n
	return n, err
}

// WritePDF implements PDFObject.
func (d PDFDocument) WritePDF(w io.Writer) error {
	c := WriteCounter{Writer: w}
	w = &c

	if _, err := io.WriteString(w, "%PDF-1.4\n\n"); err != nil {
		return err
	}

	objOffsets := []int{}

	for i, n := range d.Objects {
		objOffsets = append(objOffsets, c.Count)
		if _, err := io.WriteString(w, fmt.Sprintf("%d %d obj\n", i+1, 0)); err != nil {
			return err
		}
		if err := n.WritePDF(w); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\nendobj\n\n"); err != nil {
			return err
		}
	}

	xrefOffset := c.Count

	if _, err := io.WriteString(w, fmt.Sprintf("xref\n0 %d\n0000000000 65535 f\r\n", len(objOffsets)+1)); err != nil {
		return err
	}
	for _, offset := range objOffsets {
		if _, err := io.WriteString(w, fmt.Sprintf("%010d 00000 n\r\n", offset)); err != nil {
			return err
		}
	}

	trailer := PDFDictionary{
		PDFName("Size"): PDFNumeric(len(objOffsets) + 1),
		PDFName("Root"): PDFReference(1),
	}
	if _, err := io.WriteString(w, "\ntrailer\n"); err != nil {
		return err
	}
	if err := trailer.WritePDF(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, fmt.Sprintf("\n\nstartxref\n%d\n%%%%EOF\n", xrefOffset)); err != nil {
		return err
	}

	return nil
}
