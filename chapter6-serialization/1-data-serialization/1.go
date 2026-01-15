package __data_serialization

import (
	"bytes"
	"encoding/gob"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(gl)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decode(data []byte) (GameLog, error) {
	buffer := bytes.NewReader(data)
	decoder := gob.NewDecoder(buffer)
	var gl GameLog
	err := decoder.Decode(&gl)
	if err != nil {
		return GameLog{}, err
	}
	return gl, nil
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
