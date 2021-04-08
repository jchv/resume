# Résumé
This repository contains a program that generates my résumé.

**Important Note**: The résumé data is obfuscated. In order to unobfuscate it, you need to pass a secret into the first argument of the program.

## The Problem

Writing a résumé is hard and full of decisions. How do you produce a PDF file anymore, for example?

*   Word processor: It's fine, obviously, if all you want to do is make a PDF and move on. But this is your résumé. You want to update it over time, and maybe you want to have an HTML, or even Markdown representation of it. Updates can be cumbersome as the documents are unstructured, HTML output will be a mess, and something like Markdown is pretty much out of the question. Worst of all, really can't get any nerd cred for using Word.

*   L<small><sup>A</sup></small>T<sub>E</sub>X: Now we're playing with power. But let's be honest, does anyone even really *like* using LaTeX? Which TeX distribution should you install? Are you supposed to use PDFLaTeX or XeTeX? What in the hell is an underfull `\hbox`?

*   asciidoctor: This is pretty cool, and it works pretty nicely. ... But we get little nerd cred for this, so it's out.

## The Solution
What if we wrote our résumé in a programming language? The idea is simple. We just structure the information in our résumé as code, and then voila. Well, OK. We have to write some code to generate output. *How hard could it be?*

Armed with the PDF specification, this repository is my attempt to find out. I am working in the Go programming language simply because I find it to be an effective programming language for writing software quickly. Work began on 2021-04-06, and the original goals were accomplished on 2021-04-07. Some improvements may be on-going over time, such as the addition of a layout engine, or better PDF output.

## Index

*   `afm/` - Adobe Core14 AFM data. Contains font metrics for the 'Core 14' fonts.
*   `afm.go` - Code for parsing the font metrics. This assumes it is parsing the Core 14 AFMs for simplicity.
*   `journal.md` - A development journal tracking my journey to understand and generate PDF format documents.
*   `main.go` - The entrypoint. Contains much of the meat.
*   `obfscipher.go` - A quick cipher intended for obfuscating the résumé data. As with all hand-rolled crypto, it is, of course, unsafe for anything mission critical.
*   `pdf.go` - Contains PDF object code. Contains a lot of lower-level PDF file format logic.
