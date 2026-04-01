package domain

type Chunk struct {
	ID         string
	DocumentID string
	Text       string
}

type Chunker interface {
	Chunk(text string, documentID string) ([]Chunk, error)
}
