package main

type Config struct {
	Kafka struct {
		Brokers []string
		Topic   string
		GroupId string
	}
	Database struct {
		Host string
		Port int
		User string
		Pass string
		Name string
	}
	Network struct {
		Network [3]string
		Proxy   string
	}
	Run struct {
		Threads      int
		Debug        bool
		Redis        bool
		IsProduction bool
	}
	Redis struct {
		Host     string
		Port     int
		Db       int
		Password string
		Set      string
	}
}
