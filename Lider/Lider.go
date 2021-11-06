package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

// Direcciones

const (
	MaxPlayers      = 4
	nameNodeAddress = "dist150.inf.santiago.usm.cl:50054"
	nPort           = ":50054"
	pozoAddress     = "dist151.inf.santiago.usm.cl:50051"
	pPort           = ":50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

var terminarJuego = false

var waitCountMux sync.Mutex
var waitingCount int32 = 0

var (
	CurrentAlivePlayers    int32 = 0
	CurrentAlivePlayersMux sync.Mutex

	CurrentRondaMux sync.Mutex
	CurrentRonda    int32 = 0
)

var (
	ResponsesCount    int32 = 0
	ResponsesCountMux sync.Mutex
)

var (
	RepliesCount       int32 = 0
	RepliesCountMux    sync.Mutex
	waitingForCleaning chan int32 = make(chan int32)
)

var (
	liderInput       string
	juegoIniciadoMux sync.Mutex
	juegoIniciado    bool = false

	rondaEnJuegoMux sync.Mutex
	rondaEnProceso  bool = false
)

// Channels

var (
	JuegoActual              pb.JUEGO
	gameReadyChan            chan int32 = make(chan int32)
	gameLiderValChan         chan int32 = make(chan int32)
	nextEventWaitingChan     chan int32 = make(chan int32)
	serverGameProcessingChan chan int32 = make(chan int32)
	esperaLiderChan          chan int32 = make(chan int32)
)

var (
	ValorLider     int32
	montoAcumulado float32
)

// Variables 1er Juego "Luz Roja, Luz Verde (Luces)"

var (
	Jugadas    []*pb.EnvioJugada
	JugadasMux sync.Mutex
)

// Variables 1er Juego "Luces"

var (
	SumasJugadores []int32
)

// Variables 2do Juego "Tirar la Cuerda"

var (
	Jugadores       []int32
	Team1           []int32
	ValoresTeam1    []int32
	ValoresTeam1Mux sync.Mutex

	Team2           []int32
	ValoresTeam2    []int32
	ValoresTeam2Mux sync.Mutex

	jugadorEliminado int32 = -1

	JugadoresMux sync.Mutex
)

// Variables 3er Juego "Todo o Nada"

var (
	AllTeams       [][]int32
	AllTeamsValues [][]int32

	AllTeamsValuesMux sync.Mutex
)

type server struct {
	pb.UnimplementedLiderServer
}

// METODOS DEL LIDER SERVER

// UNIRSE------------------------------------------------------------------------------------------------------------------------------
func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Println("Recibida Solicitud de Unirse...")

	waitCountMux.Lock()

	if waitingCount > MaxPlayers {
		waitCountMux.Unlock()
		return nil
	}
	waitingCount++
	countCont := waitingCount
	jugNum := &pb.JugadorId{Val: countCont}

	waitCountMux.Unlock()

	fmt.Printf("Número Asignado a Jugador: %v -> Jugadores Totales Actualmente Esperando: %v\n", jugNum.Val, waitingCount)

	if jugNum.GetVal() <= MaxPlayers {
		res := &pb.RespuestaUnirse{
			MsgTipo:    pb.RespuestaUnirse_Esperar,
			NumJugador: jugNum,
			NumJuego:   pb.JUEGO_None,
			NumRonda:   nil,
		}
		if err := stream.Send(res); err != nil {
			return err
		}

		CurrentAlivePlayersMux.Lock()
		CurrentAlivePlayers = jugNum.Val
		CurrentAlivePlayersMux.Unlock()

		if jugNum.Val == MaxPlayers {
			fmt.Printf("Se unieron los %v/%v Jugadores. Esperando señal de inicio del Líder para comenzar\n", CurrentAlivePlayers, MaxPlayers)
		} else {
			fmt.Printf("%v/%v. Esperando más jugadores...", CurrentAlivePlayers, MaxPlayers)
		}

		JugadoresMux.Lock()
		Jugadores = append(Jugadores, jugNum.Val)
		JugadoresMux.Unlock()

		res.MsgTipo = pb.RespuestaUnirse_Comenzar
		res.NumJuego = pb.JUEGO_Luces
		res.NumRonda = &pb.RondaId{Val: 0}

		<-gameReadyChan

		if err := stream.Send(res); err != nil {
			return err
		}

		CambiarEtapa(pb.JUEGO_Luces)

		fmt.Printf("Comenzando ! Jugadores Totales: %v\n", MaxPlayers)
		fmt.Println("JUEGO: Luz Roja, Luz Verde - Ronda: 1")
		fmt.Println()
		fmt.Println()

	}
	return nil
	/*
		else if jugNum.GetVal() == MaxPlayers {
			res := &pb.RespuestaUnirse{
				MsgTipo:    pb.RespuestaUnirse_Comenzar,
				NumJugador: jugNum,
				NumJuego:   pb.JUEGO_Luces,
				NumRonda:   &pb.RondaId{Val: 0},
			}
			if err := stream.Send(res); err != nil {
				return err
			}

			JugadoresMux.Lock()
			Jugadores = append(Jugadores, jugNum.Val)
			JugadoresMux.Unlock()

			CurrentAlivePlayers = res.GetNumJugador().GetVal()

			fmt.Printf("Se unieron los %v/%v Jugadores. Esperando señal de inicio del Líder para comenzar\n", CurrentAlivePlayers, MaxPlayers)

			<-gameReadyChan

			fmt.Println()
			fmt.Println()

			CambiarEtapa(pb.JUEGO_Luces)

			fmt.Printf("Comenzando ! Jugadores Totales: %v\n", MaxPlayers)
			fmt.Println("JUEGO: Luz Roja, Luz Verde - Ronda: 1")
		}
	*/
}

