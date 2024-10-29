package notella

import "time"

type Event string

const (
	EventNewTicket       Event = "new_ticket"
	EventNewPost         Event = "new_post"
	EventGodchildRequest Event = "godchild_request"
	EventNewComment      Event = "new_comment"
	EventCommentReply    Event = "comment_reply"
)

type Message struct {
	// Unique ID for the notification scheduling request.
	Id string `json:"id"`
	// When to push the notification
	SendAt time.Time `json:"send_at"`
	// Type of event that triggered the notification
	Event Event `json:"event" jsonschema:"enum=new_ticket,enum=new_post,enum=godchild_request,enum=new_comment,enum=comment_reply"`
	// Churros ID of the ressource (the ticket, the post, the comment, etc)
	// Used to determine to whom the notification should be sent
	// For godchild_request, this is not a user id, but a godparent request id.
	ChurrosObjectId string `json:"object_id"`
	// Notification title
	Title string `json:"title"`
	// Notification body
	Body string `json:"body"`
	// URL to go to when the notification is clicked
	Action string `json:"action"`
	// Additional action buttons
	Actions []struct {
		// Label of the action button
		Label string `json:"label"`
		// URL to go to when the action button is clicked
		Action string `json:"action"`
	} `json:"actions"`
	// URL to an image to display in the notification
	Image string `json:"image"`
}
