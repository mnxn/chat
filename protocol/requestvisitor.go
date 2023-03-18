package protocol

type RequestVisitor interface {
	Keepalive(*KeepaliveRequest)
	Connect(*ConnectRequest)
	Disconnect(*DisconnectRequest)
	ListRooms(*ListRoomsRequest)
	ListUsers(*ListUsersRequest)
	MessageRoom(*MessageRoomRequest)
	MessageUser(*MessageUserRequest)
	CreateRoom(*CreateRoomRequest)
	JoinRoom(*JoinRoomRequest)
	LeaveRoom(*LeaveRoomRequest)
}

func (k *KeepaliveRequest) Accept(v RequestVisitor) { v.Keepalive(k) }

func (c *ConnectRequest) Accept(v RequestVisitor)    { v.Connect(c) }
func (d *DisconnectRequest) Accept(v RequestVisitor) { v.Disconnect(d) }

func (lr *ListRoomsRequest) Accept(v RequestVisitor) { v.ListRooms(lr) }
func (lu *ListUsersRequest) Accept(v RequestVisitor) { v.ListUsers(lu) }

func (mr *MessageRoomRequest) Accept(v RequestVisitor) { v.MessageRoom(mr) }
func (mu *MessageUserRequest) Accept(v RequestVisitor) { v.MessageUser(mu) }

func (cr *CreateRoomRequest) Accept(v RequestVisitor) { v.CreateRoom(cr) }
func (jr *JoinRoomRequest) Accept(v RequestVisitor)   { v.JoinRoom(jr) }
func (lr *LeaveRoomRequest) Accept(v RequestVisitor)  { v.LeaveRoom(lr) }