// ENVIAR JUGADA-------------------------------------------------------------------------------------------------------------------

func (s *server) EnviarJugada(stream pb.Lider_EnviarJugadaServer) error {

	for {
		// 1RA PARTE - ACUMULACION DE RESPUESTAS

		// Recibidos

		// Procesamiento estandar
		in, err := stream.Recv()

		if err == io.EOF {
			fmt.Printf("[SERVER] AL COMIENZO Un error EOF: %v\n", err)
			return nil
		}

		if err != nil {
			fmt.Printf("[SERVERAL COMIENZO Un Error no EOF: %v\n", err)
			return err
		}

		// Recoleccion de jugadas por juego
		JugadasMux.Lock()

		Jugadas = append(Jugadas, in)

		JugadasMux.Unlock()
		if JuegoActual == pb.JUEGO_Luces {
		} else if JuegoActual == pb.JUEGO_TirarCuerda {
			// Tirar Cuerda
			if in.Equipo == 1 {
				ValoresTeam1Mux.Lock()

				ValoresTeam1 = append(ValoresTeam1, in.Jugada.Val)

				ValoresTeam1Mux.Unlock()
			} else {
				ValoresTeam2Mux.Lock()

				ValoresTeam2 = append(ValoresTeam2, in.Jugada.Val)

				ValoresTeam2Mux.Unlock()
			}
		} else if JuegoActual == pb.JUEGO_TodoNada {
			// Todo o Nada

			AllTeamsValuesMux.Lock()

			var index int
			if AllTeams[in.Equipo][0] == in.NumJugador.Val {
				index = 0
			} else {
				index = 1
			}
			AllTeamsValues[in.Equipo][index] = in.Jugada.Val

			AllTeamsValuesMux.Unlock()
		}

		// Cuenta de Respuestas

		// ResponsesCountMux Lock BEGIN
		ResponsesCountMux.Lock()

		ResponsesCount++

		// Si soy la ultima respuesta esperada, avisar al lider (por medio del channel "inGameWaitingChan")
		if ResponsesCount == CurrentAlivePlayers {

			serverGameProcessingChan <- 0
		}

		ResponsesCountMux.Unlock()
		// Espacio Lock END

		// 1era PARTE END - ACUMULACION DE RESPUESTAS ------------------------------------------------------------

		// La recoleccion y respuestas ya se llevo a cabo, esperamos a que esten todas
		jugadaLider := <-gameLiderValChan

		// el lider ha procesado todas las respuestas, continuamos a la 2da parte

		// 2DA PARTE BEG - PROCESAMIENTO DE RESPUESTAS

		jugadas_len := len(Jugadas)
		num_jugador := in.NumJugador.Val

		// Creamos platilla de una respuesta, la que vamos a llenar con los datos correctos
		res := &pb.EnvioJugada{
			Tipo:     pb.EnvioJugada_Jugada,
			Rol:      pb.EnvioJugada_Lider,
			NumJuego: pb.JUEGO_Luces,
			Jugada:   &pb.Jugada{Val: jugadaLider},
			NumRonda: &pb.RondaId{Val: CurrentRonda},
		}

		// Recorremos el arreglo de jugadas, buscamos, por medio de nuestro NumJugador, la jugada que nos corresponde
		for index := 0; index < jugadas_len; index++ {
			jug := Jugadas[index]
			if num_jugador == jug.NumJugador.Val {
				// Llenamos la respuesta plantilla con el tipo de mensaje y con el estado correcto segun el arreglo "Jugadas"
				res.Tipo = jug.Tipo
				res.Estado = jug.Estado
				break
			}
		}
		// Enviamos la respuesta
		stream.Send(res)

		// Hemos enviado una respuesta a cliente (reply), la contamos
		// Cuando se hayan enviado todas las respuestas, procederemos a la limpieza
		RepliesCountMux.Lock()

		RepliesCount++
		if RepliesCount == ResponsesCount {
			waitingForCleaning <- 0
		}

		RepliesCountMux.Unlock()

		// Si la respuesa enviada fue una que termina nuestra participacion en el juego, entonces quebramos y nos vamos
		if res.Tipo == pb.EnvioJugada_Ganador ||
			res.Tipo == pb.EnvioJugada_Fin ||
			res.Estado == pb.ESTADO_Muerto ||
			res.Estado == pb.ESTADO_MuertoDefault ||
			res.Estado == pb.ESTADO_Ganador ||
			res.Estado == pb.ESTADO_Muerto21 {
			break
		}
		// Se enviaron y se cortaron respuestas, ahora esperemos a la limpieza

		// se proceso y se envio la informacion. Ahora el lider preparara el siguiente evento (limpieza), por lo que esperamos
		waitingRes := <-nextEventWaitingChan

		// el lider ya termino de limpiar, definir y preparar el siguiente evento, nos envio cual es
		// waitinRes = 2 -> Se acabo el juego, uso principalmente para debugueo
		//           = 1 -> Se viene una nueva etapa
		//           = 0 -> Se viene una nueva ronda

		// 3RA PARTE BEG - INFORMAR NUEVO EVENTO

		if waitingRes == 1 {
			var equipo int32 = 0

			res2 := &pb.EnvioJugada{
				Tipo:       pb.EnvioJugada_NuevoJuego,
				Rol:        pb.EnvioJugada_Lider,
				NumJuego:   JuegoActual,
				Jugada:     nil,
				NumJugador: in.GetNumJugador(),
				NumRonda:   &pb.RondaId{Val: CurrentRonda},
				Equipo:     equipo,
			}

			if JuegoActual == pb.JUEGO_TirarCuerda {

				searchRes := funcs.GetIndexOf(Team1, in.NumJugador.Val)
				if searchRes != -1 {
					res2.Equipo = 1
					res2.Estado = pb.ESTADO_Vivo

				} else {
					res2.Equipo = 2
					res2.Estado = pb.ESTADO_Vivo
				}

				if jugadorEliminado == in.NumJugador.Val {
					res2.Tipo = pb.EnvioJugada_Fin
					res2.Estado = pb.ESTADO_MuertoDefault
					fmt.Printf("PIUM!, hemos matado al jugador %v.\n", res2.NumJugador.Val)
					res2.Equipo = 0
					stream.Send(res2)
					return nil
				}

			} else if JuegoActual == pb.JUEGO_TodoNada {

				if jugadorEliminado == in.NumJugador.Val {
					res2.Tipo = pb.EnvioJugada_Fin
					res2.Estado = pb.ESTADO_MuertoDefault
					fmt.Printf("PIUM!, hemos matado al jugador %v.\n", res2.NumJugador.Val)
					res2.Equipo = -1
					stream.Send(res2)
					return nil
				}

				salirLoop := false
				var i int32
				var j int32
				for i = 0; i < int32(len(AllTeams)); i++ {
					for j = 0; j < 2; j++ {
						if AllTeams[i][j] == in.NumJugador.Val {
							res2.Equipo = i
							res2.Estado = pb.ESTADO_Vivo
							salirLoop = true
							break
						}
					}
					if salirLoop {
						break
					}
				}
			}
			stream.Send(res2)

		} else if waitingRes == 2 {

			fmt.Printf("Juego Acabado: Au Revoir Jugador: %v\n", in.GetNumJugador())
			fin := &pb.EnvioJugada{
				Tipo:       pb.EnvioJugada_Fin,
				Rol:        pb.EnvioJugada_Lider,
				NumJuego:   JuegoActual,
				Jugada:     &pb.Jugada{Val: jugadaLider},
				NumJugador: in.GetNumJugador(),
				Estado:     pb.ESTADO_Vivo,
			}
			stream.Send(fin)
			break

		} else {
			fmt.Printf("Iniciando Siguiente RONDA: %v - Avisando a Jugador: %v\n", CurrentRonda+1, in.NumJugador)

			nuevaRonda := &pb.EnvioJugada{
				Tipo:     pb.EnvioJugada_NuevaRonda,
				NumRonda: &pb.RondaId{Val: CurrentRonda},
				NumJuego: JuegoActual,
				Rol:      pb.EnvioJugada_Lider,
				Jugada:   &pb.Jugada{Val: jugadaLider},
			}
			stream.Send(nuevaRonda)
		}
	}
	return nil
}

