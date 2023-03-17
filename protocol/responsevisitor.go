package protocol

type ResponseVisitor interface {
	VisitError(*ErrorResponse)
	VisitFatalError(*FatalErrorResponse)
	VisitRoomList(*RoomListResponse)
	VisitUserList(*UserListResponse)
	VisitRoomMessage(*RoomMessageResponse)
	VisitUserMessage(*UserMessageResponse)
}

func (e *ErrorResponse) Accept(v ResponseVisitor)       { v.VisitError(e) }
func (fe *FatalErrorResponse) Accept(v ResponseVisitor) { v.VisitFatalError(fe) }

func (rl *RoomListResponse) Accept(v ResponseVisitor) { v.VisitRoomList(rl) }
func (ul *UserListResponse) Accept(v ResponseVisitor) { v.VisitUserList(ul) }

func (rm *RoomMessageResponse) Accept(v ResponseVisitor) { v.VisitRoomMessage(rm) }
func (um *UserMessageResponse) Accept(v ResponseVisitor) { v.VisitUserMessage(um) }
