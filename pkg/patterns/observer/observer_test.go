package observer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type event1 struct {
	EventType
	idata int
}

type event2 struct {
	EventType
	sdata string
}

func Test_Observer(t *testing.T) {
	subj := NewSubject()
	var idata int
	var sdata string
	rc := func(et EventType) {
		switch val := et.(type) {
		case event1:
			idata = val.idata
		case event2:
			sdata = val.sdata
		}
	}
	obs1 := NewObserver(rc, false, event1{})
	obs2 := NewObserver(rc, false, event2{})
	subj.ObserversAttach(obs1, obs2)
	subj.Notify(event1{idata: 100500}, event2{sdata: "100500"})
	assert.Equal(t, 100500, idata)
	assert.Equal(t, "100500", sdata)
}

func Test_ObserverAsync(t *testing.T) {
	subj := NewSubject()
	var idata int
	var sdata string
	var mx sync.Mutex
	cv := sync.NewCond(&mx)
	i := 0
	rc := func(et EventType) {
		switch val := et.(type) {
		case event1:
			idata = val.idata
		case event2:
			sdata = val.sdata
		}
		mx.Lock()
		defer mx.Unlock()
		i++
		cv.Signal()
	}
	obs1 := NewObserver(rc, true, event1{}, event2{})
	subj.ObserversAttach(obs1)
	subj.Notify(event1{idata: 100500}, event2{sdata: "100500"})
	mx.Lock()
	for i < 2 {
		cv.Wait()
	}
	mx.Unlock()
	assert.Equal(t, 100500, idata)
	assert.Equal(t, "100500", sdata)
}