func (s *server) VerMonto(ctx context.Context, in *pb.SolicitudVerMonto) (*pb.RespuestaVerMonto, error) {
	err := VerMonto()
	return &pb.RespuestaVerMonto{Monto: montoAcumulado}, err
}

// FUNCIONES NO DEL LIDER SERVER

func VerMonto() error {
	dialAddrs := pozoAddress
	if len(os.Args) == 2 {
		dialAddrs = "localhost:50051"
	}

	fmt.Println("CONTESTANDO SOLICITUD DE VER MONTO ACUMULADO al POZO")

	// Set up a connection to the server.
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return err
	}
	defer conn.Close()
	c := pb.NewPozoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.VerMonto(ctx, &pb.SolicitudVerMonto{})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return err
	}

	montoAcumulado = r.Monto
	return nil
}

// Para actuaizar el Proto file, correr
// go get -u github.com/irojas14/Lab2INF343/Proto
// O CON
// go get -u github.com/irojas14/Lab2INF343

var waitc chan int = make(chan int)

func main() {
	jugadorEliminado = -1
	go LiderService()
	<-waitc
	go LiderConsola()
	go Update()
	<-waitc
}

func Update() {
	for {
		if terminarJuego {
			fmt.Println(("Terminando Juego"))
			break
		}
		if JuegoActual == pb.JUEGO_Luces {
			JuegoLucesWaitForResponses()

		} else if JuegoActual == pb.JUEGO_TirarCuerda {

			JuegoTirarCuerdaWaitForResponses()
		} else if JuegoActual == pb.JUEGO_TodoNada {

			JuegoTodoNadaWaitForResponses()
		}
	}
}

// FUNCIONES CONSOLA  ------------------------------------------------------------------------------------------------------------------------------

func LiderConsola() {

	for {
		if terminarJuego {
			break
		}

		ShowConsola()

		fmt.Print("Ingrese Opción: ")
		fmt.Scanln(&liderInput)
		fmt.Printf("Lider Input: %v\n", liderInput)

		ProcesamientoConsola(liderInput)
		fmt.Println("==================================")
	}
}

