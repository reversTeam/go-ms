package core

import "context"

type Context struct {
	Main   context.Context
	Jeager *Jeager
}

func NewContext(jconfig JaegerConfig) *Context {
	ctx := context.Background()
	return &Context{
		Main:   ctx,
		Jeager: NewJeager(ctx, jconfig),
	}
}
