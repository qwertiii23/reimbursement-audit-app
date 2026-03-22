package response

type LoginResponse struct {
	Token string    `json:"token"`
	User  UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}
