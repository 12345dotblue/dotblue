package engine

import "fmt"

var (
	defaultRuntime Runtime
	engines        = make(map[string]Engine)
)

// RegisterEngine registers an engine implementation by name.
func RegisterEngine(e Engine) {
	engines[e.Name()] = e
}

// GetEngine returns a registered engine by name.
func GetEngine(name string) (Engine, error) {
	e, ok := engines[name]
	if !ok {
		return nil, fmt.Errorf("engine not registered: %s", name)
	}
	return e, nil
}

// SetRuntime sets the global runtime implementation.
func SetRuntime(r Runtime) {
	defaultRuntime = r
}

// GetRuntime returns the global runtime implementation.
func GetRuntime() Runtime {
	return defaultRuntime
}

// Init initializes the default runtime and registers built-in engines.
func Init() {
	RegisterEngine(&HermesEngine{})
	SetRuntime(&DockerRuntime{})
}
