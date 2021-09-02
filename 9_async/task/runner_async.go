package task

import (
	"os"
	"os/signal"
	"time"
)

type RunnerAsync struct {
	interrupt chan os.Signal
	complete  chan error
	timeout   <-chan time.Time
	tasks     []func(id int)
}

func NewRunnerAsync(d time.Duration) *RunnerAsync {
	return &RunnerAsync{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

func (this *RunnerAsync) Add(task ...func(id int)) {
	this.tasks = append(this.tasks, task...)
}

func (this *RunnerAsync) Start() error {
	signal.Notify(this.interrupt, os.Interrupt)

	go func() {
		this.complete <- this.Run()
	}()

	select {
	case err := <-this.complete:
		return err

	case <-this.timeout:
		return ErrTimeout
	}
}

func (this *RunnerAsync) Run() error {
	for id, task := range this.tasks {
		if this.gotInterrupt() {
			return ErrInterrupt
		}

		task(id)
	}

	return nil
}

func (this *RunnerAsync) gotInterrupt() bool {
	select {
	case <-this.interrupt:
		signal.Stop(this.interrupt)
		return true

	default:
		return false
	}
}
