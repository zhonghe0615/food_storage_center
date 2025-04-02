package entity

import (
	"log"
	"sync"
	"time"
)

// Order represents a food order in our system.
type Order struct {
	ID               string        // Unique order identifier.
	Name             string        // Food name.
	Temperature      string        // Temperature requirement
	Freshness        time.Duration // Freshness duration in ideal conditions.
	InitialFreshness time.Duration // Initial freshness duration in ideal conditions.
}

// StoredOrder wraps an Order along with its placement time.
type StoredOrder struct {
	Order    Order
	PlacedAt time.Time
}

// Storage represents a single storage unit with a fixed capacity.
type Storage struct {
	Name     string                  // Storage unit name.
	Capacity int                     // Maximum orders it can hold.
	Orders   map[string]*StoredOrder // Map of order IDs to stored orders.
	Lock     sync.RWMutex            // Protects access to Orders.
}

// NewStorage creates a new storage instance.
func NewStorage(name string, capacity int) *Storage {
	return &Storage{
		Name:     name,
		Capacity: capacity,
		Orders:   make(map[string]*StoredOrder),
	}
}

// Get retrieves an order by ID.
func (s *Storage) GetOrder(orderID string) (*StoredOrder, bool) {
	so, exists := s.Orders[orderID]
	return so, exists
}

// Add attempts to add an order to storage.
func (s *Storage) Add(order *StoredOrder) bool {
	log.Println("Adding order to storage, order:", order.Order.ID)
	// If the order is already present, update it.
	if _, exists := s.Orders[order.Order.ID]; exists {
		s.Orders[order.Order.ID] = order
		return true
	}
	// Otherwise, if there is room, add it.
	if len(s.Orders) < s.Capacity {
		s.Orders[order.Order.ID] = order
		return true
	}
	return false
}

// Remove deletes an order by ID.
func (s *Storage) Remove(orderID string) (*StoredOrder, bool) {
	so, exists := s.Orders[orderID]
	if exists {
		delete(s.Orders, orderID)
	}
	return so, exists
}

// Action represents an event (place, move, pickup, discard) on an order.
type Action struct {
	Timestamp int64  // Unix timestamp in microseconds.
	OrderID   string // Order identifier.
	Action    string // Action type.
}

// IsFull checks if the storage is at capacity.
func (s *Storage) IsFull() bool {
	s.Lock.RLock() // Use a read lock for read-only access.
	defer s.Lock.RUnlock()
	return len(s.Orders) >= s.Capacity
}

// ListOrders returns a snapshot of orders in storage.
func (s *Storage) ListOrders() []*StoredOrder {
	// s.Lock.RLock() // Use a read lock for read-only access.
	// defer s.Lock.RUnlock()
	orders := make([]*StoredOrder, 0, len(s.Orders))
	for _, so := range s.Orders {
		orders = append(orders, so)
	}
	return orders
}
