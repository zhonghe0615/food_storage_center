package logic

import (
	"challenge/config"
	"challenge/entity"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

///////////////////////////
// Fulfillment System    //
///////////////////////////

// FulfillmentSystem encapsulates our order processing logic.
type FulfillmentSystem struct {
	CoolerGroup *entity.StorageGroup // Storage for cold orders.
	HeaterGroup *entity.StorageGroup // Storage for hot orders.
	ShelfGroup  *entity.StorageGroup // Storage for room-temperature orders (and fallback).
	Actions     []Action             // Log of actions performed.
	aLock       sync.Mutex           // Protects the actions slice.
	mutex       sync.Mutex           // Protects the PlaceOrder function
	pickupLock  sync.Mutex           // Protects the PickupOrder function
}

// Action represents an event (place, move, pickup, discard) on an order.
type Action struct {
	Timestamp int64  // Unix timestamp in microseconds.
	OrderID   string // Order identifier.
	Action    string // Action type.
}

// NewFulfillmentSystem initializes the system based on a Config.
func NewFulfillmentSystem(cfg config.FulfillmentConfig) *FulfillmentSystem {
	// TODO: Should be better to use a factory pattern here if different types of storage diverge in initialisation.
	coolers := &entity.StorageGroup{}
	for i := 1; i <= cfg.NumCoolers; i++ {
		name := fmt.Sprintf("Cooler-%d", i)
		coolers.Storages = append(coolers.Storages, entity.NewStorage(name, cfg.CoolerCap))
		log.Printf("Created cooler: %s", name)
	}
	heaters := &entity.StorageGroup{}
	for i := 1; i <= cfg.NumHeaters; i++ {
		name := fmt.Sprintf("Heater-%d", i)
		heaters.Storages = append(heaters.Storages, entity.NewStorage(name, cfg.HeaterCap))
		log.Printf("Created heater: %s", name)
	}
	shelves := &entity.StorageGroup{}
	for i := 1; i <= cfg.NumShelves; i++ {
		name := fmt.Sprintf("Shelf-%d", i)
		shelves.Storages = append(shelves.Storages, entity.NewStorage(name, cfg.ShelfCap))
		log.Printf("Created shelf: %s", name)
	}
	return &FulfillmentSystem{
		CoolerGroup: coolers,
		HeaterGroup: heaters,
		ShelfGroup:  shelves,
		Actions:     make([]Action, 0),
		aLock:       sync.Mutex{},
		mutex:       sync.Mutex{},
		pickupLock:  sync.Mutex{},
	}
}

// logAction records an action and prints it.
func (fs *FulfillmentSystem) logAction(orderID, actionType string, executeTime time.Time) {
	fs.aLock.Lock()
	defer fs.aLock.Unlock()
	action := Action{
		Timestamp: executeTime.UnixMicro(),
		OrderID:   orderID,
		Action:    actionType,
	}
	fs.Actions = append(fs.Actions, action)
	log.Printf("Action: %-7s OrderID: %-8s Timestamp: %d", actionType, orderID, action.Timestamp)
}

// PlaceOrder implements the core logic for storing an order.
func (fs *FulfillmentSystem) PlaceOrder(order entity.Order) {
	fs.mutex.Lock()         // Lock the function
	defer fs.mutex.Unlock() // Ensure the lock is released when the function exits

	storedOrder := &entity.StoredOrder{
		Order:    order,
		PlacedAt: time.Now(), // Assuming you want to set the current time as the placement time
	}
	// For hot/cold orders, attempt ideal storage first.
	if order.Temperature == config.TEMP_TYPE_HOT || order.Temperature == config.TEMP_TYPE_COLD {
		var idealGroup *entity.StorageGroup
		if order.Temperature == config.TEMP_TYPE_HOT {
			idealGroup = fs.HeaterGroup
		} else {
			idealGroup = fs.CoolerGroup
		}
		if idealGroup.Add(storedOrder) {
			fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
			return
		}
		// If ideal storage is full, try the shelf.
		if fs.ShelfGroup.Add(storedOrder) {
			fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
			return
		}
		// After failing to add to ideal storage and initial shelf add...
		if fs.ShelfGroup.IsFull() {
			// Attempt to move orders before discarding
			if fs.tryMoveFromShelfGroup(order.Temperature) {
				if fs.ShelfGroup.Add(storedOrder) {
					fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
					return
				}
			}
		}
		// If all else fails, discard an order from the shelf to make space.
		log.Printf("Shelf is full, attempting to discard an order. Adding order: %s\n", order.ID)
		if fs.ShelfGroup.IsFull() {
			fs.discardOrderFromShelfGroup()
		}
		if fs.ShelfGroup.Add(storedOrder) {
			fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
			return
		}
	} else {
		// For room-temperature orders, use the shelf.
		if fs.ShelfGroup.Add(storedOrder) {
			fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
			return
		}
		if fs.ShelfGroup.IsFull() {
			fs.discardOrderFromShelfGroup()
		}
		if fs.ShelfGroup.Add(storedOrder) {
			fs.logAction(order.ID, config.ACTION_TYPE_PLACE, time.Now())
			return
		}
	}
}

// PickupOrder removes an order from any storage group.
func (fs *FulfillmentSystem) PickupOrder(orderID string) {
	fs.pickupLock.Lock()         // Lock the function
	defer fs.pickupLock.Unlock() // Ensure the lock is released when the function exits

	if so, ok := fs.HeaterGroup.Remove(orderID); ok {
		fs.logAction(so.Order.ID, config.ACTION_TYPE_PICKUP, time.Now())
		return
	}
	if so, ok := fs.CoolerGroup.Remove(orderID); ok {
		fs.logAction(so.Order.ID, config.ACTION_TYPE_PICKUP, time.Now())
		return
	}
	if so, ok := fs.ShelfGroup.Remove(orderID); ok {
		fs.logAction(so.Order.ID, config.ACTION_TYPE_PICKUP, time.Now())
		return
	}
	log.Printf("Order %s not found during pickup", orderID)
}

// RunHarness processes orders at the given rate and schedules pickups after a random delay.
func (fs *FulfillmentSystem) RunHarness(orders []entity.Order, orderInterval, minPickup, maxPickup time.Duration) {
	var wg sync.WaitGroup
	// Implement the background reallocation to automatically MOVE or DISCARD orders.
	stopRealloc := make(chan struct{})
	// Start background reallocation.
	go fs.ReallocateOrders(stopRealloc)
	for _, order := range orders {
		wg.Add(1)
		go func(ord entity.Order) {
			defer wg.Done()
			fs.PlaceOrder(ord)
			// Simulate pickup after a random delay between minPickup and maxPickup.
			delay := minPickup + time.Duration(rand.Int63n(int64(maxPickup-minPickup)))
			time.Sleep(delay)
			fs.PickupOrder(ord.ID)
		}(order)
		time.Sleep(orderInterval)
	}
	wg.Wait()
	//close(stopRealloc)
}

// discardOrderFromShelfGroup selects the order with the lowest remaining freshness and discards it.
func (fs *FulfillmentSystem) discardOrderFromShelfGroup() {
	candidate, found := fs.ShelfGroup.GetLeastFreshOrder()
	if !found {
		return
	}
	// Try moving an order before discarding
	if fs.tryMoveFromShelfGroup(candidate.Order.Temperature) {
		return // Order successfully moved, no need to discard
	}
	// If no order could be moved, proceed with discarding
	if _, ok := fs.ShelfGroup.Remove(candidate.Order.ID); ok {
		fs.logAction(candidate.Order.ID, config.ACTION_TYPE_DISCARD, time.Now())
	}
}

func (fs *FulfillmentSystem) tryMoveFromShelfGroup(temp string) bool {
	var idealGroup *entity.StorageGroup
	if temp == config.TEMP_TYPE_HOT {
		idealGroup = fs.HeaterGroup
	} else if temp == config.TEMP_TYPE_COLD {
		idealGroup = fs.CoolerGroup
	} else {
		return false
	}

	orders := fs.ShelfGroup.ListOrders()
	for _, so := range orders {
		if so.Order.Temperature == temp {
			for _, shelf := range fs.ShelfGroup.Storages {
				moved := fs.atomicMoveOrder(so.Order.ID, shelf, idealGroup)
				if moved {
					fs.logAction(so.Order.ID, config.ACTION_TYPE_MOVE, time.Now())
					return true
				}
			}
		}
	}
	return false
}

func (fs *FulfillmentSystem) atomicMoveOrder(orderID string, source *entity.Storage, destination *entity.StorageGroup) bool {
	// Lock the source storage
	source.Lock.Lock()
	order, exists := source.Orders[orderID]
	if !exists {
		source.Lock.Unlock()
		return false
	}

	// If the order is currently not stored under ideal conditions (i.e., stored on the shelf), update the remaining freshness.
	// Note: Only hot/cold orders require this treatment, room temperature does not.
	if order.Order.Temperature != config.TEMP_TYPE_ROOM {
		// Calculate the time t the order has been stored on the shelf
		t := time.Since(order.PlacedAt)
		// Storage under non-ideal conditions consumes freshness at twice the ideal rate
		newRemaining := order.Order.InitialFreshness - 2*t
		if newRemaining <= 0 {
			// The order has expired, do not move
			source.Lock.Unlock()
			return false
		}
		// Update the order's placement time and freshness to the remaining ideal freshness after moving
		order.PlacedAt = time.Now()
		order.Order.Freshness = newRemaining
	}
	// Try to add the order to one of the storages in the destination group
	for _, destStorage := range destination.Storages {
		// Lock the destination storage
		destStorage.Lock.Lock()
		if len(destStorage.Orders) < destStorage.Capacity {
			// Remove the order from the source
			delete(source.Orders, orderID)
			// Add to the destination storage
			destStorage.Orders[orderID] = order
			destStorage.Lock.Unlock()
			source.Lock.Unlock()
			return true
		}
		destStorage.Lock.Unlock()
	}
	source.Lock.Unlock()
	return false
}

func (fs *FulfillmentSystem) ReallocateOrders(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Only attempt reallocation if the shelf is full.
			if !fs.ShelfGroup.IsFull() {
				continue
			}
			shelfOrders := fs.ShelfGroup.ListOrders()
			for _, so := range shelfOrders {
				if so.Order.Temperature == config.TEMP_TYPE_HOT && !fs.HeaterGroup.IsFull() {
					for _, shelf := range fs.ShelfGroup.Storages {
						if fs.atomicMoveOrder(so.Order.ID, shelf, fs.HeaterGroup) {
							fs.logAction(so.Order.ID, config.ACTION_TYPE_MOVE, time.Now())
							break
						}
					}
				} else if so.Order.Temperature == config.TEMP_TYPE_COLD && !fs.CoolerGroup.IsFull() {
					for _, shelf := range fs.ShelfGroup.Storages {
						if fs.atomicMoveOrder(so.Order.ID, shelf, fs.CoolerGroup) {
							fs.logAction(so.Order.ID, config.ACTION_TYPE_MOVE, time.Now())
							break
						}
					}
				}
			}
		case <-stop:
			return
		}
	}
}
