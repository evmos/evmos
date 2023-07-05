// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package pubsub

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UnsubscribeFunc func()

type EventBus interface {
	AddTopic(name string, src <-chan coretypes.ResultEvent) error
	RemoveTopic(name string)
	Subscribe(name string) (<-chan coretypes.ResultEvent, UnsubscribeFunc, error)
	Topics() []string
}

type subscriptions struct {
	mu          sync.Mutex
	subscribers map[string]map[uint64]chan<- coretypes.ResultEvent
}

type memEventBus struct {
	topics          map[string]<-chan coretypes.ResultEvent
	topicsMux       *sync.RWMutex
	subs            subscriptions
	currentUniqueID uint64
}

func NewEventBus() EventBus {
	return &memEventBus{
		topics:    make(map[string]<-chan coretypes.ResultEvent),
		topicsMux: new(sync.RWMutex),
		subs:      subscriptions{subscribers: make(map[string]map[uint64]chan<- coretypes.ResultEvent)},
	}
}

func (m *memEventBus) GenUniqueID() uint64 {
	return atomic.AddUint64(&m.currentUniqueID, 1)
}

func (m *memEventBus) Topics() (topics []string) {
	m.topicsMux.RLock()
	defer m.topicsMux.RUnlock()

	topics = make([]string, 0, len(m.topics))
	for topicName := range m.topics {
		topics = append(topics, topicName)
	}

	return topics
}

func (m *memEventBus) AddTopic(name string, src <-chan coretypes.ResultEvent) error {
	m.topicsMux.RLock()
	_, ok := m.topics[name]
	m.topicsMux.RUnlock()

	if ok {
		return errors.New("topic already registered")
	}

	m.topicsMux.Lock()
	m.topics[name] = src
	m.topicsMux.Unlock()

	go m.publishTopic(name, src)

	return nil
}

func (m *memEventBus) RemoveTopic(name string) {
	m.topicsMux.Lock()
	delete(m.topics, name)
	m.topicsMux.Unlock()
}

func (m *memEventBus) Subscribe(name string) (<-chan coretypes.ResultEvent, UnsubscribeFunc, error) {
	m.topicsMux.RLock()
	_, ok := m.topics[name]
	m.topicsMux.RUnlock()

	if !ok {
		return nil, nil, errors.Errorf("topic not found: %s", name)
	}

	ch := make(chan coretypes.ResultEvent)
	m.subs.mu.Lock()
	defer m.subs.mu.Unlock()

	id := m.GenUniqueID()
	if _, ok := m.subs.subscribers[name]; !ok {
		m.subs.subscribers[name] = make(map[uint64]chan<- coretypes.ResultEvent)
	}
	m.subs.subscribers[name][id] = ch

	unsubscribe := func() {
		m.subs.mu.Lock()
		defer m.subs.mu.Unlock()
		delete(m.subs.subscribers[name], id)
	}

	return ch, unsubscribe, nil
}

func (m *memEventBus) publishTopic(name string, src <-chan coretypes.ResultEvent) {
	for {
		msg, ok := <-src
		if !ok {
			m.closeAllSubscribers(name)
			m.topicsMux.Lock()
			delete(m.topics, name)
			m.topicsMux.Unlock()
			return
		}
		m.publishAllSubscribers(name, msg)
	}
}

func (m *memEventBus) closeAllSubscribers(name string) {
	m.subs.mu.Lock()
	defer m.subs.mu.Unlock()

	subsribers := m.subs.subscribers[name]
	delete(m.subs.subscribers, name)
	// #nosec G705
	for _, sub := range subsribers {
		close(sub)
	}
}

func (m *memEventBus) publishAllSubscribers(name string, msg coretypes.ResultEvent) {
	m.subs.mu.Lock()
	defer m.subs.mu.Unlock()
	subsribers := m.subs.subscribers[name]
	// #nosec G705
	for _, sub := range subsribers {
		select {
		case sub <- msg:
		default:
		}
	}
}
