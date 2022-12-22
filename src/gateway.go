import (
    "log"
    "time"

    "github.com/apache/pulsar-client-go/pulsar"
)

type testJSON struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
// Create a Pulsar client
client, err := pulsar.NewClient(pulsar.ClientOptions{
    URL: "pulsar://localhost:6650",
})
if err != nil {
    log.Fatal(err)
}

defer client.Close()

// Start a separate goroutine for Prometheus metrics
// In this case, Prometheus metrics can be accessed via http://localhost:2112/metrics
go func() {
    prometheusPort := 2112
    log.Printf("Starting Prometheus metrics at http://localhost:%v/metrics\n", prometheusPort)
    http.Handle("/metrics", promhttp.Handler())
    err = http.ListenAndServe(":"+strconv.Itoa(prometheusPort), nil)
    if err != nil {
        log.Fatal(err)
    }
}()

// Create a producer
producer, err := client.CreateProducer(pulsar.ProducerOptions{
    Topic: "topic-1",
})
if err != nil {
    log.Fatal(err)
}

defer producer.Close()

ctx := context.Background()

// Write your business logic here
// In this case, you build a simple Web server. You can produce messages by requesting http://localhost:8082/produce
webPort := 8082
http.HandleFunc("/produce", func(w http.ResponseWriter, r *http.Request) {
    msgId, err := producer.Send(ctx, &pulsar.ProducerMessage{
        Payload: []byte(fmt.Sprintf("hello world")),
    })
    if err != nil {
        log.Fatal(err)
    } else {
        log.Printf("Published message: %v", msgId)
        fmt.Fprintf(w, "Published message: %v", msgId)
    }
})

err = http.ListenAndServe(":"+strconv.Itoa(webPort), nil)
if err != nil {
    log.Fatal(err)
}
}
