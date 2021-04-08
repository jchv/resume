# 2021-04-06
The project commenses.

First randevú with the actual PDF format: `PDFReference.pdf`. Provided by Adobe, it describes the PDF format in reasonable detail. The first thing that surprises me is that PDF is largely text-based. Never having actually opened one, I opened `PDFReference.pdf` in Notepad and to my surprise, it was true.

The start of `PDFReference.pdf` looks like this:

```
%PDF-1.4
%öäüß
```

The first line rather clearly signifies the version of the PDF specification. The second line is a bit more perplexing. It turns out it is a canary to detect mangling. The specification recommends putting 4 bytes >= 0x80 here, so that mangling from ASCII-clean transports (such as MIME?) can be detected. It is not required, and presumably if you write an ASCII-clean PDF document it is not useful.

The next lines are:

```
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Metadata 3 0 R
/Outlines 4 0 R
/PageLabels 5 0 R
/OpenAction 6 0 R
/PageMode /UseOutlines
/ViewerPreferences <<
/DisplayDocTitle true
>>
endobj
```

There is a lot going on here. The first line signifies 3 things:

* This is a PDF object.
* It is object number 1.
* It is revision 0 of the object.

PDF is a format designed to be modifiable. For our purposes, we only need to be concerned with revision 0, since we are just writing an initial document.

The next line begins a PDF dictionary. This is essentially just a table or map structure. The keys are PDF names, not strings. Names are signified with a /. The encoding for strings is not shown here as we have yet to come across it.

The triplets ending in "R" like "2 0 R" are indirect references. They point to other PDF objects. for "2 0 R", it points to revision 0 of object 2, which is, as you might expect, a Pages object.

From this point forward, `PDFReference.pdf` becomes a less and less useful reference of the format for a couple of reasons:

* It uses PDF's famously stupid "encryption" mechanism, so a lot of strings are gibberish.
* It is huge (> 8 MiB) so it uses the PDF format in ways that are not necessary for us. (For example, it has a tree of Pages objects.)

There are a couple more things I needed to understand about the structure of PDF documents before I could begin writing them. Namely, the xref and trailer sections. So far, looking at `PDFReference.pdf`, we've only seen the header and the body.

The xrefs section eludes me for now, but thankfully PDF readers I tested do not care. Understanding the basics of the catalog and pages object, and a minimal bit about the xref and trailer sections, I am able to produce a well-formed 0 page PDF that loads in Microsoft Edge. At this point, the PDF object code has been split into `pdf.go`.

# 2021-04-07

It is time to add a page. To do this we need to understand operators.

First, let's add a page. The page will look like this:

```
4 0 obj
<<
/Type /Page
/MediaBox [ 0 0 612 792 ]
/Contents 5 0 R
/Parent 3 0 R
/Resources <<
/Font <<
/F1 7 0 R
/F2 8 0 R
>>
/ProcSet 6 0 R
>>
>>
endobj
```

Once again, a lot going on here. This is another object, so the first line is familiar. We're dealing with a dictionary as usual.

