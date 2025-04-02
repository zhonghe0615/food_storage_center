package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	css "challenge/client"
	"challenge/config"
	"challenge/entity"
	"challenge/logic"
)

var (
	// Command-line flags for the challenge client.
	endpoint = flag.String("endpoint", "https://api.hezhong.com", "integration test endpoint")
	auth     = flag.String("auth", "", "Authentication token (required)")
	name     = flag.String("name", "", "Problem name. Leave blank (optional)")
	seed     = flag.Int64("seed", 1, "Problem seed (random if zero)")

	// Inverse order rate and pickup intervals.
	rate = flag.Duration("rate", 500*time.Millisecond, "Inverse order rate (time between order placements)")
	min  = flag.Duration("min", 4*time.Second, "Minimum pickup time")
	max  = flag.Duration("max", 8*time.Second, "Maximum pickup time")

	// Config file for storage configuration
	configFile = flag.String("config", "config/init.json", "Path to storage configuration file")
)

///////////////////////////
// Fulfillment System    //
///////////////////////////

///////////////////////////
// Integration with Client//
///////////////////////////

// main integrates our fulfillment system with the challenge client.
// It fetches orders from the server, processes them, and submits the actions.
func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	// Load storage configuration
	cfg := config.LoadConfig(*configFile)

	// Create a client using the command-line parameters
	client := css.NewClient(*endpoint, *auth)
	id, ordersFromServer, err := client.New(*name, *seed)
	if err != nil {
		log.Fatalf("Failed to fetch test problem: %v", err)
	}

	// Log received orders.
	for _, o := range ordersFromServer {
		log.Printf("Received order: %+v", o)
	}

	// Convert orders from the challenge client's type to our internal Order type.
	var orders []entity.Order
	for _, o := range ordersFromServer {
		orders = append(orders, entity.Order{
			ID:               o.ID,
			Name:             o.Name,
			Temperature:      o.Temp,                                   // Assuming client's field is Temp.
			Freshness:        time.Duration(o.Freshness) * time.Second, // Convert seconds to time.Duration.
			InitialFreshness: time.Duration(o.Freshness) * time.Second,
		})
	}

	// Initialize our fulfillment system with the configuration.
	fs := logic.NewFulfillmentSystem(cfg)

	// Run the simulation harness with command-line timing parameters
	fs.RunHarness(orders, *rate, *min, *max)

	// Convert our internal actions to the challenge client's action format.
	var actions []css.Action
	for _, a := range fs.Actions {
		actions = append(actions, css.Action{
			Timestamp: a.Timestamp,
			ID:        a.OrderID,
			Action:    a.Action,
		})
	}

	// Submit the solution using command-line timing parameters
	result, err := client.Solve(id, *rate, *min, *max, actions)
	if err != nil {
		log.Fatalf("Failed to submit test solution: %v", err)
	}

	// Highlight the test result with color - green for pass, red for fail
	if result == "pass" {
		fmt.Printf("\033[1;32mTest result: %v\033[0m\n", result)
	} else {
		fmt.Printf("\033[1;31mTest result: %v\033[0m\n", result)
	}
	time.Sleep(500 * time.Millisecond)

	// Optionally, print all captured actions.
	fmt.Println("\nCaptured Actions:")
	for _, act := range fs.Actions {
		fmt.Printf("%+v\n", act)
	}
}
