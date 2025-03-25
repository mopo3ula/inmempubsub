package inmemory

type Topic string

func (t Topic) String() string {
	return string(t)
}

func NewTopic(topicName string) Topic {
	return Topic(topicName)
}
