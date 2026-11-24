package tests

import (
	"testing"
	"time"

	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/storage/model"
)

func TestExtractionRestoresRSSAndSearch(t *testing.T) {
	dbtest(t, func(t *testing.T, db storage.Storage) {
		feed := db.CreateFeed(model.CreateFeedParams{FeedLink: "https://example.org/feed"})
		rss := "<p>originalsummary</p>"
		if !db.CreateItems([]model.Item{{
			FeedId: feed.Id, GUID: "article", Link: "https://example.org/article",
			Content: rss, Date: time.Now(), Status: model.STARRED,
		}}) {
			t.Fatal("could not create item")
		}
		filter := model.ItemFilter{FeedID: &feed.Id}
		id := db.ListItems(filter, 1, true, false)[0].Id
		assertBody := func(want string, searchable string, absent string) {
			t.Helper()
			item := db.GetItem(id)
			if item.Content != want || item.Status != model.STARRED {
				t.Fatalf("item content %q, status %v", item.Content, item.Status)
			}
			filter.Search = &searchable
			if len(db.ListItems(filter, 1, true, false)) != 1 {
				t.Errorf("missing search result for %q", searchable)
			}
			filter.Search = &absent
			if len(db.ListItems(filter, 1, true, false)) != 0 {
				t.Errorf("unexpected search result for %q", absent)
			}
		}
		setRule := func(include, exclude string) model.Feed {
			t.Helper()
			if _, err := db.UpdateFeed(feed.Id, model.UpdateFeedParams{
				ContentSelector: &include, ExcludeSelector: &exclude,
			}); err != nil {
				t.Fatal(err)
			}
			rule := *db.GetFeed(feed.Id)
			if rule.ContentSelector != include || rule.ExcludeSelector != exclude ||
				db.ListFeeds()[0].ExtractionRevision != rule.ExtractionRevision {
				t.Fatalf("rules not persisted: %+v", rule)
			}
			return rule
		}
		wantPending := func(rule model.Feed, count int) {
			t.Helper()
			items, err := db.ListPendingExtractions(rule, 4)
			if err != nil || len(items) != count {
				t.Fatalf("pending items = %v, error = %v, want %d", items, err, count)
			}
		}
		save := func(rule model.Feed, content string) {
			t.Helper()
			if err := db.RecordExtraction(rule, id, &content); err != nil {
				t.Fatal(err)
			}
		}

		wantPending(*feed, 0)
		rule := setRule("article", "")
		wantPending(rule, 1)
		first := "<p>firstextraction</p>"
		save(rule, first)
		assertBody(first, "firstextraction", "originalsummary")
		save(rule, "duplicate result")
		assertBody(first, "firstextraction", "originalsummary")
		wantPending(setRule("article", ""), 0)

		revised := setRule("article", ".ad")
		save(rule, "stale result")
		wantPending(rule, 0)
		wantPending(revised, 1)
		assertBody(rss, "originalsummary", "firstextraction")
		second := "<p>secondextraction</p>"
		save(revised, second)
		assertBody(second, "secondextraction", "originalsummary")

		disabled := setRule("", "")
		save(revised, "late result")
		wantPending(disabled, 0)
		assertBody(rss, "originalsummary", "secondextraction")
	})
}

func TestFailedExtractionsYieldToUntouchedItems(t *testing.T) {
	dbtest(t, func(t *testing.T, db storage.Storage) {
		feed := db.CreateFeed(model.CreateFeedParams{FeedLink: "https://example.org/feed", ContentSelector: "article"})
		if !db.CreateItems([]model.Item{
			{FeedId: feed.Id, GUID: "one", Date: time.Now(), Link: "https://example.org/one"},
			{FeedId: feed.Id, GUID: "two", Date: time.Now(), Link: "https://example.org/two", Content: "fallback"},
			{FeedId: feed.Id, GUID: "no-link", Date: time.Now()},
		}) {
			t.Fatal("could not create items")
		}
		next := func() model.Item {
			t.Helper()
			items, err := db.ListPendingExtractions(*feed, 1)
			if err != nil || len(items) != 1 {
				t.Fatalf("expected one pending item, got %v, %v", items, err)
			}
			return items[0]
		}
		failed := next()
		if failed.Link != "https://example.org/two" {
			t.Fatalf("expected newest linked item, got %+v", failed)
		}
		if err := db.RecordExtraction(*feed, failed.Id, nil); err != nil {
			t.Fatal(err)
		}
		if item := db.GetItem(failed.Id); item.Content != "fallback" {
			t.Fatalf("failed extraction changed body: %q", item.Content)
		}
		other := next()
		if other.Id == failed.Id {
			t.Fatal("failed item blocked next item")
		}
		body := "extracted"
		if err := db.RecordExtraction(*feed, other.Id, &body); err != nil {
			t.Fatal(err)
		}
		if next().Id != failed.Id {
			t.Fatal("failed item did not remain eligible for retry")
		}
	})
}
