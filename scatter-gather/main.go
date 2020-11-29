package main

import (
	"time"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
	"github.com/AsynkronIT/protoactor-go/scheduler"
	"github.com/kiyonori-matsumoto/protoactor-in-study/utils"
	"github.com/thoas/go-funk"
)

type PhotoMessage struct {
	id           string
	photo        string
	creationTime time.Time
	speed        int
}

type GetSpeed struct {
	pipe *actor.PID
}

func (state *GetSpeed) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *PhotoMessage:
		var newmsg PhotoMessage
		newmsg.id = msg.id
		newmsg.photo = msg.photo
		newmsg.speed = 100 // 実際は計算
		context.Send(state.pipe, &newmsg)
	}
}

type GetTime struct {
	pipe *actor.PID
}

func (state *GetTime) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *PhotoMessage:
		var newmsg PhotoMessage
		newmsg.id = msg.id
		newmsg.photo = msg.photo
		newmsg.creationTime = time.Now() // 実際は計算
		context.Send(state.pipe, &newmsg)
	}
}

type ReceipentList struct {
	receipentList []*actor.PID
}

func (state *ReceipentList) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *PhotoMessage:
		for _, receipent := range state.receipentList {
			context.Forward(receipent)
		}
	}

}

type Aggregator struct {
	timeout  time.Duration
	pipe     *actor.PID
	messages []*PhotoMessage
}

type timeout struct {
	message *PhotoMessage
}

func (state *Aggregator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *PhotoMessage:
		_target := funk.Find(state.messages, func(x *PhotoMessage) bool {
			return x.id == msg.id
		})
		if _target == nil {
			state.messages = append(state.messages, msg)
			aggregatorScheduler := scheduler.NewTimerScheduler(context)
			aggregatorScheduler.SendOnce(state.timeout, context.Self(), &timeout{message: msg})
			return
		}
		if target, ok := _target.(*PhotoMessage); ok {
			var creationTime time.Time
			if !msg.creationTime.IsZero() {
				creationTime = msg.creationTime
			} else {
				creationTime = target.creationTime
			}

			newmsg := &PhotoMessage{
				id:           msg.id,
				photo:        msg.photo,
				speed:        msg.speed + target.speed,
				creationTime: creationTime,
			}

			context.Send(state.pipe, newmsg)
			idx := funk.IndexOf(state.messages, target)
			state.messages = append(state.messages[:idx], state.messages[idx+1:]...)
		}
	case *timeout:
		_target := funk.Find(state.messages, func(x *PhotoMessage) bool {
			return x.id == msg.message.id
		})
		if _target == nil {
			// 既に処理済みのため何もしない
			return
		}
		if target, ok := _target.(*PhotoMessage); ok {
			idx := funk.IndexOf(state.messages, target)
			// 2つ目のメッセージを受信しなかった場合はじめのメッセージを送信
			context.Send(state.pipe, target)
			state.messages = append(state.messages[:idx], state.messages[idx+1:]...)
		}
	case *actor.Restarting:
		for _, target := range state.messages {
			context.Send(state.pipe, target)
			state.messages = nil
		}
	}
}

func main() {

	system := actor.NewActorSystem()

	endProbe := system.Root.Spawn(utils.NewEndProbeProps())

	aggregator := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor {
		return &Aggregator{timeout: 1 * time.Second, pipe: endProbe, messages: make([]*PhotoMessage, 0)}
	}).WithReceiverMiddleware(middleware.Logger))

	getspeed := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor { return &GetSpeed{pipe: aggregator} }).WithReceiverMiddleware(middleware.Logger))

	gettime := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor { return &GetTime{pipe: aggregator} }).WithReceiverMiddleware(middleware.Logger))

	receipent := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor { return &ReceipentList{receipentList: []*actor.PID{getspeed, gettime}} }).WithReceiverMiddleware(middleware.Logger))

	msg := &PhotoMessage{id: "123xyz", photo: "photo"}

	system.Root.Send(receipent, msg)

	_, _ = console.ReadLine()
}