The first key, `/MediaBox`, specifies the page size. (There's actually a few different 'boxes' associated with a PDF page, and for serious graphic design work we may want to specify both the MediaBox and CropBox to print into the bleed area, but for our purposes we can consider this the physical page size.) What is 612x792? Adobe likes to use units of 1/72 of an inch. 612/72 = 8.5, 792/72 = 11 - in other words, 8.5 x 11 inches, or U.S. Letter size.

`/Contents` merely points to the stream of objects for the contents of the page. `/Parent` points back to the Pages object. `/Resources` contains a list of fonts - and the indirect references in those point to Font objects. We'll be using some of the Base14 fonts that are built into the PDF specification.

`/ProcSet` is actually here for compatibility. It's not too interesting but I write it anyways just in case some older PDF readers care.

With all of that out of the way, we can talk about the contents stream.

```
5 0 obj
<< /Length 221 >>
stream
0.35 0.3 0.35 RG 58 704 m 554 704 l s
BT
0.35 0.3 0.35 rg
/F2 32.000000 Tf
58.00 710.00 Td
<4a6f686e20436861647769636b> Tj
ET
BT
0.35 0.3 0.35 rg
/F1 12.000000 Tf
58.00 690.00 Td
<536f66747761726520456e67696e656572> Tj
ET
endstream
endobj
```

The first thing that is important is the formatting of streams. Stream objects look like this:

```
<< /Length [...] >>
stream
[...]
endstream
```

It's additionally possible to encode streams in various ways, and you can specify that using additional keys to the dictionary. However, for our implementation, we'll hardcode it to work like this.

The stream itself looks daunting but it is not very scary, just very terse. Here is a breakdown:

```
0.35 0.3 0.35 RG  % Sets the color for stroked operations to rgb(0.35, 0.3, 0.35)
58 704 m          % Move the cursor to 58, 704. Note that these are coordinates from bottom-left.
554 704 l         % Line-to 554, 704.
s                 % Close + stroke line.

BT                % Begin a text object.
0.35 0.3 0.35 rg  % Sets the color for non-stroked operations to rgb(0.35, 0.3, 0.35)
/F2 32.000000 Tf  % Sets the font to /F2 at 32pt.
58.00 710.00 Td   % Sets the position to 58, 710. 
<4a6f686e2...> Tj % Draws the text. (hex encoded string)
ET
```

These operations may seem somewhat familiar: they're reminiscent of PostScript and PostScript-inspired graphics models like SVG or HTML Canvas.

For good measure, let's talk about the remaining objects.

The ProcSet object is an array specifying the types of objects to expect in the document. As mentioned before it is strictly used for compatibility purposes.

```
6 0 obj
[ /PDF /Text  ]
endobj
```

We specify a couple of font objects. They just reference Helvetica and Helvetica-Bold, part of the Base 14. For now, I am not overly concerned about encoding, though if we were writing a robust software dealing in PDF we certainly should.

```
7 0 obj
<<
/Encoding /WinAnsiEncoding
/Type /Font
/Subtype /Type1
/Name /F1
/BaseFont /Helvetica
>>
endobj

8 0 obj
<<
/Type /Font
/Subtype /Type1
/Name /F2
/BaseFont /Helvetica-Bold
/Encoding /WinAnsiEncoding
>>
endobj
```

Finally, after the objects we have the xref table, trailer, a pointer to the xref table, and the EOF.

```
xref
0 0

trailer
<<
/Size 0
/Root 1 0 R
>>

startxref
892
%%EOF
```

Bugs in my code have made me aware that PDF parsers are incredibly robust, so it is hard for me to know if my code is actually correct or not. For example, I accidentally had `startxref` set to 0, and both GhostScript and PDFium had no problem at all!

We have an empty xref table here anyways, since I do not understand it yet. The trailer is just another dictionary.

## Type Setting
At this point, it has occurred to me that I will need to write at least a minimal typesetting engine. In order to accomplish this, I will need font metrics data.

As it turns out, Adobe distributes permissively-licensed copies of the font metrics data for the Core 14 fonts. Though it is not linked anywhere on Adobe's site, that I can see, they are downloadable here:

https://download.macromedia.com/pub/developer/opentype/tech-notes/Core14_AFMs.zip

These files look pretty simple. Each line is a directive, and each directive starts with a keyword followed by some data. There's a full specification over here:

https://adobe-type-tools.github.io/font-tech-notes/pdfs/5004.AFM_Spec.pdf

However, for the sake of making life easier, we will probably very sparingly concern ourselves with the specification, and just do what is necessary to parse the Core 14 AFMs.

The actual character metrics data looks like this:

```
C 32 ; WX 600 ; N space ; B 0 0 0 0 ;
```

`C 32` specifies character 32, 0x20, ' ' (space). WX 600 specifies something a bit odd. It specifies the width of a character in thousandths of the 1/72" unit, for the font at 1pt. So, if we've done our math correctly... 600 * 12 / 1000 = 7.2 units. So at 12pt, the character should be 7.2 units wide. We're looking at Courier, which is a fixed-width font, so actually, all of the characters in the font are like this.

After some quick wrangling with scanf I have the Core 14 AFM data (at least character widths, for now) loaded into my program; it should be possible to make the most basic typesetting engine.

And with a bit of hacking...

```go
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
```

We can do basic LTR typesetting using the metrics.

## Formatting/Polish

Armed with typesetting capabilities, I now begin writing much of the actual resume layout process. The function that lays out the PDF is a little ugly, since it's essentially doing the work a layout engine would normally do, except by hand.

## Obfuscation

Because the document contains some somewhat sensitive details, I decided to obfuscate it. You can deobfuscate it by passing in the secret to the first argument of the program.

# Conclusion

Although there are many improvements that could be made, such as writing a proper layout engine, more advanced type-setting, improvements to the PDF abstractions, etc. for now I am considering this project a success. It took a total of two nights to get to this point.

Key takeaways:

*   PDF is not just binary PostScript. Not only that, it's not even necessarily *binary*. Never knew.
*   The PDF format has some wrinkles and complexities, but at its core it isn't too bad.
*   When dealing with PDF, utilizing the Core 14 fonts can let you produce simpler and smaller PDF files.
