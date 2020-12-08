package main

import (
	"bufio"        
	"encoding/gob" 
	"fmt"          
	"io"           
	"io/ioutil"    
	"log"          
	"net"          
	"os"               
	"strconv" 
	"strings" 
)

type Cliente struct {
	Nickname     string
	numeroPuerto int
}

const TAMANIOBUFFER = 65495

var cliente Cliente

func registroCliente() Cliente { 
	var nickname string
	fmt.Print("Ingrese su nickname: ")
	fmt.Scan(&nickname)
	aux := obtenerPuerto()
	cliente := Cliente{
		Nickname:     nickname,
		numeroPuerto: aux,
	}
	go conexionServidor(cliente)
	go conexionServidorMensajes()
	return cliente
}

func obtenerPuerto() int { //PUERTO :9996 PARA ASIGNAR PUERTOS UNICOS A LOS CLIENTES
	var aux int
	c, err := net.Dial("tcp", ":9996")
	if err != nil {
		fmt.Println(err)
		return aux
	}
	err = gob.NewDecoder(c).Decode(&aux)
	if err != nil {
		fmt.Println(err)
	}
	c.Close()
	return aux
}

func conexionServidor(cliente Cliente) {
	c, err := net.Dial("tcp", ":9999") //PUERTO :9999 PARA ENTRAR AL SERVIDOR
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gob.NewEncoder(c).Encode(cliente)
	if err != nil {
		fmt.Println(err)
	}
	c.Close()
}

func conexionServidorMensajes() { //SE HACE CONEXION CON EL SERVIDOR, MEDIANTE EL PUERTO ESPECIFICO DEL CLIENTE PARA RECIBIR MENSAJES
	aux := strconv.Itoa(cliente.numeroPuerto)
	puerto := ":" + aux
	s, err := net.Listen("tcp", puerto)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var mensaje string
		err = gob.NewDecoder(c).Decode(&mensaje)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(mensaje)
	}
}

func enviarMensaje(mensaje string) { //PUERTO :9998 PARA ENVIAR MENSAJES AL SERVIDOR
	c, err := net.Dial("tcp", ":9998")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Enviando mensaje. . . ")
	err = gob.NewEncoder(c).Encode(mensaje)
	if err != nil {
		fmt.Println(err)
	}
	c.Close()
}

func seleccionarArchivo() string {
	var opcion string
	var archivoSeleccionado string
	for {
		fmt.Println("------------------------------")
		fmt.Println("1) Ver archivos disponibles")
		fmt.Println("2) Enviar archvio")
		fmt.Println("3) Enviar archivo(escribir ubicación)")
		fmt.Println("4) Salir")
		fmt.Print("Opción: ")
		fmt.Scan(&opcion)
		if opcion == "1" {
			directorio, err := ioutil.ReadDir("./")
			if err != nil {
				log.Fatal(err)
			}
			for i, a := range directorio {
				fmt.Println("(" + strconv.Itoa(i) + ") -> " + a.Name())
			}
			pausa() //COMENAR O DESCOMENTAR SI ES ENCESARIO
		} else if opcion == "2" {
			directorio, err := ioutil.ReadDir("./")
			if err != nil {
				log.Fatal(err)
			}
			for i, a := range directorio {
				fmt.Println("(" + strconv.Itoa(i) + ") -> " + a.Name())
			}
			var opcionArchivo int
			fmt.Print("Seleccione el numero del archivo a enviar: ")
			fmt.Scan(&opcionArchivo)
			for i, a := range directorio {
				if opcionArchivo == i {
					archivoSeleccionado = a.Name()
					fmt.Println("Archivo seleccionado: ", archivoSeleccionado)
					break
				}
			}
			break
		} else if opcion == "3" {
			fmt.Println("Función no disponible por el momento")
			break
		} else if opcion == "4" {
			break
		} else {
			fmt.Println("Opción no valida")
		}
	}
	return archivoSeleccionado
}

func conexionServidorArchivos(archivo string) { //PUERTO :9995 PARA ENVIAR ARCHIVOS AL SERVIDOR
	c, err := net.Dial("tcp", ":9995")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	if err != nil {
		fmt.Println(err)
	}
	enviarArchivo(c, archivo)
}

func enviarArchivo(c net.Conn, archivo string) { //FUNCION PARA ENVIAR UN ARCHIVO AL SERVIDOR
	auxArchivo, err := os.Open(archivo) // se abre el archivo
	if err != nil {
		fmt.Println(err)
		return
	}

	datosArchivo, err := auxArchivo.Stat() // se lee la informacion del archivo
	if err != nil {
		fmt.Println(err)
		return
	}

	tamanioArchivo := complementarCadena(strconv.FormatInt(datosArchivo.Size(), 10), 10) //se guarda el tamaño del archivo (para enviarlo)
	
	nombreArchivo := complementarCadena(datosArchivo.Name(), 64) // se guarda el nombre del archivo (para enviarlo)

	c.Write([]byte(tamanioArchivo))//Se escribe la informacion almacenada en las strings creadas previamente
	c.Write([]byte(nombreArchivo))

	lectorBufer := make([]byte, TAMANIOBUFFER)//Se debe crear un bufer, de donde el servidor leera por paquetes.

	for {
		_, err = auxArchivo.Read(lectorBufer) //se lee cada parte del archivo y es enviada
		if err == io.EOF {
			break
		}
		c.Write(lectorBufer)
	}

	fmt.Println("Se envio el archivo: " + archivo)

	c.Close()
	return
}

