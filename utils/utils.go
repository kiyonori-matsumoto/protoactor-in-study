package utils

import (
	"fmt"
	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type EndProbe struct{}

func (*EndProbe) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	default:
		fmt.Println("received message", reflect.TypeOf(msg), msg)
	}
}

func NewEndProbeProps() *actor.Props {
	return actor.PropsFromProducer(func() actor.Actor { return &EndProbe{} })
}
