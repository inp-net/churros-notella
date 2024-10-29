export interface Message {
    event: Event;
    id:    string;
}

export enum Event {
    CommentReply = "comment_reply",
    GodchildRequest = "godchild_request",
    NewComment = "new_comment",
    NewPost = "new_post",
    NewTicket = "new_ticket",
}

