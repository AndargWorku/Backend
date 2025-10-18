// internal/models/payment.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TxRef     string    `json:"tx_ref"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
