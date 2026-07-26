package main

import (
	"agent_feedback/internal/agent"
	"agent_feedback/internal/handler"
	"agent_feedback/internal/tools"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found. Relying on system environment variables.")
	}
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel("gpt-4.1-mini"),
	)
	if err != nil {
		log.Fatalf(
			"create OpenAI model: %v",
			err,
		)
	}

	classifier := agent.NewOpenAIClassifier(llm)
	reporter := agent.NewJSONReporter(0.6)

	customerTool, err := tools.NewCustomerTool(
		"./internal/datastore/customers.json",
	)
	if err != nil {
		log.Fatalf(
			"create customer tool: %v",
			err,
		)
	}

	policyTool, err := tools.NewPolicyTool(
		"./internal/datastore/policies.json",
	)
	if err != nil {
		log.Fatalf(
			"create policy tool: %v",
			err,
		)
	}

	workflowTool, err := tools.NewWorkflowTool(
		"./internal/datastore/workflows.json",
	)
	if err != nil {
		log.Fatalf(
			"create workflow tool: %v",
			err,
		)
	}

	feedbackAgent := agent.NewAgent(
		llm,
		reporter,
		classifier,
		*customerTool,
		*policyTool,
		*workflowTool,
	)

	feedbackHandler := handler.NewFeedbackHandler(
		feedbackAgent,
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/feedback/analyze",
		feedbackHandler.Analyze,
	)

	mux.HandleFunc(
		"/health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		},
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf(
		"server listening on http://localhost%s",
		server.Addr,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf(
			"server failed: %v",
			err,
		)
	}
}

func loggingMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			start := time.Now()

			next.ServeHTTP(w, r)

			log.Printf(
				"[HTTP] method=%s path=%s duration=%s",
				r.Method,
				r.URL.Path,
				time.Since(start),
			)
		},
	)
}
