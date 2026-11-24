package worker

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nkanaev/yarr/src/content/sift"
	"github.com/nkanaev/yarr/src/storage/model"
)

type articleTestTransport func(*http.Request) *http.Response

func (rt articleTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r), nil
}

func testResponse(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRefreshExtractsNewAndStoredArticles(t *testing.T) {
	db := newTestDB(t)
	feed := db.CreateFeed(model.CreateFeedParams{Title: "Source", FeedLink: "https://source.example/feed"})
	// An empty icon marks the favicon check as already complete.
	db.UpdateFeed(feed.Id, model.UpdateFeedParams{Icon: model.SetNullable(&emptyIcon)})
	const itemA = "<item><guid>a</guid><link>https://source.example/a</link><title>A</title><description>Summary A</description></item>"
	const itemB = "<item><guid>b</guid><link>https://source.example/b</link><title>B</title><description>Summary B</description></item>"
	feedItems := itemA
	feedClient := &http.Client{Transport: articleTestTransport(func(*http.Request) *http.Response {
		if feedItems == "" {
			return testResponse(304, "", "")
		}
		return testResponse(200, "application/rss+xml", "<rss version=\"2.0\"><channel><title>Source</title>"+feedItems+"</channel></rss>")
	})}
	w := NewWorker(db, feedClient)
	requests := map[string]int{}
	w.articles = &http.Client{Transport: articleTestTransport(func(req *http.Request) *http.Response {
		path := req.URL.Path
		requests[path]++
		if path == "/a" && requests[path] == 1 {
			return testResponse(503, "text/html", "unavailable")
		}
		return testResponse(200, "text/html; charset=windows-1252", "<main>caf\xe9 "+path+"<aside>banner</aside></main>")
	})}
	refresh := func() { w.refreshFeed(*db.GetFeed(feed.Id)) }

	refresh()
	stored := db.ListItems(model.ItemFilter{FeedID: &feed.Id}, 10, false, true)
	if len(stored) != 1 || len(requests) != 0 {
		t.Fatalf("disabled rule: items=%v, article requests=%v", stored, requests)
	}
	firstID := stored[0].Id
	original := stored[0].Content
	starred := model.STARRED
	db.UpdateItem(firstID, model.UpdateItemParams{Status: &starred})

	include, exclude := "main", "aside"
	if ok, err := db.UpdateFeed(feed.Id, model.UpdateFeedParams{ContentSelector: &include, ExcludeSelector: &exclude}); err != nil || !ok {
		t.Fatalf("enable extraction: %v", err)
	}
	feedItems = "" // Existing items can be extracted while the RSS feed is unchanged.
	refresh()
	if got := db.GetItem(firstID); got.Content != original || got.Status != starred || requests["/a"] != 1 {
		t.Fatalf("failed article request: item=%+v, requests=%v", got, requests)
	}
	refresh()
	if got := db.GetItem(firstID); got.Content != "café /a" || got.Status != starred {
		t.Fatalf("retry on 304 returned %+v", got)
	}
	feedItems = itemA + itemB
	refresh()
	refresh()
	stored = db.ListItems(model.ItemFilter{FeedID: &feed.Id}, 10, false, true)
	if len(stored) != 2 || requests["/a"] != 2 || requests["/b"] != 1 {
		t.Fatalf("repeat refresh: items=%v, requests=%v", stored, requests)
	}
	for _, item := range stored {
		if item.Title == "B" && item.Content != "café /b" {
			t.Fatalf("new article content = %q", item.Content)
		}
	}
}

func TestArticleAddressPolicy(t *testing.T) {
	for _, address := range []string{"192.168.11.22:80", "127.0.0.1:80", "169.254.1.2:443", "[::1]:80", "[::ffff:10.0.0.2]:80", "[fc00::1]:80", "[fe80::1%eth0]:443", "100.64.0.1:443", "localhost:80"} {
		if err := checkArticleAddress("tcp", address, nil); err == nil {
			t.Errorf("accepted %s", address)
		}
	}
	for _, address := range []string{"1.1.1.1:443", "[2606:4700:4700::1111]:443"} {
		if err := checkArticleAddress("tcp", address, nil); err != nil {
			t.Errorf("rejected %s: %v", address, err)
		}
	}
}

func TestFetchArticleLimits(t *testing.T) {
	rule, err := sift.Compile("main", "")
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: articleTestTransport(func(*http.Request) *http.Response {
		return testResponse(200, "text/html", "<main>"+strings.Repeat("x", maxArticleBytes)+"</main>")
	})}
	_, err = fetchArticle(context.Background(), client, "https://source.example/item", rule)
	var maxErr *http.MaxBytesError
	if !errors.As(err, &maxErr) {
		t.Errorf("oversized article: %v", err)
	}
	client.Transport = articleTestTransport(func(*http.Request) *http.Response {
		return testResponse(200, "application/octet-stream", "<main>not an HTML page</main>")
	})
	if _, err := fetchArticle(context.Background(), client, "https://source.example/item", rule); err == nil {
		t.Error("accepted non-HTML media type")
	}
}

func TestRelativeArticleLinksAndStoppingDisabledBatch(t *testing.T) {
	const feedURL = "https://rss.example/feeds/news.xml"
	for _, tc := range []struct{ site, want string }{
		{"https://site.example/posts/", "https://site.example/posts/article"},
		{"", "https://rss.example/feeds/article"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			db := newTestDB(t)
			feed := db.CreateFeed(model.CreateFeedParams{
				Title: "Source", FeedLink: feedURL, Link: tc.site, ContentSelector: "main",
			})
			if !db.CreateItems([]model.Item{
				{FeedId: feed.Id, GUID: "one", Link: "article", Content: "summary"},
				{FeedId: feed.Id, GUID: "two", Link: "article", Content: "summary"},
			}) {
				t.Fatal("could not insert historical items")
			}
			w := NewWorker(db, &http.Client{Transport: articleTestTransport(func(*http.Request) *http.Response {
				return testResponse(304, "", "")
			})})
			n := 0
			w.articles.Transport = articleTestTransport(func(r *http.Request) *http.Response {
				n++
				if got := r.URL.String(); got != tc.want {
					t.Errorf("relative link resolved to %q, want %q", got, tc.want)
				}
				empty := ""
				if ok, err := db.UpdateFeed(feed.Id, model.UpdateFeedParams{ContentSelector: &empty}); err != nil || !ok {
					t.Fatalf("disable extraction: %v", err)
				}
				return testResponse(200, "text/html", "<main>obsolete content</main>")
			})
			w.refreshFeed(*feed)
			if n != 1 {
				t.Fatalf("requests after rule disabled mid-batch: got %d total", n)
			}
			for _, item := range db.ListItems(model.ItemFilter{FeedID: &feed.Id}, 10, false, true) {
				if item.Content != "summary" {
					t.Errorf("stale result stored for item %d", item.Id)
				}
			}
		})
	}
}
