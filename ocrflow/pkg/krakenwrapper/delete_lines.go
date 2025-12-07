package krakenwrapper

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func DeleteLines(src, dst string) error {
	// If src and dst are the same path, write to a temp file and rename at the end.
	tmpPath := dst
	inplace := src == dst

	if inplace {
		dir := filepath.Dir(dst)
		tmpFile, err := os.CreateTemp(dir, "deletelines-*.xml")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		tmpPath = tmpFile.Name()
		_ = tmpFile.Close() // will reopen below with os.Create
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	// Do not defer here if we need to rename; close explicitly later.
	// But still make sure it closes on early returns.
	defer func() {
		_ = out.Close()
	}()

	dec := xml.NewDecoder(in)
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")

	// Write XML header explicitly
	if _, err := io.WriteString(out, xml.Header); err != nil {
		return fmt.Errorf("write xml header: %w", err)
	}

	// Helper to skip an entire element (starting just after its StartElement token)
	skipElement := func(d *xml.Decoder) error {
		depth := 1
		for depth > 0 {
			tok, err := d.Token()
			if err != nil {
				return err
			}
			switch tok.(type) {
			case xml.StartElement:
				depth++
			case xml.EndElement:
				depth--
			}
		}
		return nil
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode token: %w", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			// Match by local name so namespaces do not matter
			if se.Name.Local == "TextLine" {
				// Skip this entire <TextLine>...</TextLine> subtree
				if err := skipElement(dec); err != nil {
					return fmt.Errorf("skip TextLine: %w", err)
				}
				// Do not write the start token or any of its contents
				continue
			}
		}

		if err := enc.EncodeToken(tok); err != nil {
			return fmt.Errorf("encode token: %w", err)
		}
	}

	if err := enc.Flush(); err != nil {
		return fmt.Errorf("flush encoder: %w", err)
	}

	// Close output before rename to be safe
	if err := out.Close(); err != nil {
		return fmt.Errorf("close dst: %w", err)
	}

	// If we wrote to a temp file, move it over the original path
	if inplace {
		if err := os.Rename(tmpPath, dst); err != nil {
			return fmt.Errorf("rename temp to dst: %w", err)
		}
	}

	return nil
}
