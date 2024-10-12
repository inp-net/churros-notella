package notella

type Event = string

var (
	EventNewTicket       Event = "new_ticket"
	EventNewPost         Event = "new_post"
	EventGodchildRequest Event = "godchild_request"
	EventNewComment      Event = "new_comment"
	EventCommentReply    Event = "comment_reply"
)
