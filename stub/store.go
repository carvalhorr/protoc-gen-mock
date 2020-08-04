package stub

import (
	"fmt"
	"sync"
)

func NewInMemoryStubsStore(allowRepeats bool) StubsStore {
	return &inMemoryStubsStore{
		Stubs:         make(map[string]map[string]*Stub, 0),
		AllowRepeated: allowRepeats,
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
	Stubs         map[string]map[string]*Stub
	AllowRepeated bool
	mutex         sync.RWMutex
}

func (s *inMemoryStubsStore) Add(e *Stub) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.Stubs[e.FullMethod]
	if !ok {
		s.Stubs[e.FullMethod] = make(map[string]*Stub, 0)
	}

	if !s.AllowRepeated && s.exists(e) {
		return fmt.Errorf("stub already exist: %s -> %s", e.FullMethod, e.Request.String())
	}

	s.Stubs[e.FullMethod][e.Request.String()] = e

	return nil
}

func (s *inMemoryStubsStore) GetStubsMapForMethod(method string) (stubs map[string]*Stub) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getStubsMapForMethod(method)
}

func (s *inMemoryStubsStore) getStubsMapForMethod(method string) (stubs map[string]*Stub) {
	return s.Stubs[method]
}

func (s *inMemoryStubsStore) GetStubsForMethod(method string) []*Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getStubssForMethod(method)
}

func (s *inMemoryStubsStore) getStubssForMethod(method string) []*Stub {
	stubs := make([]*Stub, 0)
	stubsForMethod := s.getStubsMapForMethod(method)
	for _, e := range stubsForMethod {
		stubs = append(stubs, e)
	}
	return stubs
}

func (s *inMemoryStubsStore) GetAllStubs() []*Stub {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	allStubs := make([]*Stub, 0)
	for methodName := range s.Stubs {
		allStubs = append(allStubs, s.getStubssForMethod(methodName)...)
	}

	return allStubs
}

func (s *inMemoryStubsStore) Update(e *Stub) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.exists(e) {
		return fmt.Errorf("stub does not exist: %s -> %s", e.FullMethod, e.Request.String())
	}

	s.Stubs[e.FullMethod][e.Request.String()] = e

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
	s.Stubs[method] = make(map[string]*Stub)
}

func (s *inMemoryStubsStore) DeleteAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for method := range s.Stubs {
		s.deleteAllForMethod(method)
	}
}
