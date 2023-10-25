package repository

import (
	"fmt"
	"strings"
	"time"
)

type (
	Model struct {
		EventType   string    `json:"event_type" bson:"event_type"`
		AggregateID string    `json:"aggregate_id" bson:"aggregate_id"`
		Time        time.Time `json:"time" bson:"time"`
		Data        any       `json:"data" bson:"data"`
		Hash        string    `json:"hash" bson:"hash"`
	}

	HtmlElement struct {
		name, text string
		elements   []HtmlElement
	}

	HtmlBuilder struct {
		rootName string
		root     HtmlElement
	}
)

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

func (e *HtmlElement) String() string {
	return e.string(0)
}

func (e *HtmlElement) string(indent int) string {
	sb := strings.Builder{}
	i := strings.Repeat(" ", indentSize*indent)
	sb.WriteString(fmt.Sprintf("%s<%s>\n",
		i, e.name))
	if len(e.text) > 0 {
		sb.WriteString(strings.Repeat(" ",
			indentSize*(indent+1)))
		sb.WriteString(e.text)
		sb.WriteString("\n")
	}

	for _, el := range e.elements {
		sb.WriteString(el.string(indent + 1))
	}
	sb.WriteString(fmt.Sprintf("%s</%s>\n",
		i, e.name))
	return sb.String()
}

func (b *HtmlBuilder) String() string {
	return b.root.String()
}

func (b *HtmlBuilder) AddChild(
	childName, childText string) {
	e := HtmlElement{childName, childText, []HtmlElement{}}
	b.root.elements = append(b.root.elements, e)
}

func (b *HtmlBuilder) AddChildFluent(
	childName, childText string) *HtmlBuilder {
	e := HtmlElement{childName, childText, []HtmlElement{}}
	b.root.elements = append(b.root.elements, e)
	return b
}

const styleHTML = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>balance summary</title>\n</head>\n<body style=\"border: 2px solid #D4D4D8; border-radius: 10px; padding: 20px;\">\n\n    <div style=\"display: flex; align-items: center; gap: 20px;\">\n        <img src=\"https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQ4atBZhDlfY-w4vVVHDWBTmQ1rA8ORtVtXNpZLR4M&s\" alt=\"\" style=\"width: 80px;\">\n        <h2>Balance summary</h2>\n    </div>"
