package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/Lab2INF343/Proto"
)

const (
	LiderAddres      = "alumno@dist149.inf.santiago.usm.cl"
	LiderPort        = ":50052"
	LiderFullAddress = LiderAddres + LiderPort
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(LiderFullAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewLiderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Unirse(ctx, &pb.SolicitudUnirse{Solictud: "Quiero Jugar"})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Bienvenido al Juego: Tu Numero: Juego: " + r.GetNumJugador().String() + r.GetNumJuego().String())
}

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
