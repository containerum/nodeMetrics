package influx

import (
	"encoding/json"
	"time"

	"math"

	"github.com/containerum/nodeMetrics/pkg/vector"
	"github.com/sirupsen/logrus"
)

const (
	MEM_COEFF = 1e12
)

func (flux *Influx) MemoryCurrent() (float64, error) {
	var result, err = flux.Query("SELECT MEAN(value) FROM memory_usage WHERE time > now()-5m")
	if err != nil {
		return 0, err
	}
	if len(result) < 1 {
		return 0, ErrEmptyResult
	}
	if len(result[0].Series) < 1 {
		return 0, ErrNoSeriesFound
	}
	if len(result) < 1 {
		return 0, ErrEmptyResult
	}
	if len(result[0].Series) < 1 {
		return 0, ErrNoSeriesFound
	}
	if len(result[0].Series[0].Values) < 1 {
		return 0, ErrNoValuesFound
	}
	if len(result[0].Series[0].Values[0]) < 2 {
		return 0, ErrInvalidDataPointFormat
	}
	var average, _ = result[0].Series[0].Values[0][1].(json.Number).Float64()
	average /= MEM_COEFF * flux.MemoryFactor() // TODO: remove hardcoded value
	return average, nil
}

func (flux *Influx) MemoryHistory(from, to time.Time, step time.Duration) (vector.Vec, error) {
	logrus.Debugf("getting MEMORY history from Influx")
	defer logrus.Debugf("end")
	var result, err = flux.Query("SELECT MEAN(value) FROM memory_usage WHERE time > %d AND time < %d GROUP BY TIME(%v)", from.UnixNano(), to.UnixNano(), step)
	if err != nil {
		return nil, err
	}
	if len(result) < 1 {
		logrus.Error(ErrEmptyResult)
		return vector.Vec{}, nil
	}
	if len(result[0].Series) < 1 {
		logrus.Error(ErrNoSeriesFound)
		return vector.Vec{}, nil
	}
	logrus.Debugf("parsing result")
	var values = result[0].Series[0].Values
	var history = vector.MakeVec(len(values), func(index int) float64 {
		var point = values[index]
		if len(point) < 2 {
			logrus.Errorf("invalid data point in InfluxDB response: expected >= 2 columns, got %q", point)
			return 0
		}
		switch point := point[1].(type) {
		case int:
			return float64(point)
		case float64:
			return point
		case json.Number:
			var x, err = point.Float64()
			if err != nil {
				return 0
			}
			return x
		default:
			return 0
		}
	}).DivideScalar(MEM_COEFF * flux.MemoryFactor()).
		MulScalar(100).
		Map(math.Round) // TODO: remove hardcoded value
	return history, nil
}
