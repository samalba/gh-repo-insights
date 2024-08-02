package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog/log"
)

type RepoInsights struct {
	ghClient     *github.Client
	repoName     string
	dateRange    string
	cacheResults bool
	cache        *JSONCache
}

func NewRepoInsights(repoName string, dateRange string) (*RepoInsights, error) {
	ghClient := github.NewClient(nil)
	if os.Getenv("GITHUB_PAT") != "" {
		ghClient = ghClient.WithAuthToken(os.Getenv("GITHUB_PAT"))
	}

	cache, err := NewJSONCache("cache", time.Hour*24)
	if err != nil {
		return nil, err
	}

	return &RepoInsights{
		ghClient:     ghClient,
		repoName:     repoName,
		dateRange:    dateRange,
		cacheResults: true,
		cache:        cache,
	}, nil
}

func (r *RepoInsights) SearchIssues(ctx context.Context, query string, since string) ([]*github.Issue, error) {
	fullQuery := fmt.Sprintf("repo:%s created:>%s %s", r.repoName, since, query)

	if r.cacheResults {
		issues, err := r.cache.Load(fullQuery)
		if err == nil {
			return issues, nil
		}
		log.Err(err).Msg("Error loading cache")
	}

	searchOpts := &github.SearchOptions{
		Sort:  "created",
		Order: "asc",
	}

	var issues []*github.Issue
	for {
		result, resp, err := r.ghClient.Search.Issues(ctx, fullQuery, searchOpts)

		if _, ok := err.(*github.RateLimitError); ok {
			log.Err(err).Msg("Rate limited, waiting for a minute...")
			time.Sleep(time.Minute)
			continue
		} else if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		issues = append(issues, result.Issues...)
		if resp.NextPage == 0 {
			break
		}

		searchOpts.Page = resp.NextPage
	}

	if r.cacheResults {
		err := r.cache.Save(fullQuery, issues)
		if err != nil {
			log.Err(err).Msg("Error saving cache")
		}
	}

	return issues, nil
}

func (r *RepoInsights) FilterOut(issues []*github.Issue, filter []string) []*github.Issue {
	filtered := []*github.Issue{}

	for _, issue := range issues {
		title := strings.ToLower(issue.GetTitle())
		found := false
		for _, f := range filter {
			if strings.HasPrefix(title, fmt.Sprintf("%s:", f)) ||
				strings.HasPrefix(title, fmt.Sprintf("%ss:", f)) {
				found = true
				break
			}
		}
		if !found {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func (r *RepoInsights) PrintWeekly(issues []*github.Issue) {
	weekMap := make(map[string][]*github.Issue)

	// Populate the weekMap with issues
	for _, issue := range issues {
		created := issue.GetCreatedAt()
		year, week := created.ISOWeek()
		key := fmt.Sprintf("%d-%02d", year, week)

		if _, ok := weekMap[key]; !ok {
			weekMap[key] = []*github.Issue{}
		}

		weekMap[key] = append(weekMap[key], issue)
	}

	// Print the number of issue created each week
	keys := make([]string, 0, len(weekMap))
	for key := range weekMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	row := []string{}
	for _, key := range keys {
		numIssues := len(weekMap[key])
		row = append(row, fmt.Sprintf("%d", numIssues))
		fmt.Printf("Week %s: %d issues created\n", key, numIssues)
	}

	fmt.Println("--- CSV:")
	cw := csv.NewWriter(os.Stdout)
	_ = cw.Write(keys)
	_ = cw.Write(row)
	cw.Flush()
	fmt.Println("---")
}

// PrintMonthly prints the number of issues created each month
func (r *RepoInsights) PrintMonthly(since string, issues []*github.Issue, printTitles bool) error {
	// Initialize the monthMap (to make sure all months have a value)
	monthMap := make(map[string][]*github.Issue)
	sinceDate, err := time.Parse(time.DateOnly, since)
	if err != nil {
		return err
	}
	now := time.Now()
	for i := sinceDate.Year(); i <= now.Year(); i++ {
		firstMonth := time.January
		lastMonth := time.December
		if i == now.Year() {
			firstMonth = sinceDate.Month()
			lastMonth = now.Month()
		}
		for i := sinceDate.Year(); i <= now.Year(); i++ {
			for j := firstMonth; j <= lastMonth; j++ {
				key := fmt.Sprintf("%d-%02d", i, j)
				monthMap[key] = []*github.Issue{}
			}
		}
	}

	// Populate the monthMap with issues
	for _, issue := range issues {
		created := issue.GetCreatedAt()
		year, month, _ := created.Date()
		key := fmt.Sprintf("%d-%02d", year, month)

		// all month already exist
		// if _, ok := monthMap[key]; !ok {
		// 	monthMap[key] = []*github.Issue{}
		// }

		if printTitles {
			fmt.Println(key, "--", issue.GetTitle())
		}
		monthMap[key] = append(monthMap[key], issue)
	}

	// Print the number of issues created each month
	keys := make([]string, 0, len(monthMap))
	for key := range monthMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	row := []string{}
	for _, key := range keys {
		numIssues := len(monthMap[key])
		row = append(row, fmt.Sprintf("%d", numIssues))
		// fmt.Printf("Month %s: %d issues created\n", key, numIssues)
	}

	fmt.Println("--- CSV:")
	cw := csv.NewWriter(os.Stdout)
	_ = cw.Write(keys)
	_ = cw.Write(row)
	cw.Flush()
	fmt.Println("---")

	return nil
}
