package Funciones

// 	funcs "github.com/irojas14/Lab2INF343/Funciones"

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"

	pb "github.com/irojas14/Lab2INF343/Proto"
)
func Is_in_folder(file_name string, ruta string) bool{
	all_files, err := ioutil.ReadDir(ruta)
    if err != nil {
        log.Fatal(err)
    }
    for _, f := range all_files {
    	if f.Name() == file_name{
    		return true
    	}
    }
    return false
}
func RandomInRange(min int32, max int32) int32 {
	return rand.Int31n(max-min+1) + min
}

func CrearArchivoTxt(NombreDelArchivo string) {
	// check if file exists
	var _, err = os.Stat(NombreDelArchivo)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(NombreDelArchivo)
		if isError(err) {
			return
		}
		defer file.Close()
	}

	fmt.Println("Se crea el archivo llamado: ", NombreDelArchivo)
}

func InsertarJugadasDelJugador(NombreDelArchivo string, JugadasDelJugador []int32) {
	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(NombreDelArchivo, os.O_RDWR, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	for _, jugada := range JugadasDelJugador {
		// Hacer algo
		// Write some text line-by-line to file.
		_, err = file.WriteString(FormatInt32(jugada))
		if isError(err) {
			return
		}
	}

	err = file.Sync()
	if isError(err) {
		return
	}
}

func LeerArchivo(NombreDelArchivo string) {
	// Open file for reading.
	var file, err = os.OpenFile(NombreDelArchivo, os.O_RDWR, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	// Read file, line by line
	var text = make([]byte, 1024)
	for {
		_, err = file.Read(text)

		// Break if finally arrived at end of file
		if err == io.EOF {
			break
		}

		// Break if error occured
		if err != nil && err != io.EOF {
			isError(err)
			break
		}
	}

	fmt.Println("Reading from file.")
	fmt.Println(string(text))
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func FormatInt32(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func Remove(slice []int32, num_jugador int32) []int32 {

	s := GetIndexOf(slice, num_jugador)
	return append(slice[:s], slice[s+1:]...)
}

func GetIndexOf(slice []int32, elem int32) int {
	for i, e := range slice {
		if elem == e {
			return i
		}
	}
	return -1
}

func FindEnvioJugada(slice []*pb.EnvioJugada, num_jugador int32) *pb.EnvioJugada {
	for _, jug := range slice {
		if jug.NumJugador.Val == num_jugador {
			return jug
		}
	}
	return nil
}

func ArraySum(array []int32) int32 {
	var result int32 = 0
	for _, v := range array {
		result += v
	}
	return result
}

func Absoluto(n int32) int32 {
	if n > 0 {
		return n
	}
	return -1 * n
}

func DeleteFile(path string) {
    // delete file
    var err = os.Remove(path)
    if isError(err) {
        return
    }

    fmt.Println("Archivo eliminado")
}

/*
func itos(int32) string {
	i := 10
	s1 := strconv.FormatInt(int64(i), 10)
	s2 := strconv.Itoa(i)
	return s2
	fmt.Printf("%v, %v\n", s1, s2)
}
*/