package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const userAgent = "Mozilla/5.0 (compatible; QuantaLogoCrawler/1.0; +https://example.com)"

type fetchResult struct {
	Code    string
	Symbol  string
	PageURL string
	LogoURL string
	Err     error
}

type logoFailure struct {
	Code    string `json:"code,omitempty"`
	Symbol  string `json:"symbol"`
	PageURL string `json:"page_url"`
	Error   string `json:"error"`
}

type logoMapping struct {
	GeneratedAt     string            `json:"generated_at"`
	Source          string            `json:"source"`
	OutputDir       string            `json:"output_dir"`
	CodeToSymbol    map[string]string `json:"code_to_symbol"`
	CodeFailures    map[string]string `json:"code_failures,omitempty"`
	SymbolToFile    map[string]string `json:"symbol_to_file"`
	SymbolToLogoURL map[string]string `json:"symbol_to_logo_url"`
	LogoURLToFile   map[string]string `json:"logo_url_to_file"`
	Failures        []logoFailure     `json:"failures,omitempty"`
}

type crawlConfig struct {
	prefix    string
	padDigits bool
}

func main() {
	csvPath := flag.String("csv", "data/XHKG.csv", "path to XHKG csv file")
	symbolsPath := flag.String("symbols", "", "path to TSV file containing symbols (optional)")
	symbolColumn := flag.Int("symbol-column", 1, "0-based symbol column for -symbols input")
	prefix := flag.String("prefix", "", "exchange prefix for TradingView symbol (e.g. NASDAQ, NYSE)")
	outDir := flag.String("out", "data/xxx", "output directory for logos and mapping")
	mappingFile := flag.String("mapping", "", "path to mapping json (default: <out>/logo-mapping.json)")
	workers := flag.Int("workers", 4, "number of concurrent workers")
	delay := flag.Duration("delay", 300*time.Millisecond, "delay between requests per worker")
	timeout := flag.Duration("timeout", 20*time.Second, "HTTP timeout per request")
	overwrite := flag.Bool("overwrite", false, "overwrite existing logo files")
	resume := flag.Bool("resume", true, "resume from existing mapping file if present")
	flushEvery := flag.Int("flush-every", 100, "flush mapping file every N results (0 = only at end)")
	max := flag.Int("max", 0, "max number of symbols to process (0 = all)")
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	if *workers < 1 {
		log.Fatal("workers must be >= 1")
	}

	inputMode := "xhkg"
	if strings.TrimSpace(*symbolsPath) != "" {
		inputMode = "symbols"
	}

	var codes []string
	var err error
	switch inputMode {
	case "symbols":
		codes, err = readSymbols(*symbolsPath, *symbolColumn)
	default:
		codes, err = readCodes(*csvPath)
	}
	if err != nil {
		log.Fatalf("read input: %v", err)
	}
	if *verbose {
		log.Printf("loaded %d codes", len(codes))
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	if *mappingFile == "" {
		*mappingFile = filepath.Join(*outDir, "logo-mapping.json")
	}

	mapping, err := loadMapping(*mappingFile)
	if err != nil {
		log.Fatalf("load mapping: %v", err)
	}
	if mapping.GeneratedAt == "" {
		mapping.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if mapping.Source == "" {
		mapping.Source = "tradingview"
	}
	mapping.OutputDir = *outDir

	crawlCfg := crawlConfig{
		prefix:    normalizePrefix(*prefix, inputMode),
		padDigits: inputMode == "xhkg",
	}

	if *resume {
		codes = filterPendingCodes(codes, mapping, *outDir)
	}
	if *max > 0 && *max < len(codes) {
		codes = codes[:*max]
	}
	if *verbose {
		log.Printf("pending %d codes", len(codes))
	}

	client := &http.Client{
		Timeout: *timeout,
	}

	jobs := make(chan string)
	results := make(chan fetchResult)

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for code := range jobs {
				if *verbose {
					log.Printf("start %s", code)
				}
				result := fetchLogo(context.Background(), client, code, crawlCfg)
				if *verbose {
					if result.Err != nil {
						log.Printf("done %s err=%v", code, result.Err)
					} else {
						log.Printf("done %s logo=%s", code, result.LogoURL)
					}
				}
				results <- result
				if *delay > 0 {
					time.Sleep(*delay)
				}
			}
		}()
	}

	go func() {
		for _, code := range codes {
			jobs <- code
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var processed int
	var lastFlush int
	for result := range results {
		processed++
		if result.Err != nil {
			if _, seen := mapping.CodeFailures[result.Code]; !seen {
				mapping.CodeFailures[result.Code] = result.Err.Error()
				mapping.Failures = append(mapping.Failures, logoFailure{
					Code:    result.Code,
					Symbol:  result.Symbol,
					PageURL: result.PageURL,
					Error:   result.Err.Error(),
				})
			}
			continue
		}

		mapping.CodeToSymbol[result.Code] = result.Symbol
		mapping.SymbolToLogoURL[result.Symbol] = result.LogoURL

		fileName, ok := mapping.LogoURLToFile[result.LogoURL]
		if ok {
			destPath := filepath.Join(*outDir, fileName)
			if _, err := os.Stat(destPath); err == nil {
				mapping.SymbolToFile[result.Symbol] = fileName
				continue
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				if _, seen := mapping.CodeFailures[result.Code]; !seen {
					mapping.CodeFailures[result.Code] = fmt.Sprintf("stat existing logo: %v", err)
					mapping.Failures = append(mapping.Failures, logoFailure{
						Code:    result.Code,
						Symbol:  result.Symbol,
						PageURL: result.PageURL,
						Error:   fmt.Sprintf("stat existing logo: %v", err),
					})
				}
				continue
			}
			ok = false
		}

		ext := extensionFromURL(result.LogoURL)
		fileName = result.Symbol
		if ext != "" {
			fileName = fileName + ext
		}
		destPath := filepath.Join(*outDir, fileName)

		if !*overwrite {
			if _, err := os.Stat(destPath); err == nil {
				mapping.LogoURLToFile[result.LogoURL] = fileName
				mapping.SymbolToFile[result.Symbol] = fileName
				continue
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				if _, seen := mapping.CodeFailures[result.Code]; !seen {
					mapping.CodeFailures[result.Code] = fmt.Sprintf("stat logo: %v", err)
					mapping.Failures = append(mapping.Failures, logoFailure{
						Code:    result.Code,
						Symbol:  result.Symbol,
						PageURL: result.PageURL,
						Error:   fmt.Sprintf("stat logo: %v", err),
					})
				}
				continue
			}
		}

		contentType, err := downloadLogo(context.Background(), client, result.LogoURL, destPath)
		if err != nil {
			if _, seen := mapping.CodeFailures[result.Code]; !seen {
				mapping.CodeFailures[result.Code] = err.Error()
				mapping.Failures = append(mapping.Failures, logoFailure{
					Code:    result.Code,
					Symbol:  result.Symbol,
					PageURL: result.PageURL,
					Error:   err.Error(),
				})
			}
			continue
		}

		if ext == "" {
			ext = extensionFromContentType(contentType)
			if ext == "" {
				ext = ".img"
			}
			newFileName := result.Symbol + ext
			newPath := filepath.Join(*outDir, newFileName)
			if newPath != destPath {
				if err := os.Rename(destPath, newPath); err != nil {
					if _, seen := mapping.CodeFailures[result.Code]; !seen {
						mapping.CodeFailures[result.Code] = fmt.Sprintf("rename logo: %v", err)
						mapping.Failures = append(mapping.Failures, logoFailure{
							Code:    result.Code,
							Symbol:  result.Symbol,
							PageURL: result.PageURL,
							Error:   fmt.Sprintf("rename logo: %v", err),
						})
					}
					continue
				}
				fileName = newFileName
			}
		}

		mapping.LogoURLToFile[result.LogoURL] = fileName
		mapping.SymbolToFile[result.Symbol] = fileName

		if *flushEvery > 0 && processed-lastFlush >= *flushEvery {
			if err := writeMapping(*mappingFile, mapping); err != nil {
				log.Printf("flush mapping: %v", err)
			} else {
				lastFlush = processed
			}
		}

		if processed%50 == 0 {
			log.Printf("processed %d/%d", processed, len(codes))
		}
	}

	if err := writeMapping(*mappingFile, mapping); err != nil {
		log.Fatalf("write mapping: %v", err)
	}

	log.Printf("done: %d processed, %d failures", len(codes), len(mapping.Failures))
}

func readCodes(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	var codes []string
	seen := make(map[string]struct{})
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) == 0 {
			continue
		}
		code := strings.TrimSpace(record[0])
		if code == "" || !isDigits(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}

	if len(codes) == 0 {
		return nil, errors.New("no stock codes found")
	}
	return codes, nil
}

func isDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != ""
}

