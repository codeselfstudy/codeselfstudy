package main

import (
	"net/http"
	"strings"
)

// legacyRedirect maps a legacy URL to its new destination and status. Ported
// from the old app's src/data/redirects.ts. Two deliberate upgrades over the
// TypeScript version: targets carry the trailing slash the site now uses (so a
// legacy hit is a single hop, not redirect-then-canonicalization), and the
// caller preserves the query string. "/" targets stay "/".
type legacyRedirect struct {
	to     string
	status int
}

// exactRedirects is matched first, by exact path. All statuses are 308
// (permanent) as in the source data.
var exactRedirects = map[string]legacyRedirect{
	"/book":                              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/book/":                             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/apps":                              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/apps/":                             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/autism":                            {to: "/learn/", status: http.StatusPermanentRedirect},
	"/autism/":                           {to: "/learn/", status: http.StatusPermanentRedirect},
	"/contribute":                        {to: "/learn/", status: http.StatusPermanentRedirect},
	"/contribute/":                       {to: "/learn/", status: http.StatusPermanentRedirect},
	"/sponsors":                          {to: "/credits/", status: http.StatusPermanentRedirect},
	"/sponsors/":                         {to: "/credits/", status: http.StatusPermanentRedirect},
	"/parking":                           {to: "/events/", status: http.StatusPermanentRedirect},
	"/parking/":                          {to: "/events/", status: http.StatusPermanentRedirect},
	"/school":                            {to: "/learn/", status: http.StatusPermanentRedirect},
	"/school/":                           {to: "/learn/", status: http.StatusPermanentRedirect},
	"/support-free-software":             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/support-free-software/":            {to: "/learn/", status: http.StatusPermanentRedirect},
	"/tutorials":                         {to: "/learn/", status: http.StatusPermanentRedirect},
	"/tutorials/":                        {to: "/learn/", status: http.StatusPermanentRedirect},
	"/featured":                          {to: "/learn/", status: http.StatusPermanentRedirect},
	"/featured/":                         {to: "/learn/", status: http.StatusPermanentRedirect},
	"/hire-programmers":                  {to: "/jobs/", status: http.StatusPermanentRedirect},
	"/hire-programmers/":                 {to: "/jobs/", status: http.StatusPermanentRedirect},
	"/wiki":                              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/wiki/":                             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/discounts/algoexpert":              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/discounts/algoexpert/":             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/discounts/digitalocean":            {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/discounts/digitalocean/":           {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/sponsors/digitalocean":             {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/sponsors/digitalocean/":            {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/sponsors/thplibrary":               {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/sponsors/thplibrary/":              {to: "/discounts/", status: http.StatusPermanentRedirect},
	"/home":                              {to: "/", status: http.StatusPermanentRedirect},
	"/home/":                             {to: "/", status: http.StatusPermanentRedirect},
	"/index.html":                        {to: "/", status: http.StatusPermanentRedirect},
	"/support-us/":                       {to: "/", status: http.StatusPermanentRedirect},
	"/support/":                          {to: "/", status: http.StatusPermanentRedirect},
	"/support":                           {to: "/", status: http.StatusPermanentRedirect},
	"/presentations":                     {to: "/events/", status: http.StatusPermanentRedirect},
	"/presentations/":                    {to: "/events/", status: http.StatusPermanentRedirect},
	"/presentations-schedule":            {to: "/events/", status: http.StatusPermanentRedirect},
	"/c":                                 {to: "/forum/", status: http.StatusPermanentRedirect},
	"/c/":                                {to: "/forum/", status: http.StatusPermanentRedirect},
	"/bbs":                               {to: "/forum/", status: http.StatusPermanentRedirect},
	"/join":                              {to: "/forum/", status: http.StatusPermanentRedirect},
	"/welcome":                           {to: "/events/", status: http.StatusPermanentRedirect},
	"/checkin":                           {to: "/forum/", status: http.StatusPermanentRedirect},
	"/edu":                               {to: "/learn/", status: http.StatusPermanentRedirect},
	"/edu/":                              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/comment/63":                        {to: "/", status: http.StatusPermanentRedirect},
	"/forms/feedback":                    {to: "/contact/", status: http.StatusPermanentRedirect},
	"/wiki/Main_Page":                    {to: "/learn/", status: http.StatusPermanentRedirect},
	"/w":                                 {to: "/learn/", status: http.StatusPermanentRedirect},
	"/w/":                                {to: "/learn/", status: http.StatusPermanentRedirect},
	"/programming-notes-wiki/":           {to: "/learn/", status: http.StatusPermanentRedirect},
	"/python-web-scraping-selenium.html": {to: "/learn/", status: http.StatusPermanentRedirect},
	"/Array":                             {to: "/learn/", status: http.StatusPermanentRedirect},
	// The encoded form from the source; a decoded /wiki/C++ request is also
	// covered by the /wiki/* wildcard below.
	"/wiki/C%2B%2B":   {to: "/learn/", status: http.StatusPermanentRedirect},
	"/b/page/6/":      {to: "/learn/", status: http.StatusPermanentRedirect},
	"/b":              {to: "/learn/", status: http.StatusPermanentRedirect},
	"/b/":             {to: "/learn/", status: http.StatusPermanentRedirect},
	"/tools/unicode":  {to: "/tools/", status: http.StatusPermanentRedirect},
	"/tools/unicode/": {to: "/tools/", status: http.StatusPermanentRedirect},
	// These two legacy paths exist only in trailing-slash form in the source
	// data. The caller looks up the slash-normalized path (filepath.Clean
	// strips the trailing slash), and neither is covered by a wildcard, so each
	// needs a bare-form key to match.
	"/support-us":             {to: "/", status: http.StatusPermanentRedirect},
	"/programming-notes-wiki": {to: "/learn/", status: http.StatusPermanentRedirect},
}

// wildcardRedirects is matched only after exactRedirects. A pattern "/blog/*"
// in the source becomes prefix "/blog"; it matches any path under prefix + "/"
// with a non-empty remainder (so "/blog/x" matches but "/blog/" and "/blogpost"
// do not) — identical to the old findRedirect semantics.
var wildcardRedirects = []struct {
	prefix   string
	redirect legacyRedirect
}{
	{prefix: "/blog", redirect: legacyRedirect{to: "/learn/", status: http.StatusPermanentRedirect}},
	{prefix: "/wiki", redirect: legacyRedirect{to: "/learn/", status: http.StatusPermanentRedirect}},
	{prefix: "/b", redirect: legacyRedirect{to: "/learn/", status: http.StatusPermanentRedirect}},
}

// findRedirect looks up a legacy redirect for pathname: exact match first, then
// wildcard prefixes. Mirrors src/lib/redirects.ts.
func findRedirect(pathname string) (legacyRedirect, bool) {
	if r, ok := exactRedirects[pathname]; ok {
		return r, true
	}
	for _, w := range wildcardRedirects {
		if strings.HasPrefix(pathname, w.prefix+"/") && len(pathname) > len(w.prefix)+1 {
			return w.redirect, true
		}
	}
	return legacyRedirect{}, false
}
