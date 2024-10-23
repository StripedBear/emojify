package main

import (
    "fmt"
    "html/template"
    "net/http"
    "strings"
    "unicode/utf8"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"path"},
    )
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(requestDuration)
}

func emojiToCode(text string) string {
    var result strings.Builder
    for _, r := range text {
        if utf8.RuneLen(r) > 1 {
            result.WriteString(fmt.Sprintf("\\U+%04X ", r))
        } else {
            result.WriteRune(r)
        }
    }
    return result.String()
}

func codeToEmoji(text string) string {
    parts := strings.Split(text, " ")
    var result strings.Builder
    for _, part := range parts {
        if strings.HasPrefix(part, "\\U+") {
            var r rune
            fmt.Sscanf(part, "\\U+%X", &r)
            result.WriteRune(r)
        } else {
            result.WriteString(part)
        }
    }
    return result.String()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    timer := prometheus.NewTimer(requestDuration.WithLabelValues(r.URL.Path))
    defer timer.ObserveDuration()

    httpRequestsTotal.WithLabelValues(r.URL.Path).Inc()

    tmpl := template.Must(template.New("index").Parse(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <title>Emoji Converter</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    background: linear-gradient(to bottom, #b800ff, #FFFFFF, #1E90FF);
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    height: 100vh;
                    margin: 0;
                }
                .container {
                    background: rgba(255, 255, 255, 0.8);
                    padding: 20px;
                    border-radius: 10px;
                    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
                    width: 400px;
                    text-align: center;
                }
                h1 {
                    color: #FF4500;
                }
                form {
                    margin-bottom: 20px;
                }
                textarea {
                    width: 100%;
                    padding: 10px;
                    border-radius: 5px;
                    border: 1px solid #CCCCCC;
                }
                input[type="submit"] {
                    background-color: #1E90FF;
                    color: white;
                    border: none;
                    padding: 10px 20px;
                    border-radius: 5px;
                    cursor: pointer;
                }
                input[type="submit"]:hover {
                    background-color: #4682B4;
                }
                pre {
                    background-color: #F0F8FF;
                    padding: 10px;
                    border-radius: 5px;
                    border: 1px solid #CCCCCC;
                    text-align: left;
                }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>Emoji Converter</h1>
                <form method="POST" action="/convert">
                    <textarea name="text" rows="4" cols="50" placeholder="Enter text with emojis..."></textarea>
                    <br>
                    <input type="submit" value="Convert to Codes">
                </form>

                <form method="POST" action="/reverse">
                    <textarea name="text" rows="4" cols="50" placeholder="Enter text with codes like \U+1F600..."></textarea>
                    <br>
                    <input type="submit" value="Convert to Emojis">
                </form>

                <h2>Result:</h2>
                <pre>{{.}}</pre>
            </div>
        </body>
        </html>
    `))

    var result string
    if r.Method == http.MethodPost {
        input := r.FormValue("text")
        if r.URL.Path == "/convert" {
            result = emojiToCode(input)
        } else if r.URL.Path == "/reverse" {
            result = codeToEmoji(input)
        }
    }

    tmpl.Execute(w, result)
}

func main() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/convert", indexHandler)
    http.HandleFunc("/reverse", indexHandler)
    http.Handle("/metrics", promhttp.Handler())

    fmt.Println("Server is running on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
}
