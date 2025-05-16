package events

import (
	"github.com/Ruclo/MyMeals/internal/models"
)

// OrderBroadcaster defines an interface for broadcasting orders.
type OrderBroadcaster interface {
	BroadcastOrder(order *models.Order) error
}
