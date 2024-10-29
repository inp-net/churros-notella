package notella

type Event string

const (
	EventNewTicket       Event = "new_ticket"
	EventNewPost         Event = "new_post"
	EventGodchildRequest Event = "godchild_request"
	EventNewComment      Event = "new_comment"
	EventCommentReply    Event = "comment_reply"
)

type Message struct {
	Id string `json:"id"`
	// IMPORTANT: Keep this up to date!!!
	Event Event `json:"event" jsonschema:"enum=new_ticket,enum=new_post,enum=godchild_request,enum=new_comment,enum=comment_reply"`
}
