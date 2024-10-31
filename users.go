package notella

import (
	"context"
	"fmt"

	"git.inpt.fr/churros/notella/db"
)

// AllUsers returns all the users in the database that have at least one notification subscription
func AllUsers() ([]string, error) {
	users, err := prisma.User.FindMany(
		db.User.NotificationSubscriptions.Some(
			db.NotificationSubscription.Endpoint.Not(""),
		),
	).Select(
		db.User.ID.Field(),
	).Exec(context.Background())

	if err != nil {
		return []string{}, fmt.Errorf("while getting all users: %w", err)
	}

	ids := make([]string, len(users))
	for i, user := range users {
		ids[i] = user.ID
	}

	return ids, nil
}

// Receivers determines which users to send the notification to
func Receivers(message Message) ([]string, error) {
	switch message.Event {
	case EventNewPost:
		return receiversForPost(message)
	case EventNewComment:
		return receiversForNewComment(message)
	case EventBookingPaid:
		return receiversForBookingPaid(message)
	case EventCommentReply:
		return receiversForCommentReply(message)
	case EventContributionPaid:
		return receiversForContributionPaid(message)
	case EventGodchildAccepted, EventGodchildRejected:
		return receiversForGodchildResponse(message)
	case EventGodchildRequest:
		return receiversForGodchildRequest(message)
	case EventShotgunOpensSoon:
		return receiversForShotgunOpens(message)
	case EventShotgunClosesSoon:
		return receiversForShotgunCloses(message)
	}

	// For other events, assume the message churros object id is the user id
	if message.ChurrosObjectId != "" {
		_, err := prisma.User.FindUnique(
			db.User.ID.Equals(message.ChurrosObjectId),
		).Exec(context.Background())
		if err == nil {
			return []string{message.ChurrosObjectId}, nil
		}
	}

	return []string{}, nil
}

func receiversForPost(message Message) (userIds []string, err error) {
	post, err := prisma.Article.FindUnique(
		db.Article.ID.Equals(message.ChurrosObjectId),
	).With(
		db.Article.Group.Fetch().With(
			db.Group.Members.Fetch().Select(
				db.GroupMember.MemberID.Field(),
			),
			db.Group.StudentAssociation.Fetch().With(
				db.StudentAssociation.School.Fetch().With(
					db.School.Majors.Fetch().With(
						db.Major.Students.Fetch().Select(
							db.User.ID.Field(),
						),
					),
				),
			),
		),
	).Exec(context.Background())

	if err != nil {
		return []string{}, fmt.Errorf("while getting the post %q: %w", message.Id, err)
	}

	switch post.Visibility {
	case db.VisibilityPrivate, db.VisibilityUnlisted:
		return []string{}, nil
	case db.VisibilityPublic:
		return AllUsers()
	case db.VisibilitySchoolRestricted:
		for _, major := range post.Group().StudentAssociation().School().Majors() {
			for _, student := range major.Students() {
				userIds = append(userIds, student.ID)
			}
		}
		return
	case db.VisibilityGroupRestricted:
		for _, member := range post.Group().Members() {
			userIds = append(userIds, member.MemberID)
		}
	}

	return userIds, fmt.Errorf("unknown post visibility %q", post.Visibility)
}

func receiversForNewComment(message Message) (userIds []string, err error) {
	comment, err := prisma.Comment.FindUnique(
		db.Comment.ID.Equals(message.ChurrosObjectId),
	).With(
		db.Comment.Article.Fetch(),
	).Exec(context.Background())
	if err != nil {
		err = fmt.Errorf("while getting comment: %w", err)
		return
	}

	post, ok := comment.Article()
	if !ok {
		err = fmt.Errorf("comment %q has no parent post", comment.ID)
		return
	}

	authorId, ok := post.AuthorID()
	if !ok {
		err = fmt.Errorf("post %q has no author", post.ID)
		return
	}

	return []string{authorId}, nil
}

func receiversForBookingPaid(message Message) (userIds []string, err error) {
	booking, err := prisma.Registration.FindUnique(
		db.Registration.ID.Equals(message.ChurrosObjectId),
	).Exec(context.Background())

	if err != nil {
		err = fmt.Errorf("while getting booking: %w", err)
		return
	}

	authorId, ok := booking.AuthorID()
	if ok {
		userIds = append(userIds, authorId)
	}

	beneficiaryId, ok := booking.InternalBeneficiaryID()
	if ok {
		userIds = append(userIds, beneficiaryId)
	}

	return
}