func ShowConsola() {
	fmt.Println("==========MENÚ LÍDER==========")

	// Opciones
	if !juegoIniciado {
		fmt.Println("Iniciar JUEGO: PRESIONAR 'S' + 'ENTER'")

	} else {
		fmt.Println("Iniciar SIGUIENTE EVENTO: PRESIONAR 'S' + 'ENTER'")

	}
	fmt.Println("Ver MONTO ACUMULADO: PRESIONAR 'M' + 'ENTER'")
	fmt.Println("Ver JUGADAS de JUGADOR: PRESIONAR 'J' + 'ENTER'")
	fmt.Println("Terminar Servicio: PRESIONAR 'E' + 'ENTER'")
}

func ProcesamientoConsola(liderInput string) {

	if liderInput == "m" || liderInput == "M" {
		VerMonto()

	} else if liderInput == "s" || liderInput == "S" {
		if !juegoIniciado {
			IniciarJuego()

		} else {
			SiguienteEvento()

		}
	} else if liderInput == "j" || liderInput == "J" {
		ObtenerJugadas()

	} else if liderInput == "e" || liderInput == "E" {
		terminarJuego = true
		waitc <- 0
		return

	} else {
		fmt.Printf("Opción %v no válida\n", liderInput)
	}
}

func ObtenerJugadas() {

	var numJugadorStr string

	// Preguntamos por un Numero de Jugador
	fmt.Printf("Ingrese Numero Jugador o 'E' para Volver + 'ENTER': ")
	fmt.Scanln(&numJugadorStr)

	// Si nos entrega "E", volvemos
	if numJugadorStr == "e" || numJugadorStr == "E" {
		return
	}

	// Definimos si vamos a usar dirección remota o local
	dialAddrs := nameNodeAddress
	if len(os.Args) == 2 {
		dialAddrs = "localhost" + nPort
	}

	fmt.Printf("Dial Addres: %v\n", dialAddrs)

	// Creamos una conexión, se guarda en conn
	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()

	// Creamos el NameNodeClient, para hacer solicitudes al NameNode, que está escuchando (idealmente)
	nc := pb.NewNameNodeClient(conn)

	// Convertimos el input entregado, que es un string, a int64 y luego a int32
	numJugador64, err2 := strconv.ParseInt(numJugadorStr, 10, 32)
	if err2 != nil {
		log.Fatalf("Falló la Conversión al querer Obtener las jugadas del Jugador %v: err: %v\n", numJugadorStr, err2)
		return
	}
	numJugador32 := int32(numJugador64)

	// Realizamos el Procedimiento Remoto "DevolverJugadas" al NameNode
	// Recibimos un contexto y una SolicitudDevolverJugadasNameNode, que es básicamente un Numero de Jugador
	// en la variables r se almacenará nuestra respuesta

	//r : RespuestaDevolverJugadasNameNode
	// r.JugadasJugador
	// NumJugador
	// Arreglo de JugadasJuego
	// una JugadaJuego es:
	// NumJuego
	// Arreglo de Jugadas
	// Jugadas.Val => es el valor int32 de la jugada

	r, err3 := nc.DevolverJugadas(context.Background(),
		&pb.SolicitudDevolverJugadasNameNode{
			NumJugador: &pb.JugadorId{
				Val: numJugador32,
			},
		})

	if err3 != nil {
		log.Fatalf("Error Al Consultar las Jugadas del Jugador %v: err: %v\n", numJugadorStr, err3)
		return
	}

	// Imprimimos la respuesta "r"
	// Recorremos el arreglo de JugadasJuego
	// Imprimimos el NumJuego; y
	// Recorremos el arreglo de Jugadas e imprimimos su campo Val
	fmt.Printf("MOSTRANDO Jugadas del JUGADOR %v\n", r.JugadasJugador.NumJugador.Val)
	for _, jugadaJuego := range r.JugadasJugador.JugadasJuego {
		fmt.Printf("JUEGO: %v\n", jugadaJuego.NumJuego)

		for i, jugada := range jugadaJuego.Jugadas {
			fmt.Printf("[%v]: %v\n", i, jugada.Val)
		}
	}
}

func SiguienteEvento() {
	if !juegoIniciado {
		fmt.Println("El juego aún no ha sido iniciado")
		return

	} else if rondaEnProceso {

		CurrentRondaMux.Lock()
		if CurrentRonda+1 <= 4 {
			fmt.Printf("Ronda %v está aún en Proceso\n", CurrentRonda+1)
		} else {
			fmt.Println("El desarrollo del juego aún está en proceso.")
		}
		CurrentRondaMux.Unlock()
		return
	}

	rondaEnJuegoMux.Lock()
	rondaEnProceso = true
	rondaEnJuegoMux.Unlock()

	esperaLiderChan <- 0
}
func IniciarJuego() {

	fmt.Println("INTENTANDO INICIAR JUEGO")
	CurrentAlivePlayersMux.Lock()
	defer CurrentAlivePlayersMux.Unlock()

	if juegoIniciado {
		fmt.Println("El juego ya fue iniciado")
		return
	}
	if CurrentAlivePlayers < MaxPlayers {
		fmt.Printf("Faltan jugadores. Jugadores Actuales: %v\n", CurrentAlivePlayers)
		return
	}

	juegoIniciadoMux.Lock()

	juegoIniciado = true

	juegoIniciadoMux.Unlock()

	for i := 0; i < MaxPlayers; i++ {
		gameReadyChan <- 0
	}
}

