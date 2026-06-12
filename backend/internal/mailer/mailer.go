package mailer

type Mailer interface {
	Send(to string, subject string, body string) error
}

type Noop struct{}

func (Noop) Send(_ string, _ string, _ string) error {
	return nil
}
