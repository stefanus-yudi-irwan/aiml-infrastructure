package testsuite

import (
	"aiml-infrastructure/internal/base/vectordb"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

type NamedEntity struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

func GenerateTestData(id string, embeddingLength int, sentenceLength int) vectordb.Data {
	return vectordb.Data{
		ID:         id,
		DocumentID: uuid.New().String(),
		ChunkID:    uuid.New().String(),
		Content:    GenerateRandomParagraph(sentenceLength),
		Embedding:  GenerateRandomEmbedding(embeddingLength),
		Metadata: vectordb.MetadataMap{
			"tags": []string{"finance", "banking"},
			"entities": []NamedEntity{
				{
					Text: "Bank Mandiri",
					Type: "ORGANIZATION",
				},
				{
					Text: "Charles",
					Type: "PERSON",
				},
				{
					Text: "Beach",
					Type: "PLACE",
				},
			},
		},
	}
}

func GenerateRandomEmbedding(length int) []float32 {
	embedding := make([]float32, length)
	for index := range embedding {
		embedding[index] = rand.Float32()
	}
	return embedding
}

var sentences = []string{
	"Artificial intelligence is increasingly being used to improve business processes.",
	"Machine learning models can analyze large amounts of data and identify useful patterns.",
	"Efficient AI focuses on reducing computational requirements while maintaining acceptable performance.",
	"Vector databases are commonly used to store embeddings for semantic search applications.",
	"Retrieval augmented generation combines information retrieval with large language models.",
	"Organizations need reliable data pipelines to support the deployment of AI applications.",
	"Metadata can be used to filter and organize documents during the retrieval process.",
	"Cloud infrastructure provides scalable resources for deploying machine learning workloads.",
	"Data quality and data governance are important factors in successful AI implementation.",
	"Modern recommendation systems often combine multiple retrieval and ranking techniques.",
}

func GenerateRandomParagraph(sentenceCount int) string {
	if sentenceCount <= 0 {
		return ""
	}

	result := make([]string, sentenceCount)
	for index := range result {
		result[index] = sentences[rand.Intn(len(sentences))]
	}

	return strings.Join(result, " ")
}
