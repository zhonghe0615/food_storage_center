package entity

import (
	"challenge/config"
	"log"
	"sync"
	"time"
)

type StorageGroup struct {
	Storages  []*Storage
	storeLock sync.RWMutex // Use RWMutex for the storages
}

// Call this whenever adding an order
func (sg *StorageGroup) Add(order *StoredOrder) bool {
	// Try to add the order to one of the storages
	sg.storeLock.Lock()
	defer sg.storeLock.Unlock()
	log.Println("Adding order to storage group, order:", order.Order.ID)
	for _, storage := range sg.Storages {
		log.Println("Checking storage:", storage.Name)
		if !storage.IsFull() {
			log.Println("Storage is not full, adding order to storage")
			if storage.Add(order) {
				// Successfully added to storage, now add to the priority queue
				return true
			}
		}
	}
	// If all storages are full, return false
	return false
}

// Call this whenever removing an order
func (sg *StorageGroup) Remove(orderID string) (*StoredOrder, bool) {
	sg.storeLock.Lock()
	defer sg.storeLock.Unlock()
	for _, storage := range sg.Storages {
		if removedOrder, ok := storage.Remove(orderID); ok {
			return removedOrder, true
		}
	}
	return nil, false
}

func (sg *StorageGroup) GetLeastFreshOrder() (*StoredOrder, bool) {
	sg.storeLock.RLock()
	defer sg.storeLock.RUnlock()
	var leastFreshOrder *StoredOrder
	var found bool
	//TODO: Consider using a priority queue to get the least fresh order more efficiently,
	//for now this is not a performance bottleneck though.
	for _, storage := range sg.Storages {
		for _, order := range storage.ListOrders() {
			if !found || order.RemainingFreshness() < leastFreshOrder.RemainingFreshness() {
				leastFreshOrder = order
				found = true
			}
		}
	}
	if !found {
		return nil, false
	}
	return leastFreshOrder, true
}

// Helper method to calculate remaining freshness
func (so *StoredOrder) RemainingFreshness() time.Duration {
	elapsed := time.Since(so.PlacedAt)
	if so.Order.Temperature == config.TEMP_TYPE_ROOM {
		return so.Order.Freshness - elapsed
	}
	return (so.Order.Freshness / 2) - elapsed
}

func (sg *StorageGroup) ListOrders() []*StoredOrder {
	sg.storeLock.RLock()
	defer sg.storeLock.RUnlock()
	var orders []*StoredOrder
	for _, storage := range sg.Storages {
		orders = append(orders, storage.ListOrders()...)
	}
	return orders
}

func (sg *StorageGroup) IsFull() bool {
	sg.storeLock.RLock()
	defer sg.storeLock.RUnlock()

	for _, storage := range sg.Storages {
		if !storage.IsFull() {
			return false
		}
	}
	return true
}
