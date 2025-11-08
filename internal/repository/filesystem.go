package repository

import (
	"encoding/json"
	"os"

	"github.com/akarashov/urltamer/internal/model"
)

func SaveToFile(tamers model.Tamers, fileName string) error {
	data, err := json.MarshalIndent(tamers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0666)
}
