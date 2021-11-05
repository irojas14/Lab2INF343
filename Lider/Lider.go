package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	funcs "github.com/irojas14/Lab2INF343/Funciones"

	pb "github.com/irojas14/Lab2INF343/Proto"
	"google.golang.org/grpc"
)

const (
	MaxPlayers      = 8
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress     = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

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
	JuegoActual              pb.JUEGO
	gameReadyChan            chan int32 = make(chan int32)
	gameLiderValChan         chan int32 = make(chan int32)
	nextEventWaitingChan     chan int32 = make(chan int32)
	serverGameProcessingChan chan int32 = make(chan int32)
)

// Variables 1er Juego "Luz Roja, Luz Verde (Luces)"

var (
	Jugadas    []*pb.EnvioJugada
	JugadasMux sync.Mutex
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

// UNIRSE-------------------------------------------------------------------------------------------------------------------
func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Printf("Jugador Uniéndose - jugadores actuales: %v\n", waitingCount)

	waitCountMux.Lock()

	if waitingCount > MaxPlayers {
		waitCountMux.Unlock()
		return nil
	}
	waitingCount++
	countCont := waitingCount
	jugNum := &pb.JugadorId{Val: countCont}

	waitCountMux.Unlock()

	fmt.Printf("Jugador Count: %v - jugNum: %v\n", waitingCount, jugNum)

	if jugNum.GetVal() < MaxPlayers {
		res := &pb.RespuestaUnirse{
			MsgTipo:    pb.RespuestaUnirse_Esperar,
			NumJugador: jugNum,
			NumJuego:   pb.JUEGO_None,
			NumRonda:   nil,
		}
		fmt.Printf("Sending Res: %v\n", res.String())
		if err := stream.Send(res); err != nil {
			return err
		}
		fmt.Println("Esperando más jugadores")

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

	} else if jugNum.GetVal() == MaxPlayers {
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

		fmt.Printf("CurrentAlivePlayers: %v\n", CurrentAlivePlayers)
		fmt.Println()
		fmt.Println()

		CambiarEtapa(pb.JUEGO_Luces)

		for i := 0; i < MaxPlayers-1; i++ {
			gameReadyChan <- 0
		}
	}
	return nil
}

// ENVIAR JUGADA-------------------------------------------------------------------------------------------------------------------

