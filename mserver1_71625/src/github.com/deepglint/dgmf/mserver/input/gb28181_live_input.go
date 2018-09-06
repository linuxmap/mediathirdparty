package input

import (
	"time"

	"github.com/deepglint/dgmf/mserver/core"
	"github.com/deepglint/dgmf/mserver/protocols/gb28181"
	"github.com/golang/glog"
)

type GB28181LiveInput struct {
	receiver       *gb28181.GB28181Receiver
	retryInterval  time.Duration
	receiverStatus bool
	retryStatus    bool
	openStatus     bool
}

func (this *GB28181LiveInput) Open(uri string, stream *core.LiveStream) {
	this.retryStatus = true
	this.retryInterval = 1000 * time.Millisecond
	this.openStatus = true
	pool := core.GetESPool()

	go func() {
		for this.retryStatus {
			glog.V(2).Infof("[UDP_LIVE_INPUT] [STEAM_ID]=%s MServer will open a udp live input\n", stream.StreamId)
			this.receiver = &gb28181.GB28181Receiver{}
			rtms := make(chan core.RTMessage)
			go this.receiver.Open(uri, stream.StreamId, rtms)

			rtm := <-rtms
			if rtm.Status != 200 {
				glog.Warningf("[UDP_LIVE_INPUT] [STEAM_ID]=%s Udp server create faild, MServer will retry after %s, error: %s\n", stream.StreamId, this.retryInterval.String(), rtm.Error)
				time.Sleep(this.retryInterval)
				continue
			}

			glog.V(2).Infof("[UDP_LIVE_INPUT] [STEAM_ID]=%s Udp server create successed, MServer will recevice media data\n", stream.StreamId)
			this.receiverStatus = true

			for this.receiverStatus {
				select {
				case frame := <-this.receiver.Frames():
					if frame != nil {
						stream.Width = this.receiver.Width()
						stream.Height = this.receiver.Height()
						stream.SPS = this.receiver.SPS()
						stream.PPS = this.receiver.PPS()
						stream.Index = frame.Index
						stream.Fps = this.receiver.FPS()

						if frame.IFrame == true {
							stream.IFrame = *frame
						}

						pool.Live.RLock()
						for _, session := range stream.Sessions {
							select {
							case session.Frame <- frame:
							default:
							}
						}
						pool.Live.RUnlock()
					} else {
						this.receiverStatus = false
					}
				case rtm := <-rtms:
					if rtm.Status == 201 {
						glog.V(2).Infof("[RTSP_LIVE_INPUT] [STEAM_ID]=%s Rtsp receiver get a 201 signal, MServer will stop receive media data\n", stream.StreamId)
						this.receiverStatus = false
					}
					if rtm.Status == 400 {
						glog.Warningf("[RTSP_LIVE_INPUT] [STEAM_ID]=%s Rtsp receiver get a 400 signal, MServer will stop receive media data and retry after %s, error: %s\n", stream.StreamId, this.retryInterval.String(), rtm.Error)
						this.receiverStatus = false
					}
				}
			}
		}
	}()

	return
}

func (this *GB28181LiveInput) Close() {
	this.retryStatus = false
	this.receiverStatus = false
	this.receiver.Close()
	this.openStatus = false
}

func (this *GB28181LiveInput) Receiving() bool {
	return this.receiverStatus
}

func (this *GB28181LiveInput) Retry() bool {
	return this.retryStatus
}
func (this *GB28181LiveInput) Opened() bool {
	return this.openStatus
}
