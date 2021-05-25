package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
)

//Struct of type application, this holds errorLog, infoLog values, as well as an instance of another struct of env variables we define below
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	cfg      *envConfig
}

//Below struct utilises package that makes reading env variables super easy. Below values are set to the specified env variables,
//and if those env variables arent present, theres a sensible default provided.
type envConfig struct {
	CertPath string `env:"CERT_PATH" envDefault:"/source/webhook-server-tls.crt"`
	KeyPath  string `env:"KeyPath" envDefault:"/source/webhook-server-tls.key"`
	// port     int    `env:"PORT" envDefault:"3000"`
	Port int `env:"PORT" envDefault:"3000"`
}

func main() {

	//The first thing we need to do is instanciate the application struct. To do this we will need the following:
	//1. an errorLog
	//2. an infoLog
	//3. an instantiated envConfig struct

	//1. an errorlog.
	//ToDo: Need fome further clarity aroung this. go have a look at the log.New() method to find out what the deal w/ those params is
	//can make an informed guess, but would rather actually know
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)

	//2. an infolog.
	//ToDo: same deal as 1
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)

	//3. an instantiated envConfig struct
	//So this just creates an empty struct?
	//A: Yes, because if you look at the struct you defined above, all the fields basically set themselves based off your env vars, or sensible defaults if the env vars arent present
	cfg := envConfig{}

	//Now we do an error check, looks like were calling
	//a method and passing in the above empty struct? and since we use the & before we pass it in, it passes as a pointer and therefore modifies the object?
	//if the err value returned from running env.Parse is anything other than nil, we log it as a fatal error and terminate execution?
	if err := env.Parse(&cfg); err != nil {
		errorLog.Fatalln(err)
	}

	//Now we have created all the constituent parts required to build out application struct, we cank build it
	//Not 100% nsure why 'application' needs to be passed as a pointer here? ($application), isnt this the first time we are creating it?
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		cfg:      &cfg,
	}

	//Commenting out tls cert load for now
	tlsPair, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		app.errorLog.Fatal("Error Loading TLS certs", err)
	}

	//Again go look into %v, as well as the tls.Config function
	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", cfg.Port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsPair}},
	}

	//This function will go off to another fikle, and handle all the webserver logic
	server.Handler = app.routes()

	// fmt.Println("Heres your app struct: ")
	// //This just prints fields of the struct/ stops compiler shouting at me - took it at face value from internet - maybe look into why %+v works/The Printf function
	// fmt.Printf("%+v\n", app)
	// fmt.Printf("%+v\n", server)

	//you have to look into this further :)
	//Splittng off into its own thread
	go func() {
		infoLog.Printf("Starting webserver on port %v", cfg.Port)
		app.errorLog.Println(server.ListenAndServeTLS("", ""))
		//app.errorLog.Println(server.ListenAndServe())
	}() // ie, why the () on the end? -find out

	//i know what this does in terms of high level funtionality, but zero idea whim im passing in those arguments/whats going on behind the scenes, look into it!
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	app.infoLog.Println("Got shutdown signal, shutting down web server gracefully...")

	server.Shutdown(context.Background())

}
