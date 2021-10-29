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
)

var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(ctx context.Context, in *pb.SolicitudUnirse) (*pb.RespuestaUnirse, error) {
	log.Printf("Solicitud: " + in.GetSolictud())
	jugadorCount++
	//return &pb.RespuestaUnirse{NumJugador: JugadorId(jugadorCount), NumJuego: RespuestaUnirse_JUEGO.Luces}, nil
	return nil, nil
}

func main() {
	lis, err := net.Listen("tcp", nameNodeAddress+nameNodePort)
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
