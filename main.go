package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/emicklei/dot"
	"github.com/sqweek/dialog"
)

type Automata struct {
    States       []string          `json:"states"`
    Alphabet     []string          `json:"alphabet"`
    Transitions  map[string]map[string]string `json:"transitions"`
    InitialState string            `json:"initialState"`
    FinalStates  []string          `json:"finalStates"`
}

var automata Automata

// Verifica si el automata es completo
func (a *Automata) isComplete() bool {
    for _, state := range a.States {
        for _, symbol := range a.Alphabet {
            _, ok := a.Transitions[state][symbol]
            if !ok {
                return false
            }
        }
    }
    return true
}

// Completa el automata
func (a *Automata) complete() {
    sink := "sink"
    a.States = append(a.States, sink)
    a.Transitions[sink] = make(map[string]string)
    for _, state := range a.States {
        for _, symbol := range a.Alphabet {
            _, ok := a.Transitions[state][symbol]
            if !ok {
                a.Transitions[state][symbol] = sink
            }
        }
    }
}

// Verifica si la cadena es aceptada por el automata
func (a *Automata) accepts(input string) bool {
    currentState := a.InitialState
    for _, symbol := range input {
        nextState, ok := a.Transitions[currentState][string(symbol)]
        if !ok {
            return false
        }
        currentState = nextState
    }
    for _, finalState := range a.FinalStates {
        if currentState == finalState {
            return true
        }
    }
    return false
}

// Muestra un dialogo para seleccion de archivos
func showDialog(ext []string) (string, error) {
    // Obtiene el directorio actual
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    // Abre el dialogo de selección de archivo
    fileDialog := dialog.File().Title("Seleccionar archivo").Filter(ext[0], ext[1])
    filePath, err := fileDialog.SetStartDir(dir).Load()

    if err != nil {
        return "", err
    }

    // Retorna el path absoluto del archivo seleccionado
    return filepath.Abs(filePath)
}

// Permite cargar el archivo del automata
func upload() (error) {
    // Abre el dialogo de selección de archivo
    ext := []string{".json", "json"}

    if len(automata.States) > 0 {
        automata = Automata{}
    }

    filePath, err := showDialog(ext)
    if err != nil {
    return errors.New("Error al seleccionar archivo")
}

    file, err := os.Open(filePath)
    if err != nil {
        return errors.New("Error al leer archivo JSON")
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return errors.New("Error al leer archivo JSON")
    }

    err = json.Unmarshal(bytes, &automata)
    if err != nil {
        return errors.New("Error al leer archivo JSON")
    }
    return nil
}

// Permite cargar el archivo de la cadena
func uploadString() (string, error) {
    // Abre el dialogo de selección de archivo
    ext := []string{".txt", "txt"}
    var cadena string
    filePath, err := showDialog(ext)
    if err != nil {
    return cadena, errors.New("Error al seleccionar archivo")
    }

    file, err := os.Open(filePath)
    if err != nil {
        return cadena, errors.New("Error al leer archivo TXT")
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return cadena, errors.New("Error al leer archivo TXT")
    }

    cadena = string(bytes[:])
    cadena = strings.ToLower(cadena)
    return cadena, nil
}

// Convierte el automata en un String
func toString() string {
    automataS, _ := json.MarshalIndent(automata, "", " ")
    return string(automataS)
}

// Inicia los demas metodos
func start(input string) (bool, string) {
    
    // completar autómata si es necesario
    if !automata.isComplete() {
        automata.complete()
    }

    // probar si acepta la cadena de entrada
    if automata.accepts(input) {
        return true, input

    } else {
        return false, input
    }
}

