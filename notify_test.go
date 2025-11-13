package main_test

type MockNotifier struct {
	Notifications int
}

func (m *MockNotifier) SendNotification(_ uint32, _, _ string) (uint32, error) {
	m.Notifications++
	return 0, nil
}
