package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
)

var (
	topics             = make(map[string]*Topic)
	mutex  *sync.Mutex = &sync.Mutex{}
)

type Topic struct {
	name  string
	file  string
	mutex *sync.Mutex
}

func (t *Topic) FilePath() string {
	return "topics/" + t.name
}

func GetOrCreateTopic(name string) (*Topic, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if topic, ok := topics[name]; ok {
		return topic, nil
	} else {
		if newTopic, err := NewTopic(name); err != nil {
			return nil, err
		} else {
			topics[name] = newTopic
			return newTopic, nil
		}
	}
}

func GetTopic(name string) (*Topic, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if topic, ok := topics[name]; ok {
		return topic, nil
	} else {
		return nil, errors.New("topic not found")
	}
}

func NewTopic(name string) (*Topic, error) {
	fileName := fmt.Sprintf("topics/%s", name)
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	file.Close()
	newTopic := &Topic{
		name:  name,
		file:  fileName,
		mutex: &sync.Mutex{},
	}
	return newTopic, nil
}

func LoadTopics() {
	entries, err := os.ReadDir("topics")
	if err != nil {
		panic(err)
	}
	topics = make(map[string]*Topic)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		topic := &Topic{name: name, file: name, mutex: &sync.Mutex{}}
		topics[name] = topic
	}
}

func (t *Topic) LoadEvent(event Event) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(event); err != nil {
		return err
	}
	data := buf.Bytes()
	f, err := os.OpenFile(t.FilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	size := uint32(len(data))
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	if err = binary.Write(f, binary.BigEndian, size); err != nil {
		return err
	}
	return nil
}

func (t *Topic) GetLastEvents(seek int64) (EventsList, int64, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	var results []Event

	f, err := os.Open(t.FilePath())
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	stat, _ := f.Stat()
	pos := stat.Size()
	currentSeek := pos

	for pos > seek {
		pos -= 4
		f.Seek(pos, 0)

		var size uint32
		err = binary.Read(f, binary.BigEndian, &size)
		if err != nil {
			return nil, 0, err
		}

		pos -= int64(size)
		if pos < 0 {
			break
		}
		f.Seek(pos, 0)

		data := make([]byte, size)
		f.Read(data)

		var obj Event
		dec := gob.NewDecoder(bytes.NewReader(data))
		if err = dec.Decode(&obj); err != nil {
			break
		}

		results = append(results, obj)
	}
	slices.Reverse(results)

	return results, currentSeek, nil
}
