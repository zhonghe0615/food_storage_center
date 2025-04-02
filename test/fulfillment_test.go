package test

import (
	"challenge/config"
	"challenge/entity"
	"challenge/logic"
	"testing"
	"time"

	"math/rand"
)

func TestMultipleOrderReallocation(t *testing.T) {
	// Setup: Create a fulfillment system with more complex configuration
	cfg := config.FulfillmentConfig{
		NumCoolers: 1,
		CoolerCap:  2,
		NumHeaters: 1,
		HeaterCap:  2,
		NumShelves: 1,
		ShelfCap:   4,
	}
	fs := logic.NewFulfillmentSystem(cfg)

	// Add multiple orders with different temperatures and freshness
	orders := []entity.Order{
		{ID: "1", Temperature: config.TEMP_TYPE_HOT, Freshness: 5 * time.Second},
		{ID: "2", Temperature: config.TEMP_TYPE_COLD, Freshness: 10 * time.Second},
		{ID: "3", Temperature: config.TEMP_TYPE_HOT, Freshness: 15 * time.Second},
		{ID: "4", Temperature: config.TEMP_TYPE_COLD, Freshness: 20 * time.Second},
	}

	for _, order := range orders {
		fs.PlaceOrder(order)
	}
	delay := 4 + time.Duration(rand.Int63n(int64(4)))
	time.Sleep(delay)

	// Simulate space becoming available in the heater and cooler
	fs.PickupOrder("1")
	fs.PickupOrder("2")

	// Create a stop channel and start reallocation
	// stopRealloc := make(chan struct{})
	// go fs.ReallocateOrders(stopRealloc)

	// Allow some time for reallocation to occur
	time.Sleep(2 * time.Second)

	// Send stop signal to the goroutine
	// close(stopRealloc)

	// Verify: Check if orders were moved to the correct storages
	if _, ok := fs.HeaterGroup.Storages[0].GetOrder("3"); !ok {
		t.Errorf("Order 3 was not reallocated to the heater")
	}
	if _, ok := fs.CoolerGroup.Storages[0].GetOrder("4"); !ok {
		t.Errorf("Order 4 was not reallocated to the cooler")
	}

	// Check if expired orders are discarded
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("1"); ok {
		t.Errorf("Expired order 1 was not discarded")
	}
}

func TestDiscardAllRoomTemperatureOrderFromShelfGroup(t *testing.T) {
	// Setup: Create a fulfillment system with a small shelf capacity
	cfg := config.FulfillmentConfig{
		NumCoolers: 0,
		CoolerCap:  0,
		NumHeaters: 0,
		HeaterCap:  0,
		NumShelves: 1,
		ShelfCap:   2, // Small capacity to trigger discard
	}
	fs := logic.NewFulfillmentSystem(cfg)

	// Add orders with different freshness
	order1 := entity.Order{ID: "1", Temperature: config.TEMP_TYPE_ROOM, Freshness: 5 * time.Second}
	order2 := entity.Order{ID: "2", Temperature: config.TEMP_TYPE_ROOM, Freshness: 10 * time.Second}
	order3 := entity.Order{ID: "3", Temperature: config.TEMP_TYPE_ROOM, Freshness: 15 * time.Second}

	fs.PlaceOrder(order1)
	fs.PlaceOrder(order2)

	// Attempt to place a third order, which should trigger a discard
	fs.PlaceOrder(order3)

	// Verify: Check if the order with the lowest freshness was discarded
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("1"); ok {
		t.Errorf("Order 1 was not discarded as expected")
	}

	// Verify: Check if the other orders are still present
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("2"); !ok {
		t.Errorf("Order 2 should not have been discarded")
	}
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("3"); !ok {
		t.Errorf("Order 3 should have been placed on the shelf")
	}
}

func TestDiscardHybridOrderFromShelfGroup(t *testing.T) {
	cfg := config.FulfillmentConfig{
		NumCoolers: 0,
		CoolerCap:  0,
		NumHeaters: 0,
		HeaterCap:  0,
		NumShelves: 1,
		ShelfCap:   2, // Small capacity to trigger discard
	}
	fs := logic.NewFulfillmentSystem(cfg)

	// Add orders with different freshness
	order1 := entity.Order{ID: "1", Temperature: config.TEMP_TYPE_ROOM, Freshness: 5 * time.Second}
	order2 := entity.Order{ID: "2", Temperature: config.TEMP_TYPE_HOT, Freshness: 8 * time.Second}
	order3 := entity.Order{ID: "3", Temperature: config.TEMP_TYPE_ROOM, Freshness: 15 * time.Second}

	fs.PlaceOrder(order1)
	fs.PlaceOrder(order2)

	// Attempt to place a third order, which should trigger a discard
	fs.PlaceOrder(order3)

	// Verify: Check if the order with the lowest freshness was discarded
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("2"); ok {
		t.Errorf("Order 2 was not discarded as expected")
	}

	// Verify: Check if the other orders are still present
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("1"); !ok {
		t.Errorf("Order 1 should not have been discarded")
	}
	if _, ok := fs.ShelfGroup.Storages[0].GetOrder("3"); !ok {
		t.Errorf("Order 3 should have been placed on the shelf")
	}
}
