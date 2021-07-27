# get-event-volume
This app uses New Relic's GraphQL API to generate a list of all event types found in your New Relic account.  For each event type found, it then queries for the volume of data sent, showing the results in bytes.

It's the equivalent to doing this in NRQL:

  `show event types`

Then for each event type:

  `FROM (eventType) SELECT bytecountestimate()`
  
The output is formatted *eventType, bytesCollected* with one event type per line
  
We output to the command line, so you'll need to redirect to a file if you want to view this info in an editor or spreadsheet.

The utility takes the following command line arguments:

`-apikey : [REQUIRED] A user API key that is good for the account you want to pull data from.`  
`-accountId : [REQUIRED] The ID of the account you want to pull data from.`  
`-timeframe : The number of hours prior to now that we query data for.  The default is 1 hour.`  
`-filter : The name of a file that contains event types to be excluded from this query.  This is an optional switch.`  
`-verbose : Increases the verbosity of the app's output for troubleshooting purposes.  This is an optional switch.`
