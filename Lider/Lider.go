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
	MaxPlayers      = 5
	nameNodeAddress = "dist152.inf.santiago.usm.cl:50050"
	pozoAddress     = "dist150.inf.santiago.usm.cl:50051"
	sPort           = ":50052"
	address         = "dist149.inf.santiago.usm.cl" + sPort
	local           = "localhost" + sPort
)

var waitCountMux sync.Mutex
var waitingCount int32 = 0

var (
	CurrentAlivePlayers int32 = 0
	CurrentAlivePlayersMux sync.Mutex

	CurrentRondaMux     sync.Mutex
	CurrentRonda        int32 = 0
)

var (
	ResponsesCount    int32 = 0
	ResponsesCountMux sync.Mutex
)

var (
	RepliesCount int32 = 0
	RepliesCountMux sync.Mutex
	waitingForCleaning chan int32 = make(chan int32)	
)

var (
	JuegoActual       pb.JUEGO
	gameReadyChan     chan int32 = make(chan int32)
	gameLiderValChan  chan int32 = make(chan int32)
	nextEventWaitingChan chan int32 = make(chan int32)
	serverGameProcessingChan chan int32 = make(chan int32)
)

// Variables 1er Juego "Luz Roja, Luz Verde (Luces)"

var (
	Jugadas []*pb.EnvioJugada
	JugadasMux sync.Mutex
)

// Variables 2do Juego "Tirar la Cuerda"

var (
	Jugadores []int32
	Team1 []int32
	ValoresTeam1 []int32
	
	Team2 []int32
	ValoresTeam2 []int32
	
	jugadorEliminado int32

	JugadoresMux sync.Mutex
)

type server struct {
	pb.UnimplementedLiderServer
}

// METODOS DEL LIDER SERVER

