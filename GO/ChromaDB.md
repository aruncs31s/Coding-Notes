---
id: chromadb_go
aliases:
  - ChromaDB in Go
  - Vector Database Go
tags:
  - go
  - ai
  - vector-database
  - chromadb
  - embeddings
dg-publish: true
---

# ChromaDB in Go: Vector Storage, Ingestion & Export

**ChromaDB** is an open-source vector database tailored for AI applications, Retrieval-Augmented Generation (RAG), and semantic search. It stores high-dimensional embeddings alongside unstructured text documents and structured metadata.

```
+----------------+      Text Embeddings      +-------------------------+
| Document/Text  | ────────────────────────> |        ChromaDB         |
+----------------+                           | (Collection: "docs")    |
                                             |  - Vector ID: "doc-1"   |
+----------------+      Similarity Query     |  - Embedding: [0.12...] |
| Search Query   | ────────────────────────> |  - Metadata: {"cat":...}|
+----------------+                           +-------------------------+
                                                          │
                                                    Top-K Results
                                                          ▼
                                             [ Closest Document Matches ]
```

---

## 1. Client Setup & Architecture

ChromaDB exposes a standardized HTTP REST API (default port `8000`). Go applications interact with ChromaDB either through the official client (`github.com/amikos-tech/chroma-go`) or via a lightweight HTTP wrapper.

### Installing Client Library
```bash
go get github.com/amikos-tech/chroma-go
```

---

## 2. Ingestion & Similarity Search

Below is a complete implementation showing collection creation, embedding ingestion, and semantic querying:

```go
package main

import (
	"context"
	"fmt"
	"log"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"
)

func main() {
	ctx := context.Background()

	// 1. Initialize ChromaDB client (points to running Chroma server)
	client, err := chroma.NewClient(chroma.WithBasePath("http://localhost:8000"))
	if err != nil {
		log.Fatalf("Error connecting to Chroma: %v", err)
	}

	// 2. Create or retrieve collection with Cosine distance metric
	collectionName := "knowledge_base"
	metadata := map[string]interface{}{
		"description": "Golang documentation & notes",
	}

	col, err := client.CreateCollection(
		ctx,
		collectionName,
		metadata,
		true, // createIfNotExists
		types.NewCosineDistanceFunction(),
	)
	if err != nil {
		log.Fatalf("Error creating collection: %v", err)
	}

	fmt.Printf("Collection '%s' ready!\n", col.Name)

	// 3. Prepare documents, IDs, metadata, and embeddings (e.g. 384 or 1536 dim vectors)
	// For demonstration, using dummy 3-dimensional embeddings:
	ids := []string{"doc_go_1", "doc_go_2", "doc_python_1"}
	documents := []string{
		"Go is an open-source programming language supported by Google.",
		"Go features lightweight concurrency primitives called goroutines.",
		"Python is an interpreted, high-level, general-purpose language.",
	}
	metadatas := []map[string]interface{}{
		{"language": "go", "topic": "intro"},
		{"language": "go", "topic": "concurrency"},
		{"language": "python", "topic": "intro"},
	}
	embeddings := [][]float32{
		{0.95, 0.12, 0.05},
		{0.88, 0.35, 0.10},
		{0.15, 0.85, 0.70},
	}

	// 4. Ingest / Upsert data into collection
	recordSet, err := types.NewRecordSet(
		types.WithRowIDs(ids),
		types.WithTexts(documents),
		types.WithMetadatas(metadatas),
		types.WithEmbeddings(embeddings),
	)
	if err != nil {
		log.Fatalf("Error creating record set: %v", err)
	}

	_, err = col.Add(ctx, recordSet)
	if err != nil {
		log.Fatalf("Failed to add vectors: %v", err)
	}
	fmt.Println("Successfully indexed documents.")

	// 5. Query: Find top-2 most relevant documents for a query vector
	queryVector := []float32{0.92, 0.20, 0.08} // Close to Go documents
	results, err := col.Query(
		ctx,
		[][]float32{queryVector},
		2,   // n_results
		nil, // where metadata filter (optional)
		nil, // where_document filter (optional)
		[]types.Include{types.IncludeDocuments, types.IncludeMetadatas, types.IncludeDistances},
	)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Println("\n--- Query Results ---")
	for i := range results.Documents[0] {
		fmt.Printf("Rank #%d: Doc: '%s' | Distance: %f | Metadata: %v\n",
			i+1,
			results.Documents[0][i],
			results.Distances[0][i],
			results.Metadatas[0][i],
		)
	}
}
```

---

## 3. Exporting Data from ChromaDB in Go

When performing backups, migrations, or data pipelines, you can export all stored documents, embeddings, and metadata into a structured JSON or CSV format.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"
)

type ExportRecord struct {
	ID        string                 `json:"id"`
	Document  string                 `json:"document"`
	Metadata  map[string]interface{} `json:"metadata"`
	Embedding []float32              `json:"embedding,omitempty"`
}

// ExportCollection dumps all items in a collection to a local JSON file.
func ExportCollection(ctx context.Context, client *chroma.Client, colName string, outputPath string) error {
	col, err := client.GetCollection(ctx, colName, nil)
	if err != nil {
		return fmt.Errorf("could not get collection: %w", err)
	}

	// Retrieve all records with embeddings and documents
	res, err := col.Get(
		ctx,
		nil, // all IDs
		nil, // no where filter
		nil, // no document filter
		[]types.Include{types.IncludeDocuments, types.IncludeMetadatas, types.IncludeEmbeddings},
	)
	if err != nil {
		return fmt.Errorf("failed fetching collection data: %w", err)
	}

	records := make([]ExportRecord, 0, len(res.Ids))
	for i, id := range res.Ids {
		var doc string
		if len(res.Documents) > i {
			doc = res.Documents[i]
		}

		var meta map[string]interface{}
		if len(res.Metadatas) > i {
			meta = res.Metadatas[i]
		}

		var emb []float32
		if len(res.Embeddings) > i {
			emb = res.Embeddings[i]
		}

		records = append(records, ExportRecord{
			ID:        id,
			Document:  doc,
			Metadata:  meta,
			Embedding: emb,
		})
	}

	// Write to JSON file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("failed to encode records: %w", err)
	}

	fmt.Printf("Successfully exported %d records to %s\n", len(records), outputPath)
	return nil
}
```

---

## 4. Metadata Filtering (`Where` Clauses)

ChromaDB allows combining vector similarity with relational metadata filtering:
- `$eq`: Equals
- `$ne`: Not equal
- `$gt`, `$gte`, `$lt`, `$lte`: Numeric comparisons
- `$in`, `$nin`: Inclusion
- `$and`, `$or`: Logical combinators

```go
// Query only documents where language == "go" and topic == "concurrency"
whereFilter := map[string]interface{}{
    "$and": []map[string]interface{}{
        {"language": map[string]interface{}{"$eq": "go"}},
        {"topic": map[string]interface{}{"$eq": "concurrency"}},
    },
}
```
