// utility/typemanager.go
package Utility

import (
	"reflect"
	"sync"
)

// TypeManager provides concurrent-safe registries for types and functions.
type TypeManager struct {
	mu               sync.RWMutex
	typeRegistry     map[string]reflect.Type
	functionRegistry map[string]interface{}
}

// NewTypeManager creates a new, empty manager.
func NewTypeManager() *TypeManager {
	return &TypeManager{
		typeRegistry:     make(map[string]reflect.Type),
		functionRegistry: make(map[string]interface{}),
	}
}

// RegisterType registers a type under a name (overwrites if already present).
func (tm *TypeManager) RegisterType(name string, t reflect.Type) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.typeRegistry[name] = t
}

// GetType returns a type and a boolean indicating if it exists.
func (tm *TypeManager) GetType(name string) (reflect.Type, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.typeRegistry[name]
	return t, ok
}

// RegisterInstance registers the dynamic type of a non-nil instance under a name.
func (tm *TypeManager) RegisterInstance(name string, instance interface{}) {
	if instance == nil {
		return
	}
	tm.RegisterType(name, reflect.TypeOf(instance))
}

// RegisterFunc registers a callable under a name (overwrites if already present).
func (tm *TypeManager) RegisterFunc(name string, fn interface{}) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.functionRegistry[name] = fn
}

// GetFunc returns a function and a boolean indicating if it exists.
func (tm *TypeManager) GetFunc(name string) (interface{}, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	f, ok := tm.functionRegistry[name]
	return f, ok
}

// DeleteType removes a type by name (no-op if not present).
func (tm *TypeManager) DeleteType(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.typeRegistry, name)
}

// DeleteFunc removes a function by name (no-op if not present).
func (tm *TypeManager) DeleteFunc(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.functionRegistry, name)
}

// ListTypes returns a snapshot of registered type names.
func (tm *TypeManager) ListTypes() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	keys := make([]string, 0, len(tm.typeRegistry))
	for k := range tm.typeRegistry {
		keys = append(keys, k)
	}
	return keys
}

// ListFuncs returns a snapshot of registered function names.
func (tm *TypeManager) ListFuncs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	keys := make([]string, 0, len(tm.functionRegistry))
	for k := range tm.functionRegistry {
		keys = append(keys, k)
	}
	return keys
}

// -----------------------------------------------------------------------------
// Singleton accessors (replaces the original package-level getTypeManager()).
// -----------------------------------------------------------------------------

var (
	typeManagerOnce sync.Once
	typeManager     *TypeManager
)

// getTypeManager is kept (lowercase) to mirror original code’s API surface.
func getTypeManager() *TypeManager {
	typeManagerOnce.Do(func() {
		typeManager = NewTypeManager()
	})
	return typeManager
}

// DefaultTypeManager is the public-friendly accessor if you prefer exported API.
func DefaultTypeManager() *TypeManager { return getTypeManager() }

// -----------------------------------------------------------------------------
// Back-compat shims (retain original method names/signatures).
// They now delegate to the concurrent-safe implementation above.
// -----------------------------------------------------------------------------

func (tm *TypeManager) getType(name string) (t reflect.Type, exist bool) {
	return tm.GetType(name)
}

func (tm *TypeManager) setType(name string, val reflect.Type) {
	tm.RegisterType(name, val)
}

func (tm *TypeManager) getFunction(name string) interface{} {
	if f, ok := tm.GetFunc(name); ok {
		return f
	}
	return nil
}

func (tm *TypeManager) setFunction(name string, val interface{}) {
	tm.RegisterFunc(name, val)
}

