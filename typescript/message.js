"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Event = void 0;
/**
 * Type of event that triggered the notification
 */
var Event;
(function (Event) {
    Event["BookingPaid"] = "booking_paid";
    Event["ClearSchedule"] = "clear_schedule";
    Event["ClearScheduledJobs"] = "clear_scheduled_jobs";
    Event["ClearStoredSchedule"] = "clear_stored_schedule";
    Event["CommentReply"] = "comment_reply";
    Event["ContributionPaid"] = "contribution_paid";
    Event["Custom"] = "custom";
    Event["GodchildAccepted"] = "godchild_accepted";
    Event["GodchildRejected"] = "godchild_rejected";
    Event["GodchildRequest"] = "godchild_request";
    Event["LoginStuck"] = "login_stuck";
    Event["NewComment"] = "new_comment";
    Event["NewPost"] = "new_post";
    Event["PendingSignup"] = "pending_signup";
    Event["RestoreSchedule"] = "restore_schedule";
    Event["RestoreScheduleEager"] = "restore_schedule_eager";
    Event["SaveSchedule"] = "save_schedule";
    Event["ShotgunClosesSoon"] = "shotgun_closes_soon";
    Event["ShotgunOpensSoon"] = "shotgun_opens_soon";
    Event["ShowScheduledJobs"] = "show_scheduled_jobs";
    Event["Test"] = "test";
})(Event || (exports.Event = Event = {}));
