package main

import "time"

type Fps struct {
	data   [10]*fpsData
	offset int
}

type fpsData struct {
	time       int64
	frameCount uint64
}

func (f *Fps) Add(frameCount uint64) {
	datum := &fpsData{time.Now().UnixNano(), frameCount}
	f.data[f.offset] = datum
	f.offset = (f.offset + 1) % len(f.data)
}

func (f *Fps) Current() float64 {
	var minDatum, maxDatum *fpsData

	for _, datum := range f.data {
		if datum != nil {
			if minDatum == nil || datum.time < minDatum.time {
				minDatum = datum
			}
			if maxDatum == nil || datum.time > maxDatum.time {
				maxDatum = datum
			}
		}
	}

	if minDatum == nil || maxDatum == nil || minDatum == maxDatum {
		return 0
	} else {
		return float64(maxDatum.frameCount-minDatum.frameCount) / (float64(maxDatum.time-minDatum.time) / 1000000000)
	}
}
