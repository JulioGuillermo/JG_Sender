package notification

import "gioui.org/x/notify"

func InitNotifier() notify.Notifier {
	defer func() {
		recover()
	}()
	not, e := notify.NewNotifier()
	if e != nil {
		not = nil
	}
	return not
}