func receiversForCommentReply(message Message) (userIds []string, err error) {
	comment, err := prisma.Comment.FindUnique(
		db.Comment.ID.Equals(message.ChurrosObjectId),
	).With(
		db.Comment.InReplyTo.Fetch(),
	).Exec(context.Background())
	if err != nil {
		err = fmt.Errorf("while getting comment: %w", err)
		return
	}

	parent, ok := comment.InReplyTo()
	if !ok {
		err = fmt.Errorf("comment %q has no parent", comment.ID)
		return
	}

	authorId, ok := parent.AuthorID()
	if !ok {
		err = fmt.Errorf("comment %q has no author", parent.ID)
		return
	}

	return []string{authorId}, nil
}

func receiversForContributionPaid(message Message) (userIds []string, err error) {
	contribution, err := prisma.Contribution.FindUnique(
		db.Contribution.ID.Equals(message.ChurrosObjectId),
	).Exec(context.Background())

	if err != nil {
		err = fmt.Errorf("while getting contribution: %w", err)
		return
	}

	return []string{contribution.UserID}, nil
}

func receiversForGodchildResponse(message Message) (userIds []string, err error) {
	godchildRequest, err := prisma.GodparentRequest.FindUnique(
		db.GodparentRequest.ID.Equals(message.ChurrosObjectId),
	).Exec(context.Background())

	if err != nil {
		err = fmt.Errorf("while getting godchild request: %w", err)
		return
	}

	return []string{godchildRequest.GodchildID}, nil
}

func receiversForGodchildRequest(message Message) (userIds []string, err error) {
	godchildRequest, err := prisma.GodparentRequest.FindUnique(
		db.GodparentRequest.ID.Equals(message.ChurrosObjectId),
	).Exec(context.Background())

	if err != nil {
		err = fmt.Errorf("while getting godchild request: %w", err)
		return
	}

	return []string{godchildRequest.GodparentID}, nil
}

func receiversForShotgunOpens(message Message) (userIds []string, err error) {
	shotgun, err := prisma.Event.FindUnique(
		db.Event.ID.Equals(message.ChurrosObjectId),
	).With(
		db.Event.Group.Fetch().With(
			db.Group.Members.Fetch().Select(
				db.GroupMember.MemberID.Field(),
			),
			db.Group.StudentAssociation.Fetch().With(
				db.StudentAssociation.School.Fetch().With(
					db.School.Majors.Fetch().With(
						db.Major.Students.Fetch().Select(
							db.User.ID.Field(),
						),
					),
				),
			),
		),
	).Exec(context.Background())

	userIds = make([]string, 0)

	switch shotgun.Visibility {
	case db.VisibilityPublic:
		return AllUsers()
	case db.VisibilitySchoolRestricted:
		for _, major := range shotgun.Group().StudentAssociation().School().Majors() {
			for _, student := range major.Students() {
				userIds = append(userIds, student.ID)
			}
		}
		return
	case db.VisibilityGroupRestricted:
		for _, member := range shotgun.Group().Members() {
			userIds = append(userIds, member.MemberID)
		}
		return
	}

	return
}

func receiversForShotgunCloses(message Message) (userIds []string, err error) {
	shotgun, err := prisma.Event.FindUnique(
		db.Event.ID.Equals(message.ChurrosObjectId),
	).With(
		db.Event.Group.Fetch().With(
			db.Group.Members.Fetch().Select(
				db.GroupMember.MemberID.Field(),
			),
			db.Group.StudentAssociation.Fetch().With(
				db.StudentAssociation.School.Fetch().With(
					db.School.Majors.Fetch().With(
						db.Major.Students.Fetch().Select(
							db.User.ID.Field(),
						),
					),
				),
			),
		),
		db.Event.Tickets.Fetch().With(
			db.Ticket.Registrations.Fetch().Select(
				db.Registration.AuthorID.Field(),
				db.Registration.InternalBeneficiaryID.Field(),
			),
		),
	).Exec(context.Background())

	switch shotgun.Visibility {
	case db.VisibilityPublic:
		return AllUsers()
	case db.VisibilitySchoolRestricted:
		for _, major := range shotgun.Group().StudentAssociation().School().Majors() {
			for _, student := range major.Students() {
				userIds = append(userIds, student.ID)
			}
		}
		return
	case db.VisibilityGroupRestricted:
		for _, member := range shotgun.Group().Members() {
			userIds = append(userIds, member.MemberID)
		}
		return
	}

	// Remove users that are booked to the event

	usersToRemove := make([]string, 0)

	for _, ticket := range shotgun.Tickets() {
		for _, registration := range ticket.Registrations() {
			authorId, ok := registration.AuthorID()
			if ok {
				usersToRemove = append(usersToRemove, authorId)
			}

			beneficiaryId, ok := registration.InternalBeneficiaryID()
			if ok {
				usersToRemove = append(usersToRemove, beneficiaryId)
			}
		}
	}

	for _, user := range usersToRemove {
		for i, id := range userIds {
			if id == user {
				userIds = append(userIds[:i], userIds[i+1:]...)
				break
			}
		}
	}

	return
}
