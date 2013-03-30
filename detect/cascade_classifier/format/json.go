package format

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type Feature struct {
	Size int   `json:"size"`
	PX   []int `json:"px"`
	PY   []int `json:"py"`
	PZ   []int `json:"pz"`
	NX   []int `json:"nx"`
	NY   []int `json:"ny"`
	NZ   []int `json:"nz"`
}

type StageClassifier struct {
	Count     int       `json:"count"`
	Threshold float64   `json:"threshold"`
	Feature   []Feature `json:"feature"`
	Alpha     []float64 `json:"alpha"`
}

type Cascade struct {
	Length          int               `json:"length"`
	Width           int               `json:"width"`
	Height          int               `json:"height"`
	StageClassifier []StageClassifier `json:"stage_classifier"`
}

func LoadJson(r io.Reader) (cascade *Cascade, err error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal("Could not read stdin: ", err)
	}
	if err = json.Unmarshal(data, &cascade); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %v", err)
	}
	return
}
