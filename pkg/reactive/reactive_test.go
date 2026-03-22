package reactive

import "testing"

func resetReactiveState() {
	observerStack = nil
	activeObserver = nil
	batchMu.Lock()
	batchDepth = 0
	batchQueue = nil
	batchMu.Unlock()
}

func TestSignalGetSet(t *testing.T) {
	resetReactiveState()
	s := NewSignal(0)
	s.Set(42)
	if got := s.Get(); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestSignalPeek(t *testing.T) {
	resetReactiveState()
	s := NewSignal(7)
	if got := s.Peek(); got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestSignalZeroValue(t *testing.T) {
	resetReactiveState()
	s := NewSignal("")
	if got := s.Get(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSignalBool(t *testing.T) {
	resetReactiveState()
	s := NewSignal(false)
	s.Set(true)
	if got := s.Get(); !got {
		t.Errorf("expected true, got %v", got)
	}
}

func TestSignalMultipleReads(t *testing.T) {
	resetReactiveState()
	s := NewSignal(9)
	if got := s.Get(); got != 9 {
		t.Fatalf("expected first read to return 9, got %d", got)
	}
	if got := s.Get(); got != 9 {
		t.Errorf("expected second read to return 9, got %d", got)
	}
}

func TestEffectRunsImmediately(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		s.Get()
		count++
	})
	defer e.Dispose()

	if count != 1 {
		t.Errorf("expected effect to run once on creation, got %d", count)
	}
}

func TestEffectRerunsOnSet(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		s.Get()
		count++
	})
	defer e.Dispose()

	s.Set(1)

	if count != 2 {
		t.Errorf("expected effect to run twice, got %d", count)
	}
}

func TestEffectDispose(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		s.Get()
		count++
	})

	e.Dispose()
	s.Set(1)

	if count != 1 {
		t.Errorf("expected disposed effect not to re-run, got %d", count)
	}
}

func TestEffectMultipleDeps(t *testing.T) {
	resetReactiveState()
	count := 0
	a := NewSignal(1)
	b := NewSignal(2)
	e := NewEffect(func() {
		_ = a.Get()
		_ = b.Get()
		count++
	})
	defer e.Dispose()

	a.Set(3)
	b.Set(4)

	if count != 3 {
		t.Errorf("expected effect to run three times, got %d", count)
	}
}

func TestEffectNoRerunOnPeek(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		_ = s.Peek()
		count++
	})
	defer e.Dispose()

	s.Set(1)

	if count != 1 {
		t.Errorf("expected effect not to re-run when using Peek, got %d", count)
	}
}

func TestComputedBasic(t *testing.T) {
	resetReactiveState()
	a := NewSignal(3)
	c := NewComputed(func() int {
		return a.Get() * 2
	})

	if got := c.Get(); got != 6 {
		t.Errorf("expected 6, got %d", got)
	}
}

func TestComputedLazy(t *testing.T) {
	resetReactiveState()
	calls := 0
	_ = NewComputed(func() int {
		calls++
		return 1
	})

	if calls != 0 {
		t.Errorf("expected computed to be lazy, got %d calls", calls)
	}
}

func TestComputedCaches(t *testing.T) {
	resetReactiveState()
	calls := 0
	a := NewSignal(2)
	c := NewComputed(func() int {
		calls++
		return a.Get() * 2
	})

	if c.Get() != 4 {
		t.Fatalf("expected first get to return 4")
	}
	if c.Get() != 4 {
		t.Fatalf("expected second get to return 4")
	}
	if calls != 1 {
		t.Errorf("expected computed to cache result, got %d calls", calls)
	}
}

func TestComputedInvalidates(t *testing.T) {
	resetReactiveState()
	calls := 0
	a := NewSignal(2)
	c := NewComputed(func() int {
		calls++
		return a.Get() * 2
	})

	_ = c.Get()
	a.Set(5)

	if got := c.Get(); got != 10 {
		t.Errorf("expected invalidated computed to return 10, got %d", got)
	}
	if calls != 2 {
		t.Errorf("expected computed to re-run twice total, got %d", calls)
	}
}

func TestComputedChained(t *testing.T) {
	resetReactiveState()
	a := NewSignal(2)
	double := NewComputed(func() int {
		return a.Get() * 2
	})
	quad := NewComputed(func() int {
		return double.Get() * 2
	})

	if got := quad.Get(); got != 8 {
		t.Fatalf("expected 8, got %d", got)
	}

	a.Set(3)

	if got := quad.Get(); got != 12 {
		t.Errorf("expected chained computed to update to 12, got %d", got)
	}
}

func TestBatchDefersFire(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		s.Get()
		count++
	})
	defer e.Dispose()

	Batch(func() {
		s.Set(1)
		s.Set(2)
	})

	if count != 2 {
		t.Errorf("expected effect to run once initially and once after batch, got %d", count)
	}
}

func TestBatchNestedNoop(t *testing.T) {
	resetReactiveState()
	count := 0
	s := NewSignal(0)
	e := NewEffect(func() {
		s.Get()
		count++
	})
	defer e.Dispose()

	Batch(func() {
		s.Set(1)
		Batch(func() {
			s.Set(2)
		})
	})

	if count != 2 {
		t.Errorf("expected nested batch to flush once, got %d runs", count)
	}
}
