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
	address = "dist149.inf.santiago.usm.cl:50052"
)

func main() {
	fmt.Println("COMENZANDO EL JUGADOR")
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	fmt.Println("Sgte Linea desde el Dial")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Dial Terminado")

	c := pb.NewLiderClient(conn)

	fmt.Println("Cliente Creado")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Unirse(ctx, &pb.SolicitudUnirse{Solictud: "Quiero Jugar"})
	fmt.Println("Sgte Linea desde el llamado Unirse")

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println("Llamado Remoto Finalizado")
	log.Printf("Bienvenido al Juego: Tu Numero: " + r.GetNumJugador().String() + " Juego: " + r.GetNumJuego().String() + "\n")
}

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
