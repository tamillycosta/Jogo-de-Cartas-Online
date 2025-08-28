package api

import (

	
	"jogodecartasonline/api/Request"
	"jogodecartasonline/api/Response"
	"log"
	"net"
	
)

type HandlerMethod func(req request.Request, conn net.Conn) response.Response

type Aplication struct {
    Routes map[string]HandlerMethod

}

func NewAplication() *Aplication{
	return &Aplication{Routes: make(map[string]HandlerMethod)}
}


func (api *Aplication) AddRoute(methodName string, method HandlerMethod){
	
	if _, ok := api.Routes[methodName]; ok{
		api.Forbidden("Este Metodo Ja Existe")
	}else{
		api.Routes[methodName] = method
	}
	
}


func (api *Aplication) Dispatch(req request.Request, conn net.Conn) response.Response {
    if handler, ok := api.Routes[req.Method]; ok {
        return handler(req, conn)
    }
    var resp response.Response
    return resp.MakeErrorResponse(404, "Este Metodo Não Existe", "Not Found")
}


func (api *Aplication) InternalServerError(conn net.Conn, err error) {
    resp := response.Response{}
	resp.MakeErrorResponse(500, "Houve um Erro no Servidor", "InternalServerError")
    data, _ := resp.Serialize()
    conn.Write(data)
    log.Printf("Erro interno: %v\n", err)
    conn.Close()
}


func (api *Aplication) Forbidden(forbiddenAction string) response.Response{
	resp := response.Response{}
    return resp.MakeErrorResponse( 403, forbiddenAction, "InternalServerError")
}