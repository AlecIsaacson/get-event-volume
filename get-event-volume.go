//This app returns a list of all New Relic events and the bytes consumed by them.
package main

import (
    "context"
    "fmt"
    "flag"
    "os"
    "bufio"

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
  nrEvents := flag.String("filter", "", "The file that contains events not to be processed")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
  timeframe := flag.String("since", "1", "Number of hours to get data for")
	flag.Parse()

  if *logVerbose {
    fmt.Println("Get-Event-Volume v1.0")
    fmt.Println("Verbose logging enabled")
  }
  
  //Get the New Relic stock events to be ignored.
  nrStockEvents := getStockEvents(*nrEvents)

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

  // Get the list of event types found in this account.
  var graphqlEventResponse nrNRQLEventResultStruct
  if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlEventResponse); err != nil {
      panic(err)
  }

  // fmt.Println("Results:", graphqlEventResponse)

  //Return the results and get each eventTypes volume.
  for _,result := range graphqlEventResponse.Actor.Account.Nrql.Results {
    if _, found := nrStockEvents[result.EventType]; !found {
      nrSizeQuery := "FROM `" + result.EventType + "` SELECT bytecountestimate() SINCE " + *timeframe + "hours ago" 
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

func getStockEvents(filterFile string) (map[string]string) {
  nrStockEvents := make(map[string]string)
  eventFile, err := os.Open(filterFile)
  if err != nil {
    //log.Fatal(err)
  }
  defer eventFile.Close()
  
  scanner := bufio.NewScanner(eventFile)
    for scanner.Scan() {
      nrStockEvents[scanner.Text()] = scanner.Text()
    }
  return nrStockEvents
}
