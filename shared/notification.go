package shared

import (
	"encoding/json"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"google.golang.org/appengine/urlfetch"
)

// PushNotification is a notification JSON object
type PushNotification struct {
	Title   string                  `json:"title,omitempty"`
	Options PushNotificationOptions `json:"options,omitempty"`
}

// PushNotificationOptions for the PushNotification
type PushNotificationOptions struct {
	Body  string               `json:"body,omitempty"`
	Icon  string               `json:"icon,omitempty"`  // URL String
	Image string               `json:"image,omitempty"` // URL String
	Badge string               `json:"badge,omitempty"` // URL String
	Data  PushNotificationData `json:"data,omitempty"`
}

// PushNotificationData holds data to send with the notification.
type PushNotificationData struct {
	URL string `json:"url,omitempty"`
}

// NotificationsAPI provides methods for notifications.
type NotificationsAPI interface {
	SendPushNotification(
		title string,
		msg string,
		path string,
		optIcon *string, // Optional custom icon.
	) error
}

type notificationsAPIImpl struct {
	aeAPI AppEngineAPI
}

// NewNotificationsAPI gets a NotificationsAPI implementation.
func NewNotificationsAPI(aeAPI AppEngineAPI) NotificationsAPI {
	return notificationsAPIImpl{
		aeAPI: aeAPI,
	}
}

func (n notificationsAPIImpl) SendPushNotification(
	title string,
	msg string,
	path string,
	optIcon *string, // Optional custom icon.
) error {
	icon := "/static/favicon.ico"
	if optIcon != nil {
		icon = *optIcon
	}
	notification, _ := json.Marshal(
		PushNotification{
			Title: title,
			Options: PushNotificationOptions{
				Body: msg,
				Icon: icon,
				Data: PushNotificationData{
					URL: path,
				},
			},
		},
	)
	// Push notifications
	ctx := n.aeAPI.Context()
	store := NewAppEngineDatastore(ctx, true)
	subscriptionsPrivateKey, err := GetSecret(store, "webpush-private-key")
	if err != nil {
		return err
	}
	keys, subs, err := GetSubscriptions(store)
	if err == nil {
		for i, sub := range subs {
			opts := webpush.Options{
				HTTPClient:      urlfetch.Client(ctx),
				Subscriber:      "mailto:ecosystem-infra-internal@google.com",
				VAPIDPublicKey:  subscriptionsPublicKey,
				VAPIDPrivateKey: subscriptionsPrivateKey,
			}
			resp, err := webpush.SendNotification(
				notification,
				&sub,
				&opts,
			)
			if err != nil {
				if resp.StatusCode == http.StatusGone {
					store.Delete(keys[i])
				}
				return err
			}
		}
	}
	return nil
}
