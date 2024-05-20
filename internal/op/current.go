package op

import (
	"sync"
	"time"
)

type current struct {
	current Current
	lock    sync.RWMutex
}

type Current struct {
	Movie  CurrentMovie
	Status Status
}

type CurrentMovie struct {
	ID     string
	IsLive bool
}

func newCurrent() *current {
	return &current{
		current: Current{
			Status: newStatus(),
		},
	}
}

type Status struct {
	Seek       float64   `json:"seek"`
	Rate       float64   `json:"rate"`
	Playing    bool      `json:"playing"`
	lastUpdate time.Time `json:"-"`
}

func newStatus() Status {
	return Status{
		Seek:       0,
		Rate:       1.0,
		lastUpdate: time.Now(),
	}
}

func (c *current) Current() Current {
	c.lock.RLock()
	defer c.lock.RUnlock()
	c.current.UpdateStatus()
	return c.current
}

func (c *current) SetMovie(movie CurrentMovie, play bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.current.Movie = movie
	c.current.SetSeek(0, 0)
	c.current.Status.Playing = play
}

func (c *current) Status() Status {
	c.lock.RLock()
	defer c.lock.RUnlock()
	c.current.UpdateStatus()
	return c.current.Status
}

func (c *current) SetStatus(playing bool, seek, rate, timeDiff float64) *Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := c.current.SetStatus(playing, seek, rate, timeDiff)
	return &s
}

func (c *current) SetSeekRate(seek, rate, timeDiff float64) *Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := c.current.SetSeekRate(seek, rate, timeDiff)
	return &s
}

func (c *Current) UpdateStatus() Status {
	if c.Movie.IsLive {
		c.Status.lastUpdate = time.Now()
		return c.Status
	}
	if c.Status.Playing {
		c.Status.Seek += time.Since(c.Status.lastUpdate).Seconds() * c.Status.Rate
	}
	c.Status.lastUpdate = time.Now()
	return c.Status
}

func (c *Current) setLiveStatus() Status {
	c.Status.Playing = true
	c.Status.Rate = 1.0
	c.Status.Seek = 0
	c.Status.lastUpdate = time.Now()
	return c.Status
}

func (c *Current) SetStatus(playing bool, seek, rate, timeDiff float64) Status {
	if c.Movie.IsLive {
		return c.setLiveStatus()
	}
	c.Status.Playing = playing
	c.Status.Rate = rate
	if playing {
		c.Status.Seek = seek + (timeDiff * rate)
	} else {
		c.Status.Seek = seek
	}
	c.Status.lastUpdate = time.Now()
	return c.Status
}

func (c *Current) SetSeekRate(seek, rate, timeDiff float64) Status {
	if c.Movie.IsLive {
		return c.setLiveStatus()
	}
	if c.Status.Playing {
		c.Status.Seek = seek + (timeDiff * rate)
	} else {
		c.Status.Seek = seek
	}
	c.Status.Rate = rate
	c.Status.lastUpdate = time.Now()
	return c.Status
}

func (c *Current) SetSeek(seek, timeDiff float64) Status {
	if c.Movie.IsLive {
		return c.setLiveStatus()
	}
	if c.Status.Playing {
		c.Status.Seek = seek + (timeDiff * c.Status.Rate)
	} else {
		c.Status.Seek = seek
	}
	c.Status.lastUpdate = time.Now()
	return c.Status
}
