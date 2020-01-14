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
		Proxy   string
	}
	Run struct {
		Threads int
		Debug   bool
		Redis   bool
	}
	Redis struct {
		Host     string
		Port     int
		Db       int
		Password string
		Set      string
	}
}
