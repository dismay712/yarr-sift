package worker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"time"

	"github.com/nkanaev/yarr/src/content/htmlutil"
	"github.com/nkanaev/yarr/src/content/sift"
	"github.com/nkanaev/yarr/src/storage/model"
	"golang.org/x/net/html/charset"
)

const (
	articlesPerRefresh  = 10
	articleBatchTimeout = time.Minute
	articleTimeout      = 30 * time.Second
	maxArticleBytes     = 4 << 20
)

func (w *Worker) saveItemsAndListPending(feed model.Feed, items []model.Item) ([]model.Item, error) {
	w.itemMu.Lock()
	defer w.itemMu.Unlock()
	if len(items) > 0 && !w.db.CreateItems(items) {
		return nil, errors.New("could not store feed items")
	}
	if feed.ContentSelector == "" {
		return nil, nil
	}
	// Query stored items even when the publisher's feed response was a 304.
	return w.db.ListPendingExtractions(feed, articlesPerRefresh)
}

func (w *Worker) extractArticles(feed model.Feed, pending []model.Item) {
	rule, err := sift.Compile(feed.ContentSelector, feed.ExcludeSelector)
	if err != nil {
		log.Printf("Invalid extraction rule for feed %d: %v", feed.Id, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), articleBatchTimeout)
	defer cancel()
	base := feed.Link
	if !htmlutil.IsAPossibleLink(base) {
		base = feed.FeedLink
	}
	for _, item := range pending {
		if ctx.Err() != nil {
			break
		}
		// A rule can change while another item in this batch is downloading.
		current := w.db.GetFeed(feed.Id)
		if current == nil || current.ExtractionRevision != feed.ExtractionRevision {
			break
		}
		link := item.Link
		if !htmlutil.IsAPossibleLink(link) {
			link = htmlutil.AbsoluteUrl(link, base)
		}
		content, err := fetchArticle(ctx, w.articles, link, rule)
		var result *string
		if err == nil {
			result = &content
		} else {
			log.Printf("Extraction failed for feed %d item %d: %v", feed.Id, item.Id, err)
		}
		w.itemMu.Lock()
		err = w.db.RecordExtraction(feed, item.Id, result)
		w.itemMu.Unlock()
		if err != nil {
			log.Printf("Could not record extraction for feed %d item %d: %v", feed.Id, item.Id, err)
		}
	}
}

func fetchArticle(ctx context.Context, client *http.Client, link string, rule *sift.Rule) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", err
	}
	if err := validateArticleURL(req.URL); err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Yarr-Sift")
	req.Header.Set("Accept", "text/html, application/xhtml+xml")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP status %d", res.StatusCode)
	}
	body := bufio.NewReader(http.MaxBytesReader(nil, res.Body, maxArticleBytes))
	contentType := res.Header.Get("Content-Type")
	detected := contentType
	if detected == "" {
		preview, err := body.Peek(512)
		if err != nil && err != io.EOF {
			return "", err
		}
		detected = http.DetectContentType(preview)
	}
	mediaType, _, err := mime.ParseMediaType(detected)
	if err != nil || (mediaType != "text/html" && mediaType != "application/xhtml+xml") {
		return "", fmt.Errorf("unsupported content type %q", detected)
	}
	reader, err := charset.NewReader(body, contentType)
	if err != nil {
		return "", err
	}
	return rule.Extract(reader)
}
