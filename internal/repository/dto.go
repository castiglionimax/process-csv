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

const styleHTML = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>balance summary</title>\n    <style>\n        table {\n            font-family: Arial, sans-serif;\n            border-collapse: collapse;\n            width: 50%;\n        }\n\n        th, td {\n            border: 1px solid #dddddd;\n            text-align: left;\n            padding: 8px;\n        }\n\n        th {\n            background-color: #f2f2f2;\n        }\n\n        .header{\n            display: flex;\n            align-items: center;\n            gap: 20px;\n        }  \n        .header img {\n            width: 80px;\n        }\n        .container-tables {\n         width: 100%;\n         }\n         .container-tables  table{\n         width: 100%;\n        }\n        body{\n            border: 2px solid #D4D4D8;\n          border-radius: 10px;\n          padding: 20px;\n        }\n        .table-balance thead tr th {\n            background-color: #166980;\n            color: #fff;\n            text-align: center\n        }        \n        .table-balance tbody tr td {\n            text-align: center\n        }\n        .total-amount p{\n            color: #4b5244;\n            font-size: 15px;\n            font-weight: 700;\n        }\n    </style>\n</head>\n<body >\n    <div class=\"header\">\n\n\n        <img src=\"https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQ4atBZhDlfY-w4vVVHDWBTmQ1rA8ORtVtXNpZLR4M&s\" alt=\"\" />\n\n        <h2>Balance summary</h2>\n\n    </div>"
