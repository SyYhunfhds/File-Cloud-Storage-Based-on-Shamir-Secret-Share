package visualization

import "time"

type Animation struct {
	startTime time.Time
	duration  time.Duration
	loop     bool
	reverse  bool
}

func NewAnimation(duration time.Duration, loop bool) *Animation {
	return &Animation{
		startTime: time.Now(),
		duration:  duration,
		loop:      loop,
		reverse:   false,
	}
}

func (a *Animation) Update() {
}

func (a *Animation) GetProgress() float64 {
	elapsed := time.Since(a.startTime)
	progress := float64(elapsed) / float64(a.duration)

	if a.reverse {
		progress = 1 - progress
	}

	if a.loop {
		progress = progress - float64(int(progress))
		if progress < 0 {
			progress += 1
		}
	}

	if progress > 1 {
		progress = 1
	}
	if progress < 0 {
		progress = 0
	}

	return progress
}

func (a *Animation) IsFinished() bool {
	elapsed := time.Since(a.startTime)
	return elapsed >= a.duration
}

func (a *Animation) Reset() {
	a.startTime = time.Now()
}

func (a *Animation) SetReverse(reverse bool) {
	a.reverse = reverse
}

func EaseInQuad(t float64) float64 {
	return t * t
}

func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func LerpFloat64(a, b, t float64) float64 {
	return a + (b-a)*t
}

func LerpInt64(a, b, t float64) int64 {
	return int64(LerpFloat64(float64(a), float64(b), t))
}
