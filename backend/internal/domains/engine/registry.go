package engine

import "fmt"

type Registry struct {
	defaultRuntime Runtime
	engines        map[string]Engine
}

func NewRegistry() *Registry {
	return &Registry{
		engines: make(map[string]Engine),
	}
}

func (r *Registry) RegisterEngine(e Engine) {
	if r == nil || e == nil {
		return
	}
	r.engines[e.Name()] = e
}

func (r *Registry) GetEngine(name string) (Engine, error) {
	if r == nil {
		return nil, fmt.Errorf("engine registry is not configured")
	}
	e, ok := r.engines[name]
	if !ok {
		return nil, fmt.Errorf("engine not registered: %s", name)
	}
	return e, nil
}

func (r *Registry) SetRuntime(runtime Runtime) {
	if r == nil {
		return
	}
	r.defaultRuntime = runtime
}

func (r *Registry) GetRuntime() Runtime {
	if r == nil {
		return nil
	}
	return r.defaultRuntime
}

var defaultRegistry = NewRegistry()

// RegisterEngine registers an engine implementation by name.
func RegisterEngine(e Engine) {
	defaultRegistry.RegisterEngine(e)
}

// GetEngine returns a registered engine by name.
func GetEngine(name string) (Engine, error) {
	return defaultRegistry.GetEngine(name)
}

// SetRuntime sets the global runtime implementation.
func SetRuntime(r Runtime) {
	defaultRegistry.SetRuntime(r)
}

// GetRuntime returns the global runtime implementation.
func GetRuntime() Runtime {
	return defaultRegistry.GetRuntime()
}

// Init initializes the default runtime and registers built-in engines.
func Init() {
	RegisterEngine(&HermesEngine{})
	RegisterEngine(&NanobotEngine{})
	SetRuntime(NewDockerRuntime())
}
