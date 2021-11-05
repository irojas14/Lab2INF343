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
	ClienTeam         int32
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
	stream, err := c.EnviarJugada(context.Background())

	if err != nil {
		log.Fatalf("Error al Enviar Jugada Stream: %v\n", err)
		return err
	}

	for {
		var randval int32 = funcs.RandomInRange(1, 10)
		fmt.Printf("Random Value: %v\n", randval)

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

		if in.GetEstado() == pb.ESTADO_MuertoDefault {
			fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil
		} else if in.GetEstado() == pb.ESTADO_Muerto {
			fmt.Println("Muerto, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil
		} else if in.Estado == pb.ESTADO_Ganador {

			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			stream.CloseSend()
			return nil
		}
		if in.NumRonda.Val+2 >= 5 {
			fmt.Printf("Esperando Respuesta de Inicio de juego: Tirar cuerda. \n")
		} else {
			fmt.Printf("Esperando Respuesta de Inicio de %v Ronda\n", in.NumRonda.Val+2)
		}

		// 2DA RESPUESTA

		in2, err2 := stream.Recv()

		fmt.Printf("2da Respuesta Recibida: %v\n", in2.String())
		//fmt.Printf("Respuesta de %v Ronda Recibida\n", in.NumRonda.Val+1)

		if err2 == io.EOF {
			log.Fatalf("Error EOF en in2: %v\n", err)
			return err2
		}

		if err2 != nil {
			log.Fatalf("Error No EOF en in2: %v\n", err)
			return err2
		}
		fmt.Println("Contenido Respuesta: " + in2.String())

		if in2.GetTipo() == pb.EnvioJugada_NuevaRonda {

			fmt.Println("Pasamos a la Siguiente Ronda")

		} else if in2.GetTipo() == pb.EnvioJugada_Ganador {

			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			break

		} else if in2.GetTipo() == pb.EnvioJugada_Fin {
			if in2.Estado == pb.ESTADO_Muerto {

				fmt.Println("Muerto, Cerrando Stream y Volviendo")
			} else if in2.Estado == pb.ESTADO_MuertoDefault {
				fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")

			} else {

				fmt.Printf("Se acabó el juego...por ahora\n")
			}
			stream.CloseSend()
			return nil

		} else if in2.GetTipo() == pb.EnvioJugada_NuevoJuego {
			fmt.Printf("Cambiando de Etapa\n")
			ClienTeam = in2.Equipo
			ClientNumRonda = *in2.GetNumRonda()
			break
		}
		fmt.Println()
		fmt.Println()
	}
	fmt.Println("Yendo a la 2da Etapa: Tirar la Cuerda")
	TirarCuerda(c, stream)
	return nil
}

func TirarCuerda(c pb.LiderClient, stream pb.Lider_EnviarJugadaClient) error {

	fmt.Println("En el Juego 2: Tirar la Cuerda!!")
	for {
		var randval int32 = funcs.RandomInRange(1, 4)
		fmt.Printf("Random Value: %v\n", randval)

		jugada := &pb.EnvioJugada{
			Tipo:       pb.EnvioJugada_Jugada,
			Rol:        pb.EnvioJugada_Jugador,
			NumJuego:   pb.JUEGO_TirarCuerda,
			NumRonda:   &ClientNumRonda,
			NumJugador: &ClientNumJugador,
			Jugada:     &pb.Jugada{Val: randval},
			Equipo:     ClienTeam,
		}

		fmt.Printf("Enviando Jugada al Lider: %v - Jugador: %v\n", jugada.Jugada, ClientNumJugador.Val)

		stream.Send(jugada)

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

		if in.Estado == pb.ESTADO_Muerto {
			fmt.Println("Muerto, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil

		} else if in.Estado == pb.ESTADO_MuertoDefault {
			fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil

		} else if in.Estado == pb.ESTADO_Ganador {

			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			stream.CloseSend()
			return nil
		}

		fmt.Println("Esperando Respuesta de Inicio de Nueva Etapa")

		in2, err2 := stream.Recv()

		res2, tipo := ProcesarRespuesta(stream, in2, err2, jugada)

		if tipo == TipoRespuesta_EOF || tipo == TipoRespuesta_Error || tipo == TipoRespuesta_Terminar {
			return res2

		} else if tipo == TipoRespuesta_NuevaEtapa {
			break
		}
	}
	var linea string
	fmt.Println("Nos iríamos a la 3ra Etapa: Esperando input")
	fmt.Scanln(&linea)
	// Sgte Etapa
	return nil
}

type TipoRespuesta int32

const (
	TipoRespuesta_Normal     TipoRespuesta = 0
	TipoRespuesta_EOF        TipoRespuesta = 1
	TipoRespuesta_Error      TipoRespuesta = 2
	TipoRespuesta_Terminar   TipoRespuesta = 3
	TipoRespuesta_NuevaRonda TipoRespuesta = 4
	TipoRespuesta_NuevaEtapa TipoRespuesta = 5
)

func ProcesarRespuesta(stream pb.Lider_EnviarJugadaClient, in *pb.EnvioJugada, err error, jugada *pb.EnvioJugada) (error, TipoRespuesta) {

	var tipo TipoRespuesta
	var reterr error

	if err == io.EOF {
		log.Fatalf("END OF FILE: %v\n", err)
		tipo = TipoRespuesta_EOF
		reterr = nil
	}

	if err != nil {
		log.Fatalf("Error No EOF: %v\n", err)
		tipo = TipoRespuesta_Error
		reterr = err
	}

	fmt.Printf("Respuesta Jugada: %v - Estado: %v\n", in.String(), in.Estado.String())

	if in.Estado == pb.ESTADO_Muerto {
		fmt.Println("Muerto, Cerrando Stream y Volviendo")
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil

	} else if in.Estado == pb.ESTADO_MuertoDefault {
		fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil

	} else if in.Estado == pb.ESTADO_Ganador || in.Tipo == pb.EnvioJugada_Ganador {

		fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
		fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil
	}

	if in.Tipo == pb.EnvioJugada_NuevaRonda {

		fmt.Println("Pasamos a la Siguiente Ronda")
		tipo = TipoRespuesta_NuevaRonda
		reterr = nil

	} else if in.Tipo == pb.EnvioJugada_Fin {

		fmt.Printf("Se acabó el juego...por ahora\n")
		tipo = TipoRespuesta_Terminar
		reterr = nil

	} else if in.Tipo == pb.EnvioJugada_NuevoJuego {
		fmt.Printf("Cambiando de Etapa\n")
		fmt.Println()
		fmt.Println()

		ClienTeam = in.Equipo
		ClientNumRonda = *in.GetNumRonda()
		tipo = TipoRespuesta_NuevaEtapa
		reterr = nil
	}

	fmt.Println()
	fmt.Println()

	return reterr, tipo
}

// AUXILIAR

func itos() {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	fmt.Printf("%v, %v\n", s1, s2)
}
