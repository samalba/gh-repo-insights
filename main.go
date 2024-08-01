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

	fmt.Println("===> Query: is:issue label:\"kind/bug\"")
	issues, err := repoInsights.SearchIssues(ctx, "is:issue label:\"kind/bug\"", "2024-01-01")
	if err != nil {
		log.Err(err).Msg("Error searching issues")
		return
	}
	repoInsights.PrintWeekly(issues)
	repoInsights.PrintMonthly(issues)

	fmt.Println("===> Query: is:pr \"fix\" in:title is:merged")
	issues, err = repoInsights.SearchIssues(ctx, "is:pr \"fix\" in:title is:merged", "2024-01-01")
	if err != nil {
		log.Err(err).Msg("Error searching issues")
		return
	}
	repoInsights.PrintWeekly(issues)
	repoInsights.PrintMonthly(issues)
}
