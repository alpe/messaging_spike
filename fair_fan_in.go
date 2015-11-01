package messaging_spike

// any payload type
type Payload string

// FanIn merges given input channels into a new one.
// In opposite to concurrent consumption fairness is provided by round robin consumption.
// FanIn is not blocking on empty channels but continues with the next. When all input channels
// are closed, the output channel will be closed, too.
func FanIn(in []chan Payload, out chan<- Payload) {
	for len(in) > 0 {
		for i, c := range in {
			select {
			case v, ok := <-c:
				if !ok { // remove when closed
					in = append(in[:i], in[i+1:]...)
					continue
				}
				out <- v // may block on slow consumers
			default: // don't block when no message
			}
		}
	}
	close(out)
}
