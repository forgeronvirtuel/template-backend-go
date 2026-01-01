package metrics

import "testing"

// TestNoopRecorder verifies that the no-op recorder is safe to use
// and does not panic when methods are called.
func TestNoopRecorder(t *testing.T) {
	recorder := NewNoop()

	// These should not panic
	recorder.IncCounter("test_counter", Labels{{Key: "key", Value: "value"}}, 1)
	recorder.ObserveHistogram("test_histogram", Labels{{Key: "key", Value: "value"}}, 1.5)
	recorder.SetGauge("test_gauge", Labels{{Key: "key", Value: "value"}}, 10.0)
	recorder.AddGauge("test_gauge", Labels{{Key: "key", Value: "value"}}, 5.0)
}

// TestNoopRecorderNilLabels verifies that nil labels are safe.
func TestNoopRecorderNilLabels(t *testing.T) {
	recorder := NewNoop()

	// These should not panic with nil labels
	recorder.IncCounter("test_counter", nil, 1)
	recorder.ObserveHistogram("test_histogram", nil, 1.5)
	recorder.SetGauge("test_gauge", nil, 10.0)
	recorder.AddGauge("test_gauge", nil, 5.0)
}

// TestLabels verifies Labels type behavior.
func TestLabels(t *testing.T) {
	labels := Labels{
		{Key: "method", Value: "GET"},
		{Key: "route", Value: "/users"},
		{Key: "status", Value: "200"},
	}

	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}

	// Helper to find label by key
	findLabel := func(key string) string {
		for _, l := range labels {
			if l.Key == key {
				return l.Value
			}
		}
		return ""
	}

	if method := findLabel("method"); method != "GET" {
		t.Errorf("expected method=GET, got %s", method)
	}

	if route := findLabel("route"); route != "/users" {
		t.Errorf("expected route=/users, got %s", route)
	}

	if status := findLabel("status"); status != "200" {
		t.Errorf("expected status=200, got %s", status)
	}
}
