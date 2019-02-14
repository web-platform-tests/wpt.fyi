package shared

import (
	webpush "github.com/SherClockHolmes/webpush-go"
	"google.golang.org/appengine/datastore"
)

// Generated @ https://tools.reactpwa.com/vapid?email=ecosystem-infra-internal%40google.com
const subscriptionsPublicKey = "BGf5diO3W8TqDldUEAUSFDbKLztmzAgoU14oRjrvMQZn0ceeRdq6hJCvF526DmXyljXmeVM6avvjkRyXI7PYebk"

// AddSubscription stores the given subscription, for use when pushing.
func AddSubscription(aeAPI AppEngineAPI, sub webpush.Subscription) error {
	ctx := aeAPI.Context()
	key := datastore.NewIncompleteKey(ctx, "Subscription", nil)
	_, err := datastore.Put(ctx, key, &sub)
	return err
}

// GetSubscriptions gets any stored subscriptions.
func GetSubscriptions(store Datastore) ([]Key, []webpush.Subscription, error) {
	q := store.NewQuery("Subscription")
	var subs []webpush.Subscription
	keys, err := store.GetAll(q, &subs)
	return keys, subs, err
}
