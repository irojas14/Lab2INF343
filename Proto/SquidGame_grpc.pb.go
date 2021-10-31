// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package Proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LiderClient is the client API for Lider service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LiderClient interface {
	Unirse(ctx context.Context, in *SolicitudUnirse, opts ...grpc.CallOption) (Lider_UnirseClient, error)
	VerMonto(ctx context.Context, in *SolicitudVerMonto, opts ...grpc.CallOption) (*RespuestaVerMonto, error)
	EnviarJugada(ctx context.Context, in *SolicitudEnviarJugada, opts ...grpc.CallOption) (*RespuestaEnviarJugada, error)
}

type liderClient struct {
	cc grpc.ClientConnInterface
}

func NewLiderClient(cc grpc.ClientConnInterface) LiderClient {
	return &liderClient{cc}
}

func (c *liderClient) Unirse(ctx context.Context, in *SolicitudUnirse, opts ...grpc.CallOption) (Lider_UnirseClient, error) {
	stream, err := c.cc.NewStream(ctx, &Lider_ServiceDesc.Streams[0], "/Proto.Lider/Unirse", opts...)
	if err != nil {
		return nil, err
	}
	x := &liderUnirseClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Lider_UnirseClient interface {
	Recv() (*RespuestaUnirse, error)
	grpc.ClientStream
}

type liderUnirseClient struct {
	grpc.ClientStream
}

func (x *liderUnirseClient) Recv() (*RespuestaUnirse, error) {
	m := new(RespuestaUnirse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *liderClient) VerMonto(ctx context.Context, in *SolicitudVerMonto, opts ...grpc.CallOption) (*RespuestaVerMonto, error) {
	out := new(RespuestaVerMonto)
	err := c.cc.Invoke(ctx, "/Proto.Lider/VerMonto", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liderClient) EnviarJugada(ctx context.Context, in *SolicitudEnviarJugada, opts ...grpc.CallOption) (*RespuestaEnviarJugada, error) {
	out := new(RespuestaEnviarJugada)
	err := c.cc.Invoke(ctx, "/Proto.Lider/EnviarJugada", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LiderServer is the server API for Lider service.
// All implementations must embed UnimplementedLiderServer
// for forward compatibility
type LiderServer interface {
	Unirse(*SolicitudUnirse, Lider_UnirseServer) error
	VerMonto(context.Context, *SolicitudVerMonto) (*RespuestaVerMonto, error)
	EnviarJugada(context.Context, *SolicitudEnviarJugada) (*RespuestaEnviarJugada, error)
	mustEmbedUnimplementedLiderServer()
}

// UnimplementedLiderServer must be embedded to have forward compatible implementations.
type UnimplementedLiderServer struct {
}

func (UnimplementedLiderServer) Unirse(*SolicitudUnirse, Lider_UnirseServer) error {
	return status.Errorf(codes.Unimplemented, "method Unirse not implemented")
}
func (UnimplementedLiderServer) VerMonto(context.Context, *SolicitudVerMonto) (*RespuestaVerMonto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerMonto not implemented")
}
func (UnimplementedLiderServer) EnviarJugada(context.Context, *SolicitudEnviarJugada) (*RespuestaEnviarJugada, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnviarJugada not implemented")
}
func (UnimplementedLiderServer) mustEmbedUnimplementedLiderServer() {}

// UnsafeLiderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LiderServer will
// result in compilation errors.
type UnsafeLiderServer interface {
	mustEmbedUnimplementedLiderServer()
}

func RegisterLiderServer(s grpc.ServiceRegistrar, srv LiderServer) {
	s.RegisterService(&Lider_ServiceDesc, srv)
}

func _Lider_Unirse_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SolicitudUnirse)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LiderServer).Unirse(m, &liderUnirseServer{stream})
}

type Lider_UnirseServer interface {
	Send(*RespuestaUnirse) error
	grpc.ServerStream
}

type liderUnirseServer struct {
	grpc.ServerStream
}

func (x *liderUnirseServer) Send(m *RespuestaUnirse) error {
	return x.ServerStream.SendMsg(m)
}

func _Lider_VerMonto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudVerMonto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiderServer).VerMonto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.Lider/VerMonto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiderServer).VerMonto(ctx, req.(*SolicitudVerMonto))
	}
	return interceptor(ctx, in, info, handler)
}

func _Lider_EnviarJugada_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudEnviarJugada)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiderServer).EnviarJugada(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.Lider/EnviarJugada",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiderServer).EnviarJugada(ctx, req.(*SolicitudEnviarJugada))
	}
	return interceptor(ctx, in, info, handler)
}

// Lider_ServiceDesc is the grpc.ServiceDesc for Lider service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Lider_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Proto.Lider",
	HandlerType: (*LiderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VerMonto",
			Handler:    _Lider_VerMonto_Handler,
		},
		{
			MethodName: "EnviarJugada",
			Handler:    _Lider_EnviarJugada_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Unirse",
			Handler:       _Lider_Unirse_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "Proto/SquidGame.proto",
}