func symbolCandidates(raw string, padDigits bool) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if !padDigits {
		return []string{strings.ToUpper(raw)}
	}
	if len(raw) < 4 {
		padded := strings.Repeat("0", 4-len(raw)) + raw
		if padded == raw {
			return []string{raw}
		}
		return []string{padded, raw}
	}
	return []string{raw}
}

func fetchLogo(ctx context.Context, client *http.Client, code string, cfg crawlConfig) fetchResult {
	candidates := symbolCandidates(code, cfg.padDigits)
	var lastErr error
	var lastSymbol string
	var lastPageURL string
	for _, candidate := range candidates {
		symbol := candidate
		if cfg.prefix != "" {
			symbol = cfg.prefix + symbol
		}
		pageURL := fmt.Sprintf("https://www.tradingview.com/symbols/%s/", url.PathEscape(symbol))
		logoURL, err := fetchLogoURL(ctx, client, pageURL)
		if err == nil && logoURL != "" {
			return fetchResult{
				Code:    code,
				Symbol:  symbol,
				PageURL: pageURL,
				LogoURL: logoURL,
			}
		}
		lastErr = err
		lastSymbol = symbol
		lastPageURL = pageURL
	}

	if lastErr == nil {
		lastErr = errors.New("logo not found")
	}
	return fetchResult{
		Code:    code,
		Symbol:  lastSymbol,
		PageURL: lastPageURL,
		Err:     lastErr,
	}
}

