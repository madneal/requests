package requests

type Config struct {
	Kafka struct {
		Brokers []string
		Topic   string
	}
}
