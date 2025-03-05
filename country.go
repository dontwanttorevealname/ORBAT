package main

import (
    "bufio"
    "fmt"
    "github.com/biter777/countries"
    "os"
    "strings"
)

func main() {
    allCountries := countries.All()
    for _, countryCode := range allCountries {
        country := countryCode.Info()
        fmt.Printf("Country: %-40s Alpha-3: %s\n", country.Name, country.Alpha3)
    }
    
    fmt.Print("Enter country name: ")
    reader := bufio.NewReader(os.Stdin)
    input, _ := reader.ReadString('\n')
    input = strings.TrimSpace(input)
    
    country := countries.ByName(input).Info()
    fmt.Printf("Country: %s\n", country.Name)
    fmt.Printf("Alpha-2: %s\n", country.Alpha2)
}