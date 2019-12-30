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
	Network struct {
		Network [5]string
	}
}