func JuegoLucesWaitForResponses() {
	<-serverGameProcessingChan

	// 1ERA PARTE - DEFINIR ESTADOS (VIVOS, MUERTOS, GANADOR)

	// el lider elige su numero

	/*
			PARA DEBUGEAR, SE DEJA EL RANDOM DEL  LIDER EN NUMBEROS MAS ALTOS!!!
		liderVal := funcs.RandomInRange(6, 10)
	*/
	ValorLider = funcs.RandomInRange(11, 14)
	fmt.Printf("VALOR DEL LÍDER: %v\n", ValorLider)
	// Recorremos las respuestas colocadas en el arreglo global "Jugadas", definiendo el estado correspondiente
	// Si el valor del jugador es menor que el del lider, entonces vive, si no, muere
	var vivos []*pb.EnvioJugada
	var jugadas_len = len(Jugadas)
	for index := 0; index < jugadas_len; index++ {
		jug := Jugadas[index]
		jug.Tipo = pb.EnvioJugada_Jugada
		SumasJugadores[jug.NumJugador.Val-1] += jug.Jugada.Val
		if (CurrentRonda == 3) && (SumasJugadores[jug.NumJugador.Val-1] < 21) {
			jug.Estado = pb.ESTADO_Muerto21
			fmt.Printf("PIUM!, hemos matado al jugador %v.\n", jug.NumJugador.Val)
			CurrentAlivePlayers--
			Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
		} else if jug.Jugada.Val < ValorLider {
			jug.Estado = pb.ESTADO_Vivo
			vivos = append(vivos, jug)
		} else {
			jug.Estado = pb.ESTADO_Muerto
			fmt.Printf("PIUM!, hemos matado al jugador %v.\n", jug.NumJugador.Val)
			CurrentAlivePlayers--
			Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
		}
	}

	// ENVIAR JUGADAS AL NAMENODE
	// Las jugadas y los destinos ya están listos, enviaremos al nameNode
	EnviarJugadasANameNode()

	// Revisamos si es que murieron todos o si queda un jugador
	RevisarSiFinDeJuego(vivos)

	// Debemos avisar a los clientes que esperan que el procesamiento esta listo
	// hay que avisarle a cada uno, por medio del channel "gamerLiderValChan"
	AvisarRespuestaLiderLista(ValorLider)

	// NEXT - ESPERAMOS SE HAYAN ENVIADO LAS RESPUESTAS A TODOS LOS CLIENTES PARA LIMPIAR
	<-waitingForCleaning

	if JuegoActual != pb.JUEGO_None {

		JuegoLucesCleanAndReset()
	} else {

		fmt.Println("Nos Vamos")
		JuegoLucesEndingCleaning()
	}
}

func JuegoLucesCleanAndReset() {
	// LIMPIEZA PREPARACION NUEVO EVENTO

	// 1ERA, 2DA Y 3RA PARTE En BasicCleaning
	// Basicamente reinicia ResponsesCount y RepliesCount
	BasicCleaning()

	// 4TA PARTE - SUMAR RONDA COUNT

	// Sumamos (o reiniciamos) la Ronda
	CurrentRondaMux.Lock()

	CurrentRonda++
	if CurrentRonda <= 3 {
		fmt.Printf("La Ronda a Comenzar: %v\n", CurrentRonda+1)
	} else {
		fmt.Println("Comenzará el siguiente juego.")
	}

	// 5TA PARTE - VER SI JUEGO LUCES TERMINO Y DEBEMOS CAMBIAR

	var waitingResponse int32 = 0
	var liderMsg string
	// Si completamos las rondas, entonces la reiniciamos
	// definimos el waitingResponse en 0 => Nuevo Evento: Nueva Ronda
	// en caso contratio, waitingResponse en 1 => Nuevo Evento: Nuevo Juego
	if CurrentRonda < 4 {

		waitingResponse = 0

		liderMsg = "RONDA"
	} else {

		waitingResponse = 1

		liderMsg = "NUEVA ETAPA"

		CambiarEtapa(pb.JUEGO_TirarCuerda)
	}
	CurrentRondaMux.Unlock()

	// 6TA PARTE - ESPERANDO SEÑAL DEL LIDER

	// Dependiendo de la ronda imprimimos un diferente mensaje
	EsperarAvisoLider(liderMsg)

	// 7RA PARTE - INFORMAR A LOS CLIENTES QUE ESTAN ESPERANDO
	AvisarTerminoCleanAndReset(waitingResponse)
}

func JuegoLucesEndingCleaning() {
	fmt.Println("Limpieza de Final")
	BasicCleaning()

	Jugadores = nil
	Jugadores = make([]int32, 0)
	CurrentAlivePlayers = 0
	waitingCount = 0
}

// FUNCIONES TIRAR CUERDA -------------------------------------------------------------------------------

