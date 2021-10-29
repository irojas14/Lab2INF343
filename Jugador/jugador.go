package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	LiderAddres      = "alumno@dist149.inf.santiago.usm.cl"
	LiderPort        = ":50052"
	LiderFullAddress = LiderAddres + LiderPort
)

func main() {
	fmt.Print("COMENZANDO EL JUGADOR")
	// Set up a connection to the server.
	conn, err := grpc.Dial(LiderFullAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	fmt.Print("Dial Terminado")

	c := pb.NewLiderClient(conn)

	fmt.Print("Cliente Creado")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Unirse(ctx, &pb.SolicitudUnirse{Solictud: "Quiero Jugar"})
	fmt.Print("Sgte Linea desde el llamado Unirse")

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Print("Llamado Remoto Finalizado")
	log.Printf("Bienvenido al Juego: Tu Numero: Juego: " + r.GetNumJugador().String() + r.GetNumJuego().String())
}

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
