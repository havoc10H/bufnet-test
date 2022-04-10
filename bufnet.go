package bufnet

import (
	"errors"
	"net"

	"github.com/sysdevguru/bufnet/writer"
)

const (
	defaultBandwidth = 1024
)

var (
	errConnBandwidth = errors.New("connection bandwidth should be smaller than server bandwidth")

	lockChan = make(chan struct{}, 1)
)

// BufferedListener is the buffered net.Listener
type BufferedListener struct {
	bandwidth     int
	connBandwidth int
	connCount     int
	net.Listener
}

// Listen returns buffered listener
func Listen(ln net.Listener, serverBandwidth, connBandwidth int) (*BufferedListener, error) {
	// set defaultBandwidth if negative values are provided
	if serverBandwidth < 0 {
		serverBandwidth = defaultBandwidth
	}
	if connBandwidth < 0 {
		connBandwidth = defaultBandwidth
	}
	if connBandwidth > serverBandwidth {
		return nil, errConnBandwidth
	}

	return &BufferedListener{bandwidth: serverBandwidth, connBandwidth: connBandwidth, Listener: ln}, nil
}

// BufConn makes buffered connection based on provided listener and connection
// this is used for per connection bandwidth control
func BufConn(c net.Conn, ln net.Listener, connBandwidth int) *BufferedConn {
	// set listener bandwidth as 0, no server bandwidth limit
	bl := &BufferedListener{bandwidth: 0, Listener: ln}

	if connBandwidth < 0 {
		connBandwidth = defaultBandwidth
	}
	return newBufferedConn(bl, c, connBandwidth)
}

// Accept returns buffered net.Conn
func (bl *BufferedListener) Accept() (net.Conn, error) {
	c, err := bl.Listener.Accept()
	if err != nil {
		return c, err
	}

	// update connections count
	lockChan <- struct{}{}
	bl.connCount++
	<-lockChan

	c = newBufferedConn(bl, c, bl.connBandwidth)
	return c, err
}

// BufferedConn is the wrapper for net.Conn
type BufferedConn struct {
	bandwidth        int
	bufferedListener *BufferedListener
	originBandwidth  int
	net.Conn
}

func newBufferedConn(bl *BufferedListener, c net.Conn, connBandwidth int) *BufferedConn {
	return &BufferedConn{bandwidth: connBandwidth, bufferedListener: bl, originBandwidth: connBandwidth, Conn: c}
}

// Write to buffered connection
func (bc *BufferedConn) Write(p []byte) (n int, err error) {
	// get updated bandwidth
	bc.updateBandwidth()

	// skip limiting if connection bandwidth is 0
	if bc.bandwidth == 0 {
		return bc.Conn.Write(p)
	}

	writer := writer.NewWriter(bc.Conn, bc.bandwidth)
	return writer.Write(p)
}

// Close the connection, decrease connection count of listener
func (bc *BufferedConn) Close() error {
	var err error
	if bc.Conn != nil {
		err = bc.Conn.Close()
		lockChan <- struct{}{}
		bc.bufferedListener.connCount--
		<-lockChan
		bc.Conn = nil
	}
	return err
}

func (bc *BufferedConn) updateBandwidth() {
	lockChan <- struct{}{}
	// update connection bandwidth when there is server bandwidth limit
	if bc.bufferedListener.bandwidth != 0 {
		bc.bandwidth = bc.bufferedListener.bandwidth / bc.bufferedListener.connCount

		// increase bandwidth in case connections are closed
		if bc.bufferedListener.connCount*bc.originBandwidth <= bc.bufferedListener.bandwidth {
			bc.bandwidth = bc.originBandwidth
		}
	}
	<-lockChan
}