func JuegoTirarCuerdaWaitForResponses() {
	fmt.Println("Procesando Tirar Cuerda")

	<-serverGameProcessingChan

	liderVal := funcs.RandomInRange(1, 4)
	liderPar := liderVal % 2

	st1 := funcs.ArraySum(ValoresTeam1)
	st2 := funcs.ArraySum(ValoresTeam2)

	st1Par := st1 % 2
	st2Par := st2 % 2

	fmt.Printf("LIDER - Valor: %v - Paridad: %v\n", liderVal, liderPar)
	fmt.Printf("TEAM 1: %v - Suma: %v - Paridad: %v\n", Team1, st1, st1Par)
	fmt.Printf("TEAM 2: %v - Suma: %v - Paridad: %v\n", Team2, st2, st2Par)

	team1res := st1Par == liderPar
	team2res := st2Par == liderPar

	var vivos []*pb.EnvioJugada

	if team1res && team2res {

		for _, jug := range Jugadas {

			jug.Tipo = pb.EnvioJugada_Jugada
			jug.Estado = pb.ESTADO_Vivo

			vivos = append(vivos, jug)
		}

	} else if team1res != team2res {

		fmt.Printf("Team 1: %v - Team 2: %v - Matando al Muerto\n", team1res, team2res)

		var election int32
		if team1res {
			election = 1
		} else {
			election = 2
		}

		for _, jug := range Jugadas {
			jug.Tipo = pb.EnvioJugada_Jugada

			if jug.Equipo == election {

				jug.Estado = pb.ESTADO_Vivo

				vivos = append(vivos, jug)
			} else {
				jug.Estado = pb.ESTADO_Muerto
				fmt.Printf("PIUM!, hemos matado al jugador %v.\n", jug.NumJugador.Val)
				CurrentAlivePlayers--
				Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
			}
		}

	} else if !team1res && !team2res {

		fmt.Println("AMBOS MAL! ELIGIENDO AL AZAR")
		election := funcs.RandomInRange(1, 2)

		fmt.Printf("TEAM ELEGIDO GANADOR: %v\n", election)

		for _, jug := range Jugadas {
			jug.Tipo = pb.EnvioJugada_Jugada

			if jug.Equipo == election {
				jug.Estado = pb.ESTADO_Vivo

				vivos = append(vivos, jug)
			} else {
				jug.Estado = pb.ESTADO_Muerto
				fmt.Printf("PIUM!, hemos matado al jugador %v.\n", jug.NumJugador.Val)
				CurrentAlivePlayers--
				Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
			}
		}
	}

	// ENVIAR JUGADAS AL NAMENODE
	// Las jugadas y los destinos ya están listos, enviaremos al nameNode.
	EnviarJugadasANameNode()

	// Revisamos si es que murieron todos o si queda un jugador
	RevisarSiFinDeJuego(vivos)

	// Debemos avisar a los clientes que esperan que el procesamiento esta listo
	// hay que avisarle a cada uno, por medio del channel "gamerLiderValChan"
	AvisarRespuestaLiderLista(liderVal)

	// NEXT - ESPERAMOS SE HAYAN ENVIADO LAS RESPUESTAS A TODOS LOS CLIENTES PARA LIMPIAR
	<-waitingForCleaning

	if JuegoActual != pb.JUEGO_None {

		JuegoTirarCuerdaCleanAndReset()
	} else {

		JuegoTirarCuerdaEndingCleaning()
	}
}

func JuegoTirarCuerdaCleanAndReset() {

	// 1ERA, 2DA, 3RA PARTE
	// Reiniciamos ResponsesCount, RepliesCount
	BasicCleaning()

	CurrentRondaMux.Lock()

	CurrentRonda++

	CurrentRondaMux.Unlock()

	// 5TA PARTE - VER SI EL JUEGO TERMINO Y DEBEMOS CAMBIAR
	CambiarEtapa(pb.JUEGO_TodoNada)
	liderMsg := "NUEVA ETAPA"

	// 6TA PARTE - ESPERANDO SEÑAL DEL LIDER
	// Dependiendo de la ronda imprimimos un diferente mensaje
	EsperarAvisoLider(liderMsg)

	// 7RA PARTE - INFORMAR A LOS CLIENTES QUE ESTAN ESPERANDO
	var waitingResponse int32 = 1
	AvisarTerminoCleanAndReset(waitingResponse)
}

func JuegoTirarCuerdaEndingCleaning() {
	BasicCleaning()
	jugadorEliminado = -1
}

// FUNCIONES TODO O NADA

func JuegoTodoNadaWaitForResponses() {

	fmt.Println("Procesando Todo o Nada")

	<-serverGameProcessingChan

	// PROCESAMIENTO DE LOGICA "TODO O NADA"

	var liderVal int32 = funcs.RandomInRange(1, 10)

	fmt.Printf("LIDER VAL: %v\n", liderVal)

	var vivos []*pb.EnvioJugada

	var ganadores []int32

	lenAllTeams := len(AllTeams)
	for i := 0; i < lenAllTeams; i++ {
		valTeam1 := funcs.Absoluto(AllTeamsValues[i][0] - liderVal)
		valTeam2 := funcs.Absoluto(AllTeamsValues[i][1] - liderVal)
		if valTeam1 > valTeam2 {
			ganadores = append(ganadores, AllTeams[i][1])
		} else if valTeam1 < valTeam2 {
			ganadores = append(ganadores, AllTeams[i][0])
		} else {
			ganadores = append(ganadores, AllTeams[i][0])
			ganadores = append(ganadores, AllTeams[i][1])
		}
	}
	fmt.Printf("Ganadores: %v\n", ganadores)

	// Actualizamos los estados de los clientes (en el arreglo Jugadas)
	for _, jug := range Jugadas {
		jug.Tipo = pb.EnvioJugada_Jugada

		index := funcs.GetIndexOf(ganadores, jug.NumJugador.Val)

		if index != -1 {
			jug.Tipo = pb.EnvioJugada_Ganador
			jug.Estado = pb.ESTADO_Ganador

			vivos = append(vivos, jug)

		} else {
			jug.Estado = pb.ESTADO_Muerto
			fmt.Printf("PIUM!, hemos matado al jugador %v.\n", jug.NumJugador.Val)
			CurrentAlivePlayers--
			Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
		}
	}

	// ENVIAR JUGADAS AL NAMENODE
	// Las jugadas y los destinos ya están listos, enviaremos al nameNode.
	EnviarJugadasANameNode()

	// Revisamos si es que murieron todos o si queda un jugador
	ImprimirWinners(vivos)

	// Debemos avisar a los clientes que esperan que el procesamiento esta listo
	// hay que avisarle a cada uno, por medio del channel "gamerLiderValChan"
	AvisarRespuestaLiderLista(liderVal)

	// NEXT - ESPERAMOS SE HAYAN ENVIADO LAS RESPUESTAS A TODOS LOS CLIENTES PARA LIMPIAR
	<-waitingForCleaning

	if JuegoActual != pb.JUEGO_None {

		JuegoTodoNadaCleanAndReset()
	} else {

		JuegoTodoNadaEndingCleaning()
	}

}

