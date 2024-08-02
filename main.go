package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	repoInsights, err := NewRepoInsights("dagger/dagger", "created:>2024-01-01")
	if err != nil {
		log.Err(err).Msg("Error creating repo insights")
		return
	}

	// FIXME: bugs aren't well labeled
	// need another filter to be relevant
	//
	// query := "is:issue label:\"kind/bug\""
	// fmt.Println("===> Query:", query)
	// issues, err := repoInsights.SearchIssues(ctx, query, "2024-01-01")
	// if err != nil {
	// 	log.Err(err).Msg("Error searching issues")
	// 	return
	// }
	// repoInsights.PrintWeekly(issues)
	// repoInsights.PrintMonthly(issues)

	since := "2024-01-01"
	filter := []string{"doc", "ci", "chore"}

	// Query all PRs merged
	query := "is:pr is:merged"
	fmt.Println("===> Query:", query)
	issues, err := repoInsights.SearchIssues(ctx, query, since)
	if err != nil {
		log.Err(err).Msg("Error searching issues")
		return
	}

	issues = repoInsights.FilterOut(issues, filter)
	if err := repoInsights.PrintMonthly(since, issues, false); err != nil {
		log.Err(err).Msg("Error printing monthly")
		return
	}

	// Query all PRs merged with "fix" in title
	query = "is:pr \"fix\" in:title is:merged"
	fmt.Println("===> Query:", query)
	issues, err = repoInsights.SearchIssues(ctx, query, since)
	if err != nil {
		log.Err(err).Msg("Error searching issues")
		return
	}

	issues = repoInsights.FilterOut(issues, filter)
	// repoInsights.PrintWeekly(issues)
	if err := repoInsights.PrintMonthly(since, issues, true); err != nil {
		log.Err(err).Msg("Error printing monthly")
		return
	}
}
