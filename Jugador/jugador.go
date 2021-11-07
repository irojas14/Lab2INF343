package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	port         = ":50052"
	local        = "localhost" + port
	liderAddress = "dist149.inf.santiago.usm.cl" + port
)

var dialAddrs string

var (
	ClientNumJugador    pb.JugadorId
	ClientNumRonda      pb.RondaId
	ClientCurrentGame   pb.JUEGO
	ClienTeam           int32
	ClientCurrentEstado pb.ESTADO
)

type TipoJugador int

const (
	TipoJugador_Humano TipoJugador = 0
	TipoJugador_Bot    TipoJugador = 1
)

var (
	tipo TipoJugador
)

var (
	juegoIniciadoaMux sync.Mutex
)

var (
	enRondaMux sync.Mutex
	enRonda    bool = false
)

var (
	enUnirse    bool = false
	enUnirseMux sync.Mutex
)

var (
	waitForGameReady  chan int32 = make(chan int32)
	waitForJugadaChan chan int32 = make(chan int32)
	waitc             chan int32 = make(chan int32)
)

var (
	c pb.LiderClient
)

func main() {
	rand.Seed(time.Now().UnixNano())

	dialAddrs = liderAddress
	if len(os.Args) == 2 {
		if os.Args[1] == "h" {
			dialAddrs = liderAddress
			tipo = TipoJugador_Humano
		} else {
			dialAddrs = local
			if os.Args[1] == "1" {
				tipo = TipoJugador_Humano
			} else {
				tipo = TipoJugador_Bot
			}
		}

	} else {
		tipo = TipoJugador_Bot
	}

	//fmt.Printf("Dial Addres: %v - Tipo: %v\n", dialAddrs, tipo)

	if tipo == TipoJugador_Humano {
		go Update()
		go Consola()
		<-waitc
	} else {
		UnirseJuego()
	}
}

// JUEGOS

func Luces(cr *pb.LiderClient) error {
	c := *cr
	stream, err := c.EnviarJugada(context.Background())

	fmt.Printf("\nCOMENZANDO JUEGO LUCES: %v\n\n", ClientNumJugador.Val)

	if err != nil {
		log.Fatalf("Error al Enviar Jugada Stream: %v\n", err)
		return err
	}

	for {

		var randval int32
		if tipo == TipoJugador_Humano {
			randval = <-waitForJugadaChan

		} else {
			randval = funcs.RandomInRange(1, 10)
		}

		jugada := &pb.EnvioJugada{
			Tipo:       pb.EnvioJugada_Jugada,
			Rol:        pb.EnvioJugada_Jugador,
			NumJuego:   pb.JUEGO_Luces,
			NumRonda:   &ClientNumRonda,
			NumJugador: &ClientNumJugador,
			Jugada:     &pb.Jugada{Val: randval},
		}

		fmt.Printf("Valor Elegido: %v\n", randval)

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

		fmt.Printf("Respuesta Jugada: %v - Etapa: %v - Estado: %v\n", in.String(), ClientCurrentGame, in.Estado.String())

		if in.GetEstado() == pb.ESTADO_Muerto21 {
			fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estas muerto, Cerrando Stream y Volviendo.")
			stream.CloseSend()
			return nil
		} else if in.GetEstado() == pb.ESTADO_MuertoDefault {
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

		fmt.Printf("2da Respuesta Recibida Etapa: %v -> %v \n", in2.NumJuego, in2.String())
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
			TurnEnRondaFalse()

		} else if in2.GetTipo() == pb.EnvioJugada_Ganador {

			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			break

		} else if in2.GetTipo() == pb.EnvioJugada_Fin {
			if in2.Estado == pb.ESTADO_Muerto {

				fmt.Println("Muerto, Cerrando Stream y Volviendo")
			} else if in2.Estado == pb.ESTADO_MuertoDefault {
				fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")

			} else if in2.Estado == pb.ESTADO_Muerto21 {
				fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estás muerto, Cerrando Stream y Volviendo.")

			} else {

				fmt.Printf("Se acabó el juego...por ahora\n")
			}
			stream.CloseSend()
			return nil

		} else if in2.GetTipo() == pb.EnvioJugada_NuevoJuego {
			fmt.Printf("Cambiando de Etapa\n")
			ClientCurrentGame = in2.NumJuego
			ClienTeam = in2.Equipo
			ClientNumRonda = *in2.GetNumRonda()
			TurnEnRondaFalse()
			break
		}
		fmt.Println()
		fmt.Println()
	}
	fmt.Println("Yendo a la 2da Etapa: Tirar la Cuerda")
	TirarCuerda(c, stream)
	return nil
}