func JuegoTodoNadaCleanAndReset() {

	// 1ERA, 2DA, 3RA PARTE
	// Reiniciamos ResponsesCount, RepliesCount
	BasicCleaning()

	// Cleaning TODO o NADA
	CambiarEtapa(pb.JUEGO_Fin)
	liderMsg := "FIN"

	// 6TA PARTE - ESPERANDO SEÑAL DEL LIDER
	// Dependiendo de la ronda imprimimos un diferente mensaje
	EsperarAvisoLider(liderMsg)

	// 7RA PARTE - INFORMAR A LOS CLIENTES QUE ESTAN ESPERANDO
	var waitingResponse int32 = 2
	AvisarTerminoCleanAndReset(waitingResponse)
}

func JuegoTodoNadaEndingCleaning() {

	// 1ERA, 2DA, 3RA PARTE
	// Reiniciamos ResponsesCount, RepliesCount
	BasicCleaning()
}

// FUNCIONES COMUNES ----------------------------------------------------------------------------------------------------------------------------

func RevisarSiFinDeJuego(vivos []*pb.EnvioJugada) {
	vivoCount := len(vivos)
	if vivoCount == 0 {
		// Todos los jugadores murieron, terminar juego

		fmt.Printf("Todos los Jugadores Están Muertos - No hay Ganadores: %v - VivoCount: %v\n", CurrentAlivePlayers, vivoCount)
		CambiarEtapa(pb.JUEGO_None)

	} else if vivoCount == 1 {
		// Queda solo 1, es el ganador
		// Modificamos su mensaje para que su estado sea y el mensaje a enviar sean de ganador

		fmt.Printf("TENEMOS UN GANADOR: CurrentAlive %v - VivoCount: %v !!\n", CurrentAlivePlayers, vivoCount)
		vivos[0].Tipo = pb.EnvioJugada_Ganador
		vivos[0].Estado = pb.ESTADO_Ganador
		CambiarEtapa(pb.JUEGO_None)
	}
}

func ImprimirWinners(vivos []*pb.EnvioJugada) {
	fmt.Println("LOS GANADORES!!")
	for _, elem := range vivos {
		fmt.Printf("Ganador: Jugador nº %v.\n", elem.NumJugador.Val)
	}
}

func AvisarRespuestaLiderLista(liderVal int32) {

	// Debemos avisar a los clientes que esperan que el procesamiento esta listo
	// hay que avisarle a cada uno, por medio del channel "gamerLiderValChan"
	var i int32 = 0
	for i = 0; i < ResponsesCount; i++ {

		gameLiderValChan <- liderVal
	}
}

func BasicCleaning() {
	// 1RA PARTE - BORRAMOS LAS JUGADAS DE LA RONDA (Quizas se podrian enviar desde aca)
	Jugadas = nil
	Jugadas = make([]*pb.EnvioJugada, 0)

	// 2DA PARTE - REINICIAR RESPUESTA COUNT
	ResponsesCountMux.Lock()
	ResponsesCount = 0
	ResponsesCountMux.Unlock()

	// 3RA PARTE - REINICIAR REPLIES COUNT
	RepliesCountMux.Lock()
	RepliesCount = 0
	RepliesCountMux.Unlock()
}

func AvisarTerminoCleanAndReset(waitingResponse int32) {
	// Enviamos el waiting response definido anteriormente
	CurrentAlivePlayersMux.Lock()

	var loop int32 = 0
	if jugadorEliminado != -1 {
		loop = CurrentAlivePlayers + 1
	} else {
		loop = CurrentAlivePlayers
	}

	var i int32
	for i = 0; i < loop; i++ {

		nextEventWaitingChan <- waitingResponse
	}
	CurrentAlivePlayersMux.Unlock()
}

func EsperarAvisoLider(liderMsg string) {
	/*
		var liderSignal string
		fmt.Printf("AVISE INICIO SIGUIENTE: %v (Presione ENTER)\n", liderMsg)
		// Se espera un input del humano lider
		fmt.Scanln(&liderSignal)

		fmt.Println()
		fmt.Println()*/
	rondaEnJuegoMux.Lock()
	rondaEnProceso = false
	rondaEnJuegoMux.Unlock()
	<-esperaLiderChan
}

