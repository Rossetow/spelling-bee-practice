package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"runtime"
	"context"
	"time"

	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	"github.com/hegedustibor/htgo-tts/voices"
	"github.com/joho/godotenv"
	"github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type Word struct {
	Word string 	`json:"word"`
	Used bool 		`json:"used"`
}

type Game struct {
	Words []Word	`json:"words"`
	usedWords []Word
	unusedWords []Word
}

func newGame () *Game{
	var words Game

	words.getWords()
	words.filterUsedWords()

	return &words
}

const (
    ENV_FILE = "../.env"
)

func loadEnvFile() error {

	//Get env when running locally
    if _, err := os.Stat(ENV_FILE); os.IsNotExist(err) {
        return nil
    }
	
    return godotenv.Load(ENV_FILE)
}

func getResponseFromAI(word string) string{

	// Prompt and get response from GeminiAI
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
    if err != nil {
        panic(err)
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash")
    resp, err := model.GenerateContent(ctx, genai.Text("give me a simples sentence using the word " + word + " in the context to explain the words' meaning"))
    if err != nil {
        panic(err)
    }
	return fmt.Sprint(resp.Candidates[0].Content.Parts[0])
}

func main () {

	// Load env variables
	loadEnvFile()

	// Initialize variables
	choice := -1
	g := newGame()

	

	for choice != 0 {

		// Verify if all words have already been used and if so, reset them
		if len(g.Words) == len(g.usedWords) {
			fmt.Println("Você já praticou todas as palavras! Vou resetar e começar do início!")
			fmt.Println("Aperte qualquer tecla para continuar")
	
			dummy := ""
	
			fmt.Scan(&dummy)
	
			g.resetWords()
	
			ClearScreen()
		}

		// Prompts
		fmt.Println("O que você quer fazer?")
		fmt.Println("1 - sortear palavra")
		fmt.Println("2 - ver as palavras restantes")
		fmt.Println("3 - ver as palavras usadas")
		fmt.Println("4 - ver todas as palavras")
		fmt.Println("5 - reiniciar palavras")
		fmt.Println("0 - sair")

		fmt.Scan(&choice)

		switch choice{
		case 1:
			g.spellingBee()

		case 2:
			g.listUnusedWords()

		case 3:
			g.listUsedWords()

		case 4:
			g.listAllWords()
		
		case 0:
			break

		default:
			fmt.Println("Escolha inválida!")
			time.Sleep(2500)
			continue
		}

		fmt.Println("\ncontinuar?")
		fmt.Println("1 - sim")
		fmt.Println("0 - não")

		fmt.Scan(&choice)

		if choice != 0 && choice != 1 {
			fmt.Println("Escolha inválida!")
			time.Sleep(2000)

			continue
		}

		// Clear terminal
		ClearScreen()

	}

	// Terminate the program
	fmt.Println("Ok, até mais!")
	time.Sleep(1000)
	fmt.Println("Salvando progresso das palavras...")
	os.RemoveAll("./audio")
	g.saveProgress()
	time.Sleep(3000)
}

func ClearScreen() {
    if runtime.GOOS == "windows" {
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    } else {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
}

func (g *Game) getWords(){

	// Read json file
	byteValue, err := os.ReadFile("../data/words.json")

	if err != nil {
		panic("Error reading used words file")
	}

	// Convert json bytes to structs
	json.Unmarshal(byteValue, g)

	var ret []Word

	// Populate the client
	g.Words = append(ret, g.Words...)

}

func  (g *Game) filterUsedWords(){

	g.unusedWords = []Word{}
	g.usedWords = []Word{}

	for _, nw := range g.Words {

		// Separete words in order to not repeating them
		if nw.Used {
			g.usedWords = append(g.usedWords, nw)
			continue
		}

		g.unusedWords = append(g.unusedWords, nw)
		
	}
}

func (g *Game) changeWordStatus(word Word) {

	// Set Word.Used to true
	for i, ws := range g.Words {
		if ws == word {
			g.Words[i].Used = true
			break
		}
	}

	g.filterUsedWords()
}

func (g *Game) saveProgress() {

	// Convert structs into json
	jsonFormat, err := json.MarshalIndent(g, "", "  ")

	if err != nil {
		panic("Error marshalling words")
	}

	// Overwrite the file with updated data
	err = os.WriteFile("../data/words.json", jsonFormat, 0644)

	if err != nil {
		panic("Error saving words")
	}
}

func (g *Game) spellingBee () {

	choice := -1

	wordIndex := rand.IntN(len(g.unusedWords))

	// Run text to speech lib
	tts(g.unusedWords[wordIndex].Word)

	fmt.Println("1 - ouvir a palavra novamente")
	fmt.Println("2 - ouvir a palavra em um contexto")
	fmt.Println("3 - escrever a palavra")

	fmt.Scan(&choice)

	if choice != 2 && choice != 1  && choice != 3{
		fmt.Println("Escolha inválida!")
		time.Sleep(1000)
	}

	if choice == 1 {

		tts(g.unusedWords[wordIndex].Word)
		fmt.Println("1 - ouvir a palavra em um contexto")
		fmt.Println("2 - escrever a palavra")

		fmt.Scan(&choice)

		if choice == 1 {
			tts(getResponseFromAI(g.unusedWords[wordIndex].Word))
		}
		choice = -1
	}

	if choice == 2 {
		tts(getResponseFromAI(g.unusedWords[wordIndex].Word))
		fmt.Println("1 - ouvir a palavra novamente")
		fmt.Println("2 - escrever a palavra")

		fmt.Scan(&choice)

		if choice == 1 {
			tts(g.unusedWords[wordIndex].Word)
		}
	}

	fmt.Println("Escreva a palavra: ")

	var writtenWord string

	fmt.Scan(&writtenWord)

	// Check if answer is the same as expected
	if writtenWord != g.unusedWords[wordIndex].Word {
		fmt.Printf("\nErrado! A palavra era: %v\n", g.unusedWords[wordIndex].Word)
		fmt.Printf("Você escreveu: %v\n", writtenWord)
	} else {
		fmt.Printf("\nParabéns, você acertou!! A palavra era: %v\n", g.unusedWords[wordIndex].Word)
		g.changeWordStatus(g.unusedWords[wordIndex])
	}
}

func tts (word string) {

	// Tts algorythm
	speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.Native{}}
	err := speech.Speak(word); 
	if err != nil {
	panic(err)
	}
}

func (g *Game) listUnusedWords (){

	for i, uw := range g.unusedWords {
		fmt.Printf("\n%v: %v", (i + 1), uw.Word)
	}
	fmt.Println("\n")
}

func (g *Game) listUsedWords (){

	for i, uw := range g.usedWords {
		fmt.Printf("\n%v: %v", (i + 1), uw.Word)
	}
	fmt.Println("\n")
}

func (g *Game) listAllWords (){

	for i, uw := range g.Words {

		var used string

		if uw.Used {
			used = "Usada"
		} else {
			used = "Não usada"
		}

		fmt.Printf("\n%v: %v - %v", (i + 1), uw.Word, used)
	}
	fmt.Println("\n")
}

func (g *Game) resetWords () {

	// Set all words to not used again
	for _, w := range g.Words {
		w.Used = false
	}

	g.filterUsedWords()
	g.saveProgress()
}