package object

func (e *Environment) NewEnclosedEnvironment() *Environment {
	env := newEnvironment()
	env.outer = e
	return env
}

func newEnvironment() *Environment {
	s := make(map[string]*Object)
	return &Environment{store: s, outer: nil}
}

func NewEnvironment(bases map[string]Object) *Environment {
	o := newEnvironment()
	for s, base := range bases {
		o.SetCurrent(s, base)
	}

	s := make(map[string]*Object)
	return &Environment{store: s, outer: o}
}

type Environment struct {
	store map[string]*Object
	outer *Environment
}

func (e *Environment) Inspect(num int) string { return "[ENV]" }
func (e *Environment) Type() Type      { return ENVIRONMENT }
func (e *Environment) Copy() Object    { return e }

func (e *Environment) Get(name string) (*Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) DoAlloc(Index Object) (*Object, bool) {
	if envIndex, ok := Index.(*String); ok {
		if _, ok := e.store[string(envIndex.Value)]; ok {
			return nil, false
		}
		var obj Object = nil
		e.store[string(envIndex.Value)] = &obj
		return &obj, true
	}
	return nil, false
}

func (e *Environment) SetCurrent(name string, val Object) (*Object, bool) {
	if ptr, ok := e.DoAlloc(&String{Value: []rune(name)}); ok {
		*ptr = val
		return ptr, true
	}
	return nil, false
}

func (e *Environment) DeAlloc(Index Object) bool {
	if envIndex, ok := Index.(*String); ok {
		_, ok := e.store[string(envIndex.Value)]
		if ok {
			delete(e.store, string(envIndex.Value))
			return true
		}
		if !ok && e.outer != nil {
			return e.outer.DeAlloc(envIndex)
		}
	}
	return false
}
