// Command transkribus-downloader downloads PAGE XML and facsimile images for a
// document published through Transkribus Sites.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIBase  = "https://api-sites.transkribus.eu/search/documents"
	maxResponseSize = 100 << 20 // 100 MiB, large enough for unusually detailed PAGE XML.
)

type documentRef struct {
	Site         string
	DocumentID   int64
	CollectionID int64
}

type documentMetadata struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	PageCount int    `json:"pageCount"`
}

type pageMetadata struct {
	Content string `json:"content"`
	Image   string `json:"image"`
}

type downloadMode struct {
	XML    bool
	Images bool
}

func main() {
	outputDir := flag.String("output-dir", "", "directory in which to store downloaded files (default: transkribus-<document-id>-page-xml)")
	download := flag.String("download", "both", "what to download: both, xml, or images")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <transkribus-document-link>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "Downloads PAGE XML and facsimile images one page at a time and skips files already present.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		if flag.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "\nerror: missing Transkribus document link")
		} else {
			fmt.Fprintln(os.Stderr, "\nerror: expected exactly one Transkribus document link")
		}
		os.Exit(2)
	}

	doc, err := parseDocumentURL(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	mode, err := parseDownloadMode(*download)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if *outputDir == "" {
		*outputDir = fmt.Sprintf("transkribus-%d-page-xml", doc.DocumentID)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := &http.Client{Timeout: 60 * time.Second}
	if err := downloadDocument(ctx, client, defaultAPIBase, doc, mode, *outputDir, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseDownloadMode(raw string) (downloadMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "both":
		return downloadMode{XML: true, Images: true}, nil
	case "xml":
		return downloadMode{XML: true}, nil
	case "images", "image":
		return downloadMode{Images: true}, nil
	default:
		return downloadMode{}, fmt.Errorf("invalid -download value %q: expected both, xml, or images", raw)
	}
}

func parseDocumentURL(raw string) (documentRef, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return documentRef{}, fmt.Errorf("parse document link: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return documentRef{}, errors.New("document link must use http or https")
	}
	if !strings.EqualFold(u.Hostname(), "app.transkribus.org") {
		return documentRef{}, fmt.Errorf("expected an app.transkribus.org link, got host %q", u.Hostname())
	}

	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(parts) != 4 || parts[0] != "sites" || parts[2] != "doc" {
		return documentRef{}, errors.New("expected a link shaped like https://app.transkribus.org/sites/<site>/doc/<document-id>?colId=<collection-id>")
	}
	site, err := url.PathUnescape(parts[1])
	if err != nil || site == "" || strings.Contains(site, "/") {
		return documentRef{}, errors.New("document link contains an invalid site name")
	}
	documentID, err := positiveInt64(parts[3], "document ID")
	if err != nil {
		return documentRef{}, err
	}
	collectionID, err := positiveInt64(u.Query().Get("colId"), "colId")
	if err != nil {
		return documentRef{}, err
	}

	return documentRef{Site: site, DocumentID: documentID, CollectionID: collectionID}, nil
}

func positiveInt64(raw, name string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", name, raw)
	}
	return value, nil
}

func downloadDocument(ctx context.Context, client *http.Client, apiBase string, doc documentRef, mode downloadMode, outputDir string, progress io.Writer) error {
	if !mode.XML && !mode.Images {
		return errors.New("nothing selected for download")
	}
	metadataURL, err := documentAPIURL(apiBase, doc, 0)
	if err != nil {
		return err
	}
	var metadata documentMetadata
	if err := getJSON(ctx, client, metadataURL, &metadata); err != nil {
		return fmt.Errorf("get document metadata: %w", err)
	}
	if metadata.PageCount < 0 {
		return fmt.Errorf("document metadata returned invalid page count %d", metadata.PageCount)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", outputDir, err)
	}
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		absOutputDir = outputDir
	}
	fmt.Fprintf(progress, "Document: %s (%d pages)\n", metadata.Title, metadata.PageCount)
	fmt.Fprintf(progress, "Storing in %s\n", absOutputDir)

	width := len(strconv.Itoa(metadata.PageCount))
	if width < 4 {
		width = 4
	}
	for pageNumber := 1; pageNumber <= metadata.PageCount; pageNumber++ {
		baseName := fmt.Sprintf("page-%0*d", width, pageNumber)
		xmlPath := filepath.Join(outputDir, baseName+".xml")
		imagePath := filepath.Join(outputDir, baseName+".jpg")

		needXML, err := needsDownload(xmlPath, mode.XML)
		if err != nil {
			return fmt.Errorf("check PAGE XML output for page %d: %w", pageNumber, err)
		}
		needImage, err := needsDownload(imagePath, mode.Images)
		if err != nil {
			return fmt.Errorf("check image output for page %d: %w", pageNumber, err)
		}
		if mode.XML && !needXML {
			fmt.Fprintf(progress, "Skipped PAGE XML for page %d out of %d as it is already downloaded\n", pageNumber, metadata.PageCount)
		}
		if mode.Images && !needImage {
			fmt.Fprintf(progress, "Skipped image for page %d out of %d as it is already downloaded\n", pageNumber, metadata.PageCount)
		}
		if !needXML && !needImage {
			continue
		}

		pageURL, err := documentAPIURL(apiBase, doc, pageNumber)
		if err != nil {
			return err
		}
		var page pageMetadata
		if err := getJSON(ctx, client, pageURL, &page); err != nil {
			return fmt.Errorf("get metadata for page %d: %w", pageNumber, err)
		}
		if needXML {
			if page.Content == "" {
				return fmt.Errorf("metadata for page %d has no PAGE XML content URL", pageNumber)
			}
			fmt.Fprintf(progress, "Downloading PAGE XML for page %d out of %d\n", pageNumber, metadata.PageCount)
			contents, err := getBytes(ctx, client, page.Content)
			if err != nil {
				return fmt.Errorf("download PAGE XML for page %d: %w", pageNumber, err)
			}
			if err := validatePageXML(contents); err != nil {
				return fmt.Errorf("validate PAGE XML for page %d: %w", pageNumber, err)
			}
			if err := writeFileAtomically(xmlPath, contents, 0o644); err != nil {
				return fmt.Errorf("store PAGE XML for page %d in %q: %w", pageNumber, xmlPath, err)
			}
			fmt.Fprintf(progress, "Stored PAGE XML for page %d in %s\n", pageNumber, xmlPath)
		}

		if needImage {
			if page.Image == "" {
				return fmt.Errorf("metadata for page %d has no IIIF image URL", pageNumber)
			}
			imageURL, err := fullIIIFImageURL(page.Image)
			if err != nil {
				return fmt.Errorf("build image URL for page %d: %w", pageNumber, err)
			}
			fmt.Fprintf(progress, "Downloading image for page %d out of %d\n", pageNumber, metadata.PageCount)
			contents, err := getBytes(ctx, client, imageURL)
			if err != nil {
				return fmt.Errorf("download image for page %d: %w", pageNumber, err)
			}
			if err := validateImage(contents); err != nil {
				return fmt.Errorf("validate image for page %d: %w", pageNumber, err)
			}
			if err := writeFileAtomically(imagePath, contents, 0o644); err != nil {
				return fmt.Errorf("store image for page %d in %q: %w", pageNumber, imagePath, err)
			}
			fmt.Fprintf(progress, "Stored image for page %d in %s\n", pageNumber, imagePath)
		}
	}

	fmt.Fprintf(progress, "Done: all %d pages are stored in %s\n", metadata.PageCount, absOutputDir)
	return nil
}

