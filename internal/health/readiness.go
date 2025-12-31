package health

import (
	"context"
	"sync"
	"time"
)

// CheckStatus represents the status of an individual readiness check.
type CheckStatus string

const (
	CheckStatusOK   CheckStatus = "ok"
	CheckStatusFail CheckStatus = "fail"
)

// SummaryStatus represents the overall readiness status.
type SummaryStatus string

const (
	SummaryStatusOK       SummaryStatus = "ok"
	SummaryStatusNotReady SummaryStatus = "not_ready"
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
	Name     string      `json:"name"`
	Status   CheckStatus `json:"status"`
	Critical bool        `json:"critical"`
	Error    string      `json:"error,omitempty"`
}

type Summary struct {
	Ready  bool          `json:"-"`
	Status SummaryStatus `json:"status"`
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
	out := Summary{
		Ready:  true,
		Status: SummaryStatusOK,
		Checks: make([]CheckResult, len(a.checks)),
	}

	// Policy: if no readiness checks are registered, the service is NOT ready.
	if len(a.checks) == 0 {
		out.Ready = false
		out.Status = SummaryStatusNotReady
		out.Checks = []CheckResult{}
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(a.checks))

	for idx, check := range a.checks {
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
				res.Status = CheckStatusFail
				res.Error = err.Error()
			} else {
				res.Status = CheckStatusOK
			}
			out.Checks[idx] = res
		}()
	}

	wg.Wait()

	// Aggregate readiness: only critical failures make it not ready.
	for _, r := range out.Checks {
		if r.Status == CheckStatusFail && r.Critical {
			out.Ready = false
			out.Status = SummaryStatusNotReady
			break
		}
	}

	return out
}
