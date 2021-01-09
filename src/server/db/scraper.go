package main

import (
    "fmt"
    "sync"
    "strconv"
    "strings"
    "io/ioutil"
    "encoding/json"
    "github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
)


type Pokemon struct {
    Id           int       `json:"id"`
    Name         string    `json:"name"`
    Category     string    `json:"category"`

    Generation   int       `json:"generation"`
    Descriptions []string  `json:"descriptions"`
    Types        []string  `json:"types"`

    Precedent    string    `json:"precedent"`
    Stage        string    `json:"stage"`
}


var Pokedex struct {
    Pokemons     map[int]Pokemon
    mu           sync.Mutex
}

func GetPokemon(id int) (Pokemon, bool) {
    Pokedex.mu.Lock()
    defer Pokedex.mu.Unlock()

    pokemon, found := Pokedex.Pokemons[id]
    return pokemon, found
}

func SavePokemon(id int, pokemon Pokemon) {
    Pokedex.mu.Lock()
    Pokedex.Pokemons[id] = pokemon
    Pokedex.mu.Unlock()
}


func DownloadSpecies(jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for id := range jobs {
        var pokemon Pokemon
        var data structs.PokemonSpecies
        err := fmt.Errorf("")
    
        // Asks the APIs
        for err != nil {
            data, err = pokeapi.PokemonSpecies(strconv.Itoa(id))
        }

        // Sets pokemon's id, name and generation
        pokemon.Id = id
        pokemon.Name = strings.Title(data.Name)
        pokemon.Generation, _ = strconv.Atoi(string(data.Generation.URL[37]))

        // Sets its precedent, if exists
        if precedent, ok := data.EvolvesFromSpecies.(map[string]interface{}); ok {
            pokemon.Precedent = strings.Title(precedent["name"].(string))
        }

        // Set its category
        for _, category := range data.Genera {
            if category.Language.Name == "en" {
                pokemon.Category = category.Genus
                break
            }
        }

        // Checks for descriptions
        for _, desc := range data.FlavorTextEntries {
            if desc.Language.Name != "en" {
                continue
            }

            // Removes \n and \f and converts it to lowercase
            text := strings.ToLower(desc.FlavorText)
            text = strings.ReplaceAll(text, "\n", " ")
            text = strings.ReplaceAll(text, "\f", " ")

            // Excludes the description if it contains the pokemon's name
            if strings.Contains(text, strings.ToLower(pokemon.Name)) {
                continue
            }

            // Excludes the description if it's duplicated
            found := false
            for _, description := range pokemon.Descriptions {
                if text == description {
                    found = true
                    continue
                }
            }
                
            if !found {
                pokemon.Descriptions = append(pokemon.Descriptions, text)
            }
        }
    
        if len(pokemon.Descriptions) == 0 {
            pokemon.Descriptions = []string{}
        }

        SavePokemon(id, pokemon)
    }
}

func DownloadPokemon(jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for id := range jobs {
        // Gets the saved data
        pokemon, _ := GetPokemon(id)
    
        // Asks the APIs
        var data structs.Pokemon
        err := fmt.Errorf("")
        for err != nil {
            data, err = pokeapi.Pokemon(strconv.Itoa(id))
        }

        // Sets pokemon's types
        for _, slot := range data.Types {
            pokemon.Types = append(pokemon.Types, strings.Title(slot.Type.Name))
        }
        
        SavePokemon(id, pokemon)
    }
}

func DownloadEvolutions(jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for id := range jobs {
        // Skips the chains that doesn't exist
		blacklist := map[int]int {210: 0, 222:0, 225:0, 226: 0, 227: 0, 231: 0, 238: 0, 251: 0}
		if _, found := blacklist[id]; found {
            continue
        }

        // Asks the APIs
        var data structs.EvolutionChain
        err := fmt.Errorf("")
        for err != nil {
            data, err = pokeapi.EvolutionChain(strconv.Itoa(id))
        }
        
        // Prepares the chain 
        var chain [3][]int
        
        // Adds the first level
        pid, _ := strconv.Atoi(strings.ReplaceAll(data.Chain.Species.URL, "/", "")[36:])
        if _, found := GetPokemon(pid); found {
            chain[0] = []int{pid}
        } else {
            continue
        }

        // Adds all the second levels
        for _, second := range data.Chain.EvolvesTo {
            pid, _ = strconv.Atoi(strings.ReplaceAll(second.Species.URL, "/", "")[36:])

            if _, found := GetPokemon(pid); found {
                chain[1] = append(chain[1], pid)
            } else {
                continue
            }

            // Adds all the third levels
            for _, third := range second.EvolvesTo {
                pid, _ = strconv.Atoi(strings.ReplaceAll(third.Species.URL, "/", "")[36:])

                if _, found := GetPokemon(pid); found {
                    chain[2] = append(chain[2], pid)
                } else {
                    continue
                }
            }
        }

        // Calculates the maximum
        var max int
        if len(chain[2]) > 0 {
            max = 3
        } else if len(chain[1]) > 0 {
            max = 2
        } else {
            max = 1
        }

        // Sets the stage for all the pokemons in the chain
        for level, pids := range chain {
            stage := strconv.Itoa(level + 1) + "/" + strconv.Itoa(max)
    
            // And saves them
            for _, pid := range pids {
                pokemon, _ := GetPokemon(pid)
                pokemon.Stage = stage
                SavePokemon(pid, pokemon)
            }
        }
    }
}

func FeedWorkers(worker func(<-chan int, *sync.WaitGroup), times int) {
    // Creates a pair of channels
    var wg sync.WaitGroup
    jobs := make(chan int, times)

    // Starts some workers
    for wid := 0; wid < 250; wid++ {
        wg.Add(1)
        go worker(jobs, &wg)
    }

    // Feeds them
    for id := 1; id <= times; id++ {
        jobs <- id
    }

    // Waits until they've finished
    close(jobs)
    wg.Wait()
}


func main() {
    Pokedex.Pokemons = make(map[int]Pokemon)
    fmt.Println("Starting pokémon scraper...")

    // Feeds the DownloadSpecies workers
    fmt.Print("- Downloading part 1/3... ")
    FeedWorkers(DownloadSpecies, 768)
    fmt.Println("done.")

    // Feeds the DownloadPokemons workers
    fmt.Print("- Downloading part 2/3... ")
    FeedWorkers(DownloadPokemon, 768)
    fmt.Println("done.")

    // Feeds the DownloadEvolutions workers
    fmt.Print("- Downloading part 3/3... ")
    FeedWorkers(DownloadEvolutions, 422)
    fmt.Println("done.")

    // Writes the output in a json file
    fmt.Print("- Writing to file... ")
    file, _ := json.MarshalIndent(Pokedex.Pokemons, "", "    ")
    ioutil.WriteFile("pokemons.json", file, 0644)
    fmt.Println("done.")

    fmt.Println("Scraping finished.")
}
