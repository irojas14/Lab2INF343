package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port    = ":50051"
	local   = "localhost" + port
	address = "dist151.inf.santiago.usm.cl" + port
)

var MontoAcumulado int32 = 0

type server struct {
	pb.UnimplementedPozoServer
}

func (s *server) VerMonto(ctx context.Context, in *pb.SolicitudVerMonto) (*pb.RespuestaVerMonto, error) {
    log.Println("Sirviendo Solicitud de Ver Monto")
    return &pb.RespuestaVerMonto{ Monto: float32(MontoAcumulado) }, nil
}

func main() {
	srvAddr := address
	if len(os.Args) == 2 {
		srvAddr = local
	}

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("faled to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterPozoServer(s, &server{})
	log.Printf("Pozo escuchando en %v", lis.Addr())
	
    if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
