import { useState } from "preact/hooks";
import api, { HTTPError, NetworkError } from "../api";
import type { ArticleRule, Feed } from "../api-types";
import { useI18n } from "../i18n";
import { to } from "../utils";

interface ArticleRuleFormProps {
  feed: Feed;
  onCancel: () => void;
  onSaved: (rule: ArticleRule) => void;
}

export default function ArticleRuleForm({ feed, onCancel, onSaved }: ArticleRuleFormProps) {
  const { t } = useI18n();
  const [content, setContent] = useState(feed.content_selector);
  const [exclude, setExclude] = useState(feed.exclude_selector);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function save(event: Event) {
    event.preventDefault();
    if (saving) return;
    const contentSelector = content.trim();
    const rule = {
      content_selector: contentSelector,
      exclude_selector: contentSelector ? exclude.trim() : "",
    };

    setError("");
    setSaving(true);
    const [err] = await to(api.feeds.update(feed.id, rule));
    setSaving(false);
    if (err) {
      if (err instanceof HTTPError) {
        setError(t("error_server", { code: err.status, text: err.message }));
      } else {
        setError(t(err instanceof NetworkError ? "error_network" : "fail_save_feed"));
      }
      return;
    }
    onSaved(rule);
  }

  return (
    <form className="d-flex flex-column gap-2" onSubmit={save} aria-busy={saving}>
      <h2 className="fs-5 mb-0">{t("article_extraction")}</h2>
      <p className="mb-0 text-break">{feed.title}</p>
      <p id="article-rule-help" className="small mb-1">
        {t("article_rule_help")}
      </p>
      <label htmlFor="content-selector">{t("content_selector")}</label>
      <input
        id="content-selector"
        className="c-input"
        disabled={saving}
        autoFocus
        spellcheck={false}
        autoComplete="off"
        aria-describedby="article-rule-help"
        placeholder="article .entry-content"
        value={content}
        onInput={e => setContent(e.currentTarget.value)}
      />
      <label htmlFor="exclude-selector">{t("exclude_selector")}</label>
      <input
        id="exclude-selector"
        className="c-input"
        spellcheck={false}
        autoComplete="off"
        placeholder=".ad, .related"
        value={exclude}
        disabled={saving || !content.trim()}
        onInput={e => setExclude(e.currentTarget.value)}
      />
      {error && (
        <p role="alert" className="mb-0 text-danger text-break">
          {error}
        </p>
      )}
      <div className="d-flex justify-content-end gap-2 mt-2">
        <button type="button" className="c-button is-link" onClick={onCancel}>
          {t("cancel")}
        </button>
        <button type="submit" className="c-button" disabled={saving}>
          {t("save")}
        </button>
      </div>
    </form>
  );
}
