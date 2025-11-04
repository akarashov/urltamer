package repository

import (
	"encoding/json"
	"os"

	"github.com/akarashov/urltamer/internal/model"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

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

func (p *Producer) WriteEvent(event *model.Tamer) error {
	// Encode the event (pointer to Tamer) directly. Previously this used
	// &event which produced a pointer-to-pointer and incorrect JSON.
	return p.encoder.Encode(event)
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

func (c *Consumer) ReadEvent() (*model.Tamer, error) {
	tamer := &model.Tamer{}
	if err := c.decoder.Decode(&tamer); err != nil {
		return nil, err
	}

	return tamer, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

// SaveToFile writes the full tamers slice as a JSON array to the specified file.
func SaveToFile(tamers model.Tamers, fileName string) error {
	data, err := json.MarshalIndent(tamers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0666)
}
