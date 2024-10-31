export interface Message {
    /**
     * URL to go to when the action button is clicked
     */
    action: string;
    /**
     * Additional action buttons
     */
    actions?: Action[];
    /**
     * Notification body
     */
    body: string;
    /**
     * Type of event that triggered the notification
     */
    event: Event;
    /**
     * Unique ID for the notification scheduling request.
     */
    id: string;
    /**
     * URL to an image to display in the notification
     */
    image?: string;
    /**
     * Churros ID of the ressource (the ticket, the post, the comment, etc)
     * Used to determine to whom the notification should be sent
     * For godchild_request, this is not a user id, but a godparent request id.
     */
    object_id: string;
    /**
     * When to push the notification
     */
    send_at: Date;
    /**
     * Notification title
     */
    title: string;
}

export interface Action {
    action: string;
    label:  string;
}

/**
 * Type of event that triggered the notification
 */
export enum Event {
    BookingPaid = "booking_paid",
    ClearScheduledJobs = "clear_scheduled_jobs",
    CommentReply = "comment_reply",
    ContributionPaid = "contribution_paid",
    Custom = "custom",
    GodchildAccepted = "godchild_accepted",
    GodchildRejected = "godchild_rejected",
    GodchildRequest = "godchild_request",
    LoginStuck = "login_stuck",
    NewComment = "new_comment",
    NewPost = "new_post",
    PendingSignup = "pending_signup",
    ShotgunClosesSoon = "shotgun_closes_soon",
    ShotgunOpensSoon = "shotgun_opens_soon",
    Test = "test",
}
