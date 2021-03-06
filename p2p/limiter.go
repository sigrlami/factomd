package p2p

import (
	"fmt"
	"net"
	"time"
)

// LimitedListener will block multiple connection attempts from a single ip
// within a specific timeframe
type LimitedListener struct {
	listener       net.Listener
	limit          time.Duration
	lastConnection time.Time
	history        []attempt
}

type attempt struct {
	host string
	time time.Time
}

// NewLimitedListener initializes a new listener for the specified address (host:port)
// throttling incoming connections
func NewLimitedListener(address string, limit time.Duration) (*LimitedListener, error) {
	if limit < 0 {
		return nil, fmt.Errorf("Invalid time limit (negative)")
	}
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &LimitedListener{
		listener:       l,
		limit:          limit,
		lastConnection: time.Time{},
		history:        nil,
	}, nil
}

// clearHistory truncates the history to only relevant entries
func (ll *LimitedListener) clearHistory() {
	tl := time.Now().Add(-ll.limit) // get timelimit of range to check

	// no connection made in the last X seconds
	// the vast majority of connections will proc this
	if ll.lastConnection.Before(tl) {
		ll.history = nil // reset and release to gc
	}

	if len(ll.history) > 0 {
		i := 0
		for ; i < len(ll.history); i++ {
			if ll.history[i].time.After(tl) { // inside target range
				break
			}
		}

		if i >= len(ll.history) {
			ll.history = nil
		} else {
			ll.history = ll.history[i:]
		}
	}
}

// isInHistory checks if a host (IP) has connected in the last X seconds
// clears history before checking
func (ll *LimitedListener) isInHistory(host string) bool {
	ll.clearHistory()

	for _, h := range ll.history {
		if h.host == host {
			return true
		}
	}
	return false
}

// addToHistory adds an address to the system at the current time
func (ll *LimitedListener) addToHistory(host string) {
	ll.history = append(ll.history, attempt{host: host, time: time.Now()})
	ll.lastConnection = time.Now()
}

// Accept accepts a connection if no other connection attempt from that ip has been made
// in the specified time frame
func (ll *LimitedListener) Accept() (net.Conn, error) {
	//ll.listener.SetDeadline(time.Now().Add(time.Second))
	con, err := ll.listener.Accept()
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(con.RemoteAddr().String())
	if err != nil {
		con.Close()
		return nil, err
	}

	if ll.isInHistory(host) {
		con.Close()
		return nil, fmt.Errorf("connection rate limit exceeded for %s", host)
	}

	ll.addToHistory(host)
	return con, nil
}

// Addr returns the address the listener is listening to
func (ll *LimitedListener) Addr() net.Addr {
	return ll.listener.Addr()
}

// Close closes the associated net.Listener
func (ll *LimitedListener) Close() {
	ll.listener.Close()
}
