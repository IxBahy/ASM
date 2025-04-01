package katana

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
	"github.com/projectdiscovery/katana/pkg/engine/standard"
	"github.com/projectdiscovery/katana/pkg/output"
	"github.com/projectdiscovery/katana/pkg/types"
)

type KatanaScanner struct {
	*scanners.BaseScanner
}

type CrawlResult struct {
	Domain      string      `json:"domain"`
	URLs        []URLResult `json:"urls"`
	Count       int         `json:"count"`
	TimeStamp   string      `json:"timestamp"`
	ElapsedTime string      `json:"elapsed_time"`
}

type URLResult struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Depth       int               `json:"depth"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

func NewKatanaScanner() *KatanaScanner {
	config := scanners.ScannerConfig{
		Name:           "katana",
		Version:        "embedded",
		ExecutablePath: "",
		Base_Command:   "",
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: true,
			Version:   "1.0.0",
		},
	}

	s := &KatanaScanner{
		BaseScanner: base,
	}
	return s
}

func (s *KatanaScanner) Setup() error {
	return nil
}

func (s *KatanaScanner) IsInstalled() bool {
	return true
}

func (s *KatanaScanner) RegisterInstallationStats() error {
	return nil
}

func (s *KatanaScanner) Scan(target string) (scanners.ScannerResult, error) {
	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	target = normalizeURL(target)
	domain := extractor.ExtractDomain(target)

	var crawledURLs []URLResult

	startTime := time.Now()

	options := &types.Options{
		MaxDepth:     3,
		FieldScope:   "rdn",
		BodyReadSize: math.MaxInt,
		Timeout:      10,
		Concurrency:  1,
		Parallelism:  1,
		Delay:        0,
		RateLimit:    50,
		Strategy:     "depth-first",

		OnResult: func(crawlResult output.Result) {

			params := make(map[string]string)
			if crawlResult.Request.URL != "" {
				if parts := strings.Split(crawlResult.Request.URL, "?"); len(parts) > 1 {
					for _, param := range strings.Split(parts[1], "&") {
						kv := strings.SplitN(param, "=", 2)
						if len(kv) == 2 {
							params[kv[0]] = kv[1]
						} else if len(kv) == 1 {
							params[kv[0]] = ""
						}
					}
				}
			}

			urlResult := URLResult{
				URL:         crawlResult.Request.URL,
				Method:      crawlResult.Request.Method,
				StatusCode:  crawlResult.Response.StatusCode,
				ContentType: extractContentType(crawlResult.Response.Headers),
				Depth:       crawlResult.Request.Depth,
				Parameters:  params,
			}

			crawledURLs = append(crawledURLs, urlResult)
		},
	}

	crawlerOptions, err := types.NewCrawlerOptions(options)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create crawler options: %v", err))
		return result, err
	}
	defer crawlerOptions.Close()

	crawler, err := standard.New(crawlerOptions)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create crawler: %v", err))
		return result, err
	}
	defer crawler.Close()

	err = crawler.Crawl(target)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("crawling error: %v", err))

	}

	elapsedTime := time.Since(startTime)

	crawlResult := CrawlResult{
		Domain:      domain,
		URLs:        crawledURLs,
		Count:       len(crawledURLs),
		TimeStamp:   time.Now().Format(time.RFC3339),
		ElapsedTime: elapsedTime.String(),
	}

	jsonData, err := json.MarshalIndent(crawlResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(jsonData))
	return result, nil
}

func normalizeURL(url string) string {

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}

func extractContentType(headers map[string]string) string {
	if ct, ok := headers["Content-Type"]; ok {
		return ct
	}
	if ct, ok := headers["content-type"]; ok {
		return ct
	}
	return ""
}
