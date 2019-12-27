package main

type Config struct {
	Kafka struct {
		Brokers []string
		Topic   string
	}
	Database struct {
		Host string
		Port int
		User string
		Pass string
		Name string
	}
}
