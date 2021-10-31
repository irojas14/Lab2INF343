package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"os"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	MaxPlayers = 16
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)
var jugCountMux sync.Mutex;
var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	jugCountMux.Lock()
	defer jugCountMux.Unlock()

	if (jugadorCount < MaxPlayers) {
		res := &pb.RespuestaUnirse {

		}
	}

}

func (s *server) UnirseProcessing() {
	if (jugadorCount < MaxPlayers) {

	}
}

// Para actuaizar el Proto file, correr
// go get -u github.com/irojas14/Lab2INF343/Proto
// O CON
// go get -u github.com/irojas14/Lab2INF343

func main() {
	fmt.Printf("rgs en: %v\n", len(os.Args))
	srvAddr := address
	if len(os.Args) == 2 {
		srvAddr = local
	}

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("faled tolisten: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLiderServer(s, &server{})
	log.Printf("Juego Inicializado: escuhano en %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