// PozoClient is the client API for Pozo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PozoClient interface {
	VerMonto(ctx context.Context, in *SolicitudVerMonto, opts ...grpc.CallOption) (*RespuestaVerMonto, error)
}

type pozoClient struct {
	cc grpc.ClientConnInterface
}

func NewPozoClient(cc grpc.ClientConnInterface) PozoClient {
	return &pozoClient{cc}
}

func (c *pozoClient) VerMonto(ctx context.Context, in *SolicitudVerMonto, opts ...grpc.CallOption) (*RespuestaVerMonto, error) {
	out := new(RespuestaVerMonto)
	err := c.cc.Invoke(ctx, "/Proto.Pozo/VerMonto", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PozoServer is the server API for Pozo service.
// All implementations must embed UnimplementedPozoServer
// for forward compatibility
type PozoServer interface {
	VerMonto(context.Context, *SolicitudVerMonto) (*RespuestaVerMonto, error)
	mustEmbedUnimplementedPozoServer()
}

// UnimplementedPozoServer must be embedded to have forward compatible implementations.
type UnimplementedPozoServer struct {
}

func (UnimplementedPozoServer) VerMonto(context.Context, *SolicitudVerMonto) (*RespuestaVerMonto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerMonto not implemented")
}
func (UnimplementedPozoServer) mustEmbedUnimplementedPozoServer() {}

// UnsafePozoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PozoServer will
// result in compilation errors.
type UnsafePozoServer interface {
	mustEmbedUnimplementedPozoServer()
}

func RegisterPozoServer(s grpc.ServiceRegistrar, srv PozoServer) {
	s.RegisterService(&Pozo_ServiceDesc, srv)
}

func _Pozo_VerMonto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudVerMonto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PozoServer).VerMonto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.Pozo/VerMonto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PozoServer).VerMonto(ctx, req.(*SolicitudVerMonto))
	}
	return interceptor(ctx, in, info, handler)
}

// Pozo_ServiceDesc is the grpc.ServiceDesc for Pozo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Pozo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Proto.Pozo",
	HandlerType: (*PozoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VerMonto",
			Handler:    _Pozo_VerMonto_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Proto/SquidGame.proto",
}

// NameNodeClient is the client API for NameNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NameNodeClient interface {
	RegistrarJugadas(ctx context.Context, in *SolicitudRegistrarJugadas, opts ...grpc.CallOption) (*RespuestaRegistrarJugadas, error)
	DevolverJugadas(ctx context.Context, in *SolicitudDevolverRegistro, opts ...grpc.CallOption) (*RespuestaDevolverRegistro, error)
}

type nameNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewNameNodeClient(cc grpc.ClientConnInterface) NameNodeClient {
	return &nameNodeClient{cc}
}

func (c *nameNodeClient) RegistrarJugadas(ctx context.Context, in *SolicitudRegistrarJugadas, opts ...grpc.CallOption) (*RespuestaRegistrarJugadas, error) {
	out := new(RespuestaRegistrarJugadas)
	err := c.cc.Invoke(ctx, "/Proto.NameNode/RegistrarJugadas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nameNodeClient) DevolverJugadas(ctx context.Context, in *SolicitudDevolverRegistro, opts ...grpc.CallOption) (*RespuestaDevolverRegistro, error) {
	out := new(RespuestaDevolverRegistro)
	err := c.cc.Invoke(ctx, "/Proto.NameNode/DevolverJugadas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NameNodeServer is the server API for NameNode service.
// All implementations must embed UnimplementedNameNodeServer
// for forward compatibility
type NameNodeServer interface {
	RegistrarJugadas(context.Context, *SolicitudRegistrarJugadas) (*RespuestaRegistrarJugadas, error)
	DevolverJugadas(context.Context, *SolicitudDevolverRegistro) (*RespuestaDevolverRegistro, error)
	mustEmbedUnimplementedNameNodeServer()
}

// UnimplementedNameNodeServer must be embedded to have forward compatible implementations.
type UnimplementedNameNodeServer struct {
}

func (UnimplementedNameNodeServer) RegistrarJugadas(context.Context, *SolicitudRegistrarJugadas) (*RespuestaRegistrarJugadas, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegistrarJugadas not implemented")
}
func (UnimplementedNameNodeServer) DevolverJugadas(context.Context, *SolicitudDevolverRegistro) (*RespuestaDevolverRegistro, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DevolverJugadas not implemented")
}
func (UnimplementedNameNodeServer) mustEmbedUnimplementedNameNodeServer() {}

// UnsafeNameNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NameNodeServer will
// result in compilation errors.
type UnsafeNameNodeServer interface {
	mustEmbedUnimplementedNameNodeServer()
}

func RegisterNameNodeServer(s grpc.ServiceRegistrar, srv NameNodeServer) {
	s.RegisterService(&NameNode_ServiceDesc, srv)
}

func _NameNode_RegistrarJugadas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudRegistrarJugadas)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NameNodeServer).RegistrarJugadas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.NameNode/RegistrarJugadas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NameNodeServer).RegistrarJugadas(ctx, req.(*SolicitudRegistrarJugadas))
	}
	return interceptor(ctx, in, info, handler)
}

