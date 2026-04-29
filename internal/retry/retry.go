package retry

import "time"

var delays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func Do(operation func() error, isRetriable func(error) bool) error {
	err := operation()
	if err == nil {
		return nil
	}

	if !isRetriable(err) {
		return err
	}

	for _, delay := range delays {
		time.Sleep(delay)

		err = operation()
		if err == nil {
			return nil
		}

		if !isRetriable(err) {
			return err
		}
	}

	return err
}
