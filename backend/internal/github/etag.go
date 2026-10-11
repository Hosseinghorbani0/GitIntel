package github

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	githubETagCacheCapacity = 128
	githubETagCacheMaxBody  = 512 * 1024
	githubETagCacheTTL      = 5 * time.Minute
)

type etagCacheEntry struct {
	key       string
	etag      string
	body      []byte
	header    http.Header
	expiresAt time.Time
}

type githubETagCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	lru     *list.List
}

var sharedGitHubETagCache = githubETagCache{
	entries: make(map[string]*list.Element),
	lru:     list.New(),
}

type etagTransport struct {
	base        http.RoundTripper
	tokenDigest string
}

type replayBody struct {
	reader io.Reader
	body   io.Closer
}

func (b *replayBody) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *replayBody) Close() error {
	return b.body.Close()
}

func withETagCache(client *http.Client, token string) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	clientCopy := *client
	base := clientCopy.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	tokenDigest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	clientCopy.Transport = &etagTransport{base: base, tokenDigest: hex.EncodeToString(tokenDigest[:])}
	return &clientCopy
}

func (t *etagTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet {
		return t.base.RoundTrip(request)
	}
	if strings.HasSuffix(request.URL.Path, "/rate_limit") {
		return t.base.RoundTrip(request)
	}

	key := t.tokenDigest + "|" + request.URL.String() + "|" +
		request.Header.Get("Accept") + "|" + request.Header.Get("X-GitHub-Api-Version")
	entry := sharedGitHubETagCache.get(key)
	if entry != nil && entry.etag != "" {
		request = request.Clone(request.Context())
		request.Header = request.Header.Clone()
		request.Header.Set("If-None-Match", entry.etag)
	}

	response, err := t.base.RoundTrip(request)
	if err != nil || response == nil {
		return response, err
	}
	if response.StatusCode == http.StatusNotModified && entry != nil {
		response.Body.Close()
		cached := *response
		cached.StatusCode = http.StatusOK
		cached.Status = "200 OK"
		cached.Header = entry.header.Clone()
		cached.Body = io.NopCloser(bytes.NewReader(entry.body))
		cached.ContentLength = int64(len(entry.body))
		return &cached, nil
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		sharedGitHubETagCache.delete(key)
		return response, nil
	}

	etag := response.Header.Get("ETag")
	if etag == "" {
		sharedGitHubETagCache.delete(key)
		return response, nil
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, githubETagCacheMaxBody+1))
	if err != nil {
		response.Body.Close()
		return nil, err
	}
	if len(body) > githubETagCacheMaxBody {
		response.Body = &replayBody{
			reader: io.MultiReader(bytes.NewReader(body), response.Body),
			body:   response.Body,
		}
		return response, nil
	}
	response.Body.Close()
	sharedGitHubETagCache.put(key, etag, body, response.Header)
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return response, nil
}

func (c *githubETagCache) get(key string) *etagCacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.entries[key]
	if !ok {
		return nil
	}
	entry := element.Value.(*etagCacheEntry)
	if time.Now().After(entry.expiresAt) {
		delete(c.entries, key)
		c.lru.Remove(element)
		return nil
	}
	c.lru.MoveToFront(element)
	copyEntry := *entry
	copyEntry.body = append([]byte(nil), entry.body...)
	copyEntry.header = entry.header.Clone()
	return &copyEntry
}

func (c *githubETagCache) put(key, etag string, body []byte, header http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.entries[key]; ok {
		entry := element.Value.(*etagCacheEntry)
		entry.etag = etag
		entry.body = append(entry.body[:0], body...)
		entry.header = header.Clone()
		entry.expiresAt = time.Now().Add(githubETagCacheTTL)
		c.lru.MoveToFront(element)
		return
	}
	entry := &etagCacheEntry{
		key:       key,
		etag:      etag,
		body:      append([]byte(nil), body...),
		header:    header.Clone(),
		expiresAt: time.Now().Add(githubETagCacheTTL),
	}
	c.entries[key] = c.lru.PushFront(entry)
	if c.lru.Len() > githubETagCacheCapacity {
		oldest := c.lru.Back()
		delete(c.entries, oldest.Value.(*etagCacheEntry).key)
		c.lru.Remove(oldest)
	}
}

func (c *githubETagCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.entries[key]; ok {
		delete(c.entries, key)
		c.lru.Remove(element)
	}
}
