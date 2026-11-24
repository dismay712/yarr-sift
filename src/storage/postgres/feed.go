package postgres

import (
	"database/sql"
	"log"

	"github.com/nkanaev/yarr/src/storage/model"
)

func (s *PostgresStorage) CreateFeed(params model.CreateFeedParams) *model.Feed {
	title := params.Title
	if title == "" {
		title = params.FeedLink
	}
	row := s.db.QueryRow(`
		insert into feeds (title, description, link, feed_link, folder_id, content_selector, exclude_selector)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (feed_link) do update set folder_id = $5
		returning id, folder_id, title, coalesce(description, ''), coalesce(link, ''), feed_link, icon,
			content_selector, exclude_selector, extraction_revision`,
		title,
		params.Description,
		params.Link,
		params.FeedLink,
		params.FolderID,
		params.ContentSelector,
		params.ExcludeSelector,
	)

	var f model.Feed
	err := row.Scan(&f.Id, &f.FolderId, &f.Title, &f.Description, &f.Link, &f.FeedLink,
		&f.Icon, &f.ContentSelector, &f.ExcludeSelector, &f.ExtractionRevision)
	if err != nil {
		log.Print(err)
		return nil
	}
	return &f
}

func (s *PostgresStorage) DeleteFeed(feedId int64) bool {
	result, err := s.db.Exec(`delete from feeds where id = $1`, feedId)
	if err != nil {
		log.Print(err)
		return false
	}
	nrows, err := result.RowsAffected()
	if err != nil {
		log.Print(err)
		return false
	}
	return nrows == 1
}

func (s *PostgresStorage) UpdateFeed(feedId int64, params model.UpdateFeedParams) (bool, error) {
	_, err := s.db.Exec(`
		update feeds set
			extraction_revision = case
				when coalesce($8, content_selector) != content_selector
				  or coalesce($9, exclude_selector) != exclude_selector
				then extraction_revision + 1 else extraction_revision end,
			content_selector = coalesce($8, content_selector),
			exclude_selector = coalesce($9, exclude_selector),
			title     = coalesce($2, title),
			feed_link = coalesce($3, feed_link),
			folder_id = case when $4 then $5 else folder_id end,
			icon      = case when $6 then $7 else icon end
		where id = $1
	`,
		feedId,
		params.Title,
		params.FeedLink,
		params.FolderID.Set,
		params.FolderID.Value,
		params.Icon.Set,
		params.Icon.Value,
		params.ContentSelector,
		params.ExcludeSelector,
	)
	if err != nil {
		log.Print(err)
		return false, err
	}
	return true, nil
}

func (s *PostgresStorage) ListFeeds() []model.Feed {
	result := make([]model.Feed, 0)
	rows, err := s.db.Query(`
		select id, folder_id, title, coalesce(description, ''), coalesce(link, ''), feed_link, icon,
			content_selector, exclude_selector, extraction_revision
		from feeds
		order by lower(title)
	`)
	if err != nil {
		log.Print(err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var f model.Feed
		err = rows.Scan(
			&f.Id,
			&f.FolderId,
			&f.Title,
			&f.Description,
			&f.Link,
			&f.FeedLink,
			&f.Icon,
			&f.ContentSelector,
			&f.ExcludeSelector,
			&f.ExtractionRevision,
		)
		if err != nil {
			log.Print(err)
			return result
		}
		result = append(result, f)
	}
	return result
}

func (s *PostgresStorage) GetFeed(id int64) *model.Feed {
	var f model.Feed
	err := s.db.QueryRow(`
		select
			id, folder_id, title, coalesce(description, ''), coalesce(link, ''), feed_link,
			icon, content_selector, exclude_selector, extraction_revision
		from feeds where id = $1
	`, id).Scan(
		&f.Id, &f.FolderId, &f.Title, &f.Description, &f.Link, &f.FeedLink,
		&f.Icon, &f.ContentSelector, &f.ExcludeSelector, &f.ExtractionRevision,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return nil
	}
	return &f
}
