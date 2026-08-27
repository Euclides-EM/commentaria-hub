package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDocumentURL(t *testing.T) {
	doc, err := parseDocumentURL("https://app.transkribus.org/sites/noscemus/doc/912317?colId=88639")
	if err != nil {
		t.Fatal(err)
	}
	want := documentRef{Site: "noscemus", DocumentID: 912317, CollectionID: 88639}
	if doc != want {
		t.Fatalf("got %#v, want %#v", doc, want)
	}
}

func TestParseDocumentURLRejectsInvalidLinks(t *testing.T) {
	tests := []string{
		"https://example.org/sites/noscemus/doc/912317?colId=88639",
		"https://app.transkribus.org/sites/noscemus/doc/not-a-number?colId=88639",
		"https://app.transkribus.org/sites/noscemus/doc/912317",
		"https://app.transkribus.org/doc/912317?colId=88639",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := parseDocumentURL(input); err == nil {
				t.Fatalf("parseDocumentURL(%q) unexpectedly succeeded", input)
			}
		})
	}
}

func TestParseDownloadMode(t *testing.T) {
	tests := map[string]downloadMode{
		"both":   {XML: true, Images: true},
		"xml":    {XML: true},
		"images": {Images: true},
	}
	for input, want := range tests {
		got, err := parseDownloadMode(input)
		if err != nil {
			t.Fatalf("parseDownloadMode(%q): %v", input, err)
		}
		if got != want {
			t.Errorf("parseDownloadMode(%q) = %#v, want %#v", input, got, want)
		}
	}
	if _, err := parseDownloadMode("neither"); err == nil {
		t.Fatal("invalid mode unexpectedly succeeded")
	}
}

func TestFullIIIFImageURL(t *testing.T) {
	got, err := fullIIIFImageURL("https://files.transkribus.eu/iiif/2/IMAGEKEY/info.json")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://files.transkribus.eu/iiif/2/IMAGEKEY/full/full/0/default.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDownloadDocumentStoresAndResumesPageByPage(t *testing.T) {
	pageRequests := map[string]int{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageRequests[r.URL.Path]++
		switch r.URL.Path {
		case "/documents/noscemus/912317":
			fmt.Fprint(w, `{"id":912317,"title":"Test document","pageCount":2}`)
		case "/documents/noscemus/912317/pages/1":
			fmt.Fprintf(w, `{"content":%q,"image":%q}`, server.URL+"/files/one", server.URL+"/iiif/one/info.json")
		case "/documents/noscemus/912317/pages/2":
			fmt.Fprintf(w, `{"content":%q,"image":%q}`, server.URL+"/files/two", server.URL+"/iiif/two/info.json")
		case "/files/one":
			fmt.Fprint(w, `<?xml version="1.0"?><PcGts><Page id="one"/></PcGts>`)
		case "/files/two":
			fmt.Fprint(w, `<?xml version="1.0"?><PcGts><Page id="two"/></PcGts>`)
		case "/iiif/one/full/full/0/default.jpg", "/iiif/two/full/full/0/default.jpg":
			w.Write([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	outputDir := t.TempDir()
	existing := filepath.Join(outputDir, "page-0001.xml")
	if err := os.WriteFile(existing, []byte("already here"), 0o644); err != nil {
		t.Fatal(err)
	}
	var progress bytes.Buffer
	doc := documentRef{Site: "noscemus", DocumentID: 912317, CollectionID: 88639}
	if err := downloadDocument(context.Background(), server.Client(), server.URL+"/documents", doc, downloadMode{XML: true, Images: true}, outputDir, &progress); err != nil {
		t.Fatal(err)
	}

	if pageRequests["/files/one"] != 0 {
		t.Fatal("already downloaded PAGE XML was requested")
	}
	contents, err := os.ReadFile(filepath.Join(outputDir, "page-0002.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `id="two"`) {
		t.Fatalf("unexpected page contents: %s", contents)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "page-0002.jpg")); err != nil {
		t.Fatalf("image was not stored: %v", err)
	}
	output := progress.String()
	for _, expected := range []string{
		"Storing in ",
		"Skipped PAGE XML for page 1 out of 2 as it is already downloaded",
		"Downloading image for page 1 out of 2",
		"Downloading PAGE XML for page 2 out of 2",
		"Stored image for page 2 in ",
		"Done: all 2 pages",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("progress output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestDownloadDocumentDoesNotStoreInvalidXML(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/documents/site/1":
			fmt.Fprint(w, `{"id":1,"pageCount":1}`)
		case "/documents/site/1/pages/1":
			fmt.Fprintf(w, `{"content":%q}`, server.URL+"/bad")
		case "/bad":
			fmt.Fprint(w, `<html>not PAGE XML</html>`)
		}
	}))
	defer server.Close()

	outputDir := t.TempDir()
	err := downloadDocument(context.Background(), server.Client(), server.URL+"/documents", documentRef{Site: "site", DocumentID: 1, CollectionID: 2}, downloadMode{XML: true}, outputDir, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "expected PcGts") {
		t.Fatalf("got error %v, want PcGts validation error", err)
	}
	if _, statErr := os.Stat(filepath.Join(outputDir, "page-0001.xml")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid XML was stored; stat error: %v", statErr)
	}
}
