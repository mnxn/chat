package protocol

type RequestVisitor interface {
	VisitConnect(*ConnectRequest)
	VisitDisconnect(*DisconnectRequest)
	VisitListRooms(*ListRoomsRequest)
	VisitListUsers(*ListUsersRequest)
	VisitMessageRoom(*MessageRoomRequest)
	VisitMessageUser(*MessageUserRequest)
	VisitCreateRoom(*CreateRoomRequest)
	VisitJoinRoom(*JoinRoomRequest)
	VisitLeaveRoom(*LeaveRoomRequest)
}

func (c *ConnectRequest) Accept(v RequestVisitor)    { v.VisitConnect(c) }
func (d *DisconnectRequest) Accept(v RequestVisitor) { v.VisitDisconnect(d) }

func (lr *ListRoomsRequest) Accept(v RequestVisitor) { v.VisitListRooms(lr) }
func (lu *ListUsersRequest) Accept(v RequestVisitor) { v.VisitListUsers(lu) }

func (mr *MessageRoomRequest) Accept(v RequestVisitor) { v.VisitMessageRoom(mr) }
func (mu *MessageUserRequest) Accept(v RequestVisitor) { v.VisitMessageUser(mu) }

func (cr *CreateRoomRequest) Accept(v RequestVisitor) { v.VisitCreateRoom(cr) }
func (jr *JoinRoomRequest) Accept(v RequestVisitor)   { v.VisitJoinRoom(jr) }
func (lr *LeaveRoomRequest) Accept(v RequestVisitor)  { v.VisitLeaveRoom(lr) }
