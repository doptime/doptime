package apis

type EmbeddingModel string

const EmbeddingModelChatGpt EmbeddingModel = "chatgpt"

type Text2VectorIn struct {
	Text           []string
	EmbeddingModel string
}
type Text2VectorOut struct {
	Vector [][]float32
}