// Crea el grafo con los valores del automata
func createGraph() (*dot.Graph, error) {
    g := dot.NewGraph(dot.Directed)
    a := automata

    // Generar los nodos
    for _, state := range a.States {
		node := g.Node(state)
		if state == a.InitialState {
            node.Attr("style", "filled")
			node.Attr("fillcolor", "green")
		}
		for _, finalState := range a.FinalStates {
			if state == finalState {
                node.Attr("style", "filled")
				node.Attr("fillcolor", "red")
			}
		}
	}

    // Generar las aristas
    for source, transitions := range a.Transitions {
		for symbol, target := range transitions {
			g.Edge(g.Node(source), g.Node(target)).Attr("label", symbol)

		}
	}

    return g, nil
}

// Funcion principal
func main() {    

    var app = app.New()

    window := app.NewWindow("Automatons by ACME")

    icon, err := fyne.LoadResourceFromPath("./Resources/upload.png")
    window.SetIcon(icon)

    window.Resize(fyne.NewSize(400, 400))

    importIcon, err := fyne.LoadResourceFromPath("./Resources/upload.png")
    if err != nil {
        fmt.Println("Error al leer archivo SVG:", err)
    }

    importTxt := widget.NewLabel("Sube el automata en formato JSON")

    

    startIcon, err := fyne.LoadResourceFromPath("./Resources/start.png")
    if err != nil {
        fmt.Println("Error al leer archivo SVG:", err)
        return
    }

    showIcon, err := fyne.LoadResourceFromPath("./Resources/show.png")
    if err != nil {
        fmt.Println("Error al leer archivo SVG:", err)
        return
    }

    entryTxt := widget.NewLabel("Ingresa la cadena a verificar: ")
    entry := widget.NewEntry()

    automatonTxt := widget.NewLabel("")
    
    importBtn := widget.NewButtonWithIcon("CARGAR", importIcon, func() {
        if upload() != nil{
            entry.SetText("ERROR al cargar el archivo")
        }
    })

    importStringBtn := widget.NewButtonWithIcon("CARGAR CADENA", importIcon, func() {
        cadena, err := uploadString()
        if err != nil{
            entry.SetText("ERROR al cargar el archivo")
        }
        entry.SetText(cadena)
    })


    start := widget.NewButtonWithIcon("VERIFICAR CADENA", startIcon, func() {
        cadena := strings.ToLower(entry.Text)
        status, input := start(cadena)
        if status {
            txt := fmt.Sprintf("La cadena:\n %v\n\n>> ES VALIDA <<", input)
            automatonTxt.SetText(txt)
            return
        } 
        txt := fmt.Sprintf("La cadena:\n %v\n\n>> NO ES VALIDA <<", input)
        automatonTxt.SetText(txt)
    })

    

    show := widget.NewButtonWithIcon("VER AUTOMATA", showIcon, func() {
        g, err := createGraph()
        if err != nil {
            panic(err)
        }
        popup := app.NewWindow("Grafo")

        // Eliminar el archivo grafo.png si existe
        if _, err := os.Stat("grafo.png"); err == nil {
            if err := os.Remove("grafo.png"); err != nil {
                panic(err)
            }
        }

        // Generar una imagen PNG del grafo utilizando Graphviz
        cmd := exec.Command("dot", "-Tpng")
        var out bytes.Buffer
        cmd.Stdin = io.TeeReader(bytes.NewBufferString(g.String()), &out)
        outfile, err := os.Create("grafo.png")
        if err != nil {
            panic(err)
        }
        defer outfile.Close()
        cmd.Stdout = outfile
        if err := cmd.Run(); err != nil {
            panic(err)
        }

        // Obtener el tamaño de la imagen generada
        imgFile, err := os.Open("grafo.png")
        if err != nil {
            panic(err)
        }
        defer imgFile.Close()
        imgCfg, _, err := image.DecodeConfig(imgFile)
        if err != nil {
            panic(err)
        }
        imgWidth := imgCfg.Width
        imgHeight := imgCfg.Height

        popup.SetContent(canvas.NewImageFromFile("grafo.png"))
        popup.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
        popup.Show()
        })

    content := container.NewVBox(importTxt, importBtn, entryTxt, importStringBtn, entry, start, show, automatonTxt)
    window.SetContent(content)
    window.ShowAndRun()
}