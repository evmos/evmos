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

type subscriptionsContainer struct {
	mu            sync.Mutex
	subscriptions map[string]map[uint64]chan<- coretypes.ResultEvent
}

type topicsContainer struct {
	mu     sync.RWMutex
	topics map[string]<-chan coretypes.ResultEvent
}

type memEventBus struct {
	t               topicsContainer
	s               subscriptionsContainer
	currentUniqueID uint64
}

func NewEventBus() EventBus {
	return &memEventBus{
		t: topicsContainer{topics: make(map[string]<-chan coretypes.ResultEvent)},
		s: subscriptionsContainer{subscriptions: make(map[string]map[uint64]chan<- coretypes.ResultEvent)},
	}
}

func (m *memEventBus) GenUniqueID() uint64 {
	return atomic.AddUint64(&m.currentUniqueID, 1)
}

func (m *memEventBus) Topics() (topics []string) {
	m.t.mu.RLock()
	defer m.t.mu.RUnlock()

	topics = make([]string, 0, len(m.t.topics))
	for topicName := range m.t.topics {
		topics = append(topics, topicName)
	}

	return topics
}

func (m *memEventBus) AddTopic(name string, src <-chan coretypes.ResultEvent) error {
	m.t.mu.RLock()
	_, ok := m.t.topics[name]
	m.t.mu.RUnlock()

	if ok {
		return errors.New("topic already registered")
	}

	m.t.mu.Lock()
	m.t.topics[name] = src
	m.t.mu.Unlock()

	go m.publishTopic(name, src)

	return nil
}

func (m *memEventBus) RemoveTopic(name string) {
	m.t.mu.Lock()
	defer m.t.mu.Unlock()
	delete(m.t.topics, name)
}

func (m *memEventBus) Subscribe(name string) (<-chan coretypes.ResultEvent, UnsubscribeFunc, error) {
	m.t.mu.RLock()
	_, ok := m.t.topics[name]
	m.t.mu.RUnlock()

	if !ok {
		return nil, nil, errors.Errorf("topic not found: %s", name)
	}

	ch := make(chan coretypes.ResultEvent)
	m.s.mu.Lock()
	defer m.s.mu.Unlock()

	id := m.GenUniqueID()
	if _, ok := m.s.subscriptions[name]; !ok {
		m.s.subscriptions[name] = make(map[uint64]chan<- coretypes.ResultEvent)
	}
	m.s.subscriptions[name][id] = ch

	unsubscribe := func() {
		m.s.mu.Lock()
		defer m.s.mu.Unlock()
		delete(m.s.subscriptions[name], id)
	}

	return ch, unsubscribe, nil
}

func (m *memEventBus) publishTopic(name string, src <-chan coretypes.ResultEvent) {
	for {
		msg, ok := <-src
		if !ok {
			m.closeAllSubscribers(name)
			m.t.mu.Lock()
			delete(m.t.topics, name)
			m.t.mu.Unlock()
			return
		}
		m.publishAllSubscribers(name, msg)
	}
}

func (m *memEventBus) closeAllSubscribers(name string) {
	m.s.mu.Lock()
	defer m.s.mu.Unlock()

	subsribers := m.s.subscriptions[name]
	delete(m.s.subscriptions, name)
	// #nosec G705
	for _, sub := range subsribers {
		close(sub)
	}
}

func (m *memEventBus) publishAllSubscribers(name string, msg coretypes.ResultEvent) {
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	subsribers := m.s.subscriptions[name]
	// #nosec G705
	for _, sub := range subsribers {
		select {
		case sub <- msg:
		default:
		}
	}
}
