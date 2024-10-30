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
	case db.VisibilityPrivate:
	case db.VisibilityUnlisted:
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
