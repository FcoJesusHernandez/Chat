package main

import (
	"bufio"
	"container/list"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
)

var Con Conexion = Conexion{
	Id:     999,
	Nombre: "",
	Activo: true,
}

var Men Mensaje
var lista_mensajes list.List

type Peticion struct {
	Tipo     string
	Conexion Conexion
	Mensaje  Mensaje
}

type Conexion struct {
	Id     uint64
	Nombre string
	Activo bool
}

type Mensaje struct {
	Id_conexion     uint64
	Nombre_conexion string
	Contenido       string
}

func enviarMensaje() {
	var texto string

	fmt.Println("Mensaje : ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		texto = scanner.Text()
	}

	Men.Contenido = texto
	Men.Id_conexion = Con.Id
	Men.Nombre_conexion = Con.Nombre

	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {
		//Con.Con = c

		peticion := Peticion{
			Tipo:     "MENSAJE",
			Conexion: Con,
			Mensaje:  Men,
		}

		//fmt.Println(peticion)

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func esperaMensajes() {
	puerto := "990" + strconv.FormatUint(Con.Id, 10)

	s, error := net.Listen("tcp", ":"+puerto) // c = conexion escuchando en el puerto
	if error != nil {
		fmt.Println(error)
		return
	}
	for {
		msm := Mensaje{
			Id_conexion:     uint64(999),
			Nombre_conexion: "",
			Contenido:       "",
		}
		c, error := s.Accept() //cuando acepte la conexion llamara al manejador
		if error != nil {
			fmt.Println(error)
			continue
		}

		error2 := gob.NewDecoder(c).Decode(&msm)
		if error2 != nil {
			fmt.Println(error2)
		}

		if msm.Id_conexion != 999 {
			fmt.Println("Nuevo Mensaje")
			lista_mensajes.PushBack(msm)
			muestraMensajes()
		}
	}
}

func muestraMensajes() {
	usuario_nombre := ""
	for e := lista_mensajes.Front(); e != nil; e = e.Next() {
		if e.Value.(Mensaje).Id_conexion == Con.Id {
			usuario_nombre = "Yo"
		} else {
			usuario_nombre = e.Value.(Mensaje).Nombre_conexion
		}
		fmt.Println(usuario_nombre, " : ", e.Value.(Mensaje).Contenido)
	}
}

func enviarArchivo() {
	/*c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {

	}*/
}

func clienteFin() {
	fmt.Println("Salir")

	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {

		peticion := Peticion{
			Tipo:     "FIN",
			Conexion: Con,
			Mensaje:  Men,
		}

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func clienteInicio() {
	c, error := net.Dial("tcp", ":9999")

	if error != nil {
		fmt.Println(error)
	} else {
		var nombre string

		fmt.Print("Cual es tu nombre : ")

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			nombre = scanner.Text()
		}

		if c == nil {
			fmt.Println("Conexion no encontrada")
		} else {
			Con.Nombre = nombre

			peticion := Peticion{
				Tipo:     "INICIO",
				Conexion: Con,
				Mensaje:  Men,
			}

			error := gob.NewEncoder(c).Encode(peticion)
			if error != nil {
				fmt.Println(error)
			}

			error = gob.NewDecoder(c).Decode(&Con.Id)

			if Con.Id == 999 {
				fmt.Println("Error con inicialización")
			} else {
				fmt.Println("Inicilización exitosa")
				go esperaMensajes()
			}

			if error != nil {
				fmt.Println(error)
			}

			c.Close()
		}
	}
}

func menu() uint {
	var opcion = uint(0)
	fmt.Println("Opciónes")
	fmt.Println("1- Enviar Mensaje")
	fmt.Println("2- Enviar Archivo")
	fmt.Println("3- Salir")
	fmt.Scanln(&opcion)
	return uint(opcion)
}

func main() {

	clienteInicio()

	for {
		switch opcion := menu(); {
		case opcion == uint(1):
			enviarMensaje()
		case opcion == uint(2):
			enviarArchivo()
		case opcion == uint(3):
			clienteFin()
			return
			break
		default:
			fmt.Println("Opción no valida")
		}
	}
}
