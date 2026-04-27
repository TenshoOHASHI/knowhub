package swagger

// ===== Auth =====

// RegisterRequest ユーザー登録リクエスト
type RegisterRequest struct {
	Username string `json:"username" example:"testuser"`
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
}

// LoginRequest ログインリクエスト
type LoginRequest struct {
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
}

// ===== Wiki =====

// CreateArticleRequest 記事作成リクエスト
type CreateArticleRequest struct {
	Title      string `json:"title" example:"Go入門ガイド"`
	Content    string `json:"content" example:"# Hello World\n本文です"`
	CategoryID string `json:"category_id" example:"1"`
	Visibility string `json:"visibility" example:"public"`
}

// UpdateArticleRequest 記事更新リクエスト
type UpdateArticleRequest struct {
	Title      string `json:"title" example:"更新タイトル"`
	Content    string `json:"content" example:"更新内容"`
	Visibility string `json:"visibility" example:"public"`
}

// ===== Category =====

// CreateCategoryRequest カテゴリ作成リクエスト
type CreateCategoryRequest struct {
	Name     string `json:"name" example:"Go"`
	ParentID string `json:"parent_id" example:""`
}

// ===== Profile =====

// CreateProfileRequest プロフィール作成リクエスト
type CreateProfileRequest struct {
	Title       string `json:"title" example:"Backend Engineer"`
	Bio         string `json:"bio" example:"Go が好きです"`
	GithubUrl   string `json:"github_url" example:"https://github.com/user"`
	AvatarUrl   string `json:"avatar_url" example:"https://example.com/avatar.png"`
	TwitterUrl  string `json:"twitter_url" example:"https://twitter.com/user"`
	LinkedinUrl string `json:"linkedin_url" example:"https://linkedin.com/in/user"`
	WantedlyUrl string `json:"wantedly_url" example:"https://www.wantedly.com/id/user"`
	Skills      string `json:"skills" example:"Go,TypeScript,MySQL"`
	Languages   string `json:"languages" example:"Japanese,English"`
}

// ===== Portfolio =====

// CreatePortfolioItemRequest ポートフォリオ作成リクエスト
type CreatePortfolioItemRequest struct {
	Title       string `json:"title" example:"TenHub"`
	Description string `json:"description" example:"技術ナレッジベースプラットフォーム"`
	Url         string `json:"url" example:"https://github.com/user/project"`
	Status      string `json:"status" example:"developing"`
	Category    string `json:"category" example:"Web Application"`
	TechStack   string `json:"tech_stack" example:"Go,Next.js,MySQL"`
}

// UpdatePortfolioItemRequest ポートフォリオ更新リクエスト
type UpdatePortfolioItemRequest struct {
	Title       string `json:"title" example:"更新タイトル"`
	Description string `json:"description" example:"更新説明"`
	Url         string `json:"url" example:"https://example.com"`
	Status      string `json:"status" example:"completed"`
	Category    string `json:"category" example:"CLI Tool"`
	TechStack   string `json:"tech_stack" example:"Go,Docker"`
}

// ===== Upload =====

// UploadResponse 画像アップロードレスポンス
type UploadResponse struct {
	Url string `json:"url" example:"/uploads/1714567890123_avatar.png"`
}
