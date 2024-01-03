package utils

type Set map[interface{}]bool

func (s Set) Add(element interface{}) {
	s[element] = true
}

func (s Set) Remove(element interface{}) {
	delete(s, element)
}

func (s Set) Contains(element interface{}) bool {
	return s[element]
}
