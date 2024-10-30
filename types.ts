export interface Message {
    /**
     * URL to go to when the action button is clicked
     */
    action: string;
    /**
     * Additional action buttons
     */
    actions: Action[];
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
    image: string;
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
    CommentReply = "comment_reply",
    GodchildRequest = "godchild_request",
    NewComment = "new_comment",
    NewPost = "new_post",
    NewTicket = "new_ticket",
}