func normalizePrefix(prefix, inputMode string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		if inputMode == "xhkg" {
			return "HKEX-"
		}
		return ""
	}
	trimmed = strings.ToUpper(trimmed)
	if !strings.HasSuffix(trimmed, "-") {
		trimmed += "-"
	}
	return trimmed
}

func readSymbols(path string, column int) ([]string, error) {
	if column < 0 {
		return nil, errors.New("symbol column must be >= 0")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	var symbols []string
	seen := make(map[string]struct{})
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) <= column {
			continue
		}
		symbol := strings.TrimSpace(record[column])
		if symbol == "" {
			continue
		}
		symbol = strings.ToUpper(symbol)
		if symbol == "SYMBOL" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return nil, errors.New("no symbols found")
	}
	return symbols, nil
}

func fetchLogoURL(ctx context.Context, client *http.Client, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	rawLogoURL := findLogoURL(doc)
	if rawLogoURL == "" {
		return "", errors.New("logo not found")
	}

	return resolveURL(pageURL, rawLogoURL)
}

func resolveURL(pageURL, assetURL string) (string, error) {
	assetURL = strings.TrimSpace(assetURL)
	if assetURL == "" {
		return "", errors.New("empty logo url")
	}
	if strings.HasPrefix(assetURL, "//") {
		return "https:" + assetURL, nil
	}
	parsed, err := url.Parse(assetURL)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return assetURL, nil
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}

func findLogoURL(doc *html.Node) string {
	root := findNodeByClass(doc, "js-symbol-page-header-root")
	if root == nil {
		return ""
	}
	img := findFirstElement(root, "img")
	if img == nil {
		return ""
	}
	if src := attrValue(img, "src"); src != "" {
		return src
	}
	if src := attrValue(img, "data-src"); src != "" {
		return src
	}
	if srcset := attrValue(img, "srcset"); srcset != "" {
		return firstURLFromSrcset(srcset)
	}
	if srcset := attrValue(img, "data-srcset"); srcset != "" {
		return firstURLFromSrcset(srcset)
	}
	return ""
}

func findNodeByClass(node *html.Node, class string) *html.Node {
	var match *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if match != nil {
			return
		}
		if n.Type == html.ElementNode && hasClass(n, class) {
			match = n
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return match
}

func findFirstElement(node *html.Node, tag string) *html.Node {
	var match *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if match != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == tag {
			match = n
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return match
}

func hasClass(node *html.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key != "class" {
			continue
		}
		for _, token := range strings.Fields(attr.Val) {
			if token == class {
				return true
			}
		}
	}
	return false
}

func attrValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func firstURLFromSrcset(value string) string {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}
	first := strings.TrimSpace(parts[0])
	if first == "" {
		return ""
	}
	fields := strings.Fields(first)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func extensionFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	ext := path.Ext(parsed.Path)
	if ext == "" {
		return ""
	}
	return ext
}

func extensionFromContentType(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return ""
	}
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return ""
	}
	return exts[0]
}

func downloadLogo(ctx context.Context, client *http.Client, logoURL, destPath string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/svg+xml,image/*,*/*")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: unexpected status %s", logoURL, resp.Status)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(destPath), "logo-*.tmp")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
	}()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}

	return resp.Header.Get("Content-Type"), nil
}

func loadMapping(path string) (logoMapping, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return logoMapping{
				CodeToSymbol:    make(map[string]string),
				CodeFailures:    make(map[string]string),
				SymbolToFile:    make(map[string]string),
				SymbolToLogoURL: make(map[string]string),
				LogoURLToFile:   make(map[string]string),
			}, nil
		}
		return logoMapping{}, err
	}

	var mapping logoMapping
	if err := json.Unmarshal(content, &mapping); err != nil {
		return logoMapping{}, err
	}

	if mapping.CodeToSymbol == nil {
		mapping.CodeToSymbol = make(map[string]string)
	}
	if mapping.CodeFailures == nil {
		mapping.CodeFailures = make(map[string]string)
	}
	if mapping.SymbolToFile == nil {
		mapping.SymbolToFile = make(map[string]string)
	}
	if mapping.SymbolToLogoURL == nil {
		mapping.SymbolToLogoURL = make(map[string]string)
	}
	if mapping.LogoURLToFile == nil {
		mapping.LogoURLToFile = make(map[string]string)
	}
	return mapping, nil
}

func writeMapping(path string, mapping logoMapping) error {
	payload, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func filterPendingCodes(codes []string, mapping logoMapping, outDir string) []string {
	if len(codes) == 0 {
		return codes
	}
	var pending []string
	for _, code := range codes {
		if _, failed := mapping.CodeFailures[code]; failed {
			continue
		}
		symbol := mapping.CodeToSymbol[code]
		if symbol == "" {
			pending = append(pending, code)
			continue
		}
		fileName := mapping.SymbolToFile[symbol]
		if fileName == "" {
			pending = append(pending, code)
			continue
		}
		if _, err := os.Stat(filepath.Join(outDir, fileName)); err != nil {
			pending = append(pending, code)
		}
	}
	return pending
}