func TurnEnRondaFalse() {
	enRondaMux.Lock()
	enRonda = false
	enRondaMux.Unlock()
}

func TirarCuerda(c pb.LiderClient, stream pb.Lider_EnviarJugadaClient) error {

	fmt.Println("En el Juego 2: Tirar la Cuerda!!")
	for {
		var randval int32
		if tipo == TipoJugador_Humano {
			randval = <-waitForJugadaChan
		} else {
			randval = funcs.RandomInRange(1, 4)
		}

		fmt.Printf("Valor Elegido: %v\n", randval)

		jugada := &pb.EnvioJugada{
			Tipo:       pb.EnvioJugada_Jugada,
			Rol:        pb.EnvioJugada_Jugador,
			NumJuego:   pb.JUEGO_TirarCuerda,
			NumRonda:   &ClientNumRonda,
			NumJugador: &ClientNumJugador,
			Jugada:     &pb.Jugada{Val: randval},
			Equipo:     ClienTeam,
		}

		fmt.Printf("Enviando Jugada al Lider: %v - Jugador: %v - Etapa Client: %v - Etapa Jugada: %v\n", jugada.Jugada, ClientNumJugador.Val, ClientCurrentGame, jugada.NumJuego)

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

		fmt.Printf("Respuesta Jugada: %v - Estado: %v - Client Etapa %v - in Etapa: %v\n", in.String(), in.Estado.String(), ClientCurrentGame, in.NumJuego)

		if in.Estado == pb.ESTADO_Muerto {
			fmt.Println("Muerto, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil

		} else if in.Estado == pb.ESTADO_Muerto21 {
			fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estás muerto, Cerrando Stream y Volviendo.")
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

		tipo, res2 := ProcesarRespuesta(stream, in2, err2, jugada)

		if tipo == TipoRespuesta_EOF || tipo == TipoRespuesta_Error || tipo == TipoRespuesta_Terminar {
			return res2

		} else if tipo == TipoRespuesta_NuevaEtapa {
			TurnEnRondaFalse()
			break
		}
	}
	fmt.Println("Yendo a la 3ra Etapa: Todo o Nada")
	JuegoTodoONada(c, stream)
	return nil
}

func JuegoTodoONada(c pb.LiderClient, stream pb.Lider_EnviarJugadaClient) error {

	for {
		var randval int32

		if tipo == TipoJugador_Humano {
			randval = <-waitForJugadaChan
		} else {
			randval = funcs.RandomInRange(1, 10)
		}

		fmt.Printf("Valor Elegido: %v\n", randval)

		jugada := pb.EnvioJugada{
			Tipo:       pb.EnvioJugada_Jugada,
			Rol:        pb.EnvioJugada_Jugador,
			NumJuego:   pb.JUEGO_TodoNada,
			NumRonda:   &ClientNumRonda,
			NumJugador: &ClientNumJugador,
			Jugada:     &pb.Jugada{Val: randval},
			Equipo:     ClienTeam,
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

		fmt.Printf("Respuesta Jugada: %v - Etapa: %v - Estado: %v\n", in.String(), ClientCurrentGame, in.Estado.String())

		if in.GetEstado() == pb.ESTADO_MuertoDefault {
			fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil
		} else if in.GetEstado() == pb.ESTADO_Muerto {
			fmt.Println("Muerto, Cerrando Stream y Volviendo")
			stream.CloseSend()
			return nil
		} else if in.GetEstado() == pb.ESTADO_Muerto21 {
			fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estás muerto, Cerrando Stream y Volviendo.")
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

		fmt.Printf("2da Respuesta Recibida Etapa: %v -> %v \n", in2.NumJuego, in2.String())
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

		if in2.Tipo == pb.EnvioJugada_NuevaRonda {

			fmt.Println("Pasamos a la Siguiente Ronda")
			TurnEnRondaFalse()

		} else if in2.Tipo == pb.EnvioJugada_Ganador {

			fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
			fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
			break

		} else if in2.Tipo == pb.EnvioJugada_Fin {
			if in2.Estado == pb.ESTADO_Muerto {

				fmt.Println("Muerto, Cerrando Stream y Volviendo")
			} else if in2.Estado == pb.ESTADO_MuertoDefault {
				fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")

			} else if in2.Estado == pb.ESTADO_Muerto21 {
				fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estás muerto, Cerrando Stream y Volviendo.")

			} else {

				fmt.Printf("Se acabó el juego...por ahora\n")
			}
			stream.CloseSend()
			return nil

		} else if in2.Tipo == pb.EnvioJugada_NuevoJuego {
			fmt.Printf("Cambiando de Etapa\n")
			ClientCurrentGame = in2.NumJuego
			ClienTeam = in2.Equipo
			ClientNumRonda = *in2.GetNumRonda()
			TurnEnRondaFalse()
			break
		}
		fmt.Println()
		fmt.Println()
	}
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

func ProcesarRespuesta(stream pb.Lider_EnviarJugadaClient, in *pb.EnvioJugada, err error, jugada *pb.EnvioJugada) (TipoRespuesta, error) {

	var tipo TipoRespuesta
	var reterr error

	if err == io.EOF {
		log.Fatalf("END OF FILE: %v\n", err)
		tipo = TipoRespuesta_EOF
		reterr = nil
		return tipo, reterr
	}

	if err != nil {
		log.Fatalf("Error No EOF: %v\n", err)
		tipo = TipoRespuesta_Error
		reterr = err
		return tipo, reterr
	}

	fmt.Printf("Respuesta Jugada: %v - Estado: %v\n", in.String(), in.Estado.String())

	if in.Estado == pb.ESTADO_Muerto {
		fmt.Println("Muerto, Cerrando Stream y Volviendo")
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil
		return tipo, reterr

	} else if in.Estado == pb.ESTADO_Muerto21 {
		fmt.Println("Jugaste 4 Rondas y tus azares no sumaron 21..., Estás muerto, Cerrando Stream y Volviendo.")
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil
		return tipo, reterr

	} else if in.Estado == pb.ESTADO_MuertoDefault {
		fmt.Println("Jugaste pero no tenias equipo, muerto por defecto(azares de paridad en juego 2)!, Cerrando Stream y Volviendo")
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil
		return tipo, reterr

	} else if in.Estado == pb.ESTADO_Ganador || in.Tipo == pb.EnvioJugada_Ganador {

		fmt.Println("HEMOS SOBREVIVIDO ! HEMOS GANADO ! Y AHORA SOMOS MILLONARIOS...Pero traumatizados")
		fmt.Printf("Tus números de la suerte: Jugada: %v - Numero Jugador: %v\n", jugada.Jugada.Val, jugada.NumJugador)
		stream.CloseSend()
		tipo = TipoRespuesta_Terminar
		reterr = nil
		return tipo, reterr
	}

	if in.Tipo == pb.EnvioJugada_NuevaRonda {

		fmt.Println("Pasamos a la Siguiente Ronda")
		tipo = TipoRespuesta_NuevaRonda
		reterr = nil

	} else if in.Tipo == pb.EnvioJugada_Fin {

		fmt.Printf("Se acabó el juego...por ahora\n")
		tipo = TipoRespuesta_Terminar
		reterr = nil
		stream.CloseSend()
		return tipo, reterr

	} else if in.Tipo == pb.EnvioJugada_NuevoJuego {
		fmt.Printf("Cambiando de Etapa\n")
		fmt.Println()
		fmt.Println()

		ClientCurrentGame = in.NumJuego
		ClienTeam = in.Equipo
		ClientNumRonda = *in.GetNumRonda()
		tipo = TipoRespuesta_NuevaEtapa
		reterr = nil
	}

	fmt.Println()
	fmt.Println()

	return tipo, reterr
}

var (
	terminarJuego = false
	input         string
	juegoIniciado = false
)

func Consola() {
	for {
		if terminarJuego {
			break
		}

		ShowConsola()

		fmt.Print("Ingrese Opción: ")
		fmt.Scanln(&input)
		fmt.Printf("Consola Input: %v\n", input)

		ProcesamientoConsola(input)
		fmt.Println("==================================")
	}
}

func ShowConsola() {
	fmt.Println("==========MENÚ JUGADOR==========")

	// Opciones
	if !juegoIniciado {
		fmt.Println("Unirse a JUEGO: PRESIONAR 'S' + 'ENTER'")

	} else {
		fmt.Println("Continuar al SIGUIENTE EVENTO: PRESIONAR 'S' + 'ENTER'")

	}
	fmt.Println("Ver MONTO ACUMULADO: PRESIONAR 'M' + 'ENTER'")
	fmt.Println("Terminar Servicio: PRESIONAR 'E' + 'ENTER'")
}

func Update() {

	<-waitForGameReady
	if juegoIniciado && ClientCurrentGame == pb.JUEGO_Luces {
		fmt.Println("En el Update Iniciando Luces")
		Luces(&c)
	}
}

func ProcesamientoConsola(input string) {

	if input == "m" || input == "M" {
		VerMonto()

	} else if input == "s" || input == "S" {
		if !juegoIniciado {
			Unirse()

		} else {
			SiguienteEvento()

		}
	} else if input == "e" || input == "E" {
		terminarJuego = true
		waitc <- 0
		return

	} else {
		fmt.Printf("Opción %v no válida\n", input)
	}
}

func VerMonto() {
	dialAddrs := liderAddress
	if len(os.Args) == 2 {
		dialAddrs = local
	}

	fmt.Println("CONSULTANDO MONTO ACUMULADO")

	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewLiderClient(conn)

	r, err2 := c.VerMonto(context.Background(), &pb.SolicitudVerMonto{})

	if err2 != nil {
		log.Fatalf("Error al Consultar el Monto: %v\n", err2)
		return
	}

	fmt.Printf("El MONTO ACUMULADO en el POZO Es: %v\n", r.Monto)
}

func Unirse() {
	enUnirseMux.Lock()

	if enUnirse {
		fmt.Println("Aún Esperando Unirse. Espere o Cancele")
		enUnirseMux.Unlock()
		return
	}

	enUnirse = true

	enUnirseMux.Unlock()

	go UnirseJuego()
}

func UnirseJuego() (*pb.LiderClient, error) {

	fmt.Println("INTENTADO UNIRSE A JUEGO")

	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c = pb.NewLiderClient(conn)

	stream, err := c.Unirse(context.Background(), &pb.SolicitudUnirse{})

	if err != nil {
		log.Fatalf("Error: %v\n", err)
		return nil, err
	}

	for {
		fmt.Println("A Recibir datos de Unirse del Lider")
		r, err := stream.Recv()
		fmt.Println("Información Recibida")

		if err == io.EOF {
			enUnirseMux.Lock()
			enUnirse = false
			enUnirseMux.Unlock()
			break
		}
		if err != nil {
			log.Fatalf("\nError al Recibir: %v - %v\n", c, err)
			enUnirseMux.Lock()
			enUnirse = false
			enUnirseMux.Unlock()
			return nil, err
		}
		log.Println("\nRespuesta: " + r.String())

		if r.MsgTipo == pb.RespuestaUnirse_Comenzar {
			ClientNumJugador = *r.GetNumJugador()
			ClientNumRonda = *r.GetNumRonda()
			ClientCurrentGame = r.GetNumJuego()
			ClientCurrentEstado = pb.ESTADO_Vivo
			log.Printf("Valores: %v - %v - %v\n", ClientNumJugador.GetVal(), ClientNumRonda.GetVal(), ClientCurrentGame)

			if tipo == TipoJugador_Humano {
				fmt.Println("Jugador Humano Recibió Comenzar")
				juegoIniciadoaMux.Lock()
				juegoIniciado = true
				juegoIniciadoaMux.Unlock()

				enUnirseMux.Lock()
				enUnirse = false
				enUnirseMux.Unlock()

				// Pone en marcha el update
				waitForGameReady <- 0
			}
			break
		}
	}
	if tipo == TipoJugador_Bot {
		Luces(&c)
		fmt.Println("Terminó el Juego para el Bot")
		return nil, nil
	}
	<-waitc
	return nil, nil
}

func SiguienteEvento() {
	enRondaMux.Lock()

	if enRonda {
		fmt.Println("Aún no termina la ronda")
		enRondaMux.Unlock()
		return
	}
	enRondaMux.Unlock()

	ElegirValor()
}

func ElegirValor() int32 {
	var a int32 = 1
	var b int32
	var input int32
	for {
		if ClientCurrentGame == pb.JUEGO_TirarCuerda {
			input = funcs.RandomInRange(1, 4)
			b = 4
			fmt.Printf("Tiraste la cuerda con valor %v.\n: ", input)
			if a <= input && input <= b {
				enRondaMux.Lock()
				enRonda = true
				enRondaMux.Unlock()
				waitForJugadaChan <- input
				return input
			} else {
				fmt.Printf("Valor %v elegido fuera del rango. Ingrese otra vez.\n", input)
			}
		} else {
			b = 10
		}

		fmt.Printf("Elija un Valor entre %v y %v: ", a, b)
		fmt.Scanln(&input)
		//fmt.Printf("Elegir Valor Input: %v\n", input)

		if a <= input && input <= b {
			enRondaMux.Lock()
			enRonda = true
			enRondaMux.Unlock()
			waitForJugadaChan <- input
			return input
		} else {
			fmt.Printf("Valor %v elegido fuera del rango. Ingrese otra vez.\n", input)
		}
	}
}
