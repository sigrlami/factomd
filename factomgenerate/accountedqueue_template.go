//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and 'ᐳ' so ᐸ_typenameᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

//ᐸif false ᐳ
package Dummy // this is only here to make gofmt happy and is never in the generated code
//ᐸendᐳ

//ᐸdefine "accountedqueue-imports"ᐳ
import (
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/telemetry"
)

//ᐸendᐳ

//ᐸdefine "accountedqueue"ᐳ
// Start accountedqueue generated go code

type ᐸ_typenameᐳ struct {
	common.Name
	Channel chan ᐸ_typeᐳ
}

func (q *ᐸ_typenameᐳ) Init(parent common.NamedObject, name string, size int) *ᐸ_typenameᐳ {
	q.Name.Init(parent, name)
	q.Channel = make(chan ᐸ_typeᐳ, size)
	return q
}

// construct gauge w/ proper labels
func (q *ᐸ_typenameᐳ) Metric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.GetPath(), "current")
}

// construct counter for tracking totals
func (q *ᐸ_typenameᐳ) TotalMetric() telemetry.Counter {
	return telemetry.TotalCounter.WithLabelValues(q.GetPath(), "total")
}

// Length of underlying channel
func (q ᐸ_typenameᐳ) Length() int {
	return len(q.Channel)
}

// Cap of underlying channel
func (q ᐸ_typenameᐳ) Cap() int {
	return cap(q.Channel)
}

// Enqueue adds item to channel and instruments based on type
func (q ᐸ_typenameᐳ) Enqueue(m ᐸ_typeᐳ) {
	q.Channel <- m
	q.TotalMetric().Inc()
	q.Metric().Inc()
}

// Enqueue adds item to channel and instruments based on
// returns true it it enqueues the data
func (q ᐸ_typenameᐳ) EnqueueNonBlocking(m ᐸ_typeᐳ) bool {
	select {
	case q.Channel <- m:
		q.TotalMetric().Inc()
		q.Metric().Inc()
		return true
	default:
		return false
	}
}

// Dequeue removes an item from channel
// Returns nil if nothing in // queue
func (q ᐸ_typenameᐳ) Dequeue() ᐸ_typeᐳ {
	select {
	case v := <-q.Channel:
		q.Metric().Dec()
		return v
	default:
		return nil
	}
}

// Dequeue removes an item from channel
func (q ᐸ_typenameᐳ) BlockingDequeue() ᐸ_typeᐳ {
	v := <-q.Channel
	q.Metric().Dec()
	return v
}

// End accountedqueue generated go code
// ᐸend ᐳ
