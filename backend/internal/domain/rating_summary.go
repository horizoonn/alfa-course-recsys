package domain

import (
	"errors"
	"math"
)

type RatingSummary struct {
	Average float64
	Count   *int64
}

func (r RatingSummary) Validate() error {
	if math.IsNaN(r.Average) || math.IsInf(r.Average, 0) || r.Average < 1 || r.Average > 5 {
		return errors.New("rating average must be finite and between 1 and 5")
	}
	if r.Count != nil && *r.Count <= 0 {
		return errors.New("rating count must be positive or nil")
	}
	return nil
}
