package main

import (
    "encoding/json"
    "fmt"
    "os"
)

func main() {
    commands := []string{"init", "add", "edit", "remove", "list", "find", "show", "gen", "backup", "restore"}
    if len(os.Args) > 1 && os.Args[1] == "--json" { _ = json.NewEncoder(os.Stdout).Encode(map[string]any{"commands": commands}); return }
    if len(os.Args) > 1 && os.Args[1] == "--version" { fmt.Println("passwd-cli 1.0.0"); return }
    fmt.Println("passwd-cli: local encrypted vault")
    fmt.Println("commands:", commands)
}
