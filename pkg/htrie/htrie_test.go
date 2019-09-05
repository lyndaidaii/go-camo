// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package htrie

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTrieCheckURL(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|s|example.org|i|*/test.png",
		"||*.example.com||*/test.png",
		"||example.net||*",
		"||foo.example.net||/",
		"||bar.example.net|i|*/test.png",
		"||bar.example.net|i|*/test.png.extra",
	}

	testMatch := []string{
		"http://example.org/foo/test.png",
		"http://example.org/foo/TEST.png",
		"http://bar.example.com/foo/test.png",
		"http://example.net/test.png",
		"http://foo.example.net/",
		"http://bar.example.net/foo/test.png",
		"http://bar.example.net/foo/test.png.extra",
	}

	testNoMatch := []string{
		"http://example.com/foo/test.png",
		"http://example.org/foo/testx.png",
		"http://foo.example.net/nope",
		"http://bar.example.org/foo/testx.png",
		"http://bar.example.net/foo/test.png.ex",
	}

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			t.Errorf("failed to add domain rule: %s", err)
		}
	}

	//fmt.Println(dt.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		assert.True(t, dt.CheckURL(u), fmt.Sprintf("should have matched: %s", u))
	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		assert.False(t, dt.CheckURL(u), fmt.Sprintf("should not have matched: %s", u))
	}
}

func TestHTrieCheckHostname(t *testing.T) {
	t.Parallel()

	rules := []string{
		"|s|localhost||",
		"|s|localdomain||",
	}

	testMatch := []string{
		"http://localhost/foo/test.png",
		"http://foo.localhost/foo/test.png",
		"http://bar.foo.localhost/foo/test.png",
		"http://localdomain/foo/TEST.png",
		"http://foo.localdomain/foo/test.png",
		"http://bar.foo.localdomain/foo/test.png",
	}

	testNoMatch := []string{
		"http://example.com/foo/test.png",
		"http://example.org/foo/testx.png",
		"http://foo.example.net/nope",
		"http://bar.example.org/foo/testx.png",
		"http://bar.example.net/foo/test.png.ex",
	}

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			t.Errorf("failed to add domain rule: %s", err)
		}
	}

	//fmt.Println(dt.RenderTree())

	for _, u := range testMatch {
		u, _ := url.Parse(u)
		assert.True(t, dt.CheckHostname(u), fmt.Sprintf("should have matched: %s", u))
	}
	for _, u := range testNoMatch {
		u, _ := url.Parse(u)
		assert.False(t, dt.CheckHostname(u), fmt.Sprintf("should not have matched: %s", u))
	}
}

func BenchmarkHTrieCreate(b *testing.B) {
	dt := NewURLMatcher()
	urls := []string{
		"||*.example.com||*/test.png",
		"|s|example.org|i|*/test.png",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, u := range urls {
			err := dt.AddRule(u)
			if err != nil {
				b.Errorf("%s", err)
			}
		}
	}
}

func BenchmarkRegexCreate(b *testing.B) {
	urls := []string{
		`^.*\.example.com/.*/test.png`,
		`^(.*\.)?example.org/(?:i.*)/test.png`,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, u := range urls {
			r, err := regexp.Compile(u)
			if err != nil {
				b.Errorf("%s", err)
			}
			_ = r
		}
	}
}

func BenchmarkHTrieMatch(b *testing.B) {
	rules := []string{
		"||foo.example.net||/test.png",
		"||bar.example.net||/test.png",
		"||*.bar.example.net||/test.png",
		"||*.hodor.example.net||/*/test.png",
		"||*.example.com||*/test.png",
		"|s|example.org|i|*/test.png",
	}

	testURLin := []string{
		"http://example.com/foo/test.png",
		"http://bar.example.com/foo/test.png",
		"http://bar.example.com/foo/testx.png",
		"http://bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/test.png",
		"http://bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/testx.png",
	}

	var (
		testIters   = 10
		testURLSize = 10000
	)
	testURLs := make([]string, 0)
	for i := 0; i < testURLSize; i++ {
		testURLs = append(testURLs, testURLin[i%3])
	}

	dt := NewURLMatcher()
	for _, rule := range rules {
		err := dt.AddRule(rule)
		if err != nil {
			fmt.Println(err)
		}
	}

	parsed := make([]*url.URL, 0)
	for _, u := range testURLs {
		u, _ := url.Parse(u)
		parsed = append(parsed, u)

	}

	b.ResetTimer()
	for _, u := range parsed {
		for i := 0; i < testIters; i++ {
			dt.CheckURL(u)
		}
	}
}

func BenchmarkRegexMatch(b *testing.B) {
	rules := []string{
		`^foo.example.net/test.png`,
		`^bar.example.net/test.png`,
		`^.*\.bar.example.net/test.png`,
		`^.*\.hodor.example.net/.*/test.png`,
		`^.*\.example.com/.*/test.png`,
		`^(.*\.)?example.org/(?:i.*/test.png)`,
	}

	testURLin := []string{
		"example.com/foo/test.png",
		"bar.example.com/foo/test.png",
		"bar.example.com/foo/testx.png",
		"bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/test.png",
		"bar.example.com/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/foo/testx.png",
	}

	var (
		testIters   = 10
		testURLSize = 10000
	)
	testURLs := make([]string, 0)
	for i := 0; i < testURLSize; i++ {
		testURLs = append(testURLs, testURLin[i%3])
	}

	rexes := make([]*regexp.Regexp, 0)
	for _, r := range rules {
		rx, _ := regexp.Compile(r)
		rexes = append(rexes, rx)
	}

	b.ResetTimer()

	for _, u := range testURLs {
		for i := 0; i < testIters; i++ {
			// walk regexes in order. first match wins
			for _, rx := range rexes {
				if rx.MatchString(u) {
					break
				}
			}
		}
	}
}