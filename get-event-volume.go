//This app returns a list of all New Relic entities (name and GUID) that match a particular query.
//It's useful as a feeder to another app that will do something to those entties.
package main

import (
    "context"
    "fmt"
    "flag"

    "github.com/machinebox/graphql"
)

// The NR GraphQL API returns NRQL results in this struct
type nrNRQLEventResultStruct struct {
	// Data struct {
		Actor struct {
			Account struct {
				Nrql struct {
					Results []struct {
						EventType string `json:"eventType"`
					} `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	// } `json:"data"`
	// Extensions struct {
	// 	NrOnly struct {
	// 		Docs         string `json:"_docs"`
	// 		AllCacheHits []struct {
	// 			Count int    `json:"count"`
	// 			Name  string `json:"name"`
	// 		} `json:"allCacheHits"`
	// 		DeepTrace      string `json:"deepTrace"`
	// 		HTTPRequestLog []struct {
	// 			Body string `json:"body"`
	// 			Curl string `json:"curl"`
	// 		} `json:"httpRequestLog"`
	// 	} `json:"nrOnly"`
	// } `json:"extensions"`
}

type nrNRQLByteCountResultsStruct struct {
	// Data struct {
		Actor struct {
			Account struct {
				Nrql struct {
					Results []struct {
						Bytecountestimate interface{} `json:"bytecountestimate"`
						Result            int         `json:"result"`
					} `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	// } `json:"data"`
	// Extensions struct {
	// 	NrOnly struct {
	// 		Docs         string `json:"_docs"`
	// 		AllCacheHits []struct {
	// 			Count int    `json:"count"`
	// 			Name  string `json:"name"`
	// 		} `json:"allCacheHits"`
	// 		DeepTrace      string `json:"deepTrace"`
	// 		HTTPRequestLog []struct {
	// 			Body string `json:"body"`
	// 			Curl string `json:"curl"`
	// 		} `json:"httpRequestLog"`
	// 	} `json:"nrOnly"`
	// } `json:"extensions"`
}

func main() {
  // Define command line flags and defaults.
  nrAPI := flag.String("apikey", "", "New Relic GraphQL API Key")
  nrAccount := flag.Int("accountId", 0, "New Relic account ID")
  // nrQuery := flag.String("nrql","name like '%'","A valid NRQL query")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	flag.Parse()

  if *logVerbose {
    fmt.Println("Entity finder v1.0")
    fmt.Println("Verbose logging enabled")
  }

  //Spawn a new GraphQL client
  graphqlClient := graphql.NewClient("https://api.newrelic.com/graphql")

  //Generate the GraphQL query structure.
  graphqlRequest := graphql.NewRequest(`
    query($query: Nrql!, $account: Int!)
    {
      actor {
        account(id: $account) {
          nrql(query: $query, timeout: 120) {
            results
          }
        }
      }
    }
  `)

  //Set the query and headers.
  graphqlRequest.Var("query", "show eventTypes")
  graphqlRequest.Var("account", *nrAccount)
  graphqlRequest.Header.Set("API-Key", *nrAPI)

  // Get the list of event types.
  var graphqlEventResponse nrNRQLEventResultStruct
  if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlEventResponse); err != nil {
      panic(err)
  }

  // fmt.Println("Results:", graphqlEventResponse)

  //Return the results and get each eventTypes volume.
  for _,result := range graphqlEventResponse.Actor.Account.Nrql.Results {
    if (result.EventType != "Log" && result.EventType != "LogExtendedRecord") {
      nrSizeQuery := "FROM `" + result.EventType + "` SELECT bytecountestimate()"
      graphqlRequest.Var("query", nrSizeQuery)
      graphqlRequest.Var("account", *nrAccount)
      graphqlRequest.Header.Set("API-Key", *nrAPI)

      var graphqlSizeResponse nrNRQLByteCountResultsStruct
      if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlSizeResponse); err != nil {
          panic(err)
      }
      
      for _,sizeResult := range graphqlSizeResponse.Actor.Account.Nrql.Results {
        fmt.Printf("%v,%v\n", result.EventType, sizeResult.Result)
      }
    }
  }  
}
