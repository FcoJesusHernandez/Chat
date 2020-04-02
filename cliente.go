package main

import (
	"bufio"
	"container/list"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var clear map[string]func()

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		panic("Error al limpiar pantalla")
	}
}

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
	Archivo  File
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

type File struct {
	Id_conexion    uint64
	Nombre_archivo string
	Datos          []byte
}

type Peticion_clientes struct {
	Tipo    string
	Mensaje Mensaje
	Archivo File
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
		c, error := s.Accept() //cuando acepte la conexion llamara al manejador
		if error != nil {
			fmt.Println(error)
			continue
		}

		peticion := Peticion_clientes{
			Tipo: "NULO",
			//Mensaje: nil,
			//Archivo: nil,
		}

		error2 := gob.NewDecoder(c).Decode(&peticion)
		if error2 != nil {
			fmt.Println(error2)
		}

		if peticion.Tipo == "MENSAJE" {
			msm := Mensaje{
				Id_conexion:     uint64(999),
				Nombre_conexion: "",
				Contenido:       "",
			}

			msm = peticion.Mensaje
			if msm.Id_conexion != 999 {
				CallClear()

				if msm.Id_conexion != Con.Id {
					fmt.Println("( Nuevo Mensaje ", msm.Nombre_conexion, " : ", msm.Contenido, " ) ")
				}

				menuTexto()
				lista_mensajes.PushBack(msm)
			}
		} else if peticion.Tipo == "ARCHIVO" {
			file := File{
				Id_conexion:    uint64(999),
				Nombre_archivo: "",
				Datos:          nil,
			}

			file = peticion.Archivo
			if file.Id_conexion != 999 {
				CallClear()

				if file.Id_conexion != Con.Id {
					fmt.Println("( Nuevo Archivo - ", file.Nombre_archivo, " )")

					nombre_archivo := "docs/CLIENTE_" + Con.Nombre + "_" + time.Now().Format("2006-01-02_15_04_05") + "_" + file.Nombre_archivo
					file2, error := os.Create(nombre_archivo) // retorna el puntero al archivo y si hubiera un error
					if error != nil {
						fmt.Println("No se pudo crear el archivo")
						fmt.Println(error)
						return
					}
					defer file2.Close()
					file2.Write(file.Datos)
				}

				menuTexto()
			}
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
	c, error := net.Dial("tcp", ":9999")
	if error != nil {
		fmt.Println(error)
	} else {

		var nombre_archivo string

		fmt.Println("Nombre del archivo : ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			nombre_archivo = scanner.Text()
		}

		file, error := os.Open(nombre_archivo)
		if error != nil {
			fmt.Println(error)
		}

		stat, error := file.Stat() // regresa las propiedades, estadisticas del archivo, como la cantidad de bytes
		if error != nil {
			fmt.Println("No se puede leer las propiedades del archivo")
			return
		}

		b := make([]byte, stat.Size()) // recervar memoria para el archivo
		file.Read(b)

		archivo := File{
			Id_conexion:    Con.Id,
			Nombre_archivo: nombre_archivo,
			Datos:          b,
		}

		peticion := Peticion{
			Tipo:     "ARCHIVO",
			Conexion: Con,
			Mensaje:  Men,
			Archivo:  archivo,
		}

		error = gob.NewEncoder(c).Encode(peticion)
		if error != nil {
			fmt.Println(error)
		}

		c.Close()
	}
}

func clienteFin() {
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

func menuTexto() {
	fmt.Println("Opciónes")
	fmt.Println("1- Enviar Mensaje")
	fmt.Println("2- Enviar Archivo")
	fmt.Println("3- Mostrar Mensajes")
	fmt.Println("4- Salir")
}

func menu() uint {
	var opcion = uint(0)
	fmt.Println("Opciónes")
	fmt.Println("1- Enviar Mensaje")
	fmt.Println("2- Enviar Archivo")
	fmt.Println("3- Mostrar Mensajes")
	fmt.Println("4- Salir")
	fmt.Scanln(&opcion)
	return uint(opcion)
}

func main() {

	clienteInicio()

	for {
		CallClear()
		switch opcion := menu(); {
		case opcion == uint(1):
			CallClear()
			fmt.Println(" --- Enviar mensaje --- ")
			enviarMensaje()
		case opcion == uint(2):
			CallClear()
			fmt.Println(" --- Enviar archivo --- ")
			enviarArchivo()
		case opcion == uint(3):
			CallClear()
			fmt.Println(" --- Mostrar mensajes --- ")
			muestraMensajes()
		case opcion == uint(4):
			CallClear()
			fmt.Println(" --- Salir --- ")
			clienteFin()
			return
			break
		default:
			fmt.Println("Opción no valida")
		}
		var pausa = ""
		fmt.Println("Presione una tecla y enter para continuar ")
		fmt.Scanln(&pausa)
	}
}