func (s *server) EnviarJugada(stream pb.Lider_EnviarJugadaServer) error {

	for {
		// 1RA PARTE - ACUMULACION DE RESPUESTAS

		// Recibidos

		fmt.Printf("Procesando una jugada enviada: Responses: %v\n", ResponsesCount)
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

		fmt.Printf("Response Recibida de: %v\n", in.NumJugador.Val)

		// Recoleccion de jugadas por juego
		JugadasMux.Lock()

		Jugadas = append(Jugadas, in)

		JugadasMux.Unlock()

		fmt.Printf("LIDER ETAPA: %v - JUGADOR %v EN JUEGO %v\n", JuegoActual, in.NumJugador.Val, in.NumJuego)

		if JuegoActual == pb.JUEGO_TirarCuerda {
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
			fmt.Printf("ALL TEAM VALUES: Jugador: %v Equipo: %v - Index: %v\n", in.NumJugador, in.Equipo, index)
			AllTeamsValues[in.Equipo][index] = in.Jugada.Val

			fmt.Printf("Asignación Jugador %v completa\n", in.NumJugador.Val)
			AllTeamsValuesMux.Unlock()
		}

		// Cuenta de Respuestas

		// ResponsesCountMux Lock BEGIN
		ResponsesCountMux.Lock()

		ResponsesCount++

		// Si soy la ultima respuesta esperada, avisar al lider (por medio del channel "inGameWaitingChan")
		if ResponsesCount == CurrentAlivePlayers {

			fmt.Printf("RespuestasListas - Jugador Activador: %v\n", in.GetNumJugador())
			serverGameProcessingChan <- 0
		}

		ResponsesCountMux.Unlock()
		// Espacio Lock END

		// 1era PARTE END - ACUMULACION DE RESPUESTAS ------------------------------------------------------------

		// La recoleccion y respuestas ya se llevo a cabo, esperamos a que esten todas
		fmt.Printf("Esperando Una Jugada del Lider: %v\n", in.GetNumJugador())
		jugadaLider := <-gameLiderValChan
		fmt.Printf("Respuestas del Lider recibida - Con valor: %v\n", jugadaLider)

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
				fmt.Printf("Llenando Mi tipo y estado: Jugador: %v - Tipo: %v - Estado: %v\n", in.NumJugador, res.Tipo, res.Estado)
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
			fmt.Printf("Activando Cleaning: Jugador: %v\n", in.NumJugador)
			waitingForCleaning <- 0
		}

		RepliesCountMux.Unlock()

		// Si la respuesa enviada fue una que termina nuestra participacion en el juego, entonces quebramos y nos vamos
		if res.Tipo == pb.EnvioJugada_Ganador ||
			res.Tipo == pb.EnvioJugada_Fin ||
			res.Estado == pb.ESTADO_Muerto ||
			res.Estado == pb.ESTADO_MuertoDefault ||
			res.Estado == pb.ESTADO_Ganador {
			fmt.Printf("Envio Jugada quebrado por ganador o muerto: Jugador: %v - Tipo: %v - Estado: %v\n", in.NumJugador, res.Tipo, res.Estado)
			break
		}
		// Se enviaron y se cortaron respuestas, ahora esperemos a la limpieza

		// se proceso y se envio la informacion. Ahora el lider preparara el siguiente evento (limpieza), por lo que esperamos
		fmt.Printf("Esperando El Siguiente Evento: %v - Jugador: %v\n", CurrentRonda, in.GetNumJugador())
		waitingRes := <-nextEventWaitingChan
		fmt.Printf("Siguiente Evento Arrivado: %v - Jugador: %v - valorWait: %v\n", CurrentRonda, in.GetNumJugador(), waitingRes)

		// el lider ya termino de limpiar, definir y preparar el siguiente evento, nos envio cual es
		// waitinRes = 2 -> Se acabo el juego, uso principalmente para debugueo
		//           = 1 -> Se viene una nueva etapa
		//           = 0 -> Se viene una nueva ronda

		// 3RA PARTE BEG - INFORMAR NUEVO EVENTO

		if waitingRes == 1 {

			fmt.Printf("Jugador a Nueva Etapa: %v - Lider Juego Etapa: %v - Client Juego In: %v\n", in.GetNumJugador(), JuegoActual, in.NumJuego)

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

			fmt.Printf("Jugador a Nueva Etapa: %v - Juego Etapa: %v - juego enviar en Res2: %v\n", res2.GetNumJugador(), JuegoActual, res2.NumJuego)

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

					fmt.Printf("Jugador Eliminado Al Azar: %v, cerrándolo\n", in.NumJugador.Val)
					res2.Tipo = pb.EnvioJugada_Fin
					res2.Estado = pb.ESTADO_MuertoDefault
					res2.Equipo = 0
					stream.Send(res2)
					return nil
				}
			} else if JuegoActual == pb.JUEGO_TodoNada {
				fmt.Printf("Adentro de else if JuegoActual == pb.JUEGO_TodoNada en EnviarJugada(serverside) - jugadorEliminado: %v\n", jugadorEliminado)

				if jugadorEliminado == in.NumJugador.Val {
					fmt.Printf("Jugador Eliminado Al Azar: %v, cerrándolo\n", in.NumJugador.Val)
					res2.Tipo = pb.EnvioJugada_Fin
					res2.Estado = pb.ESTADO_MuertoDefault
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
							fmt.Printf("Equipo Jugador %v: %v\n", in.NumJugador.Val, i)
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
			fmt.Printf("Enviando Aviso de Comienzo de la Siguiente Ronda: %v - Jugador: %v\n", CurrentRonda, in.NumJugador)

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

func VerMonto() {
	dialAddrs := pozoAddress
	if len(os.Args) == 2 {
		dialAddrs = "localhost:50051"
	}
	fmt.Printf("Consultando Pozo - Addr: %s", dialAddrs)
	// Set up a connection to the server.

	conn, err := grpc.Dial(dialAddrs, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPozoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.VerMonto(ctx, &pb.SolicitudVerMonto{})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Printf("\nEl Monto Acumulado Actual es: %f\n", r.GetMonto())
}

// Para actuaizar el Proto file, correr
// go get -u github.com/irojas14/Lab2INF343/Proto
// O CON
// go get -u github.com/irojas14/Lab2INF343

var waitc chan int = make(chan int)

func main() {
	jugadorEliminado = -1
	go LiderService()
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

var terminarJuego = false

func JuegoLucesWaitForResponses() {
	fmt.Printf("En Juego Luces Responses: Responses: %v - Ronda: %v\n", ResponsesCount, CurrentRonda)
	<-serverGameProcessingChan

	// 1ERA PARTE - DEFINIR ESTADOS (VIVOS, MUERTOS, GANADOR)

	// el lider elige su numero

	/*
			PARA DEBUGEAR, SE DEJA EL RANDOM DEL  LIDER EN NUMBEROS MAS ALTOS!!!
		liderVal := funcs.RandomInRange(6, 10)
	*/
	liderVal := funcs.RandomInRange(10, 14)

	// Recorremos las respuestas colocadas en el arreglo global "Jugadas", definiendo el estado correspondiente
	// Si el valor del jugador es menor que el del lider, entonces vive, si no, muere
	var vivos []*pb.EnvioJugada
	var jugadas_len = len(Jugadas)

	for index := 0; index < jugadas_len; index++ {
		jug := Jugadas[index]
		jug.Tipo = pb.EnvioJugada_Jugada

		if jug.Jugada.Val < liderVal {

			jug.Estado = pb.ESTADO_Vivo

			vivos = append(vivos, jug)
		} else {

			jug.Estado = pb.ESTADO_Muerto

			CurrentAlivePlayers--
			Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
		}
	}

	// Revisamos si es que murieron todos o si queda un jugador
	RevisarSiFinDeJuego(vivos)

	// Debemos avisar a los clientes que esperan que el procesamiento esta listo
	// hay que avisarle a cada uno, por medio del channel "gamerLiderValChan"
	AvisarRespuestaLiderLista(liderVal)

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
	fmt.Println("COMENZADO LIMPIEZA: Juego Luces")

	// 1ERA, 2DA Y 3RA PARTE En BasicCleaning
	// Basicamente reinicia ResponsesCount y RepliesCount
	BasicCleaning()

	fmt.Printf("CurrentAlivePlayers: %v- MaxPlayers: %v - Ronda Terminada: %v\n", CurrentAlivePlayers, MaxPlayers, CurrentRonda+1)

	// 4TA PARTE - SUMAR RONDA COUNT

	// Sumamos (o reiniciamos) la Ronda
	CurrentRondaMux.Lock()

	CurrentRonda++

	fmt.Printf("La Ronda a Comenzar: %v\n", CurrentRonda+1)

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

		CurrentRonda = 0
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
	fmt.Println("Processando Tirar Cuerda")

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

		fmt.Printf("El election: %v\n", election)

		for _, jug := range Jugadas {
			jug.Tipo = pb.EnvioJugada_Jugada

			if jug.Equipo == election {

				jug.Estado = pb.ESTADO_Vivo

				vivos = append(vivos, jug)
			} else {
				jug.Estado = pb.ESTADO_Muerto

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

				CurrentAlivePlayers--
				Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
			}
		}
	}

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

		fmt.Println("Nos Vamos")
		JuegoTirarCuerdaEndingCleaning()
	}
}

func JuegoTirarCuerdaCleanAndReset() {

	fmt.Println("En Juego Tirar Cuerdas Clean And Reset")
	// 1ERA, 2DA, 3RA PARTE
	// Reiniciamos ResponsesCount, RepliesCount
	BasicCleaning()

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

	fmt.Println("Processando Todo o Nada")

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
		fmt.Println()
		fmt.Printf("funcs.Absoluto(AllTeamsValues[%v][0] - liderVal)=%v\n", i, valTeam1)
		fmt.Printf("funcs.Absoluto(AllTeamsValues[%v][1] - liderVal)=%v\n", i, valTeam2)
		fmt.Printf("funcs.Absoluto(AllTeamsValues[%v][0])=%v\n", i, AllTeamsValues[i][0])
		fmt.Printf("funcs.Absoluto(AllTeamsValues[%v][1])=%v\n", i, AllTeamsValues[i][1])
		fmt.Println()
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

			CurrentAlivePlayers--
			Jugadores = funcs.Remove(Jugadores, jug.NumJugador.Val)
		}
	}

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

		fmt.Println("Nos Vamos")
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
		fmt.Printf("Ganador: %v\n", elem)
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

	fmt.Printf("Juego Actual: %v - jugadorEliminado: %v - loop: %v\n", JuegoActual, jugadorEliminado, loop)
	var i int32
	for i = 0; i < loop; i++ {

		nextEventWaitingChan <- waitingResponse
	}
	CurrentAlivePlayersMux.Unlock()
}

