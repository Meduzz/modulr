package event

type (
	subscribee struct {
		ID      string
		Address string
		Port    int
		Context string
		Path    string
	}

	subscription struct {
		Topic    string
		Routing  string
		Group    string
		Services []*subscribee
	}
)
