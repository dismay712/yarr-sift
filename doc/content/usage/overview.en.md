---
title: Overview
weight: 1
---

{{< img src="promo.png" alt="yarr overview" >}}

The main screen has three sections.

1. **Feed list** - Use this section to select a feed or a folder.
2. **Article list** - This section shows the articles from your selection.
3. **Article content** - This section shows the content of the article you select.

Each section has a top bar with actions for that section.

The top bar of the **feed list** contains:

- filters - filter the list by unread, starred, or all articles
- global settings menu - add or refresh feeds, change theme, set auto-refresh rate, choose sorting, import/export subscriptions, view shortcuts, and switch language

The top bar of the **article list** contains:

- search field - search within the current selection
- mark all read - mark all articles in the current selection as read; visible only if the unread filter is set
- settings menu for the selected feed or folder - open the website, rename, move, or delete

The top bar of the **article content** contains actions for the selected article:

- star toggle - mark the article as starred
- read toggle - mark the article as read or unread
- article appearance settings - choose the font and adjust the text size
- read here - fetch and display the full article content
- open in browser - open the article link in a new tab
- left chevron - navigate to the previous article in the list
- right chevron - navigate to the next article in the list
- close - close and unselect the current article

## Scheduled article extraction

1. Select a feed, open the **⋯** menu above its article list, and choose **Extract article text**.
2. Enter a CSS selector for the text you want. For example, use `.entry-content` for a page whose article is inside `<div class="entry-content">`. To skip ads and related stories inside the article, enter `.ad, .related` under **Exclude selector**. Use selectors that match the site's original HTML.
3. Save, then choose **Refresh Feeds** in the feed list menu to try the rule. For recurring updates, set **Auto Refresh** in that menu to an interval such as `30m` and keep Yarr Sift running.

New entries and previously saved entries are processed in batches. Yarr Sift connects directly to public HTTP(S) pages and keeps the text and line breaks from the matching elements in the server's HTML. Pages that require a login or load text with JavaScript may yield no text. When a fetch fails or no text matches, the RSS content stays available and the reader retries on a later refresh.

Each successful match is saved once per rule. Changing a rule queues the feed's entries for extraction again. Clear the content selector and save to turn extraction off and restore the original feed content.
