package main

import (
	console "github.com/AsynkronIT/goconsole"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/kiyonori-matsumoto/protoactor-in-study/utils"
)

type Photo struct {
	license string
	speed   int
}

type SpeedFilter struct {
	minSpeed int
	pipe     *actor.PID
}

func (state *SpeedFilter) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *Photo:
		if msg.speed > state.minSpeed {
			context.Send(state.pipe, msg)
		}
	}
}

type LicenseFilter struct {
	pipe *actor.PID
}

func (state *LicenseFilter) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *Photo:
		if msg.license != "" {
			context.Send(state.pipe, msg)
		}
	}
}

// type EndProbe struct{}

// func (*EndProbe) Receive(context actor.Context) {
// 	switch msg := context.Message().(type) {
// 	default:
// 		fmt.Println("received message", reflect.TypeOf(msg), msg)
// 	}
// }

// func NewEndProbeProps() *actor.Props {
// 	return actor.PropsFromProducer(func() actor.Actor { return &EndProbe{} })
// }

func main() {
	system := actor.NewActorSystem()

	endProbe := system.Root.Spawn(utils.NewEndProbeProps())

	speedFilter := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor { return &SpeedFilter{minSpeed: 50, pipe: endProbe} }))

	licenseFilter := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor { return &LicenseFilter{pipe: speedFilter} }))

	msg := &Photo{license: "123xyz", speed: 60}

	system.Root.Send(licenseFilter, msg)

	_, _ = console.ReadLine()
}
