package dispatcher

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/ltvsocket"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/golang/protobuf/proto"
	"log"
)

type errInterface interface {
	Error() string
}

var (
	msgConnected = uint32(cellnet.Type2ID(&coredef.ConnectedACK{}))
	msgAccepted  = uint32(cellnet.Type2ID(&coredef.AcceptedACK{}))
	msgClosed    = uint32(cellnet.Type2ID(&coredef.ClosedACK{}))
)

// 处理Peer的新会话及会话的消息处理
func PeerHandler(disp *PacketDispatcher) func(cellnet.CellID, interface{}) {

	return func(peer cellnet.CellID, peerev interface{}) {

		switch v := peerev.(type) {
		case ltvsocket.EventNewSession: // 新的连接生成

			var msgid uint32
			switch peerev.(type) {
			case ltvsocket.EventConnected:
				msgid = msgConnected
			case ltvsocket.EventAccepted:
				msgid = msgAccepted
			}

			ltvsocket.SpawnSession(v.Stream(), func(ses cellnet.CellID, sesev interface{}) {

				switch data := sesev.(type) {

				case cellnet.EventInit: // 初始化转通知
					disp.Call(ses, &cellnet.Packet{MsgID: msgid})
				case ltvsocket.EventClose: // 断开转通知
					disp.Call(ses, &cellnet.Packet{MsgID: msgClosed})
				case *cellnet.Packet: // 收
					disp.Call(ses, data)
				case proto.Message: // 发
					v.Stream().Write(cellnet.BuildPacket(data))
				}

			})

		case errInterface:
			log.Println(cellnet.ReflectContent(v))
		}

	}
}
