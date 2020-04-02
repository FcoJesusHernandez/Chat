package main

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"io/ioutil"
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

var lista_datos_conexiones list.List
var lista_conexiones list.List
var lista_mensajes list.List
var lista_archivos list.List
var lista_peticiones list.List

var indice uint64 = 1

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

type File_respaldo struct {
	Id_conexion             uint64
	Nombre_archivo_original string
	Nombre_adaptado         string
}

type Peticion_clientes struct {
	Tipo    string
	Mensaje Mensaje
	Archivo File
}

func server() {
	s, error := net.Listen("tcp", ":9999") // c = conexion escuchando en el puerto
	if error != nil {
		fmt.Println(error)
		return
	}
	fmt.Println("Servicio activo")
	for {
		c, error := s.Accept() //cuando acepte la conexion llamara al manejador
		if error != nil {
			fmt.Println(error)
			continue
		}

		go handleCliente(c)
	}
}

func handleCliente(c net.Conn) {
	var Peticion Peticion

	error := gob.NewDecoder(c).Decode(&Peticion)
	//fmt.Println(Peticion)
	if error != nil {
		fmt.Println(error)
	} else {
		lista_peticiones.PushBack(Peticion)

		if Peticion.Tipo == "INICIO" {
			lista_conexiones.PushBack(c)

			Peticion.Conexion.Id = indice

			error = gob.NewEncoder(c).Encode(indice)
			if error != nil {
				fmt.Println(error)
			} else {
				lista_datos_conexiones.PushBack(Peticion.Conexion)
				indice++
				CallClear()
				mostrarPeticiones()
				menuTexto()
			}
		} else if Peticion.Tipo == "FIN" {
			CallClear()
			mostrarPeticiones()
			menuTexto()
			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(Conexion).Id == Peticion.Conexion.Id {
					lista_datos_conexiones.Remove(e)
				}
			}
		} else if Peticion.Tipo == "MENSAJE" {
			CallClear()

			lista_mensajes.PushBack(Peticion.Mensaje)
			mostrarPeticiones()
			menuTexto()
			gob.NewEncoder(c).Encode(Peticion)

			Peticion_para_cliente := Peticion_clientes{
				Tipo:    "MENSAJE",
				Mensaje: Peticion.Mensaje,
				//Archivo: "",
			}

			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(Conexion).Activo == true {
					puerto := "990" + strconv.FormatUint(e.Value.(Conexion).Id, 10)

					c2, error := net.Dial("tcp", ":"+puerto)

					if error != nil {
						fmt.Println(error)
					} else {
						error2 := gob.NewEncoder(c2).Encode(Peticion_para_cliente)
						if error2 != nil {
							fmt.Println(error2)
							lista_datos_conexiones.Remove(e)
						} else {
							//fmt.Println("Mensaje enviado")
						}
					}
				}
			}
		} else if Peticion.Tipo == "ARCHIVO" {
			nombre_archivo := "docs/SERVIDOR_" + time.Now().Format("2006-01-02_15_04_05") + "_" + Peticion.Archivo.Nombre_archivo
			file, error := os.Create(nombre_archivo) // retorna el puntero al archivo y si hubiera un error
			if error != nil {
				fmt.Println("No se pudo crear el archivo")
				fmt.Println(error)
				return
			}
			defer file.Close()
			file.Write(Peticion.Archivo.Datos)

			new_archivo := File_respaldo{
				Id_conexion:             Peticion.Conexion.Id,
				Nombre_archivo_original: Peticion.Archivo.Nombre_archivo,
				Nombre_adaptado:         nombre_archivo,
			}
			lista_archivos.PushBack(new_archivo)

			Peticion_para_cliente := Peticion_clientes{
				Tipo: "ARCHIVO",
				//Mensaje: nil,
				Archivo: Peticion.Archivo,
			}

			for e := lista_datos_conexiones.Front(); e != nil; e = e.Next() {
				if e.Value.(Conexion).Activo == true {
					puerto := "990" + strconv.FormatUint(e.Value.(Conexion).Id, 10)

					c2, error := net.Dial("tcp", ":"+puerto)

					if error != nil {
						fmt.Println(error)
					} else {
						error2 := gob.NewEncoder(c2).Encode(Peticion_para_cliente)
						if error2 != nil {
							fmt.Println(error2)
							lista_datos_conexiones.Remove(e)
						} else {
							//fmt.Println("Mensaje enviado")
						}
					}
				}
			}
			CallClear()
			mostrarPeticiones()
			menuTexto()
		} else {
			fmt.Println("Petición desconocida")
		}
		c.Close()
	}
}

func menuTexto() {
	fmt.Println("Opciónes")
	fmt.Println("1- Mostrar los mensajes/nombre de los archivos enviados")
	fmt.Println("2- Opción para respaldar en un archivo de texto los mensajes/nombre de los archivos enviados.")
	fmt.Println("3- Terminar servidor")
}

func menu() uint {
	var opcion = uint(0)
	fmt.Println("Opciónes")
	fmt.Println("1- Mostrar los mensajes/nombre de los archivos enviados")
	fmt.Println("2- Opción para respaldar en un archivo de texto los mensajes/nombre de los archivos enviados.")
	fmt.Println("3- Terminar servidor")
	fmt.Scanln(&opcion)
	return uint(opcion)
}

