// Package vandargo provides a secure integration with the Vandar payment gateway
// storage.go implements storage utilities for transaction persistence
package vandargo

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStorage is a simple in-memory implementation of StorageInterface
type MemoryStorage struct {
	transactions map[string]*Transaction
	mutex        sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		transactions: make(map[string]*Transaction),
	}
}

// StoreTransaction saves a new transaction to storage
func (s *MemoryStorage) StoreTransaction(ctx context.Context, transaction *Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if transaction.ID == "" {
		return fmt.Errorf("transaction ID cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store a copy of the transaction to prevent external modifications
	transactionCopy := *transaction
	s.transactions[transaction.Token] = &transactionCopy

	return nil
}

// GetTransaction retrieves a transaction by token
func (s *MemoryStorage) GetTransaction(ctx context.Context, token string) (*Transaction, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	transaction, exists := s.transactions[token]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", token)
	}

	// Return a copy to prevent external modifications
	transactionCopy := *transaction
	return &transactionCopy, nil
}

// UpdateTransaction updates an existing transaction
func (s *MemoryStorage) UpdateTransaction(ctx context.Context, transaction *Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if transaction.ID == "" {
		return fmt.Errorf("transaction ID cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.transactions[transaction.Token]
	if !exists {
		return fmt.Errorf("transaction not found: %s", transaction.Token)
	}

	// Update the transaction
	transaction.UpdatedAt = time.Now()
	transactionCopy := *transaction
	s.transactions[transaction.Token] = &transactionCopy

	return nil
}

// GetTransactionsByStatus retrieves transactions by their status
func (s *MemoryStorage) GetTransactionsByStatus(ctx context.Context, status string) ([]*Transaction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []*Transaction

	for _, transaction := range s.transactions {
		if transaction.Status == status {
			// Create a copy to prevent external modifications
			transactionCopy := *transaction
			result = append(result, &transactionCopy)
		}
	}

	return result, nil
}
