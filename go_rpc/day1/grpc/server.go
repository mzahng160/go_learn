package geerpc

import (
	"encoding/json"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int        //MagicNumber marks this's a geerpc request
	CodeType    codec.Type //client may choose different Codec encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodeType:    codec.GobType,
}

type Server struct{}

// NewServer return a new server
func NewServer() *Server {
	return &Server{}
}

//DefaultServer is the default instance of *Server
var DefaultServer = NewServer()

//Server runs the server on a single connection
//Server blocks, serving the connection until the client hangs up
func (server *Server) ServerConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option

	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error:", err)
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %dx", opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodeType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodeType)
		return
	}

	server.serverCodec(f(conn))
}

//invalidRequest is a placeholder for reponse argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serverCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup) //wait until all request are handle

	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}

			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}

		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}

	wg.Wait()
	_ = cc.Close()
}

//request stores all information of a call
type request struct {
	h           *codec.Header //header of request
	argv, reply reflect.Value //argv and reply of request
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}

	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}

	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.reply = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.reply.Interface(), sending)
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}

		go server.ServerConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
