package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/Lab2INF343/Proto"
)

const (
	nameNodeAddress = "alumno@dist152.inf.santiago.usm.cl"
	nameNodePort    = ":50051"
	Port            = ":50052"
)

var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(ctx context.Context, in *pb.SolicitudUnirse) (*pb.RespuestaUnirse, error) {
	log.Printf("Solicitud: " + in.GetSolictud())
	jugadorCount++
	return &pb.RespuestaUnirse{NumJugador: &pb.JugadorId{Val: jugadorCount}, NumJuego: pb.RespuestaUnirse_Luces}, nil
}

func main() {
	lis, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLiderServer(s, &server{})
	log.Printf("Juego Inicializado: escuchando en %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
