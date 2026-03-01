package page

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

const (
	customElementTemplate = `<%s data-content-id="%s" data-app-name="%s" data-app-generation="%s"%s></%s>`
	navigationTemplate    = `<nav id="navigation-%s">
<script type="text/javascript">
fetch('/-/navigation/%s/[[ .Language ]]%s')
  .then(response => response.text())
  .then(data => {
    document.getElementById('navigation-%s').innerHTML = data;
  });
</script>
</nav>`
	rawHTMLTemplate = `<div data-content-id="%s">%s</div>`
)

func (p *PageHandler) CacheKey(l language.Tag) string {
	return fmt.Sprintf("%s:%s:%s", p.Name, p.Checksum(), l.String())
}

func (p *PageHandler) Checksum() string {
	if p.checksum != "" {
		return p.checksum
	}

	// Generate a checksum based on the p.Status
	// This is used to invalidate the cache when the host changes
	// We use the p.Status to ensure that the cache is invalidated when the host changes

	if p.Status == nil {
		return ""
	}

	// 1. Extract the keys
	keys := make([]string, 0, len(p.Status.Attributes))
	for k := range p.Status.Attributes {
		keys = append(keys, k)
	}

	// 2. Sort the keys alphabetically
	sort.Strings(keys)

	// 3. Create a hash object
	h := sha256.New()

	// 4. Write the status.ObservedGeneration
	h.Write([]byte(strconv.FormatInt(p.Status.ObservedGeneration, 10)))

	// 5. Write key-value pairs in the sorted order
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(p.Status.Attributes[k]))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *PageHandler) ContentToHTMLMap() map[string]string {
	items := map[string]string{}

	for slot, content := range p.Content {
		items[slot] = content.ToHTML(slot)
	}

	return items
}

func (p PageHandler) NavigationToHTMLMap() map[string]string {
	items := map[string]string{}

	for navKey := range p.Navigations {
		items[navKey] = fmt.Sprintf(navigationTemplate, navKey, navKey, p.BasePath(), navKey)
	}

	return items
}

func (p PageHandler) BasePath() string {
	if p.Page == nil {
		return ""
	}
	return p.Page.BasePath
}

func (p PageHandler) Label() string {
	if p.Page == nil {
		return ""
	}
	return p.Page.Label
}

func (p PageHandler) PatternPath() string {
	if p.Page == nil {
		return ""
	}
	return p.Page.PatternPath
}

func (r *PackedContent) ToHTML(slot string) string {
	if r.Content != "" {
		return fmt.Sprintf(rawHTMLTemplate, slot, r.Content)
	}

	var attributes strings.Builder
	for k, v := range r.Attributes {
		fmt.Fprintf(&attributes, ` %s="%s"`, k, v)
	}

	return fmt.Sprintf(customElementTemplate, r.CustomElementName, slot, r.AppName, r.AppGeneration, attributes.String(), r.CustomElementName)
}
