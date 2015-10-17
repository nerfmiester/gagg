package main

import(

	"fmt"
	"github.com/nerfmiester/gagg/Config"
	"log"

"github.com/ivpusic/toml"
)
var tomlConfig Config.TomlConfig

func getToml() {
	if _, err := toml.DecodeFile("definition.toml", &tomlConfig); err != nil {
		log.Fatal(err)
		// return
	}
	if (tomlConfig.Debug.Enabled) {
		fmt.Printf("Title: %s\n", tomlConfig.Title)
		fmt.Printf("Owner: %s (%s, %s), Born: %s\n",
			tomlConfig.Owner.Name, tomlConfig.Owner.Org, tomlConfig.Owner.Bio,
			tomlConfig.Owner.DOB)
		fmt.Printf("Database: %s %v (Max conn. %d), Enabled? %v\n",
			tomlConfig.DB.Server, tomlConfig.DB.Ports, tomlConfig.DB.ConnMax,
			tomlConfig.DB.ConnMax, tomlConfig.DB.Server)
		for serverName, server := range tomlConfig.Servers {
			fmt.Printf("Server: %s (%s, %s)\n", serverName, server.IP, server.DC)
		}
		fmt.Printf("Client data: %v\n", tomlConfig.Clients.Data)
		fmt.Printf("Client hosts: %v\n", tomlConfig.Clients.Hosts)
		fmt.Printf("Agg: %s, ", tomlConfig.Agg.Server)
	}
}
func main() {
	fmt.Println("Hello")

	for i, port := range tomlConfig.Agg.Ports {

		fmt.Printf("Port %d is %d",i,port)

	}

}