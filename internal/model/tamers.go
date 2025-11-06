package model
import (
	"encoding/json"
	"os"
)

type Tamer struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Tamers []Tamer

type TamersMap map[string]bool

func (t *Tamers) Load(fname string) error {
    data, err := os.ReadFile(fname)
    if err != nil {
        return err
    }
    if err := json.Unmarshal(data, t); err != nil {
        return err
    }
    return nil
} 
