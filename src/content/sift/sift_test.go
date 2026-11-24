package sift

import (
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	page := `<!doctype html><head><title>Page title</title></head><body>
		<div class="ad"><article>Outside advertisement</article></div>
		<article><h2>第一<em>段</em></h2>
		<p>Hello <b>world</b>!   &lt;img src=x&gt; &amp; more.</p>
		<div class="ad"><p>Inside advertisement</p></div>
		<p style="color: red">Next<br>line</p>
		<pre>One
Two</pre>
		<p hidden>Hide me</p><aside aria-hidden="true">Hidden</aside>
		<script>ignore()</script><style>body {}</style><template>Example</template>
		</article><section id="extra"><p data-kind="keep">Last paragraph</p>
		<p>Ignored</p></section></body>`
	rule, err := Compile("article, article p, #extra > p[data-kind='keep']", ".ad, .sponsor")
	if err != nil {
		t.Fatal(err)
	}
	content, err := rule.Extract(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"第一段", "Hello world! &lt;img src=x&gt; &amp; more.",
		"Next", "line", "One", "Two", "Last paragraph",
	}, "<br>\n")
	if content != want {
		t.Errorf("got %q, want %q", content, want)
	}
}

func TestCompileInvalidRules(t *testing.T) {
	for _, pair := range [][2]string{{"", ""}, {"main[", ""}, {"article", ".ad,"}} {
		if _, err := Compile(pair[0], pair[1]); err == nil {
			t.Errorf("accepted invalid selectors %q", pair)
		}
	}
}

func TestEmptyResult(t *testing.T) {
	rule, err := Compile("article p", ".ad")
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range []string{
		"<p>not under article</p>",
		"<article><p>  </p></article>",
		"<article><p class='ad'>ad</p></article>",
	} {
		if content, err := rule.Extract(strings.NewReader(page)); err == nil || content != "" {
			t.Errorf("expected empty result error, got %q, %v", content, err)
		}
	}
}
