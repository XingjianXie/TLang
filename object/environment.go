package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]*Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]*Object
	outer *Environment
}

func (e *Environment) Get(name string) (*Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) SetCurrent(name string, val Object) (*Object, bool) {
	_, ok := e.store[name]
	if ok {
		return nil, false
	}
	e.store[name] = &val
	return &val, true
}

func (e *Environment) SetAvailable(name string, val Object) (*Object, bool) {
	_, ok := e.store[name]
	if ok {
		*e.store[name] = val
		return &val, ok
	}
	if e.outer != nil {
		return e.outer.SetAvailable(name, val)
	}
	return nil, false
}

func (e *Environment) Del(name string) {
	_, ok := e.store[name]
	if ok {
		delete(e.store, name)
		return
	}
	if !ok && e.outer != nil {
		e.outer.Del(name)
	}
}
