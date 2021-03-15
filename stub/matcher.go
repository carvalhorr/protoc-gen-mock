package stub

import (
	"context"
	"google.golang.org/grpc/metadata"
	"sort"
	"strings"
)

// Search and match stubs in the StubsStore
type StubsMatcher interface {
	Match(ctx context.Context, fullMethod, requestJson string) *Stub
}

// Creates new stubs matcher
func NewStubsMatcher(store StubsStore) StubsMatcher {
	return &stubsMatcher{
		StubsStore: store,
	}
}

type stubsMatcher struct {
	StubsStore StubsStore
}

// Returns the Stub in the StubsStore that matches the method and requestJSON provided OR nil if no stub is found
func (m *stubsMatcher) Match(ctx context.Context, fullMethod, requestJson string) *Stub {
	stubsForMethod := m.StubsStore.GetStubsMapForMethod(fullMethod)
	if stubsForMethod == nil {
		return nil
	}
	var firstPartialMatch *Stub
	for _, stub := range stubsForMethod {
		switch stub.Request.Match {
		case "exact":
			if stub.Request.Content.Equals(JsonString(requestJson)) && matchMetadata(ctx, stub) {
				return stub
			}
		case "partial":
			if firstPartialMatch != nil {
				continue // if a partial match is already found continue (so it takes the first match)
			}
			if stub.Request.Content.Matches(JsonString(requestJson)) && matchMetadata(ctx, stub) {
				firstPartialMatch = stub // Use the first partial match
			}
		}
	}
	return firstPartialMatch
}

func matchMetadata(ctx context.Context, stub *Stub) bool {
	if len(stub.Request.Metadata) == 0 {
		return true
	}
	// read metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	stubMetadata := getStubMetadata(stub)
	// compare
	for key, values := range stubMetadata {
		contextMetadata := md.Get(key)
		sort.Strings(contextMetadata)
		sort.Strings(values)
		if strings.Join(values, ",") != strings.Join(contextMetadata, ",") {
			return false
		}
	}
	return true
}

func getStubMetadata(stub *Stub) (stubMetadata map[string][]string) {
	stubMetadata = make(map[string][]string, 0)
	for key, values := range stub.Request.Metadata {
		for _, value := range values {
			stubMetadata[key] = append(stubMetadata[key], strings.TrimSpace(value))
		}
	}
	return
}
