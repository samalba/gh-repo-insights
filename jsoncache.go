package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog/log"
)

type JSONCache struct {
	path string
	ttl  time.Duration
}

type CacheData struct {
	Created time.Time
	TTL     time.Time
	Query   string
	Issues  []*github.Issue
}

func NewJSONCache(path string, ttl time.Duration) (*JSONCache, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	return &JSONCache{
		path: path,
		ttl:  ttl,
	}, nil
}

func (j *JSONCache) Save(cacheKey string, issues []*github.Issue) error {
	digestKey := fmt.Sprintf("%x", sha256.Sum256([]byte(cacheKey)))
	filePath := fmt.Sprintf("%s/%s.json", j.path, digestKey)

	log.Debug().Str("filePath", filePath).Msg("Creating cache file")
	fp, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fp.Close()

	cacheData := &CacheData{
		Created: time.Now(),
		TTL:     time.Now().Add(j.ttl),
		Query:   cacheKey,
		Issues:  issues,
	}
	e := json.NewEncoder(fp)
	if err := e.Encode(cacheData); err != nil {
		return err
	}
	log.Debug().Int("len-issues", len(issues)).Dur("TTL", j.ttl).Str("query", cacheKey).Msg("Cached data")

	return nil
}

func (j *JSONCache) Load(cacheKey string) ([]*github.Issue, error) {
	digestKey := fmt.Sprintf("%x", sha256.Sum256([]byte(cacheKey)))
	filePath := fmt.Sprintf("%s/%s.json", j.path, digestKey)

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		log.Err(err).Str("filePath", filePath).Msg("Cache does not exist")
		return nil, err
	}

	fp, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	data := &CacheData{}
	d := json.NewDecoder(fp)
	if err := d.Decode(data); err != nil {
		return nil, err
	}
	log.Debug().Int("len-issues", len(data.Issues)).Str("query", cacheKey).Str("filePath", filePath).Msg("Loaded from cache")

	if time.Now().After(data.TTL) {
		os.Remove(filePath)
		log.Debug().Dur("TTL", j.ttl).Dur("since", time.Since(data.Created)).Str("query", cacheKey).Msg("Cache expired, removed file")
		return nil, errors.New("Cache expired")
	}

	return data.Issues, nil
}
