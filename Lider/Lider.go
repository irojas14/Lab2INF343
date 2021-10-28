package main

import (
	"context"
	"log"
)

const (
	nameNodeAddress = "alumno@dist152.inf.santiago.usm.cl"
	nameNodePort    = ":50051"
)

var jugadorCount int32 = 0

type server struct {
	pb.UnimplementedLiderServer
}

func (s *server) Unirse(ctx context.Context, in *pb.SolictudUnirse) (*pb.RespuestaUnirse, error) {
	log.Printf("Solicitud: " + in.GetSolictud())
	jugadorCount++
	return &pb.RespuestaUnirse{NumJugador: JugadorId(jugadorCount), NumJuego: RespuestaUnirse_JUEGO.Luces}, nil
}

func main() {
}