func complementarCadena(cadena string, longitud int) string {
	for {
		longitudCadena := len(cadena)
		if longitudCadena < longitud {
			cadena = cadena + ":"
			continue
		}
		break
	}
	return cadena
}

func envioSolicitudArchivo(archivo string) { //PUERTO :9994 PARA RECIBIR/ENVIAR LA SOLICITUD DE UN ARCHIVO
	c, err := net.Dial("tcp", ":9994")
	if err != nil {
		return
	}
	defer c.Close()
	err = gob.NewEncoder(c).Encode(archivo)
	if err != nil {
		return
	}
	recibirArchivoServidor()
}

func recibirArchivoServidor() {

	c, err := net.Dial("tcp", ":9993") //PUERTO :9993 PARA RECIBIR ARCHIVO DEL SERVIDOR
	if err != nil {
		return
	}
	defer c.Close()
	
	buferNombreArchivo := make([]byte, 64)//Se crean dos bufers para la creacion del archivo.
	buferTamanioArchivo := make([]byte, 10)

	c.Read(buferTamanioArchivo) //se lee el tamaño del archivo
	tamanioArchivo, _ := strconv.ParseInt(strings.Trim(string(buferTamanioArchivo), ":"), 10, 64) //se guarda el tamaño del archivo

	c.Read(buferNombreArchivo) //se lee el nombre del archivo
	nombreArchivo := strings.Trim(string(buferNombreArchivo), ":") // se guarda el nombre del archivo

	archivo, err := os.Create(nombreArchivo) //se crea un archivo 
	if err != nil {
		return
	}
	defer archivo.Close()

	var bytesLeidos int64

	for { //se lee toda la información del archivo y se guarda en el archivo creado
		if (tamanioArchivo - bytesLeidos) < TAMANIOBUFFER {
			io.CopyN(archivo, c, (tamanioArchivo - bytesLeidos))
			c.Read(make([]byte, (bytesLeidos+TAMANIOBUFFER)-tamanioArchivo))
			break
		}
		io.CopyN(archivo, c, TAMANIOBUFFER)
		bytesLeidos += TAMANIOBUFFER
	}

	fmt.Println("Archivo recibido: ", nombreArchivo)
}

func terminarCliente(nickname string) { //PUERTO :9997 PARA TERMINAR CONEXION DE UN CLIENTE
	c, err := net.Dial("tcp", ":9997")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gob.NewEncoder(c).Encode(nickname)
	if err != nil {
		fmt.Println(err)
	}
	c.Close()
}

func pausa() {
	var salto string
	fmt.Scanln(&salto)
	fmt.Println("Presionce ENTER para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func menu() {
	cliente = registroCliente()
	continuar := true
	var opcion int
	for continuar != false {
		fmt.Println("----------------------------------------")
		fmt.Println("NICKNAME: ", cliente.Nickname)
		fmt.Println("1) Enviar mensaje")
		fmt.Println("2) Enviar Archivo")
		fmt.Println("3) Recibir Archivo")
		fmt.Println("4) Salir")
		fmt.Print("Opción: \n")
		fmt.Scan(&opcion)
		switch opcion {
		case 1://opción para enviar un mensaje
			var salto string
			fmt.Scanln(&salto)
			fmt.Print("Ingrese el mensaje: ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				mensaje := scanner.Text()
				mensaje = cliente.Nickname + " dice: " + mensaje
				enviarMensaje(mensaje) //se envia el mensaje
			}
		case 2://funcion para enviar un archivo
			var archivo string = seleccionarArchivo() //se selecciona un archivo
			if len(archivo) > 0 {
				conexionServidorArchivos(archivo) // se hace la conexión y se envia
			}
		case 3://Opción para obtener archivos desde el servidor
			var archivoSeleccionado string
			fmt.Println("Archivos disponibles en el servidor")
			archivosS, err := ioutil.ReadDir("C:\\Users\\Usuario\\Documents\\lalo\\Distribuidos\\Examen1\\Servidor")
			if err != nil {
				log.Fatal(err)
			}
			for i, a := range archivosS {
				fmt.Println("(" + strconv.Itoa(i) + ") ->" + a.Name()) //muestro los archivos existentes
			}
			var opcionArchivo int
			fmt.Print("Seleccione el numero del archivo a recibir: ")
			fmt.Scan(&opcionArchivo)
			for i, a := range archivosS {
				if opcionArchivo == i {
					archivoSeleccionado = a.Name()
					fmt.Println("Selecciono el archivo: ", archivoSeleccionado)
				}
			}
			envioSolicitudArchivo(archivoSeleccionado)
			pausa() //COMENTAR O DESCOMENTAR
		case 4: //salir
			fmt.Println("Saliendo. . .")
			terminarCliente(cliente.Nickname)
			continuar = false
			pausa()
		default:
			fmt.Println("Opción no valida")
			pausa()
		}
		
	}
}

func main() {
	menu()
}