func EsperarAvisoLider(liderMsg string) {

	var liderSignal string
	fmt.Printf("AVISE INICIO SIGUIENTE: %v\n", liderMsg)
	// Se espera un input del humano lider
	fmt.Scanln(&liderSignal)

	fmt.Println()
	fmt.Println()

}

func CambiarEtapa(nEtapa pb.JUEGO) {
	var n, m int
	fmt.Printf("CAMBIO DE ETAPA -> EstadoActual: %v - Nuevo: %v\n", JuegoActual, nEtapa)
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

	} else if JuegoActual == pb.JUEGO_TirarCuerda {
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
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/*
	if in.GetNumJuego() == pb.JUEGO_Luces {
		fmt.Printf("VALORES ELEGIDO POR EL LÍDER: %v - EN LA RONDA: %v\n", jugadaLider, CurrentRonda)
		if CurrentRonda < 4 {
			res := &pb.EnvioJugada{
				Tipo:     pb.EnvioJugada_Jugada,
				Rol:      pb.EnvioJugada_Lider,
				NumJuego: pb.JUEGO_Luces,
				Jugada:   &pb.Jugada{Val: jugadaLider},
				NumRonda: &pb.RondaId{Val: CurrentRonda},
			}

			if jugadaLider < in.GetJugada().GetVal() {
				estado = pb.ESTADO_Vivo
				res.Estado = estado
				stream.Send(res)
			} else {
				estado = pb.ESTADO_Muerto
				res.Estado = estado
				stream.Send(res)

				// Jugadores Lock BEG
				JugadoresMux.Lock()
				if len(Jugadores) != 0 {
					Jugadores = funcs.Remove(Jugadores, in.GetNumJugador().Val)
					CurrentAlivePlayers--
				}
				JugadoresMux.Unlock()
				// Jugadores Lock END
				break
			}
		}
	}
	// CONDICIÓN DE TERMINO: GANADOR


	if CurrentAlivePlayers == 1 && estado != pb.ESTADO_Muerto {
		fmt.Printf("Ganaste Jugador: %v\n", in.GetNumJugador())

		youReWinner := &pb.EnvioJugada{
			Tipo:     pb.EnvioJugada_Ganador,
			Rol:      pb.EnvioJugada_Lider,
			NumJuego: pb.JUEGO_Luces,
			Jugada:   &pb.Jugada{Val: jugadaLider},
			NumJugador: in.GetNumJugador(),
			NumRonda: &pb.RondaId{Val: CurrentRonda},
		}
		stream.Send(youReWinner)
		break
	}
*/
