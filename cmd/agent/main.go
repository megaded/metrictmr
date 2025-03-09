package main

import (
	"net/http"

	"github.com/megaded/metrictmr/internal/agent"
)

func main() {
	a := agent.CreateAgent()
	a.StartSend()
	http.HandleFunc()
	f := func(http.ResponseWriter, *http.Request){}
	http.Handle("1", f)


	


		
	
}
