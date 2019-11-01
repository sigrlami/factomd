//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and 'ᐳ' so ᐸ_typenameᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

package generated_test // this is only here to make gofmt happy and is never in the generated code

//go:generate go run ./generate.go

//ᐸdefine "accountedqueue_test-imports"ᐳ
import (
	"testing"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/generated"
)

//ᐸendᐳ

//ᐸdefine "accountedqueue_test"ᐳ
// Start accountedqueue_test generated go code

func TestAccountedQueue(t *testing.T) {
	q := new(generated.ᐸ_typenameᐳ).Init(common.NilName, "Test", 10)

	if q.Dequeue() != nil {
		t.Fatal("empty dequeue return non-nil")
	}

	for i := 0; i < 10; i++ {
		q.Enqueue(ᐸ_testelementᐳ)
	}

	// commented out because it requires a modern prometheus package
	//if testutil.ToFloat64(q.TotalMetric()) != float64(10) {
	//	t.Fatal("TotalMetric fail")
	//}

	for i := 9; i >= 0; i-- {
		q.Dequeue()
		// commented out because it requires a modern prometheus package
		//if testutil.ToFloat64(q.Metric()) != float64(i) {
		//	t.Fatal("Metric fail")
		//}
	}

	if q.Dequeue() != nil {
		t.Fatal("empty dequeue return non-nil")
	}
}

//ᐸendᐳ
