package notella

import "time"

type Event string

const (
	EventClearStoredSchedule Event = "clear_stored_schedule"
	EventShowScheduledJobs   Event = "show_scheduled_jobs"
	EventSaveSchedule        Event = "save_schedule"
	EventRestoreSchedule     Event = "restore_schedule"
	// Like restore_schedule, but also re-schedules events that have send_at in the past
	EventRestoreScheduleEager Event = "restore_schedule_eager"
	EventClearSchedule        Event = "clear_schedule"
	EventNewPost              Event = "new_post"
	EventGodchildRequest      Event = "godchild_request"
	EventCustom               Event = "custom"
	EventTest                 Event = "test"
	EventGodchildAccepted     Event = "godchild_accepted"
	EventGodchildRejected     Event = "godchild_rejected"
	EventPendingSignup        Event = "pending_signup"
	EventLoginStuck           Event = "login_stuck"
	EventBookingPaid          Event = "booking_paid"
	EventContributionPaid     Event = "contribution_paid"
	EventShotgunOpensSoon     Event = "shotgun_opens_soon"
	EventShotgunClosesSoon    Event = "shotgun_closes_soon"
)

type Message struct {
	// Unique ID for the notification scheduling request.
	Id string `json:"id"`
	// When to push the notification
	SendAt time.Time `json:"send_at"`
	// Clear scheduled jobs for the given event types on the object_id before scheduling notifications
	// TODO: generate jsonschema annotation
	ClearScheduleFor []Event `json:"clear_schedule_for"`
	// Type of event that triggered the notification
	// next-line-generate event-enum-jsonschema-values
	Event Event `json:"event" jsonschema:"enum=clear_stored_schedule,enum=show_scheduled_jobs,enum=save_schedule,enum=restore_schedule,enum=restore_schedule_eager,enum=clear_schedule,enum=new_post,enum=godchild_request,enum=custom,enum=test,enum=godchild_accepted,enum=godchild_rejected,enum=pending_signup,enum=login_stuck,enum=booking_paid,enum=contribution_paid,enum=shotgun_opens_soon,enum=shotgun_closes_soon"`
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
	} `json:"actions,omitempty"`
	// URL to an image to display in the notification
	Image string `json:"image,omitempty"`
}