// UNIRSE-------------------------------------------------------------------------------------------------------------------
func (s *server) Unirse(in *pb.SolicitudUnirse, stream pb.Lider_UnirseServer) error {
	fmt.Printf("Jugador Uniéndose - jugadores actuales: %v\n", waitingCount)

	if waitingCount > MaxPlayers {
		return nil
	}

	waitCountMux.Lock()
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
		
		fmt.Printf("Procesando una jugada enviada: Responses: %v\n", ResponsesCount)
		
		// Recibidos

		// Procesamiento estandar
		in, err := stream.Recv()

		if err == io.EOF {
			fmt.Printf("Un error EOF: %v\n", err)
			return nil
		}

		if err != nil {
			fmt.Printf("Un Error no EOF: %v\n", err)
			return err
		}
		
		// Recoleccion de jugadas por juego
		JugadasMux.Lock()
		
		Jugadas = append(Jugadas, in)
		
		JugadasMux.Unlock()
		
		if (JuegoActual == pb.JUEGO_TirarCuerda) {
			// Tirar Cuerda

			if (in.Equipo == 1) {
				ValoresTeam1 = append(ValoresTeam1, in.Jugada.Val)

			} else {
				ValoresTeam2 = append(ValoresTeam2, in.Jugada.Val)
			}
		}

		// Cuenta de Respuestas

		// ResponsesCountMux Lock BEGIN
		ResponsesCountMux.Lock()
		
		ResponsesCount++
		//fmt.Printf("Responses Tomada en Cuenta: %v - Jugador: %v\n", ResponsesCount, in.GetNumJugador())
		//fmt.Printf("Responses == Current?: %v\n", ResponsesCount == CurrentAlivePlayers)
		
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
		for index:= 0; index < jugadas_len; index++ {
			jug := Jugadas[index]
			if (num_jugador == jug.NumJugador.Val) {
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
			waitingForCleaning<- 0
		}

		RepliesCountMux.Unlock()

		// Si la respuesa enviada fue una que termina nuestra participacion en el juego, entonces quebramos y nos vamos
		if (res.Tipo == pb.EnvioJugada_Ganador ||
		res.Tipo == pb.EnvioJugada_Fin ||
		res.Estado == pb.ESTADO_Muerto ||
		res.Estado == pb.ESTADO_Ganador) {
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

			fmt.Printf("Nueva Etapa: %v\n", in.GetNumJugador())

			var equipo int32 = 0;

			res2 := &pb.EnvioJugada{
				Tipo:     pb.EnvioJugada_NuevoJuego,
				Rol:      pb.EnvioJugada_Lider,
				NumJuego: JuegoActual,
				Jugada:   nil,
				NumJugador: in.GetNumJugador(),
				NumRonda: &pb.RondaId{Val: CurrentRonda},
				Equipo: equipo,
			}
			if (JuegoActual == pb.JUEGO_TirarCuerda) {

				searchRes := funcs.GetIndexOf(Team1, in.NumJugador.Val)
				if (searchRes != -1) {

					res2.Equipo = 1;
				} else {

					res2.Equipo = 2
				}

				if (jugadorEliminado == in.NumJugador.Val) {

					res2.Tipo = pb.EnvioJugada_Fin
					res2.Estado = pb.ESTADO_Muerto
					res2.Equipo = 0
					stream.Send(res2)
					break
				}
			}
			stream.Send(res2)						

		} else if waitingRes == 2 {
			
			fmt.Printf("Juego Acabado: Au Revoir Jugador: %v\n", in.GetNumJugador())
			fin := &pb.EnvioJugada{
				Tipo:     pb.EnvioJugada_Fin,
				Rol:      pb.EnvioJugada_Lider,
				NumJuego: pb.JUEGO_Luces,
				Jugada:   &pb.Jugada{Val: jugadaLider},
				NumJugador: in.GetNumJugador(),
			}
			stream.Send(fin)
			break

		} else {
			fmt.Printf("Enviando Aviso de Comienzo de la Siguiente Ronda: %v - Jugador: %v\n", CurrentRonda, in.NumJugador)
			nuevaRonda := &pb.EnvioJugada{
				Tipo:     pb.EnvioJugada_NuevaRonda,
				NumRonda: &pb.RondaId{Val: CurrentRonda},
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

	go LiderService()
	go Update()
	<-waitc
}

func Update() {
	for {
		if terminarJuego {
			fmt.Println(("Terminando Juego"))
			break;
		}
		if JuegoActual == pb.JUEGO_Luces {
			JuegoLucesWaitForResponses()

		} else if JuegoActual == pb.JUEGO_TirarCuerda {
			JuegoTirarCuerdaWaitForResponses()
		}
	}
}

var terminarJuego = false;

func JuegoLucesWaitForResponses() {
	fmt.Printf("En Juego Luces Responses: Responses: %v - Ronda: %v\n", ResponsesCount, CurrentRonda)
	<-serverGameProcessingChan

	// 1ERA PARTE - DEFINIR ESTADOS (VIVOS, MUERTOS, GANADOR)

	// el lider elige su numero
	liderVal := funcs.RandomInRange(6, 10)

	// Recorremos las respuestas colocadas en el arreglo global "Jugadas", definiendo el estado correspondiente
	// Si el valor del jugador es menor que el del lider, entonces vive, si no, muere
	var vivos []*pb.EnvioJugada;	
	var jugadas_len = len(Jugadas)
	
	for index := 0; index < jugadas_len; index++ {
		jug := Jugadas[index]
		jug.Tipo = pb.EnvioJugada_Jugada

		if (jug.Jugada.Val < liderVal) {

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

	if (JuegoActual != pb.JUEGO_None) {

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

	fmt.Printf("CurrentAlivePlayers: %v- MaxPlayers: %v - Ronda Terminada: %v\n", CurrentAlivePlayers, MaxPlayers, CurrentRonda)

	// 4TA PARTE - SUMAR RONDA COUNT

	// Sumamos (o reiniciamos) la Ronda
	CurrentRondaMux.Lock()
	
	CurrentRonda++

	fmt.Printf("La Ronda a Comenzar: %v\n", CurrentRonda)

	// 5TA PARTE - VER SI JUEGO LUCES TERMINO Y DEBEMOS CAMBIAR

	var waitingResponse int32
	var  liderMsg string
	// Si completamos las rondas, entonces la reiniciamos
		// definimos el waitingResponse en 0 => Nuevo Evento: Nueva Ronda
	// en caso contratio, waitingResponse en 1 => Nuevo Evento: Nuevo Juego
	if (CurrentRonda < 4) {

		waitingResponse = 0
		
		liderMsg = "RONDA"
	} else {

		fmt.Println("Juego Luces Terminado")
		
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

	var vivos []*pb.EnvioJugada;
	
	if (team1res) {
		fmt.Println("TEAM 1 VIVO")

		for _, num_jugador := range(Team1) {
			jug := funcs.FindEnvioJugada(Jugadas, num_jugador)
			
			jug.Tipo = pb.EnvioJugada_Jugada
			jug.Estado = pb.ESTADO_Vivo
			vivos = append(vivos, jug)
		}
	}
	if (team2res) {
		fmt.Println("TEAM 2 VIVO")

		for _, num_jugador := range(Team2) {
			jug := funcs.FindEnvioJugada(Jugadas, num_jugador)
			
			jug.Tipo = pb.EnvioJugada_Jugada
			jug.Estado = pb.ESTADO_Vivo
			vivos = append(vivos, jug)
		}

	}

	if (!team1res && !team2res) {

		fmt.Println("AMBOS MAL! ELIGIENDO AL AZAR")
		election := funcs.RandomInRange(1, 2)
		
		var selTeam []int32
		if election == 1 {
			selTeam = Team1
		} else {
			selTeam = Team2
		}

		fmt.Printf("TEAM ELEGIDO GANADOR: %v -> %v\n", election, selTeam)

		for _, jug := range(Jugadas) {
			jug.Tipo = pb.EnvioJugada_Jugada

			indexres := funcs.GetIndexOf(selTeam, jug.NumJugador.Val)
			if (indexres != -1) {

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

	if (JuegoActual != pb.JUEGO_None) {

		JuegoTirarCuerdaCleanAndReset()
	} else {

		fmt.Println("Nos Vamos")
		JuegoTirarCuerdaEndingCleaning()
	}
}

func JuegoTirarCuerdaCleanAndReset() {

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
}

// FUNCIONES TODO O NADA

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
		var i int32	
		for i = 0; i < CurrentAlivePlayers; i++ {
	
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
	fmt.Printf("EstadoActual: %v - Nuevo: %v\n", JuegoActual, nEtapa)

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
		if (largo_jugadores % 2 == 1){
			jugadorEliminado = Jugadores[largo_jugadores-1]

			CurrentAlivePlayers--
			Jugadores = Jugadores[:largo_jugadores-1] //elimina el ultimo (despues del shuffle por lo que igual es random)
		}
		Team1 = Jugadores[:largo_jugadores/2]
		Team2 = Jugadores[largo_jugadores/2:]

		fmt.Printf("TEAM 1: %v\n", Team1)
		fmt.Printf("TEAM 2: %v\n", Team2)

	} else if JuegoActual == pb.JUEGO_TodoNada {

	} else if JuegoActual == pb.JUEGO_Fin {
		
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