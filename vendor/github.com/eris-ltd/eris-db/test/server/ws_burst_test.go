package server

import (
	"testing"
	"time"

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/server"
	"github.com/stretchr/testify/assert"
)

const CONNS uint16 = 100
const MESSAGES = 1000

// To keep track of new websocket sessions on the server.
type SessionCounter struct {
	opened int
	closed int
}

func (this *SessionCounter) Run(oChan, cChan <-chan *server.WSSession) {
	go func() {
		for {
			select {
			case <-oChan:
				this.opened++
				break
			case <-cChan:
				this.closed++
				break
			}
		}
	}()
}

func (this *SessionCounter) Report() (int, int, int) {
	return this.opened, this.closed, this.opened - this.closed
}

// Testing to ensure that websocket server does not crash, and that it
// cleans up after itself.
func TestWsFlooding(t *testing.T) {

	// New websocket server.
	wsServer := NewScumsocketServer(CONNS)

	// Keep track of sessions.
	sc := &SessionCounter{}

	// Register the observer.
	oChan := wsServer.SessionManager().SessionOpenEventChannel()
	cChan := wsServer.SessionManager().SessionCloseEventChannel()

	sc.Run(oChan, cChan)

	serveProcess, err := NewServeScumSocket(wsServer)
	assert.NoError(t, err, "Failed to serve new websocket.")
	errServe := serveProcess.Start()
	assert.NoError(t, errServe, "ScumSocketed!")
	t.Logf("Flooding...")
	// Run. Blocks.
	errRun := runWs()
	stopC := serveProcess.StopEventChannel()
	errStop := serveProcess.Stop(0)
	<-stopC
	assert.NoError(t, errRun, "ScumSocketed!")
	assert.NoError(t, errStop, "ScumSocketed!")
	o, c, a := sc.Report()
	assert.Equal(t, uint16(o), CONNS, "Server registered '%d' opened conns out of '%d'", o, CONNS)
	assert.Equal(t, uint16(c), CONNS, "Server registered '%d' closed conns out of '%d'", c, CONNS)
	assert.Equal(t, uint16(a), uint16(0), "Server registered '%d' conns still active after shutting down.", a)
}

func runWs() error {
	doneChan := make(chan bool)
	errChan := make(chan error)
	for i := uint16(0); i < CONNS; i++ {
		go wsClient(doneChan, errChan)
	}
	runners := uint16(0)
	for runners < CONNS {
		select {
		case _ = <-doneChan:
			runners++
		case err := <-errChan:
			return err
		}
	}
	return nil
}

func wsClient(doneChan chan bool, errChan chan error) {
	client := client.NewWSClient("ws://localhost:31401/scumsocket")
	_, err := client.Dial()
	if err != nil {
		errChan <- err
		return
	}
	readChan := client.StartRead()
	i := 0
	for i < MESSAGES {
		client.WriteMsg([]byte("test"))
		<-readChan
		i++
	}
	client.Close()
	time.Sleep(100 * time.Millisecond)

	doneChan <- true
}
