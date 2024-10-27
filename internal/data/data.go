package data

import (
	"encoding/json"
	"io"
	"os"
)

type Event struct {
	ID     uint   `json:"id"`
	Short  string `json:"short_url"`
	Long   string `json:"original_url"`
	UserID string `json:"user_id"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

var P *Producer

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteEvent(event *Event) error {
	return p.encoder.Encode(&event)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadAllEvents() ([]Event, error) {
	var events []Event

	for {
		event := &Event{}
		if err := c.decoder.Decode(&event); err != nil {
			if err == io.EOF { // Достигнут конец файла
				break
			}
			return nil, err // Возвращаем ошибку, если она не EOF
		}
		events = append(events, *event)
	}

	return events, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
