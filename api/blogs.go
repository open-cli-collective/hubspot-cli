package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// BlogPost represents a HubSpot blog post
type BlogPost struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Slug                 string                 `json:"slug,omitempty"`
	State                string                 `json:"state,omitempty"`
	AuthorName           string                 `json:"authorName,omitempty"`
	BlogAuthorID         string                 `json:"blogAuthorId,omitempty"`
	ContentGroupID       string                 `json:"contentGroupId,omitempty"`
	Domain               string                 `json:"domain,omitempty"`
	HTMLTitle            string                 `json:"htmlTitle,omitempty"`
	MetaDescription      string                 `json:"metaDescription,omitempty"`
	PostBody             string                 `json:"postBody,omitempty"`
	PostSummary          string                 `json:"postSummary,omitempty"`
	FeaturedImage        string                 `json:"featuredImage,omitempty"`
	FeaturedImageAltText string                 `json:"featuredImageAltText,omitempty"`
	PublishDate          string                 `json:"publishDate,omitempty"`
	CreatedAt            string                 `json:"created,omitempty"`
	UpdatedAt            string                 `json:"updated,omitempty"`
	Archived             bool                   `json:"archived,omitempty"`
	ArchivedAt           string                 `json:"archivedAt,omitempty"`
	CurrentState         string                 `json:"currentState,omitempty"`
	TagIDs               []int64                `json:"tagIds,omitempty"`
	LayoutSections       map[string]interface{} `json:"layoutSections,omitempty"`
}

// BlogPostList represents a paginated list of blog posts
type BlogPostList struct {
	Results []BlogPost `json:"results"`
	Paging  *Paging    `json:"paging,omitempty"`
	Total   int        `json:"total,omitempty"`
}

// BlogAuthor represents a HubSpot blog author
type BlogAuthor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"fullName,omitempty"`
	Email       string `json:"email,omitempty"`
	Slug        string `json:"slug,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Bio         string `json:"bio,omitempty"`
	Website     string `json:"website,omitempty"`
	Twitter     string `json:"twitter,omitempty"`
	Facebook    string `json:"facebook,omitempty"`
	LinkedIn    string `json:"linkedin,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	CreatedAt   string `json:"created,omitempty"`
	UpdatedAt   string `json:"updated,omitempty"`
}

// BlogAuthorList represents a paginated list of blog authors
type BlogAuthorList struct {
	Results []BlogAuthor `json:"results"`
	Paging  *Paging      `json:"paging,omitempty"`
	Total   int          `json:"total,omitempty"`
}

// BlogTag represents a HubSpot blog tag
type BlogTag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug,omitempty"`
	Language  string `json:"language,omitempty"`
	CreatedAt string `json:"created,omitempty"`
	UpdatedAt string `json:"updated,omitempty"`
}

// BlogTagList represents a paginated list of blog tags
type BlogTagList struct {
	Results []BlogTag `json:"results"`
	Paging  *Paging   `json:"paging,omitempty"`
	Total   int       `json:"total,omitempty"`
}

// ListBlogPosts retrieves blog posts with pagination
func (c *Client) ListBlogPosts(opts ListOptions) (*BlogPostList, error) {
	url := fmt.Sprintf("%s/cms/v3/blogs/posts", c.BaseURL)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.After != "" {
		params["after"] = opts.After
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result BlogPostList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog posts response: %w", err)
	}

	return &result, nil
}

// GetBlogPost retrieves a single blog post by ID
func (c *Client) GetBlogPost(postID string) (*BlogPost, error) {
	if postID == "" {
		return nil, fmt.Errorf("post ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/blogs/posts/%s", c.BaseURL, postID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result BlogPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog post response: %w", err)
	}

	return &result, nil
}

// CreateBlogPost creates a new blog post
func (c *Client) CreateBlogPost(post map[string]interface{}) (*BlogPost, error) {
	url := fmt.Sprintf("%s/cms/v3/blogs/posts", c.BaseURL)

	body, err := c.post(url, post)
	if err != nil {
		return nil, err
	}

	var result BlogPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog post response: %w", err)
	}

	return &result, nil
}

// UpdateBlogPost updates an existing blog post
func (c *Client) UpdateBlogPost(postID string, updates map[string]interface{}) (*BlogPost, error) {
	if postID == "" {
		return nil, fmt.Errorf("post ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/blogs/posts/%s", c.BaseURL, postID)

	body, err := c.patch(url, updates)
	if err != nil {
		return nil, err
	}

	var result BlogPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog post response: %w", err)
	}

	return &result, nil
}

// DeleteBlogPost archives a blog post
func (c *Client) DeleteBlogPost(postID string) error {
	if postID == "" {
		return fmt.Errorf("post ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/blogs/posts/%s", c.BaseURL, postID)

	_, err := c.delete(url)
	return err
}

// ListBlogAuthors retrieves blog authors with pagination
func (c *Client) ListBlogAuthors(opts ListOptions) (*BlogAuthorList, error) {
	url := fmt.Sprintf("%s/cms/v3/blogs/authors", c.BaseURL)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.After != "" {
		params["after"] = opts.After
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result BlogAuthorList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog authors response: %w", err)
	}

	return &result, nil
}

// ListBlogTags retrieves blog tags with pagination
func (c *Client) ListBlogTags(opts ListOptions) (*BlogTagList, error) {
	url := fmt.Sprintf("%s/cms/v3/blogs/tags", c.BaseURL)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.After != "" {
		params["after"] = opts.After
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result BlogTagList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blog tags response: %w", err)
	}

	return &result, nil
}
