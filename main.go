package main


import(
"jogodecartasonline/uteis"
)



func main() {
	var server = uteis.Server{}
	server = *uteis.NewServer()
	
	go server.PrintStats()
	// Inicia o servidor na porta 8080
	server.Start("8080")
}