func needsDownload(path string, selected bool) (bool, error) {
	if !selected {
		return false, nil
	}
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	return false, fmt.Errorf("stat %q: %w", path, err)
}

func fullIIIFImageURL(infoURL string) (string, error) {
	u, err := url.Parse(infoURL)
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("invalid IIIF info URL %q", infoURL)
	}
	if !strings.HasSuffix(u.Path, "/info.json") {
		return "", fmt.Errorf("IIIF URL does not end in /info.json: %q", infoURL)
	}
	u.Path = strings.TrimSuffix(u.Path, "/info.json") + "/full/full/0/default.jpg"
	u.RawPath = ""
	u.Fragment = ""
	return u.String(), nil
}

func documentAPIURL(apiBase string, doc documentRef, pageNumber int) (string, error) {
	base, err := url.Parse(strings.TrimRight(apiBase, "/"))
	if err != nil {
		return "", fmt.Errorf("parse API base URL: %w", err)
	}
	base.Path += "/" + doc.Site + "/" + strconv.FormatInt(doc.DocumentID, 10)
	if pageNumber > 0 {
		base.Path += "/pages/" + strconv.Itoa(pageNumber)
	}
	query := base.Query()
	query.Set("collection", strconv.FormatInt(doc.CollectionID, 10))
	query.Set("url", doc.Site)
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	contents, err := getBytes(ctx, client, endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, target); err != nil {
		return fmt.Errorf("decode response from %s: %w", endpoint, err)
	}
	return nil
}

func getBytes(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %s: %w", endpoint, err)
	}
	req.Header.Set("User-Agent", "ocrflow-transkribus-downloader/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseSize+1)
	contents, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", endpoint, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(contents))
		if len(message) > 500 {
			message = message[:500] + "..."
		}
		if message == "" {
			return nil, fmt.Errorf("GET %s returned %s", endpoint, resp.Status)
		}
		return nil, fmt.Errorf("GET %s returned %s: %s", endpoint, resp.Status, message)
	}
	if len(contents) > maxResponseSize {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", endpoint, maxResponseSize)
	}
	return contents, nil
}

func validatePageXML(contents []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(contents))
	foundRoot := false
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !foundRoot {
					return errors.New("response contains no XML root element")
				}
				return nil
			}
			return err
		}
		if start, ok := token.(xml.StartElement); ok && !foundRoot {
			if start.Name.Local != "PcGts" {
				return fmt.Errorf("expected PcGts root element, got %q", start.Name.Local)
			}
			foundRoot = true
		}
	}
}

func validateImage(contents []byte) error {
	contentType := http.DetectContentType(contents)
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("expected image data, got %s", contentType)
	}
	return nil
}

func writeFileAtomically(path string, contents []byte, mode os.FileMode) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".page-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if _, err := tmp.Write(contents); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
