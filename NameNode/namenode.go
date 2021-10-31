package main

import (
	"context"
	"log"

	pb "github.com/irojas14/Lab2INF343/Proto"
)

const (
	port    = ":50054"
	local   = "localhost" + port
	address = "dist149.inf.santiago.usm.cl" + port
)

var JugadodasDeJugadores = "JugadasDeJugadores.txt"


type server struct {
	pb.UnimplementedNameNodeServer
}

func (s *server) RegistrarJugadas(ctx context.Context, in *pb.SolicitudRegistrarJugadas) (*pb.RespuestaRegistrarJugadas, error) {
    log.Println("Sirviendo Solicitud de Registrar Jugada")
    return &pb.RespuestaRegistrarJugadas{ JugadorId: &pb.JugadorId{Val: in.JugadasJugador.NumJugador.Val} }, nil
}

func (s *server) DevolverJugadas(ctx context.Context, in *pb.SolicitudDevolverRegistro) (*pb.RespuestaDevolverRegistro, error) {
	log.Println("Sirviendo Solicitud de Devolver Jugada")
	return &pb.RespuestaDevolverRegistro{ JugadasJugador: &pb.JugadasJugador{Val: in.}}
}