func _NameNode_DevolverJugadas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudDevolverRegistro)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NameNodeServer).DevolverJugadas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.NameNode/DevolverJugadas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NameNodeServer).DevolverJugadas(ctx, req.(*SolicitudDevolverRegistro))
	}
	return interceptor(ctx, in, info, handler)
}

// NameNode_ServiceDesc is the grpc.ServiceDesc for NameNode service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NameNode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Proto.NameNode",
	HandlerType: (*NameNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegistrarJugadas",
			Handler:    _NameNode_RegistrarJugadas_Handler,
		},
		{
			MethodName: "DevolverJugadas",
			Handler:    _NameNode_DevolverJugadas_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Proto/SquidGame.proto",
}

// DataNodeClient is the client API for DataNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataNodeClient interface {
	RegistrarJugadas(ctx context.Context, in *SolicitudRegistrarJugadas, opts ...grpc.CallOption) (*RespuestaRegistrarJugadas, error)
	DevolverJugadas(ctx context.Context, in *SolicitudDevolverRegistro, opts ...grpc.CallOption) (*RespuestaDevolverRegistro, error)
}

type dataNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewDataNodeClient(cc grpc.ClientConnInterface) DataNodeClient {
	return &dataNodeClient{cc}
}

func (c *dataNodeClient) RegistrarJugadas(ctx context.Context, in *SolicitudRegistrarJugadas, opts ...grpc.CallOption) (*RespuestaRegistrarJugadas, error) {
	out := new(RespuestaRegistrarJugadas)
	err := c.cc.Invoke(ctx, "/Proto.DataNode/RegistrarJugadas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataNodeClient) DevolverJugadas(ctx context.Context, in *SolicitudDevolverRegistro, opts ...grpc.CallOption) (*RespuestaDevolverRegistro, error) {
	out := new(RespuestaDevolverRegistro)
	err := c.cc.Invoke(ctx, "/Proto.DataNode/DevolverJugadas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataNodeServer is the server API for DataNode service.
// All implementations must embed UnimplementedDataNodeServer
// for forward compatibility
type DataNodeServer interface {
	RegistrarJugadas(context.Context, *SolicitudRegistrarJugadas) (*RespuestaRegistrarJugadas, error)
	DevolverJugadas(context.Context, *SolicitudDevolverRegistro) (*RespuestaDevolverRegistro, error)
	mustEmbedUnimplementedDataNodeServer()
}

// UnimplementedDataNodeServer must be embedded to have forward compatible implementations.
type UnimplementedDataNodeServer struct {
}

func (UnimplementedDataNodeServer) RegistrarJugadas(context.Context, *SolicitudRegistrarJugadas) (*RespuestaRegistrarJugadas, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegistrarJugadas not implemented")
}
func (UnimplementedDataNodeServer) DevolverJugadas(context.Context, *SolicitudDevolverRegistro) (*RespuestaDevolverRegistro, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DevolverJugadas not implemented")
}
func (UnimplementedDataNodeServer) mustEmbedUnimplementedDataNodeServer() {}

// UnsafeDataNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataNodeServer will
// result in compilation errors.
type UnsafeDataNodeServer interface {
	mustEmbedUnimplementedDataNodeServer()
}

func RegisterDataNodeServer(s grpc.ServiceRegistrar, srv DataNodeServer) {
	s.RegisterService(&DataNode_ServiceDesc, srv)
}

func _DataNode_RegistrarJugadas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudRegistrarJugadas)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataNodeServer).RegistrarJugadas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.DataNode/RegistrarJugadas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataNodeServer).RegistrarJugadas(ctx, req.(*SolicitudRegistrarJugadas))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataNode_DevolverJugadas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SolicitudDevolverRegistro)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataNodeServer).DevolverJugadas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Proto.DataNode/DevolverJugadas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataNodeServer).DevolverJugadas(ctx, req.(*SolicitudDevolverRegistro))
	}
	return interceptor(ctx, in, info, handler)
}

// DataNode_ServiceDesc is the grpc.ServiceDesc for DataNode service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataNode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Proto.DataNode",
	HandlerType: (*DataNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegistrarJugadas",
			Handler:    _DataNode_RegistrarJugadas_Handler,
		},
		{
			MethodName: "DevolverJugadas",
			Handler:    _DataNode_DevolverJugadas_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Proto/SquidGame.proto",
}