func mostrarPeticiones() {
	fmt.Println("Registro de peticiones")
	fmt.Println("-------------------------------------")
	fmt.Println("")
	for e := lista_peticiones.Front(); e != nil; e = e.Next() {
		if e.Value.(Peticion).Tipo == "INICIO" {
			fmt.Println("Usuario conectado - ", e.Value.(Peticion).Conexion.Nombre)
		} else if e.Value.(Peticion).Tipo == "FIN" {
			fmt.Println("Usuario desconectado - ", e.Value.(Peticion).Conexion.Nombre)
		} else if e.Value.(Peticion).Tipo == "MENSAJE" {
			fmt.Println("Mensaje - ", e.Value.(Peticion).Mensaje.Nombre_conexion, " : ", e.Value.(Peticion).Mensaje.Contenido)
		} else if e.Value.(Peticion).Tipo == "ARCHIVO" {
			fmt.Println("Archivo - ", e.Value.(Peticion).Archivo.Nombre_archivo)
		} else {
			fmt.Println("Petición desconocida - ")
		}
	}
	fmt.Println("")
	fmt.Println("-------------------------------------")
	fmt.Println("")
}

func mostrarMensajes() {
	fmt.Println("Mensajes")
	fmt.Println("-------------------------------------")
	fmt.Println("")
	for e := lista_mensajes.Front(); e != nil; e = e.Next() {
		origen := ""
		if e.Value.(Mensaje).Id_conexion == 999 {
			origen = " ( Archivado )"
		}
		fmt.Println(e.Value.(Mensaje).Nombre_conexion, origen, " dice : ", e.Value.(Mensaje).Contenido)
	}
	fmt.Println("")
	fmt.Println("-------------------------------------")
	fmt.Println("")
}

func respaldarMensajes() {
	os.Remove("mensajes.txt")

	file, error := os.Create("mensajes.txt")
	if error != nil {
		fmt.Println("No se pudo crear el archivo")
		return
	}

	for e := lista_mensajes.Front(); e != nil; e = e.Next() {
		file.WriteString(e.Value.(Mensaje).Nombre_conexion + " | " + e.Value.(Mensaje).Contenido + " \n")
	}

	file.Close()
}

func cargarMensajes() {
	data, error := ioutil.ReadFile("mensajes.txt")
	if error != nil {
		fmt.Println("No se puede leer el archivo")
		return
	}
	s := string(data)

	nombre, mensaje := "", ""
	terminado1 := false

	for i := 0; i < len(data); i++ {
		if s[i] != '\n' {
			if !terminado1 {
				if s[i] == '|' {
					terminado1 = true
					continue
				}
				nombre = nombre + string(s[i])
			} else {
				mensaje = mensaje + string(s[i])
			}
		} else {
			msm := Mensaje{
				Id_conexion:     999,
				Nombre_conexion: nombre,
				Contenido:       mensaje,
			}
			lista_mensajes.PushBack(msm)

			nombre = ""
			mensaje = ""
			terminado1 = false
		}
	}
}

func respaldarArchivos() {
	os.Remove("archivos.txt")

	file, error := os.Create("archivos.txt")
	if error != nil {
		fmt.Println("No se pudo crear el archivo")
		return
	}

	for e := lista_archivos.Front(); e != nil; e = e.Next() {
		file.WriteString(e.Value.(File_respaldo).Nombre_archivo_original + " | " + e.Value.(File_respaldo).Nombre_adaptado + " \n")
	}

	file.Close()
}

func cargarArchivos() {
	data, error := ioutil.ReadFile("archivos.txt")
	if error != nil {
		fmt.Println("No se puede leer el archivo")
		return
	}
	s := string(data)

	nombre_archivo_original, nombre_adaptado := "", ""
	terminado1 := false

	for i := 0; i < len(data); i++ {
		if s[i] != '\n' {
			if !terminado1 {
				if s[i] == '|' {
					terminado1 = true
					continue
				}
				nombre_archivo_original = nombre_archivo_original + string(s[i])
			} else {
				nombre_adaptado = nombre_adaptado + string(s[i])
			}
		} else {

			archivo := File_respaldo{
				Id_conexion:             uint64(999),
				Nombre_archivo_original: nombre_archivo_original,
				Nombre_adaptado:         nombre_adaptado,
			}
			lista_archivos.PushBack(archivo)

			nombre_archivo_original = ""
			nombre_adaptado = ""
			terminado1 = false
		}
	}
}

func mostrarArchivos() {
	fmt.Println("Archivos")
	fmt.Println("-------------------------------------")
	fmt.Println("")
	for e := lista_archivos.Front(); e != nil; e = e.Next() {
		origen := ""
		if e.Value.(File_respaldo).Id_conexion == 999 {
			origen = " ( Archivado )"
		}

		fmt.Println(e.Value.(File_respaldo).Nombre_archivo_original, origen, " renombrado como : ", e.Value.(File_respaldo).Nombre_adaptado)
	}
	fmt.Println("")
	fmt.Println("-------------------------------------")
	fmt.Println("")
}

func main() {
	go server()
	cargarMensajes()
	cargarArchivos()
	for {
		CallClear()
		mostrarPeticiones()
		switch opcion := menu(); {
		case opcion == uint(1):
			CallClear()
			mostrarMensajes()
			mostrarArchivos()
		case opcion == uint(2):
			CallClear()
			respaldarMensajes()
			respaldarArchivos()
		case opcion == uint(3):
			CallClear()
			respaldarMensajes()
			respaldarArchivos()
			//salir()
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
