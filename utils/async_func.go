package utils

// RunTaskAsync transform simple func into async like function that return bool chan
func RunTaskAsync(handler func()) <-chan bool {
	r := make(chan bool)

	go func() {
		defer close(r)

		// Simulate a workload.
		handler()

		r <- true
	}()

	return r
}
