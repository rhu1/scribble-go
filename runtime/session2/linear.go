package session2

import "sync"

type LinearResource struct {
	sync.Mutex
	used bool
}

func (res *LinearResource) Use() {
	res.Lock()
	defer res.Unlock()
	if res.used {
		panic("Linear resource already used.")
	}
	res.used = true
}
