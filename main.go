package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fyne.io/fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
func showDialog() (string, error) {
    // Obtiene el directorio actual
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    // Abre el dialogo de selección de archivo
    fileDialog := dialog.File().Title("Seleccionar archivo").Filter("JSON", "json")
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
    filePath, err := showDialog()
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

// Funcion principal
func main() {    

    var app = app.New()

    window := app.NewWindow("Automatons by ACME")

    icon, err := fyne.LoadResourceFromPath("./Resources/upload.png")
    window.SetIcon(icon)

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


    start := widget.NewButtonWithIcon("VERIFICAR CADENA", startIcon, func() {
        status, input := start(entry.Text)
        if status {
            txt := fmt.Sprintf("La cadena %v es VALIDA", input)
            automatonTxt.SetText(txt)
            return
        } 
        txt := fmt.Sprintf("La cadena %v NO es VALIDA", input)
        automatonTxt.SetText(txt)
    })

    show := widget.NewButtonWithIcon("VER AUTOMATA", showIcon, func() {
        automatonTxt.SetText(toString())
    })


    content := container.NewVBox(importTxt, importBtn, entryTxt, entry, start, show, automatonTxt)

    window.SetContent(content)
    window.ShowAndRun()
}