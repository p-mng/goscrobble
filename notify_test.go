package main_test

type FakeNotifier struct {
	Notifications int
}

func (m *FakeNotifier) SendNotification(_ uint32, _, _ string) (uint32, error) {
	m.Notifications++
	return 0, nil
}
