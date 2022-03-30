package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
)

type BalancingStrategy interface {
	Init([]*Backend)
	GetNextBackend(IncomingReq) *Backend
	RegisterBackend(*Backend)
	PrintTopology()
}

type RRBalancingStrategy struct {
	Index    int
	Backends []*Backend
}

type StaticBalancingStrategy struct {
	Index    int
	Backends []*Backend
}

type HashedBalancingStrategy struct {
	OccupiedSlots []int
	Backends      []*Backend
}

func (s *RRBalancingStrategy) Init(backends []*Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *RRBalancingStrategy) GetNextBackend(_ IncomingReq) *Backend {
	s.Index = (s.Index + 1) % len(s.Backends)
	return s.Backends[s.Index]
}

func (s *RRBalancingStrategy) RegisterBackend(backend *Backend) {
	s.Backends = append(s.Backends, backend)
}

func (s *RRBalancingStrategy) PrintTopology() {
	for index, backend := range s.Backends {
		fmt.Println(fmt.Sprintf("      [%d] %s", index, backend))
	}
}

func NewRRBalancingStrategy(backends []*Backend) *RRBalancingStrategy {
	strategy := new(RRBalancingStrategy)
	strategy.Init(backends)
	return strategy
}

func (s *StaticBalancingStrategy) Init(backends []*Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *StaticBalancingStrategy) GetNextBackend(_ IncomingReq) *Backend {
	return s.Backends[s.Index]
}

func (s *StaticBalancingStrategy) RegisterBackend(backend *Backend) {
	s.Backends = append(s.Backends, backend)
}

func (s *StaticBalancingStrategy) PrintTopology() {
	for index, backend := range s.Backends {
		if index == s.Index {
			fmt.Println(fmt.Sprintf("      [%s] %s", "x", backend))
		} else {
			fmt.Println(fmt.Sprintf("      [%s] %s", " ", backend))
		}
	}
}

func NewStaticBalancingStrategy(backends []*Backend) *StaticBalancingStrategy {
	strategy := new(StaticBalancingStrategy)
	strategy.Init(backends)
	return strategy
}

func hash(s string) int {
	h := md5.New()
	var sum int = 0
	io.WriteString(h, s)
	for _, b := range h.Sum(nil) {
		sum += int(b)
	}
	return sum % 19
}

func (s *HashedBalancingStrategy) Init(backends []*Backend) {
	s.OccupiedSlots = []int{}
	s.Backends = []*Backend{}
	for _, backend := range backends {
		key := hash(backend.String())

		if len(s.OccupiedSlots) == 0 {
			s.OccupiedSlots = append(s.OccupiedSlots, key)
			s.Backends = append(s.Backends, backend)
			continue
		}

		index := sort.Search(len(s.OccupiedSlots), func(i int) bool {
			return s.OccupiedSlots[i] >= key
		})

		if index == len(s.OccupiedSlots) {
			s.OccupiedSlots = append(s.OccupiedSlots, key)
		} else {
			s.OccupiedSlots = append(s.OccupiedSlots[:index+1], s.OccupiedSlots[index:]...)
			s.OccupiedSlots[index] = key
		}

		if index == len(s.Backends) {
			s.Backends = append(s.Backends, backend)
		} else {
			s.Backends = append(s.Backends[:index+1], s.Backends[index:]...)
			s.Backends[index] = backend
		}
	}
}

func (s *HashedBalancingStrategy) GetNextBackend(req IncomingReq) *Backend {
	slot := hash(req.reqId)
	index := sort.Search(len(s.OccupiedSlots), func(i int) bool { return s.OccupiedSlots[i] > slot })
	return s.Backends[index%len(s.Backends)]
}

func (s *HashedBalancingStrategy) RegisterBackend(backend *Backend) {
	key := hash(backend.String())
	index := sort.Search(len(s.OccupiedSlots), func(i int) bool { return s.OccupiedSlots[i] >= key })

	if index == len(s.OccupiedSlots) {
		s.OccupiedSlots = append(s.OccupiedSlots, key)
	} else {
		s.OccupiedSlots = append(s.OccupiedSlots[:index+1], s.OccupiedSlots[index:]...)
		s.OccupiedSlots[index] = key
	}

	if index == len(s.Backends) {
		s.Backends = append(s.Backends, backend)
	} else {
		s.Backends = append(s.Backends[:index+1], s.Backends[index:]...)
		s.Backends[index] = backend
	}
}

func (s *HashedBalancingStrategy) PrintTopology() {
	index, i := 0, 0
	for i = 0; i < 19; i++ {
		if index < len(s.OccupiedSlots) && s.OccupiedSlots[index] == i {
			fmt.Println(fmt.Sprintf("      [%2d] %s", i, s.Backends[index]))
			index++
		} else {
			fmt.Println(fmt.Sprintf("      [%2d] -", i))
		}
	}
	for ; i < 19; i++ {
		fmt.Println(fmt.Sprintf("      [%2d] -", i))
	}
}

func NewHashedBalancingStrategy(backends []*Backend) *HashedBalancingStrategy {
	strategy := new(HashedBalancingStrategy)
	strategy.Init(backends)
	return strategy
}
