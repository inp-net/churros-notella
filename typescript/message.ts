import * as z from "zod";

// Type of event that triggered the notification

export const EventSchema = z.enum([
    "booking_paid",
    "clear_schedule",
    "clear_scheduled_jobs",
    "clear_stored_schedule",
    "comment_reply",
    "contribution_paid",
    "custom",
    "godchild_accepted",
    "godchild_rejected",
    "godchild_request",
    "login_stuck",
    "new_comment",
    "new_post",
    "pending_signup",
    "restore_schedule",
    "restore_schedule_eager",
    "save_schedule",
    "shotgun_closes_soon",
    "shotgun_opens_soon",
    "show_scheduled_jobs",
    "test",
]);
export type Event = z.infer<typeof EventSchema>;

export const ActionSchema = z.object({
    "action": z.string(),
    "label": z.string(),
});
export type Action = z.infer<typeof ActionSchema>;

export const MessageSchema = z.object({
    "action": z.string(),
    "actions": z.array(ActionSchema).optional(),
    "body": z.string(),
    "event": EventSchema,
    "id": z.string(),
    "image": z.string().optional(),
    "object_id": z.string(),
    "send_at": z.coerce.date(),
    "title": z.string(),
});
export type Message = z.infer<typeof MessageSchema>;
