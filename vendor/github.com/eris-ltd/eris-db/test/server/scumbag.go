package server

import (
	"encoding/json"
	"os"
	"runtime"

	rpc "github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"
	"github.com/gin-gonic/gin"
	"github.com/tendermint/log15"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlWarn,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}

type ScumbagServer struct {
	running bool
}

func NewScumbagServer() server.Server {
	return &ScumbagServer{}
}

func (this *ScumbagServer) Start(sc *server.ServerConfig, g *gin.Engine) {
	g.GET("/scumbag", func(c *gin.Context) {
		c.String(200, "Scumbag")
	})
	this.running = true
}

func (this *ScumbagServer) Running() bool {
	return this.running
}

func (this *ScumbagServer) ShutDown() {
	// fmt.Println("Scumbag...")
}

type ScumSocketService struct{}

func (this *ScumSocketService) Process(data []byte, session *server.WSSession) {
	resp := rpc.NewRPCResponse("1", "Scumbag")
	bts, _ := json.Marshal(resp)
	session.Write(bts)
}

func NewScumsocketServer(maxConnections uint16) *server.WebSocketServer {
	sss := &ScumSocketService{}
	return server.NewWebSocketServer(maxConnections, sss)
}

func NewServeScumbag() (*server.ServeProcess, error) {
	cfg := server.DefaultServerConfig()
	cfg.Bind.Port = uint16(31400)
	return server.NewServeProcess(cfg, NewScumbagServer())
}

func NewServeScumSocket(wsServer *server.WebSocketServer) (*server.ServeProcess,
	error) {
	cfg := server.DefaultServerConfig()
	cfg.WebSocket.WebSocketEndpoint = "/scumsocket"
	cfg.Bind.Port = uint16(31401)
	return server.NewServeProcess(cfg, wsServer)
}
