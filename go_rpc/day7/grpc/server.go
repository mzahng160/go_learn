package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int        //MagicNumber marks this's a geerpc request
	CodecType      codec.Type //client may choose different Codec encode body
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

type Server struct {
	serviceMap sync.Map
}

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
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}

	server.serverCodec(f(conn), &opt)
}

//invalidRequest is a placeholder for reponse argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serverCodec(cc codec.Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup) //wait until all request are handle

	for {
		fmt.Print("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$\n")

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
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
	}

	wg.Wait()
	_ = cc.Close()
}

//request stores all information of a call
type request struct {
	h              *codec.Header //header of request
	argv, replyval reflect.Value //argv and reply of request
	mtype          *methodType
	svc            *service
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

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {

	fmt.Printf("serviceMethod %s\n", serviceMethod)

	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed:" + serviceMethod)
		return
	}

	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: cannot find service: " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: cannot find method: " + methodName)
	}
	return
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}

	req.argv = req.mtype.newArgv()
	req.replyval = req.mtype.newReplyv()

	fmt.Println("###################req.argv: ", req.argv, "|req.argv:", req.replyval)

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	fmt.Println("###################argvi: ", argvi)

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}

	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {

	fmt.Print("###########Server sendResponse\n")

	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex,
	wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	fmt.Print("###########Server sendResponse\n")

	called := make(chan struct{})
	sent := make(chan struct{})

	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyval)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}

		server.sendResponse(cc, req.h, req.replyval.Interface(), sending)
		sent <- struct{}{}
	}()

	fmt.Printf("###########Server timeout: %v\n", timeout)

	//timeout = 1

	if timeout == 0 {
		fmt.Print("################## timeout 0$$$$$$$$$\n")

		<-called
		<-sent

		fmt.Print("################## timeout 0*********\n")

		return
	}

	select {
	case <-time.After(timeout):
		fmt.Printf("##################  handleRequest select timeout\n")
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout:expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		fmt.Printf("################## handleRequest select called\n")
		<-sent
	}

	fmt.Print("##################Server handleRequest finish!\n")
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

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	fmt.Println("Server Register name ", s.name)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined:" + s.name)
	}
	return nil
}

//register publishes the receiver's methods in the defaultserver
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

const (
	connected        = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geerpc_"
	defaultDebugPath = "/debug/geerpc"
)

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	fmt.Println("ServeHTTP req.Method:", req.Method)

	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ":", err.Error())
		return
	}

	fmt.Println("ServeHTTP connected:", connected)

	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServerConn(conn)
}

func (server *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
