package testserv

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// An Instruction tells the Handler how to serve an HTTP response.
type Instruction struct {
	// HeaderDelay tells the Handler how long to pause before returning the
	// response headers.
	HeaderDelay time.Duration
	// Header tells the Handler what HTTP headers to return.
	//
	// If Header is missing the "Content-Length" header, and Body is not nil,
	// then the "Content-Length" header will be added and set to the length
	// of Body.
	Header http.Header
	// StatusCode tells the Handler what HTTP status code to return.
	StatusCode int
	// BodyDelay tells the Handler how long to pause before beginning to write
	// the response body.
	BodyDelay time.Duration
	// BodyServiceTime tells the Handler how long it should take to transmit the
	// entire response body.
	BodyServiceTime time.Duration
	// Body tells the Handler what to return for the response body.
	//
	// The Handler will spend at least BodyServiceTime writing the response body
	// bytes, flushing the response writer after each byte written and allocating
	// a proportional amount of the body service time to each byte written.
	Body []byte
}

// A Handler implements a Go standard http.Handler for handling HTTP requests
// within an http.Server. The Handler's instructions tell it how to serve each
// incoming request.
type Handler struct {
	// N is the number of requests served so far.
	N int

	// Inst contains the instructions for serving requests.
	//
	// When N is less than the length of Inst, the next request will be served
	// using the instruction in Inst[N]. Once N reaches the length of Inst, all
	// subsequent requests are served a 400 Bad Request response.
	Inst []Instruction

	lock sync.Mutex
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	n := h.N
	h.N++
	h.lock.Unlock()

	// Handle running out of instructions.
	if n >= len(h.Inst) {
		body := []byte(fmt.Sprintf("Out of instructions: N[%d] >= len(Inst)[%d]", n, len(h.Inst)))
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(400)
		_, _ = w.Write(body)
		return
	}

	// Current instruction.
	inst := h.Inst[n]

	// Parley the response writer into a flusher. We use the flusher so we can
	// flush work in progress down to the network stack as soon as it is ready,
	// to ensure the client sees a steady stream of data rather than a big bang
	// at the end.
	f, ok := w.(http.Flusher)
	if !ok {
		panic("testserv: w is not a Flusher")
	}

	// Pause for the delay time specified before writing the header.
	time.Sleep(inst.HeaderDelay)

	// Copy the instruction header to the response header.
	hdr := w.Header()
	for k, v := range inst.Header {
		hdr[k] = v
	}

	// Insert a content-length field into the header if needed.
	body := inst.Body
	if body != nil {
		if _, has := inst.Header["Content-Length"]; !has {
			hdr.Set("Content-Length", strconv.Itoa(len(body)))
		}
	}

	// Write the header and flush.
	w.WriteHeader(inst.StatusCode)
	f.Flush()

	// Pause for the delay time specified before writing the body.
	time.Sleep(inst.BodyDelay)

	// Write the body one byte at a time, pausing and flushing between each byte
	// and overall ensuring the time taken to write the body takes the specified
	// service duration.
	serviceTime := inst.BodyServiceTime
	bytePause := serviceTime / time.Duration(len(body))
	serviceStart := time.Now()
	for i := 0; i < len(body)-1; i++ {
		time.Sleep(bytePause)
		_, err := w.Write(body[i : i+1])
		if err != nil {
			return
		}
		f.Flush()
	}
	time.Sleep(serviceTime - time.Now().Sub(serviceStart))
	_, _ = w.Write(body[len(body)-1:])
}
