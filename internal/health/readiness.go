package health

import (
	"context"
	"sync"
	"time"
)

// ReadinessCheck represents a dependency check used by the /ready endpoint.
// Critical checks failing make the service NOT ready (HTTP 503).
// Non-critical checks failing do not block readiness (HTTP 200).
type ReadinessCheck interface {
	Name() string
	Critical() bool
	Check(ctx context.Context) error
}

type CheckResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // "ok" | "fail"
	Critical bool   `json:"critical"`
	Error    string `json:"error,omitempty"`
}

type Summary struct {
	Ready  bool          `json:"-"`
	Status string        `json:"status"` // "ok" | "not_ready"
	Checks []CheckResult `json:"checks"`
}

// Aggregator runs readiness checks.
// It is deterministic: results keep the same order as the checks slice.
type Aggregator struct {
	checks          []ReadinessCheck
	perCheckTimeout time.Duration
}

func NewReadinessAggregator(checks []ReadinessCheck, perCheckTimeout time.Duration) *Aggregator {
	if perCheckTimeout <= 0 {
		perCheckTimeout = 1 * time.Second
	}
	copied := make([]ReadinessCheck, len(checks))
	copy(copied, checks)

	return &Aggregator{
		checks:          copied,
		perCheckTimeout: perCheckTimeout,
	}
}

func (a *Aggregator) Run(ctx context.Context) Summary {
	// Default to ready when there are no checks.
	out := Summary{
		Ready:  true,
		Status: "ok",
		Checks: make([]CheckResult, len(a.checks)),
	}
	if len(a.checks) == 0 {
		out.Checks = []CheckResult{}
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(a.checks))

	for idx, check := range a.checks {
		idx := idx
		check := check

		go func() {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, a.perCheckTimeout)
			defer cancel()

			err := check.Check(checkCtx)
			res := CheckResult{
				Name:     check.Name(),
				Critical: check.Critical(),
			}
			if err != nil {
				res.Status = "fail"
				res.Error = err.Error()
			} else {
				res.Status = "ok"
			}
			out.Checks[idx] = res
		}()
	}

	wg.Wait()

	// Aggregate readiness: only critical failures make it not ready.
	for _, r := range out.Checks {
		if r.Status == "fail" && r.Critical {
			out.Ready = false
			out.Status = "not_ready"
			break
		}
	}

	return out
}
