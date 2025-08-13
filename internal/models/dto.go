package models

type CreatePostDTO struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}

type AddCommentDTO struct {
	CommentMsg string `json:"comment"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestToken struct {
	Token string `json:"token"`
}
