package core

import "context"

type Context struct {
	Name   string
	Main   context.Context
	Jeager *Jeager
}

func NewContext(name string, jconfig JaegerConfig) *Context {
	ctx := context.Background()
	return &Context{
		Name:   name,
		Main:   ctx,
		Jeager: NewJeager(ctx, name, jconfig),
	}
}
