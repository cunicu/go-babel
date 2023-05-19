package table_test

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/go-babel/internal/table"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table suite")
}

var _ = Describe("Table", func() {
	var t table.Table[int, int]

	BeforeEach(func() {
		t.Clear()
	})

	It("can insert", func() {
		t.Insert(4, 5)
		Expect(t.Len()).To(Equal(1))

		val, ok := t.Lookup(4)
		Expect(ok).To(BeTrue())
		Expect(val).To(Equal(5))
	})

	It("can lookup non-existing key", func() {
		_, ok := t.Lookup(100)
		Expect(ok).To(BeFalse())
	})

	It("can clear", func() {
		t.Insert(6, 7)
		val, ok := t.Lookup(6)
		Expect(ok).To(BeTrue())
		Expect(val).To(Equal(7))

		t.Clear()
		Expect(t.Empty()).To(BeTrue())
	})

	It("Empty", func() {
		Expect(t.Empty()).To(BeTrue())

		t.Insert(1, 1)

		Expect(t.Empty()).To(BeFalse())
	})

	It("Len", func() {
		Expect(t.Len()).To(Equal(0))

		t.Insert(1, 1)
		Expect(t.Len()).To(Equal(1))

		t.Insert(1, 1)
		Expect(t.Len()).To(Equal(1))

		t.Insert(2, 1)
		Expect(t.Len()).To(Equal(2))
	})

	It("can update", func() {
		t.Insert(1, 1)

		val, ok := t.Lookup(1)
		Expect(ok).To(BeTrue())
		Expect(val).To(Equal(1))

		m := map[int]int{
			1: 100,
		}

		t.Update(m)

		Expect(t.Len()).To(Equal(1))

		val, ok = t.Lookup(1)
		Expect(ok).To(BeTrue())
		Expect(val).To(Equal(100))
	})

	It("can iterate", func() {
		m := map[int]int{
			1: 100,
			2: 200,
			3: 300,
		}

		t.Update(m)

		f := map[int]int{}

		err := t.Foreach(func(k, v int) error {
			f[k] = v
			return nil
		})
		Expect(err).To(Succeed())

		Expect(f).To(Equal(m))
	})

	It("aborts iteration on error", func() {
		m := map[int]int{
			1: 100,
			2: 200,
			3: 300,
		}

		e := map[int]int{
			1: 100,
			2: 200,
		}

		t.Update(m)

		f := map[int]int{}

		errAbort := errors.New("abort here")

		err := t.Foreach(func(k, v int) error {
			if k == 3 {
				return errAbort
			}
			f[k] = v
			return nil
		})
		Expect(err).To(MatchError(errAbort))

		Expect(f).To(Equal(e))
	})
})
