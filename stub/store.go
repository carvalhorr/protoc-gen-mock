package stub

import (
	"fmt"
	"sync"
)

func NewInMemoryStubsStore() StubsStore {
	return &inMemoryStubsStore{
		Stubs:         make(map[string]map[string][]*Stub, 0),
		AllowRepeated: false,
	}
}

func NewRecordingsStore() RecordingsStore {
	return &inMemoryStubsStore{
		Stubs:         make(map[string]map[string][]*Stub, 0),
		AllowRepeated: true,
	}
}

type StubsStore interface {
	Add(e *Stub) error
	GetStubsMapForMethod(method string) map[string]*Stub
	GetStubsForMethod(method string) []*Stub
	GetAllStubs() []*Stub
	Update(e *Stub) error
	DeleteAllForMethod(method string)
	DeleteAll()
	Delete(e *Stub) error
	Exists(e *Stub) bool
}

type RecordingsStore interface {
	Add(e *Stub) error
	GetAllStubs() []*Stub
}

type inMemoryStubsStore struct {
	// Stores the stubs registered.
	// First map's key is a full method name
	// Second map's key is a gRPC request payload in JSON format
	// The data stub here would look like:
	// /carvalhorr.proto.test.TestProtobuf/GetProtoTest ->
	//               {\"customerId\":1593510,\"siteId\":10153291} -> stub1
	//               {\"customerId\":1545,\"siteId\":10153291} -> stub2
	// /full method name 2 ->
	//               request 1 -> stub3
	//               request 2 -> stub4
	Stubs         map[string]map[string][]*Stub
	AllowRepeated bool
	mutex         sync.RWMutex
}

func (s *inMemoryStubsStore) Add(e *Stub) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.Stubs[e.FullMethod]
	if !ok {
		s.Stubs[e.FullMethod] = make(map[string][]*Stub, 0)
	}

	if !s.AllowRepeated && s.exists(e) {
		return fmt.Errorf("stub already exist: %s -> %s", e.FullMethod, e.Request.String())
	}

	s.Stubs[e.FullMethod][e.Request.String()] = append(s.Stubs[e.FullMethod][e.Request.String()], e)

	return nil
}

func (s *inMemoryStubsStore) GetStubsMapForMethod(method string) (stubs map[string]*Stub) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getStubsMapForMethod(method)
}

func (s *inMemoryStubsStore) getStubsMapForMethod(method string) (stubs map[string]*Stub) {
	response := make(map[string]*Stub)
	for req, stubs := range s.Stubs[method] {
		response[req] = stubs[0]
	}
	return response
}

func (s *inMemoryStubsStore) GetStubsForMethod(method string) []*Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getStubsForMethod(method)
}

func (s *inMemoryStubsStore) getStubsForMethod(method string) []*Stub {
	resp := make([]*Stub, 0)
	for _, stubs := range s.Stubs[method] {
		for _, e := range stubs {
			resp = append(resp, e)
		}
	}
	return resp
}

func (s *inMemoryStubsStore) GetAllStubs() []*Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	allStubs := make([]*Stub, 0)
	for methodName := range s.Stubs {
		allStubs = append(allStubs, s.getStubsForMethod(methodName)...)
	}

	return allStubs
}

func (s *inMemoryStubsStore) Update(e *Stub) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.exists(e) {
		return fmt.Errorf("stub does not exist: %s -> %s", e.FullMethod, e.Request.String())
	}

	s.Stubs[e.FullMethod][e.Request.String()][0] = e

	return nil
}

func (s *inMemoryStubsStore) Delete(e *Stub) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.exists(e) {
		return fmt.Errorf("stub does not exist: %s -> %s", e.FullMethod, e.Request.String())
	}

	delete(s.Stubs[e.FullMethod], e.Request.String())

	return nil
}

func (s *inMemoryStubsStore) Exists(e *Stub) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.exists(e)
}

func (s *inMemoryStubsStore) exists(e *Stub) bool {
	stubsPerMethod := s.Stubs[e.FullMethod]
	foundStub := stubsPerMethod[e.Request.String()]
	return foundStub != nil
}

func (s *inMemoryStubsStore) DeleteAllForMethod(method string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deleteAllForMethod(method)
}

func (s *inMemoryStubsStore) deleteAllForMethod(method string) {
	s.Stubs[method] = make(map[string][]*Stub)
}

func (s *inMemoryStubsStore) DeleteAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for method := range s.Stubs {
		s.deleteAllForMethod(method)
	}
}