func CambiarEtapa(nEtapa pb.JUEGO) {
	var n, m int
	jugadorEliminado = -1

	// Pre-procesamiento
	if JuegoActual == pb.JUEGO_None {

	} else if JuegoActual == pb.JUEGO_Luces {

	} else if JuegoActual == pb.JUEGO_TirarCuerda {

	} else if JuegoActual == pb.JUEGO_TodoNada {

	} else if JuegoActual == pb.JUEGO_Fin {

	}

	JuegoActual = nEtapa

	// Post-procesamiento
	if JuegoActual == pb.JUEGO_None {

	} else if JuegoActual == pb.JUEGO_Luces {
		SumasJugadores = make([]int32, len(Jugadores))
	} else if JuegoActual == pb.JUEGO_TirarCuerda {
		fmt.Printf("Jugadores que siguen vivos:%v\n", Jugadores)

		fmt.Println("CAMBIANDO Y PREPARANDO JUEGO TIRAR CUERDA")

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(Jugadores), func(i, j int) { Jugadores[i], Jugadores[j] = Jugadores[j], Jugadores[i] }) // se desordenan
		var largo_jugadores int = len(Jugadores)
		if largo_jugadores%2 == 1 {
			jugadorEliminado = Jugadores[0]
			fmt.Printf("Al apretar enter en la siguiente ronda se eliminara por azares el jugador:%v\n", jugadorEliminado)
			CurrentAlivePlayers--
			largo_jugadores--
			Jugadores = Jugadores[1:] //elimina el primero (despues del shuffle por lo que igual es random)
		}
		Team1 = Jugadores[:largo_jugadores/2]
		Team2 = Jugadores[largo_jugadores/2:]

		fmt.Printf("TEAM 1: %v\n", Team1)
		fmt.Printf("TEAM 2: %v\n", Team2)

	} else if JuegoActual == pb.JUEGO_TodoNada {
		fmt.Printf("Jugadores que siguen vivos:%v\n", Jugadores)
		fmt.Println("CAMBIANDO Y PREPARANDO JUEGO TODO O NADA")

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(Jugadores), func(i, j int) { Jugadores[i], Jugadores[j] = Jugadores[j], Jugadores[i] }) // se desordenan
		var largo_jugadores int = len(Jugadores)
		if largo_jugadores%2 == 1 {
			jugadorEliminado = Jugadores[0]
			fmt.Printf("Al apretar enter en la siguiente ronda se eliminara por azares el jugador:%v\n", jugadorEliminado)
			CurrentAlivePlayers--
			largo_jugadores--
			Jugadores = Jugadores[1:] //elimina el primero (despues del shuffle por lo que igual es random)
		}
		//Creando equipos:
		n = largo_jugadores / 2
		m = 2
		AllTeams = make([][]int32, n)
		for i := range AllTeams {
			AllTeams[i] = make([]int32, m)
		}
		AllTeamsValues = make([][]int32, n)
		for i := range AllTeamsValues {
			AllTeamsValues[i] = make([]int32, m)
		}
		for i := 0; i < largo_jugadores; i = i + 2 {
			AllTeams[i/2][0] = Jugadores[i]
			AllTeams[i/2][1] = Jugadores[i+1]
		}
		//Mostrando equipos creados:
		for i := 0; i < len(AllTeams); i++ {
			fmt.Printf("Team %v: [ ", i+1)
			for j := 0; j < len(AllTeams[i]); j++ {
				fmt.Printf("%v ", AllTeams[i][j])
			}
			fmt.Println("]")
		}

	} else if JuegoActual == pb.JUEGO_Fin {
		fmt.Printf("SE ACABO EL JUEGO")
	}
}

func EnviarJugadasANameNode() {
	nAddrs := nameNodeAddress
	if len(os.Args) == 2 {
		nAddrs = "localhost" + nPort
	}

	fmt.Printf("Conectándose al NameNode - Addr: %s\n", nAddrs)

	nconn, nerr := grpc.Dial(nAddrs, grpc.WithInsecure(), grpc.WithBlock())

	if nerr != nil {
		log.Fatalf("did not connect: %v\n", nerr)
	}
	defer nconn.Close()

	nc := pb.NewNameNodeClient(nconn)

	for _, jug := range Jugadas {

		jugadasArray := []*pb.Jugada{jug.Jugada}
		jugJuegoArray := []*pb.JugadasJuego{
			{
				NumJuego: JuegoActual,
				Jugadas:  jugadasArray,
			},
		}

		SolicitudDeRegistro := &pb.SolicitudRegistrarJugadas{
			JugadasJugador: &pb.JugadasJugador{
				NumJugador:   jug.NumJugador,
				JugadasJuego: jugJuegoArray,
			},
		}

		r, err := nc.RegistrarJugadas(context.Background(), SolicitudDeRegistro)

		if err != nil {
			fmt.Printf("Error al solicitar registrar jugadas del jugador: %v - err: %v\n", jug.NumJugador.Val, err)
			continue
		}
		fmt.Printf("Se registró la jugada del Jugador: %v\n", r.NumJugador.Val)
	}
	fmt.Printf("Se registraron las jugadas del Juego: %v - Ronda: %v\n", JuegoActual, CurrentRonda+1)

}

func LiderService() {
	srvAddr := address
	if len(os.Args) == 2 {
		srvAddr = local
	}

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	pb.RegisterLiderServer(s, &server{})
	log.Printf("Juego Inicializado: escuchando en %v\n", lis.Addr())
	waitc <- 0
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
