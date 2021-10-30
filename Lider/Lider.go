package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/irojas14/Lab2INF343/Proto"

	"google.golang.org/grpc"
)

const (
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(ctx context.Context, in *pb.SolicitudUnirse) (*pb.RespuestaUnirse, error) {
	log.Printf("Solicitud: " + in.GetSolictud())
	jugadorCount++
	return &pb.RespuestaUnirse{NumJugador: &pb.JugadorId{Val: jugadorCount}, NumJuego: pb.JUEGO_Luces, NumRonda: &pb.RondaId{Val: 1}}, nil
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
		log.Fatalf("faled to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLiderServer(s, &server{})
	log.Printf("Juego Inicializado: escuhando en %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
