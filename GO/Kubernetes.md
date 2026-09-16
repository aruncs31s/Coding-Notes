---
id: kubernetes_go
aliases:
  - Kubernetes in Go
  - Deploying Go on K8s
tags:
  - go
  - kubernetes
  - k8s
  - devops
  - containerization
dg-publish: true
---

# Kubernetes in Go: Containers, Probes & Graceful Shutdown

Deploying Go applications to **Kubernetes (K8s)** provides automated scaling, self-healing, rolling updates, and service discovery. Go binaries compile down to a single statically linked executable, making them exceptionally lightweight and fast to start in containers.

```
+-------------------------------------------------------------+
|                     Kubernetes Cluster                      |
|                                                             |
|  +-------------------+        +---------------------------+ |
|  | Service (Port 80) | ─────> | Pod (Go App Container)    | |
|  +-------------------+        |  - Liveness:  /healthz    | |
|                               |  - Readiness: /ready      | |
|                               |  - SIGTERM -> Shutdown    | |
|                               +---------------------------+ |
+-------------------------------------------------------------+
```

---

## 1. Multi-Stage Production Dockerfile

Statically compiled binaries with `CGO_ENABLED=0` allow deploying to `distroless` or `scratch` images with image sizes often under **20 MB**.

```dockerfile
# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Leverage Docker layer caching for modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static binary without CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server ./main.go

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /app/server /server

# Run as non-root user
USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
```

---

## 2. Production Go Server: Probes & Graceful Shutdown

Kubernetes sends `SIGTERM` before killing a pod. If an application immediately halts, inflight requests are dropped. A graceful shutdown listens for `SIGTERM` and gives inflight requests time to complete.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type AppServer struct {
	server  *http.Server
	isReady atomic.Bool
}

func NewAppServer(addr string) *AppServer {
	app := &AppServer{}

	mux := http.NewServeMux()

	// 1. Liveness Probe: tells K8s if process is alive (restarts if fails)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// 2. Readiness Probe: tells K8s if app can accept user traffic (pulls out of service if fails)
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if app.isReady.Load() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "READY")
		} else {
			http.Error(w, "NOT_READY", http.StatusServiceUnavailable)
		}
	})

	// Business API
	mux.HandleFunc("GET /api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "pong"}`)
	})

	app.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return app
}

func main() {
	app := NewAppServer(":8080")

	// Start server in background
	go func() {
		log.Printf("Starting HTTP server on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server Listen error: %v", err)
		}
	}()

	// Simulate initialization (e.g. connecting to DB, warming cache)
	time.Sleep(1 * time.Second)
	app.isReady.Store(true)
	log.Println("Application is ready to receive traffic.")

	// Listen for OS signals (SIGTERM from Kubernetes, SIGINT from Ctrl+C)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM, os.Interrupt)

	sig := <-shutdownChan
	log.Printf("Received termination signal [%s]. Initiating graceful shutdown...", sig)

	// 1. Immediately fail readiness probe so K8s stops routing new traffic
	app.isReady.Store(false)

	// 2. Allow existing connections up to 15 seconds to drain
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped. Exiting.")
}
```

---

## 3. Kubernetes Deployment & Service Manifests

Save as `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-backend-deployment
  labels:
    app: go-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-backend
  template:
    metadata:
      labels:
        app: go-backend
    spec:
      containers:
        - name: go-backend
          image: myregistry.com/myuser/go-backend:v1.0.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          # Health Probes
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 2
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: go-backend-service
spec:
  type: ClusterIP
  selector:
    app: go-backend
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

---

## 4. Programmatic Cluster Interaction via `client-go`

If you are developing custom Kubernetes Operators or tools interacting with cluster APIs, use official `k8s.io/client-go`:

```go
package main

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ListPodsInCluster() {
	// In-cluster configuration (reads service account token inside pod)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Cannot load in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		fmt.Printf("Pod: %s (Status: %s)\n", pod.Name, pod.Status.Phase)
	}
}
```
