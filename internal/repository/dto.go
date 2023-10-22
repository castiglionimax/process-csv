package repository

import "time"

type Model struct {
	EventType   string    `json:"event_type" bson:"event_type"`
	AggregateID string    `json:"aggregate_id" bson:"aggregate_id"`
	Time        time.Time `json:"time" bson:"time"`
	Data        any       `json:"data" bson:"data"`
	Hash        string    `json:"hash" bson:"hash"`
}

func newModel(eventType, aggregateID string, data any, hash string) Model {
	timestamp := time.Now()
	return Model{
		EventType:   eventType,
		AggregateID: aggregateID,
		Time:        timestamp,
		Data:        data,
		Hash:        hash,
	}
}
