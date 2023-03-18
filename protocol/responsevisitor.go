package protocol

type ResponseVisitor interface {
	Error(*ErrorResponse)
	FatalError(*FatalErrorResponse)
	RoomList(*RoomListResponse)
	UserList(*UserListResponse)
	RoomMessage(*RoomMessageResponse)
	UserMessage(*UserMessageResponse)
}

func (e *ErrorResponse) Accept(v ResponseVisitor)       { v.Error(e) }
func (fe *FatalErrorResponse) Accept(v ResponseVisitor) { v.FatalError(fe) }

func (rl *RoomListResponse) Accept(v ResponseVisitor) { v.RoomList(rl) }
func (ul *UserListResponse) Accept(v ResponseVisitor) { v.UserList(ul) }

func (rm *RoomMessageResponse) Accept(v ResponseVisitor) { v.RoomMessage(rm) }
func (um *UserMessageResponse) Accept(v ResponseVisitor) { v.UserMessage(um) }
