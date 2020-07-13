// Demonstrates how to use a BigQuery aggregator to aggregate match data from
// the Riot API into a queryable table.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/datastore"
	"github.com/ChoiKyubum/riot/analytics/data_aggregation"
	"github.com/ChoiKyubum/riot/analytics/data_aggregation/bigquery_aggregator"
	"github.com/ChoiKyubum/riot/apiclient"
	"github.com/ChoiKyubum/riot/cachedclient"
	"github.com/ChoiKyubum/riot/cachedclient/google"
	"github.com/ChoiKyubum/riot/constants/queue"
	"github.com/ChoiKyubum/riot/constants/region"
	"github.com/ChoiKyubum/riot/ratelimit"
)

func main() {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	key := os.Getenv("RIOT_APIKEY")
	ctx := context.Background()
	ds, err := datastore.NewClient(ctx, project)
	if err != nil {
		log.Fatal(err)
	}
	bq, err := bigquery.NewClient(ctx, project)
	if err != nil {
		log.Fatal(err)
	}
	ba := bigquery_aggregator.New("TestAggregator", "TestAggregatorDataset", "matches", ds, bq)

	limiter := ratelimit.NewLimiter()
	httpClient := http.DefaultClient
	underlying := apiclient.New(key, httpClient, limiter)
	cache, err := google.NewDatastore(ctx, project, "TestAggregatorCache")
	if err != nil {
		log.Fatal(err)
	}
	client := cachedclient.New(underlying, cache)
	agg := data_aggregation.NewAggregator(client, ba)
	err = agg.AggregateChallengerLeagueMatches(ctx, region.NA1, queue.RankedSolo5x5, time.Now().Add(-10*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	log.Print("saved challenger matches")
}
