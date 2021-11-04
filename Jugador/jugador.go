package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port    = ":50052"
	local   = "localhost" + port
	address = "dist149.inf.santiago.usm.cl" + port
)

var (
	ClientNumJugador  pb.JugadorId
	ClientNumRonda    pb.RondaId
	ClientCurrentGame pb.JUEGO
)

func main() {
	rand.Seed(time.Now().UnixNano())
	dialAddrs := address
	if len(os.Args) == 2 {
		dialAddrs = local
	}

	fmt.Printf("COMENZANDO EL JUGADOR - Addr: %s\n", dialAddrs)
	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewLiderClient(conn)

	stream, err := c.Unirse(context.Background(), &pb.SolicitudUnirse{})

	if err != nil {
		log.Fatalf("Error: %v\n", err)
		return
	}
	for {
		fmt.Println("A Recibir datos de Unirse del Lider")
		r, err := stream.Recv()
		fmt.Println("Información Recibida")	
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("\nError al Recibir: %v - %v\n", c, err)
			break
		}
		log.Println("\nRespuesta: " + r.String())

		if r.MsgTipo == pb.RespuestaUnirse_Comenzar {
			ClientNumJugador = *r.GetNumJugador()
			ClientNumRonda = *r.GetNumRonda()
			ClientCurrentGame = r.GetNumJuego()
			log.Printf("Valores: %v - %v - %v\n", ClientNumJugador.GetVal(), ClientNumRonda.GetVal(), ClientCurrentGame)
			break
		}
	}
	if ClientCurrentGame == pb.JUEGO_Luces {
		Luces(c)
	}
}

// JUEGOS

func Luces(c pb.LiderClient) error {
	var randval int32 = funcs.RandomInRange(1, 10)
	fmt.Printf("Random Value: %v\n", randval)

	stream, err := c.EnviarJugada(context.Background())

	if err != nil {
		log.Fatalf("Error al Enviar Jugada Stream: %v\n", err)
		return err
	}

	for {
		jugada := pb.EnvioJugada{
			Tipo:       pb.EnvioJugada_Jugada,
			Rol:        pb.EnvioJugada_Jugador,
			NumJuego:   pb.JUEGO_Luces,
			NumRonda:   &ClientNumRonda,
			NumJugador: &ClientNumJugador,
			Jugada:     &pb.Jugada{Val: randval},
		}

		fmt.Printf("Enviando Jugada al Lider: %v - Jugador: %v\n", jugada.Jugada, ClientNumJugador.Val)
		
		stream.Send(&jugada)

		fmt.Println("Jugada Enviada, Esperando Resolución")

		in, err := stream.Recv()

		fmt.Println("Resolución Recibida")

		if err == io.EOF {
			log.Fatalf("END OF FILE: %v\n", err)
			return nil
		}

		if err != nil {
			log.Fatalf("Error No EOF: %v\n", err)
			return err
		}

		fmt.Printf("Respuesta Jugada: %v - Estado: %v\n", in.String(), in.Estado.String())

		if in.GetEstado() == pb.ESTADO_Muerto {
			fmt.Println("Muerto, Cerrando Stream y Volviendo")
			break
		}

		fmt.Println("Esperando Respuesta de Inicio de 2da Ronda")
		
		in2, err2 := stream.Recv()

		fmt.Println("Respuesta de 2da Ronda Recibida")

		if (err2 == io.EOF) {
			log.Fatalf("Error EOF en in2: %v\n", err)
			return err2			
		}

		if err2 != nil {
			log.Fatalf("Error No EOF en in2: %v\n", err)
			return err2
		}

		fmt.Println("Contenido 2da Respuesta: " + in2.String())

		if in2.GetTipo() == pb.EnvioJugada_NuevaRonda {
			fmt.Println("Pasamos a la Siguiente Ronda")
		} else if  in2.GetTipo() == pb.EnvioJugada_Ganador {
			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumarizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			break
		} else if in2.GetTipo() == pb.EnvioJugada_Fin {
			fmt.Printf("Se acabo el juego...por ahora\n")
			break
		} else if in2.GetTipo() == pb.EnvioJugada_NuevoJuego {
			fmt.Printf("Cambiando de Etapa\n")
			break
		}
	}
	stream.CloseSend()
	return nil
}

/*s
func TirarCuerda(c pb.LiderClient) error {
	var randval int32 = funcs.RandomInRange(1, 4)
	fmt.Printf("Random Value: %v\n", randval)

	r, err := c.EnviarJugada(context.Background(), &pb.SolicitudEnviarJugada{
		JugadaInfo: &pb.PaqueteJugada{
			NumJugador: &pb.JugadorId{Val: ClientNumJugador.Val},
			NumJuego:   ClientCurrentGame,
			NumRonda:   &pb.RondaId{Val: ClientNumRonda.Val},
			Jugada:     &pb.Jugada{Val: randval}}})

	if err != nil {
		log.Fatalf("Error al jugar tirar cierda: %v\n", err)
		return err
	}
	fmt.Printf("Respuesta Jugada: %v", r.String())
	waitc <- "done"
	return nil
}
*/

// AUXILIAR

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
